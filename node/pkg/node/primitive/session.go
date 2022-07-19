// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

type SessionID uint64

// Sessions provides access to open sessions
type Sessions interface {
	// Get gets a session by ID
	Get(SessionID) (Session, bool)
	// List lists all open sessions
	List() []Session
}

// Session is a service session
type Session interface {
	// ID returns the session identifier
	ID() SessionID
	// State returns the current session state
	State() SessionState
	// Watch watches the session state
	Watch(f func(SessionState)) Watcher
	// Commands returns the session commands
	Commands() Commands
}

type SessionState int

const (
	SessionClosed SessionState = iota
	SessionOpen
)

type OperationID string

// Operation is a command or query operation
type Operation interface {
	// OperationID returns the operation identifier
	OperationID() OperationID
	// Session returns the session executing the operation
	Session() Session
	// Input returns the operation input
	Input() []byte
	// Output returns the operation output
	Output([]byte, error)
	// Close closes the operation
	Close()
}

// Commands provides access to pending commands
type Commands interface {
	// Get gets a command by ID
	Get(CommandID) (Command, bool)
	// List lists all open commands
	List(OperationID) []Command
}

type CommandID uint64

// Command is a command operation
type Command interface {
	Operation
	// ID returns the command identifier
	ID() CommandID
	// State returns the current command state
	State() CommandState
	// Watch watches the command state
	Watch(f func(CommandState)) Watcher
}

type CommandState int

const (
	CommandPending CommandState = iota
	CommandRunning
	CommandComplete
)

// Query is a query operation
type Query interface {
	Operation
}

// Watcher is a context for a Watch call
type Watcher interface {
	// Cancel cancels the watcher
	Cancel()
}
