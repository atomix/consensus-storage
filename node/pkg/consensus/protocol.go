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

type ReplicaID uint32

type ShardID uint32

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
	case *Event_ReplicaReady:
		if partition, ok := n.partitions[protocol.PartitionID(e.ReplicaReady.ShardID)]; ok {
			partition.setReady()
		}
	case *Event_LeaderUpdated:
		if partition, ok := n.partitions[protocol.PartitionID(e.LeaderUpdated.ShardID)]; ok {
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
						ReplicaEvent: ReplicaEvent{
							ShardID:   ShardID(partition.ID()),
							ReplicaID: partition.replicaID,
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
				Event: &Event_ReplicaReady{
					ReplicaReady: &ReplicaReadyEvent{
						ReplicaEvent: ReplicaEvent{
							ShardID:   ShardID(partition.ID()),
							ReplicaID: partition.replicaID,
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

func (n *Protocol) GetConfig(shardID ShardID) (ShardConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	membership, err := n.host.SyncGetClusterMembership(ctx, uint64(shardID))
	if err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return ShardConfig{}, errors.NewNotFound(err.Error())
		}
		return ShardConfig{}, wrapError(err)
	}

	var replicas []ReplicaConfig
	for replicaID, address := range membership.Nodes {
		member, err := getReplica(replicaID, address, ReplicaRole_MEMBER)
		if err != nil {
			return ShardConfig{}, err
		}
		replicas = append(replicas, member)
	}
	for replicaID, address := range membership.Observers {
		member, err := getReplica(replicaID, address, ReplicaRole_OBSERVER)
		if err != nil {
			return ShardConfig{}, err
		}
		replicas = append(replicas, member)
	}
	for replicaID, address := range membership.Witnesses {
		member, err := getReplica(replicaID, address, ReplicaRole_WITNESS)
		if err != nil {
			return ShardConfig{}, err
		}
		replicas = append(replicas, member)
	}
	for replicaID := range membership.Removed {
		replicas = append(replicas, ReplicaConfig{
			ReplicaID: ReplicaID(replicaID),
			Role:      ReplicaRole_REMOVED,
		})
	}
	return ShardConfig{
		ShardID:  shardID,
		Replicas: replicas,
	}, nil
}

func getReplica(replicaID uint64, address string, role ReplicaRole) (ReplicaConfig, error) {
	parts := strings.Split(address, ":")
	host, portS := parts[0], parts[1]
	port, err := strconv.Atoi(portS)
	if err != nil {
		return ReplicaConfig{}, err
	}
	return ReplicaConfig{
		ReplicaID: ReplicaID(replicaID),
		Host:      host,
		Port:      int32(port),
		Role:      role,
	}, nil
}

func (n *Protocol) BootstrapShard(ctx context.Context, shardID ShardID, replicaID ReplicaID, config RaftConfig, replicas ...ReplicaConfig) error {
	var member *ReplicaConfig
	for _, r := range replicas {
		if r.ReplicaID == replicaID {
			member = &r
			break
		}
	}

	if member == nil {
		return errors.NewInvalid("unknown member %d", replicaID)
	}

	raftConfig := n.getRaftConfig(shardID, replicaID, member.Role, config)
	targets := make(map[uint64]dragonboat.Target)
	for _, member := range replicas {
		targets[uint64(member.ReplicaID)] = fmt.Sprintf("%s:%d", member.Host, member.Port)
	}
	if err := n.host.StartCluster(targets, false, n.newStateMachine, raftConfig); err != nil {
		if err == dragonboat.ErrClusterAlreadyExist {
			return nil
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) AddReplica(ctx context.Context, shardID ShardID, member ReplicaConfig, version uint64) error {
	address := fmt.Sprintf("%s:%d", member.Host, member.Port)
	if err := n.host.SyncRequestAddNode(ctx, uint64(shardID), uint64(member.ReplicaID), address, version); err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return errors.NewNotFound(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) RemoveReplica(ctx context.Context, shardID ShardID, replicaID ReplicaID, version uint64) error {
	if err := n.host.SyncRequestDeleteNode(ctx, uint64(shardID), uint64(replicaID), version); err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return errors.NewNotFound(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) JoinShard(ctx context.Context, shardID ShardID, replicaID ReplicaID, config RaftConfig) error {
	raftConfig := n.getRaftConfig(shardID, replicaID, ReplicaRole_MEMBER, config)
	if err := n.host.StartCluster(map[uint64]dragonboat.Target{}, true, n.newStateMachine, raftConfig); err != nil {
		if err == dragonboat.ErrClusterAlreadyExist {
			return errors.NewAlreadyExists(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) LeaveShard(ctx context.Context, shardID ShardID) error {
	if err := n.host.StopCluster(uint64(shardID)); err != nil {
		if err == dragonboat.ErrClusterNotFound {
			return errors.NewNotFound(err.Error())
		}
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) DeleteData(ctx context.Context, shardID ShardID, replicaID ReplicaID) error {
	if err := n.host.SyncRemoveData(ctx, uint64(shardID), uint64(replicaID)); err != nil {
		return wrapError(err)
	}
	return nil
}

func (n *Protocol) newStateMachine(clusterID, replicaID uint64) dbstatemachine.IStateMachine {
	streams := newContext()
	partition := newPartition(protocol.PartitionID(clusterID), ReplicaID(replicaID), n.host, streams)
	n.mu.Lock()
	n.partitions[partition.ID()] = partition
	n.mu.Unlock()
	return newStateMachine(streams, n.registry)
}

func (n *Protocol) getRaftConfig(shardID ShardID, replicaID ReplicaID, role ReplicaRole, config RaftConfig) raftconfig.Config {
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
		NodeID:                 uint64(replicaID),
		ClusterID:              uint64(shardID),
		ElectionRTT:            electionRTT,
		HeartbeatRTT:           heartbeatRTT,
		CheckQuorum:            true,
		SnapshotEntries:        snapshotEntries,
		CompactionOverhead:     compactionOverhead,
		MaxInMemLogSize:        config.MaxInMemLogSize,
		OrderedConfigChange:    config.OrderedConfigChange,
		DisableAutoCompactions: config.DisableAutoCompactions,
		IsObserver:             role == ReplicaRole_OBSERVER,
		IsWitness:              role == ReplicaRole_WITNESS,
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
