// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	statemachine "github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	dbsm "github.com/lni/dragonboat/v3/statemachine"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

func newStateMachine(protocol *protocolContext, types *primitive.TypeRegistry) dbsm.IStateMachine {
	context := &stateMachineContext{
		scheduler: newScheduler(),
	}
	sm := &stateMachine{
		stateMachineContext: context,
		protocol:            protocol,
		sm: statemachine.NewStateMachine(context, func(smCtx statemachine.Context) statemachine.SessionManager {
			return session.NewManager(context, func(sessionCtx session.Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) session.PrimitiveManager {
				return primitive.NewManager(sessionCtx, types)
			})
		}),
	}
	return sm
}

type stateMachineContext struct {
	index     atomic.Uint64
	time      atomic.Value
	scheduler *stateMachineScheduler
}

func (c *stateMachineContext) Log() logging.Logger {
	return log
}

func (c *stateMachineContext) update(ts time.Time) (statemachine.Index, time.Time) {
	index := statemachine.Index(c.index.Add(1))
	t := c.time.Load().(time.Time)
	if ts.After(t) {
		c.time.Store(ts)
		return index, ts
	}
	return index, t
}

func (c *stateMachineContext) Index() statemachine.Index {
	return statemachine.Index(c.index.Load())
}

func (c *stateMachineContext) Time() time.Time {
	return c.time.Load().(time.Time)
}

func (c *stateMachineContext) Scheduler() statemachine.Scheduler {
	return c.scheduler
}

var _ statemachine.Context = (*stateMachineContext)(nil)

type stateMachine struct {
	*stateMachineContext
	protocol *protocolContext
	queryID  atomic.Uint64
	sm       *statemachine.StateMachine
}

func (s *stateMachine) Update(bytes []byte) (dbsm.Result, error) {
	proposal := &multiraftv1.RaftProposal{}
	if err := proto.Unmarshal(bytes, proposal); err != nil {
		return dbsm.Result{}, err
	}

	// Update the timestamp
	index, ts := s.update(proposal.Timestamp)

	// Run scheduled tasks for the updated timestamp
	s.scheduler.tick(ts)

	// Get the local output stream (if any)
	stream := s.protocol.getStream(proposal.Term, proposal.SequenceNum)

	// Submit the proposal to the state machine
	s.sm.Propose(newProposal(
		statemachine.ProposalID(index),
		proposal.Proposal,
		stream))

	// Run scheduled tasks for this index
	s.scheduler.tock(index)
	return dbsm.Result{}, nil
}

func (s *stateMachine) Lookup(value interface{}) (interface{}, error) {
	query := value.(*protocolQuery)
	s.sm.Query(newQuery(statemachine.QueryID(s.queryID.Add(1)), query.input, query.stream))
	return nil, nil
}

func (s *stateMachine) SaveSnapshot(writer io.Writer, collection dbsm.ISnapshotFileCollection, i <-chan struct{}) error {
	sw := snapshot.NewWriter(writer)
	log.Infow("Persisting state to snapshot",
		logging.Uint64("Index", uint64(s.Index())),
		logging.Time("Time", s.Time()))
	snapshot := &multiraftv1.Snapshot{
		Index:     multiraftv1.Index(s.Index()),
		Timestamp: s.Time(),
	}
	if err := sw.WriteMessage(snapshot); err != nil {
		return err
	}
	return s.sm.Snapshot(sw)
}

func (s *stateMachine) RecoverFromSnapshot(reader io.Reader, files []dbsm.SnapshotFile, i <-chan struct{}) error {
	sr := snapshot.NewReader(reader)
	snapshot := &multiraftv1.Snapshot{}
	if err := sr.ReadMessage(snapshot); err != nil {
		return err
	}
	log.Infow("Recovering state from snapshot",
		logging.Uint64("Index", uint64(snapshot.Index)),
		logging.Time("Time", snapshot.Timestamp))
	s.index.Store(uint64(snapshot.Index))
	s.time.Store(snapshot.Timestamp)
	return s.sm.Recover(sr)
}

func (s *stateMachine) Close() error {
	return nil
}

func newExecution[T statemachine.ExecutionID, I, O proto.Message](id T, input I, stream streams.WriteStream[O]) *stateMachineExecution[T, I, O] {
	return &stateMachineExecution[T, I, O]{
		id:     id,
		input:  input,
		stream: stream,
	}
}

type stateMachineExecution[T statemachine.ExecutionID, I, O proto.Message] struct {
	id       T
	input    I
	stream   streams.WriteStream[O]
	state    atomic.Int32
	watching atomic.Bool
	watchers map[string]statemachine.WatchFunc[statemachine.Phase]
	mu       sync.RWMutex
}

func (e stateMachineExecution[T, I, O]) ID() T {
	return e.id
}

func (e stateMachineExecution[T, I, O]) Log() logging.Logger {
	return log
}

func (e *stateMachineExecution[T, I, O]) Watch(watcher statemachine.WatchFunc[statemachine.Phase]) statemachine.CancelFunc {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.watching.CompareAndSwap(false, true) {
		e.watchers = make(map[string]statemachine.WatchFunc[statemachine.Phase])
	}
	id := uuid.New().String()
	e.watchers[id] = watcher
	return func() {
		delete(e.watchers, id)
	}
}

func (e stateMachineExecution[T, I, O]) Input() I {
	return e.input
}

func (e stateMachineExecution[T, I, O]) Output(output O) {
	e.stream.Value(output)
}

func (e stateMachineExecution[T, I, O]) Error(err error) {
	e.stream.Error(err)
}

func (e *stateMachineExecution[T, I, O]) close(phase statemachine.Phase) {
	if e.state.CompareAndSwap(int32(statemachine.Runnnig), int32(phase)) {
		if e.watching.Load() {
			e.mu.RLock()
			defer e.mu.RUnlock()
			for _, watcher := range e.watchers {
				watcher(phase)
			}
		}
	}
}

func (e *stateMachineExecution[T, I, O]) Cancel() {
	e.stream.Close()
	e.close(statemachine.Canceled)
}

func (e stateMachineExecution[T, I, O]) Close() {
	e.stream.Close()
	e.close(statemachine.Complete)
}

func newProposal(id statemachine.ProposalID, input *multiraftv1.StateMachineProposalInput, stream streams.WriteStream[*multiraftv1.StateMachineProposalOutput]) *stateMachineProposal {
	return &stateMachineProposal{
		stateMachineExecution: newExecution[statemachine.ProposalID](id, input, stream),
	}
}

type stateMachineProposal struct {
	*stateMachineExecution[statemachine.ProposalID, *multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput]
}

var _ statemachine.Proposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput] = (*stateMachineProposal)(nil)

func newQuery(id statemachine.QueryID, input *multiraftv1.StateMachineQueryInput, stream streams.WriteStream[*multiraftv1.StateMachineQueryOutput]) *stateMachineQuery {
	return &stateMachineQuery{
		stateMachineExecution: newExecution[statemachine.QueryID, *multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput](id, input, stream),
	}
}

type stateMachineQuery struct {
	*stateMachineExecution[statemachine.QueryID, *multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput]
}

var _ statemachine.Query[*multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput] = (*stateMachineQuery)(nil)
