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
)

type primitiveDelegate interface {
	snapshot.Recoverable
	propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput])
	query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput])
}

func newGenericPrimitive[I, O any](context *primitiveContext, primitiveType Type[I, O]) *genericPrimitive[I, O] {
	primitive := &genericPrimitive[I, O]{
		primitiveContext: context,
		codec:            primitiveType.Codec(),
	}
	primitive.sm = primitiveType.NewStateMachine(primitive)
	return primitive
}

type genericPrimitive[I, O any] struct {
	*primitiveContext
	codec Codec[I, O]
	sm    Primitive[I, O]
}

func (p *genericPrimitive[I, O]) Sessions() Sessions[I, O] {
	return newGenericSessions[I, O](p, p.primitiveContext.Sessions())
}

func (p *genericPrimitive[I, O]) Proposals() Proposals[I, O] {
	return newGenericProposals[I, O](p, p.primitiveContext.Proposals())
}

func (p *genericPrimitive[I, O]) Snapshot(writer *snapshot.Writer) error {
	return p.sm.Snapshot(writer)
}

func (p *genericPrimitive[I, O]) Recover(reader *snapshot.Reader) error {
	return p.sm.Recover(reader)
}

func (p *genericPrimitive[I, O]) propose(parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	if proposal, ok := newGenericProposal[I, O](p, parent); ok {
		p.sm.Propose(proposal)
	}
}

func (p *genericPrimitive[I, O]) query(parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	if query, ok := newGenericQuery[I, O](p, parent); ok {
		p.sm.Query(query)
	}
}

var _ primitiveDelegate = (*genericPrimitive[any, any])(nil)

func newGenericSessions[I, O any](primitive *genericPrimitive[I, O], parent session.Sessions) Sessions[I, O] {
	return &genericSessions[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type genericSessions[I, O any] struct {
	primitive *genericPrimitive[I, O]
	parent    session.Sessions
}

func (p *genericSessions[I, O]) Get(id SessionID) (Session[I, O], bool) {
	session, ok := p.parent.Get(session.ID(id))
	if !ok {
		return nil, false
	}
	return newGenericSession[I, O](p.primitive, session), true
}

func (p *genericSessions[I, O]) List() []Session[I, O] {
	parents := p.parent.List()
	sessions := make([]Session[I, O], 0, len(parents))
	for _, parent := range parents {
		sessions = append(sessions, newGenericSession[I, O](p.primitive, parent))
	}
	return sessions
}

func newGenericSession[I, O any](primitive *genericPrimitive[I, O], parent session.Session) Session[I, O] {
	return &genericSession[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type genericSession[I, O any] struct {
	primitive *genericPrimitive[I, O]
	parent    session.Session
}

func (s *genericSession[I, O]) Log() logging.Logger {
	return s.parent.Log()
}

func (s *genericSession[I, O]) ID() SessionID {
	return SessionID(s.parent.ID())
}

func (s *genericSession[I, O]) State() SessionState {
	return SessionState(s.parent.State())
}

func (s *genericSession[I, O]) Watch(watcher statemachine.WatchFunc[SessionState]) statemachine.CancelFunc {
	return s.parent.Watch(func(state session.State) {
		watcher(SessionState(state))
	})
}

func (s *genericSession[I, O]) Proposals() Proposals[I, O] {
	return newGenericProposals[I, O](s.primitive, s.parent.Proposals())
}

func newGenericProposals[I, O any](primitive *genericPrimitive[I, O], parent session.Proposals) Proposals[I, O] {
	return &genericProposals[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type genericProposals[I, O any] struct {
	primitive *genericPrimitive[I, O]
	parent    session.Proposals
}

func (p *genericProposals[I, O]) Get(id statemachine.ProposalID) (Proposal[I, O], bool) {
	proposal, ok := p.parent.Get(id)
	if !ok {
		return nil, false
	}
	return newGenericProposal[I, O](p.primitive, proposal)
}

func (p *genericProposals[I, O]) List() []Proposal[I, O] {
	parents := p.parent.List()
	proposals := make([]Proposal[I, O], 0, len(parents))
	for _, parent := range parents {
		proposal, ok := newGenericProposal[I, O](p.primitive, parent)
		if ok {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

func newGenericProposal[I, O any](primitive *genericPrimitive[I, O], parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) (Proposal[I, O], bool) {
	input, err := primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		parent.Error(errors.NewInternal("failed decoding proposal input: %s", err))
		return nil, false
	}
	return &genericProposal[I, O]{
		primitive: primitive,
		parent:    parent,
		input:     input,
	}, true
}

type genericProposal[I, O any] struct {
	primitive *genericPrimitive[I, O]
	parent    session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
	input     I
}

func (e *genericProposal[I, O]) ID() statemachine.ProposalID {
	return e.parent.ID()
}

func (e *genericProposal[I, O]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *genericProposal[I, O]) Session() Session[I, O] {
	return newGenericSession[I, O](e.primitive, e.parent.Session())
}

func (e *genericProposal[I, O]) Watch(watcher statemachine.WatchFunc[statemachine.Phase]) statemachine.CancelFunc {
	return e.parent.Watch(watcher)
}

func (e *genericProposal[I, O]) Input() I {
	return e.input
}

func (e *genericProposal[I, O]) Output(output O) {
	payload, err := e.primitive.codec.EncodeOutput(output)
	if err != nil {
		e.parent.Error(errors.NewInternal("failed encoding proposal output: %s", err))
	} else {
		e.parent.Output(&multiraftv1.PrimitiveProposalOutput{
			Payload: payload,
		})
	}
}

func (e *genericProposal[I, O]) Error(err error) {
	e.parent.Error(err)
}

func (e *genericProposal[I, O]) Cancel() {
	e.parent.Cancel()
}

func (e *genericProposal[I, O]) Close() {
	e.parent.Close()
}

func newGenericQuery[I, O any](primitive *genericPrimitive[I, O], parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) (Query[I, O], bool) {
	input, err := primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		parent.Error(errors.NewInternal("failed decoding proposal input: %s", err))
		return nil, false
	}
	return &genericQuery[I, O]{
		primitive: primitive,
		parent:    parent,
		input:     input,
	}, true
}

type genericQuery[I, O any] struct {
	primitive *genericPrimitive[I, O]
	parent    session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]
	input     I
}

func (e *genericQuery[I, O]) ID() statemachine.QueryID {
	return e.parent.ID()
}

func (e *genericQuery[I, O]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *genericQuery[I, O]) Session() Session[I, O] {
	return newGenericSession[I, O](e.primitive, e.parent.Session())
}

func (e *genericQuery[I, O]) Watch(watcher statemachine.WatchFunc[statemachine.Phase]) statemachine.CancelFunc {
	return e.parent.Watch(watcher)
}

func (e *genericQuery[I, O]) Input() I {
	return e.input
}

func (e *genericQuery[I, O]) Output(output O) {
	payload, err := e.primitive.codec.EncodeOutput(output)
	if err != nil {
		e.parent.Error(errors.NewInternal("failed encoding proposal output: %s", err))
	} else {
		e.parent.Output(&multiraftv1.PrimitiveQueryOutput{
			Payload: payload,
		})
	}
}

func (e *genericQuery[I, O]) Error(err error) {
	e.parent.Error(err)
}

func (e *genericQuery[I, O]) Cancel() {
	e.parent.Cancel()
}

func (e *genericQuery[I, O]) Close() {
	e.parent.Close()
}
