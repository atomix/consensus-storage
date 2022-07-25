// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitives

import (
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	statemachine "github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

func RegisterCounterType(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*counterv1.CounterInput, *counterv1.CounterOutput](registry)(CounterType)
}

var CounterType = statemachine.NewPrimitiveType[*counterv1.CounterInput, *counterv1.CounterOutput]("Counter", "v1", counterCodec, newStateMachine)

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

func newStateMachine(ctx statemachine.PrimitiveContext[*counterv1.CounterInput, *counterv1.CounterOutput]) statemachine.Primitive[*counterv1.CounterInput, *counterv1.CounterOutput] {
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

func (s *CounterStateMachine) Update(command statemachine.Command[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	switch command.Input().Input.(type) {
	case *counterv1.CounterInput_Set:
		s.set(command)
	case *counterv1.CounterInput_CompareAndSet:
		s.compareAndSet(command)
	case *counterv1.CounterInput_Increment:
		s.increment(command)
	case *counterv1.CounterInput_Decrement:
		s.decrement(command)
	default:
		command.Error(errors.NewNotSupported("command not supported"))
	}
}

func (s *CounterStateMachine) set(command statemachine.Command[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer command.Close()
	s.value = command.Input().GetSet().Value
	command.Output(&counterv1.CounterOutput{
		Output: &counterv1.CounterOutput_Set{
			Set: &counterv1.SetOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) compareAndSet(command statemachine.Command[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer command.Close()
	if s.value != command.Input().GetCompareAndSet().Compare {
		command.Error(errors.NewConflict("optimistic lock failure"))
	} else {
		s.value = command.Input().GetCompareAndSet().Update
		command.Output(&counterv1.CounterOutput{
			Output: &counterv1.CounterOutput_CompareAndSet{
				CompareAndSet: &counterv1.CompareAndSetOutput{
					Value: s.value,
				},
			},
		})
	}
}

func (s *CounterStateMachine) increment(command statemachine.Command[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer command.Close()
	s.value += command.Input().GetIncrement().Delta
	command.Output(&counterv1.CounterOutput{
		Output: &counterv1.CounterOutput_Increment{
			Increment: &counterv1.IncrementOutput{
				Value: s.value,
			},
		},
	})
}

func (s *CounterStateMachine) decrement(command statemachine.Command[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	defer command.Close()
	s.value -= command.Input().GetDecrement().Delta
	command.Output(&counterv1.CounterOutput{
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
