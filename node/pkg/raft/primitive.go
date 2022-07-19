// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package raft

import (
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/node/pkg/primitive"
	"github.com/atomix/multi-raft/node/pkg/snapshot"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/google/uuid"
	"time"
)

func newPrimitiveManager(registry *primitive.Registry, context sessionManagerContext) *primitiveManager {
	return &primitiveManager{
		registry:   registry,
		context:    context,
		primitives: make(map[primitive.ID]*primitiveContext),
	}
}

type primitiveManager struct {
	registry   *primitive.Registry
	context    sessionManagerContext
	primitives map[primitive.ID]*primitiveContext
}

func (m *primitiveManager) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.primitives)); err != nil {
		return err
	}
	for _, primitive := range m.primitives {
		if err := primitive.snapshot(writer); err != nil {
			return err
		}
	}
	return nil
}

func (m *primitiveManager) Recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		primitive := newPrimitive(m)
		if err := primitive.recover(reader); err != nil {
			return nil
		}
	}
	return nil
}

func (m *primitiveManager) Create(command CreatePrimitive) {
	primitive := newPrimitive(m)
	primitive.create(command)
}

func (m *primitiveManager) Command(command PrimitiveCommand) {
	primitive, ok := m.primitives[primitive.ID(command.Input().PrimitiveID)]
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

func (m *primitiveManager) Query(query PrimitiveQuery) {
	primitive, ok := m.primitives[primitive.ID(query.Input().PrimitiveID)]
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

func (m *primitiveManager) Close(command ClosePrimitive) {
	primitive, ok := m.primitives[primitive.ID(command.Input().PrimitiveID)]
	if !ok {
		command.Output(&multiraftv1.ClosePrimitiveOutput{}, nil)
		command.Close()
		return
	}
	primitive.close(command)
}

func newPrimitive(manager *primitiveManager) *primitiveContext {
	return &primitiveContext{
		manager:  manager,
		sessions: newPrimitiveSessions(),
		commands: newPrimitiveCommands(),
	}
}

type primitiveContext struct {
	manager       *primitiveManager
	sessions      *primitiveSessions
	primitiveID   primitive.ID
	spec          multiraftv1.PrimitiveSpec
	primitiveType primitive.Type
	commands      *primitiveCommands
	stateMachine  primitive.StateMachine
}

func (p *primitiveContext) ID() primitive.ID {
	return p.primitiveID
}

func (p *primitiveContext) Type() primitive.Type {
	return p.primitiveType
}

func (p *primitiveContext) Namespace() string {
	return p.spec.Namespace
}

func (p *primitiveContext) Name() string {
	return p.spec.Name
}

func (p *primitiveContext) Index() primitive.Index {
	return primitive.Index(p.manager.context.Index())
}

func (p *primitiveContext) Time() time.Time {
	return p.manager.context.Time()
}

func (p *primitiveContext) Scheduler() primitive.Scheduler {
	return p.manager.context.Scheduler()
}

func (p *primitiveContext) Sessions() primitive.Sessions {
	return p.sessions
}

func (p *primitiveContext) Commands() primitive.Commands {
	return p.commands
}

func (p *primitiveContext) snapshot(writer *snapshot.Writer) error {
	snapshot := &multiraftv1.PrimitiveSnapshot{
		PrimitiveID: multiraftv1.PrimitiveID(p.primitiveID),
		Spec:        p.spec,
	}
	if err := writer.WriteMessage(snapshot); err != nil {
		return err
	}

	if err := writer.WriteVarInt(len(p.sessions.sessions)); err != nil {
		return err
	}
	for _, session := range p.sessions.sessions {
		if err := session.snapshot(writer); err != nil {
			return err
		}
	}
	return p.stateMachine.Snapshot(writer)
}

func (p *primitiveContext) recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.PrimitiveSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	p.primitiveID = primitive.ID(snapshot.PrimitiveID)
	p.spec = snapshot.Spec
	p.primitiveType = p.manager.registry.Get(snapshot.Spec.Type.Name, snapshot.Spec.Type.ApiVersion)

	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		session := newPrimitiveSession(p)
		if err := session.recover(reader); err != nil {
			return err
		}
	}
	return p.stateMachine.Recover(reader)
}

func (p *primitiveContext) create(command CreatePrimitive) {
	var existing *primitiveContext
	for _, p := range p.manager.primitives {
		if p.Namespace() == command.Input().Namespace && p.Name() == command.Input().Name {
			if p.Type().Name() != command.Input().Type.Name || p.Type().APIVersion() != command.Input().Type.ApiVersion {
				command.Output(nil, errors.NewForbidden("cannot create primitive of a different type with the same name"))
				command.Close()
				return
			}
			existing = p
			break
		}
	}

	if existing != nil {
		command.Output(&multiraftv1.CreatePrimitiveOutput{
			PrimitiveID: multiraftv1.PrimitiveID(existing.primitiveID),
		}, nil)
		command.Close()
		return
	}

	p.primitiveID = primitive.ID(p.manager.context.Index())
	p.spec = command.Input().PrimitiveSpec
	p.primitiveType = p.manager.registry.Get(command.Input().Type.Name, command.Input().Type.ApiVersion)
	p.stateMachine = p.primitiveType.NewStateMachine(p)

	session := newPrimitiveSession(p)
	session.create(command)
}

func (p *primitiveContext) command(command PrimitiveCommand) {
	session, ok := p.sessions.sessions[primitive.SessionID(command.Session().ID())]
	if !ok {
		command.Output(nil, errors.NewFault("session not found"))
		command.Close()
		return
	}
	session.command(command)
}

func (p *primitiveContext) query(query PrimitiveQuery) {
	session, ok := p.sessions.sessions[primitive.SessionID(query.Session().ID())]
	if !ok {
		query.Output(nil, errors.NewFault("session not found"))
		query.Close()
		return
	}
	session.query(query)
}

func (p *primitiveContext) close(command ClosePrimitive) {
	session, ok := p.sessions.sessions[primitive.SessionID(command.Session().ID())]
	if !ok {
		command.Output(&multiraftv1.ClosePrimitiveOutput{}, nil)
		command.Close()
		return
	}
	session.close(command)
}

func newPrimitiveSessions() *primitiveSessions {
	return &primitiveSessions{
		sessions: make(map[primitive.SessionID]*primitiveSession),
	}
}

type primitiveSessions struct {
	sessions map[primitive.SessionID]*primitiveSession
}

func (s *primitiveSessions) add(session *primitiveSession) {
	s.sessions[session.ID()] = session
}

func (s *primitiveSessions) remove(sessionID primitive.SessionID) {
	delete(s.sessions, sessionID)
}

func (s *primitiveSessions) Get(id primitive.SessionID) (primitive.Session, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *primitiveSessions) List() []primitive.Session {
	sessions := make([]primitive.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func newPrimitiveSession(context *primitiveContext) *primitiveSession {
	return &primitiveSession{
		primitive: context,
		commands:  newPrimitiveCommands(),
		watchers:  make(map[string]func(primitive.SessionState)),
	}
}

type primitiveSession struct {
	primitive     *primitiveContext
	session       Session
	commands      *primitiveCommands
	state         primitive.SessionState
	watchers      map[string]func(primitive.SessionState)
	parentWatcher Watcher
}

func (s *primitiveSession) ID() primitive.SessionID {
	return primitive.SessionID(s.session.ID())
}

func (s *primitiveSession) State() primitive.SessionState {
	return primitive.SessionState(s.session.State())
}

func (s *primitiveSession) Watch(f func(primitive.SessionState)) primitive.Watcher {
	id := uuid.New().String()
	s.watchers[id] = f
	return newSessionWatcher(func() {
		delete(s.watchers, id)
	})
}

func (s *primitiveSession) Commands() primitive.Commands {
	return s.commands
}

func (s *primitiveSession) snapshot(writer *snapshot.Writer) error {
	commands := make([]multiraftv1.Index, 0, len(s.commands.commands))
	for _, command := range s.commands.commands {
		commands = append(commands, command.command.Index())
	}
	snapshot := &multiraftv1.PrimitiveSessionSnapshot{
		SessionID: s.session.ID(),
		Commands:  commands,
	}
	return writer.WriteMessage(snapshot)
}

func (s *primitiveSession) recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.PrimitiveSessionSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	session, ok := s.primitive.manager.context.Sessions().Get(snapshot.SessionID)
	if !ok {
		return errors.NewFault("session %d not found", snapshot.SessionID)
	}
	s.session = session
	s.state = primitive.SessionOpen
	s.parentWatcher = s.session.Watch(func(state SessionState) {
		if s.state == primitive.SessionOpen && state == SessionClosed {
			s.primitive.sessions.remove(s.ID())
			s.state = primitive.SessionClosed
			for _, watcher := range s.watchers {
				watcher(primitive.SessionClosed)
			}
		}
	})
	s.primitive.sessions.add(s)
	for _, index := range snapshot.Commands {
		command, ok := session.Commands().Get(index)
		if !ok {
			return errors.NewFault("session %d command %d not found", snapshot.SessionID, index)
		}
		newPrimitiveCommand(s, command.Operation())
	}
	return nil
}

func (s *primitiveSession) create(command CreatePrimitive) {
	s.session = command.Session()
	s.state = primitive.SessionOpen
	s.parentWatcher = s.session.Watch(func(state SessionState) {
		if s.state == primitive.SessionOpen && state == SessionClosed {
			s.primitive.sessions.remove(s.ID())
			s.state = primitive.SessionClosed
			for _, watcher := range s.watchers {
				watcher(primitive.SessionClosed)
			}
		}
	})
	s.primitive.sessions.add(s)
	command.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(s.primitive.primitiveID),
	}, nil)
	command.Close()
}

func (s *primitiveSession) command(command PrimitiveCommand) {
	s.primitive.stateMachine.Command(newPrimitiveCommand(s, command))
}

func (s *primitiveSession) query(query PrimitiveQuery) {
	s.primitive.stateMachine.Query(newPrimitiveQuery(s, query))
}

func (s *primitiveSession) close(command ClosePrimitive) {
	s.parentWatcher.Cancel()
	s.primitive.sessions.remove(s.ID())
	s.state = primitive.SessionClosed
	for _, watcher := range s.watchers {
		watcher(primitive.SessionClosed)
	}
	command.Output(&multiraftv1.ClosePrimitiveOutput{}, nil)
	command.Close()
}

func newPrimitiveCommands() *primitiveCommands {
	return &primitiveCommands{
		commands: make(map[primitive.CommandID]*primitiveCommand),
	}
}

type primitiveCommands struct {
	commands map[primitive.CommandID]*primitiveCommand
}

func (c *primitiveCommands) add(command *primitiveCommand) {
	c.commands[command.ID()] = command
}

func (c *primitiveCommands) remove(commandID primitive.CommandID) {
	delete(c.commands, commandID)
}

func (c *primitiveCommands) Get(commandID primitive.CommandID) (primitive.Command, bool) {
	command, ok := c.commands[commandID]
	return command, ok
}

func (c *primitiveCommands) List(id primitive.OperationID) []primitive.Command {
	var commands []primitive.Command
	for _, command := range c.commands {
		if command.OperationID() == id {
			commands = append(commands, command)
		}
	}
	return commands
}

func newPrimitiveCommand(session *primitiveSession, command PrimitiveCommand) primitive.Command {
	c := &primitiveCommand{
		session: session,
		command: command,
	}
	session.commands.add(c)
	command.Watch(func(state CommandState) {
		session.commands.remove(c.ID())
	})
	return c
}

type primitiveCommand struct {
	session *primitiveSession
	command PrimitiveCommand
}

func (c *primitiveCommand) ID() primitive.CommandID {
	return primitive.CommandID(c.command.Index())
}

func (c *primitiveCommand) State() primitive.CommandState {
	return primitive.CommandState(c.command.State())
}

func (c *primitiveCommand) Watch(f func(primitive.CommandState)) primitive.Watcher {
	return c.command.Watch(func(state CommandState) {
		f(primitive.CommandState(state))
	})
}

func (c *primitiveCommand) OperationID() primitive.OperationID {
	return primitive.OperationID(c.command.Input().ID)
}

func (c *primitiveCommand) Session() primitive.Session {
	return c.session
}

func (c *primitiveCommand) Input() []byte {
	return c.command.Input().Payload
}

func (c *primitiveCommand) Output(payload []byte, err error) {
	c.command.Output(&multiraftv1.PrimitiveOperationOutput{
		OperationOutput: multiraftv1.OperationOutput{
			Status:  getStatus(err),
			Message: getMessage(err),
			Payload: payload,
		},
	}, nil)
}

func (c *primitiveCommand) Close() {
	c.command.Close()
}

func newPrimitiveQuery(session *primitiveSession, query PrimitiveQuery) primitive.Query {
	return &primitiveQuery{
		session: session,
		query:   query,
	}
}

type primitiveQuery struct {
	session *primitiveSession
	query   PrimitiveQuery
}

func (q *primitiveQuery) OperationID() primitive.OperationID {
	return primitive.OperationID(q.query.Input().ID)
}

func (q *primitiveQuery) Session() primitive.Session {
	return q.session
}

func (q *primitiveQuery) Input() []byte {
	return q.query.Input().Payload
}

func (q *primitiveQuery) Output(payload []byte, err error) {
	q.query.Output(&multiraftv1.PrimitiveOperationOutput{
		OperationOutput: multiraftv1.OperationOutput{
			Status:  getStatus(err),
			Message: getMessage(err),
			Payload: payload,
		},
	}, nil)
}

func (q *primitiveQuery) Close() {
	q.query.Close()
}
