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
	"github.com/google/uuid"
)

func newPrimitive(parent session.Context, id ID, spec multiraftv1.PrimitiveSpec, registry *TypeRegistry) (*primitiveContext, bool) {
	factory, ok := registry.lookup(spec.Service)
	if !ok {
		return nil, false
	}
	primitive := &primitiveContext{
		Context:  parent,
		id:       id,
		spec:     spec,
		sessions: newPrimitiveSessions(),
		log: parent.Log().WithFields(
			logging.String("Service", spec.Service),
			logging.Uint64("Primitive", uint64(id)),
			logging.String("Namespace", spec.Namespace),
			logging.String("Name", spec.Name)),
	}
	primitive.sm = factory(primitive)
	return primitive, true
}

type primitiveContext struct {
	session.Context
	id       ID
	spec     multiraftv1.PrimitiveSpec
	sessions *primitiveSessions
	log      logging.Logger
	sm       primitiveDelegate
}

func (p *primitiveContext) Log() logging.Logger {
	return p.log
}

func (p *primitiveContext) ID() ID {
	return p.id
}

func (p *primitiveContext) Service() string {
	return p.spec.Service
}

func (p *primitiveContext) Namespace() string {
	return p.spec.Namespace
}

func (p *primitiveContext) Name() string {
	return p.spec.Name
}

func (p *primitiveContext) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(p.sessions.sessions)); err != nil {
		return err
	}
	for _, session := range p.sessions.list() {
		if err := writer.WriteVarUint64(uint64(session.ID())); err != nil {
			return err
		}
	}
	return p.sm.Snapshot(writer)
}

func (p *primitiveContext) Recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		sessionID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}
		parent, ok := p.Context.Sessions().Get(session.ID(sessionID))
		if !ok {
			return errors.NewFault("session %d not found", sessionID)
		}
		p.sessions.add(newPrimitiveSession(p, parent))
	}
	return p.sm.Recover(reader)
}

func (p *primitiveContext) Sessions() session.Sessions {
	return p.sessions
}

func (p *primitiveContext) Proposals() session.Proposals {
	return newPrimitiveProposals(p.id, p.Context.Proposals(), p.sessions)
}

func (p *primitiveContext) open(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {
	p.sessions.add(newPrimitiveSession(p, proposal.Session()))
	proposal.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(p.ID()),
	})
	proposal.Close()
}

func (p *primitiveContext) close(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {
	p.sessions.remove(proposal.Session().ID())
	proposal.Output(&multiraftv1.ClosePrimitiveOutput{})
	proposal.Close()
}

func (p *primitiveContext) propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	session, ok := p.sessions.get(proposal.Session().ID())
	if !ok {
		proposal.Error(errors.NewForbidden("session not found"))
		proposal.Close()
	} else {
		p.sm.propose(newPrimitiveProposal(session, proposal))
	}
}

func (p *primitiveContext) query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	session, ok := p.sessions.get(query.Session().ID())
	if !ok {
		query.Error(errors.NewForbidden("session not found"))
		query.Close()
	} else {
		p.sm.query(newPrimitiveQuery(session, query))
	}
}

func newPrimitiveSession(
	context *primitiveContext,
	parent session.Session) *primitiveSession {
	s := &primitiveSession{
		primitive: context,
		parent:    parent,
		state:     session.Open,
		watchers:  make(map[uuid.UUID]statemachine.WatchFunc[session.State]),
		log: parent.Log().WithFields(
			logging.String("Service", context.spec.Service),
			logging.Uint64("Primitive", uint64(context.id)),
			logging.String("Namespace", context.spec.Namespace),
			logging.String("Name", context.spec.Name)),
	}
	s.cancel = parent.Watch(func(state session.State) {
		if state == session.Closed {
			s.close()
		}
	})
	return s
}

type primitiveSession struct {
	primitive *primitiveContext
	parent    session.Session
	state     session.State
	watchers  map[uuid.UUID]statemachine.WatchFunc[session.State]
	cancel    statemachine.CancelFunc
	log       logging.Logger
}

func (s *primitiveSession) Log() logging.Logger {
	return s.log
}

func (s *primitiveSession) ID() session.ID {
	return s.parent.ID()
}

func (s *primitiveSession) State() session.State {
	return s.state
}

func (s *primitiveSession) Watch(watcher statemachine.WatchFunc[session.State]) statemachine.CancelFunc {
	id := uuid.New()
	s.watchers[id] = watcher
	return func() {
		delete(s.watchers, id)
	}
}

func (s *primitiveSession) Proposals() session.Proposals {
	return newPrimitiveSessionProposals(s)
}

func (s *primitiveSession) close() {
	s.cancel()
	s.primitive.sessions.remove(s.ID())
	s.state = session.Closed
	for _, watcher := range s.watchers {
		watcher(session.Closed)
	}
}

var _ session.Session = (*primitiveSession)(nil)

func newPrimitiveSessions() *primitiveSessions {
	return &primitiveSessions{
		sessions: make(map[session.ID]*primitiveSession),
	}
}

type primitiveSessions struct {
	sessions map[session.ID]*primitiveSession
}

func (s *primitiveSessions) Get(id session.ID) (session.Session, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *primitiveSessions) List() []session.Session {
	sessions := make([]session.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *primitiveSessions) add(session *primitiveSession) {
	s.sessions[session.ID()] = session
}

func (s *primitiveSessions) remove(sessionID session.ID) bool {
	if _, ok := s.sessions[sessionID]; ok {
		delete(s.sessions, sessionID)
		return true
	}
	return false
}

func (s *primitiveSessions) get(id session.ID) (*primitiveSession, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *primitiveSessions) list() []*primitiveSession {
	sessions := make([]*primitiveSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func newPrimitiveQuery(
	session *primitiveSession,
	parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) *primitiveQuery {
	return &primitiveQuery{
		Query:   parent,
		session: session,
	}
}

type primitiveQuery struct {
	session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]
	session *primitiveSession
}

func (q *primitiveQuery) Session() session.Session {
	return q.session
}

var _ session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput] = (*primitiveQuery)(nil)

func newPrimitiveProposal(
	session *primitiveSession,
	parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) *primitiveProposal {
	return &primitiveProposal{
		Proposal: parent,
		session:  session,
	}
}

type primitiveProposal struct {
	session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
	session *primitiveSession
}

func (p *primitiveProposal) Session() session.Session {
	return p.session
}

var _ session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] = (*primitiveProposal)(nil)

func newPrimitiveProposals(primitiveID ID, proposals session.Proposals, sessions *primitiveSessions) *primitiveProposals {
	return &primitiveProposals{
		primitiveID: primitiveID,
		proposals:   proposals,
		sessions:    sessions,
	}
}

type primitiveProposals struct {
	primitiveID ID
	proposals   session.Proposals
	sessions    *primitiveSessions
}

func (p *primitiveProposals) Get(id statemachine.ProposalID) (session.PrimitiveProposal, bool) {
	parent, ok := p.proposals.Get(id)
	if !ok {
		return nil, false
	}
	if ID(parent.Input().PrimitiveID) != p.primitiveID {
		return nil, false
	}
	session, ok := p.sessions.get(parent.Session().ID())
	if !ok {
		return nil, false
	}
	return newPrimitiveProposal(session, parent), true
}

func (p *primitiveProposals) List() []session.PrimitiveProposal {
	parents := p.proposals.List()
	proposals := make([]session.PrimitiveProposal, 0, len(parents))
	for _, parent := range parents {
		if ID(parent.Input().PrimitiveID) != p.primitiveID {
			continue
		}
		session, ok := p.sessions.get(parent.Session().ID())
		if !ok {
			continue
		}
		proposals = append(proposals, newPrimitiveProposal(session, parent))
	}
	return proposals
}

func newPrimitiveSessionProposals(session *primitiveSession) *primitiveSessionProposals {
	return &primitiveSessionProposals{
		session: session,
	}
}

type primitiveSessionProposals struct {
	session *primitiveSession
}

func (p *primitiveSessionProposals) Get(id statemachine.ProposalID) (session.PrimitiveProposal, bool) {
	parent, ok := p.session.parent.Proposals().Get(id)
	if !ok {
		return nil, false
	}
	if ID(parent.Input().PrimitiveID) != p.session.primitive.id {
		return nil, false
	}
	return newPrimitiveProposal(p.session, parent), true
}

func (p *primitiveSessionProposals) List() []session.PrimitiveProposal {
	parents := p.session.parent.Proposals().List()
	proposals := make([]session.PrimitiveProposal, 0, len(parents))
	for _, parent := range parents {
		if ID(parent.Input().PrimitiveID) != p.session.primitive.id {
			continue
		}
		proposals = append(proposals, newPrimitiveProposal(p.session, parent))
	}
	return proposals
}
