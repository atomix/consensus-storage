// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	electionv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/election/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.runtime.election.v1.LeaderElection"

func Register(registry *primitive.TypeRegistry) {
	primitive.RegisterType[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput](registry)(Type)
}

var Type = primitive.NewType[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput](Service, lockCodec, newLeaderElectionStateMachine)

var lockCodec = primitive.NewCodec[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput](
	func(bytes []byte) (*electionv1.LeaderElectionInput, error) {
		input := &electionv1.LeaderElectionInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *electionv1.LeaderElectionOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newLeaderElectionStateMachine(ctx primitive.Context[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput]) primitive.Primitive[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput] {
	sm := &LeaderElectionStateMachine{
		Context:  ctx,
		watchers: make(map[primitive.QueryID]primitive.Query[*electionv1.WatchInput, *electionv1.WatchOutput]),
	}
	sm.init()
	return sm
}

type LeaderElectionStateMachine struct {
	primitive.Context[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput]
	enter    primitive.Proposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.EnterInput, *electionv1.EnterOutput]
	withdraw primitive.Proposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.WithdrawInput, *electionv1.WithdrawOutput]
	anoint   primitive.Proposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.AnointInput, *electionv1.AnointOutput]
	promote  primitive.Proposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.PromoteInput, *electionv1.PromoteOutput]
	demote   primitive.Proposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.DemoteInput, *electionv1.DemoteOutput]
	evict    primitive.Proposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.EvictInput, *electionv1.EvictOutput]
	getTerm  primitive.Querier[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.GetTermInput, *electionv1.GetTermOutput]
	watch    primitive.Querier[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.WatchInput, *electionv1.WatchOutput]
	electionv1.LeaderElectionSnapshot
	cancel   primitive.CancelFunc
	watchers map[primitive.QueryID]primitive.Query[*electionv1.WatchInput, *electionv1.WatchOutput]
	mu       sync.RWMutex
}

func (s *LeaderElectionStateMachine) init() {
	s.enter = primitive.NewProposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.EnterInput, *electionv1.EnterOutput](s).
		Name("Enter").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.EnterInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Enter); ok {
				return set.Enter, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.EnterOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Enter{
					Enter: output,
				},
			}
		}).
		Build(s.doEnter)
	s.withdraw = primitive.NewProposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.WithdrawInput, *electionv1.WithdrawOutput](s).
		Name("Withdraw").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.WithdrawInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Withdraw); ok {
				return set.Withdraw, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.WithdrawOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Withdraw{
					Withdraw: output,
				},
			}
		}).
		Build(s.doWithdraw)
	s.anoint = primitive.NewProposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.AnointInput, *electionv1.AnointOutput](s).
		Name("Anoint").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.AnointInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Anoint); ok {
				return set.Anoint, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.AnointOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Anoint{
					Anoint: output,
				},
			}
		}).
		Build(s.doAnoint)
	s.promote = primitive.NewProposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.PromoteInput, *electionv1.PromoteOutput](s).
		Name("Promote").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.PromoteInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Promote); ok {
				return set.Promote, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.PromoteOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Promote{
					Promote: output,
				},
			}
		}).
		Build(s.doPromote)
	s.demote = primitive.NewProposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.DemoteInput, *electionv1.DemoteOutput](s).
		Name("Demote").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.DemoteInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Demote); ok {
				return set.Demote, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.DemoteOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Demote{
					Demote: output,
				},
			}
		}).
		Build(s.doDemote)
	s.evict = primitive.NewProposer[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.EvictInput, *electionv1.EvictOutput](s).
		Name("Evict").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.EvictInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Evict); ok {
				return set.Evict, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.EvictOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Evict{
					Evict: output,
				},
			}
		}).
		Build(s.doEvict)
	s.getTerm = primitive.NewQuerier[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.GetTermInput, *electionv1.GetTermOutput](s).
		Name("GetTerm").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.GetTermInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_GetTerm); ok {
				return set.GetTerm, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.GetTermOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_GetTerm{
					GetTerm: output,
				},
			}
		}).
		Build(s.doGetTerm)
	s.watch = primitive.NewQuerier[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput, *electionv1.WatchInput, *electionv1.WatchOutput](s).
		Name("Watch").
		Decoder(func(input *electionv1.LeaderElectionInput) (*electionv1.WatchInput, bool) {
			if set, ok := input.Input.(*electionv1.LeaderElectionInput_Watch); ok {
				return set.Watch, true
			}
			return nil, false
		}).
		Encoder(func(output *electionv1.WatchOutput) *electionv1.LeaderElectionOutput {
			return &electionv1.LeaderElectionOutput{
				Output: &electionv1.LeaderElectionOutput_Watch{
					Watch: output,
				},
			}
		}).
		Build(s.doWatch)
}

func (s *LeaderElectionStateMachine) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteMessage(&s.LeaderElectionSnapshot); err != nil {
		return err
	}
	return nil
}

func (s *LeaderElectionStateMachine) Recover(reader *snapshot.Reader) error {
	if err := reader.ReadMessage(&s.LeaderElectionSnapshot); err != nil {
		return err
	}
	if s.Leader != nil {
		if err := s.watchSession(primitive.SessionID(s.Leader.SessionID)); err != nil {
			return err
		}
	}
	return nil
}

func (s *LeaderElectionStateMachine) watchSession(sessionID primitive.SessionID) error {
	if s.cancel != nil {
		s.cancel()
	}
	session, ok := s.Sessions().Get(sessionID)
	if !ok {
		return errors.NewFault("unknown session")
	}
	s.cancel = session.Watch(func(state primitive.SessionState) {
		if state == primitive.SessionClosed {
			s.Leader = nil
			for s.Candidates != nil {
				candidate := s.Candidates[0]
				s.Candidates = s.Candidates[1:]
				if primitive.SessionID(candidate.SessionID) != session.ID() {
					s.Leader = &candidate
					s.Term++
					s.notify(s.term())
					break
				}
			}
		}
	})
	return nil
}

func (s *LeaderElectionStateMachine) term() electionv1.Term {
	term := electionv1.Term{
		Index: s.Term,
	}
	if s.Leader != nil {
		term.Leader = s.Leader.Name
	}
	for _, candidate := range s.Candidates {
		term.Candidates = append(term.Candidates, candidate.Name)
	}
	return term
}

func (s *LeaderElectionStateMachine) Propose(proposal primitive.Proposal[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput]) {
	switch proposal.Input().Input.(type) {
	case *electionv1.LeaderElectionInput_Enter:
		s.enter.Execute(proposal)
	case *electionv1.LeaderElectionInput_Withdraw:
		s.withdraw.Execute(proposal)
	case *electionv1.LeaderElectionInput_Anoint:
		s.anoint.Execute(proposal)
	case *electionv1.LeaderElectionInput_Promote:
		s.promote.Execute(proposal)
	case *electionv1.LeaderElectionInput_Demote:
		s.demote.Execute(proposal)
	case *electionv1.LeaderElectionInput_Evict:
		s.evict.Execute(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
	}
}

func (s *LeaderElectionStateMachine) candidateExists(name string) bool {
	for _, candidate := range s.Candidates {
		if candidate.Name == name {
			return true
		}
	}
	return false
}

func (s *LeaderElectionStateMachine) doEnter(proposal primitive.Proposal[*electionv1.EnterInput, *electionv1.EnterOutput]) {
	defer proposal.Close()
	if s.Leader == nil {
		s.Term++
		s.Leader = &electionv1.LeaderElectionCandidate{
			Name:      proposal.Input().Candidate,
			SessionID: multiraftv1.SessionID(proposal.Session().ID()),
		}
		if err := s.watchSession(proposal.Session().ID()); err != nil {
			panic(err)
		}
		term := s.term()
		s.notify(term)
		proposal.Output(&electionv1.EnterOutput{
			Term: term,
		})
	} else if !s.candidateExists(proposal.Input().Candidate) {
		s.Candidates = append(s.Candidates, electionv1.LeaderElectionCandidate{
			Name:      proposal.Input().Candidate,
			SessionID: multiraftv1.SessionID(proposal.Session().ID()),
		})
		term := s.term()
		s.notify(term)
		proposal.Output(&electionv1.EnterOutput{
			Term: term,
		})
	} else {
		proposal.Output(&electionv1.EnterOutput{
			Term: s.term(),
		})
	}
}

func (s *LeaderElectionStateMachine) doWithdraw(proposal primitive.Proposal[*electionv1.WithdrawInput, *electionv1.WithdrawOutput]) {
	defer proposal.Close()
	if s.Leader != nil && primitive.SessionID(s.Leader.SessionID) == proposal.Session().ID() {
		s.Leader = nil
		if len(s.Candidates) > 0 {
			candidate := s.Candidates[0]
			s.Candidates = s.Candidates[1:]
			s.Leader = &candidate
			s.Term++
		}
		term := s.term()
		s.notify(term)
		proposal.Output(&electionv1.WithdrawOutput{
			Term: term,
		})
	} else {
		proposal.Output(&electionv1.WithdrawOutput{
			Term: s.term(),
		})
	}
}

func (s *LeaderElectionStateMachine) doAnoint(proposal primitive.Proposal[*electionv1.AnointInput, *electionv1.AnointOutput]) {
	defer proposal.Close()
	if s.Leader == nil || s.Leader.Name == proposal.Input().Candidate || !s.candidateExists(proposal.Input().Candidate) {
		proposal.Output(&electionv1.AnointOutput{
			Term: s.term(),
		})
		return
	}

	candidates := make([]electionv1.LeaderElectionCandidate, 0, len(s.Candidates))
	for _, candidate := range s.Candidates {
		if candidate.Name == proposal.Input().Candidate {
			s.Term++
			s.Leader = &candidate
		} else {
			candidates = append(candidates, candidate)
		}
	}
	s.Candidates = candidates

	term := s.term()
	s.notify(term)
	proposal.Output(&electionv1.AnointOutput{
		Term: term,
	})
}

func (s *LeaderElectionStateMachine) doPromote(proposal primitive.Proposal[*electionv1.PromoteInput, *electionv1.PromoteOutput]) {
	defer proposal.Close()
	if s.Leader == nil || s.Leader.Name == proposal.Input().Candidate || !s.candidateExists(proposal.Input().Candidate) {
		proposal.Output(&electionv1.PromoteOutput{
			Term: s.term(),
		})
		return
	}

	if s.Candidates[0].Name == proposal.Input().Candidate {
		s.Term++
		s.Leader = &s.Candidates[0]
		s.Candidates = s.Candidates[1:]
	} else {
		var index int
		for i, candidate := range s.Candidates {
			if candidate.Name == proposal.Input().Candidate {
				index = i
				break
			}
		}

		candidates := make([]electionv1.LeaderElectionCandidate, 0, len(s.Candidates))
		for i, candidate := range s.Candidates {
			if i < index-1 {
				candidates[i] = candidate
			} else if i == index-1 {
				candidates[i] = s.Candidates[index]
			} else if i == index {
				candidates[i] = s.Candidates[i-1]
			} else {
				candidates[i] = candidate
			}
		}
		s.Candidates = candidates
	}

	term := s.term()
	s.notify(term)
	proposal.Output(&electionv1.PromoteOutput{
		Term: term,
	})
}

func (s *LeaderElectionStateMachine) doDemote(proposal primitive.Proposal[*electionv1.DemoteInput, *electionv1.DemoteOutput]) {
	defer proposal.Close()
	if s.Leader == nil || s.Leader.Name == proposal.Input().Candidate || !s.candidateExists(proposal.Input().Candidate) {
		proposal.Output(&electionv1.DemoteOutput{
			Term: s.term(),
		})
		return
	}

	if s.Leader != nil && s.Leader.Name == proposal.Input().Candidate {
		if len(s.Candidates) == 0 {
			proposal.Output(&electionv1.DemoteOutput{
				Term: s.term(),
			})
			return
		}
		leader := s.Leader
		s.Term++
		s.Leader = &s.Candidates[0]
		s.Candidates = append([]electionv1.LeaderElectionCandidate{*leader}, s.Candidates[1:]...)
	} else if s.Candidates[len(s.Candidates)-1].Name != proposal.Input().Candidate {
		var index int
		for i, candidate := range s.Candidates {
			if candidate.Name == proposal.Input().Candidate {
				index = i
				break
			}
		}

		candidates := make([]electionv1.LeaderElectionCandidate, 0, len(s.Candidates))
		for i, candidate := range s.Candidates {
			if i < index+1 {
				candidates[i] = candidate
			} else if i == index+1 {
				candidates[i] = s.Candidates[index]
			} else if i == index {
				candidates[i] = s.Candidates[i+1]
			} else {
				candidates[i] = candidate
			}
		}
		s.Candidates = candidates
	}

	term := s.term()
	s.notify(term)
	proposal.Output(&electionv1.DemoteOutput{
		Term: term,
	})
}

func (s *LeaderElectionStateMachine) doEvict(proposal primitive.Proposal[*electionv1.EvictInput, *electionv1.EvictOutput]) {
	defer proposal.Close()
	if s.Leader == nil || (s.Leader.Name != proposal.Input().Candidate && !s.candidateExists(proposal.Input().Candidate)) {
		proposal.Output(&electionv1.EvictOutput{
			Term: s.term(),
		})
		return
	}

	if s.Leader.Name == proposal.Input().Candidate {
		s.Leader = nil
		if len(s.Candidates) > 0 {
			candidate := s.Candidates[0]
			s.Candidates = s.Candidates[1:]
			s.Leader = &candidate
			s.Term++
		}
	} else {
		candidates := make([]electionv1.LeaderElectionCandidate, 0, len(s.Candidates))
		for _, candidate := range s.Candidates {
			if candidate.Name != proposal.Input().Candidate {
				candidates = append(candidates, candidate)
			}
		}
		s.Candidates = candidates
	}

	term := s.term()
	s.notify(term)
	proposal.Output(&electionv1.EvictOutput{
		Term: term,
	})
}

func (s *LeaderElectionStateMachine) Query(query primitive.Query[*electionv1.LeaderElectionInput, *electionv1.LeaderElectionOutput]) {
	switch query.Input().Input.(type) {
	case *electionv1.LeaderElectionInput_GetTerm:
		s.getTerm.Execute(query)
	case *electionv1.LeaderElectionInput_Watch:
		s.watch.Execute(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *LeaderElectionStateMachine) doGetTerm(query primitive.Query[*electionv1.GetTermInput, *electionv1.GetTermOutput]) {
	defer query.Close()
	query.Output(&electionv1.GetTermOutput{
		Term: s.term(),
	})
}

func (s *LeaderElectionStateMachine) doWatch(query primitive.Query[*electionv1.WatchInput, *electionv1.WatchOutput]) {
	s.mu.Lock()
	s.watchers[query.ID()] = query
	s.mu.Unlock()
}

func (s *LeaderElectionStateMachine) notify(term electionv1.Term) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, watcher := range s.watchers {
		watcher.Output(&electionv1.WatchOutput{
			Term: term,
		})
	}
}
