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
	Name() string
	APIVersion() string
	Codec() Codec[I, O]
	NewPrimitive(PrimitiveContext[I, O]) Primitive[I, O]
}

func NewPrimitiveType[I, O any](name, apiVersion string, codec Codec[I, O], factory func(PrimitiveContext[I, O]) Primitive[I, O]) PrimitiveType[I, O] {
	return &primitiveType[I, O]{
		name:       name,
		apiVersion: apiVersion,
		codec:      codec,
		factory:    factory,
	}
}

type primitiveType[I, O any] struct {
	name       string
	apiVersion string
	codec      Codec[I, O]
	factory    func(PrimitiveContext[I, O]) Primitive[I, O]
}

func (t *primitiveType[I, O]) Name() string {
	return t.name
}

func (t *primitiveType[I, O]) APIVersion() string {
	return t.apiVersion
}

func (t *primitiveType[I, O]) Codec() Codec[I, O] {
	return t.codec
}

func (t *primitiveType[I, O]) NewPrimitive(context PrimitiveContext[I, O]) Primitive[I, O] {
	return t.factory(context)
}

type PrimitiveContext[I, O any] interface {
	// PrimitiveID returns the service identifier
	PrimitiveID() multiraftv1.PrimitiveID
	// Type returns the service type
	Type() PrimitiveType[I, O]
	// Namespace returns the service namespace
	Namespace() string
	// Name returns the service name
	Name() string
	// Index returns the current service index
	Index() multiraftv1.Index
	// Time returns the current service time
	Time() time.Time
	// Scheduler returns the service scheduler
	Scheduler() *Scheduler
	// Sessions returns the open sessions
	Sessions() Sessions[I, O]
	// Commands returns the pending commands
	Commands() Commands[I, O]
}

// Primitive is a primitive state machine
type Primitive[I, O any] interface {
	Snapshot(writer *snapshot.Writer) error
	Recover(reader *snapshot.Reader) error
	Update(command Command[I, O])
	Read(query Query[I, O])
}

// Session is a service session
type Session[I, O any] interface {
	// ID returns the session identifier
	ID() multiraftv1.SessionID
	// State returns the current session state
	State() multiraftv1.SessionSnapshot_State
	// Watch watches the session state
	Watch(f SessionWatcher) CancelFunc
	// Commands returns the session commands
	Commands() Commands[I, O]
}

type SessionWatcher func(multiraftv1.SessionSnapshot_State)

type CancelFunc func()

// Sessions provides access to open sessions
type Sessions[I, O any] interface {
	// Get gets a session by ID
	Get(multiraftv1.SessionID) (Session[I, O], bool)
	// List lists all open sessions
	List() []Session[I, O]
}

// Operation is a command or query operation
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

// Command is a command operation
type Command[I, O any] interface {
	Operation[I, O]
	// Index returns the command index
	Index() multiraftv1.Index
	// State returns the current command state
	State() multiraftv1.CommandSnapshot_State
	// Watch watches the command state
	Watch(f CommandWatcher) CancelFunc
	// Close closes the command
	Close()
}

type CommandWatcher func(state multiraftv1.CommandSnapshot_State)

// Commands provides access to pending commands
type Commands[I, O any] interface {
	// Get gets a command by ID
	Get(multiraftv1.Index) (Command[I, O], bool)
	// List lists all open commands
	List() []Command[I, O]
}

// Query is a query operation
type Query[I, O any] interface {
	Operation[I, O]
}
