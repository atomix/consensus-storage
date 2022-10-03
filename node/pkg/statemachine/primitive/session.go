// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/google/uuid"
	"sync"
)

func newPrimitiveSession[I, O any](primitive *primitiveExecutor[I, O]) *primitiveSession[I, O] {
	s := &primitiveSession[I, O]{
		primitive: primitive,
		proposals: make(map[ProposalID]*primitiveProposal[I, O]),
		queries:   make(map[QueryID]*primitiveQuery[I, O]),
		watchers:  make(map[uuid.UUID]func(SessionState)),
	}
	return s
}

type primitiveSession[I, O any] struct {
	primitive *primitiveExecutor[I, O]
	parent    session.Session
	proposals map[ProposalID]*primitiveProposal[I, O]
	queries   map[QueryID]*primitiveQuery[I, O]
	queriesMu sync.Mutex
	state     SessionState
	watchers  map[uuid.UUID]func(SessionState)
	cancel    session.CancelFunc
	log       logging.Logger
}

func (s *primitiveSession[I, O]) Log() logging.Logger {
	return s.log
}

func (s *primitiveSession[I, O]) ID() SessionID {
	return SessionID(s.parent.ID())
}

func (s *primitiveSession[I, O]) State() SessionState {
	return s.state
}

func (s *primitiveSession[I, O]) Watch(watcher func(SessionState)) CancelFunc {
	id := uuid.New()
	s.watchers[id] = watcher
	return func() {
		delete(s.watchers, id)
	}
}

func (s *primitiveSession[I, O]) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarUint64(uint64(s.ID())); err != nil {
		return err
	}
	return nil
}

func (s *primitiveSession[I, O]) Recover(reader *snapshot.Reader) error {
	sessionID, err := reader.ReadVarUint64()
	if err != nil {
		return err
	}
	parent, ok := s.primitive.Context.Sessions().Get(session.ID(sessionID))
	if !ok {
		return errors.NewFault("session not found")
	}
	s.init(parent)
	return nil
}

func (s *primitiveSession[I, O]) registerProposal(proposal *primitiveProposal[I, O]) {
	s.proposals[proposal.ID()] = proposal
}

func (s *primitiveSession[I, O]) unregisterProposal(proposalID ProposalID) {
	delete(s.proposals, proposalID)
}

func (s *primitiveSession[I, O]) registerQuery(query *primitiveQuery[I, O]) {
	s.queriesMu.Lock()
	s.queries[query.ID()] = query
	s.queriesMu.Unlock()
}

func (s *primitiveSession[I, O]) unregisterQuery(queryID QueryID) {
	s.queriesMu.Lock()
	delete(s.queries, queryID)
	s.queriesMu.Unlock()
}

func (s *primitiveSession[I, O]) init(parent session.Session) {
	s.parent = parent
	s.state = SessionOpen
	s.log = s.primitive.Log().WithFields(logging.Uint64("SessionID", uint64(parent.ID())))
	s.cancel = parent.Watch(func(state session.State) {
		if state == session.Closed {
			s.destroy()
		}
	})
	s.primitive.sessions.add(s)
}

func (s *primitiveSession[I, O]) open(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {
	s.init(proposal.Session())
	proposal.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(s.ID()),
	})
	proposal.Close()
}

func (s *primitiveSession[I, O]) propose(parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	proposal := newPrimitiveProposal[I, O](s)
	proposal.execute(parent)
}

func (s *primitiveSession[I, O]) query(parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	query := newPrimitiveQuery[I, O](s)
	query.execute(parent)
}

func (s *primitiveSession[I, O]) close(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {
	s.destroy()
	proposal.Output(&multiraftv1.ClosePrimitiveOutput{})
	proposal.Close()
}

func (s *primitiveSession[I, O]) destroy() {
	s.cancel()
	s.primitive.sessions.remove(s.ID())
	s.state = SessionClosed
	for _, proposal := range s.proposals {
		proposal.Cancel()
	}
	for _, query := range s.queries {
		query.Cancel()
	}
	for _, watcher := range s.watchers {
		watcher(SessionClosed)
	}
}

var _ Session = (*primitiveSession[any, any])(nil)

func newPrimitiveSessions[I, O any]() *primitiveSessions[I, O] {
	return &primitiveSessions[I, O]{
		sessions: make(map[SessionID]*primitiveSession[I, O]),
	}
}

type primitiveSessions[I, O any] struct {
	sessions map[SessionID]*primitiveSession[I, O]
}

func (s *primitiveSessions[I, O]) Get(id SessionID) (Session, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *primitiveSessions[I, O]) List() []Session {
	sessions := make([]Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *primitiveSessions[I, O]) add(session *primitiveSession[I, O]) {
	s.sessions[session.ID()] = session
}

func (s *primitiveSessions[I, O]) remove(sessionID SessionID) bool {
	if _, ok := s.sessions[sessionID]; ok {
		delete(s.sessions, sessionID)
		return true
	}
	return false
}

func (s *primitiveSessions[I, O]) get(id SessionID) (*primitiveSession[I, O], bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *primitiveSessions[I, O]) list() []*primitiveSession[I, O] {
	sessions := make([]*primitiveSession[I, O], 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
