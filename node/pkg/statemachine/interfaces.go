// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"time"
)

type NewSessionManagerFunc func(SessionManagerContext) SessionManager

type SessionManager interface {
	snapshot.Recoverable
	OpenSession(proposal OpenSessionProposal)
	KeepAlive(proposal KeepAliveProposal)
	CloseSession(proposal CloseSessionProposal)
	Propose(proposal SessionProposal)
	Query(query SessionQuery)
}

type OpenSessionProposal Proposal[*multiraftv1.OpenSessionInput, *multiraftv1.OpenSessionOutput]
type KeepAliveProposal Proposal[*multiraftv1.KeepAliveInput, *multiraftv1.KeepAliveOutput]
type CloseSessionProposal Proposal[*multiraftv1.CloseSessionInput, *multiraftv1.CloseSessionOutput]
type SessionProposal Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]
type SessionQuery Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]

type Index uint64

type SessionManagerContext interface {
	Context
	// Log returns the primitive logger
	Log() logging.Logger
	// Time returns the current service time
	Time() time.Time
	// Scheduler returns the service scheduler
	Scheduler() Scheduler
}

type CancelFunc func()

type Scheduler interface {
	Time() time.Time
	Await(index Index, f func()) CancelFunc
	Delay(d time.Duration, f func()) CancelFunc
	Schedule(t time.Time, f func()) CancelFunc
}

type CallID interface {
	ProposalID | QueryID
}

// Call is a proposal or query call context
type Call[T CallID, I, O any] interface {
	// ID returns the execution identifier
	ID() T
	// Log returns the operation log
	Log() logging.Logger
	// Input returns the input
	Input() I
	// Output returns the output
	Output(O)
	// Error returns a failure error
	Error(error)
	// Close closes the execution
	Close()
}

type ProposalID uint64

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Call[ProposalID, I, O]
}

type QueryID uint64

// Query is a read operation
type Query[I, O any] interface {
	Call[QueryID, I, O]
}

func NewTranscodingExecution[T CallID, I1, O1, I2, O2 any](execution Call[T, I1, O1], input I2, transcoder func(O2) O1) Call[T, I2, O2] {
	return &transcodingExecution[T, I1, O1, I2, O2]{
		parent:     execution,
		input:      input,
		transcoder: transcoder,
	}
}

type transcodingExecution[T CallID, I1, O1, I2, O2 any] struct {
	parent     Call[T, I1, O1]
	input      I2
	transcoder func(O2) O1
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) ID() T {
	return e.parent.ID()
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Input() I2 {
	return e.input
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Output(output O2) {
	e.parent.Output(e.transcoder(output))
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Error(err error) {
	e.parent.Error(err)
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Close() {
	e.parent.Close()
}

func NewTranscodingProposal[I1, O1, I2, O2 any](proposal Proposal[I1, O1], input I2, transcoder func(O2) O1) Proposal[I2, O2] {
	return &transcodingProposal[I1, O1, I2, O2]{
		Call: NewTranscodingExecution[ProposalID, I1, O1, I2, O2](proposal, input, transcoder),
	}
}

type transcodingProposal[I1, O1, I2, O2 any] struct {
	Call[ProposalID, I2, O2]
}

func NewTranscodingQuery[I1, O1, I2, O2 any](proposal Query[I1, O1], input I2, transcoder func(O2) O1) Query[I2, O2] {
	return &transcodingQuery[I1, O1, I2, O2]{
		Call: NewTranscodingExecution[QueryID, I1, O1, I2, O2](proposal, input, transcoder),
	}
}

type transcodingQuery[I1, O1, I2, O2 any] struct {
	Call[QueryID, I2, O2]
}
