// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	counterv1 "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
)

func newPrimitiveStateMachine(context Context) StateMachine {
	return &counterService{
		Context: context,
	}
}

// counterService is a state machine for a counter primitive
type counterService struct {
	Context
	value int64
}

func (c *counterService) Backup(writer *snapshot.Writer) error {
	return writer.WriteVarInt64(c.value)
}

func (c *counterService) Restore(reader *snapshot.Reader) error {
	i, err := reader.ReadVarInt64()
	if err != nil {
		return err
	}
	c.value = i
	return nil
}

func (c *counterService) Set(set SetProposal) (*counterv1.SetOutput, error) {
	c.value = set.Input().Value
	return &counterv1.SetOutput{
		Value: c.value,
	}, nil
}

func (c *counterService) CompareAndSet(set CompareAndSetProposal) (*counterv1.CompareAndSetOutput, error) {
	if c.value != set.Input().Compare {
		return nil, errors.NewConflict("optimistic lock failure")
	}
	c.value = set.Input().Update
	return &counterv1.CompareAndSetOutput{
		Value: c.value,
	}, nil
}

func (c *counterService) Get(GetQuery) (*counterv1.GetOutput, error) {
	return &counterv1.GetOutput{
		Value: c.value,
	}, nil
}

func (c *counterService) Increment(increment IncrementProposal) (*counterv1.IncrementOutput, error) {
	c.value += increment.Input().Delta
	return &counterv1.IncrementOutput{
		Value: c.value,
	}, nil
}

func (c *counterService) Decrement(decrement DecrementProposal) (*counterv1.DecrementOutput, error) {
	c.value -= decrement.Input().Delta
	return &counterv1.DecrementOutput{
		Value: c.value,
	}, nil
}
