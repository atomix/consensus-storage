// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	lockv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/lock/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"time"
)

const Service = "atomix.runtime.lock.v1.Lock"

func Register(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*lockv1.LockInput, *lockv1.LockOutput](registry)(Type)
}

var Type = statemachine.NewPrimitiveType[*lockv1.LockInput, *lockv1.LockOutput](Service, lockCodec, newLockStateMachine)

var lockCodec = statemachine.NewCodec[*lockv1.LockInput, *lockv1.LockOutput](
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

func newLockStateMachine(ctx statemachine.PrimitiveContext[*lockv1.LockInput, *lockv1.LockOutput]) statemachine.Primitive[*lockv1.LockInput, *lockv1.LockOutput] {
	sm := &LockStateMachine{
		PrimitiveContext: ctx,
		proposals:        make(map[statemachine.ProposalID]statemachine.CancelFunc),
		sessions:         make(map[statemachine.SessionID]statemachine.CancelFunc),
	}
	sm.init()
	return sm
}

type Waiter struct {
	statemachine.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]
	expire *time.Time
}

type LockStateMachine struct {
	statemachine.PrimitiveContext[*lockv1.LockInput, *lockv1.LockOutput]
	lock      statemachine.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]
	queue     []statemachine.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]
	proposals map[statemachine.ProposalID]statemachine.CancelFunc
	sessions  map[statemachine.SessionID]statemachine.CancelFunc
	acquire   statemachine.Updater[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.AcquireInput, *lockv1.AcquireOutput]
	release   statemachine.Updater[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.ReleaseInput, *lockv1.ReleaseOutput]
	get       statemachine.Reader[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.GetInput, *lockv1.GetOutput]
}

func (s *LockStateMachine) init() {
	s.acquire = statemachine.NewUpdater[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.AcquireInput, *lockv1.AcquireOutput](s).
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
	s.release = statemachine.NewUpdater[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.ReleaseInput, *lockv1.ReleaseOutput](s).
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
	s.get = statemachine.NewReader[*lockv1.LockInput, *lockv1.LockOutput, *lockv1.GetInput, *lockv1.GetOutput](s).
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
		if err := writer.WriteVarUint64(uint64(s.lock.ID())); err != nil {
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
		proposal, ok := s.acquire.Proposals().Get(statemachine.ProposalID(proposalID))
		if !ok {
			return errors.NewFault("proposal not found")
		}
		s.lock = proposal

		n, err := reader.ReadVarInt()
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			proposalID, err := reader.ReadVarUint64()
			if err != nil {
				return err
			}
			proposal, ok := s.acquire.Proposals().Get(statemachine.ProposalID(proposalID))
			if !ok {
				return errors.NewFault("proposal not found")
			}
			s.queue = append(s.queue, proposal)
		}
	}
	return nil
}

func (s *LockStateMachine) Update(proposal statemachine.Proposal[*lockv1.LockInput, *lockv1.LockOutput]) {
	switch proposal.Input().Input.(type) {
	case *lockv1.LockInput_Acquire:
		s.acquire.Update(proposal)
	case *lockv1.LockInput_Release:
		s.release.Update(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
	}
}

func (s *LockStateMachine) doAcquire(proposal statemachine.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]) {
	if s.lock == nil {
		defer proposal.Close()
		s.lock = proposal
		proposal.Output(&lockv1.AcquireOutput{
			Index: multiraftv1.Index(proposal.ID()),
		})
	} else {
		s.proposals[proposal.ID()] = proposal.Watch(func(state statemachine.OperationState) {
			if state == statemachine.Canceled {
				for i, waiter := range s.queue {
					if waiter.ID() == proposal.ID() {
						s.queue = append(s.queue[:i], s.queue[i+1:]...)
						break
					}
				}
			}
			delete(s.proposals, proposal.ID())
		})
		if _, ok := s.sessions[proposal.Session().ID()]; !ok {
			s.sessions[proposal.Session().ID()] = proposal.Session().Watch(func(state statemachine.SessionState) {
				if state == statemachine.SessionClosed {
					var queue []statemachine.Proposal[*lockv1.AcquireInput, *lockv1.AcquireOutput]
					for _, waiter := range s.queue {
						if waiter.Session().ID() == proposal.Session().ID() {
							queue = append(queue, waiter)
						}
					}
					s.queue = queue
					if s.lock != nil && s.lock.Session().ID() == proposal.Session().ID() {
						s.lock = nil
						if s.queue != nil {
							s.lock = s.queue[0]
							s.queue = s.queue[1:]
							s.lock.Output(&lockv1.AcquireOutput{
								Index: multiraftv1.Index(s.lock.ID()),
							})
							s.lock.Close()
						}
					}
				}
				delete(s.sessions, proposal.Session().ID())
			})
		}
		s.queue = append(s.queue, proposal)
	}
}

func (s *LockStateMachine) doRelease(proposal statemachine.Proposal[*lockv1.ReleaseInput, *lockv1.ReleaseOutput]) {
	defer proposal.Close()
	if s.lock == nil {
		proposal.Error(errors.NewConflict("lock not held by client"))
		return
	}

	if multiraftv1.Index(s.lock.ID()) != proposal.Input().Index {
		proposal.Error(errors.NewConflict("lock not held by client"))
		return
	}

	s.lock = nil
	if s.queue != nil {
		s.lock = s.queue[0]
		s.queue = s.queue[1:]
		s.lock.Output(&lockv1.AcquireOutput{
			Index: multiraftv1.Index(s.lock.ID()),
		})
		s.lock.Close()
	}
}

func (s *LockStateMachine) Read(query statemachine.Query[*lockv1.LockInput, *lockv1.LockOutput]) {
	switch query.Input().Input.(type) {
	case *lockv1.LockInput_Get:
		s.get.Read(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *LockStateMachine) doGet(query statemachine.Query[*lockv1.GetInput, *lockv1.GetOutput]) {
	defer query.Close()
	if s.lock != nil {
		query.Output(&lockv1.GetOutput{
			Index: multiraftv1.Index(s.lock.ID()),
		})
	} else {
		query.Error(errors.NewNotFound("local not held"))
	}
}
