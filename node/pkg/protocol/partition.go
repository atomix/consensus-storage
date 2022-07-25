// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	statemachine "github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/stream"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/lni/dragonboat/v3"
	raftconfig "github.com/lni/dragonboat/v3/config"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"google.golang.org/grpc/metadata"
	"sync"
	"sync/atomic"
	"time"
)

func newPartition(id multiraftv1.PartitionID, node *Node) *Partition {
	return &Partition{
		id:        id,
		node:      node,
		listeners: make(map[int]chan<- multiraftv1.PartitionEvent),
		streams:   stream.NewRegistry(),
	}
}

type Partition struct {
	id         multiraftv1.PartitionID
	node       *Node
	leader     uint64
	term       uint64
	replicas   map[multiraftv1.NodeID]multiraftv1.ReplicaConfig
	listeners  map[int]chan<- multiraftv1.PartitionEvent
	listenerID int
	streams    *stream.Registry
	mu         sync.RWMutex
}

func (p *Partition) ID() multiraftv1.PartitionID {
	return p.id
}

func (p *Partition) setLeader(term multiraftv1.Term, leader multiraftv1.NodeID) {
	atomic.StoreUint64(&p.term, uint64(term))
	atomic.StoreUint64(&p.leader, uint64(leader))
}

func (p *Partition) getLeader() (multiraftv1.Term, multiraftv1.NodeID) {
	return multiraftv1.Term(atomic.LoadUint64(&p.term)), multiraftv1.NodeID(atomic.LoadUint64(&p.leader))
}

func (p *Partition) getServiceConfig() multiraftv1.ServiceConfig {
	var config multiraftv1.ServiceConfig
	_, leaderID := p.getLeader()
	leader, ok := p.replicas[leaderID]
	if ok {
		config.Leader = fmt.Sprintf("%s:%d", leader.Host, leader.ApiPort)
	}
	for _, replica := range p.replicas {
		if replica.NodeID != leaderID {
			config.Followers = append(config.Followers, fmt.Sprintf("%s:%d", replica.Host, replica.ApiPort))
		}
	}
	return config
}

func (p *Partition) publish(event *multiraftv1.PartitionEvent) {
	log.Infow("Publish PartitionEvent",
		logging.Stringer("PartitionEvent", event))
	switch e := event.Event.(type) {
	case *multiraftv1.PartitionEvent_LeaderUpdated:
		p.setLeader(e.LeaderUpdated.Term, e.LeaderUpdated.Leader)
		p.mu.Lock()
		defer p.mu.Unlock()
		for _, listener := range p.listeners {
			listener <- *event
		}

		serviceConfigEvent := &multiraftv1.PartitionEvent{
			Timestamp:   event.Timestamp,
			PartitionID: event.PartitionID,
			Event: &multiraftv1.PartitionEvent_ServiceConfigChanged{
				ServiceConfigChanged: &multiraftv1.ServiceConfigChangedEvent{
					Config: p.getServiceConfig(),
				},
			},
		}
		log.Infow("Publish PartitionEvent",
			logging.Stringer("PartitionEvent", serviceConfigEvent))
		for _, listener := range p.listeners {
			listener <- *serviceConfigEvent
		}
	default:
		p.mu.RLock()
		defer p.mu.RUnlock()
		for _, listener := range p.listeners {
			listener <- *event
		}
	}
}

func (p *Partition) Watch(ctx context.Context, ch chan<- multiraftv1.PartitionEvent) {
	p.listenerID++
	id := p.listenerID
	p.mu.Lock()
	p.listeners[id] = ch
	p.mu.Unlock()

	p.mu.RLock()
	if p.replicas != nil {
		go func() {
			ch <- multiraftv1.PartitionEvent{
				Timestamp:   time.Now(),
				PartitionID: p.id,
				Event: &multiraftv1.PartitionEvent_ServiceConfigChanged{
					ServiceConfigChanged: &multiraftv1.ServiceConfigChangedEvent{
						Config: p.getServiceConfig(),
					},
				},
			}
			p.mu.RUnlock()
		}()
	} else {
		p.mu.RUnlock()
	}

	go func() {
		<-ctx.Done()
		p.mu.Lock()
		close(ch)
		delete(p.listeners, id)
		p.mu.Unlock()
	}()
}

func (p *Partition) newStateMachine(clusterID, nodeID uint64) dbstatemachine.IStateMachine {
	return statemachine.NewStateMachine(p.streams, p.node.registry)
}

func (p *Partition) bootstrap(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updateConfig(cluster, config)
	if _, ok := p.replicas[p.node.id]; !ok {
		return nil
	}
	members := make(map[uint64]dragonboat.Target)
	for _, replica := range p.replicas {
		members[uint64(replica.NodeID)] = fmt.Sprintf("%s:%d", replica.Host, replica.RaftPort)
	}
	raftConfig, ok := p.getRaftConfig(config)
	if !ok {
		return nil
	}
	if err := p.node.host.StartCluster(members, false, p.newStateMachine, raftConfig); err != nil {
		return err
	}
	return nil
}

func (p *Partition) join(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updateConfig(cluster, config)
	if _, ok := p.replicas[p.node.id]; !ok {
		return nil
	}
	members := make(map[uint64]dragonboat.Target)
	for _, replica := range p.replicas {
		members[uint64(replica.NodeID)] = fmt.Sprintf("%s:%d", replica.Host, replica.RaftPort)
	}
	raftConfig, ok := p.getRaftConfig(config)
	if !ok {
		return nil
	}
	return p.node.host.StartCluster(members, true, p.newStateMachine, raftConfig)
}

func (p *Partition) leave() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.node.host.StopCluster(uint64(p.id))
}

func (p *Partition) shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.node.host.StopNode(uint64(p.id), uint64(p.node.id)); err != nil {
		return err
	}
	return nil
}

func (p *Partition) updateConfig(cluster multiraftv1.ClusterConfig, config multiraftv1.PartitionConfig) {
	p.replicas = make(map[multiraftv1.NodeID]multiraftv1.ReplicaConfig)
	for _, member := range config.Members {
		for _, replica := range cluster.Replicas {
			if replica.NodeID == member.NodeID {
				p.replicas[replica.NodeID] = replica
				break
			}
		}
	}
}

func (p *Partition) getRaftConfig(config multiraftv1.PartitionConfig) (raftconfig.Config, bool) {
	var member *multiraftv1.MemberConfig
	for _, memberConfig := range config.Members {
		if memberConfig.NodeID == p.node.id {
			member = &memberConfig
			break
		}
	}

	if member == nil {
		return raftconfig.Config{}, false
	}

	var rtt uint64 = 250
	if p.node.config.HeartbeatPeriod != nil {
		rtt = uint64(p.node.config.HeartbeatPeriod.Milliseconds())
	}

	electionRTT := uint64(10)
	if p.node.config.ElectionTimeout != nil {
		electionRTT = uint64(p.node.config.ElectionTimeout.Milliseconds()) / rtt
	}

	return raftconfig.Config{
		NodeID:             uint64(p.node.id),
		ClusterID:          uint64(config.PartitionID),
		ElectionRTT:        electionRTT,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    p.node.config.SnapshotEntryThreshold,
		CompactionOverhead: p.node.config.CompactionRetainEntries,
		IsObserver:         member.Role == multiraftv1.MemberConfig_OBSERVER,
		IsWitness:          member.Role == multiraftv1.MemberConfig_WITNESS,
	}, true
}

func (p *Partition) Command(ctx context.Context, command *multiraftv1.CommandInput) (*multiraftv1.CommandOutput, error) {
	resultCh := make(chan streams.Result[*multiraftv1.CommandOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.commitCommand(ctx, command, streams.NewChannelStream[*multiraftv1.CommandOutput](resultCh)); err != nil {
			errCh <- err
		}
	}()

	select {
	case result, ok := <-resultCh:
		if !ok {
			err := errors.NewCanceled("stream closed")
			return nil, err
		}

		if result.Failed() {
			return nil, result.Error
		}

		return result.Value, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Partition) StreamCommand(ctx context.Context, input *multiraftv1.CommandInput, out streams.WriteStream[*multiraftv1.CommandOutput]) error {
	in := streams.NewBufferedStream[*multiraftv1.CommandOutput]()
	go func() {
		if err := p.commitCommand(ctx, input, in); err != nil {
			in.Error(err)
			return
		}
	}()
	go func() {
		for {
			result, ok := in.Receive()
			if !ok {
				out.Close()
				return
			}
			out.Send(result)
		}
	}()
	return nil
}

func (p *Partition) commitCommand(ctx context.Context, input *multiraftv1.CommandInput, stream streams.WriteStream[*multiraftv1.CommandOutput]) error {
	term, leader := p.getLeader()
	if leader != p.node.id {
		return errors.NewUnavailable("not the leader")
	}

	streamID := p.streams.Register(term, stream)
	entry := &multiraftv1.RaftLogEntry{
		StreamID: streamID,
		Command:  *input,
	}

	bytes, err := proto.Marshal(entry)
	if err != nil {
		return errors.NewInternal("failed to marshal RaftLogEntry: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, clientTimeout)
	defer cancel()
	if _, err := p.node.host.SyncPropose(ctx, p.node.host.GetNoOPSession(uint64(p.id)), bytes); err != nil {
		return wrapError(err)
	}
	return nil
}

func (p *Partition) Query(ctx context.Context, query *multiraftv1.QueryInput) (*multiraftv1.QueryOutput, error) {
	resultCh := make(chan streams.Result[*multiraftv1.QueryOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.applyQuery(ctx, query, streams.NewChannelStream[*multiraftv1.QueryOutput](resultCh)); err != nil {
			errCh <- err
		}
	}()

	select {
	case result, ok := <-resultCh:
		if !ok {
			err := errors.NewCanceled("stream closed")
			return nil, err
		}

		if result.Failed() {
			return nil, result.Error
		}

		return result.Value, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Partition) StreamQuery(ctx context.Context, input *multiraftv1.QueryInput, out streams.WriteStream[*multiraftv1.QueryOutput]) error {
	in := streams.NewBufferedStream[*multiraftv1.QueryOutput]()
	go func() {
		if err := p.applyQuery(ctx, input, in); err != nil {
			in.Error(err)
			return
		}
	}()
	go func() {
		for {
			result, ok := in.Receive()
			if !ok {
				out.Close()
				return
			}
			out.Send(result)
		}
	}()
	return nil
}

func (p *Partition) applyQuery(ctx context.Context, input *multiraftv1.QueryInput, output streams.WriteStream[*multiraftv1.QueryOutput]) error {
	query := &stream.Query{
		Input:  input,
		Stream: output,
	}
	md, _ := metadata.FromIncomingContext(ctx)
	sync := md["Sync"] != nil
	if sync {
		ctx, cancel := context.WithTimeout(ctx, clientTimeout)
		defer cancel()
		if _, err := p.node.host.SyncRead(ctx, uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	} else {
		if _, err := p.node.host.StaleRead(uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	}
	return nil
}
