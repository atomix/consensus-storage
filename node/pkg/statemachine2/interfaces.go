// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine2/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"time"
)

type NewSessionManagerFunc func(Context) SessionManager

type SessionManager interface {
	snapshot.Recoverable
	OpenSession(proposal Proposal[*multiraftv1.OpenSessionInput, *multiraftv1.OpenSessionOutput])
	KeepAlive(proposal Proposal[*multiraftv1.KeepAliveInput, *multiraftv1.KeepAliveOutput])
	CloseSession(proposal Proposal[*multiraftv1.CloseSessionInput, *multiraftv1.CloseSessionOutput])
	Propose(proposal Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput])
	Query(query Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput])
}

type Index uint64

type Context interface {
	// Log returns the primitive logger
	Log() logging.Logger
	// Index returns the current service index
	Index() Index
	// Time returns the current service time
	Time() time.Time
	// Scheduler returns the service scheduler
	Scheduler() Scheduler
}

// Timer is a cancellable timer
type Timer interface {
	// Cancel cancels the timer, preventing it from running in the future
	Cancel()
}

type Scheduler interface {
	Time() time.Time
	Await(index Index, f func()) Timer
	Delay(d time.Duration, f func()) Timer
	Schedule(t time.Time, f func()) Timer
}

type WatchFunc[T any] func(T)

type CancelFunc func()

type Watchable[T any] interface {
	Watch(watcher WatchFunc[T]) CancelFunc
}

type ExecutionID interface {
	ProposalID | QueryID
}

type Phase int

const (
	Pending Phase = iota
	Runnnig
	Complete
	Canceled
)

// Execution is a proposal or query execution context
type Execution[T ExecutionID, I, O any] interface {
	Watchable[Phase]
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
	// Cancel cancels the execution
	Cancel()
}

type ProposalID uint64

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Execution[ProposalID, I, O]
}

type QueryID uint64

// Query is a read operation
type Query[I, O any] interface {
	Execution[QueryID, I, O]
}

func NewTranscodingExecution[T ExecutionID, I1, O1, I2, O2 any](execution Execution[T, I1, O1], input I2, transcoder func(O2) O1) Execution[T, I2, O2] {
	return &transcodingExecution[T, I1, O1, I2, O2]{
		parent:     execution,
		input:      input,
		transcoder: transcoder,
	}
}

type transcodingExecution[T ExecutionID, I1, O1, I2, O2 any] struct {
	parent     Execution[T, I1, O1]
	input      I2
	transcoder func(O2) O1
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) ID() T {
	return e.parent.ID()
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *transcodingExecution[T, I1, O1, I2, O2]) Watch(watcher WatchFunc[Phase]) CancelFunc {
	return e.parent.Watch(watcher)
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

func (e *transcodingExecution[T, I1, O1, I2, O2]) Cancel() {
	e.parent.Cancel()
}

func NewTranscodingProposal[I1, O1, I2, O2 any](proposal Proposal[I1, O1], input I2, transcoder func(O2) O1) Proposal[I2, O2] {
	return &transcodingProposal[I1, O1, I2, O2]{
		Execution: NewTranscodingExecution[ProposalID, I1, O1, I2, O2](proposal, input, transcoder),
	}
}

type transcodingProposal[I1, O1, I2, O2 any] struct {
	Execution[ProposalID, I2, O2]
}

func NewTranscodingQuery[I1, O1, I2, O2 any](proposal Query[I1, O1], input I2, transcoder func(O2) O1) Query[I2, O2] {
	return &transcodingQuery[I1, O1, I2, O2]{
		Execution: NewTranscodingExecution[QueryID, I1, O1, I2, O2](proposal, input, transcoder),
	}
}

type transcodingQuery[I1, O1, I2, O2 any] struct {
	Execution[QueryID, I2, O2]
}
