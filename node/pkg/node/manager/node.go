// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/protocol"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
	"github.com/atomix/runtime/sdk/pkg/async"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	"sync"
)

var log = logging.GetLogger()

const (
	defaultDataDir = "/var/lib/atomix/data"
)

func NewNodeManager(registry *primitive.Registry, config multiraftv1.NodeConfig) *NodeManager {
	var rtt uint64 = 250
	if config.HeartbeatPeriod != nil {
		rtt = uint64(config.HeartbeatPeriod.Milliseconds())
	}

	dataDir := config.DataDir
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	node := &NodeManager{
		id:         config.NodeID,
		registry:   registry,
		config:     config,
		partitions: make(map[multiraftv1.PartitionID]*PartitionManager),
		listeners:  make(map[int]chan<- multiraftv1.NodeEvent),
	}

	listener := newEventListener(node)
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

	node.host = host
	return node
}

type NodeManager struct {
	id         multiraftv1.NodeID
	host       *dragonboat.NodeHost
	config     multiraftv1.NodeConfig
	cluster    *multiraftv1.ClusterConfig
	protocol   *protocol.NodeProtocol
	registry   *primitive.Registry
	partitions map[multiraftv1.PartitionID]*PartitionManager
	listeners  map[int]chan<- multiraftv1.NodeEvent
	listenerID int
	mu         sync.RWMutex
}

func (m *NodeManager) ID() multiraftv1.NodeID {
	return m.id
}

func (m *NodeManager) Protocol() *protocol.NodeProtocol {
	return m.protocol
}

func (m *NodeManager) publish(event *multiraftv1.NodeEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	log.Infow("Publish NodeEvent",
		logging.Stringer("NodeEvent", event))
	for _, listener := range m.listeners {
		listener <- *event
	}
}

func (m *NodeManager) Partition(partitionID multiraftv1.PartitionID) (*PartitionManager, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	partition, ok := m.partitions[partitionID]
	if !ok {
		return nil, errors.NewNotFound("partition %d not found", partitionID)
	}
	return partition, nil
}

func (m *NodeManager) Watch(ctx context.Context, ch chan<- multiraftv1.NodeEvent) {
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

func (m *NodeManager) Bootstrap(config multiraftv1.ClusterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cluster = &config

	if len(m.partitions) > 0 {
		return errors.NewConflict("node is already bootstrapped")
	}

	partitionIDs := make([]multiraftv1.PartitionID, 0, len(config.Partitions))
	for _, partitionConfig := range config.Partitions {
		partitionIDs = append(partitionIDs, partitionConfig.PartitionID)
	}
	m.protocol = protocol.NewNodeProtocol(m.id, m.host, m.registry)

	for _, partitionID := range partitionIDs {
		partition, ok := m.protocol.Partition(partitionID)
		if !ok {
			panic("partition not found")
		}
		m.partitions[partitionID] = newPartitionManager(partitionID, m, partition)
	}

	return async.IterAsync(len(config.Partitions), func(i int) error {
		partitionConfig := config.Partitions[i]
		return m.partitions[partitionConfig.PartitionID].bootstrap(config, partitionConfig)
	})
}

func (m *NodeManager) Join(config multiraftv1.PartitionConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cluster == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	partition, ok := m.partitions[config.PartitionID]
	if !ok {
		return errors.NewFault("partition %d not found", config.PartitionID)
	}
	return partition.join(*m.cluster, config)
}

func (m *NodeManager) Leave(partitionID multiraftv1.PartitionID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cluster == nil {
		return errors.NewForbidden("node has not yet been bootstrapped")
	}

	partition, ok := m.partitions[partitionID]
	if !ok {
		return errors.NewFault("partition %d not found", partitionID)
	}
	return partition.leave()
}

func (m *NodeManager) Shutdown(partitionID multiraftv1.PartitionID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.host.StopNode(uint64(partitionID), uint64(m.id)); err != nil {
		return err
	}
	return nil
}
