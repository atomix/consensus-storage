// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	statemachine "github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

type NewPrimitiveManagerFunc func(Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) PrimitiveManager

type PrimitiveManager interface {
	snapshot.Recoverable
	CreatePrimitive(proposal Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput])
	ClosePrimitive(proposal Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput])
	Propose(proposal Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput])
	Query(query Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput])
}

type Context[I, O any] interface {
	statemachine.Context
	// Sessions returns the open sessions
	Sessions() Sessions[I, O]
	// Proposals returns the pending proposals
	Proposals() Proposals[I, O]
}

type ID uint64

type State int

const (
	Open State = iota
	Closed
)

// Sessionized is an interface for types that are associated with a session
type Sessionized[I, O any] interface {
	Session() Session[I, O]
}

// Session is a service session
type Session[I, O any] interface {
	statemachine.Watchable[State]
	// Log returns the session log
	Log() logging.Logger
	// ID returns the session identifier
	ID() ID
	// State returns the current session state
	State() State
	// Proposals returns the session proposals
	Proposals() Proposals[I, O]
}

// Sessions provides access to open sessions
type Sessions[I, O any] interface {
	// Get gets a session by ID
	Get(ID) (Session[I, O], bool)
	// List lists all open sessions
	List() []Session[I, O]
}

// Execution is a proposal or query execution
type Execution[T statemachine.ExecutionID, I, O any] interface {
	statemachine.Execution[T, I, O]
	Sessionized[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
}

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

// Query is a read operation
type Query[I, O any] interface {
	Execution[statemachine.QueryID, I, O]
}

func NewTranscodingExecution[T statemachine.ExecutionID, I1, O1, I2, O2 any](
	execution statemachine.Execution[T, I1, O1],
	session Session[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput],
	input I2,
	transcoder func(O2) O1) Execution[T, I2, O2] {
	return &transcodingExecution[T, I1, O1, I2, O2]{
		Execution: statemachine.NewTranscodingExecution[T, I1, O1, I2, O2](execution, input, transcoder),
		session:   session,
	}
}

type transcodingExecution[T statemachine.ExecutionID, I1, O1, I2, O2 any] struct {
	statemachine.Execution[T, I2, O2]
	session Session[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Session() Session[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] {
	return e.session
}

func NewTranscodingProposal[I1, O1, I2, O2 any](
	proposal statemachine.Proposal[I1, O1],
	session Session[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput],
	input I2,
	transcoder func(O2) O1) Proposal[I2, O2] {
	return &transcodingProposal[I1, O1, I2, O2]{
		Execution: NewTranscodingExecution[statemachine.ProposalID, I1, O1, I2, O2](proposal, session, input, transcoder),
	}
}

type transcodingProposal[I1, O1, I2, O2 any] struct {
	Execution[statemachine.ProposalID, I2, O2]
}

func NewTranscodingQuery[I1, O1, I2, O2 any](
	proposal statemachine.Query[I1, O1],
	session Session[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput],
	input I2,
	transcoder func(O2) O1) Query[I2, O2] {
	return &transcodingQuery[I1, O1, I2, O2]{
		Execution: NewTranscodingExecution[statemachine.QueryID, I1, O1, I2, O2](proposal, session, input, transcoder),
	}
}

type transcodingQuery[I1, O1, I2, O2 any] struct {
	Execution[statemachine.QueryID, I2, O2]
}
