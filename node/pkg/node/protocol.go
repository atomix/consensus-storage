// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	"fmt"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/protocol/node"
	"github.com/atomix/runtime/sdk/pkg/protocol/statemachine"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"sync"
	"time"
)

var log = logging.GetLogger()

type Index uint64

type Term uint64

type MemberID uint32

type GroupID uint32

type SequenceNum uint64

const (
	defaultDataDir                 = "/var/lib/atomix/data"
	defaultSnapshotEntryThreshold  = 10000
	defaultCompactionRetainEntries = 1000
	defaultClientTimeout           = time.Minute
)

func newProtocol(config NodeConfig, registry *statemachine.PrimitiveTypeRegistry) *Protocol {
	var rtt uint64 = 250
	if config.HeartbeatPeriod != nil {
		rtt = uint64(config.HeartbeatPeriod.Milliseconds())
	}

	dataDir := config.DataDir
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	protocol := &Protocol{
		config:     config,
		registry:   registry,
		partitions: make(map[protocol.PartitionID]*Partition),
		watchers:   make(map[int]chan<- Event),
	}

	listener := newEventListener(protocol)
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	nodeConfig := raftconfig.NodeHostConfig{
		WALDir:              dataDir,
		NodeHostDir:         dataDir,
		RTTMillisecond:      rtt,
		RaftAddress:         address,
		RaftEventListener:   listener,
		SystemEventListener: listener,
	}

	host, err := dragonboat.NewNodeHost(nodeConfig)
	if err != nil {
		panic(err)
	}

	protocol.host = host
	return protocol
}

type Protocol struct {
	host       *dragonboat.NodeHost
	config     NodeConfig
	registry   *statemachine.PrimitiveTypeRegistry
	partitions map[protocol.PartitionID]*Partition
	watchers   map[int]chan<- Event
	watcherID  int
	mu         sync.RWMutex
}

func (n *Protocol) publish(event *Event) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	switch e := event.Event.(type) {
	case *Event_MemberReady:
		if partition, ok := n.partitions[protocol.PartitionID(e.MemberReady.GroupID)]; ok {
			partition.setReady()
		}
	case *Event_LeaderUpdated:
		if partition, ok := n.partitions[protocol.PartitionID(e.LeaderUpdated.GroupID)]; ok {
			partition.setLeader(e.LeaderUpdated.Term, e.LeaderUpdated.Leader)
		}
	}
	log.Infow("Publish Event",
		logging.Stringer("Event", event))
	for _, listener := range n.watchers {
		listener <- *event
	}
}

func (n *Protocol) Partition(partitionID protocol.PartitionID) (node.Partition, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	partition, ok := n.partitions[partitionID]
	return partition, ok
}

func (n *Protocol) Partitions() []node.Partition {
	n.mu.RLock()
	defer n.mu.RUnlock()
	partitions := make([]node.Partition, 0, len(n.partitions))
	for _, partition := range n.partitions {
		partitions = append(partitions, partition)
	}
	return partitions
}

func (n *Protocol) Watch(ctx context.Context, watcher chan<- Event) {
	n.watcherID++
	id := n.watcherID
	n.mu.Lock()
	n.watchers[id] = watcher
	for _, partition := range n.partitions {
		term, leader := partition.getLeader()
		if term > 0 {
			watcher <- Event{
				Event: &Event_LeaderUpdated{
					LeaderUpdated: &LeaderUpdatedEvent{
						MemberEvent: MemberEvent{
							GroupID:  GroupID(partition.ID()),
							MemberID: partition.memberID,
						},
						Term:   term,
						Leader: leader,
					},
				},
			}
		}
		ready := partition.getReady()
		if ready {
			watcher <- Event{
				Event: &Event_MemberReady{
					MemberReady: &MemberReadyEvent{
						MemberEvent: MemberEvent{
							GroupID:  GroupID(partition.ID()),
							MemberID: partition.memberID,
						},
					},
				},
			}
		}
	}
	n.mu.Unlock()

	go func() {
		<-ctx.Done()
		n.mu.Lock()
		close(watcher)
		delete(n.watchers, id)
		n.mu.Unlock()
	}()
}

func (n *Protocol) Bootstrap(config GroupConfig) error {
	raftConfig := n.getRaftConfig(config)
	members := make(map[uint64]dragonboat.Target)
	for _, member := range config.Members {
		members[uint64(member.MemberID)] = fmt.Sprintf("%s:%d", member.Host, member.Port)
	}
	if err := n.host.StartCluster(members, false, n.newStateMachine, raftConfig); err != nil {
		if err == dragonboat.ErrClusterAlreadyExist {
			return nil
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) Join(config GroupConfig) error {
	raftConfig := n.getRaftConfig(config)
	members := make(map[uint64]dragonboat.Target)
	for _, member := range config.Members {
		members[uint64(member.MemberID)] = fmt.Sprintf("%s:%d", member.Host, member.Port)
	}
	if err := n.host.StartCluster(members, true, n.newStateMachine, raftConfig); err != nil {
		if err == dragonboat.ErrClusterAlreadyExist {
			return nil
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) Leave(groupID GroupID) error {
	return n.host.StopCluster(uint64(groupID))
}

func (n *Protocol) newStateMachine(clusterID, nodeID uint64) dbstatemachine.IStateMachine {
	streams := newContext()
	partition := newPartition(protocol.PartitionID(clusterID), MemberID(nodeID), n.host, streams)
	n.mu.Lock()
	n.partitions[partition.ID()] = partition
	n.mu.Unlock()
	return newStateMachine(streams, n.registry)
}

func (n *Protocol) Shutdown() error {
	n.host.Stop()
	return nil
}

func (n *Protocol) getRaftConfig(config GroupConfig) raftconfig.Config {
	var rtt uint64 = 250
	if n.config.HeartbeatPeriod != nil {
		rtt = uint64(n.config.HeartbeatPeriod.Milliseconds())
	}

	electionRTT := uint64(10)
	if n.config.ElectionTimeout != nil {
		electionRTT = uint64(n.config.ElectionTimeout.Milliseconds()) / rtt
	}

	snapshotEntryThreshold := n.config.SnapshotEntryThreshold
	if snapshotEntryThreshold == 0 {
		snapshotEntryThreshold = defaultSnapshotEntryThreshold
	}
	compactionRetainEntries := n.config.CompactionRetainEntries
	if compactionRetainEntries == 0 {
		compactionRetainEntries = defaultCompactionRetainEntries
	}

	return raftconfig.Config{
		NodeID:             uint64(config.MemberID),
		ClusterID:          uint64(config.GroupID),
		ElectionRTT:        electionRTT,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    snapshotEntryThreshold,
		CompactionOverhead: compactionRetainEntries,
		IsObserver:         config.Role == MemberRole_OBSERVER,
		IsWitness:          config.Role == MemberRole_WITNESS,
	}
}

func wrapError(err error) error {
	switch err {
	case dragonboat.ErrClusterNotFound,
		dragonboat.ErrClusterNotBootstrapped,
		dragonboat.ErrClusterNotInitialized,
		dragonboat.ErrClusterNotReady,
		dragonboat.ErrClusterClosed:
		return errors.NewUnavailable(err.Error())
	case dragonboat.ErrSystemBusy,
		dragonboat.ErrBadKey:
		return errors.NewUnavailable(err.Error())
	case dragonboat.ErrClosed,
		dragonboat.ErrNodeRemoved:
		return errors.NewUnavailable(err.Error())
	case dragonboat.ErrInvalidSession,
		dragonboat.ErrInvalidTarget,
		dragonboat.ErrInvalidAddress,
		dragonboat.ErrInvalidOperation:
		return errors.NewInvalid(err.Error())
	case dragonboat.ErrPayloadTooBig,
		dragonboat.ErrTimeoutTooSmall:
		return errors.NewForbidden(err.Error())
	case dragonboat.ErrDeadlineNotSet,
		dragonboat.ErrInvalidDeadline:
		return errors.NewInternal(err.Error())
	case dragonboat.ErrDirNotExist:
		return errors.NewInternal(err.Error())
	case dragonboat.ErrTimeout:
		return errors.NewTimeout(err.Error())
	case dragonboat.ErrCanceled:
		return errors.NewCanceled(err.Error())
	case dragonboat.ErrRejected:
		return errors.NewForbidden(err.Error())
	default:
		return errors.NewUnknown(err.Error())
	}
}
