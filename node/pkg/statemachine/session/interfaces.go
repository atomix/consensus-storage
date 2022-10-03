// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"time"
)

type NewPrimitiveManagerFunc func(Context) PrimitiveManager

type PrimitiveManager interface {
	snapshot.Recoverable
	CreatePrimitive(proposal CreatePrimitiveProposal)
	ClosePrimitive(proposal ClosePrimitiveProposal)
	Propose(proposal PrimitiveProposal)
	Query(query PrimitiveQuery)
}

type CreatePrimitiveProposal Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]
type ClosePrimitiveProposal Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]
type PrimitiveProposal Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
type PrimitiveQuery Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]

type Context interface {
	statemachine.SessionManagerContext
	// Sessions returns the open sessions
	Sessions() Sessions
	// Proposals returns the pending proposals
	Proposals() Proposals
}

type ID uint64

type State int

const (
	Open State = iota
	Closed
)

// Sessionized is an interface for types that are associated with a session
type Sessionized interface {
	Session() Session
}

type WatchFunc[T any] func(T)

type CancelFunc func()

type Watchable[T any] interface {
	Watch(watcher WatchFunc[T]) CancelFunc
}

// Session is a service session
type Session interface {
	Watchable[State]
	// Log returns the session log
	Log() logging.Logger
	// ID returns the session identifier
	ID() ID
	// State returns the current session state
	State() State
}

// Sessions provides access to open sessions
type Sessions interface {
	// Get gets a session by ID
	Get(ID) (Session, bool)
	// List lists all open sessions
	List() []Session
}

type Phase int

const (
	Pending Phase = iota
	Running
	Complete
	Canceled
)

// Execution is a proposal or query execution
type Execution[T statemachine.ExecutionID, I, O any] interface {
	statemachine.Execution[T, I, O]
	// Time returns the state machine time at the time of execution
	Time() time.Time
	Sessionized
}

type ProposalPhase = Phase

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Execution[statemachine.ProposalID, I, O]
	Watchable[ProposalPhase]
}

// Proposals provides access to pending proposals
type Proposals interface {
	// Get gets a proposal by ID
	Get(statemachine.ProposalID) (PrimitiveProposal, bool)
	// List lists all open proposals
	List() []PrimitiveProposal
}

type QueryPhase = Phase

// Query is a read operation
type Query[I, O any] interface {
	Execution[statemachine.QueryID, I, O]
	Watchable[QueryPhase]
}
