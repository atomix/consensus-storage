// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

type NewPrimitiveFunc[I, O any] func(Context[I, O]) Primitive[I, O]

type Primitive[I, O any] interface {
	snapshot.Recoverable
	Propose(proposal Proposal[I, O])
	Query(query Query[I, O])
}

type AnyPrimitive Primitive[any, any]

type Type[I, O any] interface {
	Service() string
	Codec() Codec[I, O]
	NewStateMachine(Context[I, O]) Primitive[I, O]
}

type AnyType Type[any, any]

func NewType[I, O any](service string, codec Codec[I, O], factory NewPrimitiveFunc[I, O]) Type[I, O] {
	return &primitiveType[I, O]{
		service: service,
		codec:   codec,
		factory: factory,
	}
}

type primitiveType[I, O any] struct {
	service string
	codec   Codec[I, O]
	factory func(Context[I, O]) Primitive[I, O]
}

func (t *primitiveType[I, O]) Service() string {
	return t.service
}

func (t *primitiveType[I, O]) Codec() Codec[I, O] {
	return t.codec
}

func (t *primitiveType[I, O]) NewStateMachine(context Context[I, O]) Primitive[I, O] {
	return t.factory(context)
}

type Info interface {
	// ID returns the service identifier
	ID() ID
	// Log returns the service logger
	Log() logging.Logger
	// Service returns the service name
	Service() string
	// Namespace returns the service namespace
	Namespace() string
	// Name returns the service name
	Name() string
}

type Context[I, O any] interface {
	statemachine.SessionManagerContext
	Info
	// Sessions returns the open sessions
	Sessions() session.Sessions
	// Proposals returns the pending proposals
	Proposals() Proposals[I, O]
}

type AnyContext Context[any, any]

type ID uint64

type managedPrimitive interface {
	snapshot.Recoverable
	Info
	open(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput])
	close(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput])
	propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput])
	query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput])
}

func newPrimitiveContext[I, O any](parent session.Context, id ID, namespace string, name string, primitiveType Type[I, O]) *primitiveContext[I, O] {
	return &primitiveContext[I, O]{
		Context:   parent,
		id:        id,
		namespace: namespace,
		name:      name,
		service:   primitiveType.Service(),
		sessions:  newPrimitiveSessions[I, O](),
		proposals: newPrimitiveProposals[I, O](),
		codec:     primitiveType.Codec(),
		log: log.WithFields(
			logging.String("Service", primitiveType.Service()),
			logging.Uint64("Primitive", uint64(id)),
			logging.String("Namespace", namespace),
			logging.String("Name", name)),
	}
}

type primitiveContext[I, O any] struct {
	session.Context
	id        ID
	namespace string
	name      string
	service   string
	codec     Codec[I, O]
	sessions  *primitiveSessions[I, O]
	proposals *primitiveProposals[I, O]
	log       logging.Logger
}

func (c *primitiveContext[I, O]) Log() logging.Logger {
	return c.log
}

func (c *primitiveContext[I, O]) ID() ID {
	return c.id
}

func (c *primitiveContext[I, O]) Service() string {
	return c.service
}

func (c *primitiveContext[I, O]) Namespace() string {
	return c.namespace
}

func (c *primitiveContext[I, O]) Name() string {
	return c.name
}

func (c *primitiveContext[I, O]) Sessions() Sessions {
	return c.sessions
}

func (c *primitiveContext[I, O]) Proposals() Proposals[I, O] {
	return c.proposals
}

func newPrimitive[I, O any](parent session.Context, id ID, namespace string, name string, primitiveType Type[I, O]) managedPrimitive {
	context := newPrimitiveContext[I, O](parent, id, namespace, name, primitiveType)
	return &primitiveExecutor[I, O]{
		primitiveContext: context,
		sm:               primitiveType.NewStateMachine(context),
	}
}

type primitiveExecutor[I, O any] struct {
	*primitiveContext[I, O]
	log logging.Logger
	sm  Primitive[I, O]
}

func (p *primitiveExecutor[I, O]) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(p.sessions.sessions)); err != nil {
		return err
	}
	for _, session := range p.sessions.list() {
		if err := session.Snapshot(writer); err != nil {
			return err
		}
	}
	return p.sm.Snapshot(writer)
}

func (p *primitiveExecutor[I, O]) Recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		session := newPrimitiveSession[I, O](p)
		if err := session.Recover(reader); err != nil {
			return err
		}
	}
	return p.sm.Recover(reader)
}

func (p *primitiveExecutor[I, O]) open(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {
	session := newPrimitiveSession[I, O](p)
	session.open(proposal.Session())
	proposal.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(p.ID()),
	})
	proposal.Close()
}

func (p *primitiveExecutor[I, O]) close(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {
	session, ok := p.sessions.get(proposal.Session().ID())
	if !ok {
		proposal.Error(errors.NewForbidden("session not found"))
		proposal.Close()
	} else {
		session.close()
		proposal.Output(&multiraftv1.ClosePrimitiveOutput{})
		proposal.Close()
	}
}

func (p *primitiveExecutor[I, O]) propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	session, ok := p.sessions.get(proposal.Session().ID())
	if !ok {
		proposal.Error(errors.NewForbidden("session not found"))
		proposal.Close()
	} else {
		session.propose(proposal)
	}
}

func (p *primitiveExecutor[I, O]) query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	session, ok := p.sessions.get(query.Session().ID())
	if !ok {
		query.Error(errors.NewForbidden("session not found"))
		query.Close()
	} else {
		session.query(query)
	}
}
