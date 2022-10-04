// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	lockv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/lock/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

const Service = "atomix.runtime.lock.v1.Lock"

func Register(registry *primitive.TypeRegistry) {
	primitive.RegisterType[*lockv1.LockInput, *lockv1.LockOutput](registry)(Type)
}

var Type = primitive.NewType[*lockv1.LockInput, *lockv1.LockOutput](Service, lockCodec, newLockStateMachine)

var lockCodec = primitive.NewCodec[*lockv1.LockInput, *lockv1.LockOutput](
	func(bytes []byte) (*lockv1.LockInput, error) {
		input := &lockv1.LockInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *lockv1.LockOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newLockStateMachine(ctx primitive.Context[*lockv1.LockInput, *lockv1.LockOutput]) primitive.Primitive[*lockv1.LockInput, *lockv1.LockOutput] {
	sm := &LockStateMachine{
		Context:   ctx,
		proposals: make(map[primitive.ProposalID]primitive.CancelFunc),
		timers:    make(map[primitive.ProposalID]primitive.CancelFunc),
	}
	sm.init()
	return sm
}

type lock struct {
	proposalID primitive.ProposalID
	sessionID  primitive.SessionID
	watcher    primitive.CancelFunc
}

type LockStateMachine struct {
	primitive.Context[*lockv1.LockInput, *lockv1.LockOutput]
	lock      *lock
	queue     []primitive.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]
	proposals map[primitive.ProposalID]primitive.CancelFunc
	timers    map[primitive.ProposalID]primitive.CancelFunc
	acquire   primitive.Proposer[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.AcquireInput, *lockv1.AcquireOutput]
	release   primitive.Proposer[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.ReleaseInput, *lockv1.ReleaseOutput]
	get       primitive.Querier[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.GetInput, *lockv1.GetOutput]
}

func (s *LockStateMachine) init() {
	s.acquire = primitive.NewProposer[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.AcquireInput, *lockv1.AcquireOutput](s).
		Name("Acquire").
		Decoder(func(input *lockv1.LockInput) (*lockv1.AcquireInput, bool) {
			if set, ok := input.Input.(*lockv1.LockInput_Acquire); ok {
				return set.Acquire, true
			}
			return nil, false
		}).
		Encoder(func(output *lockv1.AcquireOutput) *lockv1.LockOutput {
			return &lockv1.LockOutput{
				Output: &lockv1.LockOutput_Acquire{
					Acquire: output,
				},
			}
		}).
		Build(s.doAcquire)
	s.release = primitive.NewProposer[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.ReleaseInput, *lockv1.ReleaseOutput](s).
		Name("Release").
		Decoder(func(input *lockv1.LockInput) (*lockv1.ReleaseInput, bool) {
			if set, ok := input.Input.(*lockv1.LockInput_Release); ok {
				return set.Release, true
			}
			return nil, false
		}).
		Encoder(func(output *lockv1.ReleaseOutput) *lockv1.LockOutput {
			return &lockv1.LockOutput{
				Output: &lockv1.LockOutput_Release{
					Release: output,
				},
			}
		}).
		Build(s.doRelease)
	s.get = primitive.NewQuerier[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.GetInput, *lockv1.GetOutput](s).
		Name("Get").
		Decoder(func(input *lockv1.LockInput) (*lockv1.GetInput, bool) {
			if set, ok := input.Input.(*lockv1.LockInput_Get); ok {
				return set.Get, true
			}
			return nil, false
		}).
		Encoder(func(output *lockv1.GetOutput) *lockv1.LockOutput {
			return &lockv1.LockOutput{
				Output: &lockv1.LockOutput_Get{
					Get: output,
				},
			}
		}).
		Build(s.doGet)
}

func (s *LockStateMachine) Snapshot(writer *snapshot.Writer) error {
	if s.lock != nil {
		if err := writer.WriteBool(true); err != nil {
			return err
		}
		if err := writer.WriteVarUint64(uint64(s.lock.proposalID)); err != nil {
			return err
		}
		if err := writer.WriteVarUint64(uint64(s.lock.sessionID)); err != nil {
			return err
		}
		if err := writer.WriteVarInt(len(s.queue)); err != nil {
			return err
		}
		for _, waiter := range s.queue {
			if err := writer.WriteVarUint64(uint64(waiter.ID())); err != nil {
				return err
			}
		}
	} else {
		if err := writer.WriteBool(false); err != nil {
			return err
		}
	}
	return nil
}

func (s *LockStateMachine) Recover(reader *snapshot.Reader) error {
	locked, err := reader.ReadBool()
	if err != nil {
		return err
	}

	if locked {
		proposalID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}
		sessionID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}

		session, ok := s.Sessions().Get(primitive.SessionID(sessionID))
		if !ok {
			return errors.NewFault("session not found")
		}

		s.lock = &lock{
			proposalID: primitive.ProposalID(proposalID),
			sessionID:  primitive.SessionID(sessionID),
			watcher: session.Watch(func(state primitive.SessionState) {
				if state == primitive.SessionClosed {
					s.nextRequest()
				}
			}),
		}

		n, err := reader.ReadVarInt()
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			proposalID, err := reader.ReadVarUint64()
			if err != nil {
				return err
			}
			proposal, ok := s.acquire.Proposals().Get(primitive.ProposalID(proposalID))
			if !ok {
				return errors.NewFault("proposal not found")
			}
			s.enqueueRequest(proposal)
		}
	}
	return nil
}

func (s *LockStateMachine) Propose(proposal primitive.Proposal[*lockv1.LockInput, *lockv1.LockOutput]) {
	switch proposal.Input().Input.(type) {
	case *lockv1.LockInput_Acquire:
		s.acquire.Execute(proposal)
	case *lockv1.LockInput_Release:
		s.release.Execute(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
	}
}

func (s *LockStateMachine) enqueueRequest(proposal primitive.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]) {
	s.queue = append(s.queue, proposal)
	s.watchRequest(proposal)
}

func (s *LockStateMachine) dequeueRequest(proposal primitive.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]) {
	s.unwatchRequest(proposal.ID())
	queue := make([]primitive.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput], 0, len(s.queue))
	for _, waiter := range s.queue {
		if waiter.ID() != proposal.ID() {
			queue = append(queue, waiter)
		}
	}
	s.queue = queue
}

func (s *LockStateMachine) nextRequest() {
	s.lock = nil
	if s.queue == nil {
		return
	}
	proposal := s.queue[0]
	s.lock = &lock{
		proposalID: proposal.ID(),
		sessionID:  proposal.Session().ID(),
		watcher: proposal.Session().Watch(func(state primitive.SessionState) {
			if state == primitive.SessionClosed {
				s.nextRequest()
			}
		}),
	}
	s.queue = s.queue[1:]
	s.unwatchRequest(proposal.ID())
	proposal.Output(&lockv1.AcquireOutput{
		Index: multiraftv1.Index(proposal.ID()),
	})
	proposal.Close()
}

func (s *LockStateMachine) watchRequest(proposal primitive.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]) {
	s.proposals[proposal.ID()] = proposal.Watch(func(state primitive.ProposalState) {
		if primitive.IsDone(state) {
			s.dequeueRequest(proposal)
		}
	})

	if proposal.Input().Timeout != nil {
		s.timers[proposal.ID()] = s.Scheduler().Delay(*proposal.Input().Timeout, func() {
			s.dequeueRequest(proposal)
			proposal.Error(errors.NewConflict("lock already held"))
			proposal.Close()
		})
	}
}

func (s *LockStateMachine) unwatchRequest(proposalID primitive.ProposalID) {
	if cancel, ok := s.proposals[proposalID]; ok {
		cancel()
		delete(s.proposals, proposalID)
	}
	if cancel, ok := s.timers[proposalID]; ok {
		cancel()
		delete(s.timers, proposalID)
	}
}

func (s *LockStateMachine) doAcquire(proposal primitive.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]) {
	if s.lock == nil {
		s.lock = &lock{
			proposalID: proposal.ID(),
			sessionID:  proposal.Session().ID(),
			watcher: proposal.Session().Watch(func(state primitive.SessionState) {
				if state == primitive.SessionClosed {
					s.nextRequest()
				}
			}),
		}
		proposal.Output(&lockv1.AcquireOutput{
			Index: multiraftv1.Index(proposal.ID()),
		})
		proposal.Close()
	} else {
		s.enqueueRequest(proposal)
	}
}

func (s *LockStateMachine) doRelease(proposal primitive.Proposal[*lockv1.ReleaseInput, *lockv1.ReleaseOutput]) {
	defer proposal.Close()
	if s.lock == nil || s.lock.sessionID != proposal.Session().ID() {
		proposal.Error(errors.NewConflict("lock not held by client"))
	} else {
		s.nextRequest()
		proposal.Output(&lockv1.ReleaseOutput{})
	}
}

func (s *LockStateMachine) Query(query primitive.Query[*lockv1.LockInput, *lockv1.LockOutput]) {
	switch query.Input().Input.(type) {
	case *lockv1.LockInput_Get:
		s.get.Execute(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *LockStateMachine) doGet(query primitive.Query[*lockv1.GetInput, *lockv1.GetOutput]) {
	defer query.Close()
	if s.lock != nil {
		query.Output(&lockv1.GetOutput{
			Index: multiraftv1.Index(s.lock.proposalID),
		})
	} else {
		query.Error(errors.NewNotFound("local not held"))
	}
}
