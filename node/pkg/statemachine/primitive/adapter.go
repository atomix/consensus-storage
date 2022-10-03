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
	"time"
)

type primitiveDelegate interface {
	snapshot.Recoverable
	propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput])
	query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput])
}

func newPrimitiveAdapter[I, O any](context *managedPrimitive, primitiveType Type[I, O]) *primitiveAdapter[I, O] {
	primitive := &primitiveAdapter[I, O]{
		managedPrimitive: context,
		codec:            primitiveType.Codec(),
	}
	primitive.sm = primitiveType.NewStateMachine(primitive)
	return primitive
}

type primitiveAdapter[I, O any] struct {
	*managedPrimitive
	codec Codec[I, O]
	sm    Primitive[I, O]
}

func (p *primitiveAdapter[I, O]) Sessions() Sessions {
	return newSessionsAdapter[I, O](p, p.managedPrimitive.Sessions())
}

func (p *primitiveAdapter[I, O]) Proposals() Proposals[I, O] {
	return newProposalsAdapter[I, O](p, p.managedPrimitive.Proposals())
}

func (p *primitiveAdapter[I, O]) Snapshot(writer *snapshot.Writer) error {
	return p.sm.Snapshot(writer)
}

func (p *primitiveAdapter[I, O]) Recover(reader *snapshot.Reader) error {
	return p.sm.Recover(reader)
}

func (p *primitiveAdapter[I, O]) propose(parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	if proposal, ok := newProposalAdapter[I, O](p, parent); ok {
		p.sm.Propose(proposal)
	}
}

func (p *primitiveAdapter[I, O]) query(parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	if query, ok := newQueryAdapter[I, O](p, parent); ok {
		p.sm.Query(query)
	}
}

var _ primitiveDelegate = (*primitiveAdapter[any, any])(nil)

func newSessionsAdapter[I, O any](primitive *primitiveAdapter[I, O], parent session.Sessions) Sessions {
	return &sessionsAdapter[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type sessionsAdapter[I, O any] struct {
	primitive *primitiveAdapter[I, O]
	parent    session.Sessions
}

func (p *sessionsAdapter[I, O]) Get(id SessionID) (Session, bool) {
	session, ok := p.parent.Get(session.ID(id))
	if !ok {
		return nil, false
	}
	return newSessionAdapter[I, O](p.primitive, session), true
}

func (p *sessionsAdapter[I, O]) List() []Session {
	parents := p.parent.List()
	sessions := make([]Session, 0, len(parents))
	for _, parent := range parents {
		sessions = append(sessions, newSessionAdapter[I, O](p.primitive, parent))
	}
	return sessions
}

func newSessionAdapter[I, O any](primitive *primitiveAdapter[I, O], parent session.Session) Session {
	return &sessionAdapter[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type sessionAdapter[I, O any] struct {
	primitive *primitiveAdapter[I, O]
	parent    session.Session
}

func (s *sessionAdapter[I, O]) Log() logging.Logger {
	return s.parent.Log()
}

func (s *sessionAdapter[I, O]) ID() SessionID {
	return SessionID(s.parent.ID())
}

func (s *sessionAdapter[I, O]) State() SessionState {
	return SessionState(s.parent.State())
}

func (s *sessionAdapter[I, O]) Watch(watcher WatchFunc[SessionState]) CancelFunc {
	cancel := s.parent.Watch(func(state session.State) {
		watcher(SessionState(state))
	})
	return func() {
		cancel()
	}
}

func newProposalsAdapter[I, O any](primitive *primitiveAdapter[I, O], parent session.Proposals) Proposals[I, O] {
	return &proposalsAdapter[I, O]{
		primitive: primitive,
		parent:    parent,
	}
}

type proposalsAdapter[I, O any] struct {
	primitive *primitiveAdapter[I, O]
	parent    session.Proposals
}

func (p *proposalsAdapter[I, O]) Get(id statemachine.ProposalID) (Proposal[I, O], bool) {
	proposal, ok := p.parent.Get(id)
	if !ok {
		return nil, false
	}
	return newProposalAdapter[I, O](p.primitive, proposal)
}

func (p *proposalsAdapter[I, O]) List() []Proposal[I, O] {
	parents := p.parent.List()
	proposals := make([]Proposal[I, O], 0, len(parents))
	for _, parent := range parents {
		proposal, ok := newProposalAdapter[I, O](p.primitive, parent)
		if ok {
			proposals = append(proposals, proposal)
		}
	}
	return proposals
}

func newProposalAdapter[I, O any](primitive *primitiveAdapter[I, O], parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) (Proposal[I, O], bool) {
	input, err := primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		parent.Error(errors.NewInternal("failed decoding proposal input: %s", err))
		return nil, false
	}
	return &proposalAdapter[I, O]{
		primitive: primitive,
		parent:    parent,
		input:     input,
	}, true
}

type proposalAdapter[I, O any] struct {
	primitive *primitiveAdapter[I, O]
	parent    session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
	input     I
}

func (e *proposalAdapter[I, O]) ID() statemachine.ProposalID {
	return e.parent.ID()
}

func (e *proposalAdapter[I, O]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *proposalAdapter[I, O]) Time() time.Time {
	return e.parent.Time()
}

func (e *proposalAdapter[I, O]) Session() Session {
	return newSessionAdapter[I, O](e.primitive, e.parent.Session())
}

func (e *proposalAdapter[I, O]) Watch(watcher WatchFunc[ProposalPhase]) CancelFunc {
	cancel := e.parent.Watch(func(phase session.ProposalPhase) {
		watcher(ProposalPhase(phase))
	})
	return func() {
		cancel()
	}
}

func (e *proposalAdapter[I, O]) Input() I {
	return e.input
}

func (e *proposalAdapter[I, O]) Output(output O) {
	payload, err := e.primitive.codec.EncodeOutput(output)
	if err != nil {
		e.parent.Error(errors.NewInternal("failed encoding proposal output: %s", err))
	} else {
		e.parent.Output(&multiraftv1.PrimitiveProposalOutput{
			Payload: payload,
		})
	}
}

func (e *proposalAdapter[I, O]) Error(err error) {
	e.parent.Error(err)
}

func (e *proposalAdapter[I, O]) Close() {
	e.parent.Close()
}

func newQueryAdapter[I, O any](primitive *primitiveAdapter[I, O], parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) (Query[I, O], bool) {
	input, err := primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		parent.Error(errors.NewInternal("failed decoding proposal input: %s", err))
		return nil, false
	}
	return &queryAdapter[I, O]{
		primitive: primitive,
		parent:    parent,
		input:     input,
	}, true
}

type queryAdapter[I, O any] struct {
	primitive *primitiveAdapter[I, O]
	parent    session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]
	input     I
}

func (e *queryAdapter[I, O]) ID() statemachine.QueryID {
	return e.parent.ID()
}

func (e *queryAdapter[I, O]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *queryAdapter[I, O]) Time() time.Time {
	return e.parent.Time()
}

func (e *queryAdapter[I, O]) Session() Session {
	return newSessionAdapter[I, O](e.primitive, e.parent.Session())
}

func (e *queryAdapter[I, O]) Watch(watcher WatchFunc[QueryPhase]) CancelFunc {
	cancel := e.parent.Watch(func(phase session.QueryPhase) {
		watcher(QueryPhase(phase))
	})
	return func() {
		cancel()
	}
}

func (e *queryAdapter[I, O]) Input() I {
	return e.input
}

func (e *queryAdapter[I, O]) Output(output O) {
	payload, err := e.primitive.codec.EncodeOutput(output)
	if err != nil {
		e.parent.Error(errors.NewInternal("failed encoding proposal output: %s", err))
	} else {
		e.parent.Output(&multiraftv1.PrimitiveQueryOutput{
			Payload: payload,
		})
	}
}

func (e *queryAdapter[I, O]) Error(err error) {
	e.parent.Error(err)
}

func (e *queryAdapter[I, O]) Close() {
	e.parent.Close()
}
