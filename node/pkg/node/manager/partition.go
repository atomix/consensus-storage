// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/protocol"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"sync"
	"time"
)

func newPartitionManager(id multiraftv1.PartitionID, node *NodeManager, protocol *protocol.PartitionProtocol) *PartitionManager {
	return &PartitionManager{
		id:        id,
		node:      node,
		protocol:  protocol,
		listeners: make(map[int]chan<- multiraftv1.PartitionEvent),
	}
}

type PartitionManager struct {
	id         multiraftv1.PartitionID
	node       *NodeManager
	protocol   *protocol.PartitionProtocol
	leader     multiraftv1.NodeID
	replicas   map[multiraftv1.NodeID]multiraftv1.ReplicaConfig
	listeners  map[int]chan<- multiraftv1.PartitionEvent
	listenerID int
	mu         sync.RWMutex
}

func (m *PartitionManager) ID() multiraftv1.PartitionID {
	return m.id
}

func (m *PartitionManager) Protocol() *protocol.PartitionProtocol {
	return m.protocol
}

func (m *PartitionManager) getServiceConfig() multiraftv1.ServiceConfig {
	var config multiraftv1.ServiceConfig
	leader, ok := m.replicas[m.leader]
	if ok {
		config.Leader = fmt.Sprintf("%s:%d", leader.Host, leader.ApiPort)
	}
	for _, replica := range m.replicas {
		if replica.NodeID != m.leader {
			config.Followers = append(config.Followers, fmt.Sprintf("%s:%d", replica.Host, replica.ApiPort))
		}
	}
	return config
}

func (m *PartitionManager) publish(event *multiraftv1.PartitionEvent) {
	log.Infow("Publish PartitionEvent",
		logging.Stringer("PartitionEvent", event))
	switch e := event.Event.(type) {
	case *multiraftv1.PartitionEvent_LeaderUpdated:
		m.protocol.SetLeader(e.LeaderUpdated.Term, e.LeaderUpdated.Leader)
		m.mu.Lock()
		defer m.mu.Unlock()
		for _, listener := range m.listeners {
			listener <- *event
		}

		m.leader = e.LeaderUpdated.Leader
		serviceConfigEvent := &multiraftv1.PartitionEvent{
			Timestamp:   event.Timestamp,
			PartitionID: event.PartitionID,
			Event: &multiraftv1.PartitionEvent_ServiceConfigChanged{
				ServiceConfigChanged: &multiraftv1.ServiceConfigChangedEvent{
					Config: m.getServiceConfig(),
				},
			},
		}
		log.Infow("Publish PartitionEvent",
			logging.Stringer("PartitionEvent", serviceConfigEvent))
		for _, listener := range m.listeners {
			listener <- *serviceConfigEvent
		}
	default:
		m.mu.RLock()
		defer m.mu.RUnlock()
		for _, listener := range m.listeners {
			listener <- *event
		}
	}
}

func (m *PartitionManager) Watch(ctx context.Context, ch chan<- multiraftv1.PartitionEvent) {
	m.listenerID++
	id := m.listenerID
	m.mu.Lock()
	m.listeners[id] = ch
	m.mu.Unlock()

	m.mu.RLock()
	if m.replicas != nil {
		go func() {
			ch <- multiraftv1.PartitionEvent{
				Timestamp:   time.Now(),
				PartitionID: m.id,
				Event: &multiraftv1.PartitionEvent_ServiceConfigChanged{
					ServiceConfigChanged: &multiraftv1.ServiceConfigChangedEvent{
						Config: m.getServiceConfig(),
					},
				},
			}
			m.mu.RUnlock()
		}()
	} else {
		m.mu.RUnlock()
	}

	go func() {
		<-ctx.Done()
		m.mu.Lock()
		close(ch)
		delete(m.listeners, id)
		m.mu.Unlock()
	}()
}

func (m *PartitionManager) newStateMachine(clusterID, nodeID uint64) dbstatemachine.IStateMachine {
	return protocol.NewStateMachine(m.protocol)
}

func (m *PartitionManager) bootstrap(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateConfig(cluster, config)
	if _, ok := m.replicas[m.node.id]; !ok {
		return nil
	}
	members := make(map[uint64]dragonboat.Target)
	for _, replica := range m.replicas {
		members[uint64(replica.NodeID)] = fmt.Sprintf("%s:%d", replica.Host, replica.RaftPort)
	}
	raftConfig, ok := m.getRaftConfig(config)
	if !ok {
		return nil
	}
	if err := m.node.host.StartCluster(members, false, m.newStateMachine, raftConfig); err != nil {
		return err
	}
	return nil
}

func (m *PartitionManager) join(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateConfig(cluster, config)
	if _, ok := m.replicas[m.node.id]; !ok {
		return nil
	}
	members := make(map[uint64]dragonboat.Target)
	for _, replica := range m.replicas {
		members[uint64(replica.NodeID)] = fmt.Sprintf("%s:%d", replica.Host, replica.RaftPort)
	}
	raftConfig, ok := m.getRaftConfig(config)
	if !ok {
		return nil
	}
	return m.node.host.StartCluster(members, true, m.newStateMachine, raftConfig)
}

func (m *PartitionManager) leave() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.node.host.StopCluster(uint64(m.id))
}

func (m *PartitionManager) shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.node.host.StopNode(uint64(m.id), uint64(m.node.id)); err != nil {
		return err
	}
	return nil
}

func (m *PartitionManager) updateConfig(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) {
	m.replicas = make(map[multiraftv1.NodeID]multiraftv1.ReplicaConfig)
	for _, member := range config.Members {
		for _, replica := range cluster.Replicas {
			if replica.NodeID == member.NodeID {
				m.replicas[replica.NodeID] = replica
				break
			}
		}
	}
}

func (m *PartitionManager) getRaftConfig(config multiraftv1.PartitionConfig) (raftconfig.Config, bool) {
	var member *multiraftv1.MemberConfig
	for _, memberConfig := range config.Members {
		if memberConfig.NodeID == m.node.id {
			member = &memberConfig
			break
		}
	}

	if member == nil {
		return raftconfig.Config{}, false
	}

	var rtt uint64 = 250
	if m.node.config.HeartbeatPeriod != nil {
		rtt = uint64(m.node.config.HeartbeatPeriod.Milliseconds())
	}

	electionRTT := uint64(10)
	if m.node.config.ElectionTimeout != nil {
		electionRTT = uint64(m.node.config.ElectionTimeout.Milliseconds()) / rtt
	}

	return raftconfig.Config{
		NodeID:             uint64(m.node.id),
		ClusterID:          uint64(config.PartitionID),
		ElectionRTT:        electionRTT,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    m.node.config.SnapshotEntryThreshold,
		CompactionOverhead: m.node.config.CompactionRetainEntries,
		IsObserver:         member.Role == multiraftv1.MemberConfig_OBSERVER,
		IsWitness:          member.Role == multiraftv1.MemberConfig_WITNESS,
	}, true
}
