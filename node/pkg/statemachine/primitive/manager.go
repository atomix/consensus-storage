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

var log = logging.GetLogger()

func NewManager(ctx session.Context, registry *TypeRegistry) session.PrimitiveManager {
	return &primitiveManager{
		Context:    ctx,
		registry:   registry,
		primitives: make(map[ID]*managedPrimitive),
	}
}

type primitiveManager struct {
	session.Context
	registry   *TypeRegistry
	primitives map[ID]*managedPrimitive
}

func (m *primitiveManager) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.primitives)); err != nil {
		return err
	}
	for _, primitive := range m.primitives {
		snapshot := &multiraftv1.PrimitiveSnapshot{
			PrimitiveID: multiraftv1.PrimitiveID(primitive.ID()),
			Spec:        primitive.spec,
		}
		if err := writer.WriteMessage(snapshot); err != nil {
			return err
		}
		if err := primitive.Snapshot(writer); err != nil {
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
		snapshot := &multiraftv1.PrimitiveSnapshot{}
		if err := reader.ReadMessage(snapshot); err != nil {
			return err
		}
		primitive, ok := newManagedPrimitive(m.Context, ID(snapshot.PrimitiveID), snapshot.Spec, m.registry)
		if !ok {
			return errors.NewFault("primitive type not found")
		}
		m.primitives[primitive.ID()] = primitive
		if err := primitive.Recover(reader); err != nil {
			return err
		}
	}
	return nil
}

func (m *primitiveManager) Propose(proposal session.PrimitiveProposal) {
	primitive, ok := m.primitives[ID(proposal.Input().PrimitiveID)]
	if !ok {
		proposal.Error(errors.NewForbidden("primitive %d not found", proposal.Input().PrimitiveID))
		proposal.Close()
	} else {
		primitive.propose(proposal)
	}
}

func (m *primitiveManager) CreatePrimitive(proposal session.CreatePrimitiveProposal) {
	var primitive *managedPrimitive
	for _, p := range m.primitives {
		if p.spec.Namespace == proposal.Input().Namespace &&
			p.spec.Name == proposal.Input().Name {
			if p.spec.Service != proposal.Input().Service {
				proposal.Error(errors.NewForbidden("cannot create primitive of a different type with the same name"))
				proposal.Close()
				return
			}
			primitive = p
			break
		}
	}

	if primitive == nil {
		if p, ok := newManagedPrimitive(m.Context, ID(proposal.ID()), proposal.Input().PrimitiveSpec, m.registry); !ok {
			proposal.Error(errors.NewForbidden("unknown primitive type"))
			proposal.Close()
			return
		} else {
			m.primitives[p.ID()] = p
			primitive = p
		}
	}

	primitive.open(proposal)
}

func (m *primitiveManager) ClosePrimitive(proposal session.ClosePrimitiveProposal) {
	primitive, ok := m.primitives[ID(proposal.Input().PrimitiveID)]
	if !ok {
		proposal.Error(errors.NewForbidden("primitive %d not found", proposal.Input().PrimitiveID))
		proposal.Close()
	} else {
		primitive.close(proposal)
	}
}

func (m *primitiveManager) Query(query session.PrimitiveQuery) {
	primitive, ok := m.primitives[ID(query.Input().PrimitiveID)]
	if !ok {
		query.Error(errors.NewForbidden("primitive %d not found", query.Input().PrimitiveID))
		query.Close()
	} else {
		primitive.query(query)
	}
}

func newManagedPrimitive(parent session.Context, id ID, spec multiraftv1.PrimitiveSpec, registry *TypeRegistry) (*managedPrimitive, bool) {
	factory, ok := registry.lookup(spec.Service)
	if !ok {
		return nil, false
	}
	primitive := &managedPrimitive{
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

type managedPrimitive struct {
	session.Context
	id       ID
	spec     multiraftv1.PrimitiveSpec
	sessions *primitiveSessions
	log      logging.Logger
	sm       primitiveDelegate
}

func (p *managedPrimitive) Log() logging.Logger {
	return p.log
}

func (p *managedPrimitive) ID() ID {
	return p.id
}

func (p *managedPrimitive) Service() string {
	return p.spec.Service
}

func (p *managedPrimitive) Namespace() string {
	return p.spec.Namespace
}

func (p *managedPrimitive) Name() string {
	return p.spec.Name
}

func (p *managedPrimitive) Snapshot(writer *snapshot.Writer) error {
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

func (p *managedPrimitive) Recover(reader *snapshot.Reader) error {
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

func (p *managedPrimitive) Sessions() session.Sessions {
	return p.sessions
}

func (p *managedPrimitive) Proposals() session.Proposals {
	return newPrimitiveProposals(p.id, p.Context.Proposals(), p.sessions)
}

func (p *managedPrimitive) open(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {
	p.sessions.add(newPrimitiveSession(p, proposal.Session()))
	proposal.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(p.ID()),
	})
	proposal.Close()
}

func (p *managedPrimitive) close(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {
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

func (p *managedPrimitive) propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	session, ok := p.sessions.get(proposal.Session().ID())
	if !ok {
		proposal.Error(errors.NewForbidden("session not found"))
		proposal.Close()
	} else {
		p.sm.propose(newPrimitiveProposal(session, proposal))
	}
}

func (p *managedPrimitive) query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	session, ok := p.sessions.get(query.Session().ID())
	if !ok {
		query.Error(errors.NewForbidden("session not found"))
		query.Close()
	} else {
		p.sm.query(newPrimitiveQuery(session, query))
	}
}

func newPrimitiveSession(
	context *managedPrimitive,
	parent session.Session) *primitiveSession {
	s := &primitiveSession{
		primitive: context,
		parent:    parent,
		state:     session.Open,
		watchers:  make(map[uuid.UUID]session.WatchFunc[session.State]),
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
	primitive *managedPrimitive
	parent    session.Session
	state     session.State
	watchers  map[uuid.UUID]session.WatchFunc[session.State]
	cancel    session.CancelFunc
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

func (s *primitiveSession) Watch(watcher session.WatchFunc[session.State]) session.CancelFunc {
	id := uuid.New()
	s.watchers[id] = watcher
	return func() {
		delete(s.watchers, id)
	}
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
	parent session.PrimitiveQuery) *primitiveQuery {
	return &primitiveQuery{
		PrimitiveQuery: parent,
		session:        session,
	}
}

type primitiveQuery struct {
	session.PrimitiveQuery
	session *primitiveSession
}

func (q *primitiveQuery) Session() session.Session {
	return q.session
}

var _ session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput] = (*primitiveQuery)(nil)

func newPrimitiveProposal(
	session *primitiveSession,
	parent session.PrimitiveProposal) *primitiveProposal {
	return &primitiveProposal{
		PrimitiveProposal: parent,
		session:           session,
	}
}

type primitiveProposal struct {
	session.PrimitiveProposal
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
