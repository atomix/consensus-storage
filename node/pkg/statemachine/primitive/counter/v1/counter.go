// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

const Service = "atomix.runtime.counter.v1.Counter"

func Register(registry *primitive.TypeRegistry) {
	primitive.RegisterType[*counterv1.CounterInput, *counterv1.CounterOutput](registry)(Type)
}

var Type = primitive.NewType[*counterv1.CounterInput, *counterv1.CounterOutput](Service, counterCodec, newCounterStateMachine)

var counterCodec = primitive.NewCodec[*counterv1.CounterInput, *counterv1.CounterOutput](
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

func newCounterStateMachine(ctx primitive.PrimitiveContext[*counterv1.CounterInput, *counterv1.CounterOutput]) primitive.Primitive[*counterv1.CounterInput, *counterv1.CounterOutput] {
	sm := &CounterStateMachine{
		PrimitiveContext: ctx,
	}
	sm.init()
	return sm
}

type CounterStateMachine struct {
	primitive.PrimitiveContext[*counterv1.CounterInput, *counterv1.CounterOutput]
	value     int64
	set       primitive.Proposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.SetInput, *counterv1.SetOutput]
	update    primitive.Proposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.UpdateInput, *counterv1.UpdateOutput]
	increment primitive.Proposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.IncrementInput, *counterv1.IncrementOutput]
	decrement primitive.Proposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.DecrementInput, *counterv1.DecrementOutput]
	get       primitive.Querier[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.GetInput, *counterv1.GetOutput]
}

func (s *CounterStateMachine) init() {
	s.set = primitive.NewProposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.SetInput, *counterv1.SetOutput](s).
		Name("Set").
		Decoder(func(input *counterv1.CounterInput) (*counterv1.SetInput, bool) {
			if set, ok := input.Input.(*counterv1.CounterInput_Set); ok {
				return set.Set, true
			}
			return nil, false
		}).
		Encoder(func(output *counterv1.SetOutput) *counterv1.CounterOutput {
			return &counterv1.CounterOutput{
				Output: &counterv1.CounterOutput_Set{
					Set: output,
				},
			}
		}).
		Build(s.doSet)
	s.update = primitive.NewProposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.UpdateInput, *counterv1.UpdateOutput](s).
		Name("Update").
		Decoder(func(input *counterv1.CounterInput) (*counterv1.UpdateInput, bool) {
			if set, ok := input.Input.(*counterv1.CounterInput_Update); ok {
				return set.Update, true
			}
			return nil, false
		}).
		Encoder(func(output *counterv1.UpdateOutput) *counterv1.CounterOutput {
			return &counterv1.CounterOutput{
				Output: &counterv1.CounterOutput_Update{
					Update: output,
				},
			}
		}).
		Build(s.doUpdate)
	s.increment = primitive.NewProposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.IncrementInput, *counterv1.IncrementOutput](s).
		Name("Increment").
		Decoder(func(input *counterv1.CounterInput) (*counterv1.IncrementInput, bool) {
			if set, ok := input.Input.(*counterv1.CounterInput_Increment); ok {
				return set.Increment, true
			}
			return nil, false
		}).
		Encoder(func(output *counterv1.IncrementOutput) *counterv1.CounterOutput {
			return &counterv1.CounterOutput{
				Output: &counterv1.CounterOutput_Increment{
					Increment: output,
				},
			}
		}).
		Build(s.doIncrement)
	s.decrement = primitive.NewProposer[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.DecrementInput, *counterv1.DecrementOutput](s).
		Name("Decrement").
		Decoder(func(input *counterv1.CounterInput) (*counterv1.DecrementInput, bool) {
			if set, ok := input.Input.(*counterv1.CounterInput_Decrement); ok {
				return set.Decrement, true
			}
			return nil, false
		}).
		Encoder(func(output *counterv1.DecrementOutput) *counterv1.CounterOutput {
			return &counterv1.CounterOutput{
				Output: &counterv1.CounterOutput_Decrement{
					Decrement: output,
				},
			}
		}).
		Build(s.doDecrement)
	s.get = primitive.NewQuerier[*counterv1.CounterInput, *counterv1.CounterOutput, *counterv1.GetInput, *counterv1.GetOutput](s).
		Name("Get").
		Decoder(func(input *counterv1.CounterInput) (*counterv1.GetInput, bool) {
			if set, ok := input.Input.(*counterv1.CounterInput_Get); ok {
				return set.Get, true
			}
			return nil, false
		}).
		Encoder(func(output *counterv1.GetOutput) *counterv1.CounterOutput {
			return &counterv1.CounterOutput{
				Output: &counterv1.CounterOutput_Get{
					Get: output,
				},
			}
		}).
		Build(s.doGet)
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

func (s *CounterStateMachine) Propose(proposal primitive.Proposal[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	switch proposal.Input().Input.(type) {
	case *counterv1.CounterInput_Set:
		s.set.Execute(proposal)
	case *counterv1.CounterInput_Update:
		s.update.Execute(proposal)
	case *counterv1.CounterInput_Increment:
		s.increment.Execute(proposal)
	case *counterv1.CounterInput_Decrement:
		s.decrement.Execute(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
	}
}

func (s *CounterStateMachine) doSet(proposal primitive.Proposal[*counterv1.SetInput, *counterv1.SetOutput]) {
	defer proposal.Close()
	s.value = proposal.Input().Value
	proposal.Output(&counterv1.SetOutput{
		Value: s.value,
	})
}

func (s *CounterStateMachine) doUpdate(proposal primitive.Proposal[*counterv1.UpdateInput, *counterv1.UpdateOutput]) {
	defer proposal.Close()
	if s.value != proposal.Input().Compare {
		proposal.Error(errors.NewConflict("optimistic lock failure"))
	} else {
		s.value = proposal.Input().Update
		proposal.Output(&counterv1.UpdateOutput{
			Value: s.value,
		})
	}
}

func (s *CounterStateMachine) doIncrement(proposal primitive.Proposal[*counterv1.IncrementInput, *counterv1.IncrementOutput]) {
	defer proposal.Close()
	s.value += proposal.Input().Delta
	proposal.Output(&counterv1.IncrementOutput{
		Value: s.value,
	})
}

func (s *CounterStateMachine) doDecrement(proposal primitive.Proposal[*counterv1.DecrementInput, *counterv1.DecrementOutput]) {
	defer proposal.Close()
	s.value -= proposal.Input().Delta
	proposal.Output(&counterv1.DecrementOutput{
		Value: s.value,
	})
}

func (s *CounterStateMachine) Query(query primitive.Query[*counterv1.CounterInput, *counterv1.CounterOutput]) {
	switch query.Input().Input.(type) {
	case *counterv1.CounterInput_Get:
		s.get.Execute(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *CounterStateMachine) doGet(query primitive.Query[*counterv1.GetInput, *counterv1.GetOutput]) {
	defer query.Close()
	query.Output(&counterv1.GetOutput{
		Value: s.value,
	})
}
