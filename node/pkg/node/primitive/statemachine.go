// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	"github.com/atomix/multi-raft/node/pkg/node/snapshot"
	"time"
)

// Scheduler provides deterministic scheduling for a state machine
type Scheduler interface {
	// Time returns the current time
	Time() time.Time

	// Run executes a function asynchronously
	Run(f func())

	// RunAfter schedules a function to be run once after the given delay
	RunAfter(d time.Duration, f func()) Timer

	// RepeatAfter schedules a function to run repeatedly every interval starting after the given delay
	RepeatAfter(d time.Duration, i time.Duration, f func()) Timer

	// RunAt schedules a function to be run once after the given delay
	RunAt(t time.Time, f func()) Timer

	// RepeatAt schedules a function to run repeatedly every interval starting after the given delay
	RepeatAt(t time.Time, i time.Duration, f func()) Timer
}

// Timer is a cancellable timer
type Timer interface {
	// Cancel cancels the timer, preventing it from running in the future
	Cancel()
}

type ID uint64

type Index uint64

// Context is a primitive context
type Context interface {
	// ID returns the service identifier
	ID() ID
	// Type returns the service type
	Type() Type
	// Namespace returns the service namespace
	Namespace() string
	// Name returns the service name
	Name() string
	// Index returns the current service index
	Index() Index
	// Time returns the current service time
	Time() time.Time
	// Scheduler returns the service scheduler
	Scheduler() Scheduler
	// Sessions returns the open sessions
	Sessions() Sessions
	// Commands returns the pending commands
	Commands() Commands
}

// StateMachine is a primitive state machine
type StateMachine interface {
	// Snapshot is called to take a snapshot of the service state
	Snapshot(writer *snapshot.Writer) error
	// Recover is called to restore the service state from a snapshot
	Recover(reader *snapshot.Reader) error
	// Command executes a service command
	Command(Command)
	// Query executes a service query
	Query(Query)
}
