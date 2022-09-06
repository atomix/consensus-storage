// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
)

func NewStateMachine(ctx statemachine.PrimitiveContext[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput]) statemachine.StateMachine[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput] {
	return &StateMachine{
		PrimitiveContext: ctx,
		...
	}
}

type StateMachine struct {
	statemachine.PrimitiveContext[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput]
	create statemachine.Updater[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]
	close  statemachine.Updater[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]
	update statemachine.Updater[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput]
	read   statemachine.Reader[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput]
}

func (s *StateMachine) init() {
	s.create = statemachine.NewUpdater[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput](s).
		Name("CreatePrimitive").
		Decoder(func(input *multiraftv1.StateMachineInput) (*multiraftv1.CreatePrimitiveInput, bool) {
			if create, ok := input.Input.(*multiraftv1.StateMachineInput_CreatePrimitive); ok {
				return create.CreatePrimitive, true
			}
			return nil, false
		}).
		Encoder(func(output *multiraftv1.CreatePrimitiveOutput) *multiraftv1.StateMachineOutput {
			return &multiraftv1.StateMachineOutput{
				Output: &multiraftv1.StateMachineOutput_CreatePrimitive{
					CreatePrimitive: output,
				},
			}
		}).
		Build(s.doCreate)
	s.close = statemachine.NewUpdater[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput](s).
		Name("ClosePrimitive").
		Decoder(func(input *multiraftv1.StateMachineInput) (*multiraftv1.ClosePrimitiveInput, bool) {
			if close, ok := input.Input.(*multiraftv1.StateMachineInput_ClosePrimitive); ok {
				return close.ClosePrimitive, true
			}
			return nil, false
		}).
		Encoder(func(output *multiraftv1.ClosePrimitiveOutput) *multiraftv1.StateMachineOutput {
			return &multiraftv1.StateMachineOutput{
				Output: &multiraftv1.StateMachineOutput_ClosePrimitive{
					ClosePrimitive: output,
				},
			}
		}).
		Build(s.doClose)
	s.update = statemachine.NewUpdater[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput](s).
		Name("Update").
		Decoder(func(input *multiraftv1.StateMachineInput) (*multiraftv1.PrimitiveOperationInput, bool) {
			if update, ok := input.Input.(*multiraftv1.StateMachineInput_Operation); ok {
				return update.Operation, true
			}
			return nil, false
		}).
		Encoder(func(output *multiraftv1.PrimitiveOperationOutput) *multiraftv1.StateMachineOutput {
			return &multiraftv1.StateMachineOutput{
				Output: &multiraftv1.StateMachineOutput_Operation{
					Operation: output,
				},
			}
		}).
		Build(s.doUpdate)
	s.read = statemachine.NewReader[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput, *multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput](s).
		Name("Read").
		Decoder(func(input *multiraftv1.StateMachineInput) (*multiraftv1.PrimitiveOperationInput, bool) {
			if update, ok := input.Input.(*multiraftv1.StateMachineInput_Operation); ok {
				return update.Operation, true
			}
			return nil, false
		}).
		Encoder(func(output *multiraftv1.PrimitiveOperationOutput) *multiraftv1.StateMachineOutput {
			return &multiraftv1.StateMachineOutput{
				Output: &multiraftv1.StateMachineOutput_Operation{
					Operation: output,
				},
			}
		}).
		Build(s.doRead)
}

func (s *StateMachine) Snapshot(writer *snapshot.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (s *StateMachine) Recover(reader *snapshot.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (s *StateMachine) Update(proposal statemachine.Proposal[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput]) {
	switch proposal.Input().Input.(type) {
	case *multiraftv1.StateMachineInput_CreatePrimitive:
		s.create.Update(proposal)
	case *multiraftv1.StateMachineInput_ClosePrimitive:
		s.close.Update(proposal)
	case *multiraftv1.StateMachineInput_Operation:
		s.update.Update(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *StateMachine) doCreate(proposal statemachine.Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]) {

}

func (s *StateMachine) doClose(proposal statemachine.Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]) {

}

func (s *StateMachine) doUpdate(proposal statemachine.Proposal[*multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput]) {

}

func (s *StateMachine) Read(query statemachine.Query[*multiraftv1.StateMachineInput, *multiraftv1.StateMachineOutput]) {
	switch query.Input().Input.(type) {
	case *multiraftv1.StateMachineInput_Operation:
		s.read.Read(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
		query.Close()
	}
}

func (s *StateMachine) doRead(query statemachine.Query[*multiraftv1.PrimitiveOperationInput, *multiraftv1.PrimitiveOperationOutput]) {

}
