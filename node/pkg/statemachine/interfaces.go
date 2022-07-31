// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
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
	Get(multiraftv1.SessionID) (Session[I, O], bool)
	// List lists all open sessions
	List() []Session[I, O]
}

// Operation is a proposal or read operation
type Operation[I, O any] interface {
	// Session returns the session executing the operation
	Session() Session[I, O]
	// Input returns the operation input
	Input() I
	// Output returns the operation output
	Output(O)
	// Error returns a failure error
	Error(error)
}

type ProposalID uint64

type ProposalState int

const (
	ProposalPending ProposalState = iota
	ProposalRunning
	ProposalComplete
)

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Operation[I, O]
	// ID returns the proposal ID
	ID() ProposalID
	// State returns the current proposal state
	State() ProposalState
	// Watch watches the proposal state
	Watch(f ProposalWatcher) CancelFunc
	// Close closes the proposal
	Close()
}

type ProposalWatcher func(ProposalState)

// Proposals provides access to pending proposals
type Proposals[I, O any] interface {
	// Get gets a proposal by ID
	Get(ProposalID) (Proposal[I, O], bool)
	// List lists all open proposals
	List() []Proposal[I, O]
}

// Query is a read operation
type Query[I, O any] interface {
	Operation[I, O]
}
