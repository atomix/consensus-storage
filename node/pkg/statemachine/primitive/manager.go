// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	statemachine "github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/google/uuid"
)

var log = logging.GetLogger()

func NewManager(ctx session.Context, registry *TypeRegistry) session.PrimitiveManager {
	return &primitiveManagerStateMachine{
		Context:    ctx,
		registry:   registry,
		primitives: make(map[ID]*managedPrimitive),
	}
}

type primitiveManagerStateMachine struct {
	session.Context
	registry   *TypeRegistry
	primitives map[ID]*managedPrimitive
}

func (m *primitiveManagerStateMachine) Snapshot(writer *snapshot.Writer) error {
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

func (m *primitiveManagerStateMachine) Recover(reader *snapshot.Reader) error {
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

		context := newManagedContext(m.Context, ID(snapshot.PrimitiveID), snapshot.Spec)
		primitive := newManagedPrimitive(context, factory(context))
		m.primitives[primitive.ID()] = primitive
		if err := primitive.Recover(reader); err != nil {
			return nil
		}
	}
	return nil
}

func (m *primitiveManagerStateMachine) Propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	primitive, ok := m.primitives[ID(proposal.Input().PrimitiveID)]
	if !ok {
		proposal.Error(errors.NewForbidden("primitive %d not found", proposal.Input().PrimitiveID))
		proposal.Close()
	} else {
		primitive.propose(proposal)
	}
}

func (m *primitiveManagerStateMachine) CreatePrimitive(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {
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
		factory, ok := m.registry.lookup(proposal.Input().Service)
		if !ok {
			proposal.Error(errors.NewForbidden("unknown primitive type"))
			proposal.Close()
			return
		}
		context := newManagedContext(m.Context, ID(proposal.ID()), proposal.Input().PrimitiveSpec)
		primitive = newManagedPrimitive(context, factory(context))
		m.primitives[primitive.ID()] = primitive
	}

	primitive.open(proposal)
}

func (m *primitiveManagerStateMachine) ClosePrimitive(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {
	primitive, ok := m.primitives[ID(proposal.Input().PrimitiveID)]
	if !ok {
		proposal.Error(errors.NewForbidden("primitive %d not found", proposal.Input().PrimitiveID))
		proposal.Close()
	} else {
		primitive.close(proposal)
	}
}

func (m *primitiveManagerStateMachine) Query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	primitive, ok := m.primitives[ID(query.Input().PrimitiveID)]
	if !ok {
		query.Error(errors.NewFault("primitive %d not found", query.Input().PrimitiveID))
		query.Close()
	} else {
		primitive.query(query)
	}
}

func newManagedPrimitive(context *managedContext, sm primitiveDelegate) *managedPrimitive {
	return &managedPrimitive{
		managedContext: context,
		sm:             sm,
	}
}

type managedPrimitive struct {
	*managedContext
	sm primitiveDelegate
}

func (p *managedPrimitive) Snapshot(writer *snapshot.Writer) error {
	if err := p.managedContext.Snapshot(writer); err != nil {
		return err
	}
	if err := p.sm.Snapshot(writer); err != nil {
		return err
	}
	return nil
}

func (p *managedPrimitive) Recover(reader *snapshot.Reader) error {
	if err := p.managedContext.Recover(reader); err != nil {
		return err
	}
	if err := p.sm.Recover(reader); err != nil {
		return err
	}
	return nil
}

func (p *managedPrimitive) open(proposal session.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {
	p.sessions.add(newManagedSession(p.managedContext, proposal.Session()))
	proposal.Output(&multiraftv1.CreatePrimitiveOutput{
		PrimitiveID: multiraftv1.PrimitiveID(p.ID()),
	})
	proposal.Close()
}

func (p *managedPrimitive) close(proposal session.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {
	p.sessions.remove(proposal.Session().ID())
	proposal.Output(&multiraftv1.ClosePrimitiveOutput{})
	proposal.Close()
}

func (p *managedPrimitive) propose(proposal session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	session, ok := p.sessions.get(proposal.Session().ID())
	if !ok {
		proposal.Error(errors.NewForbidden("session not found"))
		proposal.Close()
	} else {
		p.sm.propose(newManagedProposal(session, proposal))
	}
}

func (p *managedPrimitive) query(query session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	session, ok := p.sessions.get(query.Session().ID())
	if !ok {
		query.Error(errors.NewForbidden("session not found"))
		query.Close()
	} else {
		p.sm.query(newManagedQuery(session, query))
	}
}

func newManagedContext(parent session.Context, id ID, spec multiraftv1.PrimitiveSpec) *managedContext {
	return &managedContext{
		Context:  parent,
		id:       id,
		spec:     spec,
		sessions: newManagedSessions(),
		log: parent.Log().WithFields(
			logging.String("Service", spec.Service),
			logging.Uint64("Primitive", uint64(id)),
			logging.String("Namespace", spec.Namespace),
			logging.String("Name", spec.Name)),
	}
}

type managedContext struct {
	session.Context
	id       ID
	spec     multiraftv1.PrimitiveSpec
	sessions *managedSessions
	log      logging.Logger
}

func (c *managedContext) Log() logging.Logger {
	return c.log
}

func (c *managedContext) ID() ID {
	return c.id
}

func (c *managedContext) Service() string {
	return c.spec.Service
}

func (c *managedContext) Namespace() string {
	return c.spec.Namespace
}

func (c *managedContext) Name() string {
	return c.spec.Name
}

func (c *managedContext) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(c.sessions.sessions)); err != nil {
		return err
	}
	for _, session := range c.sessions.list() {
		if err := writer.WriteVarUint64(uint64(session.ID())); err != nil {
			return err
		}
	}
	return nil
}

func (c *managedContext) Recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		sessionID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}
		parent, ok := c.Context.Sessions().Get(session.ID(sessionID))
		if !ok {
			return errors.NewFault("session %d not found", sessionID)
		}
		c.sessions.add(newManagedSession(c, parent))
	}
	return nil
}

func (c *managedContext) Sessions() session.Sessions {
	return c.sessions
}

func (c *managedContext) Proposals() session.Proposals {
	return newManagedContextProposals(c.id, c.Context.Proposals(), c.sessions)
}

func newManagedSession(
	context *managedContext,
	parent session.Session) *managedSession {
	s := &managedSession{
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

type managedSession struct {
	primitive *managedContext
	parent    session.Session
	state     session.State
	watchers  map[uuid.UUID]statemachine.WatchFunc[session.State]
	cancel    statemachine.CancelFunc
	log       logging.Logger
}

func (s *managedSession) Log() logging.Logger {
	return s.log
}

func (s *managedSession) ID() session.ID {
	return s.parent.ID()
}

func (s *managedSession) State() session.State {
	return s.state
}

func (s *managedSession) Watch(watcher statemachine.WatchFunc[session.State]) statemachine.CancelFunc {
	id := uuid.New()
	s.watchers[id] = watcher
	return func() {
		delete(s.watchers, id)
	}
}

func (s *managedSession) Proposals() session.Proposals {
	return newManagedSessionProposals(s)
}

func (s *managedSession) close() {
	s.cancel()
	s.primitive.sessions.remove(s.ID())
	s.state = session.Closed
	for _, watcher := range s.watchers {
		watcher(session.Closed)
	}
}

var _ session.Session = (*managedSession)(nil)

func newManagedSessions() *managedSessions {
	return &managedSessions{
		sessions: make(map[session.ID]*managedSession),
	}
}

type managedSessions struct {
	sessions map[session.ID]*managedSession
}

func (s *managedSessions) Get(id session.ID) (session.Session, bool) {
	session, ok := s.sessions[id]
	return session, ok
}

func (s *managedSessions) List() []session.Session {
	sessions := make([]session.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *managedSessions) add(session *managedSession) {
	s.sessions[session.ID()] = session
}

func (s *managedSessions) remove(sessionID session.ID) bool {
	if _, ok := s.sessions[sessionID]; ok {
		delete(s.sessions, sessionID)
		return true
	}
	return false
}

func (s *managedSessions) get(id session.ID) (*managedSession, bool) {
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

func newManagedQuery(
	session *managedSession,
	parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) *managedQuery {
	return &managedQuery{
		Query:   parent,
		session: session,
	}
}

type managedQuery struct {
	session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]
	session *managedSession
}

var _ session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput] = (*managedQuery)(nil)

func newManagedProposal(
	session *managedSession,
	parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) *managedProposal {
	return &managedProposal{
		Proposal: parent,
		session:  session,
	}
}

type managedProposal struct {
	session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
	session *managedSession
}

var _ session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] = (*managedProposal)(nil)

func newManagedContextProposals(primitiveID ID, proposals session.Proposals, sessions *managedSessions) *managedContextProposals {
	return &managedContextProposals{
		primitiveID: primitiveID,
		proposals:   proposals,
		sessions:    sessions,
	}
}

type managedContextProposals struct {
	primitiveID ID
	proposals   session.Proposals
	sessions    *managedSessions
}

func (p *managedContextProposals) Get(id statemachine.ProposalID) (session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput], bool) {
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
	return newManagedProposal(session, parent), true
}

func (p *managedContextProposals) List() []session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] {
	parents := p.proposals.List()
	proposals := make([]session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput], 0, len(parents))
	for _, parent := range parents {
		if ID(parent.Input().PrimitiveID) != p.primitiveID {
			continue
		}
		session, ok := p.sessions.get(parent.Session().ID())
		if !ok {
			continue
		}
		proposals = append(proposals, newManagedProposal(session, parent))
	}
	return proposals
}

func newManagedSessionProposals(session *managedSession) *managedSessionProposals {
	return &managedSessionProposals{
		session: session,
	}
}

type managedSessionProposals struct {
	session *managedSession
}

func (p *managedSessionProposals) Get(id statemachine.ProposalID) (session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput], bool) {
	parent, ok := p.session.parent.Proposals().Get(id)
	if !ok {
		return nil, false
	}
	if ID(parent.Input().PrimitiveID) != p.session.primitive.id {
		return nil, false
	}
	return newManagedProposal(p.session, parent), true
}

func (p *managedSessionProposals) List() []session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] {
	parents := p.session.parent.Proposals().List()
	proposals := make([]session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput], 0, len(parents))
	for _, parent := range parents {
		if ID(parent.Input().PrimitiveID) != p.session.primitive.id {
			continue
		}
		proposals = append(proposals, newManagedProposal(p.session, parent))
	}
	return proposals
}
