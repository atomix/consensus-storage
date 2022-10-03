// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"encoding/binary"
	"encoding/json"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/google/uuid"
	"sync"
	"time"
)

type ID uint64

type State int

const (
	Open State = iota
	Closed
)

type CancelFunc = statemachine.CancelFunc

// Session is a service session
type Session interface {
	// Log returns the session log
	Log() logging.Logger
	// ID returns the session identifier
	ID() ID
	// State returns the current session state
	State() State
	// Watch watches the session state for changes
	Watch(watcher func(State)) CancelFunc
}

// Sessions provides access to open sessions
type Sessions interface {
	// Get gets a session by ID
	Get(ID) (Session, bool)
	// List lists all open sessions
	List() []Session
}

func newManagedSession(manager *sessionManager) *managedSession {
	return &managedSession{
		manager:   manager,
		proposals: make(map[multiraftv1.SequenceNum]*sessionProposal),
		queries:   make(map[multiraftv1.SequenceNum]*sessionQuery),
		watchers:  make(map[uuid.UUID]func(State)),
	}
}

type managedSession struct {
	manager          *sessionManager
	proposals        map[multiraftv1.SequenceNum]*sessionProposal
	queries          map[multiraftv1.SequenceNum]*sessionQuery
	queriesMu        sync.Mutex
	log              logging.Logger
	id               ID
	state            State
	watchers         map[uuid.UUID]func(State)
	timeout          time.Duration
	lastUpdated      time.Time
	expireCancelFunc CancelFunc
}

func (s *managedSession) Log() logging.Logger {
	return s.log
}

func (s *managedSession) ID() ID {
	return s.id
}

func (s *managedSession) State() State {
	return s.state
}

func (s *managedSession) Watch(f func(State)) CancelFunc {
	id := uuid.New()
	s.watchers[id] = f
	return func() {
		delete(s.watchers, id)
	}
}

func (s *managedSession) propose(parent statemachine.Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]) {
	if proposal, ok := s.proposals[parent.Input().SequenceNum]; ok {
		proposal.replay(parent)
	} else {
		proposal := newSessionProposal(s)
		s.proposals[parent.Input().SequenceNum] = proposal
		proposal.execute(parent)
	}
}

func (s *managedSession) query(parent statemachine.Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]) {
	query := newSessionQuery(s)
	query.execute(parent)
	if query.state == Running {
		s.queriesMu.Lock()
		s.queries[query.Input().SequenceNum] = query
		s.queriesMu.Unlock()
	}
}

func (s *managedSession) Snapshot(writer *snapshot.Writer) error {
	s.Log().Debug("Persisting session to snapshot")

	var state multiraftv1.SessionSnapshot_State
	switch s.state {
	case Open:
		state = multiraftv1.SessionSnapshot_OPEN
	case Closed:
		state = multiraftv1.SessionSnapshot_CLOSED
	}

	snapshot := &multiraftv1.SessionSnapshot{
		SessionID:   multiraftv1.SessionID(s.id),
		State:       state,
		Timeout:     s.timeout,
		LastUpdated: s.lastUpdated,
	}
	if err := writer.WriteMessage(snapshot); err != nil {
		return err
	}

	if err := writer.WriteVarInt(len(s.proposals)); err != nil {
		return err
	}
	for _, proposal := range s.proposals {
		if err := proposal.snapshot(writer); err != nil {
			return err
		}
	}
	return nil
}

func (s *managedSession) Recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.SessionSnapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}

	s.id = ID(snapshot.SessionID)
	s.timeout = snapshot.Timeout
	s.lastUpdated = snapshot.LastUpdated

	s.log = s.manager.Log().WithFields(logging.Uint64("Session", uint64(s.id)))
	s.Log().Debug("Recovering session from snapshot")

	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		proposal := newSessionProposal(s)
		if err := proposal.recover(reader); err != nil {
			return err
		}
		s.proposals[proposal.input.SequenceNum] = proposal
	}

	switch snapshot.State {
	case multiraftv1.SessionSnapshot_OPEN:
		s.state = Open
	case multiraftv1.SessionSnapshot_CLOSED:
		s.state = Closed
	}
	s.manager.sessions.add(s)
	s.scheduleExpireTimer()
	return nil
}

func (s *managedSession) expire() {
	s.Log().Warnf("Session expired after %s", s.manager.Time().Sub(s.lastUpdated))
	s.destroy()
}

func (s *managedSession) scheduleExpireTimer() {
	if s.expireCancelFunc != nil {
		s.expireCancelFunc()
	}
	expireTime := s.lastUpdated.Add(s.timeout)
	s.expireCancelFunc = s.manager.Scheduler().Schedule(expireTime, s.expire)
	s.Log().Debugw("Scheduled expire time", logging.Time("Expire", expireTime))
}

func (s *managedSession) open(open statemachine.Proposal[*multiraftv1.OpenSessionInput, *multiraftv1.OpenSessionOutput]) {
	defer open.Close()
	s.id = ID(open.ID())
	s.state = Open
	s.lastUpdated = s.manager.Time()
	s.timeout = open.Input().Timeout
	s.log = s.manager.Log().WithFields(logging.Uint64("Session", uint64(s.id)))
	s.manager.sessions.add(s)
	s.scheduleExpireTimer()
	s.Log().Infow("Opened session", logging.Duration("Timeout", s.timeout))
	open.Output(&multiraftv1.OpenSessionOutput{
		SessionID: multiraftv1.SessionID(s.ID()),
	})
}

func (s *managedSession) keepAlive(keepAlive statemachine.Proposal[*multiraftv1.KeepAliveInput, *multiraftv1.KeepAliveOutput]) {
	defer keepAlive.Close()

	openInputs := &bloom.BloomFilter{}
	if err := json.Unmarshal(keepAlive.Input().InputFilter, openInputs); err != nil {
		s.Log().Warn("Failed to decode request filter", err)
		keepAlive.Error(errors.NewInvalid("invalid request filter", err))
		return
	}

	s.Log().Debug("Processing keep-alive")
	for _, proposal := range s.proposals {
		if keepAlive.Input().LastInputSequenceNum < proposal.Input().SequenceNum {
			continue
		}
		sequenceNumBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceNumBytes, uint64(proposal.Input().SequenceNum))
		if !openInputs.Test(sequenceNumBytes) {
			proposal.Cancel()
			delete(s.proposals, proposal.Input().SequenceNum)
		} else {
			if outputSequenceNum, ok := keepAlive.Input().LastOutputSequenceNums[proposal.Input().SequenceNum]; ok {
				proposal.ack(outputSequenceNum)
			}
		}
	}

	s.queriesMu.Lock()
	for sn, query := range s.queries {
		if keepAlive.Input().LastInputSequenceNum < query.Input().SequenceNum {
			continue
		}
		sequenceNumBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceNumBytes, uint64(query.Input().SequenceNum))
		if !openInputs.Test(sequenceNumBytes) {
			query.Cancel()
			delete(s.queries, sn)
		}
	}
	s.queriesMu.Unlock()

	keepAlive.Output(&multiraftv1.KeepAliveOutput{})

	s.lastUpdated = s.manager.Time()
	s.scheduleExpireTimer()
}

func (s *managedSession) close(close statemachine.Proposal[*multiraftv1.CloseSessionInput, *multiraftv1.CloseSessionOutput]) {
	defer close.Close()
	s.destroy()
	close.Output(&multiraftv1.CloseSessionOutput{})
}

func (s *managedSession) destroy() {
	s.manager.sessions.remove(s.id)
	s.expireCancelFunc()
	s.state = Closed
	for _, proposal := range s.proposals {
		proposal.Cancel()
	}
	s.queriesMu.Lock()
	for _, query := range s.queries {
		query.Cancel()
	}
	s.queriesMu.Unlock()
	for _, watcher := range s.watchers {
		watcher(Closed)
	}
}

func newManagedSessions() *managedSessions {
	return &managedSessions{
		sessions: make(map[ID]*managedSession),
	}
}

type managedSessions struct {
	sessions map[ID]*managedSession
}

func (s *managedSessions) Get(id ID) (Session, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *managedSessions) List() []Session {
	sessions := make([]Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *managedSessions) add(session *managedSession) bool {
	if _, ok := s.sessions[session.ID()]; !ok {
		s.sessions[session.ID()] = session
		return true
	}
	return false
}

func (s *managedSessions) remove(id ID) bool {
	if _, ok := s.sessions[id]; ok {
		delete(s.sessions, id)
		return true
	}
	return false
}

func (s *managedSessions) get(id ID) (*managedSession, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *managedSessions) list() []*managedSession {
	sessions := make([]*managedSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
