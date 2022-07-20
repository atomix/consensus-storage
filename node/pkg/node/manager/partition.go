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

func (m *PartitionManager) publish(event *multiraftv1.PartitionEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	log.Infow("Publish PartitionEvent",
		logging.Stringer("PartitionEvent", event))
	for _, listener := range m.listeners {
		listener <- *event
	}
}

func (m *PartitionManager) Watch(ctx context.Context, ch chan<- multiraftv1.PartitionEvent) {
	m.listenerID++
	id := m.listenerID
	m.mu.Lock()
	m.listeners[id] = ch
	m.mu.Unlock()

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
	raftConfig, ok := m.getRaftConfig(config)
	if !ok {
		return nil
	}
	initialMembers := m.getInitialMembers(cluster, config)
	return m.node.host.StartCluster(initialMembers, false, m.newStateMachine, raftConfig)
}

func (m *PartitionManager) join(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raftConfig, ok := m.getRaftConfig(config)
	if !ok {
		return nil
	}
	initialMembers := m.getInitialMembers(cluster, config)
	return m.node.host.StartCluster(initialMembers, true, m.newStateMachine, raftConfig)
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

func (m *PartitionManager) getInitialMembers(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) map[uint64]dragonboat.Target {
	replicaConfigs := make(map[multiraftv1.NodeID]multiraftv1.ReplicaConfig)
	for _, replicaConfig := range cluster.Replicas {
		replicaConfigs[replicaConfig.NodeID] = replicaConfig
	}

	initialMembers := make(map[uint64]dragonboat.Target)
	for _, memberConfig := range config.Members {
		if replicaConfig, ok := replicaConfigs[memberConfig.NodeID]; ok {
			initialMembers[uint64(replicaConfig.NodeID)] = fmt.Sprintf("%s:%d", replicaConfig.Host, replicaConfig.RaftPort)
		}
	}
	return initialMembers
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
		NodeID:             uint64(m.id),
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
