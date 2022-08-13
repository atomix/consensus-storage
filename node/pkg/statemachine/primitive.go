// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/google/uuid"
	"time"
)

func newPrimitiveManager(registry *PrimitiveTypeRegistry, context *sessionManager) *primitiveManager {
	return &primitiveManager{
		registry:   registry,
		context:    context,
		primitives: make(map[multiraftv1.PrimitiveID]primitiveExecutor),
	}
}

type primitiveManager struct {
	registry   *PrimitiveTypeRegistry
	context    *sessionManager
	primitives map[multiraftv1.PrimitiveID]primitiveExecutor
}

func (m *primitiveManager) snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.primitives)); err != nil {
		return err
	}
	for _, primitive := range m.primitives {
		snapshot := &multiraftv1.PrimitiveSnapshot{
			PrimitiveID: multiraftv1.PrimitiveID(primitive.info().id),
			Spec:        primitive.info().spec,
		}
		if err := writer.WriteMessage(snapshot); err != nil {
			return err
		}
		if err := primitive.snapshot(writer); err != nil {
			return err
		}
	}
	return nil
}

func (m *primitiveManager) recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		snapshot := &multiraftv1.PrimitiveSnapshot{}
		if err := reader.ReadMessage(snapshot); err != nil {
			return err
		}
		factory, ok := m.registry.lookup(snapshot.Spec.Service)
		if !ok {
			return errors.NewFault("primitive type not found")
		}

		info := primitiveInfo{
			id:   snapshot.PrimitiveID,
			spec: snapshot.Spec,
		}
		primitive := factory(m, info)
		m.primitives[info.id] = primitive
		if err := primitive.recover(reader); err != nil {
			return nil
		}
	}
	return nil
}

func (m *primitiveManager) create(command *raftSessionCreatePrimitiveCommand) {
	var executor primitiveExecutor
	for _, p := range m.primitives {
		if p.info().spec.Namespace == command.Input().Namespace && p.info().spec.Name == command.Input().Name {
			if p.info().spec.Service != command.Input().Service {
				command.Error(errors.NewForbidden("cannot create primitive of a different type with the same name"))
				command.Close()
				return
			}
			executor = p
			break
		}
	}

	if executor == nil {
		factory, ok := m.registry.lookup(command.Input().Service)
		if !ok {
			command.Error(errors.NewForbidden("unknown primitive type"))
			command.Close()
			return
		}

		info := primitiveInfo{
			id:   multiraftv1.PrimitiveID(command.index),
			spec: command.Input().PrimitiveSpec,
		}
		executor = factory(m, info)
		m.primitives[info.id] = executor
	}
	executor.create(command)
}

func (m *primitiveManager) update(proposal *raftSessionOperationCommand) {
	primitive, ok := m.primitives[proposal.Input().PrimitiveID]
	if !ok {
		proposal.Error(errors.NewFault("primitive not found"))
		proposal.Close()
		return
	}
	primitive.update(proposal)
}

func (m *primitiveManager) read(query *raftSessionOperationQuery) {
	primitive, ok := m.primitives[query.Input().PrimitiveID]
	if !ok {
		query.Error(errors.NewFault("primitive not found"))
		query.Close()
		return
	}
	primitive.read(query)
}

func (m *primitiveManager) close(command *raftSessionClosePrimitiveCommand) {
	primitive, ok := m.primitives[command.Input().PrimitiveID]
	if !ok {
		command.Output(&multiraftv1.ClosePrimitiveOutput{})
		command.Close()
		return
	}
	primitive.close(command)
}

type primitiveExecutor interface {
	info() primitiveInfo
	snapshot(*snapshot.Writer) error
	recover(*snapshot.Reader) error
	create(*raftSessionCreatePrimitiveCommand)
	update(*raftSessionOperationCommand)
	read(*raftSessionOperationQuery)
	close(*raftSessionClosePrimitiveCommand)
}

type primitiveInfo struct {
	id   multiraftv1.PrimitiveID
	spec multiraftv1.PrimitiveSpec
}

func newPrimitiveContext[I, O any](manager *primitiveManager, info primitiveInfo, primitiveType PrimitiveType[I, O]) *primitiveContext[I, O] {
	return &primitiveContext[I, O]{
		primitiveInfo: info,
		manager:       manager,
		primitiveType: primitiveType,
		sessions:      newPrimitiveSessions[I, O](),
		proposals:     newPrimitiveProposals[I, O](),
	}
}

type primitiveContext[I, O any] struct {
	primitiveInfo
	manager       *primitiveManager
	sessions      *primitiveSessions[I, O]
	primitiveType PrimitiveType[I, O]
	proposals     *primitiveProposals[I, O]
}

func (p *primitiveContext[I, O]) info() primitiveInfo {
	return p.primitiveInfo
}

func (p *primitiveContext[I, O]) PrimitiveID() PrimitiveID {
	return PrimitiveID(p.id)
}

func (p *primitiveContext[I, O]) Namespace() string {
	return p.spec.Namespace
}

func (p *primitiveContext[I, O]) Name() string {
	return p.spec.Name
}

func (p *primitiveContext[I, O]) Type() PrimitiveType[I, O] {
	return p.primitiveType
}

func (p *primitiveContext[I, O]) Index() Index {
	return Index(p.manager.context.context.index)
}

func (p *primitiveContext[I, O]) Time() time.Time {
	return p.manager.context.context.time
}

func (p *primitiveContext[I, O]) Scheduler() *Scheduler {
	return p.manager.context.context.scheduler
}

func (p *primitiveContext[I, O]) Sessions() Sessions[I, O] {
	return p.sessions
}

func (p *primitiveContext[I, O]) Proposals() Proposals[I, O] {
	return p.proposals
}

func newPrimitiveStateMachineExecutor[I, O any](context *primitiveContext[I, O], stateMachine Primitive[I, O]) primitiveExecutor {
	return &primitiveStateMachineExecutor[I, O]{
		primitiveContext: context,
		stateMachine:     stateMachine,
	}
}

type primitiveStateMachineExecutor[I, O any] struct {
	*primitiveContext[I, O]
	stateMachine Primitive[I, O]
}

func (p *primitiveStateMachineExecutor[I, O]) snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(p.sessions.sessions)); err != nil {
		return err
	}
	for _, session := range p.sessions.sessions {
		if err := writer.WriteVarUint64(uint64(session.ID())); err != nil {
			return err
		}
	}
	return p.stateMachine.Snapshot(writer)
}

func (p *primitiveStateMachineExecutor[I, O]) recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		sessionID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}
		parent, ok := p.manager.context.sessions[multiraftv1.SessionID(sessionID)]
		if !ok {
			return errors.NewFault("session %d not found", sessionID)
		}
		session := newPrimitiveSession(p, parent)
		session.open()
		for _, parentCommand := range parent.commands {
			if parentCommand.state != multiraftv1.CommandSnapshot_RUNNING {
				continue
			}
			operation := parentCommand.Operation()
			if operation.Input().PrimitiveID == multiraftv1.PrimitiveID(p.id) {
				command, err := newPrimitiveProposal[I, O](session, operation)
				if err != nil {
					return err
				}
				command.open()
			}
		}
	}
	return p.stateMachine.Recover(reader)
}

func (p *primitiveStateMachineExecutor[I, O]) create(command *raftSessionCreatePrimitiveCommand) {
	_, ok := p.sessions.sessions[command.session.sessionID]
	if !ok {
		session := newPrimitiveSession[I, O](p, command.session)
		session.open()
	}
	command.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: p.id,
	})
	command.Close()
}

func (p *primitiveStateMachineExecutor[I, O]) update(command *raftSessionOperationCommand) {
	primitiveSession, ok := p.sessions.sessions[command.session.sessionID]
	if !ok {
		command.Error(errors.NewFault("session not found"))
		command.Close()
		return
	}
	primitiveCommand, err := newPrimitiveProposal[I, O](primitiveSession, command)
	if err != nil {
		command.Error(err)
		command.Close()
	} else {
		primitiveCommand.open()
		p.stateMachine.Update(primitiveCommand)
	}
}

func (p *primitiveStateMachineExecutor[I, O]) read(query *raftSessionOperationQuery) {
	primitiveSession, ok := p.sessions.sessions[query.session.sessionID]
	if !ok {
		query.Error(errors.NewFault("session not found"))
		query.Close()
		return
	}
	primitiveQuery, err := newPrimitiveQuery[I, O](primitiveSession, query)
	if err != nil {
		query.Error(err)
		query.Close()
	} else {
		p.stateMachine.Read(primitiveQuery)
	}
}

func (p *primitiveStateMachineExecutor[I, O]) close(command *raftSessionClosePrimitiveCommand) {
	session, ok := p.sessions.sessions[command.session.sessionID]
	if ok {
		session.close()
	}
	command.Output(&multiraftv1.ClosePrimitiveOutput{})
	command.Close()
}

func newPrimitiveSessions[I, O any]() *primitiveSessions[I, O] {
	return &primitiveSessions[I, O]{
		sessions: make(map[multiraftv1.SessionID]*primitiveSession[I, O]),
	}
}

type primitiveSessions[I, O any] struct {
	sessions map[multiraftv1.SessionID]*primitiveSession[I, O]
}

func (s *primitiveSessions[I, O]) add(session *primitiveSession[I, O]) {
	s.sessions[session.session.sessionID] = session
}

func (s *primitiveSessions[I, O]) remove(sessionID multiraftv1.SessionID) {
	delete(s.sessions, sessionID)
}

func (s *primitiveSessions[I, O]) Get(id multiraftv1.SessionID) (Session[I, O], bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *primitiveSessions[I, O]) List() []Session[I, O] {
	sessions := make([]Session[I, O], 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func newPrimitiveSession[I, O any](context *primitiveStateMachineExecutor[I, O], parent *raftSession) *primitiveSession[I, O] {
	return &primitiveSession[I, O]{
		primitive: context,
		session:   parent,
		proposals: newPrimitiveProposals[I, O](),
		state:     multiraftv1.SessionSnapshot_OPEN,
	}
}

type primitiveSession[I, O any] struct {
	primitive *primitiveStateMachineExecutor[I, O]
	session   *raftSession
	proposals *primitiveProposals[I, O]
	state     multiraftv1.SessionSnapshot_State
	watchers  map[string]SessionWatcher
}

func (s *primitiveSession[I, O]) ID() SessionID {
	return SessionID(s.session.sessionID)
}

func (s *primitiveSession[I, O]) State() SessionState {
	switch s.session.state {
	case multiraftv1.SessionSnapshot_OPEN:
		return SessionOpen
	case multiraftv1.SessionSnapshot_CLOSED:
		return SessionClosed
	default:
		panic("unknown session state")
	}
}

func (s *primitiveSession[I, O]) Watch(f SessionWatcher) CancelFunc {
	if s.watchers == nil {
		s.watchers = make(map[string]SessionWatcher)
	}
	id := uuid.New().String()
	s.watchers[id] = f
	return func() {
		delete(s.watchers, id)
	}
}

func (s *primitiveSession[I, O]) Proposals() Proposals[I, O] {
	return s.proposals
}

func (s *primitiveSession[I, O]) open() {
	s.primitive.sessions.add(s)
	s.session.closers[multiraftv1.PrimitiveID(s.primitive.id)] = s.close
	s.state = multiraftv1.SessionSnapshot_OPEN
}

func (s *primitiveSession[I, O]) close() {
	s.primitive.sessions.remove(s.session.sessionID)
	delete(s.session.closers, s.primitive.id)
	s.state = multiraftv1.SessionSnapshot_CLOSED
	if s.watchers != nil {
		for _, watcher := range s.watchers {
			watcher(SessionClosed)
		}
	}
}

func newPrimitiveProposals[I, O any]() *primitiveProposals[I, O] {
	return &primitiveProposals[I, O]{
		proposals: make(map[multiraftv1.Index]*primitiveProposal[I, O]),
	}
}

type primitiveProposals[I, O any] struct {
	proposals map[multiraftv1.Index]*primitiveProposal[I, O]
}

func (c *primitiveProposals[I, O]) add(proposal *primitiveProposal[I, O]) {
	c.proposals[proposal.command.index] = proposal
}

func (c *primitiveProposals[I, O]) remove(index multiraftv1.Index) {
	delete(c.proposals, index)
}

func (c *primitiveProposals[I, O]) Get(id ProposalID) (Proposal[I, O], bool) {
	proposal, ok := c.proposals[multiraftv1.Index(id)]
	return proposal, ok
}

func (c *primitiveProposals[I, O]) List() []Proposal[I, O] {
	commands := make([]Proposal[I, O], 0, len(c.proposals))
	for _, command := range c.proposals {
		commands = append(commands, command)
	}
	return commands
}

func newPrimitiveProposal[I, O any](session *primitiveSession[I, O], command *raftSessionOperationCommand) (*primitiveProposal[I, O], error) {
	input, err := session.primitive.primitiveType.Codec().DecodeInput(command.Input().Payload)
	if err != nil {
		return nil, errors.NewFault(err.Error())
	}
	proposal := &primitiveProposal[I, O]{
		session: session,
		command: command,
		input:   input,
	}
	command.closer = proposal.close
	return proposal, nil
}

type primitiveProposal[I, O any] struct {
	session  *primitiveSession[I, O]
	command  *raftSessionOperationCommand
	input    I
	watchers map[string]OperationWatcher
}

func (c *primitiveProposal[I, O]) ID() ProposalID {
	return ProposalID(c.command.index)
}

func (c *primitiveProposal[I, O]) Watch(f OperationWatcher) CancelFunc {
	if c.watchers == nil {
		c.watchers = make(map[string]OperationWatcher)
	}
	id := uuid.New().String()
	c.watchers[id] = f
	return func() {
		delete(c.watchers, id)
	}
}

func (c *primitiveProposal[I, O]) Session() Session[I, O] {
	return c.session
}

func (c *primitiveProposal[I, O]) Input() I {
	return c.input
}

func (c *primitiveProposal[I, O]) Output(output O) {
	bytes, err := c.session.primitive.primitiveType.Codec().EncodeOutput(output)
	if err != nil {
		c.command.Error(errors.NewInternal(err.Error()))
	} else {
		c.command.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Payload: bytes,
			},
		})
	}
}

func (c *primitiveProposal[I, O]) Error(err error) {
	c.command.Error(err)
}

func (c *primitiveProposal[I, O]) open() {
	c.session.primitive.proposals.add(c)
	c.session.proposals.add(c)
}

func (c *primitiveProposal[I, O]) close() {
	c.session.proposals.remove(c.command.index)
	c.session.primitive.proposals.remove(c.command.index)
	if c.watchers != nil {
		for _, watcher := range c.watchers {
			watcher(Complete)
		}
	}
}

func (c *primitiveProposal[I, O]) Close() {
	c.command.Close()
}

func newPrimitiveQuery[I, O any](session *primitiveSession[I, O], query *raftSessionOperationQuery) (*primitiveQuery[I, O], error) {
	input, err := session.primitive.primitiveType.Codec().DecodeInput(query.Input().Payload)
	if err != nil {
		return nil, errors.NewInternal(err.Error())
	}
	return &primitiveQuery[I, O]{
		session: session,
		query:   query,
		input:   input,
	}, nil
}

type primitiveQuery[I, O any] struct {
	session *primitiveSession[I, O]
	query   *raftSessionOperationQuery
	input   I
}

func (q *primitiveQuery[I, O]) ID() QueryID {
	return QueryID(q.query.sequenceNum)
}

func (q *primitiveQuery[I, O]) Watch(f OperationWatcher) CancelFunc {
	return q.query.Watch(f)
}

func (q *primitiveQuery[I, O]) Session() Session[I, O] {
	return q.session
}

func (q *primitiveQuery[I, O]) Input() I {
	return q.input
}

func (q *primitiveQuery[I, O]) Output(output O) {
	bytes, err := q.session.primitive.primitiveType.Codec().EncodeOutput(output)
	if err != nil {
		q.query.Error(errors.NewInternal(err.Error()))
	} else {
		q.query.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Payload: bytes,
			},
		})
	}
}

func (q *primitiveQuery[I, O]) Error(err error) {
	q.query.Error(err)
}

func (q *primitiveQuery[I, O]) Close() {
	q.query.Close()
}
