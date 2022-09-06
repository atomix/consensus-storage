// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/gogo/protobuf/proto"
)

type NewPrimitiveFunc[I, O any] func(PrimitiveContext[I, O]) Primitive[I, O]

type Primitive[I, O any] interface {
	snapshot.Recoverable
	Propose(proposal Proposal[I, O])
	Query(query Query[I, O])
}

type Type[I, O any] interface {
	Service() string
	Codec() Codec[I, O]
	NewStateMachine(PrimitiveContext[I, O]) Primitive[I, O]
}

func NewType[I, O any](service string, codec Codec[I, O], factory NewPrimitiveFunc[I, O]) Type[I, O] {
	return &primitiveType[I, O]{
		service: service,
		codec:   codec,
		factory: factory,
	}
}

type primitiveType[I, O any] struct {
	service string
	codec   Codec[I, O]
	factory func(PrimitiveContext[I, O]) Primitive[I, O]
}

func (t *primitiveType[I, O]) Service() string {
	return t.service
}

func (t *primitiveType[I, O]) Codec() Codec[I, O] {
	return t.codec
}

func (t *primitiveType[I, O]) NewStateMachine(context PrimitiveContext[I, O]) Primitive[I, O] {
	return t.factory(context)
}

type PrimitiveContext[I, O any] interface {
	statemachine.SessionManagerContext
	// ID returns the service identifier
	ID() ID
	// Service returns the service name
	Service() string
	// Namespace returns the service namespace
	Namespace() string
	// Name returns the service name
	Name() string
	// Sessions returns the open sessions
	Sessions() Sessions[I, O]
	// Proposals returns the pending proposals
	Proposals() Proposals[I, O]
}

type ID uint64

type SessionState int

const (
	SessionOpen SessionState = iota
	SessionClosed
)

// Sessionized is an interface for types that are associated with a session
type Sessionized[I, O any] interface {
	Session() Session[I, O]
}

type SessionID uint64

// Session is a service session
type Session[I, O any] interface {
	statemachine.Watchable[SessionState]
	// Log returns the session log
	Log() logging.Logger
	// ID returns the session identifier
	ID() SessionID
	// State returns the current session state
	State() SessionState
	// Proposals returns the session proposals
	Proposals() Proposals[I, O]
}

// Sessions provides access to open sessions
type Sessions[I, O any] interface {
	// Get gets a session by ID
	Get(SessionID) (Session[I, O], bool)
	// List lists all open sessions
	List() []Session[I, O]
}

// Execution is a proposal or query execution
type Execution[T statemachine.ExecutionID, I, O any] interface {
	statemachine.Execution[T, I, O]
	Sessionized[I, O]
}

type ProposalID = statemachine.ProposalID

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Execution[statemachine.ProposalID, I, O]
}

// Proposals provides access to pending proposals
type Proposals[I, O any] interface {
	// Get gets a proposal by ID
	Get(statemachine.ProposalID) (Proposal[I, O], bool)
	// List lists all open proposals
	List() []Proposal[I, O]
}

type QueryID = statemachine.QueryID

// Query is a read operation
type Query[I, O any] interface {
	Execution[statemachine.QueryID, I, O]
}

type Executor[T Execution[U, I, O], U statemachine.ExecutionID, I, O any] interface {
	Execute(T)
}

type Proposer[I1, O1, I2, O2 any] interface {
	Executor[Proposal[I1, O1], statemachine.ProposalID, I1, O1]
	Proposals() Proposals[I2, O2]
}

func NewProposer[I1, O1, I2, O2 proto.Message](ctx PrimitiveContext[I1, O1]) *ProposerBuilder[I1, O1, I2, O2] {
	return &ProposerBuilder[I1, O1, I2, O2]{
		ctx: ctx,
	}
}

type ProposerBuilder[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (b *ProposerBuilder[I1, O1, I2, O2]) Name(name string) *ProposerBuilder[I1, O1, I2, O2] {
	b.name = name
	return b
}

func (b *ProposerBuilder[I1, O1, I2, O2]) Decoder(f func(I1) (I2, bool)) *ProposerBuilder[I1, O1, I2, O2] {
	b.decoder = f
	return b
}

func (b *ProposerBuilder[I1, O1, I2, O2]) Encoder(f func(O2) O1) *ProposerBuilder[I1, O1, I2, O2] {
	b.encoder = f
	return b
}

func (b *ProposerBuilder[I1, O1, I2, O2]) Build(f func(Proposal[I2, O2])) Proposer[I1, O1, I2, O2] {
	return &transcodingProposer[I1, O1, I2, O2]{
		ctx:     b.ctx,
		decoder: b.decoder,
		encoder: b.encoder,
		name:    b.name,
		f:       f,
	}
}

var _ ExecutorBuilder[
	Proposal[proto.Message, proto.Message],
	statemachine.ProposalID,
	proto.Message,
	proto.Message,
	Proposer[proto.Message, proto.Message, proto.Message, proto.Message]] = (*ProposerBuilder[proto.Message, proto.Message, proto.Message, proto.Message])(nil)

type ExecutorBuilder[
	T Execution[U, I, O],
	U statemachine.ExecutionID,
	I proto.Message,
	O proto.Message,
	E Executor[T, U, I, O]] interface {
	Build(f func(T)) E
}

type Querier[I1, O1, I2, O2 any] interface {
	Executor[Query[I1, O1], statemachine.QueryID, I1, O1]
}

func NewQuerier[I1, O1, I2, O2 proto.Message](ctx PrimitiveContext[I1, O1]) *QuerierBuilder[I1, O1, I2, O2] {
	return &QuerierBuilder[I1, O1, I2, O2]{
		ctx: ctx,
	}
}

type QuerierBuilder[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (b *QuerierBuilder[I1, O1, I2, O2]) Name(name string) *QuerierBuilder[I1, O1, I2, O2] {
	b.name = name
	return b
}

func (b *QuerierBuilder[I1, O1, I2, O2]) Decoder(f func(I1) (I2, bool)) *QuerierBuilder[I1, O1, I2, O2] {
	b.decoder = f
	return b
}

func (b *QuerierBuilder[I1, O1, I2, O2]) Encoder(f func(O2) O1) *QuerierBuilder[I1, O1, I2, O2] {
	b.encoder = f
	return b
}

func (b *QuerierBuilder[I1, O1, I2, O2]) Build(f func(Query[I2, O2])) Querier[I1, O1, I2, O2] {
	return &transcodingQuerier[I1, O1, I2, O2]{
		ctx:     b.ctx,
		decoder: b.decoder,
		encoder: b.encoder,
		name:    b.name,
		f:       f,
	}
}

var _ ExecutorBuilder[
	Query[proto.Message, proto.Message],
	statemachine.QueryID,
	proto.Message,
	proto.Message,
	Querier[proto.Message, proto.Message, proto.Message, proto.Message]] = (*QuerierBuilder[proto.Message, proto.Message, proto.Message, proto.Message])(nil)

type transcodingProposer[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	name    string
	f       func(Proposal[I2, O2])
}

func (p *transcodingProposer[I1, O1, I2, O2]) Proposals() Proposals[I2, O2] {
	return newTranscodingProposals[I1, O1, I2, O2](p.ctx.Proposals(), p.decoder, p.encoder)
}

func (p *transcodingProposer[I1, O1, I2, O2]) Execute(parent Proposal[I1, O1]) {
	input, ok := p.decoder(parent.Input())
	if !ok {
		return
	}
	proposal := newTranscodingProposal[I1, O1, I2, O2](parent, input, p.decoder, p.encoder, parent.Log().WithFields(logging.String("Method", p.name)))
	proposal.Log().Debugw("Applying proposal", logging.Stringer("Input", proposal.Input()))
	p.f(proposal)
}

type transcodingQuerier[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	name    string
	f       func(Query[I2, O2])
}

func (q *transcodingQuerier[I1, O1, I2, O2]) Execute(parent Query[I1, O1]) {
	input, ok := q.decoder(parent.Input())
	if !ok {
		return
	}
	query := newTranscodingQuery[I1, O1, I2, O2](parent, input, q.decoder, q.encoder, parent.Log().WithFields(logging.String("Method", q.name)))
	query.Log().Debugw("Applying query", logging.Stringer("Input", query.Input()))
	q.f(query)
}

func newTranscodingSession[I1, O1, I2, O2 any](parent Session[I1, O1], decoder func(I1) (I2, bool), encoder func(O2) O1) Session[I2, O2] {
	return &transcodingSession[I1, O1, I2, O2]{
		parent:  parent,
		decoder: decoder,
		encoder: encoder,
	}
}

type transcodingSession[I1, O1, I2, O2 any] struct {
	parent  Session[I1, O1]
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (s *transcodingSession[I1, O1, I2, O2]) Log() logging.Logger {
	return s.parent.Log()
}

func (s *transcodingSession[I1, O1, I2, O2]) ID() SessionID {
	return s.parent.ID()
}

func (s *transcodingSession[I1, O1, I2, O2]) State() SessionState {
	return s.parent.State()
}

func (s *transcodingSession[I1, O1, I2, O2]) Watch(watcher statemachine.WatchFunc[SessionState]) statemachine.CancelFunc {
	return s.parent.Watch(watcher)
}

func (s *transcodingSession[I1, O1, I2, O2]) Proposals() Proposals[I2, O2] {
	return newTranscodingProposals[I1, O1, I2, O2](s.parent.Proposals(), s.decoder, s.encoder)
}

func newTranscodingProposals[I1, O1, I2, O2 any](parent Proposals[I1, O1], decoder func(I1) (I2, bool), encoder func(O2) O1) Proposals[I2, O2] {
	return &transcodingProposals[I1, O1, I2, O2]{
		parent:  parent,
		decoder: decoder,
		encoder: encoder,
	}
}

type transcodingProposals[I1, O1, I2, O2 any] struct {
	parent  Proposals[I1, O1]
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (p *transcodingProposals[I1, O1, I2, O2]) Get(id statemachine.ProposalID) (Proposal[I2, O2], bool) {
	parent, ok := p.parent.Get(id)
	if !ok {
		return nil, false
	}
	if input, ok := p.decoder(parent.Input()); ok {
		return newTranscodingProposal[I1, O1, I2, O2](parent, input, p.decoder, p.encoder, parent.Log()), true
	}
	return nil, false
}

func (p *transcodingProposals[I1, O1, I2, O2]) List() []Proposal[I2, O2] {
	parents := p.parent.List()
	proposals := make([]Proposal[I2, O2], 0, len(parents))
	for _, parent := range parents {
		if input, ok := p.decoder(parent.Input()); ok {
			proposals = append(proposals, newTranscodingProposal[I1, O1, I2, O2](parent, input, p.decoder, p.encoder, parent.Log()))
		}
	}
	return proposals
}

func newTranscodingExecution[T statemachine.ExecutionID, I1, O1, I2, O2 any](
	parent Execution[T, I1, O1],
	input I2,
	decoder func(I1) (I2, bool),
	encoder func(O2) O1,
	log logging.Logger) Execution[T, I2, O2] {
	return &transcodingExecution[T, I1, O1, I2, O2]{
		parent:  parent,
		input:   input,
		decoder: decoder,
		encoder: encoder,
		log:     log,
	}
}

type transcodingExecution[T statemachine.ExecutionID, I1, O1, I2, O2 any] struct {
	parent  Execution[T, I1, O1]
	input   I2
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	log     logging.Logger
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) ID() T {
	return p.parent.ID()
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Log() logging.Logger {
	return p.log
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Session() Session[I2, O2] {
	return newTranscodingSession[I1, O1, I2, O2](p.parent.Session(), p.decoder, p.encoder)
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Watch(watcher statemachine.WatchFunc[statemachine.Phase]) statemachine.CancelFunc {
	return p.parent.Watch(watcher)
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Input() I2 {
	return p.input
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Output(output O2) {
	p.parent.Output(p.encoder(output))
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Error(err error) {
	p.parent.Error(err)
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Cancel() {
	p.parent.Cancel()
}

func (p *transcodingExecution[T, I1, O1, I2, O2]) Close() {
	p.parent.Close()
}

func newTranscodingProposal[I1, O1, I2, O2 any](
	parent Proposal[I1, O1],
	input I2,
	decoder func(I1) (I2, bool),
	encoder func(O2) O1,
	log logging.Logger) Proposal[I2, O2] {
	return &transcodingProposal[I2, O2]{
		Execution: newTranscodingExecution[statemachine.ProposalID, I1, O1, I2, O2](parent, input, decoder, encoder, log),
	}
}

type transcodingProposal[I, O any] struct {
	Execution[statemachine.ProposalID, I, O]
}

func newTranscodingQuery[I1, O1, I2, O2 any](
	parent Query[I1, O1],
	input I2,
	decoder func(I1) (I2, bool),
	encoder func(O2) O1,
	log logging.Logger) Query[I2, O2] {
	return &transcodingQuery[I2, O2]{
		Execution: newTranscodingExecution[statemachine.QueryID, I1, O1, I2, O2](parent, input, decoder, encoder, log),
	}
}

type transcodingQuery[I, O any] struct {
	Execution[statemachine.QueryID, I, O]
}
