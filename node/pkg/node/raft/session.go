// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package raft

import (
	"container/list"
	"encoding/binary"
	"encoding/json"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/node/pkg/node/statemachines3/primitive"
	"github.com/atomix/multi-raft/node/pkg/node/statemachines3/snapshot"
	"github.com/atomix/runtime/pkg/errors"
	streams "github.com/atomix/runtime/pkg/stream"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"time"
)

type SessionManagerContext interface {
	// Index returns the current partition index
	Index() multiraftv1.Index
	// Time returns the current partition time
	Time() time.Time
	// Scheduler returns the partition scheduler
	Scheduler() *Scheduler
	// Sessions returns the open sessions
	Sessions() Sessions
}

// Sessions provides access to open sessions
type Sessions interface {
	// Get gets a session by ID
	Get(multiraftv1.SessionID) (Session, bool)
	// List lists all open sessions
	List() []Session
}

// Session is a service session
type Session interface {
	// ID returns the session identifier
	ID() multiraftv1.SessionID
	// State returns the current session state
	State() SessionState
	// Watch watches the session state
	Watch(f func(SessionState)) Watcher
	// Commands returns the session commands
	Commands() SessionCommands
}

type SessionState int

const (
	SessionClosed SessionState = iota
	SessionOpen
)

// Operation is a command or query operation
type Operation[I, O proto.Message] interface {
	// Session returns the session executing the operation
	Session() Session
	// Input returns the operation input
	Input() I
	// Output returns the operation output
	Output(O, error)
	// Close closes the operation
	Close()
}

// SessionCommands provides access to pending commands
type SessionCommands interface {
	// Get gets a command by ID
	Get(multiraftv1.Index) (SessionCommand, bool)
}

// Command is a command operation
type Command[I, O proto.Message] interface {
	Operation[I, O]
	// Index returns the command index
	Index() multiraftv1.Index
	// State returns the current command state
	State() CommandState
	// Watch watches the command state
	Watch(f func(CommandState)) Watcher
}

type SessionCommand interface {
	Command[*multiraftv1.SessionCommandInput, *multiraftv1.SessionCommandOutput]
	CreatePrimitive() CreatePrimitive
	ClosePrimitive() ClosePrimitive
	Operation() PrimitiveCommand
}

type CreatePrimitive = Command[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]

type ClosePrimitive = Command[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]

type PrimitiveCommand = Command[*multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput]

type CommandState int

const (
	CommandPending CommandState = iota
	CommandRunning
	CommandComplete
)

// Query is a query operation
type Query[I, O proto.Message] interface {
	Operation[I, O]
}

type SessionQuery interface {
	Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]
	Operation() PrimitiveQuery
}

type PrimitiveQuery = Query[*multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput]

// Watcher is a context for a Watch call
type Watcher interface {
	// Cancel cancels the watcher
	Cancel()
}

type SessionManager interface {
	Snapshot(writer *snapshot.Writer) error
	Recover(reader *snapshot.Reader) error
	OpenSession(input *multiraftv1.OpenSessionInput, stream streams.WriteStream[*multiraftv1.OpenSessionOutput])
	KeepAlive(input *multiraftv1.KeepAliveInput, stream streams.WriteStream[*multiraftv1.KeepAliveOutput])
	CloseSession(input *multiraftv1.CloseSessionInput, stream streams.WriteStream[*multiraftv1.CloseSessionOutput])
	CommandSession(input *multiraftv1.SessionCommandInput, stream streams.WriteStream[*multiraftv1.SessionCommandOutput])
	QuerySession(input *multiraftv1.SessionQueryInput, stream streams.WriteStream[*multiraftv1.SessionQueryOutput])
}

func newSessionManager(registry *primitive.Registry, context Context) SessionManager {
	sessionManager := &raftSessionManager{
		context:  context,
		sessions: newSessions(),
	}
	sessionManager.primitives = newPrimitiveManager(registry, sessionManager)
	return sessionManager
}

type raftSessionManager struct {
	context    Context
	sessions   *raftSessions
	primitives PrimitiveManager
	prevTime   time.Time
}

func (m *raftSessionManager) Index() multiraftv1.Index {
	return m.context.Index()
}

func (m *raftSessionManager) Time() time.Time {
	return m.context.Time()
}

func (m *raftSessionManager) Scheduler() *Scheduler {
	return m.context.Scheduler()
}

func (m *raftSessionManager) Sessions() Sessions {
	return m.sessions
}

func (m *raftSessionManager) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.sessions.sessions)); err != nil {
		return err
	}
	for _, session := range m.sessions.sessions {
		if err := session.snapshot(writer); err != nil {
			return err
		}
	}
	return m.primitives.Snapshot(writer)
}

func (m *raftSessionManager) Recover(reader *snapshot.Reader) error {
	m.sessions = newSessions()
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		session := newSession(m)
		if err := session.recover(reader); err != nil {
			return err
		}
	}
	return m.primitives.Recover(reader)
}

func (m *raftSessionManager) OpenSession(input *multiraftv1.OpenSessionInput, stream streams.WriteStream[*multiraftv1.OpenSessionOutput]) {
	session := newSession(m)
	session.open(input, stream)
	m.prevTime = m.Time()
}

func (m *raftSessionManager) KeepAlive(input *multiraftv1.KeepAliveInput, stream streams.WriteStream[*multiraftv1.KeepAliveOutput]) {
	session, ok := m.sessions.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	session.keepAlive(input, stream)

	// Compute the minimum session timeout
	var minSessionTimeout time.Duration
	for _, session := range m.sessions.sessions {
		if session.timeout > minSessionTimeout {
			minSessionTimeout = session.timeout
		}
	}

	// Compute the maximum time at which sessions may be expired.
	// If no keep-alive has been received from any session for more than the minimum session
	// timeout, suspect a stop-the-world pause may have occurred. We decline to expire any
	// of the sessions in this scenario, instead resetting the timestamps for all the sessions.
	// Only expire a session if keep-alives have been received from other sessions during the
	// session's expiration period.
	maxExpireTime := m.prevTime.Add(minSessionTimeout)
	for _, session := range m.sessions.sessions {
		if m.Time().After(maxExpireTime) {
			session.resetTime(m.Time())
		}
		if m.Time().After(session.expireTime()) {
			log.Infof("Session %d expired after %s", session.sessionID, m.Time().Sub(session.lastUpdated))
			session.expire()
		}
	}
	m.prevTime = m.Time()
}

func (m *raftSessionManager) CloseSession(input *multiraftv1.CloseSessionInput, stream streams.WriteStream[*multiraftv1.CloseSessionOutput]) {
	session, ok := m.sessions.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	session.close(input, stream)
	m.prevTime = m.Time()
}

func (m *raftSessionManager) CommandSession(input *multiraftv1.SessionCommandInput, stream streams.WriteStream[*multiraftv1.SessionCommandOutput]) {
	session, ok := m.sessions.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	command := newSessionCommand(session)
	command.execute(input, stream)
	m.prevTime = m.Time()
}

func (m *raftSessionManager) QuerySession(input *multiraftv1.SessionQueryInput, stream streams.WriteStream[*multiraftv1.SessionQueryOutput]) {
	session, ok := m.sessions.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	query := newSessionQuery(session)
	query.execute(input, stream)
}

func newSessions() *raftSessions {
	return &raftSessions{
		sessions: make(map[multiraftv1.SessionID]*raftSession),
	}
}

type raftSessions struct {
	sessions map[multiraftv1.SessionID]*raftSession
}

func (s *raftSessions) add(session *raftSession) {
	s.sessions[session.sessionID] = session
}

func (s *raftSessions) remove(sessionID multiraftv1.SessionID) {
	delete(s.sessions, sessionID)
}

func (s *raftSessions) Get(id multiraftv1.SessionID) (Session, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *raftSessions) List() []Session {
	sessions := make([]Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func newSession(manager *raftSessionManager) *raftSession {
	return &raftSession{
		manager:  manager,
		watchers: make(map[string]func(SessionState)),
	}
}

type raftSession struct {
	manager     *raftSessionManager
	commands    *raftSessionCommands
	sessionID   multiraftv1.SessionID
	timeout     time.Duration
	reset       bool
	lastUpdated time.Time
	state       SessionState
	watchers    map[string]func(SessionState)
}

func (s *raftSession) ID() multiraftv1.SessionID {
	return s.sessionID
}

func (s *raftSession) State() SessionState {
	return s.state
}

func (s *raftSession) Watch(f func(SessionState)) Watcher {
	id := uuid.New().String()
	s.watchers[id] = f
	return newSessionWatcher(func() {
		delete(s.watchers, id)
	})
}

func (s *raftSession) Commands() SessionCommands {
	return s.commands
}

func (s *raftSession) resetTime(t time.Time) {
	if !s.reset {
		s.lastUpdated = t
		s.reset = true
	}
}

func (s *raftSession) expireTime() time.Time {
	return s.lastUpdated.Add(s.timeout)
}

func (s *raftSession) open(input *multiraftv1.OpenSessionInput, stream streams.WriteStream[*multiraftv1.OpenSessionOutput]) {
	s.sessionID = multiraftv1.SessionID(s.manager.Index())
	s.lastUpdated = s.manager.Time()
	s.timeout = input.Timeout
	s.state = SessionOpen
	s.manager.sessions.add(s)
	stream.Value(&multiraftv1.OpenSessionOutput{
		SessionID: s.sessionID,
	})
	stream.Close()
}

func (s *raftSession) keepAlive(input *multiraftv1.KeepAliveInput, stream streams.WriteStream[*multiraftv1.KeepAliveOutput]) {
	openInputs := &bloom.BloomFilter{}
	if err := json.Unmarshal(input.InputFilter, openInputs); err != nil {
		log.Warn("Failed to decode request filter", err)
		stream.Error(errors.NewInvalid("invalid request filter", err))
		return
	}

	log.Debugf("Keep-alive %s", s)
	for _, command := range s.commands.commands {
		if input.LastInputSequenceNum < command.input.SequenceNum {
			continue
		}
		sequenceNumBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceNumBytes, uint64(command.input.SequenceNum))
		if !openInputs.Test(sequenceNumBytes) {
			switch command.state {
			case CommandRunning:
				log.Debugf("Canceled %s", command)
				command.Close()
			case CommandComplete:
				log.Debugf("Acked %s", command)
			}
			s.commands.remove(command.index)
		} else {
			if outputSequenceNum, ok := input.LastOutputSequenceNums[command.input.SequenceNum]; ok {
				log.Debugf("Acked %s responses up to %d", command, outputSequenceNum)
				command.ack(outputSequenceNum)
			}
		}
	}

	stream.Value(&multiraftv1.KeepAliveOutput{})
	stream.Close()
}

func (s *raftSession) close(input *multiraftv1.CloseSessionInput, stream streams.WriteStream[*multiraftv1.CloseSessionOutput]) {
	s.manager.sessions.remove(s.sessionID)
	s.state = SessionClosed
	for _, watcher := range s.watchers {
		watcher(SessionClosed)
	}
	stream.Value(&multiraftv1.CloseSessionOutput{})
	stream.Close()
}

func (s *raftSession) expire() {
	s.manager.sessions.remove(s.sessionID)
	s.state = SessionClosed
	for _, watcher := range s.watchers {
		watcher(SessionClosed)
	}
}

func (s *raftSession) snapshot(writer *snapshot.Writer) error {
	snapshot := &multiraftv1.SessionSnapshot{
		SessionID:   s.sessionID,
		Timeout:     s.timeout,
		LastUpdated: s.lastUpdated,
	}
	if err := writer.WriteMessage(snapshot); err != nil {
		return err
	}
	if err := writer.WriteVarInt(len(s.commands.commands)); err != nil {
		return err
	}
	for _, command := range s.commands.commands {
		if err := command.snapshot(writer); err != nil {
			return err
		}
	}
	return nil
}

func (s *raftSession) recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.SessionSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	s.sessionID = snapshot.SessionID
	s.timeout = snapshot.Timeout
	s.lastUpdated = snapshot.LastUpdated
	s.commands = newCommands()
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		command := newSessionCommand(s)
		if err := command.recover(reader); err != nil {
			return err
		}
	}
	s.state = SessionOpen
	s.manager.sessions.add(s)
	return nil
}

func newCommands() *raftSessionCommands {
	return &raftSessionCommands{
		commands: make(map[multiraftv1.Index]*raftSessionCommand),
	}
}

type raftSessionCommands struct {
	commands map[multiraftv1.Index]*raftSessionCommand
}

func (c *raftSessionCommands) add(command *raftSessionCommand) {
	c.commands[command.index] = command
}

func (c *raftSessionCommands) remove(index multiraftv1.Index) {
	delete(c.commands, index)
}

func (c *raftSessionCommands) Get(index multiraftv1.Index) (SessionCommand, bool) {
	command, ok := c.commands[index]
	return command, ok
}

func newSessionCommand(session *raftSession) *raftSessionCommand {
	return &raftSessionCommand{
		session: session,
	}
}

type raftSessionCommand struct {
	session  *raftSession
	index    multiraftv1.Index
	state    CommandState
	watchers map[string]func(CommandState)
	input    *multiraftv1.SessionCommandInput
	outputs  *list.List
	stream   streams.WriteStream[*multiraftv1.SessionCommandOutput]
}

func (c *raftSessionCommand) Index() multiraftv1.Index {
	return c.index
}

func (c *raftSessionCommand) Session() Session {
	return c.session
}

func (c *raftSessionCommand) State() CommandState {
	return c.state
}

func (c *raftSessionCommand) Watch(f func(CommandState)) Watcher {
	id := uuid.New().String()
	c.watchers[id] = f
	return newSessionWatcher(func() {
		delete(c.watchers, id)
	})
}

func (c *raftSessionCommand) Operation() PrimitiveCommand {
	return newSessionOperationCommand(c)
}

func (c *raftSessionCommand) CreatePrimitive() CreatePrimitive {
	return newSessionCreatePrimitiveCommand(c)
}

func (c *raftSessionCommand) ClosePrimitive() ClosePrimitive {
	return newSessionClosePrimitiveCommand(c)
}

func (c *raftSessionCommand) Input() *multiraftv1.SessionCommandInput {
	return c.input
}

func (c *raftSessionCommand) execute(input *multiraftv1.SessionCommandInput, stream streams.WriteStream[*multiraftv1.SessionCommandOutput]) {
	c.stream = stream
	switch c.state {
	case CommandPending:
		c.index = c.session.manager.Index()
		c.input = input
		c.outputs = list.New()
		c.session.commands.add(c)
		c.state = CommandRunning
		log.Debugf("Executing %s: %.250s", c, input)
		switch input.Input.(type) {
		case *multiraftv1.SessionCommandInput_Operation:
			c.session.manager.primitives.Command(c.Operation())
		case *multiraftv1.SessionCommandInput_CreatePrimitive:
			c.session.manager.primitives.Create(c.CreatePrimitive())
		case *multiraftv1.SessionCommandInput_ClosePrimitive:
			c.session.manager.primitives.Close(c.ClosePrimitive())
		}
	case CommandComplete:
		defer stream.Close()
		fallthrough
	case CommandRunning:
		if c.outputs.Len() > 0 {
			log.Debugf("Replaying %d responses for %s: %.250s", c.outputs.Len(), c, input)
			elem := c.outputs.Front()
			for elem != nil {
				output := elem.Value.(*multiraftv1.SessionCommandOutput)
				stream.Value(output)
				elem = elem.Next()
			}
		}
	}
}

func (c *raftSessionCommand) snapshot(writer *snapshot.Writer) error {
	pendingOutputs := make([]*multiraftv1.SessionCommandOutput, 0, c.outputs.Len())
	elem := c.outputs.Front()
	for elem != nil {
		pendingOutputs = append(pendingOutputs, elem.Value.(*multiraftv1.SessionCommandOutput))
		elem = elem.Next()
	}
	var state multiraftv1.CommandSnapshot_State
	switch c.state {
	case CommandRunning:
		state = multiraftv1.CommandSnapshot_OPEN
	case CommandComplete:
		state = multiraftv1.CommandSnapshot_COMPLETE
	}
	snapshot := &multiraftv1.CommandSnapshot{
		Index:          c.index,
		State:          state,
		Input:          c.input,
		PendingOutputs: pendingOutputs,
	}
	return writer.WriteMessage(snapshot)
}

func (c *raftSessionCommand) recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.CommandSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	c.index = snapshot.Index
	c.input = snapshot.Input
	c.outputs = list.New()
	for _, output := range snapshot.PendingOutputs {
		r := output
		c.outputs.PushBack(&r)
	}
	c.stream = streams.NewNilStream[*multiraftv1.SessionCommandOutput]()
	switch snapshot.State {
	case multiraftv1.CommandSnapshot_OPEN:
		c.state = CommandRunning
		c.session.commands.add(c)
	case multiraftv1.CommandSnapshot_COMPLETE:
		c.state = CommandComplete
	}
	return nil
}

func (c *raftSessionCommand) Output(output *multiraftv1.SessionCommandOutput, err error) {
	if c.state == CommandComplete {
		return
	}
	c.outputs.PushBack(output)
	c.stream.Value(output)
}

func (c *raftSessionCommand) ack(outputSequenceNum multiraftv1.SequenceNum) {
	elem := c.outputs.Front()
	for elem != nil && elem.Value.(*multiraftv1.SessionCommandOutput).SequenceNum <= outputSequenceNum {
		next := elem.Next()
		c.outputs.Remove(elem)
		elem = next
	}
}

func (c *raftSessionCommand) Close() {
	c.session.commands.remove(c.index)
	c.state = CommandComplete
	for _, watcher := range c.watchers {
		watcher(CommandComplete)
	}
	c.stream.Close()
}

func newSessionOperationCommand(parent *raftSessionCommand) *raftSessionOperationCommand {
	return &raftSessionOperationCommand{
		raftSessionCommand: parent,
	}
}

type raftSessionOperationCommand struct {
	*raftSessionCommand
	outputSeqNum multiraftv1.SequenceNum
}

func (c *raftSessionOperationCommand) Input() *multiraftv1.PrimitiveOperationInput {
	return c.raftSessionCommand.Input().GetOperation()
}

func (c *raftSessionOperationCommand) Output(output *multiraftv1.PrimitiveOperationOutput, err error) {
	if err == nil {
		c.outputSeqNum++
		c.raftSessionCommand.Output(&multiraftv1.SessionCommandOutput{
			SequenceNum: c.outputSeqNum,
			Output: &multiraftv1.SessionCommandOutput_Operation{
				Operation: output,
			},
		}, err)
	} else {
		c.raftSessionCommand.Output(nil, err)
	}
}

func newSessionCreatePrimitiveCommand(parent *raftSessionCommand) *raftSessionCreatePrimitiveCommand {
	return &raftSessionCreatePrimitiveCommand{
		raftSessionCommand: parent,
	}
}

type raftSessionCreatePrimitiveCommand struct {
	*raftSessionCommand
	outputSeqNum multiraftv1.SequenceNum
}

func (c *raftSessionCreatePrimitiveCommand) Input() *multiraftv1.CreatePrimitiveInput {
	return c.raftSessionCommand.Input().GetCreatePrimitive()
}

func (c *raftSessionCreatePrimitiveCommand) Output(output *multiraftv1.CreatePrimitiveOutput, err error) {
	if err == nil {
		c.outputSeqNum++
		c.raftSessionCommand.Output(&multiraftv1.SessionCommandOutput{
			SequenceNum: c.outputSeqNum,
			Output: &multiraftv1.SessionCommandOutput_CreatePrimitive{
				CreatePrimitive: output,
			},
		}, err)
	} else {
		c.raftSessionCommand.Output(nil, err)
	}
}

func newSessionClosePrimitiveCommand(parent *raftSessionCommand) *raftSessionClosePrimitiveCommand {
	return &raftSessionClosePrimitiveCommand{
		raftSessionCommand: parent,
	}
}

type raftSessionClosePrimitiveCommand struct {
	*raftSessionCommand
	outputSeqNum multiraftv1.SequenceNum
}

func (c *raftSessionClosePrimitiveCommand) Input() *multiraftv1.ClosePrimitiveInput {
	return c.raftSessionCommand.Input().GetClosePrimitive()
}

func (c *raftSessionClosePrimitiveCommand) Output(output *multiraftv1.ClosePrimitiveOutput, err error) {
	if err == nil {
		c.outputSeqNum++
		c.raftSessionCommand.Output(&multiraftv1.SessionCommandOutput{
			SequenceNum: c.outputSeqNum,
			Output: &multiraftv1.SessionCommandOutput_ClosePrimitive{
				ClosePrimitive: output,
			},
		}, err)
	} else {
		c.raftSessionCommand.Output(nil, err)
	}
}

func newSessionQuery(session *raftSession) *raftSessionQuery {
	return &raftSessionQuery{
		session: session,
	}
}

type raftSessionQuery struct {
	session *raftSession
	input   *multiraftv1.SessionQueryInput
	stream  streams.WriteStream[*multiraftv1.SessionQueryOutput]
}

func (q *raftSessionQuery) Session() Session {
	return q.session
}

func (q *raftSessionQuery) Operation() PrimitiveQuery {
	return newSessionOperationQuery(q)
}

func (q *raftSessionQuery) Input() *multiraftv1.SessionQueryInput {
	return q.input
}

func (q *raftSessionQuery) Output(output *multiraftv1.SessionQueryOutput, err error) {
	q.stream.Result(output, err)
}

func (q *raftSessionQuery) execute(input *multiraftv1.SessionQueryInput, stream streams.WriteStream[*multiraftv1.SessionQueryOutput]) {
	q.input = input
	q.stream = stream
	q.session.manager.primitives.Query(q.Operation())
}

func (q *raftSessionQuery) Close() {
	q.stream.Close()
}

func newSessionOperationQuery(parent *raftSessionQuery) *raftSessionOperationQuery {
	return &raftSessionOperationQuery{
		raftSessionQuery: parent,
	}
}

type raftSessionOperationQuery struct {
	*raftSessionQuery
}

func (q *raftSessionOperationQuery) Input() *multiraftv1.PrimitiveOperationInput {
	return q.raftSessionQuery.Input().GetOperation()
}

func (q *raftSessionOperationQuery) Output(output *multiraftv1.PrimitiveOperationOutput, err error) {
	if err == nil {
		q.raftSessionQuery.Output(&multiraftv1.SessionQueryOutput{
			Output: &multiraftv1.SessionQueryOutput_Operation{
				Operation: output,
			},
		}, err)
	} else {
		q.raftSessionQuery.Output(nil, err)
	}
}

func newSessionWatcher(f func()) Watcher {
	return &raftSessionWatcher{f}
}

type raftSessionWatcher struct {
	f func()
}

func (w *raftSessionWatcher) Cancel() {
	w.f()
}

var _ Watcher = (*raftSessionWatcher)(nil)
