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
		primitives: make(map[PrimitiveID]primitiveExecutor),
	}
}

type primitiveManager struct {
	registry   *PrimitiveTypeRegistry
	context    *sessionManager
	primitives map[PrimitiveID]primitiveExecutor
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
		factory, ok := m.registry.lookup(snapshot.Spec.Type.Name, snapshot.Spec.Type.ApiVersion)
		if !ok {
			return errors.NewFault("primitive type not found")
		}

		info := primitiveInfo{
			id:   PrimitiveID(snapshot.PrimitiveID),
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
			if p.info().spec.Type.Name != command.Input().Type.Name || p.info().spec.Type.ApiVersion != command.Input().Type.ApiVersion {
				command.Output(nil, errors.NewForbidden("cannot create primitive of a different type with the same name"))
				command.Close()
				return
			}
			executor = p
			break
		}
	}

	if executor == nil {
		factory, ok := m.registry.lookup(command.Input().Type.Name, command.Input().Type.ApiVersion)
		if !ok {
			command.Output(nil, errors.NewForbidden("unknown primitive type"))
			command.Close()
			return
		}

		info := primitiveInfo{
			id:   PrimitiveID(command.index),
			spec: command.Input().PrimitiveSpec,
		}
		executor = factory(m, info)
		m.primitives[info.id] = executor
	}
	executor.create(command)
}

func (m *primitiveManager) command(command *raftSessionOperationCommand) {
	primitive, ok := m.primitives[PrimitiveID(command.Input().PrimitiveID)]
	if !ok {
		command.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Status:  multiraftv1.OperationOutput_FAULT,
				Message: "primitive not found",
			},
		}, nil)
		command.Close()
		return
	}
	primitive.command(command)
}

func (m *primitiveManager) query(query *raftSessionOperationQuery) {
	primitive, ok := m.primitives[PrimitiveID(query.Input().PrimitiveID)]
	if !ok {
		query.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Status:  multiraftv1.OperationOutput_FAULT,
				Message: "primitive not found",
			},
		}, nil)
		query.Close()
		return
	}
	primitive.query(query)
}

func (m *primitiveManager) close(command *raftSessionClosePrimitiveCommand) {
	primitive, ok := m.primitives[PrimitiveID(command.Input().PrimitiveID)]
	if !ok {
		command.Output(&multiraftv1.ClosePrimitiveOutput{}, nil)
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
	command(*raftSessionOperationCommand)
	query(*raftSessionOperationQuery)
	close(*raftSessionClosePrimitiveCommand)
}

type primitiveInfo struct {
	id   PrimitiveID
	spec multiraftv1.PrimitiveSpec
}

func newPrimitiveContext[I, O any](manager *primitiveManager, info primitiveInfo, primitiveType PrimitiveType[I, O]) *primitiveContext[I, O] {
	return &primitiveContext[I, O]{
		primitiveInfo: info,
		manager:       manager,
		primitiveType: primitiveType,
		sessions:      newPrimitiveSessions[I, O](),
		commands:      newPrimitiveCommands[I, O](),
	}
}

type primitiveContext[I, O any] struct {
	primitiveInfo
	manager       *primitiveManager
	sessions      *primitiveSessions[I, O]
	primitiveType PrimitiveType[I, O]
	commands      *primitiveCommands[I, O]
}

func (p *primitiveContext[I, O]) info() primitiveInfo {
	return p.primitiveInfo
}

func (p *primitiveContext[I, O]) PrimitiveID() PrimitiveID {
	return p.id
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

func (p *primitiveContext[I, O]) Commands() Commands[I, O] {
	return p.commands
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
				command, err := newPrimitiveCommand[I, O](session, operation)
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
	_, ok := p.sessions.sessions[SessionID(command.session.sessionID)]
	if !ok {
		session := newPrimitiveSession[I, O](p, command.session)
		session.open()
	}
	command.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(p.id),
	}, nil)
	command.Close()
}

func (p *primitiveStateMachineExecutor[I, O]) command(command *raftSessionOperationCommand) {
	primitiveSession, ok := p.sessions.sessions[SessionID(command.session.sessionID)]
	if !ok {
		command.Output(nil, errors.NewFault("session not found"))
		command.Close()
		return
	}
	primitiveCommand, err := newPrimitiveCommand[I, O](primitiveSession, command)
	if err != nil {
		command.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Status:  getStatus(err),
				Message: err.Error(),
			},
		}, nil)
		command.Close()
	} else {
		primitiveCommand.open()
		p.stateMachine.Update(primitiveCommand)
	}
}

func (p *primitiveStateMachineExecutor[I, O]) query(query *raftSessionOperationQuery) {
	primitiveSession, ok := p.sessions.sessions[SessionID(query.session.sessionID)]
	if !ok {
		query.Output(nil, errors.NewFault("session not found"))
		query.Close()
		return
	}
	primitiveQuery, err := newPrimitiveQuery[I, O](primitiveSession, query)
	if err != nil {
		query.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Status:  getStatus(err),
				Message: err.Error(),
			},
		}, nil)
		query.Close()
	} else {
		p.stateMachine.Read(primitiveQuery)
		query.Close()
	}
}

func (p *primitiveStateMachineExecutor[I, O]) close(command *raftSessionClosePrimitiveCommand) {
	session, ok := p.sessions.sessions[SessionID(command.session.sessionID)]
	if ok {
		session.close()
	}
	command.Output(&multiraftv1.ClosePrimitiveOutput{}, nil)
	command.Close()
}

func newPrimitiveSessions[I, O any]() *primitiveSessions[I, O] {
	return &primitiveSessions[I, O]{
		sessions: make(map[SessionID]*primitiveSession[I, O]),
	}
}

type primitiveSessions[I, O any] struct {
	sessions map[SessionID]*primitiveSession[I, O]
}

func (s *primitiveSessions[I, O]) add(session *primitiveSession[I, O]) {
	s.sessions[session.ID()] = session
}

func (s *primitiveSessions[I, O]) remove(sessionID SessionID) {
	delete(s.sessions, sessionID)
}

func (s *primitiveSessions[I, O]) Get(id SessionID) (Session[I, O], bool) {
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
		commands:  newPrimitiveCommands[I, O](),
		state:     SessionOpen,
	}
}

type primitiveSession[I, O any] struct {
	primitive *primitiveStateMachineExecutor[I, O]
	session   *raftSession
	commands  *primitiveCommands[I, O]
	state     SessionState
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

func (s *primitiveSession[I, O]) Commands() Commands[I, O] {
	return s.commands
}

func (s *primitiveSession[I, O]) open() {
	s.primitive.sessions.add(s)
	s.session.closers[multiraftv1.PrimitiveID(s.primitive.id)] = s.close
	s.state = SessionOpen
}

func (s *primitiveSession[I, O]) close() {
	s.primitive.sessions.remove(s.ID())
	delete(s.session.closers, multiraftv1.PrimitiveID(s.primitive.id))
	s.state = SessionClosed
	if s.watchers != nil {
		for _, watcher := range s.watchers {
			watcher(SessionClosed)
		}
	}
}

func newPrimitiveCommands[I, O any]() *primitiveCommands[I, O] {
	return &primitiveCommands[I, O]{
		commands: make(map[CommandID]*primitiveCommand[I, O]),
	}
}

type primitiveCommands[I, O any] struct {
	commands map[CommandID]*primitiveCommand[I, O]
}

func (c *primitiveCommands[I, O]) add(command *primitiveCommand[I, O]) {
	c.commands[command.ID()] = command
}

func (c *primitiveCommands[I, O]) remove(commandID CommandID) {
	delete(c.commands, commandID)
}

func (c *primitiveCommands[I, O]) Get(commandID CommandID) (Command[I, O], bool) {
	command, ok := c.commands[commandID]
	return command, ok
}

func (c *primitiveCommands[I, O]) List() []Command[I, O] {
	commands := make([]Command[I, O], 0, len(c.commands))
	for _, command := range c.commands {
		commands = append(commands, command)
	}
	return commands
}

func newPrimitiveCommand[I, O any](session *primitiveSession[I, O], command *raftSessionOperationCommand) (*primitiveCommand[I, O], error) {
	input, err := session.primitive.primitiveType.Codec().DecodeInput(command.Input().Payload)
	if err != nil {
		return nil, errors.NewFault(err.Error())
	}
	c := &primitiveCommand[I, O]{
		session: session,
		command: command,
		input:   input,
	}
	command.closer = c.close
	return c, nil
}

type primitiveCommand[I, O any] struct {
	session  *primitiveSession[I, O]
	command  *raftSessionOperationCommand
	input    I
	watchers map[string]func(CommandState)
}

func (c *primitiveCommand[I, O]) ID() CommandID {
	return CommandID(c.command.index)
}

func (c *primitiveCommand[I, O]) State() CommandState {
	switch c.command.state {
	case multiraftv1.CommandSnapshot_RUNNING:
		return CommandRunning
	case multiraftv1.CommandSnapshot_COMPLETE:
		return CommandComplete
	default:
		panic("unknown command state")
	}
}

func (c *primitiveCommand[I, O]) Watch(f CommandWatcher) CancelFunc {
	if c.watchers == nil {
		c.watchers = make(map[string]func(CommandState))
	}
	id := uuid.New().String()
	c.watchers[id] = f
	return func() {
		delete(c.watchers, id)
	}
}

func (c *primitiveCommand[I, O]) Session() Session[I, O] {
	return c.session
}

func (c *primitiveCommand[I, O]) Input() I {
	return c.input
}

func (c *primitiveCommand[I, O]) Output(output O) {
	bytes, err := c.session.primitive.primitiveType.Codec().EncodeOutput(output)
	if err != nil {
		c.command.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Status:  multiraftv1.OperationOutput_INTERNAL,
				Message: err.Error(),
			},
		}, nil)
	} else {
		c.command.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Payload: bytes,
			},
		}, nil)
	}
}

func (c *primitiveCommand[I, O]) Error(err error) {
	c.command.Output(&multiraftv1.PrimitiveOperationOutput{
		OperationOutput: multiraftv1.OperationOutput{
			Status:  getStatus(err),
			Message: getMessage(err),
		},
	}, nil)
}

func (c *primitiveCommand[I, O]) open() {
	c.session.primitive.commands.add(c)
	c.session.commands.add(c)
}

func (c *primitiveCommand[I, O]) close() {
	c.session.commands.remove(c.ID())
	c.session.primitive.commands.remove(c.ID())
	if c.watchers != nil {
		for _, watcher := range c.watchers {
			watcher(CommandComplete)
		}
	}
}

func (c *primitiveCommand[I, O]) Close() {
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

func (q *primitiveQuery[I, O]) Session() Session[I, O] {
	return q.session
}

func (q *primitiveQuery[I, O]) Input() I {
	return q.input
}

func (q *primitiveQuery[I, O]) Output(output O) {
	bytes, err := q.session.primitive.primitiveType.Codec().EncodeOutput(output)
	if err != nil {
		q.query.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Status:  multiraftv1.OperationOutput_INTERNAL,
				Message: err.Error(),
			},
		}, nil)
	} else {
		q.query.Output(&multiraftv1.PrimitiveOperationOutput{
			OperationOutput: multiraftv1.OperationOutput{
				Payload: bytes,
			},
		}, nil)
	}
}

func (q *primitiveQuery[I, O]) Error(err error) {
	q.query.Output(&multiraftv1.PrimitiveOperationOutput{
		OperationOutput: multiraftv1.OperationOutput{
			Status:  getStatus(err),
			Message: getMessage(err),
		},
	}, nil)
}
