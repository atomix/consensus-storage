// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/raft"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"io"
	"sync"
)

const dataDir = "/var/lib/atomix/data"

func newManager(protocol *Protocol, registry *primitive.Registry, options Options) *Manager {
	var rtt uint64 = 250
	if options.Config.HeartbeatPeriod != nil {
		rtt = uint64(options.Config.HeartbeatPeriod.Milliseconds())
	}

	listener := newEventListener()
	address := fmt.Sprintf("%s:%d", options.RaftService.Host, options.RaftService.Port)
	nodeConfig := raftconfig.NodeHostConfig{
		WALDir:              dataDir,
		NodeHostDir:         dataDir,
		RTTMillisecond:      rtt,
		RaftAddress:         address,
		RaftEventListener:   listener,
		SystemEventListener: listener,
	}

	node, err := dragonboat.NewNodeHost(nodeConfig)
	if err != nil {
		panic(err)
	}

	return &Manager{
		Options:  options,
		id:       options.NodeID,
		host:     node,
		protocol: protocol,
		registry: registry,
		listener: listener,
	}
}

type Manager struct {
	Options
	id       multiraftv1.NodeID
	host     *dragonboat.NodeHost
	config   *multiraftv1.ClusterConfig
	protocol *Protocol
	registry *primitive.Registry
	listener *eventListener
	mu       sync.RWMutex
}

func (m *Manager) ID() multiraftv1.NodeID {
	return m.id
}

func (m *Manager) Watch(ctx context.Context, ch chan<- multiraftv1.MultiRaftEvent) {
	m.listener.listen(ctx, ch)
}

func (m *Manager) Bootstrap(config multiraftv1.ClusterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a listener to wait for a leader to be elected
	eventCh := make(chan multiraftv1.MultiRaftEvent)
	m.Watch(context.Background(), eventCh)

	m.config = &config

	partitionIDs := make([]multiraftv1.PartitionID, 0, len(config.Partitions))
	for _, partitionConfig := range config.Partitions {
		raftConfig, ok := m.getRaftConfig(partitionConfig)
		if !ok {
			continue
		}
		initialMembers := m.getInitialMembers(partitionConfig)
		if err := m.host.StartCluster(initialMembers, false, m.newStateMachine, raftConfig); err != nil {
			return err
		}
		partitionIDs = append(partitionIDs, partitionConfig.PartitionID)
	}
	m.waitForPartitionLeaders(eventCh, partitionIDs...)
	return nil
}

func (m *Manager) Join(config multiraftv1.PartitionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	raftConfig, ok := m.getRaftConfig(config)
	if !ok {
		return nil
	}
	initialMembers := m.getInitialMembers(config)

	// Create a listener to wait for a leader to be elected
	eventCh := make(chan multiraftv1.MultiRaftEvent)
	m.Watch(context.Background(), eventCh)

	if err := m.host.StartCluster(initialMembers, true, m.newStateMachine, raftConfig); err != nil {
		return err
	}

	m.waitForPartitionLeaders(eventCh, config.PartitionID)
	return nil
}

func (m *Manager) Leave(partitionID multiraftv1.PartitionID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	if err := m.host.StopCluster(uint64(partitionID)); err != nil {
		return err
	}
	m.protocol.removePartition(partitionID)
	return nil
}

func (m *Manager) getRaftConfig(config multiraftv1.PartitionConfig) (raftconfig.Config, bool) {
	var member *multiraftv1.MemberConfig
	for _, memberConfig := range config.Members {
		if memberConfig.NodeID == m.id {
			member = &memberConfig
			break
		}
	}

	if member == nil {
		return raftconfig.Config{}, false
	}

	var rtt uint64 = 250
	if m.Config.HeartbeatPeriod != nil {
		rtt = uint64(m.Config.HeartbeatPeriod.Milliseconds())
	}

	electionRTT := uint64(10)
	if m.Config.ElectionTimeout != nil {
		electionRTT = uint64(m.Config.ElectionTimeout.Milliseconds()) / rtt
	}

	return raftconfig.Config{
		NodeID:             uint64(m.id),
		ClusterID:          uint64(config.PartitionID),
		ElectionRTT:        electionRTT,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    m.Config.SnapshotEntryThreshold,
		CompactionOverhead: m.Config.CompactionRetainEntries,
		IsObserver:         member.Role == multiraftv1.MemberConfig_OBSERVER,
		IsWitness:          member.Role == multiraftv1.MemberConfig_WITNESS,
	}, true
}

func (m *Manager) getInitialMembers(config multiraftv1.PartitionConfig) map[uint64]dragonboat.Target {
	replicaConfigs := make(map[multiraftv1.NodeID]multiraftv1.ReplicaConfig)
	for _, replicaConfig := range m.config.Replicas {
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

func (m *Manager) newStateMachine(clusterID, nodeID uint64) dbstatemachine.IStateMachine {
	partition := newPartition(m.protocol, multiraftv1.PartitionID(clusterID))
	m.protocol.addPartition(partition)
	return newStateMachine(raft.NewStateMachine(m.registry), partition.streams)
}

func (m *Manager) waitForPartitionLeaders(eventCh chan multiraftv1.MultiRaftEvent, partitions ...multiraftv1.PartitionID) {
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
				partition := m.protocol.partitions[leader.LeaderUpdated.PartitionID]
				partition.setLeader(leader.LeaderUpdated.Term, leader.LeaderUpdated.Leader)
			}
		}
	}()
	<-startedCh
}

func (m *Manager) Shutdown(ctx context.Context, partitionID multiraftv1.PartitionID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.host.StopNode(uint64(partitionID), uint64(m.id)); err != nil {
		return err
	}
	return nil
}

func newStateMachine(state raft.StateMachine, streams *streamRegistry) dbstatemachine.IStateMachine {
	return &stateMachine{
		state:   state,
		streams: streams,
	}
}

type stateMachine struct {
	state   raft.StateMachine
	streams *streamRegistry
}

func (s *stateMachine) Update(bytes []byte) (dbstatemachine.Result, error) {
	logEntry := &multiraftv1.RaftLogEntry{}
	if err := proto.Unmarshal(bytes, logEntry); err != nil {
		return dbstatemachine.Result{}, err
	}

	stream := s.streams.lookup(logEntry.StreamID)
	s.state.Command(&logEntry.Command, stream)
	return dbstatemachine.Result{}, nil
}

func (s *stateMachine) Lookup(value interface{}) (interface{}, error) {
	query := value.(queryContext)
	s.state.Query(query.input, query.stream)
	return nil, nil
}

func (s *stateMachine) SaveSnapshot(writer io.Writer, collection dbstatemachine.ISnapshotFileCollection, i <-chan struct{}) error {
	return s.state.Snapshot(snapshot.NewWriter(writer))
}

func (s *stateMachine) RecoverFromSnapshot(reader io.Reader, files []dbstatemachine.SnapshotFile, i <-chan struct{}) error {
	return s.state.Recover(snapshot.NewReader(reader))
}

func (s *stateMachine) Close() error {
	return nil
}

type queryContext struct {
	input  *multiraftv1.QueryInput
	stream stream.WriteStream[*multiraftv1.QueryOutput]
}

func newNodeServer(node *Manager) multiraftv1.NodeServer {
	return &nodeServer{
		node: node,
	}
}

type nodeServer struct {
	node *Manager
}

func (s *nodeServer) Bootstrap(ctx context.Context, request *multiraftv1.BootstrapRequest) (*multiraftv1.BootstrapResponse, error) {
	log.Debugw("Bootstrap",
		logging.Stringer("BootstrapRequest", request))
	if err := s.node.Bootstrap(request.Cluster); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Bootstrap",
			logging.Stringer("BootstrapRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.BootstrapResponse{}
	log.Debugw("Bootstrap",
		logging.Stringer("BootstrapRequest", request),
		logging.Stringer("BootstrapResponse", response))
	return response, nil
}

func (s *nodeServer) Join(ctx context.Context, request *multiraftv1.JoinRequest) (*multiraftv1.JoinResponse, error) {
	log.Debugw("Join",
		logging.Stringer("JoinRequest", request))
	if err := s.node.Join(request.Partition); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Join",
			logging.Stringer("JoinRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.JoinResponse{}
	log.Debugw("Join",
		logging.Stringer("JoinRequest", request),
		logging.Stringer("JoinResponse", response))
	return response, nil
}

func (s *nodeServer) Leave(ctx context.Context, request *multiraftv1.LeaveRequest) (*multiraftv1.LeaveResponse, error) {
	log.Debugw("Leave",
		logging.Stringer("LeaveRequest", request))
	if err := s.node.Leave(request.PartitionID); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Leave",
			logging.Stringer("LeaveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.LeaveResponse{}
	log.Debugw("Leave",
		logging.Stringer("LeaveRequest", request),
		logging.Stringer("LeaveResponse", response))
	return response, nil
}

func (s *nodeServer) Watch(request *multiraftv1.WatchNodeRequest, server multiraftv1.Node_WatchServer) error {
	ch := make(chan multiraftv1.MultiRaftEvent)
	s.node.Watch(server.Context(), ch)
	for event := range ch {
		response := &multiraftv1.WatchNodeResponse{
			Timestamp: event.Timestamp,
		}
		switch t := event.Event.(type) {
		case *multiraftv1.MultiRaftEvent_MemberReady:
			response.Event = &multiraftv1.WatchNodeResponse_MemberReady{
				MemberReady: t.MemberReady,
			}
		case *multiraftv1.MultiRaftEvent_LeaderUpdated:
			response.Event = &multiraftv1.WatchNodeResponse_LeaderUpdated{
				LeaderUpdated: t.LeaderUpdated,
			}
		case *multiraftv1.MultiRaftEvent_MembershipChanged:
			response.Event = &multiraftv1.WatchNodeResponse_MembershipChanged{
				MembershipChanged: t.MembershipChanged,
			}
		case *multiraftv1.MultiRaftEvent_SendSnapshotStarted:
			response.Event = &multiraftv1.WatchNodeResponse_SendSnapshotStarted{
				SendSnapshotStarted: t.SendSnapshotStarted,
			}
		case *multiraftv1.MultiRaftEvent_SendSnapshotCompleted:
			response.Event = &multiraftv1.WatchNodeResponse_SendSnapshotCompleted{
				SendSnapshotCompleted: t.SendSnapshotCompleted,
			}
		case *multiraftv1.MultiRaftEvent_SendSnapshotAborted:
			response.Event = &multiraftv1.WatchNodeResponse_SendSnapshotAborted{
				SendSnapshotAborted: t.SendSnapshotAborted,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotReceived:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotReceived{
				SnapshotReceived: t.SnapshotReceived,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotRecovered:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotRecovered{
				SnapshotRecovered: t.SnapshotRecovered,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotCreated:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotCreated{
				SnapshotCreated: t.SnapshotCreated,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotCompacted:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotCompacted{
				SnapshotCompacted: t.SnapshotCompacted,
			}
		case *multiraftv1.MultiRaftEvent_LogCompacted:
			response.Event = &multiraftv1.WatchNodeResponse_LogCompacted{
				LogCompacted: t.LogCompacted,
			}
		case *multiraftv1.MultiRaftEvent_LogdbCompacted:
			response.Event = &multiraftv1.WatchNodeResponse_LogdbCompacted{
				LogdbCompacted: t.LogdbCompacted,
			}
		case *multiraftv1.MultiRaftEvent_ConnectionEstablished:
			response.Event = &multiraftv1.WatchNodeResponse_ConnectionEstablished{
				ConnectionEstablished: t.ConnectionEstablished,
			}
		case *multiraftv1.MultiRaftEvent_ConnectionFailed:
			response.Event = &multiraftv1.WatchNodeResponse_ConnectionFailed{
				ConnectionFailed: t.ConnectionFailed,
			}
		}
		err := server.Send(response)
		if err != nil {
			return err
		}
	}
	return nil
}
