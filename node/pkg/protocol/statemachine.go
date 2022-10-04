// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
	dbsm "github.com/lni/dragonboat/v3/statemachine"
	"io"
	"sync/atomic"
)

func newStateMachine(protocol *protocolContext, types *primitive.TypeRegistry) dbsm.IStateMachine {
	context := &stateMachineContext{}
	return &stateMachine{
		stateMachineContext: context,
		protocol:            protocol,
		sm: statemachine.NewStateMachine(context, func(smCtx statemachine.SessionManagerContext) statemachine.SessionManager {
			return session.NewManager(smCtx, func(sessionCtx session.Context) session.PrimitiveManager {
				return primitive.NewManager(sessionCtx, types)
			})
		}),
	}
}

type stateMachineContext struct {
	index atomic.Uint64
}

func (c *stateMachineContext) Index() statemachine.Index {
	return statemachine.Index(c.index.Load())
}

func (c *stateMachineContext) update() statemachine.Index {
	return statemachine.Index(c.index.Add(1))
}

func (c *stateMachineContext) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarUint64(c.index.Load()); err != nil {
		return err
	}
	return nil
}

func (c *stateMachineContext) Recover(reader *snapshot.Reader) error {
	index, err := reader.ReadVarUint64()
	if err != nil {
		return err
	}
	c.index.Store(index)
	return nil
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
	s.sm.Propose(newProposal(
		statemachine.ProposalID(s.stateMachineContext.update()),
		proposal.Proposal,
		s.protocol.getStream(proposal.Term, proposal.SequenceNum)))
	return dbsm.Result{}, nil
}

func (s *stateMachine) Lookup(value interface{}) (interface{}, error) {
	query := value.(*protocolQuery)
	s.sm.Query(newQuery(statemachine.QueryID(s.queryID.Add(1)), query.input, query.stream))
	return nil, nil
}

func (s *stateMachine) SaveSnapshot(w io.Writer, collection dbsm.ISnapshotFileCollection, i <-chan struct{}) error {
	writer := snapshot.NewWriter(w)
	log.Infow("Persisting state to snapshot", logging.Uint64("Index", uint64(s.Index())))
	if err := s.stateMachineContext.Snapshot(writer); err != nil {
		return err
	}
	return s.sm.Snapshot(writer)
}

func (s *stateMachine) RecoverFromSnapshot(r io.Reader, files []dbsm.SnapshotFile, i <-chan struct{}) error {
	reader := snapshot.NewReader(r)
	if err := s.stateMachineContext.Recover(reader); err != nil {
		return err
	}
	log.Infow("Recovering state from snapshot", logging.Uint64("Index", uint64(s.Index())))
	return s.sm.Recover(reader)
}

func (s *stateMachine) Close() error {
	return nil
}

func newCall[T statemachine.CallID, I, O proto.Message](id T, input I, stream streams.WriteStream[O]) *stateMachineCall[T, I, O] {
	return &stateMachineCall[T, I, O]{
		id:     id,
		input:  input,
		stream: stream,
	}
}

type stateMachineCall[T statemachine.CallID, I, O proto.Message] struct {
	id     T
	input  I
	stream streams.WriteStream[O]
	state  atomic.Int32
}

func (c stateMachineCall[T, I, O]) ID() T {
	return c.id
}

func (c stateMachineCall[T, I, O]) Log() logging.Logger {
	return log
}

func (c stateMachineCall[T, I, O]) Input() I {
	return c.input
}

func (c stateMachineCall[T, I, O]) Output(output O) {
	c.stream.Value(output)
}

func (c stateMachineCall[T, I, O]) Error(err error) {
	c.stream.Error(err)
}

func (c stateMachineCall[T, I, O]) Close() {
	c.stream.Close()
}

func newProposal(id statemachine.ProposalID, input *multiraftv1.StateMachineProposalInput, stream streams.WriteStream[*multiraftv1.StateMachineProposalOutput]) *stateMachineProposal {
	return &stateMachineProposal{
		stateMachineCall: newCall[statemachine.ProposalID](id, input, stream),
	}
}

type stateMachineProposal struct {
	*stateMachineCall[statemachine.ProposalID, *multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput]
}

var _ statemachine.Proposal[*multiraftv1.StateMachineProposalInput, *multiraftv1.StateMachineProposalOutput] = (*stateMachineProposal)(nil)

func newQuery(id statemachine.QueryID, input *multiraftv1.StateMachineQueryInput, stream streams.WriteStream[*multiraftv1.StateMachineQueryOutput]) *stateMachineQuery {
	return &stateMachineQuery{
		stateMachineCall: newCall[statemachine.QueryID, *multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput](id, input, stream),
	}
}

type stateMachineQuery struct {
	*stateMachineCall[statemachine.QueryID, *multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput]
}

var _ statemachine.Query[*multiraftv1.StateMachineQueryInput, *multiraftv1.StateMachineQueryOutput] = (*stateMachineQuery)(nil)
