// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	setv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/set/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.runtime.set.v1.Set"

func Register(registry *statemachine.PrimitiveTypeRegistry) {
	statemachine.RegisterPrimitiveType[*setv1.SetInput, *setv1.SetOutput](registry)(SetType)
}

var SetType = statemachine.NewPrimitiveType[*setv1.SetInput, *setv1.SetOutput](Service, setCodec, newSetStateMachine)

var setCodec = statemachine.NewCodec[*setv1.SetInput, *setv1.SetOutput](
	func(bytes []byte) (*setv1.SetInput, error) {
		input := &setv1.SetInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *setv1.SetOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newSetStateMachine(ctx statemachine.PrimitiveContext[*setv1.SetInput, *setv1.SetOutput]) statemachine.Primitive[*setv1.SetInput, *setv1.SetOutput] {
	sm := &SetStateMachine{
		PrimitiveContext: ctx,
		listeners:        make(map[statemachine.ProposalID]bool),
		entries:          make(map[string]*setv1.SetElement),
		timers:           make(map[string]statemachine.Timer),
		watchers:         make(map[statemachine.QueryID]statemachine.Query[*setv1.ElementsInput, *setv1.ElementsOutput]),
	}
	sm.init()
	return sm
}

type SetStateMachine struct {
	statemachine.PrimitiveContext[*setv1.SetInput, *setv1.SetOutput]
	listeners map[statemachine.ProposalID]bool
	entries   map[string]*setv1.SetElement
	timers    map[string]statemachine.Timer
	watchers  map[statemachine.QueryID]statemachine.Query[*setv1.ElementsInput, *setv1.ElementsOutput]
	mu        sync.RWMutex
	add       statemachine.Updater[*setv1.SetInput, *setv1.SetOutput, *setv1.AddInput, *setv1.AddOutput]
	remove    statemachine.Updater[*setv1.SetInput, *setv1.SetOutput, *setv1.RemoveInput, *setv1.RemoveOutput]
	clear     statemachine.Updater[*setv1.SetInput, *setv1.SetOutput, *setv1.ClearInput, *setv1.ClearOutput]
	events    statemachine.Updater[*setv1.SetInput, *setv1.SetOutput, *setv1.EventsInput, *setv1.EventsOutput]
	size      statemachine.Reader[*setv1.SetInput, *setv1.SetOutput, *setv1.SizeInput, *setv1.SizeOutput]
	contains  statemachine.Reader[*setv1.SetInput, *setv1.SetOutput, *setv1.ContainsInput, *setv1.ContainsOutput]
	elements  statemachine.Reader[*setv1.SetInput, *setv1.SetOutput, *setv1.ElementsInput, *setv1.ElementsOutput]
}

func (s *SetStateMachine) init() {
	s.add = statemachine.NewUpdater[*setv1.SetInput, *setv1.SetOutput, *setv1.AddInput, *setv1.AddOutput](s).
		Name("Add").
		Decoder(func(input *setv1.SetInput) (*setv1.AddInput, bool) {
			if put, ok := input.Input.(*setv1.SetInput_Add); ok {
				return put.Add, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.AddOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Add{
					Add: output,
				},
			}
		}).
		Build(s.doAdd)
	s.remove = statemachine.NewUpdater[*setv1.SetInput, *setv1.SetOutput, *setv1.RemoveInput, *setv1.RemoveOutput](s).
		Name("Remove").
		Decoder(func(input *setv1.SetInput) (*setv1.RemoveInput, bool) {
			if remove, ok := input.Input.(*setv1.SetInput_Remove); ok {
				return remove.Remove, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.RemoveOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Remove{
					Remove: output,
				},
			}
		}).
		Build(s.doRemove)
	s.clear = statemachine.NewUpdater[*setv1.SetInput, *setv1.SetOutput, *setv1.ClearInput, *setv1.ClearOutput](s).
		Name("Clear").
		Decoder(func(input *setv1.SetInput) (*setv1.ClearInput, bool) {
			if clear, ok := input.Input.(*setv1.SetInput_Clear); ok {
				return clear.Clear, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.ClearOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Clear{
					Clear: output,
				},
			}
		}).
		Build(s.doClear)
	s.events = statemachine.NewUpdater[*setv1.SetInput, *setv1.SetOutput, *setv1.EventsInput, *setv1.EventsOutput](s).
		Name("Events").
		Decoder(func(input *setv1.SetInput) (*setv1.EventsInput, bool) {
			if events, ok := input.Input.(*setv1.SetInput_Events); ok {
				return events.Events, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.EventsOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Events{
					Events: output,
				},
			}
		}).
		Build(s.doEvents)
	s.size = statemachine.NewReader[*setv1.SetInput, *setv1.SetOutput, *setv1.SizeInput, *setv1.SizeOutput](s).
		Name("Size").
		Decoder(func(input *setv1.SetInput) (*setv1.SizeInput, bool) {
			if size, ok := input.Input.(*setv1.SetInput_Size_); ok {
				return size.Size_, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.SizeOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Size_{
					Size_: output,
				},
			}
		}).
		Build(s.doSize)
	s.contains = statemachine.NewReader[*setv1.SetInput, *setv1.SetOutput, *setv1.ContainsInput, *setv1.ContainsOutput](s).
		Name("Contains").
		Decoder(func(input *setv1.SetInput) (*setv1.ContainsInput, bool) {
			if get, ok := input.Input.(*setv1.SetInput_Contains); ok {
				return get.Contains, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.ContainsOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Contains{
					Contains: output,
				},
			}
		}).
		Build(s.doContains)
	s.elements = statemachine.NewReader[*setv1.SetInput, *setv1.SetOutput, *setv1.ElementsInput, *setv1.ElementsOutput](s).
		Name("Elements").
		Decoder(func(input *setv1.SetInput) (*setv1.ElementsInput, bool) {
			if entries, ok := input.Input.(*setv1.SetInput_Elements); ok {
				return entries.Elements, true
			}
			return nil, false
		}).
		Encoder(func(output *setv1.ElementsOutput) *setv1.SetOutput {
			return &setv1.SetOutput{
				Output: &setv1.SetOutput_Elements{
					Elements: output,
				},
			}
		}).
		Build(s.doElements)
}

func (s *SetStateMachine) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(s.listeners)); err != nil {
		return err
	}
	for proposalID := range s.listeners {
		if err := writer.WriteVarUint64(uint64(proposalID)); err != nil {
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

func (s *SetStateMachine) Recover(reader *snapshot.Reader) error {
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
		s.listeners[proposal.ID()] = true
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
		element := &setv1.SetElement{}
		if err := reader.ReadMessage(element); err != nil {
			return err
		}
		s.entries[key] = element
		s.scheduleTTL(key, element)
	}
	return nil
}

func (s *SetStateMachine) Update(proposal statemachine.Proposal[*setv1.SetInput, *setv1.SetOutput]) {
	switch proposal.Input().Input.(type) {
	case *setv1.SetInput_Add:
		s.add.Update(proposal)
	case *setv1.SetInput_Remove:
		s.remove.Update(proposal)
	case *setv1.SetInput_Clear:
		s.clear.Update(proposal)
	case *setv1.SetInput_Events:
		s.events.Update(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *SetStateMachine) doAdd(proposal statemachine.Proposal[*setv1.AddInput, *setv1.AddOutput]) {
	defer proposal.Close()

	value := proposal.Input().Element.Value
	if _, ok := s.entries[value]; ok {
		proposal.Error(errors.NewAlreadyExists("value already exists in set"))
		return
	}

	element := &setv1.SetElement{}
	if proposal.Input().TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().TTL)
		element.Expire = &expire
	}

	// Create a new entry value and set it in the set.
	s.entries[value] = element

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(value, element)

	s.notify(value, &setv1.EventsOutput{
		Event: setv1.Event{
			Event: &setv1.Event_Added_{
				Added: &setv1.Event_Added{
					Element: setv1.Element{
						Value: value,
					},
				},
			},
		},
	})
	proposal.Output(&setv1.AddOutput{})
}

func (s *SetStateMachine) doRemove(proposal statemachine.Proposal[*setv1.RemoveInput, *setv1.RemoveOutput]) {
	defer proposal.Close()

	value := proposal.Input().Element.Value
	if _, ok := s.entries[value]; !ok {
		proposal.Error(errors.NewNotFound("value not found in set"))
		return
	}

	s.cancelTTL(value)
	delete(s.entries, value)

	s.notify(value, &setv1.EventsOutput{
		Event: setv1.Event{
			Event: &setv1.Event_Removed_{
				Removed: &setv1.Event_Removed{
					Element: setv1.Element{
						Value: value,
					},
				},
			},
		},
	})
	proposal.Output(&setv1.RemoveOutput{})
}

func (s *SetStateMachine) doClear(proposal statemachine.Proposal[*setv1.ClearInput, *setv1.ClearOutput]) {
	defer proposal.Close()
	for value := range s.entries {
		s.notify(value, &setv1.EventsOutput{
			Event: setv1.Event{
				Event: &setv1.Event_Removed_{
					Removed: &setv1.Event_Removed{
						Element: setv1.Element{
							Value: value,
						},
					},
				},
			},
		})
		s.cancelTTL(value)
		delete(s.entries, value)
	}
	proposal.Output(&setv1.ClearOutput{})
}

func (s *SetStateMachine) doEvents(proposal statemachine.Proposal[*setv1.EventsInput, *setv1.EventsOutput]) {
	// Output an empty event to ack the request
	proposal.Output(&setv1.EventsOutput{})

	s.listeners[proposal.ID()] = true
	proposal.Watch(func(state statemachine.OperationState) {
		if state == statemachine.Complete {
			delete(s.listeners, proposal.ID())
		}
	})
}

func (s *SetStateMachine) Read(query statemachine.Query[*setv1.SetInput, *setv1.SetOutput]) {
	switch query.Input().Input.(type) {
	case *setv1.SetInput_Size_:
		s.size.Read(query)
	case *setv1.SetInput_Contains:
		s.contains.Read(query)
	case *setv1.SetInput_Elements:
		s.elements.Read(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *SetStateMachine) doSize(query statemachine.Query[*setv1.SizeInput, *setv1.SizeOutput]) {
	defer query.Close()
	query.Output(&setv1.SizeOutput{
		Size_: uint32(len(s.entries)),
	})
}

func (s *SetStateMachine) doContains(query statemachine.Query[*setv1.ContainsInput, *setv1.ContainsOutput]) {
	defer query.Close()
	if _, ok := s.entries[query.Input().Element.Value]; ok {
		query.Output(&setv1.ContainsOutput{
			Contains: true,
		})
	} else {
		query.Output(&setv1.ContainsOutput{
			Contains: false,
		})
	}
}

func (s *SetStateMachine) doElements(query statemachine.Query[*setv1.ElementsInput, *setv1.ElementsOutput]) {
	for value := range s.entries {
		query.Output(&setv1.ElementsOutput{
			Element: setv1.Element{
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

func (s *SetStateMachine) notify(value string, event *setv1.EventsOutput) {
	for proposalID := range s.listeners {
		proposal, ok := s.events.Proposals().Get(proposalID)
		if ok {
			proposal.Output(event)
		} else {
			delete(s.listeners, proposalID)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, watcher := range s.watchers {
		watcher.Output(&setv1.ElementsOutput{
			Element: setv1.Element{
				Value: value,
			},
		})
	}
}

func (s *SetStateMachine) scheduleTTL(value string, element *setv1.SetElement) {
	s.cancelTTL(value)
	if element.Expire != nil {
		s.timers[value] = s.Scheduler().RunAt(*element.Expire, func() {
			delete(s.entries, value)
			s.notify(value, &setv1.EventsOutput{
				Event: setv1.Event{
					Event: &setv1.Event_Removed_{
						Removed: &setv1.Event_Removed{
							Element: setv1.Element{
								Value: value,
							},
							Expired: true,
						},
					},
				},
			})
		})
	}
}

func (s *SetStateMachine) cancelTTL(key string) {
	timer, ok := s.timers[key]
	if ok {
		timer.Cancel()
	}
}
