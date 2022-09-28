// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	multimapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/multimap/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.runtime.multimap.v1.MultiMap"

func Register(registry *primitive.TypeRegistry) {
	primitive.RegisterType[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput](registry)(Type)
}

var Type = primitive.NewType[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput](Service, multiMapCodec, newMultiMapStateMachine)

var multiMapCodec = primitive.NewCodec[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput](
	func(bytes []byte) (*multimapv1.MultiMapInput, error) {
		input := &multimapv1.MultiMapInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *multimapv1.MultiMapOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newMultiMapStateMachine(ctx primitive.Context[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput]) primitive.Primitive[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput] {
	sm := &MultiMapStateMachine{
		Context:   ctx,
		listeners: make(map[statemachine.ProposalID]*multimapv1.MultiMapListener),
		entries:   make(map[string]map[string]bool),
		watchers:  make(map[statemachine.QueryID]statemachine.Query[*multimapv1.EntriesInput, *multimapv1.EntriesOutput]),
	}
	sm.init()
	return sm
}

type MultiMapStateMachine struct {
	primitive.Context[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput]
	listeners map[statemachine.ProposalID]*multimapv1.MultiMapListener
	entries   map[string]map[string]bool
	watchers  map[statemachine.QueryID]statemachine.Query[*multimapv1.EntriesInput, *multimapv1.EntriesOutput]
	mu        sync.RWMutex
	put       primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.PutInput, *multimapv1.PutOutput]
	putAll    primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.PutAllInput, *multimapv1.PutAllOutput]
	replace   primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.ReplaceInput, *multimapv1.ReplaceOutput]
	remove    primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.RemoveInput, *multimapv1.RemoveOutput]
	removeAll primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.RemoveAllInput, *multimapv1.RemoveAllOutput]
	clear     primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.ClearInput, *multimapv1.ClearOutput]
	events    primitive.Proposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.EventsInput, *multimapv1.EventsOutput]
	size      primitive.Querier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.SizeInput, *multimapv1.SizeOutput]
	contains  primitive.Querier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.ContainsInput, *multimapv1.ContainsOutput]
	get       primitive.Querier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.GetInput, *multimapv1.GetOutput]
	list      primitive.Querier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.EntriesInput, *multimapv1.EntriesOutput]
}

func (s *MultiMapStateMachine) init() {
	s.put = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.PutInput, *multimapv1.PutOutput](s).
		Name("Put").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.PutInput, bool) {
			if put, ok := input.Input.(*multimapv1.MultiMapInput_Put); ok {
				return put.Put, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.PutOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Put{
					Put: output,
				},
			}
		}).
		Build(s.doPut)
	s.putAll = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.PutAllInput, *multimapv1.PutAllOutput](s).
		Name("PutAll").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.PutAllInput, bool) {
			if put, ok := input.Input.(*multimapv1.MultiMapInput_PutAll); ok {
				return put.PutAll, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.PutAllOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_PutAll{
					PutAll: output,
				},
			}
		}).
		Build(s.doPutAll)
	s.replace = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.ReplaceInput, *multimapv1.ReplaceOutput](s).
		Name("Replace").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.ReplaceInput, bool) {
			if put, ok := input.Input.(*multimapv1.MultiMapInput_Replace); ok {
				return put.Replace, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.ReplaceOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Replace{
					Replace: output,
				},
			}
		}).
		Build(s.doReplace)
	s.remove = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.RemoveInput, *multimapv1.RemoveOutput](s).
		Name("Remove").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.RemoveInput, bool) {
			if put, ok := input.Input.(*multimapv1.MultiMapInput_Remove); ok {
				return put.Remove, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.RemoveOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Remove{
					Remove: output,
				},
			}
		}).
		Build(s.doRemove)
	s.removeAll = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.RemoveAllInput, *multimapv1.RemoveAllOutput](s).
		Name("RemoveAll").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.RemoveAllInput, bool) {
			if put, ok := input.Input.(*multimapv1.MultiMapInput_RemoveAll); ok {
				return put.RemoveAll, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.RemoveAllOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_RemoveAll{
					RemoveAll: output,
				},
			}
		}).
		Build(s.doRemoveAll)
	s.clear = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.ClearInput, *multimapv1.ClearOutput](s).
		Name("Clear").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.ClearInput, bool) {
			if clear, ok := input.Input.(*multimapv1.MultiMapInput_Clear); ok {
				return clear.Clear, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.ClearOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Clear{
					Clear: output,
				},
			}
		}).
		Build(s.doClear)
	s.events = primitive.NewProposer[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.EventsInput, *multimapv1.EventsOutput](s).
		Name("Events").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.EventsInput, bool) {
			if events, ok := input.Input.(*multimapv1.MultiMapInput_Events); ok {
				return events.Events, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.EventsOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Events{
					Events: output,
				},
			}
		}).
		Build(s.doEvents)
	s.size = primitive.NewQuerier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.SizeInput, *multimapv1.SizeOutput](s).
		Name("Size").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.SizeInput, bool) {
			if size, ok := input.Input.(*multimapv1.MultiMapInput_Size_); ok {
				return size.Size_, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.SizeOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Size_{
					Size_: output,
				},
			}
		}).
		Build(s.doSize)
	s.contains = primitive.NewQuerier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.ContainsInput, *multimapv1.ContainsOutput](s).
		Name("Contains").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.ContainsInput, bool) {
			if get, ok := input.Input.(*multimapv1.MultiMapInput_Contains); ok {
				return get.Contains, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.ContainsOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Contains{
					Contains: output,
				},
			}
		}).
		Build(s.doContains)
	s.get = primitive.NewQuerier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.GetInput, *multimapv1.GetOutput](s).
		Name("Get").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.GetInput, bool) {
			if get, ok := input.Input.(*multimapv1.MultiMapInput_Get); ok {
				return get.Get, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.GetOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Get{
					Get: output,
				},
			}
		}).
		Build(s.doGet)
	s.list = primitive.NewQuerier[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput, *multimapv1.EntriesInput, *multimapv1.EntriesOutput](s).
		Name("Entries").
		Decoder(func(input *multimapv1.MultiMapInput) (*multimapv1.EntriesInput, bool) {
			if entries, ok := input.Input.(*multimapv1.MultiMapInput_Entries); ok {
				return entries.Entries, true
			}
			return nil, false
		}).
		Encoder(func(output *multimapv1.EntriesOutput) *multimapv1.MultiMapOutput {
			return &multimapv1.MultiMapOutput{
				Output: &multimapv1.MultiMapOutput_Entries{
					Entries: output,
				},
			}
		}).
		Build(s.doEntries)
}

func (s *MultiMapStateMachine) Snapshot(writer *snapshot.Writer) error {
	s.Log().Infow("Persisting MultiMap to snapshot")
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
	for key, values := range s.entries {
		if err := writer.WriteString(key); err != nil {
			return err
		}
		if err := writer.WriteVarInt(len(values)); err != nil {
			return err
		}
		for value := range values {
			if err := writer.WriteString(value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MultiMapStateMachine) Recover(reader *snapshot.Reader) error {
	s.Log().Infow("Recovering MultiMap from snapshot")
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
		listener := &multimapv1.MultiMapListener{}
		if err := reader.ReadMessage(listener); err != nil {
			return err
		}
		s.listeners[proposal.ID()] = listener
		proposal.Watch(func(phase primitive.ProposalPhase) {
			if phase == primitive.ProposalComplete {
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
		m, err := reader.ReadVarInt()
		if err != nil {
			return err
		}
		values := make(map[string]bool)
		for j := 0; j < m; j++ {
			value, err := reader.ReadString()
			if err != nil {
				return err
			}
			values[value] = true
		}
		s.entries[key] = values
	}
	return nil
}

func (s *MultiMapStateMachine) Propose(proposal primitive.Proposal[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput]) {
	switch proposal.Input().Input.(type) {
	case *multimapv1.MultiMapInput_Put:
		s.put.Execute(proposal)
	case *multimapv1.MultiMapInput_PutAll:
		s.putAll.Execute(proposal)
	case *multimapv1.MultiMapInput_Replace:
		s.replace.Execute(proposal)
	case *multimapv1.MultiMapInput_Remove:
		s.remove.Execute(proposal)
	case *multimapv1.MultiMapInput_RemoveAll:
		s.removeAll.Execute(proposal)
	case *multimapv1.MultiMapInput_Clear:
		s.clear.Execute(proposal)
	case *multimapv1.MultiMapInput_Events:
		s.events.Execute(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *MultiMapStateMachine) doPut(proposal primitive.Proposal[*multimapv1.PutInput, *multimapv1.PutOutput]) {
	defer proposal.Close()

	values, ok := s.entries[proposal.Input().Key]
	if !ok {
		values = make(map[string]bool)
		s.entries[proposal.Input().Key] = values
	}

	if _, ok := values[proposal.Input().Value]; ok {
		proposal.Error(errors.NewAlreadyExists("entry already exists"))
		return
	}

	values[proposal.Input().Value] = true

	s.notify(proposal.Input().Key, proposal.Input().Value, &multimapv1.EventsOutput{
		Event: multimapv1.Event{
			Key: proposal.Input().Key,
			Event: &multimapv1.Event_Added_{
				Added: &multimapv1.Event_Added{
					Value: proposal.Input().Value,
				},
			},
		},
	})
	proposal.Output(&multimapv1.PutOutput{})
}

func (s *MultiMapStateMachine) doPutAll(proposal primitive.Proposal[*multimapv1.PutAllInput, *multimapv1.PutAllOutput]) {
	defer proposal.Close()

	values, ok := s.entries[proposal.Input().Key]
	if !ok {
		values = make(map[string]bool)
		s.entries[proposal.Input().Key] = values
	}

	updated := false
	for _, value := range proposal.Input().Values {
		if _, ok := values[value]; !ok {
			values[value] = true
			s.notify(proposal.Input().Key, value, &multimapv1.EventsOutput{
				Event: multimapv1.Event{
					Key: proposal.Input().Key,
					Event: &multimapv1.Event_Added_{
						Added: &multimapv1.Event_Added{
							Value: value,
						},
					},
				},
			})
			updated = true
		}
	}

	proposal.Output(&multimapv1.PutAllOutput{
		Updated: updated,
	})
}

func (s *MultiMapStateMachine) doReplace(proposal primitive.Proposal[*multimapv1.ReplaceInput, *multimapv1.ReplaceOutput]) {
	defer proposal.Close()

	if len(proposal.Input().Values) == 0 {

	}

	oldValues, ok := s.entries[proposal.Input().Key]
	if !ok {
		oldValues = make(map[string]bool)
	}

	newValues := make(map[string]bool)
	s.entries[proposal.Input().Key] = newValues

	for _, value := range proposal.Input().Values {
		if _, ok := oldValues[value]; !ok {
			s.notify(proposal.Input().Key, value, &multimapv1.EventsOutput{
				Event: multimapv1.Event{
					Key: proposal.Input().Key,
					Event: &multimapv1.Event_Added_{
						Added: &multimapv1.Event_Added{
							Value: value,
						},
					},
				},
			})
		}
		newValues[value] = true
	}

	prevValues := make([]string, 0, len(oldValues))
	for value := range oldValues {
		if _, ok := newValues[value]; !ok {
			s.notify(proposal.Input().Key, value, &multimapv1.EventsOutput{
				Event: multimapv1.Event{
					Key: proposal.Input().Key,
					Event: &multimapv1.Event_Removed_{
						Removed: &multimapv1.Event_Removed{
							Value: value,
						},
					},
				},
			})
		}
		prevValues = append(prevValues, value)
	}

	proposal.Output(&multimapv1.ReplaceOutput{
		PrevValues: prevValues,
	})
}

func (s *MultiMapStateMachine) doRemove(proposal primitive.Proposal[*multimapv1.RemoveInput, *multimapv1.RemoveOutput]) {
	defer proposal.Close()

	values, ok := s.entries[proposal.Input().Key]
	if !ok {
		proposal.Error(errors.NewNotFound("entry not found"))
		return
	}

	if _, ok := values[proposal.Input().Value]; !ok {
		proposal.Error(errors.NewNotFound("entry not found"))
		return
	}

	delete(values, proposal.Input().Value)

	if len(values) == 0 {
		delete(s.entries, proposal.Input().Key)
	}

	s.notify(proposal.Input().Key, proposal.Input().Value, &multimapv1.EventsOutput{
		Event: multimapv1.Event{
			Key: proposal.Input().Key,
			Event: &multimapv1.Event_Removed_{
				Removed: &multimapv1.Event_Removed{
					Value: proposal.Input().Value,
				},
			},
		},
	})
	proposal.Output(&multimapv1.RemoveOutput{})
}

func (s *MultiMapStateMachine) doRemoveAll(proposal primitive.Proposal[*multimapv1.RemoveAllInput, *multimapv1.RemoveAllOutput]) {
	defer proposal.Close()

	values, ok := s.entries[proposal.Input().Key]
	if !ok {
		proposal.Output(&multimapv1.RemoveAllOutput{})
		return
	}

	removedValues := make([]string, 0, len(values))
	for value := range values {
		s.notify(proposal.Input().Key, value, &multimapv1.EventsOutput{
			Event: multimapv1.Event{
				Key: proposal.Input().Key,
				Event: &multimapv1.Event_Removed_{
					Removed: &multimapv1.Event_Removed{
						Value: value,
					},
				},
			},
		})
		removedValues = append(removedValues, value)
	}
	proposal.Output(&multimapv1.RemoveAllOutput{
		Values: removedValues,
	})
}

func (s *MultiMapStateMachine) doClear(proposal primitive.Proposal[*multimapv1.ClearInput, *multimapv1.ClearOutput]) {
	defer proposal.Close()
	for key, values := range s.entries {
		for value := range values {
			s.notify(key, value, &multimapv1.EventsOutput{
				Event: multimapv1.Event{
					Key: key,
					Event: &multimapv1.Event_Removed_{
						Removed: &multimapv1.Event_Removed{
							Value: value,
						},
					},
				},
			})
		}
		delete(s.entries, key)
	}
	proposal.Output(&multimapv1.ClearOutput{})
}

func (s *MultiMapStateMachine) doEvents(proposal primitive.Proposal[*multimapv1.EventsInput, *multimapv1.EventsOutput]) {
	listener := &multimapv1.MultiMapListener{
		Key: proposal.Input().Key,
	}
	s.listeners[proposal.ID()] = listener
	proposal.Watch(func(phase primitive.ProposalPhase) {
		if phase == primitive.ProposalComplete {
			delete(s.listeners, proposal.ID())
		}
	})
}

func (s *MultiMapStateMachine) Query(query primitive.Query[*multimapv1.MultiMapInput, *multimapv1.MultiMapOutput]) {
	switch query.Input().Input.(type) {
	case *multimapv1.MultiMapInput_Size_:
		s.size.Execute(query)
	case *multimapv1.MultiMapInput_Contains:
		s.contains.Execute(query)
	case *multimapv1.MultiMapInput_Get:
		s.get.Execute(query)
	case *multimapv1.MultiMapInput_Entries:
		s.list.Execute(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *MultiMapStateMachine) doSize(query primitive.Query[*multimapv1.SizeInput, *multimapv1.SizeOutput]) {
	defer query.Close()
	query.Output(&multimapv1.SizeOutput{
		Size_: uint32(len(s.entries)),
	})
}

func (s *MultiMapStateMachine) doContains(query primitive.Query[*multimapv1.ContainsInput, *multimapv1.ContainsOutput]) {
	defer query.Close()
	values, ok := s.entries[query.Input().Key]
	if !ok {
		query.Output(&multimapv1.ContainsOutput{
			Result: false,
		})
	} else if _, ok := values[query.Input().Value]; !ok {
		query.Output(&multimapv1.ContainsOutput{
			Result: false,
		})
	} else {
		query.Output(&multimapv1.ContainsOutput{
			Result: true,
		})
	}
}

func (s *MultiMapStateMachine) doGet(query primitive.Query[*multimapv1.GetInput, *multimapv1.GetOutput]) {
	defer query.Close()
	values, ok := s.entries[query.Input().Key]
	if !ok {
		query.Output(&multimapv1.GetOutput{})
	} else {
		getValues := make([]string, 0, len(values))
		for value := range values {
			getValues = append(getValues, value)
		}
		query.Output(&multimapv1.GetOutput{
			Values: getValues,
		})
	}
}

func (s *MultiMapStateMachine) doEntries(query primitive.Query[*multimapv1.EntriesInput, *multimapv1.EntriesOutput]) {
	for key, values := range s.entries {
		for value := range values {
			query.Output(&multimapv1.EntriesOutput{
				Entry: multimapv1.Entry{
					Key:   key,
					Value: value,
				},
			})
		}
	}

	if query.Input().Watch {
		s.mu.Lock()
		s.watchers[query.ID()] = query
		s.mu.Unlock()
		query.Watch(func(phase primitive.QueryPhase) {
			if phase == primitive.QueryComplete {
				s.mu.Lock()
				delete(s.watchers, query.ID())
				s.mu.Unlock()
			}
		})
	} else {
		query.Close()
	}
}

func (s *MultiMapStateMachine) notify(key string, value string, event *multimapv1.EventsOutput) {
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
		watcher.Output(&multimapv1.EntriesOutput{
			Entry: multimapv1.Entry{
				Key:   key,
				Value: value,
			},
		})
	}
}
