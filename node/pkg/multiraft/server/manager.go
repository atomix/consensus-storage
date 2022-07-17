// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/node/pkg/multiraft/statemachine"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"io"
	"sync"
)

const dataDir = "/var/lib/atomix/data"

func newNodeManager(id multiraftv1.NodeID, options Options) *NodeManager {
	return &NodeManager{
		Options:  options,
		id:       id,
		listener: newEventListener(),
	}
}

type NodeManager struct {
	Options
	id       multiraftv1.NodeID
	node     *dragonboat.NodeHost
	config   *multiraftv1.ClusterConfig
	service  *NodeService
	listener *eventListener
	mu       sync.RWMutex
}

func (n *NodeManager) ID() multiraftv1.NodeID {
	return n.id
}

func (n *NodeManager) Watch(ctx context.Context, ch chan<- multiraftv1.MultiRaftEvent) {
	n.listener.listen(ctx, ch)
}

func (n *NodeManager) Bootstrap(config multiraftv1.ClusterConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Create a listener to wait for a leader to be elected
	eventCh := make(chan multiraftv1.MultiRaftEvent)
	n.Watch(context.Background(), eventCh)

	var rtt uint64 = 250
	if n.Config.HeartbeatPeriod != nil {
		rtt = uint64(n.Config.HeartbeatPeriod.Milliseconds())
	}

	address := fmt.Sprintf("%s:%d", n.RaftService.Host, n.RaftService.Port)
	nodeConfig := raftconfig.NodeHostConfig{
		WALDir:              dataDir,
		NodeHostDir:         dataDir,
		RTTMillisecond:      rtt,
		RaftAddress:         address,
		RaftEventListener:   n.listener,
		SystemEventListener: n.listener,
	}

	node, err := dragonboat.NewNodeHost(nodeConfig)
	if err != nil {
		return err
	}
	n.service = newNodeService(n.id, node)
	n.config = &config

	partitionIDs := make([]multiraftv1.PartitionID, 0, len(config.Partitions))
	for _, partitionConfig := range config.Partitions {
		raftConfig, ok := n.getRaftConfig(partitionConfig)
		if !ok {
			continue
		}
		initialMembers := n.getInitialMembers(partitionConfig)
		if err := n.node.StartCluster(initialMembers, false, n.newStateMachine, raftConfig); err != nil {
			return err
		}
		partitionIDs = append(partitionIDs, partitionConfig.PartitionID)
	}
	n.waitForPartitionLeaders(eventCh, partitionIDs...)
	return nil
}

func (n *NodeManager) Join(config multiraftv1.PartitionConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.config == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	raftConfig, ok := n.getRaftConfig(config)
	if !ok {
		return nil
	}
	initialMembers := n.getInitialMembers(config)

	// Create a listener to wait for a leader to be elected
	eventCh := make(chan multiraftv1.MultiRaftEvent)
	n.Watch(context.Background(), eventCh)

	if err := n.node.StartCluster(initialMembers, true, n.newStateMachine, raftConfig); err != nil {
		return err
	}

	n.waitForPartitionLeaders(eventCh, config.PartitionID)
	return nil
}

func (n *NodeManager) Leave(partitionID multiraftv1.PartitionID) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.config == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	if err := n.node.StopCluster(uint64(partitionID)); err != nil {
		return err
	}
	n.service.removePartition(partitionID)
	return nil
}

func (n *NodeManager) getRaftConfig(config multiraftv1.PartitionConfig) (raftconfig.Config, bool) {
	var member *multiraftv1.MemberConfig
	for _, memberConfig := range config.Members {
		if memberConfig.NodeID == n.id {
			member = &memberConfig
			break
		}
	}

	if member == nil {
		return raftconfig.Config{}, false
	}

	var rtt uint64 = 250
	if n.Config.HeartbeatPeriod != nil {
		rtt = uint64(n.Config.HeartbeatPeriod.Milliseconds())
	}

	electionRTT := uint64(10)
	if n.Config.ElectionTimeout != nil {
		electionRTT = uint64(n.Config.ElectionTimeout.Milliseconds()) / rtt
	}

	return raftconfig.Config{
		NodeID:             uint64(n.id),
		ClusterID:          uint64(config.PartitionID),
		ElectionRTT:        electionRTT,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    n.Config.SnapshotEntryThreshold,
		CompactionOverhead: n.Config.CompactionRetainEntries,
		IsObserver:         member.Role == multiraftv1.MemberConfig_OBSERVER,
		IsWitness:          member.Role == multiraftv1.MemberConfig_WITNESS,
	}, true
}

func (n *NodeManager) getInitialMembers(config multiraftv1.PartitionConfig) map[uint64]dragonboat.Target {
	replicaConfigs := make(map[multiraftv1.NodeID]multiraftv1.ReplicaConfig)
	for _, replicaConfig := range n.config.Replicas {
		replicaConfigs[replicaConfig.NodeID] = replicaConfig
	}

	initialMembers := make(map[uint64]dragonboat.Target)
	for _, memberConfig := range config.Members {
		if replicaConfig, ok := replicaConfigs[memberConfig.NodeID]; ok {
			initialMembers[uint64(replicaConfig.NodeID)] = fmt.Sprintf("%s:%d", replicaConfig.Host, replicaConfig.Port)
		}
	}
	return initialMembers
}

func (n *NodeManager) newStateMachine(clusterID, nodeID uint64) dbstatemachine.IStateMachine {
	partition := newPartitionService(n.service, multiraftv1.PartitionID(clusterID))
	n.service.addPartition(partition)
	return newStateMachine(statemachine.NewManager(), partition.streams)
}

func (n *NodeManager) waitForPartitionLeaders(eventCh chan<- multiraftv1.MultiRaftEvent, partitions ...multiraftv1.PartitionID) {
	startedCh := make(chan struct{})
	go func() {
		startedPartitions := make(map[multiraftv1.PartitionID]bool)
		started := false
		for event := range eventCh {
			if leader, ok := event.Event.(*multiraftv1.MultiRaftEvent_LeaderUpdated); ok &&
				leader.LeaderUpdated.Term > 0 && leader.LeaderUpdated.Leader != 0 {
				startedPartitions[leader.LeaderUpdated.PartitionID] = true
				if !started && len(startedPartitions) == len(partitions) {
					close(startedCh)
					started = true
				}
				partition := n.service.partitions[leader.LeaderUpdated.PartitionID]
				partition.setLeader(leader.LeaderUpdated.Term, leader.LeaderUpdated.Leader)
			}
		}
	}()
	<-startedCh
}

func (n *NodeManager) Shutdown(ctx context.Context, partitionID multiraftv1.PartitionID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if err := n.node.StopNode(uint64(partitionID), uint64(n.id)); err != nil {
		return err
	}
	return nil
}

func newStateMachine(manager *statemachine.Manager, streams *streamRegistry) dbstatemachine.IStateMachine {
	return &stateMachine{
		manager: manager,
		streams: streams,
	}
}

type stateMachine struct {
	manager *statemachine.Manager
	streams *streamRegistry
}

func (s *stateMachine) Update(bytes []byte) (dbstatemachine.Result, error) {
	logEntry := &multiraftv1.RaftLogEntry{}
	if err := proto.Unmarshal(bytes, logEntry); err != nil {
		return dbstatemachine.Result{}, err
	}

	stream := s.streams.lookup(logEntry.StreamID)
	s.manager.Command(&logEntry.Command, stream)
	return dbstatemachine.Result{}, nil
}

func (s *stateMachine) Lookup(value interface{}) (interface{}, error) {
	query := value.(queryContext)
	s.manager.Query(query.input, query.stream)
	return nil, nil
}

func (s *stateMachine) SaveSnapshot(writer io.Writer, collection dbstatemachine.ISnapshotFileCollection, i <-chan struct{}) error {
	return s.manager.Snapshot(writer)
}

func (s *stateMachine) RecoverFromSnapshot(reader io.Reader, files []dbstatemachine.SnapshotFile, i <-chan struct{}) error {
	return s.manager.Restore(reader)
}

func (s *stateMachine) Close() error {
	return nil
}

type queryContext struct {
	input  *multiraftv1.QueryInput
	stream stream.WriteStream
}
