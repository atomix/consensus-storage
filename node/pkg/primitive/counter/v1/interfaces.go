// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"fmt"
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/gogo/protobuf/proto"
)

type StateMachine interface {
	Context
	Backup(*snapshot.Writer) error
	Restore(*snapshot.Reader) error
	// Set sets the counter value
	Set(SetProposal) (*counterv1.SetOutput, error)
	CompareAndSet(proposal CompareAndSetProposal) (*counterv1.CompareAndSetOutput, error)
	// Get gets the current counter value
	Get(GetQuery) (*counterv1.GetOutput, error)
	// Increment increments the counter value
	Increment(IncrementProposal) (*counterv1.IncrementOutput, error)
	// Decrement decrements the counter value
	Decrement(DecrementProposal) (*counterv1.DecrementOutput, error)
}

type Context interface {
	Scheduler() primitive.Scheduler
	Sessions() Sessions
	Proposals() Proposals
}

func newPrimitiveContext(service primitive.Context) Context {
	return &serviceContext{
		scheduler: service.Scheduler(),
		sessions:  newSessions(service.Sessions()),
		proposals: newProposals(service.Commands()),
	}
}

type serviceContext struct {
	scheduler primitive.Scheduler
	sessions  Sessions
	proposals Proposals
}

func (s *serviceContext) Scheduler() primitive.Scheduler {
	return s.scheduler
}

func (s *serviceContext) Sessions() Sessions {
	return s.sessions
}

func (s *serviceContext) Proposals() Proposals {
	return s.proposals
}

var _ Context = &serviceContext{}

type Sessions interface {
	Get(SessionID) (Session, bool)
	List() []Session
}

func newSessions(sessions primitive.Sessions) Sessions {
	return &serviceSessions{
		sessions: sessions,
	}
}

type serviceSessions struct {
	sessions primitive.Sessions
}

func (s *serviceSessions) Get(id SessionID) (Session, bool) {
	session, ok := s.sessions.Get(primitive.SessionID(id))
	if !ok {
		return nil, false
	}
	return newSession(session), true
}

func (s *serviceSessions) List() []Session {
	serviceSessions := s.sessions.List()
	sessions := make([]Session, len(serviceSessions))
	for i, serviceSession := range serviceSessions {
		sessions[i] = newSession(serviceSession)
	}
	return sessions
}

var _ Sessions = &serviceSessions{}

type SessionID uint64

type SessionState int

const (
	SessionClosed SessionState = iota
	SessionOpen
)

type Watcher interface {
	Cancel()
}

func newWatcher(watcher primitive.Watcher) Watcher {
	return &serviceWatcher{
		watcher: watcher,
	}
}

type serviceWatcher struct {
	watcher primitive.Watcher
}

func (s *serviceWatcher) Cancel() {
	s.watcher.Cancel()
}

var _ Watcher = &serviceWatcher{}

type Session interface {
	ID() SessionID
	State() SessionState
	Watch(func(SessionState)) Watcher
	Proposals() Proposals
}

func newSession(session primitive.Session) Session {
	return &serviceSession{
		session:   session,
		proposals: newProposals(session.Commands()),
	}
}

type serviceSession struct {
	session   primitive.Session
	proposals Proposals
}

func (s *serviceSession) ID() SessionID {
	return SessionID(s.session.ID())
}

func (s *serviceSession) Proposals() Proposals {
	return s.proposals
}

func (s *serviceSession) State() SessionState {
	return SessionState(s.session.State())
}

func (s *serviceSession) Watch(f func(SessionState)) Watcher {
	return newWatcher(s.session.Watch(func(state primitive.SessionState) {
		f(SessionState(state))
	}))
}

var _ Session = &serviceSession{}

type Proposals interface {
	Set() SetProposals
	CompareAndSet() CompareAndSetProposals
	Increment() IncrementProposals
	Decrement() DecrementProposals
}

func newProposals(commands primitive.Commands) Proposals {
	return &serviceProposals{
		setProposals:           newSetProposals(commands),
		compareAndSetProposals: newCompareAndSetProposals(commands),
		incrementProposals:     newIncrementProposals(commands),
		decrementProposals:     newDecrementProposals(commands),
	}
}

type serviceProposals struct {
	setProposals           SetProposals
	compareAndSetProposals CompareAndSetProposals

	incrementProposals IncrementProposals
	decrementProposals DecrementProposals
}

func (s *serviceProposals) Set() SetProposals {
	return s.setProposals
}
func (s *serviceProposals) CompareAndSet() CompareAndSetProposals {
	return s.compareAndSetProposals
}
func (s *serviceProposals) Increment() IncrementProposals {
	return s.incrementProposals
}
func (s *serviceProposals) Decrement() DecrementProposals {
	return s.decrementProposals
}

var _ Proposals = &serviceProposals{}

type ProposalID uint64

type ProposalState int

const (
	ProposalComplete ProposalState = iota
	ProposalOpen
)

type Proposal interface {
	fmt.Stringer
	ID() ProposalID
	Session() Session
	State() ProposalState
	Watch(func(ProposalState)) Watcher
}

func newProposal(command primitive.Command) Proposal {
	return &serviceProposal{
		command: command,
	}
}

type serviceProposal struct {
	command primitive.Command
}

func (p *serviceProposal) ID() ProposalID {
	return ProposalID(p.command.ID())
}

func (p *serviceProposal) Session() Session {
	return newSession(p.command.Session())
}

func (p *serviceProposal) State() ProposalState {
	return ProposalState(p.command.State())
}

func (p *serviceProposal) Watch(f func(ProposalState)) Watcher {
	return newWatcher(p.command.Watch(func(state primitive.CommandState) {
		f(ProposalState(state))
	}))
}

func (p *serviceProposal) String() string {
	return fmt.Sprintf("ProposalID: %d, SessionID: %d", p.ID(), p.Session().ID())
}

var _ Proposal = &serviceProposal{}

type Query interface {
	fmt.Stringer
	Session() Session
}

func newQuery(query primitive.Query) Query {
	return &serviceQuery{
		query: query,
	}
}

type serviceQuery struct {
	query primitive.Query
}

func (p *serviceQuery) Session() Session {
	return newSession(p.query.Session())
}

func (p *serviceQuery) String() string {
	return fmt.Sprintf("SessionID: %d", p.Session().ID())
}

var _ Query = &serviceQuery{}

type SetProposals interface {
	Get(ProposalID) (SetProposal, bool)
	List() []SetProposal
}

func newSetProposals(commands primitive.Commands) SetProposals {
	return &setProposals{
		commands: commands,
	}
}

type setProposals struct {
	commands primitive.Commands
}

func (p *setProposals) Get(id ProposalID) (SetProposal, bool) {
	command, ok := p.commands.Get(primitive.CommandID(id))
	if !ok {
		return nil, false
	}
	proposal, err := newSetProposal(command)
	if err != nil {
		log.Error(err)
		return nil, false
	}
	return proposal, true
}

func (p *setProposals) List() []SetProposal {
	commands := p.commands.List(primitive.OperationID(1))
	proposals := make([]SetProposal, len(commands))
	for i, command := range commands {
		proposal, err := newSetProposal(command)
		if err != nil {
			log.Error(err)
		} else {
			proposals[i] = proposal
		}
	}
	return proposals
}

var _ SetProposals = &setProposals{}

type SetProposal interface {
	Proposal
	Input() *counterv1.SetInput
}

func newSetProposal(command primitive.Command) (SetProposal, error) {
	request := &counterv1.SetInput{}
	if err := proto.Unmarshal(command.Input(), request); err != nil {
		return nil, err
	}
	return &setProposal{
		Proposal: newProposal(command),
		command:  command,
		request:  request,
	}, nil
}

type setProposal struct {
	Proposal
	command primitive.Command
	request *counterv1.SetInput
}

func (p *setProposal) Input() *counterv1.SetInput {
	return p.request
}

func (p *setProposal) String() string {
	return fmt.Sprintf("ProposalID=%d, SessionID=%d", p.ID(), p.Session().ID())
}

var _ SetProposal = &setProposal{}

type CompareAndSetProposals interface {
	Get(ProposalID) (CompareAndSetProposal, bool)
	List() []CompareAndSetProposal
}

func newCompareAndSetProposals(commands primitive.Commands) CompareAndSetProposals {
	return &compareAndSetProposals{
		commands: commands,
	}
}

type compareAndSetProposals struct {
	commands primitive.Commands
}

func (p *compareAndSetProposals) Get(id ProposalID) (CompareAndSetProposal, bool) {
	command, ok := p.commands.Get(primitive.CommandID(id))
	if !ok {
		return nil, false
	}
	proposal, err := newCompareAndSetProposal(command)
	if err != nil {
		log.Error(err)
		return nil, false
	}
	return proposal, true
}

func (p *compareAndSetProposals) List() []CompareAndSetProposal {
	commands := p.commands.List(setOp)
	proposals := make([]CompareAndSetProposal, len(commands))
	for i, command := range commands {
		proposal, err := newCompareAndSetProposal(command)
		if err != nil {
			log.Error(err)
		} else {
			proposals[i] = proposal
		}
	}
	return proposals
}

var _ CompareAndSetProposals = &compareAndSetProposals{}

type CompareAndSetProposal interface {
	Proposal
	Input() *counterv1.CompareAndSetInput
}

func newCompareAndSetProposal(command primitive.Command) (CompareAndSetProposal, error) {
	request := &counterv1.CompareAndSetInput{}
	if err := proto.Unmarshal(command.Input(), request); err != nil {
		return nil, err
	}
	return &compareAndSetProposal{
		Proposal: newProposal(command),
		command:  command,
		request:  request,
	}, nil
}

type compareAndSetProposal struct {
	Proposal
	command primitive.Command
	request *counterv1.CompareAndSetInput
}

func (p *compareAndSetProposal) Input() *counterv1.CompareAndSetInput {
	return p.request
}

func (p *compareAndSetProposal) String() string {
	return fmt.Sprintf("ProposalID=%d, SessionID=%d", p.ID(), p.Session().ID())
}

var _ CompareAndSetProposal = &compareAndSetProposal{}

type GetQuery interface {
	Query
	Input() *counterv1.GetInput
}

func newGetQuery(query primitive.Query) (GetQuery, error) {
	request := &counterv1.GetInput{}
	if err := proto.Unmarshal(query.Input(), request); err != nil {
		return nil, err
	}
	return &getQuery{
		Query:   newQuery(query),
		query:   query,
		request: request,
	}, nil
}

type getQuery struct {
	Query
	query   primitive.Query
	request *counterv1.GetInput
}

func (p *getQuery) Input() *counterv1.GetInput {
	return p.request
}

func (p *getQuery) String() string {
	return fmt.Sprintf("SessionID=%d", p.Session().ID())
}

var _ GetQuery = &getQuery{}

type IncrementProposals interface {
	Get(ProposalID) (IncrementProposal, bool)
	List() []IncrementProposal
}

func newIncrementProposals(commands primitive.Commands) IncrementProposals {
	return &incrementProposals{
		commands: commands,
	}
}

type incrementProposals struct {
	commands primitive.Commands
}

func (p *incrementProposals) Get(id ProposalID) (IncrementProposal, bool) {
	command, ok := p.commands.Get(primitive.CommandID(id))
	if !ok {
		return nil, false
	}
	proposal, err := newIncrementProposal(command)
	if err != nil {
		log.Error(err)
		return nil, false
	}
	return proposal, true
}

func (p *incrementProposals) List() []IncrementProposal {
	commands := p.commands.List(primitive.OperationID(3))
	proposals := make([]IncrementProposal, len(commands))
	for i, command := range commands {
		proposal, err := newIncrementProposal(command)
		if err != nil {
			log.Error(err)
		} else {
			proposals[i] = proposal
		}
	}
	return proposals
}

var _ IncrementProposals = &incrementProposals{}

type IncrementProposal interface {
	Proposal
	Input() *counterv1.IncrementInput
}

func newIncrementProposal(command primitive.Command) (IncrementProposal, error) {
	request := &counterv1.IncrementInput{}
	if err := proto.Unmarshal(command.Input(), request); err != nil {
		return nil, err
	}
	return &incrementProposal{
		Proposal: newProposal(command),
		command:  command,
		request:  request,
	}, nil
}

type incrementProposal struct {
	Proposal
	command primitive.Command
	request *counterv1.IncrementInput
}

func (p *incrementProposal) Input() *counterv1.IncrementInput {
	return p.request
}

func (p *incrementProposal) String() string {
	return fmt.Sprintf("ProposalID=%d, SessionID=%d", p.ID(), p.Session().ID())
}

var _ IncrementProposal = &incrementProposal{}

type DecrementProposals interface {
	Get(ProposalID) (DecrementProposal, bool)
	List() []DecrementProposal
}

func newDecrementProposals(commands primitive.Commands) DecrementProposals {
	return &decrementProposals{
		commands: commands,
	}
}

type decrementProposals struct {
	commands primitive.Commands
}

func (p *decrementProposals) Get(id ProposalID) (DecrementProposal, bool) {
	command, ok := p.commands.Get(primitive.CommandID(id))
	if !ok {
		return nil, false
	}
	proposal, err := newDecrementProposal(command)
	if err != nil {
		log.Error(err)
		return nil, false
	}
	return proposal, true
}

func (p *decrementProposals) List() []DecrementProposal {
	commands := p.commands.List(primitive.OperationID(4))
	proposals := make([]DecrementProposal, len(commands))
	for i, command := range commands {
		proposal, err := newDecrementProposal(command)
		if err != nil {
			log.Error(err)
		} else {
			proposals[i] = proposal
		}
	}
	return proposals
}

var _ DecrementProposals = &decrementProposals{}

type DecrementProposal interface {
	Proposal
	Input() *counterv1.DecrementInput
}

func newDecrementProposal(command primitive.Command) (DecrementProposal, error) {
	request := &counterv1.DecrementInput{}
	if err := proto.Unmarshal(command.Input(), request); err != nil {
		return nil, err
	}
	return &decrementProposal{
		Proposal: newProposal(command),
		command:  command,
		request:  request,
	}, nil
}

type decrementProposal struct {
	Proposal
	command primitive.Command
	request *counterv1.DecrementInput
}

func (p *decrementProposal) Input() *counterv1.DecrementInput {
	return p.request
}

func (p *decrementProposal) String() string {
	return fmt.Sprintf("ProposalID=%d, SessionID=%d", p.ID(), p.Session().ID())
}

var _ DecrementProposal = &decrementProposal{}
