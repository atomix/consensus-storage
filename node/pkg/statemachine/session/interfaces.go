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

type CancelFunc = statemachine.CancelFunc

// Session is a service session
type Session interface {
	// Log returns the session log
	Log() logging.Logger
	// ID returns the session identifier
	ID() ID
	// State returns the current session state
	State() State
	// Watch watches the session state for changes
	Watch(watcher func(State)) CancelFunc
}

// Sessions provides access to open sessions
type Sessions interface {
	// Get gets a session by ID
	Get(ID) (Session, bool)
	// List lists all open sessions
	List() []Session
}

type CallState int

const (
	Pending CallState = iota
	Running
	Complete
	Canceled
)

// Call is a proposal or query call
type Call[T statemachine.CallID, I, O any] interface {
	statemachine.Call[T, I, O]
	// Time returns the state machine time at the time of the call
	Time() time.Time
	// Session returns the call session
	Session() Session
	// State returns the call state
	State() CallState
	// Watch watches the call state for changes
	Watch(watcher func(CallState)) CancelFunc
	// Cancel cancels the call
	Cancel()
}

type ProposalID = statemachine.ProposalID

type ProposalState = CallState

// Proposal is a proposal operation
type Proposal[I, O any] Call[ProposalID, I, O]

// Proposals provides access to pending proposals
type Proposals interface {
	// Get gets a proposal by ID
	Get(id ProposalID) (PrimitiveProposal, bool)
	// List lists all open proposals
	List() []PrimitiveProposal
}

type QueryID = statemachine.QueryID

type QueryState = CallState

// Query is a read operation
type Query[I, O any] Call[QueryID, I, O]
