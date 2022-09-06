// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	mapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.runtime.map.v1.Map"

func Register(registry *primitive.TypeRegistry) {
	primitive.RegisterType[*mapv1.MapInput, *mapv1.MapOutput](registry)(Type)
}

var Type = primitive.NewType[*mapv1.MapInput, *mapv1.MapOutput](Service, mapCodec, newMapStateMachine)

var mapCodec = primitive.NewCodec[*mapv1.MapInput, *mapv1.MapOutput](
	func(bytes []byte) (*mapv1.MapInput, error) {
		input := &mapv1.MapInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *mapv1.MapOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newMapStateMachine(ctx primitive.PrimitiveContext[*mapv1.MapInput, *mapv1.MapOutput]) primitive.Primitive[*mapv1.MapInput, *mapv1.MapOutput] {
	sm := &MapStateMachine{
		PrimitiveContext: ctx,
		listeners:        make(map[primitive.ProposalID]*mapv1.MapListener),
		entries:          make(map[string]*mapv1.MapEntry),
		timers:           make(map[string]statemachine.Timer),
		watchers:         make(map[primitive.QueryID]primitive.Query[*mapv1.EntriesInput, *mapv1.EntriesOutput]),
	}
	sm.init()
	return sm
}

type MapStateMachine struct {
	primitive.PrimitiveContext[*mapv1.MapInput, *mapv1.MapOutput]
	listeners map[primitive.ProposalID]*mapv1.MapListener
	entries   map[string]*mapv1.MapEntry
	timers    map[string]statemachine.Timer
	watchers  map[primitive.QueryID]primitive.Query[*mapv1.EntriesInput, *mapv1.EntriesOutput]
	mu        sync.RWMutex
	put       primitive.Proposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.PutInput, *mapv1.PutOutput]
	insert    primitive.Proposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.InsertInput, *mapv1.InsertOutput]
	update    primitive.Proposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.UpdateInput, *mapv1.UpdateOutput]
	remove    primitive.Proposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.RemoveInput, *mapv1.RemoveOutput]
	clear     primitive.Proposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.ClearInput, *mapv1.ClearOutput]
	events    primitive.Proposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.EventsInput, *mapv1.EventsOutput]
	size      primitive.Querier[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.SizeInput, *mapv1.SizeOutput]
	get       primitive.Querier[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.GetInput, *mapv1.GetOutput]
	list      primitive.Querier[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.EntriesInput, *mapv1.EntriesOutput]
}

func (s *MapStateMachine) init() {
	s.put = primitive.NewProposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.PutInput, *mapv1.PutOutput](s).
		Name("Put").
		Decoder(func(input *mapv1.MapInput) (*mapv1.PutInput, bool) {
			if put, ok := input.Input.(*mapv1.MapInput_Put); ok {
				return put.Put, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.PutOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Put{
					Put: output,
				},
			}
		}).
		Build(s.doPut)
	s.insert = primitive.NewProposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.InsertInput, *mapv1.InsertOutput](s).
		Name("Insert").
		Decoder(func(input *mapv1.MapInput) (*mapv1.InsertInput, bool) {
			if insert, ok := input.Input.(*mapv1.MapInput_Insert); ok {
				return insert.Insert, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.InsertOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Insert{
					Insert: output,
				},
			}
		}).
		Build(s.doInsert)
	s.update = primitive.NewProposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.UpdateInput, *mapv1.UpdateOutput](s).
		Name("Update").
		Decoder(func(input *mapv1.MapInput) (*mapv1.UpdateInput, bool) {
			if update, ok := input.Input.(*mapv1.MapInput_Update); ok {
				return update.Update, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.UpdateOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Update{
					Update: output,
				},
			}
		}).
		Build(s.doUpdate)
	s.remove = primitive.NewProposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.RemoveInput, *mapv1.RemoveOutput](s).
		Name("Remove").
		Decoder(func(input *mapv1.MapInput) (*mapv1.RemoveInput, bool) {
			if remove, ok := input.Input.(*mapv1.MapInput_Remove); ok {
				return remove.Remove, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.RemoveOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Remove{
					Remove: output,
				},
			}
		}).
		Build(s.doRemove)
	s.clear = primitive.NewProposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.ClearInput, *mapv1.ClearOutput](s).
		Name("Clear").
		Decoder(func(input *mapv1.MapInput) (*mapv1.ClearInput, bool) {
			if clear, ok := input.Input.(*mapv1.MapInput_Clear); ok {
				return clear.Clear, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.ClearOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Clear{
					Clear: output,
				},
			}
		}).
		Build(s.doClear)
	s.events = primitive.NewProposer[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.EventsInput, *mapv1.EventsOutput](s).
		Name("Events").
		Decoder(func(input *mapv1.MapInput) (*mapv1.EventsInput, bool) {
			if events, ok := input.Input.(*mapv1.MapInput_Events); ok {
				return events.Events, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.EventsOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Events{
					Events: output,
				},
			}
		}).
		Build(s.doEvents)
	s.size = primitive.NewQuerier[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.SizeInput, *mapv1.SizeOutput](s).
		Name("Size").
		Decoder(func(input *mapv1.MapInput) (*mapv1.SizeInput, bool) {
			if size, ok := input.Input.(*mapv1.MapInput_Size_); ok {
				return size.Size_, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.SizeOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Size_{
					Size_: output,
				},
			}
		}).
		Build(s.doSize)
	s.get = primitive.NewQuerier[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.GetInput, *mapv1.GetOutput](s).
		Name("Get").
		Decoder(func(input *mapv1.MapInput) (*mapv1.GetInput, bool) {
			if get, ok := input.Input.(*mapv1.MapInput_Get); ok {
				return get.Get, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.GetOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Get{
					Get: output,
				},
			}
		}).
		Build(s.doGet)
	s.list = primitive.NewQuerier[*mapv1.MapInput, *mapv1.MapOutput, *mapv1.EntriesInput, *mapv1.EntriesOutput](s).
		Name("Entries").
		Decoder(func(input *mapv1.MapInput) (*mapv1.EntriesInput, bool) {
			if entries, ok := input.Input.(*mapv1.MapInput_Entries); ok {
				return entries.Entries, true
			}
			return nil, false
		}).
		Encoder(func(output *mapv1.EntriesOutput) *mapv1.MapOutput {
			return &mapv1.MapOutput{
				Output: &mapv1.MapOutput_Entries{
					Entries: output,
				},
			}
		}).
		Build(s.doEntries)
}

func (s *MapStateMachine) Snapshot(writer *snapshot.Writer) error {
	s.Log().Infow("Persisting Map to snapshot")
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
	for _, entry := range s.entries {
		if err := writer.WriteMessage(entry); err != nil {
			return err
		}
	}
	return nil
}

func (s *MapStateMachine) Recover(reader *snapshot.Reader) error {
	s.Log().Infow("Recovering Map from snapshot")
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		proposalID, err := reader.ReadVarUint64()
		if err != nil {
			return err
		}
		proposal, ok := s.Proposals().Get(primitive.ProposalID(proposalID))
		if !ok {
			return errors.NewFault("cannot find proposal %d", proposalID)
		}
		listener := &mapv1.MapListener{}
		if err := reader.ReadMessage(listener); err != nil {
			return err
		}
		s.listeners[proposal.ID()] = listener
		proposal.Watch(func(phase statemachine.Phase) {
			if phase == statemachine.Complete {
				delete(s.listeners, proposal.ID())
			}
		})
	}

	n, err = reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		entry := &mapv1.MapEntry{}
		if err := reader.ReadMessage(entry); err != nil {
			return err
		}
		s.entries[entry.Key] = entry
		s.scheduleTTL(entry.Key, entry)
	}
	return nil
}

func (s *MapStateMachine) Propose(proposal primitive.Proposal[*mapv1.MapInput, *mapv1.MapOutput]) {
	switch proposal.Input().Input.(type) {
	case *mapv1.MapInput_Put:
		s.put.Execute(proposal)
	case *mapv1.MapInput_Insert:
		s.insert.Execute(proposal)
	case *mapv1.MapInput_Update:
		s.update.Execute(proposal)
	case *mapv1.MapInput_Remove:
		s.remove.Execute(proposal)
	case *mapv1.MapInput_Clear:
		s.clear.Execute(proposal)
	case *mapv1.MapInput_Events:
		s.events.Execute(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *MapStateMachine) doPut(proposal primitive.Proposal[*mapv1.PutInput, *mapv1.PutOutput]) {
	defer proposal.Close()

	oldEntry := s.entries[proposal.Input().Key]

	// If the value is equal to the current value, return a no-op.
	if oldEntry != nil && bytes.Equal(oldEntry.Value.Value, proposal.Input().Value) {
		proposal.Output(&mapv1.PutOutput{
			Index: oldEntry.Value.Index,
		})
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.MapEntry{
		Key: proposal.Input().Key,
		Value: &mapv1.MapValue{
			Value: proposal.Input().Value,
			Index: multiraftv1.Index(s.Index()),
		},
	}
	if proposal.Input().TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().Key, newEntry)

	// Publish an event to listener streams.
	if oldEntry != nil {
		s.notify(newEntry, &mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: newEntry.Key,
				Event: &mapv1.Event_Updated_{
					Updated: &mapv1.Event_Updated{
						Value: mapv1.IndexedValue{
							Value: newEntry.Value.Value,
							Index: newEntry.Value.Index,
						},
						PrevValue: mapv1.IndexedValue{
							Value: oldEntry.Value.Value,
							Index: oldEntry.Value.Index,
						},
					},
				},
			},
		})
		proposal.Output(&mapv1.PutOutput{
			Index: newEntry.Value.Index,
			PrevValue: &mapv1.IndexedValue{
				Value: oldEntry.Value.Value,
				Index: oldEntry.Value.Index,
			},
		})
	} else {
		s.notify(newEntry, &mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: newEntry.Key,
				Event: &mapv1.Event_Inserted_{
					Inserted: &mapv1.Event_Inserted{
						Value: mapv1.IndexedValue{
							Value: newEntry.Value.Value,
							Index: newEntry.Value.Index,
						},
					},
				},
			},
		})
		proposal.Output(&mapv1.PutOutput{
			Index: newEntry.Value.Index,
		})
	}
}

func (s *MapStateMachine) doInsert(proposal primitive.Proposal[*mapv1.InsertInput, *mapv1.InsertOutput]) {
	defer proposal.Close()

	if _, ok := s.entries[proposal.Input().Key]; ok {
		proposal.Error(errors.NewAlreadyExists("key '%s' already exists", proposal.Input().Key))
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.MapEntry{
		Key: proposal.Input().Key,
		Value: &mapv1.MapValue{
			Value: proposal.Input().Value,
			Index: multiraftv1.Index(s.Index()),
		},
	}
	if proposal.Input().TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().Key, newEntry)

	// Publish an event to listener streams.
	s.notify(newEntry, &mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: newEntry.Key,
			Event: &mapv1.Event_Inserted_{
				Inserted: &mapv1.Event_Inserted{
					Value: mapv1.IndexedValue{
						Value: newEntry.Value.Value,
						Index: newEntry.Value.Index,
					},
				},
			},
		},
	})

	proposal.Output(&mapv1.InsertOutput{
		Index: newEntry.Value.Index,
	})
}

func (s *MapStateMachine) doUpdate(proposal primitive.Proposal[*mapv1.UpdateInput, *mapv1.UpdateOutput]) {
	defer proposal.Close()

	oldEntry, ok := s.entries[proposal.Input().Key]
	if !ok {
		proposal.Error(errors.NewNotFound("key '%s' not found", proposal.Input().Key))
		return
	}

	if proposal.Input().PrevIndex > 0 && oldEntry.Value.Index != proposal.Input().PrevIndex {
		proposal.Error(errors.NewConflict("entry index %d does not match remove index %d", oldEntry.Value.Index, proposal.Input().PrevIndex))
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.MapEntry{
		Key: proposal.Input().Key,
		Value: &mapv1.MapValue{
			Value: proposal.Input().Value,
			Index: multiraftv1.Index(s.Index()),
		},
	}
	if proposal.Input().TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().Key, newEntry)

	// Publish an event to listener streams.
	s.notify(newEntry, &mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: newEntry.Key,
			Event: &mapv1.Event_Updated_{
				Updated: &mapv1.Event_Updated{
					Value: mapv1.IndexedValue{
						Value: newEntry.Value.Value,
						Index: newEntry.Value.Index,
					},
					PrevValue: mapv1.IndexedValue{
						Value: oldEntry.Value.Value,
						Index: oldEntry.Value.Index,
					},
				},
			},
		},
	})

	proposal.Output(&mapv1.UpdateOutput{
		Index: newEntry.Value.Index,
		PrevValue: mapv1.IndexedValue{
			Value: oldEntry.Value.Value,
			Index: oldEntry.Value.Index,
		},
	})
}

func (s *MapStateMachine) doRemove(proposal primitive.Proposal[*mapv1.RemoveInput, *mapv1.RemoveOutput]) {
	defer proposal.Close()
	entry, ok := s.entries[proposal.Input().Key]
	if !ok {
		proposal.Error(errors.NewNotFound("key '%s' not found", proposal.Input().Key))
		return
	}

	if proposal.Input().PrevIndex > 0 && entry.Value.Index != proposal.Input().PrevIndex {
		proposal.Error(errors.NewConflict("entry index %d does not match remove index %d", entry.Value.Index, proposal.Input().PrevIndex))
		return
	}

	delete(s.entries, proposal.Input().Key)

	// Schedule the timeout for the value if necessary.
	s.cancelTTL(entry.Key)

	// Publish an event to listener streams.
	s.notify(entry, &mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: entry.Key,
			Event: &mapv1.Event_Removed_{
				Removed: &mapv1.Event_Removed{
					Value: mapv1.IndexedValue{
						Value: entry.Value.Value,
						Index: entry.Value.Index,
					},
				},
			},
		},
	})

	proposal.Output(&mapv1.RemoveOutput{
		Value: mapv1.IndexedValue{
			Value: entry.Value.Value,
			Index: entry.Value.Index,
		},
	})
}

func (s *MapStateMachine) doClear(proposal primitive.Proposal[*mapv1.ClearInput, *mapv1.ClearOutput]) {
	defer proposal.Close()
	for key, entry := range s.entries {
		s.notify(entry, &mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: entry.Key,
				Event: &mapv1.Event_Removed_{
					Removed: &mapv1.Event_Removed{
						Value: mapv1.IndexedValue{
							Value: entry.Value.Value,
							Index: entry.Value.Index,
						},
					},
				},
			},
		})
		s.cancelTTL(key)
		delete(s.entries, key)
	}
	proposal.Output(&mapv1.ClearOutput{})
}

func (s *MapStateMachine) doEvents(proposal primitive.Proposal[*mapv1.EventsInput, *mapv1.EventsOutput]) {
	listener := &mapv1.MapListener{
		Key: proposal.Input().Key,
	}
	s.listeners[proposal.ID()] = listener
	proposal.Watch(func(phase statemachine.Phase) {
		if phase == statemachine.Complete {
			delete(s.listeners, proposal.ID())
		}
	})
}

func (s *MapStateMachine) Query(query primitive.Query[*mapv1.MapInput, *mapv1.MapOutput]) {
	switch query.Input().Input.(type) {
	case *mapv1.MapInput_Size_:
		s.size.Execute(query)
	case *mapv1.MapInput_Get:
		s.get.Execute(query)
	case *mapv1.MapInput_Entries:
		s.list.Execute(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *MapStateMachine) doSize(query primitive.Query[*mapv1.SizeInput, *mapv1.SizeOutput]) {
	defer query.Close()
	query.Output(&mapv1.SizeOutput{
		Size_: uint32(len(s.entries)),
	})
}

func (s *MapStateMachine) doGet(query primitive.Query[*mapv1.GetInput, *mapv1.GetOutput]) {
	defer query.Close()
	entry, ok := s.entries[query.Input().Key]
	if !ok {
		query.Error(errors.NewNotFound("key %s not found", query.Input().Key))
	} else {
		query.Output(&mapv1.GetOutput{
			Value: mapv1.IndexedValue{
				Value: entry.Value.Value,
				Index: entry.Value.Index,
			},
		})
	}
}

func (s *MapStateMachine) doEntries(query primitive.Query[*mapv1.EntriesInput, *mapv1.EntriesOutput]) {
	for _, entry := range s.entries {
		query.Output(&mapv1.EntriesOutput{
			Entry: mapv1.Entry{
				Key: entry.Key,
				Value: &mapv1.IndexedValue{
					Value: entry.Value.Value,
					Index: entry.Value.Index,
				},
			},
		})
	}

	if query.Input().Watch {
		s.mu.Lock()
		s.watchers[query.ID()] = query
		s.mu.Unlock()
		query.Watch(func(phase statemachine.Phase) {
			if phase == statemachine.Canceled {
				s.mu.Lock()
				delete(s.watchers, query.ID())
				s.mu.Unlock()
			}
		})
	} else {
		query.Close()
	}
}

func (s *MapStateMachine) notify(entry *mapv1.MapEntry, event *mapv1.EventsOutput) {
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
		if entry.Value != nil {
			watcher.Output(&mapv1.EntriesOutput{
				Entry: mapv1.Entry{
					Key: entry.Key,
					Value: &mapv1.IndexedValue{
						Value: entry.Value.Value,
						Index: entry.Value.Index,
					},
				},
			})
		} else {
			watcher.Output(&mapv1.EntriesOutput{
				Entry: mapv1.Entry{
					Key: entry.Key,
				},
			})
		}
	}
}

func (s *MapStateMachine) scheduleTTL(key string, entry *mapv1.MapEntry) {
	s.cancelTTL(key)
	if entry.Value.Expire != nil {
		s.timers[key] = s.Scheduler().Schedule(*entry.Value.Expire, func() {
			delete(s.entries, key)
			s.notify(entry, &mapv1.EventsOutput{
				Event: mapv1.Event{
					Key: key,
					Event: &mapv1.Event_Removed_{
						Removed: &mapv1.Event_Removed{
							Value: mapv1.IndexedValue{
								Value: entry.Value.Value,
								Index: entry.Value.Index,
							},
							Expired: true,
						},
					},
				},
			})
		})
	}
}

func (s *MapStateMachine) cancelTTL(key string) {
	timer, ok := s.timers[key]
	if ok {
		timer.Cancel()
	}
}
