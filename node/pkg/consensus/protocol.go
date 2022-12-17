// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package consensus

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
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = logging.GetLogger()

type Index uint64

type Term uint64

type MemberID uint32

type GroupID uint32

type SequenceNum uint64

func NewProtocol(config NodeConfig, registry *statemachine.PrimitiveTypeRegistry, opts ...Option) *Protocol {
	var options Options
	options.apply(opts...)

	protocol := &Protocol{
		config:     config,
		registry:   registry,
		partitions: make(map[protocol.PartitionID]*Partition),
		watchers:   make(map[int]chan<- Event),
	}

	listener := newEventListener(protocol)
	address := fmt.Sprintf("%s:%d", options.Host, options.Port)
	nodeConfig := raftconfig.NodeHostConfig{
		WALDir:              config.GetDataDir(),
		NodeHostDir:         config.GetDataDir(),
		RTTMillisecond:      uint64(config.GetRTT().Milliseconds()),
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

func (n *Protocol) GetConfig(groupID GroupID) (GroupConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	membership, err := n.host.SyncGetClusterMembership(ctx, uint64(groupID))
	if err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return GroupConfig{}, errors.NewNotFound(err.Error())
		}
		return GroupConfig{}, wrapError(err)
	}

	var members []MemberConfig
	for nodeID, address := range membership.Nodes {
		member, err := getMember(nodeID, address, MemberRole_MEMBER)
		if err != nil {
			return GroupConfig{}, err
		}
		members = append(members, member)
	}
	for nodeID, address := range membership.Observers {
		member, err := getMember(nodeID, address, MemberRole_OBSERVER)
		if err != nil {
			return GroupConfig{}, err
		}
		members = append(members, member)
	}
	for nodeID, address := range membership.Witnesses {
		member, err := getMember(nodeID, address, MemberRole_WITNESS)
		if err != nil {
			return GroupConfig{}, err
		}
		members = append(members, member)
	}
	for nodeID := range membership.Removed {
		members = append(members, MemberConfig{
			MemberID: MemberID(nodeID),
			Role:     MemberRole_REMOVED,
		})
	}
	return GroupConfig{
		GroupID: groupID,
		Members: members,
	}, nil
}

func getMember(nodeID uint64, address string, role MemberRole) (MemberConfig, error) {
	parts := strings.Split(address, ":")
	host, portS := parts[0], parts[1]
	port, err := strconv.Atoi(portS)
	if err != nil {
		return MemberConfig{}, err
	}
	return MemberConfig{
		MemberID: MemberID(nodeID),
		Host:     host,
		Port:     int32(port),
		Role:     MemberRole_MEMBER,
	}, nil
}

func (n *Protocol) BootstrapGroup(ctx context.Context, groupID GroupID, memberID MemberID, config RaftConfig, members ...MemberConfig) error {
	var member *MemberConfig
	for _, m := range members {
		if m.MemberID == memberID {
			member = &m
			break
		}
	}

	if member == nil {
		return errors.NewInvalid("unknown member %d", memberID)
	}

	raftConfig := n.getRaftConfig(groupID, memberID, member.Role, config)
	targets := make(map[uint64]dragonboat.Target)
	for _, member := range members {
		targets[uint64(member.MemberID)] = fmt.Sprintf("%s:%d", member.Host, member.Port)
	}
	if err := n.host.StartCluster(targets, false, n.newStateMachine, raftConfig); err != nil {
		if err == dragonboat.ErrClusterAlreadyExist {
			return nil
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) AddMember(ctx context.Context, groupID GroupID, member MemberConfig, version uint64) error {
	address := fmt.Sprintf("%s:%d", member.Host, member.Port)
	if err := n.host.SyncRequestAddNode(ctx, uint64(groupID), uint64(member.MemberID), address, version); err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return errors.NewNotFound(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) RemoveMember(ctx context.Context, groupID GroupID, memberID MemberID, version uint64) error {
	if err := n.host.SyncRequestDeleteNode(ctx, uint64(groupID), uint64(memberID), version); err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return errors.NewNotFound(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) JoinGroup(ctx context.Context, groupID GroupID, memberID MemberID, config RaftConfig) error {
	raftConfig := n.getRaftConfig(groupID, memberID, MemberRole_MEMBER, config)
	if err := n.host.StartCluster(map[uint64]dragonboat.Target{}, true, n.newStateMachine, raftConfig); err != nil {
		if err == dragonboat.ErrClusterAlreadyExist {
			return errors.NewAlreadyExists(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) LeaveGroup(ctx context.Context, groupID GroupID) error {
	if err := n.host.StopCluster(uint64(groupID)); err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return errors.NewNotFound(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) DeleteData(ctx context.Context, groupID GroupID, memberID MemberID) error {
	if err := n.host.SyncRemoveData(ctx, uint64(groupID), uint64(memberID)); err != nil {
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) newStateMachine(clusterID, nodeID uint64) dbstatemachine.IStateMachine {
	streams := newContext()
	partition := newPartition(protocol.PartitionID(clusterID), MemberID(nodeID), n.host, streams)
	n.mu.Lock()
	n.partitions[partition.ID()] = partition
	n.mu.Unlock()
	return newStateMachine(streams, n.registry)
}

func (n *Protocol) getRaftConfig(groupID GroupID, memberID MemberID, role MemberRole, config RaftConfig) raftconfig.Config {
	electionRTT := config.ElectionRTT
	if electionRTT == 0 {
		electionRTT = defaultElectionRTT
	}
	heartbeatRTT := config.HeartbeatRTT
	if heartbeatRTT == 0 {
		heartbeatRTT = defaultHeartbeatRTT
	}
	snapshotEntries := config.SnapshotEntries
	if snapshotEntries == 0 {
		snapshotEntries = defaultSnapshotEntries
	}
	compactionOverhead := config.CompactionOverhead
	if compactionOverhead == 0 {
		compactionOverhead = defaultCompactionOverhead
	}
	return raftconfig.Config{
		NodeID:                 uint64(memberID),
		ClusterID:              uint64(groupID),
		ElectionRTT:            electionRTT,
		HeartbeatRTT:           heartbeatRTT,
		CheckQuorum:            true,
		SnapshotEntries:        snapshotEntries,
		CompactionOverhead:     compactionOverhead,
		MaxInMemLogSize:        config.MaxInMemLogSize,
		OrderedConfigChange:    config.OrderedConfigChange,
		DisableAutoCompactions: config.DisableAutoCompactions,
		IsObserver:             role == MemberRole_OBSERVER,
		IsWitness:              role == MemberRole_WITNESS,
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
