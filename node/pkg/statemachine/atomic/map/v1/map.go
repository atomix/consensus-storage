// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	mapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/map/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.multiraft.atomic.map.v1.AtomicMap"

func Register(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput](registry)(MapType)
}

var MapType = statemachine.NewPrimitiveType[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput](Service, mapCodec, newMapStateMachine)

var mapCodec = statemachine.NewCodec[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput](
	func(bytes []byte) (*mapv1.AtomicMapInput, error) {
		input := &mapv1.AtomicMapInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *mapv1.AtomicMapOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newMapStateMachine(ctx statemachine.PrimitiveContext[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) statemachine.Primitive[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput] {
	return &MapStateMachine{
		PrimitiveContext: ctx,
		listeners:        make(map[statemachine.ProposalID]*mapv1.AtomicMapListener),
		entries:          make(map[string]*mapv1.AtomicMapEntry),
		timers:           make(map[string]statemachine.Timer),
		watchers:         make(map[statemachine.QueryID]statemachine.Query[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]),
	}
}

type MapStateMachine struct {
	statemachine.PrimitiveContext[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]
	listeners map[statemachine.ProposalID]*mapv1.AtomicMapListener
	entries   map[string]*mapv1.AtomicMapEntry
	timers    map[string]statemachine.Timer
	watchers  map[statemachine.QueryID]statemachine.Query[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]
	mu        sync.RWMutex
}

func (s *MapStateMachine) Snapshot(writer *snapshot.Writer) error {
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
		listener := &mapv1.AtomicMapListener{}
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
		entry := &mapv1.AtomicMapEntry{}
		if err := reader.ReadMessage(entry); err != nil {
			return err
		}
		s.entries[entry.Key] = entry
		s.scheduleTTL(entry.Key, entry)
	}
	return nil
}

func (s *MapStateMachine) Update(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	switch proposal.Input().Input.(type) {
	case *mapv1.AtomicMapInput_Put:
		s.put(proposal)
	case *mapv1.AtomicMapInput_Insert:
		s.insert(proposal)
	case *mapv1.AtomicMapInput_Update:
		s.update(proposal)
	case *mapv1.AtomicMapInput_Remove:
		s.remove(proposal)
	case *mapv1.AtomicMapInput_Clear:
		s.clear(proposal)
	case *mapv1.AtomicMapInput_Events:
		s.events(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *MapStateMachine) put(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer proposal.Close()

	oldEntry := s.entries[proposal.Input().GetPut().Key]

	// If the value is equal to the current value, return a no-op.
	if oldEntry != nil && bytes.Equal(oldEntry.Value.Value, proposal.Input().GetPut().Value.Value) {
		proposal.Output(&mapv1.AtomicMapOutput{
			Output: &mapv1.AtomicMapOutput_Put{
				Put: &mapv1.PutOutput{
					Index: oldEntry.Value.Index,
				},
			},
		})
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.AtomicMapEntry{
		Key: proposal.Input().GetPut().Key,
		Value: &mapv1.AtomicMapValue{
			Value: proposal.Input().GetPut().Value.Value,
			Index: multiraftv1.Index(s.Index()),
		},
	}
	if proposal.Input().GetPut().Value.TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().GetPut().Value.TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().GetPut().Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().GetPut().Key, newEntry)

	// Publish an event to listener streams.
	if oldEntry != nil {
		s.notify(newEntry, &mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: newEntry.Key,
				Event: &mapv1.Event_Updated_{
					Updated: &mapv1.Event_Updated{
						NewValue:  *s.newValue(newEntry.Value),
						PrevValue: *s.newValue(oldEntry.Value),
					},
				},
			},
		})
		proposal.Output(&mapv1.AtomicMapOutput{
			Output: &mapv1.AtomicMapOutput_Put{
				Put: &mapv1.PutOutput{
					Index:     newEntry.Value.Index,
					PrevValue: s.newValue(oldEntry.Value),
				},
			},
		})
	} else {
		s.notify(newEntry, &mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: newEntry.Key,
				Event: &mapv1.Event_Inserted_{
					Inserted: &mapv1.Event_Inserted{
						Value: *s.newValue(newEntry.Value),
					},
				},
			},
		})
		proposal.Output(&mapv1.AtomicMapOutput{
			Output: &mapv1.AtomicMapOutput_Put{
				Put: &mapv1.PutOutput{
					Index: newEntry.Value.Index,
				},
			},
		})
	}
}

func (s *MapStateMachine) insert(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer proposal.Close()

	if _, ok := s.entries[proposal.Input().GetInsert().Key]; ok {
		proposal.Error(errors.NewAlreadyExists("key '%s' already exists", proposal.Input().GetInsert().Key))
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.AtomicMapEntry{
		Key: proposal.Input().GetInsert().Key,
		Value: &mapv1.AtomicMapValue{
			Value: proposal.Input().GetInsert().Value.Value,
			Index: multiraftv1.Index(s.Index()),
		},
	}
	if proposal.Input().GetInsert().Value.TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().GetInsert().Value.TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().GetInsert().Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().GetInsert().Key, newEntry)

	// Publish an event to listener streams.
	s.notify(newEntry, &mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: newEntry.Key,
			Event: &mapv1.Event_Inserted_{
				Inserted: &mapv1.Event_Inserted{
					Value: *s.newValue(newEntry.Value),
				},
			},
		},
	})

	proposal.Output(&mapv1.AtomicMapOutput{
		Output: &mapv1.AtomicMapOutput_Insert{
			Insert: &mapv1.InsertOutput{
				Index: newEntry.Value.Index,
			},
		},
	})
}

func (s *MapStateMachine) update(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer proposal.Close()

	oldEntry, ok := s.entries[proposal.Input().GetUpdate().Key]
	if !ok {
		proposal.Error(errors.NewNotFound("key '%s' not found", proposal.Input().GetUpdate().Key))
		return
	}

	if proposal.Input().GetUpdate().Index > 0 && oldEntry.Value.Index != proposal.Input().GetUpdate().Index {
		proposal.Error(errors.NewConflict("entry index %d does not match remove index %d", oldEntry.Value.Index, proposal.Input().GetUpdate().Key))
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.AtomicMapEntry{
		Key: proposal.Input().GetUpdate().Key,
		Value: &mapv1.AtomicMapValue{
			Value: proposal.Input().GetUpdate().Value.Value,
			Index: multiraftv1.Index(s.Index()),
		},
	}
	if proposal.Input().GetUpdate().Value.TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().GetUpdate().Value.TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().GetUpdate().Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().GetUpdate().Key, newEntry)

	// Publish an event to listener streams.
	s.notify(newEntry, &mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: newEntry.Key,
			Event: &mapv1.Event_Updated_{
				Updated: &mapv1.Event_Updated{
					NewValue:  *s.newValue(newEntry.Value),
					PrevValue: *s.newValue(oldEntry.Value),
				},
			},
		},
	})

	proposal.Output(&mapv1.AtomicMapOutput{
		Output: &mapv1.AtomicMapOutput_Update{
			Update: &mapv1.UpdateOutput{
				Index:     newEntry.Value.Index,
				PrevValue: *s.newValue(oldEntry.Value),
			},
		},
	})
}

func (s *MapStateMachine) remove(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer proposal.Close()
	entry, ok := s.entries[proposal.Input().GetRemove().Key]
	if !ok {
		proposal.Error(errors.NewNotFound("key '%s' not found", proposal.Input().GetRemove().Key))
		return
	}

	if proposal.Input().GetRemove().Index > 0 && entry.Value.Index != proposal.Input().GetRemove().Index {
		proposal.Error(errors.NewConflict("entry index %d does not match remove index %d", entry.Value.Index, proposal.Input().GetRemove().Key))
		return
	}

	delete(s.entries, proposal.Input().GetRemove().Key)

	// Schedule the timeout for the value if necessary.
	s.cancelTTL(entry.Key)

	// Publish an event to listener streams.
	s.notify(&mapv1.AtomicMapEntry{
		Key: entry.Key,
	}, &mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: entry.Key,
			Event: &mapv1.Event_Removed_{
				Removed: &mapv1.Event_Removed{
					Value: *s.newValue(entry.Value),
				},
			},
		},
	})

	proposal.Output(&mapv1.AtomicMapOutput{
		Output: &mapv1.AtomicMapOutput_Remove{
			Remove: &mapv1.RemoveOutput{
				Value: *s.newValue(entry.Value),
			},
		},
	})
}

func (s *MapStateMachine) clear(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer proposal.Close()
	for key, entry := range s.entries {
		s.notify(&mapv1.AtomicMapEntry{
			Key: entry.Key,
		}, &mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: entry.Key,
				Event: &mapv1.Event_Removed_{
					Removed: &mapv1.Event_Removed{
						Value: *s.newValue(entry.Value),
					},
				},
			},
		})
		s.cancelTTL(key)
		delete(s.entries, key)
	}
	proposal.Output(&mapv1.AtomicMapOutput{
		Output: &mapv1.AtomicMapOutput_Clear{
			Clear: &mapv1.ClearOutput{},
		},
	})
}

func (s *MapStateMachine) events(proposal statemachine.Proposal[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	// Output an empty event to ack the request
	proposal.Output(&mapv1.AtomicMapOutput{
		Output: &mapv1.AtomicMapOutput_Events{
			Events: &mapv1.EventsOutput{},
		},
	})

	listener := &mapv1.AtomicMapListener{
		Key: proposal.Input().GetEvents().Key,
	}
	s.listeners[proposal.ID()] = listener
	proposal.Watch(func(state statemachine.OperationState) {
		if state == statemachine.Complete {
			delete(s.listeners, proposal.ID())
		}
	})
}

func (s *MapStateMachine) Read(query statemachine.Query[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	switch query.Input().Input.(type) {
	case *mapv1.AtomicMapInput_Size_:
		s.size(query)
	case *mapv1.AtomicMapInput_Get:
		s.get(query)
	case *mapv1.AtomicMapInput_Entries:
		s.list(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *MapStateMachine) size(query statemachine.Query[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer query.Close()
	query.Output(&mapv1.AtomicMapOutput{
		Output: &mapv1.AtomicMapOutput_Size_{
			Size_: &mapv1.SizeOutput{
				Size_: uint32(len(s.entries)),
			},
		},
	})
}

func (s *MapStateMachine) get(query statemachine.Query[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	defer query.Close()
	entry, ok := s.entries[query.Input().GetGet().Key]
	if !ok {
		query.Error(errors.NewNotFound("key %s not found", query.Input().GetGet().Key))
	} else {
		query.Output(&mapv1.AtomicMapOutput{
			Output: &mapv1.AtomicMapOutput_Get{
				Get: &mapv1.GetOutput{
					Value: *s.newValue(entry.Value),
				},
			},
		})
	}
}

func (s *MapStateMachine) list(query statemachine.Query[*mapv1.AtomicMapInput, *mapv1.AtomicMapOutput]) {
	for _, entry := range s.entries {
		query.Output(&mapv1.AtomicMapOutput{
			Output: &mapv1.AtomicMapOutput_Entries{
				Entries: &mapv1.EntriesOutput{
					Entry: *s.newEntry(entry),
				},
			},
		})
	}

	if query.Input().GetEntries().Watch {
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

func (s *MapStateMachine) notify(entry *mapv1.AtomicMapEntry, event *mapv1.EventsOutput) {
	for proposalID, listener := range s.listeners {
		if listener.Key == "" || listener.Key == event.Event.Key {
			proposal, ok := s.Proposals().Get(proposalID)
			if ok {
				proposal.Output(&mapv1.AtomicMapOutput{
					Output: &mapv1.AtomicMapOutput_Events{
						Events: event,
					},
				})
			} else {
				delete(s.listeners, proposalID)
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, watcher := range s.watchers {
		watcher.Output(&mapv1.AtomicMapOutput{
			Output: &mapv1.AtomicMapOutput_Entries{
				Entries: &mapv1.EntriesOutput{
					Entry: *s.newEntry(entry),
				},
			},
		})
	}
}

func (s *MapStateMachine) scheduleTTL(key string, entry *mapv1.AtomicMapEntry) {
	s.cancelTTL(key)
	if entry.Value.Expire != nil {
		s.timers[key] = s.Scheduler().RunAt(*entry.Value.Expire, func() {
			delete(s.entries, key)
			s.notify(&mapv1.AtomicMapEntry{
				Key: entry.Key,
			}, &mapv1.EventsOutput{
				Event: mapv1.Event{
					Key: key,
					Event: &mapv1.Event_Expired_{
						Expired: &mapv1.Event_Expired{
							Value: *s.newValue(entry.Value),
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

func (s *MapStateMachine) newEntry(state *mapv1.AtomicMapEntry) *mapv1.Entry {
	entry := &mapv1.Entry{
		Key: state.Key,
	}
	if state.Value != nil {
		entry.Value = *s.newValue(state.Value)
		if state.Value.Expire != nil {
			ttl := s.Scheduler().Time().Sub(*state.Value.Expire)
			entry.Value.TTL = &ttl
		}
	}
	return entry
}

func (s *MapStateMachine) newValue(state *mapv1.AtomicMapValue) *mapv1.Value {
	value := &mapv1.Value{
		Value: state.Value,
		Index: state.Index,
	}
	if state.Expire != nil {
		ttl := s.Scheduler().Time().Sub(*state.Expire)
		value.TTL = &ttl
	}
	return value
}
