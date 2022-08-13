// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	mapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

const Service = "atomix.multiraft.map.v1.Map"

func Register(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*mapv1.MapInput, *mapv1.MapOutput](registry)(MapType)
}

var MapType = statemachine.NewPrimitiveType[*mapv1.MapInput, *mapv1.MapOutput](Service, mapCodec, newMapStateMachine)

var mapCodec = statemachine.NewCodec[*mapv1.MapInput, *mapv1.MapOutput](
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

func newMapStateMachine(ctx statemachine.PrimitiveContext[*mapv1.MapInput, *mapv1.MapOutput]) statemachine.Primitive[*mapv1.MapInput, *mapv1.MapOutput] {
	return &MapStateMachine{
		PrimitiveContext: ctx,
		listeners:        make(map[statemachine.ProposalID]*mapv1.MapListener),
		entries:          make(map[string]*mapv1.MapEntry),
		timers:           make(map[string]statemachine.Timer),
	}
}

type MapStateMachine struct {
	statemachine.PrimitiveContext[*mapv1.MapInput, *mapv1.MapOutput]
	listeners map[statemachine.ProposalID]*mapv1.MapListener
	entries   map[string]*mapv1.MapEntry
	timers    map[string]statemachine.Timer
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
		listener := &mapv1.MapListener{}
		if err := reader.ReadMessage(listener); err != nil {
			return err
		}
		s.listeners[proposal.ID()] = listener
		proposal.Watch(func(state statemachine.ProposalState) {
			if state == statemachine.ProposalComplete {
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

func (s *MapStateMachine) Update(proposal statemachine.Proposal[*mapv1.MapInput, *mapv1.MapOutput]) {
	switch proposal.Input().Input.(type) {
	case *mapv1.MapInput_Put:
		s.proposePut(proposal)
	case *mapv1.MapInput_Remove:
		s.proposeRemove(proposal)
	case *mapv1.MapInput_Clear:
		s.proposeClear(proposal)
	case *mapv1.MapInput_Events:
		s.proposeEvents(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *MapStateMachine) proposePut(proposal statemachine.Proposal[*mapv1.MapInput, *mapv1.MapOutput]) {
	defer proposal.Close()

	oldEntry := s.entries[proposal.Input().GetPut().Entry.Key]

	// If the value is equal to the current value, return a no-op.
	if oldEntry != nil && bytes.Equal(oldEntry.Value.Value, proposal.Input().GetPut().Entry.Value.Value) {
		proposal.Output(&mapv1.MapOutput{
			Output: &mapv1.MapOutput_Put{
				Put: &mapv1.PutOutput{},
			},
		})
		return
	}

	// Create a new entry and increment the revision number
	newEntry := &mapv1.MapEntry{
		Key: proposal.Input().GetPut().Entry.Key,
		Value: &mapv1.MapValue{
			Value: proposal.Input().GetPut().Entry.Value.Value,
		},
	}
	if proposal.Input().GetPut().Entry.Value.TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().GetPut().Entry.Value.TTL)
		newEntry.Value.Expire = &expire
	}

	// Create a new entry value and set it in the map.
	s.entries[proposal.Input().GetPut().Entry.Key] = newEntry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(proposal.Input().GetPut().Entry.Key, newEntry)

	// Publish an event to listener streams.
	if oldEntry != nil {
		s.notify(&mapv1.EventsOutput{
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
	} else {
		s.notify(&mapv1.EventsOutput{
			Event: mapv1.Event{
				Key: newEntry.Key,
				Event: &mapv1.Event_Inserted_{
					Inserted: &mapv1.Event_Inserted{
						Value: *s.newValue(newEntry.Value),
					},
				},
			},
		})
	}

	proposal.Output(&mapv1.MapOutput{
		Output: &mapv1.MapOutput_Put{
			Put: &mapv1.PutOutput{},
		},
	})
}

func (s *MapStateMachine) proposeRemove(proposal statemachine.Proposal[*mapv1.MapInput, *mapv1.MapOutput]) {
	defer proposal.Close()
	entry, ok := s.entries[proposal.Input().GetRemove().Key]
	if !ok {
		proposal.Error(errors.NewNotFound("key '%s' not found", proposal.Input().GetRemove().Key))
		return
	}

	delete(s.entries, proposal.Input().GetRemove().Key)

	// Schedule the timeout for the value if necessary.
	s.cancelTTL(entry.Key)

	// Publish an event to listener streams.
	s.notify(&mapv1.EventsOutput{
		Event: mapv1.Event{
			Key: entry.Key,
			Event: &mapv1.Event_Removed_{
				Removed: &mapv1.Event_Removed{
					Value: *s.newValue(entry.Value),
				},
			},
		},
	})

	proposal.Output(&mapv1.MapOutput{
		Output: &mapv1.MapOutput_Remove{
			Remove: &mapv1.RemoveOutput{
				Value: s.newValue(entry.Value),
			},
		},
	})
}

func (s *MapStateMachine) proposeClear(proposal statemachine.Proposal[*mapv1.MapInput, *mapv1.MapOutput]) {
	defer proposal.Close()
	for key, entry := range s.entries {
		s.notify(&mapv1.EventsOutput{
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
	proposal.Output(&mapv1.MapOutput{
		Output: &mapv1.MapOutput_Clear{
			Clear: &mapv1.ClearOutput{},
		},
	})
}

func (s *MapStateMachine) proposeEvents(proposal statemachine.Proposal[*mapv1.MapInput, *mapv1.MapOutput]) {
	// Output an empty event to ack the request
	proposal.Output(&mapv1.MapOutput{
		Output: &mapv1.MapOutput_Events{
			Events: &mapv1.EventsOutput{},
		},
	})

	listener := &mapv1.MapListener{
		Key: proposal.Input().GetEvents().Key,
	}
	s.listeners[proposal.ID()] = listener
	proposal.Watch(func(state statemachine.ProposalState) {
		if state == statemachine.ProposalComplete {
			delete(s.listeners, proposal.ID())
		}
	})
}

func (s *MapStateMachine) Read(query statemachine.Query[*mapv1.MapInput, *mapv1.MapOutput]) {
	switch query.Input().Input.(type) {
	case *mapv1.MapInput_Size_:
		s.querySize(query)
	case *mapv1.MapInput_Get:
		s.queryGet(query)
	case *mapv1.MapInput_Entries:
		s.queryEntries(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *MapStateMachine) querySize(query statemachine.Query[*mapv1.MapInput, *mapv1.MapOutput]) {
	query.Output(&mapv1.MapOutput{
		Output: &mapv1.MapOutput_Size_{
			Size_: &mapv1.SizeOutput{
				Size_: uint32(len(s.entries)),
			},
		},
	})
}

func (s *MapStateMachine) queryGet(query statemachine.Query[*mapv1.MapInput, *mapv1.MapOutput]) {
	entry, ok := s.entries[query.Input().GetGet().Key]
	if !ok {
		query.Error(errors.NewNotFound("key %s not found", query.Input().GetGet().Key))
	} else {
		query.Output(&mapv1.MapOutput{
			Output: &mapv1.MapOutput_Get{
				Get: &mapv1.GetOutput{
					Value: s.newValue(entry.Value),
				},
			},
		})
	}
}

func (s *MapStateMachine) queryEntries(query statemachine.Query[*mapv1.MapInput, *mapv1.MapOutput]) {
	for _, entry := range s.entries {
		query.Output(&mapv1.MapOutput{
			Output: &mapv1.MapOutput_Entries{
				Entries: &mapv1.EntriesOutput{
					Entry: s.newEntry(entry),
				},
			},
		})
	}
}

func (s *MapStateMachine) notify(event *mapv1.EventsOutput) {
	for proposalID, listener := range s.listeners {
		if listener.Key == "" || listener.Key == event.Event.Key {
			proposal, ok := s.Proposals().Get(proposalID)
			if ok {
				proposal.Output(&mapv1.MapOutput{
					Output: &mapv1.MapOutput_Events{
						Events: event,
					},
				})
			} else {
				delete(s.listeners, proposalID)
			}
		}
	}
}

func (s *MapStateMachine) scheduleTTL(key string, entry *mapv1.MapEntry) {
	s.cancelTTL(key)
	if entry.Value.Expire != nil {
		s.timers[key] = s.Scheduler().RunAt(*entry.Value.Expire, func() {
			delete(s.entries, key)
			s.notify(&mapv1.EventsOutput{
				Event: mapv1.Event{
					Key: entry.Key,
					Event: &mapv1.Event_Removed_{
						Removed: &mapv1.Event_Removed{
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

func (s *MapStateMachine) newEntry(state *mapv1.MapEntry) *mapv1.Entry {
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

func (s *MapStateMachine) newValue(state *mapv1.MapValue) *mapv1.Value {
	value := &mapv1.Value{
		Value: state.Value,
	}
	if state.Expire != nil {
		ttl := s.Scheduler().Time().Sub(*state.Expire)
		value.TTL = &ttl
	}
	return value
}
