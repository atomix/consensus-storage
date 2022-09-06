// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/gogo/protobuf/proto"
)

type primitiveDelegate interface {
	snapshot.Recoverable
	propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput])
	query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput])
}

func newPrimitiveDelegate[I, O proto.Message](context *managedContext, primitiveType Type[I, O]) primitiveDelegate {
	primitive := &primitiveStateMachine[I, O]{
		managedContext: context,
		codec:          primitiveType.Codec(),
	}
	primitive.sm = primitiveType.NewStateMachine(primitive)
	return primitive
}

type primitiveStateMachine[I, O proto.Message] struct {
	*managedContext
	codec Codec[I, O]
	sm    Primitive[I, O]
}

func (p *primitiveStateMachine[I, O]) Sessions() Sessions[I, O] {
	return newPrimitiveSessions[I, O](p, p.Context.Sessions())
}

func (p *primitiveStateMachine[I, O]) Proposals() Proposals[I, O] {
	return newPrimitiveProposals[I, O](p, p.Context.Proposals())
}

func (p *primitiveStateMachine[I, O]) Snapshot(writer *snapshot.Writer) error {
	return p.sm.Snapshot(writer)
}

func (p *primitiveStateMachine[I, O]) Recover(reader *snapshot.Reader) error {
	return p.sm.Recover(reader)
}

func (p *primitiveStateMachine[I, O]) propose(parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	if proposal, ok := newPrimitiveProposal[I, O](p, parent); ok {
		p.sm.Propose(proposal)
	}
}

func (p *primitiveStateMachine[I, O]) query(parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	if query, ok := newPrimitiveQuery[I, O](p, parent); ok {
		p.sm.Query(query)
	}
}

var _ primitiveDelegate = (*primitiveStateMachine[any, any])(nil)

func newPrimitiveSessions[I, O proto.Message](primitive *primitiveStateMachine[I, O], parent session.Sessions) Sessions[I, O] {
	return &primitiveSessions[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type primitiveSessions[I, O proto.Message] struct {
	primitive *primitiveStateMachine[I, O]
	parent    session.Sessions
}

func (p *primitiveSessions[I, O]) Get(id SessionID) (Session[I, O], bool) {
	proposal, ok := p.parent.Get(session.ID(id))
	if !ok {
		return nil, false
	}
	return newPrimitiveSession[I, O](p.primitive, proposal), true
}

func (p *primitiveSessions[I, O]) List() []Session[I, O] {
	parents := p.primitive.Context.Sessions().List()
	proposals := make([]Session[I, O], 0, len(parents))
	for _, parent := range parents {
		proposals = append(proposals, newPrimitiveSession[I, O](p.primitive, parent))
	}
	return proposals
}

func newPrimitiveSession[I, O proto.Message](primitive *primitiveStateMachine[I, O], parent session.Session) Session[I, O] {
	return &primitiveSession[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type primitiveSession[I, O proto.Message] struct {
	primitive *primitiveStateMachine[I, O]
	parent    session.Session
}

func (s *primitiveSession[I, O]) Log() logging.Logger {
	return s.parent.Log()
}

func (s *primitiveSession[I, O]) ID() SessionID {
	return SessionID(s.parent.ID())
}

func (s *primitiveSession[I, O]) State() SessionState {
	return SessionState(s.parent.State())
}

func (s *primitiveSession[I, O]) Watch(watcher statemachine.WatchFunc[SessionState]) statemachine.CancelFunc {
	return s.parent.Watch(func(state session.State) {
		watcher(SessionState(state))
	})
}

func (s *primitiveSession[I, O]) Proposals() Proposals[I, O] {
	return newPrimitiveProposals[I, O](s.primitive, s.parent.Proposals())
}

func newPrimitiveProposals[I, O proto.Message](primitive *primitiveStateMachine[I, O], parent session.Proposals) Proposals[I, O] {
	return &primitiveProposals[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type primitiveProposals[I, O proto.Message] struct {
	primitive *primitiveStateMachine[I, O]
	parent    session.Proposals
}

func (p *primitiveProposals[I, O]) Get(id statemachine.ProposalID) (Proposal[I, O], bool) {
	proposal, ok := p.parent.Get(id)
	if !ok {
		return nil, false
	}
	return newPrimitiveProposal[I, O](p.primitive, proposal)
}

func (p *primitiveProposals[I, O]) List() []Proposal[I, O] {
	parents := p.parent.List()
	proposals := make([]Proposal[I, O], 0, len(parents))
	for _, parent := range parents {
		proposal, ok := newPrimitiveProposal[I, O](p.primitive, parent)
		if ok {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

func newPrimitiveProposal[I, O proto.Message](primitive *primitiveStateMachine[I, O], parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) (Proposal[I, O], bool) {
	input, err := primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		parent.Error(errors.NewInternal("failed decoding proposal input: %s", err))
		return nil, false
	}
	return &primitiveProposal[I, O]{
		primitive: primitive,
		parent:    parent,
		input:     input,
	}, true
}

type primitiveProposal[I, O proto.Message] struct {
	primitive *primitiveStateMachine[I, O]
	parent    session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
	input     I
}

func (e *primitiveProposal[I, O]) ID() statemachine.ProposalID {
	return e.parent.ID()
}

func (e *primitiveProposal[I, O]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *primitiveProposal[I, O]) Session() Session[I, O] {
	return newPrimitiveSession[I, O](e.primitive, e.parent.Session())
}

func (e *primitiveProposal[I, O]) Watch(watcher statemachine.WatchFunc[statemachine.ProposalPhase]) statemachine.CancelFunc {
	return e.parent.Watch(watcher)
}

func (e *primitiveProposal[I, O]) Input() I {
	return e.input
}

func (e *primitiveProposal[I, O]) Output(output O) {
	payload, err := e.primitive.codec.EncodeOutput(output)
	if err != nil {
		e.parent.Error(errors.NewInternal("failed encoding proposal output: %s", err))
	} else {
		e.parent.Output(&multiraftv1.PrimitiveProposalOutput{
			Payload: payload,
		})
	}
}

func (e *primitiveProposal[I, O]) Error(err error) {
	e.parent.Error(err)
}

func (e *primitiveProposal[I, O]) Cancel() {
	e.parent.Cancel()
}

func (e *primitiveProposal[I, O]) Close() {
	e.parent.Close()
}

func newPrimitiveQuery[I, O proto.Message](primitive *primitiveStateMachine[I, O], parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) (Query[I, O], bool) {
	input, err := primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		parent.Error(errors.NewInternal("failed decoding proposal input: %s", err))
		return nil, false
	}
	return &primitiveQuery[I, O]{
		primitive: primitive,
		parent:    parent,
		input:     input,
	}, true
}

type primitiveQuery[I, O proto.Message] struct {
	primitive *primitiveStateMachine[I, O]
	parent    session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]
	input     I
}

func (e *primitiveQuery[I, O]) ID() statemachine.QueryID {
	return e.parent.ID()
}

func (e *primitiveQuery[I, O]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *primitiveQuery[I, O]) Session() Session[I, O] {
	return newPrimitiveSession[I, O](e.primitive, e.parent.Session())
}

func (e *primitiveQuery[I, O]) Watch(watcher statemachine.WatchFunc[statemachine.QueryPhase]) statemachine.CancelFunc {
	return e.parent.Watch(watcher)
}

func (e *primitiveQuery[I, O]) Input() I {
	return e.input
}

func (e *primitiveQuery[I, O]) Output(output O) {
	payload, err := e.primitive.codec.EncodeOutput(output)
	if err != nil {
		e.parent.Error(errors.NewInternal("failed encoding proposal output: %s", err))
	} else {
		e.parent.Output(&multiraftv1.PrimitiveQueryOutput{
			Payload: payload,
		})
	}
}

func (e *primitiveQuery[I, O]) Error(err error) {
	e.parent.Error(err)
}

func (e *primitiveQuery[I, O]) Cancel() {
	e.parent.Cancel()
}

func (e *primitiveQuery[I, O]) Close() {
	e.parent.Close()
}
