// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitives

import (
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

func RegisterCounterType(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*counterv1.CounterInput, *counterv1.CounterOutput](registry)(CounterType)
}

var CounterType = statemachine.NewPrimitiveType[*counterv1.CounterInput, *counterv1.CounterOutput]("Counter", "v1", counterCodec, newCounterStateMachine)

var counterCodec = statemachine.NewCodec[*counterv1.CounterInput, *counterv1.CounterOutput](
	func(bytes []byte) (*counterv1.CounterInput, error) {
		input := &counterv1.CounterInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *counterv1.CounterOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newCounterStateMachine(ctx statemachine.PrimitiveContext[*counterv1.CounterInput, *counterv1.CounterOutput]) statemachine.Primitive[*counterv1.CounterInput, *counterv1.CounterOutput] {
	return &CounterStateMachine{
		PrimitiveContext: ctx,
	}
}

type CounterStateMachine struct {
	statemachine.PrimitiveContext[*counterv1.CounterInput, *counterv1.CounterOutput]
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

func (s *CounterStateMachine) Update(proposal statemachine.Proposal[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	switch proposal.Input().Input.(type) {
	case *counterv1.CounterInput_Set:
		s.set(proposal)
	case *counterv1.CounterInput_CompareAndSet:
		s.compareAndSet(proposal)
	case *counterv1.CounterInput_Increment:
		s.increment(proposal)
	case *counterv1.CounterInput_Decrement:
		s.decrement(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
	}
}

func (s *CounterStateMachine) set(proposal statemachine.Proposal[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer proposal.Close()
	s.value = proposal.Input().GetSet().Value
	proposal.Output(&counterv1.CounterOutput{
		Output: &counterv1.CounterOutput_Set{
			Set: &counterv1.SetOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) compareAndSet(proposal statemachine.Proposal[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer proposal.Close()
	if s.value != proposal.Input().GetCompareAndSet().Compare {
		proposal.Error(errors.NewConflict("optimistic lock failure"))
	} else {
		s.value = proposal.Input().GetCompareAndSet().Update
		proposal.Output(&counterv1.CounterOutput{
			Output: &counterv1.CounterOutput_CompareAndSet{
				CompareAndSet: &counterv1.CompareAndSetOutput{
					Value: s.value,
				},
			},
		})
	}
}

func (s *CounterStateMachine) increment(proposal statemachine.Proposal[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer proposal.Close()
	s.value += proposal.Input().GetIncrement().Delta
	proposal.Output(&counterv1.CounterOutput{
		Output: &counterv1.CounterOutput_Increment{
			Increment: &counterv1.IncrementOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) decrement(proposal statemachine.Proposal[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer proposal.Close()
	s.value -= proposal.Input().GetDecrement().Delta
	proposal.Output(&counterv1.CounterOutput{
		Output: &counterv1.CounterOutput_Decrement{
			Decrement: &counterv1.DecrementOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) Read(query statemachine.Query[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	switch query.Input().Input.(type) {
	case *counterv1.CounterInput_Get:
		s.get(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *CounterStateMachine) get(query statemachine.Query[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	query.Output(&counterv1.CounterOutput{
		Output: &counterv1.CounterOutput_Get{
			Get: &counterv1.GetOutput{
				Value: s.value,
			},
		},
	})
}
