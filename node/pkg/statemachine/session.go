// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"container/list"
	"context"
	"encoding/binary"
	"encoding/json"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/google/uuid"
	"sync"
	"time"
)

func newSessionManager(registry *PrimitiveTypeRegistry, context *stateManager) *sessionManager {
	sessionManager := &sessionManager{
		context:  context,
		sessions: make(map[multiraftv1.SessionID]*raftSession),
	}
	sessionManager.primitives = newPrimitiveManager(registry, sessionManager)
	return sessionManager
}

type sessionManager struct {
	context    *stateManager
	sessions   map[multiraftv1.SessionID]*raftSession
	primitives *primitiveManager
	prevTime   time.Time
}

func (m *sessionManager) snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.sessions)); err != nil {
		return err
	}
	for _, session := range m.sessions {
		if err := session.snapshot(writer); err != nil {
			return err
		}
	}
	return m.primitives.snapshot(writer)
}

func (m *sessionManager) recover(reader *snapshot.Reader) error {
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
	return m.primitives.recover(reader)
}

func (m *sessionManager) openSession(input *multiraftv1.OpenSessionInput, stream streams.WriteStream[*multiraftv1.OpenSessionOutput]) {
	session := newSession(m)
	session.open(input, stream)
	m.prevTime = m.context.time
}

func (m *sessionManager) keepAlive(input *multiraftv1.KeepAliveInput, stream streams.WriteStream[*multiraftv1.KeepAliveOutput]) {
	session, ok := m.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	session.keepAlive(input, stream)

	// Compute the minimum session timeout
	var minSessionTimeout time.Duration
	for _, session := range m.sessions {
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
	for _, session := range m.sessions {
		if m.context.time.After(maxExpireTime) {
			session.resetTime(m.context.time)
		}
		if m.context.time.After(session.expireTime()) {
			log.Warnf("Session %d expired after %s", session.sessionID, m.context.time.Sub(session.lastUpdated))
			session.expire()
		}
	}
	m.prevTime = m.context.time
}

func (m *sessionManager) closeSession(input *multiraftv1.CloseSessionInput, stream streams.WriteStream[*multiraftv1.CloseSessionOutput]) {
	session, ok := m.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	session.close(input, stream)
	m.prevTime = m.context.time
}

func (m *sessionManager) updateSession(input *multiraftv1.SessionCommandInput, stream streams.WriteStream[*multiraftv1.SessionCommandOutput]) {
	session, ok := m.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	command := newSessionCommand(session)
	command.execute(input, stream)
	m.prevTime = m.context.time
}

func (m *sessionManager) readSession(ctx context.Context, input *multiraftv1.SessionQueryInput, stream streams.WriteStream[*multiraftv1.SessionQueryOutput]) {
	session, ok := m.sessions[input.SessionID]
	if !ok {
		stream.Error(errors.NewFault("session not found"))
		stream.Close()
		return
	}
	query := newSessionQuery(session)
	query.execute(ctx, input, stream)
}

func newSession(manager *sessionManager) *raftSession {
	return &raftSession{
		manager:  manager,
		commands: make(map[multiraftv1.Index]*raftSessionCommand),
		closers:  make(map[multiraftv1.PrimitiveID]func()),
	}
}

type raftSession struct {
	manager     *sessionManager
	commands    map[multiraftv1.Index]*raftSessionCommand
	sessionID   multiraftv1.SessionID
	timeout     time.Duration
	reset       bool
	lastUpdated time.Time
	state       multiraftv1.SessionSnapshot_State
	closers     map[multiraftv1.PrimitiveID]func()
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
	s.sessionID = multiraftv1.SessionID(s.manager.context.index)
	log.Infow("Opened session",
		logging.Uint64("Session", uint64(s.sessionID)),
		logging.Duration("Timeout", input.Timeout))
	s.lastUpdated = s.manager.context.time
	s.timeout = input.Timeout
	s.state = multiraftv1.SessionSnapshot_OPEN
	s.manager.sessions[s.sessionID] = s
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

	log.Debugw("Processing keep-alive",
		logging.Uint64("Session", uint64(s.sessionID)))
	for _, command := range s.commands {
		if input.LastInputSequenceNum < command.input.SequenceNum {
			continue
		}
		sequenceNumBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceNumBytes, uint64(command.input.SequenceNum))
		if !openInputs.Test(sequenceNumBytes) {
			command.cleanup()
		} else {
			if outputSequenceNum, ok := input.LastOutputSequenceNums[command.input.SequenceNum]; ok {
				command.ack(outputSequenceNum)
			}
		}
	}

	stream.Value(&multiraftv1.KeepAliveOutput{})
	stream.Close()

	s.lastUpdated = s.manager.context.time
	s.reset = false
}

func (s *raftSession) close(input *multiraftv1.CloseSessionInput, stream streams.WriteStream[*multiraftv1.CloseSessionOutput]) {
	log.Infow("Closing session",
		logging.Uint64("Session", uint64(input.SessionID)))
	delete(s.manager.sessions, s.sessionID)
	s.state = multiraftv1.SessionSnapshot_CLOSED
	for _, closer := range s.closers {
		closer()
	}
	stream.Value(&multiraftv1.CloseSessionOutput{})
	stream.Close()
}

func (s *raftSession) expire() {
	log.Infow("Session expired",
		logging.Uint64("Session", uint64(s.sessionID)),
		logging.Duration("Timeout", s.timeout))
	delete(s.manager.sessions, s.sessionID)
	s.state = multiraftv1.SessionSnapshot_CLOSED
	for _, closer := range s.closers {
		closer()
	}
}

func (s *raftSession) snapshot(writer *snapshot.Writer) error {
	log.Debugw("Persisting session to snapshot",
		logging.Uint64("Session", uint64(s.sessionID)))
	snapshot := &multiraftv1.SessionSnapshot{
		SessionID:   s.sessionID,
		Timeout:     s.timeout,
		LastUpdated: s.lastUpdated,
	}
	if err := writer.WriteMessage(snapshot); err != nil {
		return err
	}
	if err := writer.WriteVarInt(len(s.commands)); err != nil {
		return err
	}
	for _, command := range s.commands {
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
	log.Infow("Recovering session from snapshot",
		logging.Uint64("Session", uint64(snapshot.SessionID)))
	s.sessionID = snapshot.SessionID
	s.timeout = snapshot.Timeout
	s.lastUpdated = snapshot.LastUpdated
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
	s.state = snapshot.State
	s.manager.sessions[s.sessionID] = s
	return nil
}

func newSessionCommand(session *raftSession) *raftSessionCommand {
	return &raftSessionCommand{
		session: session,
	}
}

type raftSessionCommand struct {
	session      *raftSession
	index        multiraftv1.Index
	state        multiraftv1.CommandSnapshot_State
	input        *multiraftv1.SessionCommandInput
	outputs      *list.List
	outputSeqNum multiraftv1.SequenceNum
	stream       streams.WriteStream[*multiraftv1.SessionCommandOutput]
	closer       func()
}

func (c *raftSessionCommand) Operation() *raftSessionOperationCommand {
	return newSessionOperationCommand(c)
}

func (c *raftSessionCommand) CreatePrimitive() *raftSessionCreatePrimitiveCommand {
	return newSessionCreatePrimitiveCommand(c)
}

func (c *raftSessionCommand) ClosePrimitive() *raftSessionClosePrimitiveCommand {
	return newSessionClosePrimitiveCommand(c)
}

func (c *raftSessionCommand) Input() *multiraftv1.SessionCommandInput {
	return c.input
}

func (c *raftSessionCommand) execute(input *multiraftv1.SessionCommandInput, stream streams.WriteStream[*multiraftv1.SessionCommandOutput]) {
	log.Debugw("Executing command",
		logging.Uint64("Session", uint64(c.session.sessionID)),
		logging.Uint64("Command", uint64(c.input.SequenceNum)))
	c.stream = stream
	switch c.state {
	case multiraftv1.CommandSnapshot_PENDING:
		c.open(input)
		switch input.Input.(type) {
		case *multiraftv1.SessionCommandInput_Operation:
			c.session.manager.primitives.update(c.Operation())
		case *multiraftv1.SessionCommandInput_CreatePrimitive:
			c.session.manager.primitives.create(c.CreatePrimitive())
		case *multiraftv1.SessionCommandInput_ClosePrimitive:
			c.session.manager.primitives.close(c.ClosePrimitive())
		}
	default:
		c.replay()
	}
}

func (c *raftSessionCommand) open(input *multiraftv1.SessionCommandInput) {
	c.index = c.session.manager.context.index
	c.input = input
	c.outputs = list.New()
	c.session.commands[c.index] = c
	c.state = multiraftv1.CommandSnapshot_RUNNING
}

func (c *raftSessionCommand) replay() {
	if c.outputs.Len() > 0 {
		log.Debugw("Replaying command responses",
			logging.Uint64("Session", uint64(c.session.sessionID)),
			logging.Uint64("Command", uint64(c.input.SequenceNum)))
		elem := c.outputs.Front()
		for elem != nil {
			output := elem.Value.(*multiraftv1.SessionCommandOutput)
			c.stream.Value(output)
			elem = elem.Next()
		}
	}
	if c.state == multiraftv1.CommandSnapshot_COMPLETE {
		c.stream.Close()
	}
}

func (c *raftSessionCommand) snapshot(writer *snapshot.Writer) error {
	log.Infow("Persisting command to snapshot",
		logging.Uint64("Session", uint64(c.session.sessionID)),
		logging.Uint64("Command", uint64(c.input.SequenceNum)))
	pendingOutputs := make([]*multiraftv1.SessionCommandOutput, 0, c.outputs.Len())
	elem := c.outputs.Front()
	for elem != nil {
		pendingOutputs = append(pendingOutputs, elem.Value.(*multiraftv1.SessionCommandOutput))
		elem = elem.Next()
	}
	snapshot := &multiraftv1.CommandSnapshot{
		Index:                 c.index,
		State:                 c.state,
		Input:                 c.input,
		PendingOutputs:        pendingOutputs,
		LastOutputSequenceNum: c.outputSeqNum,
	}
	return writer.WriteMessage(snapshot)
}

func (c *raftSessionCommand) recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.CommandSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	log.Infow("Recovering command from snapshot",
		logging.Uint64("Session", uint64(c.session.sessionID)),
		logging.Uint64("Command", uint64(c.input.SequenceNum)))
	c.index = snapshot.Index
	c.input = snapshot.Input
	c.outputs = list.New()
	for _, output := range snapshot.PendingOutputs {
		r := output
		c.outputs.PushBack(r)
	}
	c.outputSeqNum = snapshot.LastOutputSequenceNum
	c.stream = streams.NewNilStream[*multiraftv1.SessionCommandOutput]()
	c.state = snapshot.State
	c.session.commands[c.index] = c
	return nil
}

func (c *raftSessionCommand) nextSequenceNum() multiraftv1.SequenceNum {
	c.outputSeqNum++
	return c.outputSeqNum
}

func (c *raftSessionCommand) Output(output *multiraftv1.SessionCommandOutput) {
	if c.state == multiraftv1.CommandSnapshot_COMPLETE {
		return
	}
	log.Debugw("Cached command output",
		logging.Uint64("Session", uint64(c.session.sessionID)),
		logging.Uint64("Command", uint64(c.input.SequenceNum)),
		logging.Uint64("SequenceNum", uint64(output.SequenceNum)))
	c.outputs.PushBack(output)
	c.stream.Value(output)
}

func (c *raftSessionCommand) Error(err error) {
	if c.state == multiraftv1.CommandSnapshot_COMPLETE {
		return
	}
	c.Output(&multiraftv1.SessionCommandOutput{
		SequenceNum: c.nextSequenceNum(),
		Failure:     getFailure(err),
	})
}

func (c *raftSessionCommand) ack(outputSequenceNum multiraftv1.SequenceNum) {
	log.Debugw("Acked command output",
		logging.Uint64("Session", uint64(c.session.sessionID)),
		logging.Uint64("Command", uint64(c.input.SequenceNum)),
		logging.Uint64("SequenceNum", uint64(outputSequenceNum)))
	elem := c.outputs.Front()
	for elem != nil && elem.Value.(*multiraftv1.SessionCommandOutput).SequenceNum <= outputSequenceNum {
		next := elem.Next()
		c.outputs.Remove(elem)
		elem = next
	}
}

func (c *raftSessionCommand) cleanup() {
	switch c.state {
	case multiraftv1.CommandSnapshot_RUNNING:
		log.Debugw("Canceled command",
			logging.Uint64("Session", uint64(c.session.sessionID)),
			logging.Uint64("Command", uint64(c.input.SequenceNum)))
		c.close()
	}
	delete(c.session.commands, c.index)
}

func (c *raftSessionCommand) close() {
	c.state = multiraftv1.CommandSnapshot_COMPLETE
	if c.closer != nil {
		c.closer()
	}
	c.stream.Close()
}

func (c *raftSessionCommand) Close() {
	log.Debugw("Closed command",
		logging.Uint64("Session", uint64(c.session.sessionID)),
		logging.Uint64("Command", uint64(c.input.SequenceNum)))
	c.close()
}

func newSessionOperationCommand(parent *raftSessionCommand) *raftSessionOperationCommand {
	return &raftSessionOperationCommand{
		raftSessionCommand: parent,
	}
}

type raftSessionOperationCommand struct {
	*raftSessionCommand
}

func (c *raftSessionOperationCommand) Input() *multiraftv1.PrimitiveOperationInput {
	return c.raftSessionCommand.Input().GetOperation()
}

func (c *raftSessionOperationCommand) Output(output *multiraftv1.PrimitiveOperationOutput) {
	c.raftSessionCommand.Output(&multiraftv1.SessionCommandOutput{
		SequenceNum: c.nextSequenceNum(),
		Output: &multiraftv1.SessionCommandOutput_Operation{
			Operation: output,
		},
	})
}

func newSessionCreatePrimitiveCommand(parent *raftSessionCommand) *raftSessionCreatePrimitiveCommand {
	return &raftSessionCreatePrimitiveCommand{
		raftSessionCommand: parent,
	}
}

type raftSessionCreatePrimitiveCommand struct {
	*raftSessionCommand
}

func (c *raftSessionCreatePrimitiveCommand) Input() *multiraftv1.CreatePrimitiveInput {
	return c.raftSessionCommand.Input().GetCreatePrimitive()
}

func (c *raftSessionCreatePrimitiveCommand) Output(output *multiraftv1.CreatePrimitiveOutput) {
	c.raftSessionCommand.Output(&multiraftv1.SessionCommandOutput{
		SequenceNum: c.nextSequenceNum(),
		Output: &multiraftv1.SessionCommandOutput_CreatePrimitive{
			CreatePrimitive: output,
		},
	})
}

func newSessionClosePrimitiveCommand(parent *raftSessionCommand) *raftSessionClosePrimitiveCommand {
	return &raftSessionClosePrimitiveCommand{
		raftSessionCommand: parent,
	}
}

type raftSessionClosePrimitiveCommand struct {
	*raftSessionCommand
}

func (c *raftSessionClosePrimitiveCommand) Input() *multiraftv1.ClosePrimitiveInput {
	return c.raftSessionCommand.Input().GetClosePrimitive()
}

func (c *raftSessionClosePrimitiveCommand) Output(output *multiraftv1.ClosePrimitiveOutput) {
	c.raftSessionCommand.Output(&multiraftv1.SessionCommandOutput{
		SequenceNum: c.nextSequenceNum(),
		Output: &multiraftv1.SessionCommandOutput_ClosePrimitive{
			ClosePrimitive: output,
		},
	})
}

func newSessionQuery(session *raftSession) *raftSessionQuery {
	return &raftSessionQuery{
		session: session,
		closeCh: make(chan struct{}),
	}
}

type raftSessionQuery struct {
	sequenceNum multiraftv1.SequenceNum
	session     *raftSession
	ctx         context.Context
	input       *multiraftv1.SessionQueryInput
	stream      streams.WriteStream[*multiraftv1.SessionQueryOutput]
	state       OperationState
	watchers    map[string]OperationWatcher
	watchersMu  sync.RWMutex
	closeCh     chan struct{}
}

func (q *raftSessionQuery) ID() QueryID {
	return QueryID(q.sequenceNum)
}

func (q *raftSessionQuery) Watch(f OperationWatcher) CancelFunc {
	q.watchersMu.Lock()
	defer q.watchersMu.Unlock()
	if q.watchers == nil {
		q.watchers = make(map[string]OperationWatcher)
		go func() {
			select {
			case <-q.closeCh:
				return
			case <-q.ctx.Done():
				log.Debugw("Canceled query",
					logging.Uint64("Session", uint64(q.session.sessionID)),
					logging.Uint64("Query", uint64(q.sequenceNum)))
				q.session.manager.context.queriesMu.Lock()
				q.session.manager.context.canceledQueries = append(q.session.manager.context.canceledQueries, q)
				q.session.manager.context.queriesMu.Unlock()
			}
		}()
	}
	id := uuid.New().String()
	q.watchers[id] = f
	return func() {
		delete(q.watchers, id)
	}
}

func (q *raftSessionQuery) Operation() *raftSessionOperationQuery {
	return newSessionOperationQuery(q)
}

func (q *raftSessionQuery) Input() *multiraftv1.SessionQueryInput {
	return q.input
}

func (q *raftSessionQuery) Output(output *multiraftv1.SessionQueryOutput) {
	q.stream.Value(output)
}

func (q *raftSessionQuery) Error(err error) {
	q.Output(&multiraftv1.SessionQueryOutput{
		Failure: getFailure(err),
	})
}

func (q *raftSessionQuery) execute(ctx context.Context, input *multiraftv1.SessionQueryInput, stream streams.WriteStream[*multiraftv1.SessionQueryOutput]) {
	log.Debugw("Executing query",
		logging.Uint64("Session", uint64(q.session.sessionID)),
		logging.Uint64("Query", uint64(q.sequenceNum)))
	q.sequenceNum = multiraftv1.SequenceNum(q.session.manager.context.sequenceNum.Add(1))
	q.ctx = ctx
	q.input = input
	q.stream = stream
	q.session.manager.primitives.read(q.Operation())
}

func (q *raftSessionQuery) Close() {
	log.Debugw("Closed query",
		logging.Uint64("Session", uint64(q.session.sessionID)),
		logging.Uint64("Query", uint64(q.sequenceNum)))
	q.stream.Close()
	close(q.closeCh)
	q.watchersMu.RLock()
	defer q.watchersMu.RUnlock()
	if q.watchers != nil {
		for _, watcher := range q.watchers {
			watcher(Complete)
		}
	}
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

func (q *raftSessionOperationQuery) Output(output *multiraftv1.PrimitiveOperationOutput) {
	q.raftSessionQuery.Output(&multiraftv1.SessionQueryOutput{
		Output: &multiraftv1.SessionQueryOutput_Operation{
			Operation: output,
		},
	})
}
