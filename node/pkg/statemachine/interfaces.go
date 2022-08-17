// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/gogo/protobuf/proto"
	"time"
)

type PrimitiveType[I, O any] interface {
	Service() string
	Codec() Codec[I, O]
	NewPrimitive(PrimitiveContext[I, O]) Primitive[I, O]
}

func NewPrimitiveType[I, O any](service string, codec Codec[I, O], factory func(PrimitiveContext[I, O]) Primitive[I, O]) PrimitiveType[I, O] {
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

func (t *primitiveType[I, O]) NewPrimitive(context PrimitiveContext[I, O]) Primitive[I, O] {
	return t.factory(context)
}

type PrimitiveID uint64

type Index uint64

type PrimitiveContext[I, O any] interface {
	// Log returns the primitive logger
	Log() logging.Logger
	// PrimitiveID returns the service identifier
	PrimitiveID() PrimitiveID
	// Type returns the service type
	Type() PrimitiveType[I, O]
	// Namespace returns the service namespace
	Namespace() string
	// Name returns the service name
	Name() string
	// Index returns the current service index
	Index() Index
	// Time returns the current service time
	Time() time.Time
	// Scheduler returns the service scheduler
	Scheduler() *Scheduler
	// Sessions returns the open sessions
	Sessions() Sessions[I, O]
	// Proposals returns the pending proposals
	Proposals() Proposals[I, O]
}

// Primitive is a primitive state machine
type Primitive[I, O any] interface {
	Snapshot(writer *snapshot.Writer) error
	Recover(reader *snapshot.Reader) error
	Update(proposal Proposal[I, O])
	Read(query Query[I, O])
}

type SessionID uint64

type SessionState int

const (
	SessionOpen SessionState = iota
	SessionClosed
)

// Session is a service session
type Session[I, O any] interface {
	// Log returns the session log
	Log() logging.Logger
	// ID returns the session identifier
	ID() SessionID
	// State returns the current session state
	State() SessionState
	// Watch watches the session state
	Watch(f SessionWatcher) CancelFunc
	// Proposals returns the session proposals
	Proposals() Proposals[I, O]
}

type SessionWatcher func(SessionState)

type CancelFunc func()

// Sessions provides access to open sessions
type Sessions[I, O any] interface {
	// Get gets a session by ID
	Get(SessionID) (Session[I, O], bool)
	// List lists all open sessions
	List() []Session[I, O]
}

type OperationState int

const (
	Pending OperationState = iota
	Runnnig
	Complete
)

type OperationWatcher func(OperationState)

// Operation is a proposal or query operation
type Operation[I, O any] interface {
	// Log returns the operation log
	Log() logging.Logger
	// Session returns the session executing the operation
	Session() Session[I, O]
	// Input returns the operation input
	Input() I
	// Output returns the operation output
	Output(O)
	// Error returns a failure error
	Error(error)
	// Watch watches the operation state
	Watch(f OperationWatcher) CancelFunc
	// Close closes the proposal
	Close()
}

type ProposalID uint64

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Operation[I, O]
	// ID returns the proposal ID
	ID() ProposalID
}

func NewUpdater[I1, O1, I2, O2 proto.Message](ctx PrimitiveContext[I1, O1]) *UpdaterBuilder[I1, O1, I2, O2] {
	return &UpdaterBuilder[I1, O1, I2, O2]{
		ctx: ctx,
	}
}

type UpdaterBuilder[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (b *UpdaterBuilder[I1, O1, I2, O2]) Name(name string) *UpdaterBuilder[I1, O1, I2, O2] {
	b.name = name
	return b
}

func (b *UpdaterBuilder[I1, O1, I2, O2]) Decoder(f func(I1) (I2, bool)) *UpdaterBuilder[I1, O1, I2, O2] {
	b.decoder = f
	return b
}

func (b *UpdaterBuilder[I1, O1, I2, O2]) Encoder(f func(O2) O1) *UpdaterBuilder[I1, O1, I2, O2] {
	b.encoder = f
	return b
}

func (b *UpdaterBuilder[I1, O1, I2, O2]) Build(f func(proposal Proposal[I2, O2])) Updater[I1, O1, I2, O2] {
	return &transcodingUpdater[I1, O1, I2, O2]{
		ctx:     b.ctx,
		name:    b.name,
		decoder: b.decoder,
		encoder: b.encoder,
		f:       f,
	}
}

func NewReader[I1, O1, I2, O2 proto.Message](ctx PrimitiveContext[I1, O1]) *ReaderBuilder[I1, O1, I2, O2] {
	return &ReaderBuilder[I1, O1, I2, O2]{
		ctx: ctx,
	}
}

type ReaderBuilder[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (b *ReaderBuilder[I1, O1, I2, O2]) Name(name string) *ReaderBuilder[I1, O1, I2, O2] {
	b.name = name
	return b
}

func (b *ReaderBuilder[I1, O1, I2, O2]) Decoder(f func(I1) (I2, bool)) *ReaderBuilder[I1, O1, I2, O2] {
	b.decoder = f
	return b
}

func (b *ReaderBuilder[I1, O1, I2, O2]) Encoder(f func(O2) O1) *ReaderBuilder[I1, O1, I2, O2] {
	b.encoder = f
	return b
}

func (b *ReaderBuilder[I1, O1, I2, O2]) Build(f func(proposal Query[I2, O2])) Reader[I1, O1, I2, O2] {
	return &transcodingReader[I1, O1, I2, O2]{
		ctx:     b.ctx,
		name:    b.name,
		decoder: b.decoder,
		encoder: b.encoder,
		f:       f,
	}
}

type Updater[I1, O1, I2, O2 any] interface {
	Proposals() Proposals[I2, O2]
	Update(proposal Proposal[I1, O1])
}

type Reader[I1, O1, I2, O2 any] interface {
	Read(query Query[I1, O1])
}

type transcodingUpdater[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	f       func(Proposal[I2, O2])
}

func (e *transcodingUpdater[I1, O1, I2, O2]) Proposals() Proposals[I2, O2] {
	return newTranscodingProposals[I1, O1, I2, O2](e.ctx.Proposals(), e.name, e.decoder, e.encoder)
}

func (e *transcodingUpdater[I1, O1, I2, O2]) Update(parent Proposal[I1, O1]) {
	input, ok := e.decoder(parent.Input())
	if !ok {
		return
	}
	proposal := newTranscodingProposal[I1, O1, I2, O2](parent, e.name, input, e.decoder, e.encoder)
	proposal.Log().Debugw("Proposal input", logging.Stringer("Input", proposal.Input()))
	e.f(proposal)
}

type transcodingReader[I1, O1, I2, O2 proto.Message] struct {
	ctx     PrimitiveContext[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	f       func(Query[I2, O2])
}

func (e *transcodingReader[I1, O1, I2, O2]) Read(parent Query[I1, O1]) {
	input, ok := e.decoder(parent.Input())
	if !ok {
		return
	}
	query := newTranscodingQuery[I1, O1, I2, O2](parent, e.name, input, e.decoder, e.encoder)
	query.Log().Debugw("Query input", logging.Stringer("Input", query.Input()))
	e.f(query)
}

func newTranscodingSessions[I1, O1, I2, O2 proto.Message](parent Sessions[I1, O1], name string, decoder func(I1) (I2, bool), encoder func(O2) O1) Sessions[I2, O2] {
	return &transcodingSessions[I1, O1, I2, O2]{
		parent:  parent,
		name:    name,
		decoder: decoder,
		encoder: encoder,
	}
}

type transcodingSessions[I1, O1, I2, O2 proto.Message] struct {
	parent  Sessions[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (p *transcodingSessions[I1, O1, I2, O2]) Get(id SessionID) (Session[I2, O2], bool) {
	session, ok := p.parent.Get(id)
	if !ok {
		return nil, false
	}
	return newTranscodingSession[I1, O1, I2, O2](session, p.name, p.decoder, p.encoder), true
}

func (p *transcodingSessions[I1, O1, I2, O2]) List() []Session[I2, O2] {
	var sessions []Session[I2, O2]
	for _, session := range p.parent.List() {
		sessions = append(sessions, newTranscodingSession[I1, O1, I2, O2](session, p.name, p.decoder, p.encoder))
	}
	return sessions
}

func newTranscodingSession[I1, O1, I2, O2 proto.Message](parent Session[I1, O1], name string, decoder func(I1) (I2, bool), encoder func(O2) O1) Session[I2, O2] {
	return &transcodingSession[I1, O1, I2, O2]{
		parent:  parent,
		name:    name,
		decoder: decoder,
		encoder: encoder,
	}
}

type transcodingSession[I1, O1, I2, O2 proto.Message] struct {
	parent  Session[I1, O1]
	name    string
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

func (s *transcodingSession[I1, O1, I2, O2]) Watch(f SessionWatcher) CancelFunc {
	return s.parent.Watch(f)
}

func (s *transcodingSession[I1, O1, I2, O2]) Proposals() Proposals[I2, O2] {
	return newTranscodingProposals[I1, O1, I2, O2](s.parent.Proposals(), s.name, s.decoder, s.encoder)
}

func newTranscodingProposals[I1, O1, I2, O2 proto.Message](parent Proposals[I1, O1], name string, decoder func(I1) (I2, bool), encoder func(O2) O1) Proposals[I2, O2] {
	return &transcodingProposals[I1, O1, I2, O2]{
		parent:  parent,
		name:    name,
		decoder: decoder,
		encoder: encoder,
	}
}

type transcodingProposals[I1, O1, I2, O2 proto.Message] struct {
	parent  Proposals[I1, O1]
	name    string
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
}

func (p *transcodingProposals[I1, O1, I2, O2]) Get(id ProposalID) (Proposal[I2, O2], bool) {
	proposal, ok := p.parent.Get(id)
	if !ok {
		return nil, false
	}
	input, ok := p.decoder(proposal.Input())
	if !ok {
		return nil, false
	}
	return newTranscodingProposal[I1, O1, I2, O2](proposal, p.name, input, p.decoder, p.encoder), true
}

func (p *transcodingProposals[I1, O1, I2, O2]) List() []Proposal[I2, O2] {
	var proposals []Proposal[I2, O2]
	for _, proposal := range p.parent.List() {
		if input, ok := p.decoder(proposal.Input()); ok {
			proposals = append(proposals, newTranscodingProposal[I1, O1, I2, O2](proposal, p.name, input, p.decoder, p.encoder))
		}
	}
	return proposals
}

func newTranscodingProposal[I1, O1, I2, O2 proto.Message](parent Proposal[I1, O1], name string, input I2, decoder func(I1) (I2, bool), encoder func(O2) O1) Proposal[I2, O2] {
	return &transcodingProposal[I1, O1, I2, O2]{
		parent:  parent,
		name:    name,
		input:   input,
		decoder: decoder,
		encoder: encoder,
		log:     parent.Log().WithFields(logging.String("Operation", name)),
	}
}

type transcodingProposal[I1, O1, I2, O2 proto.Message] struct {
	parent  Proposal[I1, O1]
	name    string
	input   I2
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	log     logging.Logger
}

func (p *transcodingProposal[I1, O1, I2, O2]) ID() ProposalID {
	return p.parent.ID()
}

func (p *transcodingProposal[I1, O1, I2, O2]) Log() logging.Logger {
	return p.log
}

func (p *transcodingProposal[I1, O1, I2, O2]) Session() Session[I2, O2] {
	return newTranscodingSession[I1, O1, I2, O2](p.parent.Session(), p.name, p.decoder, p.encoder)
}

func (p *transcodingProposal[I1, O1, I2, O2]) Input() I2 {
	return p.input
}

func (p *transcodingProposal[I1, O1, I2, O2]) Output(output O2) {
	p.log.Debugw("Proposal output", logging.Stringer("Output", output))
	p.parent.Output(p.encoder(output))
}

func (p *transcodingProposal[I1, O1, I2, O2]) Error(err error) {
	p.log.Debugw("Proposal error", logging.Error("Error", err))
	p.parent.Error(err)
}

func (p *transcodingProposal[I1, O1, I2, O2]) Watch(f OperationWatcher) CancelFunc {
	return p.parent.Watch(f)
}

func (p *transcodingProposal[I1, O1, I2, O2]) Close() {
	p.log.Debugw("Proposal closed")
	p.parent.Close()
}

func newTranscodingQuery[I1, O1, I2, O2 proto.Message](parent Query[I1, O1], name string, input I2, decoder func(I1) (I2, bool), encoder func(O2) O1) Query[I2, O2] {
	return &transcodingQuery[I1, O1, I2, O2]{
		parent:  parent,
		name:    name,
		input:   input,
		decoder: decoder,
		encoder: encoder,
		log:     parent.Log().WithFields(logging.String("Operation", name)),
	}
}

type transcodingQuery[I1, O1, I2, O2 proto.Message] struct {
	parent  Query[I1, O1]
	name    string
	input   I2
	decoder func(I1) (I2, bool)
	encoder func(O2) O1
	log     logging.Logger
}

func (p *transcodingQuery[I1, O1, I2, O2]) ID() QueryID {
	return p.parent.ID()
}

func (p *transcodingQuery[I1, O1, I2, O2]) Log() logging.Logger {
	return p.log
}

func (p *transcodingQuery[I1, O1, I2, O2]) Session() Session[I2, O2] {
	return newTranscodingSession[I1, O1, I2, O2](p.parent.Session(), p.name, p.decoder, p.encoder)
}

func (p *transcodingQuery[I1, O1, I2, O2]) Input() I2 {
	return p.input
}

func (p *transcodingQuery[I1, O1, I2, O2]) Output(output O2) {
	p.log.Debugw("Query output", logging.Stringer("Output", output))
	p.parent.Output(p.encoder(output))
}

func (p *transcodingQuery[I1, O1, I2, O2]) Error(err error) {
	p.log.Debugw("Query error", logging.Error("Error", err))
	p.parent.Error(err)
}

func (p *transcodingQuery[I1, O1, I2, O2]) Watch(f OperationWatcher) CancelFunc {
	return p.parent.Watch(f)
}

func (p *transcodingQuery[I1, O1, I2, O2]) Close() {
	p.log.Debugw("Query closed")
	p.parent.Close()
}

// Proposals provides access to pending proposals
type Proposals[I, O any] interface {
	// Get gets a proposal by ID
	Get(ProposalID) (Proposal[I, O], bool)
	// List lists all open proposals
	List() []Proposal[I, O]
}

type QueryID uint64

// Query is a read operation
type Query[I, O any] interface {
	Operation[I, O]
	// ID returns the query ID
	ID() QueryID
}
