// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/counter/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

const Service = "atomix.multiraft.atomic.counter.v1.AtomicCounter"

func Register(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput](registry)(CounterType)
}

var CounterType = statemachine.NewPrimitiveType[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput](Service, counterCodec, newCounterStateMachine)

var counterCodec = statemachine.NewCodec[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput](
	func(bytes []byte) (*counterv1.AtomicCounterInput, error) {
		input := &counterv1.AtomicCounterInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *counterv1.AtomicCounterOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newCounterStateMachine(ctx statemachine.PrimitiveContext[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) statemachine.Primitive[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput] {
	return &CounterStateMachine{
		PrimitiveContext: ctx,
	}
}

type CounterStateMachine struct {
	statemachine.PrimitiveContext[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]
	value int64
}

func (s *CounterStateMachine) Snapshot(writer *snapshot.Writer) error {
	return writer.WriteVarInt64(s.value)
}

func (s *CounterStateMachine) Recover(reader *snapshot.Reader) error {
	i, err := reader.ReadVarInt64()
	if err != nil {
		return err
	}
	s.value = i
	return nil
}

func (s *CounterStateMachine) Update(proposal statemachine.Proposal[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	switch proposal.Input().Input.(type) {
	case *counterv1.AtomicCounterInput_Set:
		s.set(proposal)
	case *counterv1.AtomicCounterInput_Update:
		s.update(proposal)
	case *counterv1.AtomicCounterInput_Increment:
		s.increment(proposal)
	case *counterv1.AtomicCounterInput_Decrement:
		s.decrement(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
	}
}

func (s *CounterStateMachine) set(proposal statemachine.Proposal[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	defer proposal.Close()
	s.value = proposal.Input().GetSet().Value
	proposal.Output(&counterv1.AtomicCounterOutput{
		Output: &counterv1.AtomicCounterOutput_Set{
			Set: &counterv1.SetOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) update(proposal statemachine.Proposal[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	defer proposal.Close()
	if s.value != proposal.Input().GetUpdate().Compare {
		proposal.Error(errors.NewConflict("optimistic lock failure"))
	} else {
		s.value = proposal.Input().GetUpdate().Update
		proposal.Output(&counterv1.AtomicCounterOutput{
			Output: &counterv1.AtomicCounterOutput_Update{
				Update: &counterv1.UpdateOutput{
					Value: s.value,
				},
			},
		})
	}
}

func (s *CounterStateMachine) increment(proposal statemachine.Proposal[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	defer proposal.Close()
	s.value += proposal.Input().GetIncrement().Delta
	proposal.Output(&counterv1.AtomicCounterOutput{
		Output: &counterv1.AtomicCounterOutput_Increment{
			Increment: &counterv1.IncrementOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) decrement(proposal statemachine.Proposal[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	defer proposal.Close()
	s.value -= proposal.Input().GetDecrement().Delta
	proposal.Output(&counterv1.AtomicCounterOutput{
		Output: &counterv1.AtomicCounterOutput_Decrement{
			Decrement: &counterv1.DecrementOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) Read(query statemachine.Query[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	switch query.Input().Input.(type) {
	case *counterv1.AtomicCounterInput_Get:
		s.get(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *CounterStateMachine) get(query statemachine.Query[*counterv1.AtomicCounterInput, *counterv1.AtomicCounterOutput]) {
	query.Output(&counterv1.AtomicCounterOutput{
		Output: &counterv1.AtomicCounterOutput_Get{
			Get: &counterv1.GetOutput{
				Value: s.value,
			},
		},
	})
}
