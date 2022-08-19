// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	countermapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/countermap/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.multiraft.atomic.map.v1.CounterMap"

func Register(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput](registry)(MapType)
}

var MapType = statemachine.NewPrimitiveType[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput](Service, mapCodec, newMapStateMachine)

var mapCodec = statemachine.NewCodec[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput](
	func(bytes []byte) (*countermapv1.CounterMapInput, error) {
		input := &countermapv1.CounterMapInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *countermapv1.CounterMapOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newMapStateMachine(ctx statemachine.PrimitiveContext[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput]) statemachine.Primitive[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput] {
	sm := &MapStateMachine{
		PrimitiveContext: ctx,
		listeners:        make(map[statemachine.ProposalID]*countermapv1.CounterMapListener),
		entries:          make(map[string]int64),
		watchers:         make(map[statemachine.QueryID]statemachine.Query[*countermapv1.EntriesInput, *countermapv1.EntriesOutput]),
	}
	sm.init()
	return sm
}

type MapStateMachine struct {
	statemachine.PrimitiveContext[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput]
	listeners map[statemachine.ProposalID]*countermapv1.CounterMapListener
	entries   map[string]int64
	watchers  map[statemachine.QueryID]statemachine.Query[*countermapv1.EntriesInput, *countermapv1.EntriesOutput]
	mu        sync.RWMutex
	put       statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.SetInput, *countermapv1.SetOutput]
	insert    statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.InsertInput, *countermapv1.InsertOutput]
	update    statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.UpdateInput, *countermapv1.UpdateOutput]
	remove    statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.RemoveInput, *countermapv1.RemoveOutput]
	increment statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.IncrementInput, *countermapv1.IncrementOutput]
	decrement statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.DecrementInput, *countermapv1.DecrementOutput]
	clear     statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.ClearInput, *countermapv1.ClearOutput]
	events    statemachine.Updater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.EventsInput, *countermapv1.EventsOutput]
	size      statemachine.Reader[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.SizeInput, *countermapv1.SizeOutput]
	get       statemachine.Reader[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.GetInput, *countermapv1.GetOutput]
	list      statemachine.Reader[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.EntriesInput, *countermapv1.EntriesOutput]
}

func (s *MapStateMachine) init() {
	s.put = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.SetInput, *countermapv1.SetOutput](s).
		Name("Set").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.SetInput, bool) {
			if put, ok := input.Input.(*countermapv1.CounterMapInput_Set); ok {
				return put.Set, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.SetOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Set{
					Set: output,
				},
			}
		}).
		Build(s.doSet)
	s.insert = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.InsertInput, *countermapv1.InsertOutput](s).
		Name("Insert").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.InsertInput, bool) {
			if insert, ok := input.Input.(*countermapv1.CounterMapInput_Insert); ok {
				return insert.Insert, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.InsertOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Insert{
					Insert: output,
				},
			}
		}).
		Build(s.doInsert)
	s.update = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.UpdateInput, *countermapv1.UpdateOutput](s).
		Name("Update").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.UpdateInput, bool) {
			if update, ok := input.Input.(*countermapv1.CounterMapInput_Update); ok {
				return update.Update, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.UpdateOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Update{
					Update: output,
				},
			}
		}).
		Build(s.doUpdate)
	s.remove = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.RemoveInput, *countermapv1.RemoveOutput](s).
		Name("Remove").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.RemoveInput, bool) {
			if remove, ok := input.Input.(*countermapv1.CounterMapInput_Remove); ok {
				return remove.Remove, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.RemoveOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Remove{
					Remove: output,
				},
			}
		}).
		Build(s.doRemove)
	s.increment = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.IncrementInput, *countermapv1.IncrementOutput](s).
		Name("Increment").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.IncrementInput, bool) {
			if remove, ok := input.Input.(*countermapv1.CounterMapInput_Increment); ok {
				return remove.Increment, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.IncrementOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Increment{
					Increment: output,
				},
			}
		}).
		Build(s.doIncrement)
	s.decrement = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.DecrementInput, *countermapv1.DecrementOutput](s).
		Name("Decrement").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.DecrementInput, bool) {
			if remove, ok := input.Input.(*countermapv1.CounterMapInput_Decrement); ok {
				return remove.Decrement, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.DecrementOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Decrement{
					Decrement: output,
				},
			}
		}).
		Build(s.doDecrement)
	s.clear = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.ClearInput, *countermapv1.ClearOutput](s).
		Name("Clear").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.ClearInput, bool) {
			if clear, ok := input.Input.(*countermapv1.CounterMapInput_Clear); ok {
				return clear.Clear, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.ClearOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Clear{
					Clear: output,
				},
			}
		}).
		Build(s.doClear)
	s.events = statemachine.NewUpdater[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.EventsInput, *countermapv1.EventsOutput](s).
		Name("Events").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.EventsInput, bool) {
			if events, ok := input.Input.(*countermapv1.CounterMapInput_Events); ok {
				return events.Events, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.EventsOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Events{
					Events: output,
				},
			}
		}).
		Build(s.doEvents)
	s.size = statemachine.NewReader[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.SizeInput, *countermapv1.SizeOutput](s).
		Name("Size").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.SizeInput, bool) {
			if size, ok := input.Input.(*countermapv1.CounterMapInput_Size_); ok {
				return size.Size_, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.SizeOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Size_{
					Size_: output,
				},
			}
		}).
		Build(s.doSize)
	s.get = statemachine.NewReader[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.GetInput, *countermapv1.GetOutput](s).
		Name("Get").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.GetInput, bool) {
			if get, ok := input.Input.(*countermapv1.CounterMapInput_Get); ok {
				return get.Get, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.GetOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Get{
					Get: output,
				},
			}
		}).
		Build(s.doGet)
	s.list = statemachine.NewReader[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput, *countermapv1.EntriesInput, *countermapv1.EntriesOutput](s).
		Name("Entries").
		Decoder(func(input *countermapv1.CounterMapInput) (*countermapv1.EntriesInput, bool) {
			if entries, ok := input.Input.(*countermapv1.CounterMapInput_Entries); ok {
				return entries.Entries, true
			}
			return nil, false
		}).
		Encoder(func(output *countermapv1.EntriesOutput) *countermapv1.CounterMapOutput {
			return &countermapv1.CounterMapOutput{
				Output: &countermapv1.CounterMapOutput_Entries{
					Entries: output,
				},
			}
		}).
		Build(s.doEntries)
}

func (s *MapStateMachine) Snapshot(writer *snapshot.Writer) error {
	s.Log().Infow("Persisting CounterMap to snapshot")
	if err := writer.WriteVarInt(len(s.listeners)); err != nil {
		return err
	}
	for proposalID, listener := range s.listeners {
		if err := writer.WriteVarUint64(uint64(proposalID)); err != nil {
			return err
		}
		if err := writer.WriteMessage(listener); err != nil {
			return err
		}
	}

	if err := writer.WriteVarInt(len(s.entries)); err != nil {
		return err
	}
	for key, value := range s.entries {
		if err := writer.WriteString(key); err != nil {
			return err
		}
		if err := writer.WriteVarInt64(value); err != nil {
			return err
		}
	}
	return nil
}

func (s *MapStateMachine) Recover(reader *snapshot.Reader) error {
	s.Log().Infow("Recovering CounterMap from snapshot")
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		proposalID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}
		proposal, ok := s.Proposals().Get(statemachine.ProposalID(proposalID))
		if !ok {
			return errors.NewFault("cannot find proposal %d", proposalID)
		}
		listener := &countermapv1.CounterMapListener{}
		if err := reader.ReadMessage(listener); err != nil {
			return err
		}
		s.listeners[proposal.ID()] = listener
		proposal.Watch(func(state statemachine.OperationState) {
			if state == statemachine.Complete {
				delete(s.listeners, proposal.ID())
			}
		})
	}

	n, err = reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		key, err := reader.ReadString()
		if err != nil {
			return err
		}
		value, err := reader.ReadVarInt64()
		if err != nil {
			return err
		}
		s.entries[key] = value
	}
	return nil
}

func (s *MapStateMachine) Update(proposal statemachine.Proposal[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput]) {
	switch proposal.Input().Input.(type) {
	case *countermapv1.CounterMapInput_Set:
		s.put.Update(proposal)
	case *countermapv1.CounterMapInput_Insert:
		s.insert.Update(proposal)
	case *countermapv1.CounterMapInput_Update:
		s.update.Update(proposal)
	case *countermapv1.CounterMapInput_Remove:
		s.remove.Update(proposal)
	case *countermapv1.CounterMapInput_Clear:
		s.clear.Update(proposal)
	case *countermapv1.CounterMapInput_Events:
		s.events.Update(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *MapStateMachine) doSet(proposal statemachine.Proposal[*countermapv1.SetInput, *countermapv1.SetOutput]) {
	defer proposal.Close()

	key := proposal.Input().Key
	newValue := proposal.Input().Value
	oldValue, updated := s.entries[key]
	s.entries[key] = newValue

	// Publish an event to listener streams.
	if updated {
		s.notify(key, newValue, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Updated_{
					Updated: &countermapv1.Event_Updated{
						Value:     newValue,
						PrevValue: oldValue,
					},
				},
			},
		})
		proposal.Output(&countermapv1.SetOutput{
			PrevValue: oldValue,
		})
	} else {
		s.notify(key, newValue, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Inserted_{
					Inserted: &countermapv1.Event_Inserted{
						Value: newValue,
					},
				},
			},
		})
		proposal.Output(&countermapv1.SetOutput{
			PrevValue: oldValue,
		})
	}
}

func (s *MapStateMachine) doInsert(proposal statemachine.Proposal[*countermapv1.InsertInput, *countermapv1.InsertOutput]) {
	defer proposal.Close()

	key := proposal.Input().Key
	value := proposal.Input().Value

	if _, ok := s.entries[key]; ok {
		proposal.Error(errors.NewAlreadyExists("key '%s' already exists", key))
		return
	}

	s.entries[key] = value

	s.notify(key, value, &countermapv1.EventsOutput{
		Event: countermapv1.Event{
			Key: key,
			Event: &countermapv1.Event_Inserted_{
				Inserted: &countermapv1.Event_Inserted{
					Value: value,
				},
			},
		},
	})
	proposal.Output(&countermapv1.InsertOutput{})
}

func (s *MapStateMachine) doUpdate(proposal statemachine.Proposal[*countermapv1.UpdateInput, *countermapv1.UpdateOutput]) {
	defer proposal.Close()

	key := proposal.Input().Key
	newValue := proposal.Input().Value

	oldValue, updated := s.entries[key]
	if !updated {
		proposal.Error(errors.NewNotFound("key '%s' not found", key))
		return
	}

	s.entries[key] = newValue

	// Publish an event to listener streams.
	s.notify(key, newValue, &countermapv1.EventsOutput{
		Event: countermapv1.Event{
			Key: key,
			Event: &countermapv1.Event_Updated_{
				Updated: &countermapv1.Event_Updated{
					Value:     newValue,
					PrevValue: oldValue,
				},
			},
		},
	})
	proposal.Output(&countermapv1.UpdateOutput{
		PrevValue: oldValue,
	})
}

func (s *MapStateMachine) doIncrement(proposal statemachine.Proposal[*countermapv1.IncrementInput, *countermapv1.IncrementOutput]) {
	defer proposal.Close()

	key := proposal.Input().Key
	delta := proposal.Input().Delta
	if delta == 0 {
		delta = 1
	}

	oldValue, updated := s.entries[key]
	if !updated {
		proposal.Error(errors.NewNotFound("key '%s' not found", key))
		return
	}

	newValue := oldValue + delta
	s.entries[key] = newValue

	// Publish an event to listener streams.
	if updated {
		s.notify(key, delta, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Updated_{
					Updated: &countermapv1.Event_Updated{
						Value:     newValue,
						PrevValue: oldValue,
					},
				},
			},
		})
	} else {
		s.notify(key, delta, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Inserted_{
					Inserted: &countermapv1.Event_Inserted{
						Value: newValue,
					},
				},
			},
		})
	}
	proposal.Output(&countermapv1.IncrementOutput{
		PrevValue: oldValue,
	})
}

func (s *MapStateMachine) doDecrement(proposal statemachine.Proposal[*countermapv1.DecrementInput, *countermapv1.DecrementOutput]) {
	defer proposal.Close()

	key := proposal.Input().Key
	delta := proposal.Input().Delta
	if delta == 0 {
		delta = 1
	}

	oldValue, updated := s.entries[key]
	if !updated {
		proposal.Error(errors.NewNotFound("key '%s' not found", key))
		return
	}

	newValue := oldValue - delta
	s.entries[key] = newValue

	// Publish an event to listener streams.
	if updated {
		s.notify(key, delta, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Updated_{
					Updated: &countermapv1.Event_Updated{
						Value:     newValue,
						PrevValue: oldValue,
					},
				},
			},
		})
	} else {
		s.notify(key, delta, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Inserted_{
					Inserted: &countermapv1.Event_Inserted{
						Value: newValue,
					},
				},
			},
		})
	}
	proposal.Output(&countermapv1.DecrementOutput{
		PrevValue: oldValue,
	})
}

func (s *MapStateMachine) doRemove(proposal statemachine.Proposal[*countermapv1.RemoveInput, *countermapv1.RemoveOutput]) {
	defer proposal.Close()

	key := proposal.Input().Key
	value, ok := s.entries[key]
	if !ok {
		proposal.Error(errors.NewNotFound("key '%s' not found", key))
		return
	}

	if proposal.Input().PrevValue != value {
		proposal.Error(errors.NewConflict("entry value %d does not match remove value %d", value, proposal.Input().PrevValue))
		return
	}

	delete(s.entries, key)

	// Publish an event to listener streams.
	s.notify(key, value, &countermapv1.EventsOutput{
		Event: countermapv1.Event{
			Key: key,
			Event: &countermapv1.Event_Removed_{
				Removed: &countermapv1.Event_Removed{
					Value: value,
				},
			},
		},
	})
	proposal.Output(&countermapv1.RemoveOutput{
		Value: value,
	})
}

func (s *MapStateMachine) doClear(proposal statemachine.Proposal[*countermapv1.ClearInput, *countermapv1.ClearOutput]) {
	defer proposal.Close()
	for key, value := range s.entries {
		s.notify(key, value, &countermapv1.EventsOutput{
			Event: countermapv1.Event{
				Key: key,
				Event: &countermapv1.Event_Removed_{
					Removed: &countermapv1.Event_Removed{
						Value: value,
					},
				},
			},
		})
		delete(s.entries, key)
	}
	proposal.Output(&countermapv1.ClearOutput{})
}

func (s *MapStateMachine) doEvents(proposal statemachine.Proposal[*countermapv1.EventsInput, *countermapv1.EventsOutput]) {
	// Output an empty event to ack the request
	proposal.Output(&countermapv1.EventsOutput{})

	listener := &countermapv1.CounterMapListener{
		Key: proposal.Input().Key,
	}
	s.listeners[proposal.ID()] = listener
	proposal.Watch(func(state statemachine.OperationState) {
		if state == statemachine.Complete {
			delete(s.listeners, proposal.ID())
		}
	})
}

func (s *MapStateMachine) Read(query statemachine.Query[*countermapv1.CounterMapInput, *countermapv1.CounterMapOutput]) {
	switch query.Input().Input.(type) {
	case *countermapv1.CounterMapInput_Size_:
		s.size.Read(query)
	case *countermapv1.CounterMapInput_Get:
		s.get.Read(query)
	case *countermapv1.CounterMapInput_Entries:
		s.list.Read(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *MapStateMachine) doSize(query statemachine.Query[*countermapv1.SizeInput, *countermapv1.SizeOutput]) {
	defer query.Close()
	query.Output(&countermapv1.SizeOutput{
		Size_: uint32(len(s.entries)),
	})
}

func (s *MapStateMachine) doGet(query statemachine.Query[*countermapv1.GetInput, *countermapv1.GetOutput]) {
	defer query.Close()
	value, ok := s.entries[query.Input().Key]
	if !ok {
		query.Error(errors.NewNotFound("key %s not found", query.Input().Key))
	} else {
		query.Output(&countermapv1.GetOutput{
			Value: value,
		})
	}
}

func (s *MapStateMachine) doEntries(query statemachine.Query[*countermapv1.EntriesInput, *countermapv1.EntriesOutput]) {
	for key, value := range s.entries {
		query.Output(&countermapv1.EntriesOutput{
			Entry: countermapv1.Entry{
				Key:   key,
				Value: value,
			},
		})
	}

	if query.Input().Watch {
		s.mu.Lock()
		s.watchers[query.ID()] = query
		s.mu.Unlock()
		query.Watch(func(state statemachine.OperationState) {
			if state == statemachine.Complete {
				s.mu.Lock()
				delete(s.watchers, query.ID())
				s.mu.Unlock()
			}
		})
	} else {
		query.Close()
	}
}

func (s *MapStateMachine) notify(key string, value int64, event *countermapv1.EventsOutput) {
	for proposalID, listener := range s.listeners {
		if listener.Key == "" || listener.Key == event.Event.Key {
			proposal, ok := s.events.Proposals().Get(proposalID)
			if ok {
				proposal.Output(event)
			} else {
				delete(s.listeners, proposalID)
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, watcher := range s.watchers {
		watcher.Output(&countermapv1.EntriesOutput{
			Entry: countermapv1.Entry{
				Key:   key,
				Value: value,
			},
		})
	}
}
