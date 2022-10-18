// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	indexedmapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/indexedmap/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const Service = "atomix.runtime.indexedmap.v1.IndexedMap"

func Register(registry *primitive.TypeRegistry) {
	primitive.RegisterType[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput](registry)(Type)
}

var Type = primitive.NewType[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput](Service, indexedMapCodec, newIndexedMapStateMachine)

var indexedMapCodec = primitive.NewCodec[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput](
	func(bytes []byte) (*indexedmapv1.IndexedMapInput, error) {
		input := &indexedmapv1.IndexedMapInput{}
		if err := proto.Unmarshal(bytes, input); err != nil {
			return nil, err
		}
		return input, nil
	},
	func(output *indexedmapv1.IndexedMapOutput) ([]byte, error) {
		return proto.Marshal(output)
	})

func newIndexedMapStateMachine(ctx primitive.Context[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput]) primitive.Primitive[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput] {
	sm := &IndexedMapStateMachine{
		Context: ctx,
	}
	sm.init()
	return sm
}

// LinkedMapEntryValue is a doubly linked MapEntryValue
type LinkedMapEntryValue struct {
	*indexedmapv1.IndexedMapEntry
	Prev *LinkedMapEntryValue
	Next *LinkedMapEntryValue
}

type IndexedMapStateMachine struct {
	primitive.Context[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput]
	lastIndex  uint64
	keys       map[string]*LinkedMapEntryValue
	indexes    map[uint64]*LinkedMapEntryValue
	firstEntry *LinkedMapEntryValue
	lastEntry  *LinkedMapEntryValue
	streams    map[primitive.ProposalID]*indexedmapv1.IndexedMapListener
	timers     map[string]primitive.CancelFunc
	watchers   map[primitive.QueryID]primitive.Query[*indexedmapv1.EntriesInput, *indexedmapv1.EntriesOutput]
	mu         sync.RWMutex
	append     primitive.Proposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.AppendInput, *indexedmapv1.AppendOutput]
	update     primitive.Proposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.UpdateInput, *indexedmapv1.UpdateOutput]
	remove     primitive.Proposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.RemoveInput, *indexedmapv1.RemoveOutput]
	clear      primitive.Proposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.ClearInput, *indexedmapv1.ClearOutput]
	events     primitive.Proposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.EventsInput, *indexedmapv1.EventsOutput]
	size       primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.SizeInput, *indexedmapv1.SizeOutput]
	get        primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.GetInput, *indexedmapv1.GetOutput]
	first      primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.FirstEntryInput, *indexedmapv1.FirstEntryOutput]
	last       primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.LastEntryInput, *indexedmapv1.LastEntryOutput]
	next       primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.NextEntryInput, *indexedmapv1.NextEntryOutput]
	prev       primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.PrevEntryInput, *indexedmapv1.PrevEntryOutput]
	list       primitive.Querier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.EntriesInput, *indexedmapv1.EntriesOutput]
}

func (s *IndexedMapStateMachine) init() {
	s.reset()
	s.append = primitive.NewProposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.AppendInput, *indexedmapv1.AppendOutput](s).
		Name("Append").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.AppendInput, bool) {
			if put, ok := input.Input.(*indexedmapv1.IndexedMapInput_Append); ok {
				return put.Append, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.AppendOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Append{
					Append: output,
				},
			}
		}).
		Build(s.doAppend)
	s.update = primitive.NewProposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.UpdateInput, *indexedmapv1.UpdateOutput](s).
		Name("Update").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.UpdateInput, bool) {
			if update, ok := input.Input.(*indexedmapv1.IndexedMapInput_Update); ok {
				return update.Update, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.UpdateOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Update{
					Update: output,
				},
			}
		}).
		Build(s.doUpdate)
	s.remove = primitive.NewProposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.RemoveInput, *indexedmapv1.RemoveOutput](s).
		Name("Remove").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.RemoveInput, bool) {
			if remove, ok := input.Input.(*indexedmapv1.IndexedMapInput_Remove); ok {
				return remove.Remove, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.RemoveOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Remove{
					Remove: output,
				},
			}
		}).
		Build(s.doRemove)
	s.clear = primitive.NewProposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.ClearInput, *indexedmapv1.ClearOutput](s).
		Name("Clear").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.ClearInput, bool) {
			if clear, ok := input.Input.(*indexedmapv1.IndexedMapInput_Clear); ok {
				return clear.Clear, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.ClearOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Clear{
					Clear: output,
				},
			}
		}).
		Build(s.doClear)
	s.events = primitive.NewProposer[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.EventsInput, *indexedmapv1.EventsOutput](s).
		Name("Events").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.EventsInput, bool) {
			if events, ok := input.Input.(*indexedmapv1.IndexedMapInput_Events); ok {
				return events.Events, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.EventsOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Events{
					Events: output,
				},
			}
		}).
		Build(s.doEvents)
	s.size = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.SizeInput, *indexedmapv1.SizeOutput](s).
		Name("Size").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.SizeInput, bool) {
			if size, ok := input.Input.(*indexedmapv1.IndexedMapInput_Size_); ok {
				return size.Size_, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.SizeOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Size_{
					Size_: output,
				},
			}
		}).
		Build(s.doSize)
	s.get = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.GetInput, *indexedmapv1.GetOutput](s).
		Name("Get").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.GetInput, bool) {
			if get, ok := input.Input.(*indexedmapv1.IndexedMapInput_Get); ok {
				return get.Get, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.GetOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Get{
					Get: output,
				},
			}
		}).
		Build(s.doGet)
	s.first = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.FirstEntryInput, *indexedmapv1.FirstEntryOutput](s).
		Name("FirstEntry").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.FirstEntryInput, bool) {
			if get, ok := input.Input.(*indexedmapv1.IndexedMapInput_FirstEntry); ok {
				return get.FirstEntry, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.FirstEntryOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_FirstEntry{
					FirstEntry: output,
				},
			}
		}).
		Build(s.doFirstEntry)
	s.last = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.LastEntryInput, *indexedmapv1.LastEntryOutput](s).
		Name("LastEntry").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.LastEntryInput, bool) {
			if get, ok := input.Input.(*indexedmapv1.IndexedMapInput_LastEntry); ok {
				return get.LastEntry, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.LastEntryOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_LastEntry{
					LastEntry: output,
				},
			}
		}).
		Build(s.doLastEntry)
	s.next = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.NextEntryInput, *indexedmapv1.NextEntryOutput](s).
		Name("NextEntry").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.NextEntryInput, bool) {
			if get, ok := input.Input.(*indexedmapv1.IndexedMapInput_NextEntry); ok {
				return get.NextEntry, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.NextEntryOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_NextEntry{
					NextEntry: output,
				},
			}
		}).
		Build(s.doNextEntry)
	s.prev = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.PrevEntryInput, *indexedmapv1.PrevEntryOutput](s).
		Name("PrevEntry").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.PrevEntryInput, bool) {
			if get, ok := input.Input.(*indexedmapv1.IndexedMapInput_PrevEntry); ok {
				return get.PrevEntry, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.PrevEntryOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_PrevEntry{
					PrevEntry: output,
				},
			}
		}).
		Build(s.doPrevEntry)
	s.list = primitive.NewQuerier[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput, *indexedmapv1.EntriesInput, *indexedmapv1.EntriesOutput](s).
		Name("Entries").
		Decoder(func(input *indexedmapv1.IndexedMapInput) (*indexedmapv1.EntriesInput, bool) {
			if entries, ok := input.Input.(*indexedmapv1.IndexedMapInput_Entries); ok {
				return entries.Entries, true
			}
			return nil, false
		}).
		Encoder(func(output *indexedmapv1.EntriesOutput) *indexedmapv1.IndexedMapOutput {
			return &indexedmapv1.IndexedMapOutput{
				Output: &indexedmapv1.IndexedMapOutput_Entries{
					Entries: output,
				},
			}
		}).
		Build(s.doEntries)
}

func (s *IndexedMapStateMachine) reset() {
	if s.timers != nil {
		for _, cancel := range s.timers {
			cancel()
		}
	}
	s.timers = make(map[string]primitive.CancelFunc)
	if s.watchers != nil {
		for _, watcher := range s.watchers {
			watcher.Cancel()
		}
	}
	s.watchers = make(map[primitive.QueryID]primitive.Query[*indexedmapv1.EntriesInput, *indexedmapv1.EntriesOutput])
	s.streams = make(map[primitive.ProposalID]*indexedmapv1.IndexedMapListener)
	s.keys = make(map[string]*LinkedMapEntryValue)
	s.indexes = make(map[uint64]*LinkedMapEntryValue)
	s.firstEntry = nil
	s.lastEntry = nil
}

func (s *IndexedMapStateMachine) Snapshot(writer *snapshot.Writer) error {
	s.Log().Infow("Persisting IndexedMap to snapshot")
	if err := s.snapshotEntries(writer); err != nil {
		return err
	}
	if err := s.snapshotStreams(writer); err != nil {
		return err
	}
	return nil
}

func (s *IndexedMapStateMachine) snapshotEntries(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(s.keys)); err != nil {
		return err
	}
	entry := s.firstEntry
	for entry != nil {
		if err := writer.WriteMessage(s.firstEntry); err != nil {
			return err
		}
		entry = entry.Next
	}
	if err := writer.WriteVarUint64(s.lastIndex); err != nil {
		return err
	}
	return nil
}

func (s *IndexedMapStateMachine) snapshotStreams(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(s.streams)); err != nil {
		return err
	}
	for proposalID, listener := range s.streams {
		if err := writer.WriteVarUint64(uint64(proposalID)); err != nil {
			return err
		}
		if err := writer.WriteMessage(listener); err != nil {
			return err
		}
	}
	return nil
}

func (s *IndexedMapStateMachine) Recover(reader *snapshot.Reader) error {
	s.Log().Infow("Recovering IndexedMap from snapshot")
	s.reset()
	if err := s.recoverEntries(reader); err != nil {
		return err
	}
	if err := s.recoverStreams(reader); err != nil {
		return err
	}
	return nil
}

func (s *IndexedMapStateMachine) recoverEntries(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}

	var prevEntry *LinkedMapEntryValue
	for i := 0; i < n; i++ {
		entry := &LinkedMapEntryValue{}
		if err := reader.ReadMessage(entry); err != nil {
			return err
		}
		s.keys[entry.Key] = entry
		s.indexes[entry.Index] = entry
		if s.firstEntry == nil {
			s.firstEntry = entry
		}
		if prevEntry != nil {
			prevEntry.Next = entry
			entry.Prev = prevEntry
		}
		prevEntry = entry
		s.lastEntry = entry
		s.scheduleTTL(entry)
	}

	i, err := reader.ReadVarUint64()
	if err != nil {
		return err
	}
	s.lastIndex = i
	return nil
}

func (s *IndexedMapStateMachine) recoverStreams(reader *snapshot.Reader) error {
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
		listener := &indexedmapv1.IndexedMapListener{}
		if err := reader.ReadMessage(listener); err != nil {
			return err
		}
		s.streams[proposal.ID()] = listener
		proposal.Watch(func(state primitive.ProposalState) {
			if primitive.IsDone(state) {
				delete(s.streams, proposal.ID())
			}
		})
	}
	return nil
}

func (s *IndexedMapStateMachine) Propose(proposal primitive.Proposal[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput]) {
	switch proposal.Input().Input.(type) {
	case *indexedmapv1.IndexedMapInput_Append:
		s.append.Execute(proposal)
	case *indexedmapv1.IndexedMapInput_Update:
		s.update.Execute(proposal)
	case *indexedmapv1.IndexedMapInput_Remove:
		s.remove.Execute(proposal)
	case *indexedmapv1.IndexedMapInput_Clear:
		s.clear.Execute(proposal)
	case *indexedmapv1.IndexedMapInput_Events:
		s.events.Execute(proposal)
	default:
		proposal.Error(errors.NewNotSupported("proposal not supported"))
		proposal.Close()
	}
}

func (s *IndexedMapStateMachine) doAppend(proposal primitive.Proposal[*indexedmapv1.AppendInput, *indexedmapv1.AppendOutput]) {
	defer proposal.Close()

	// Check that the key does not already exist in the map
	if entry, ok := s.keys[proposal.Input().Key]; ok {
		proposal.Error(errors.NewAlreadyExists("key %s already exists at index %d", proposal.Input().Key, entry.Index))
		return
	}

	// Increment the map index
	s.lastIndex++
	index := s.lastIndex

	// Create a new entry value and set it in the map.
	entry := &LinkedMapEntryValue{
		IndexedMapEntry: &indexedmapv1.IndexedMapEntry{
			Index: index,
			Key:   proposal.Input().Key,
			Value: &indexedmapv1.IndexedMapValue{
				Value:   proposal.Input().Value,
				Version: uint64(proposal.ID()),
			},
		},
	}
	if proposal.Input().TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().TTL)
		entry.Value.Expire = &expire
	}
	s.keys[entry.Key] = entry
	s.indexes[entry.Index] = entry

	// Set the first entry if not set
	if s.firstEntry == nil {
		s.firstEntry = entry
	}

	// If the last entry is set, link it to the new entry
	if s.lastEntry != nil {
		s.lastEntry.Next = entry
		entry.Prev = s.lastEntry
	}

	// Update the last entry
	s.lastEntry = entry

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(entry)

	s.notify(entry, &indexedmapv1.Event{
		Key:   entry.Key,
		Index: entry.Index,
		Event: &indexedmapv1.Event_Inserted_{
			Inserted: &indexedmapv1.Event_Inserted{
				Value: *newValue(entry.Value),
			},
		},
	})

	proposal.Output(&indexedmapv1.AppendOutput{
		Entry: newEntry(entry.IndexedMapEntry),
	})
}

func (s *IndexedMapStateMachine) doUpdate(proposal primitive.Proposal[*indexedmapv1.UpdateInput, *indexedmapv1.UpdateOutput]) {
	defer proposal.Close()

	// Get the current entry value by key, index, or both
	var oldEntry *LinkedMapEntryValue
	if proposal.Input().Index != 0 {
		if e, ok := s.indexes[proposal.Input().Index]; !ok {
			proposal.Error(errors.NewNotFound("index %d not found", proposal.Input().Index))
			return
		} else if proposal.Input().Key != "" && e.Key != proposal.Input().Key {
			proposal.Error(errors.NewFault("key at index %d does not match proposed Update key %s", proposal.Input().Index, proposal.Input().Key))
			return
		} else {
			oldEntry = e
		}
	} else if proposal.Input().Key != "" {
		if e, ok := s.keys[proposal.Input().Key]; !ok {
			proposal.Error(errors.NewNotFound("key %s not found", proposal.Input().Key))
			return
		} else {
			oldEntry = e
		}
	} else {
		proposal.Error(errors.NewInvalid("must specify either a key or index to update"))
		return
	}

	// If a prev_version was specified, check that the previous version matches
	if proposal.Input().PrevVersion != 0 && oldEntry.Value.Version != proposal.Input().PrevVersion {
		proposal.Error(errors.NewConflict("key %s version %d does not match prev_version %d", oldEntry.Key, oldEntry.Value.Version, proposal.Input().PrevVersion))
		return
	}

	// If the value is equal to the current value, return a no-op.
	if bytes.Equal(oldEntry.Value.Value, proposal.Input().Value) {
		proposal.Output(&indexedmapv1.UpdateOutput{
			Entry: newEntry(oldEntry.IndexedMapEntry),
		})
		return
	}

	// Create a new entry value and set it in the map.
	entry := &LinkedMapEntryValue{
		IndexedMapEntry: &indexedmapv1.IndexedMapEntry{
			Index: oldEntry.Index,
			Key:   oldEntry.Key,
			Value: &indexedmapv1.IndexedMapValue{
				Value:   proposal.Input().Value,
				Version: uint64(proposal.ID()),
			},
		},
		Prev: oldEntry.Prev,
		Next: oldEntry.Next,
	}
	if proposal.Input().TTL != nil {
		expire := s.Scheduler().Time().Add(*proposal.Input().TTL)
		entry.Value.Expire = &expire
	}
	s.keys[entry.Key] = entry
	s.indexes[entry.Index] = entry

	// Update links for previous and next entries
	if oldEntry.Prev != nil {
		oldEntry.Prev.Next = entry
	} else {
		s.firstEntry = entry
	}
	if oldEntry.Next != nil {
		oldEntry.Next.Prev = entry
	} else {
		s.lastEntry = entry
	}

	// Schedule the timeout for the value if necessary.
	s.scheduleTTL(entry)

	s.notify(entry, &indexedmapv1.Event{
		Key:   entry.Key,
		Index: entry.Index,
		Event: &indexedmapv1.Event_Updated_{
			Updated: &indexedmapv1.Event_Updated{
				Value: *newValue(entry.Value),
			},
		},
	})

	proposal.Output(&indexedmapv1.UpdateOutput{
		Entry: newEntry(entry.IndexedMapEntry),
	})
}

func (s *IndexedMapStateMachine) doRemove(proposal primitive.Proposal[*indexedmapv1.RemoveInput, *indexedmapv1.RemoveOutput]) {
	defer proposal.Close()

	var entry *LinkedMapEntryValue
	var ok bool
	if proposal.Input().Index != 0 {
		if entry, ok = s.indexes[proposal.Input().Index]; !ok {
			proposal.Error(errors.NewNotFound("no entry found at index %d", proposal.Input().Index))
			return
		}
	} else {
		if entry, ok = s.keys[proposal.Input().Key]; !ok {
			proposal.Error(errors.NewNotFound("no entry found at key %s", proposal.Input().Key))
			return
		}
	}

	if proposal.Input().PrevVersion != 0 && entry.Value.Version != proposal.Input().PrevVersion {
		proposal.Error(errors.NewConflict("key %s version %d does not match prev_version %d", entry.Key, entry.Value.Version, proposal.Input().PrevVersion))
		return
	}

	// Delete the entry from the map.
	delete(s.keys, entry.Key)
	delete(s.indexes, entry.Index)

	// Cancel any TTLs.
	s.cancelTTL(proposal.Input().Key)

	// Update links for previous and next entries
	if entry.Prev != nil {
		entry.Prev.Next = entry.Next
	} else {
		s.firstEntry = entry.Next
	}
	if entry.Next != nil {
		entry.Next.Prev = entry.Prev
	} else {
		s.lastEntry = entry.Prev
	}

	s.notify(entry, &indexedmapv1.Event{
		Key:   entry.Key,
		Index: entry.Index,
		Event: &indexedmapv1.Event_Removed_{
			Removed: &indexedmapv1.Event_Removed{
				Value: *newValue(entry.IndexedMapEntry.Value),
			},
		},
	})

	proposal.Output(&indexedmapv1.RemoveOutput{
		Entry: newEntry(entry.IndexedMapEntry),
	})
}

func (s *IndexedMapStateMachine) doClear(proposal primitive.Proposal[*indexedmapv1.ClearInput, *indexedmapv1.ClearOutput]) {
	defer proposal.Close()
	for key, entry := range s.keys {
		s.notify(entry, &indexedmapv1.Event{
			Key:   entry.Key,
			Index: entry.Index,
			Event: &indexedmapv1.Event_Removed_{
				Removed: &indexedmapv1.Event_Removed{
					Value: *newValue(entry.IndexedMapEntry.Value),
				},
			},
		})
		s.cancelTTL(key)
	}
	s.keys = make(map[string]*LinkedMapEntryValue)
	s.indexes = make(map[uint64]*LinkedMapEntryValue)
	s.firstEntry = nil
	s.lastEntry = nil
	proposal.Output(&indexedmapv1.ClearOutput{})
}

func (s *IndexedMapStateMachine) doEvents(proposal primitive.Proposal[*indexedmapv1.EventsInput, *indexedmapv1.EventsOutput]) {
	listener := &indexedmapv1.IndexedMapListener{
		Key: proposal.Input().Key,
	}
	s.streams[proposal.ID()] = listener
	proposal.Watch(func(state primitive.ProposalState) {
		if primitive.IsDone(state) {
			delete(s.streams, proposal.ID())
		}
	})
}

func (s *IndexedMapStateMachine) Query(query primitive.Query[*indexedmapv1.IndexedMapInput, *indexedmapv1.IndexedMapOutput]) {
	switch query.Input().Input.(type) {
	case *indexedmapv1.IndexedMapInput_Size_:
		s.size.Execute(query)
	case *indexedmapv1.IndexedMapInput_Get:
		s.get.Execute(query)
	case *indexedmapv1.IndexedMapInput_FirstEntry:
		s.first.Execute(query)
	case *indexedmapv1.IndexedMapInput_LastEntry:
		s.last.Execute(query)
	case *indexedmapv1.IndexedMapInput_NextEntry:
		s.next.Execute(query)
	case *indexedmapv1.IndexedMapInput_PrevEntry:
		s.prev.Execute(query)
	case *indexedmapv1.IndexedMapInput_Entries:
		s.list.Execute(query)
	default:
		query.Error(errors.NewNotSupported("query not supported"))
	}
}

func (s *IndexedMapStateMachine) doSize(query primitive.Query[*indexedmapv1.SizeInput, *indexedmapv1.SizeOutput]) {
	defer query.Close()
	query.Output(&indexedmapv1.SizeOutput{
		Size_: uint32(len(s.keys)),
	})
}

func (s *IndexedMapStateMachine) doGet(query primitive.Query[*indexedmapv1.GetInput, *indexedmapv1.GetOutput]) {
	defer query.Close()

	var entry *LinkedMapEntryValue
	var ok bool
	if query.Input().Index > 0 {
		if entry, ok = s.indexes[query.Input().Index]; !ok {
			query.Error(errors.NewNotFound("no entry found at index %d", query.Input().Index))
			return
		}
	} else {
		if entry, ok = s.keys[query.Input().Key]; !ok {
			query.Error(errors.NewNotFound("no entry found at key %s", query.Input().Key))
			return
		}
	}

	query.Output(&indexedmapv1.GetOutput{
		Entry: newEntry(entry.IndexedMapEntry),
	})
}

func (s *IndexedMapStateMachine) doFirstEntry(query primitive.Query[*indexedmapv1.FirstEntryInput, *indexedmapv1.FirstEntryOutput]) {
	defer query.Close()
	if s.firstEntry == nil {
		query.Error(errors.NewNotFound("map is empty"))
	} else {
		query.Output(&indexedmapv1.FirstEntryOutput{
			Entry: newEntry(s.firstEntry.IndexedMapEntry),
		})
	}
}

func (s *IndexedMapStateMachine) doLastEntry(query primitive.Query[*indexedmapv1.LastEntryInput, *indexedmapv1.LastEntryOutput]) {
	defer query.Close()
	if s.lastEntry == nil {
		query.Error(errors.NewNotFound("map is empty"))
	} else {
		query.Output(&indexedmapv1.LastEntryOutput{
			Entry: newEntry(s.lastEntry.IndexedMapEntry),
		})
	}
}

func (s *IndexedMapStateMachine) doNextEntry(query primitive.Query[*indexedmapv1.NextEntryInput, *indexedmapv1.NextEntryOutput]) {
	defer query.Close()
	entry, ok := s.indexes[query.Input().Index]
	if !ok {
		entry = s.firstEntry
		if entry == nil {
			query.Error(errors.NewNotFound("map is empty"))
			return
		}
		for entry != nil && entry.Index >= query.Input().Index {
			entry = entry.Next
		}
	}
	entry = entry.Next
	if entry == nil {
		query.Error(errors.NewNotFound("no entry found after index %d", query.Input().Index))
	} else {
		query.Output(&indexedmapv1.NextEntryOutput{
			Entry: newEntry(entry.IndexedMapEntry),
		})
	}
}

func (s *IndexedMapStateMachine) doPrevEntry(query primitive.Query[*indexedmapv1.PrevEntryInput, *indexedmapv1.PrevEntryOutput]) {
	defer query.Close()
	entry, ok := s.indexes[query.Input().Index]
	if !ok {
		entry = s.lastEntry
		if entry == nil {
			query.Error(errors.NewNotFound("map is empty"))
			return
		}
		for entry != nil && entry.Index >= query.Input().Index {
			entry = entry.Prev
		}
	}
	entry = entry.Prev
	if entry == nil {
		query.Error(errors.NewNotFound("no entry found prior to index %d", query.Input().Index))
	} else {
		query.Output(&indexedmapv1.PrevEntryOutput{
			Entry: newEntry(entry.IndexedMapEntry),
		})
	}
}

func (s *IndexedMapStateMachine) doEntries(query primitive.Query[*indexedmapv1.EntriesInput, *indexedmapv1.EntriesOutput]) {
	defer query.Close()
	entry := s.firstEntry
	for entry != nil {
		query.Output(&indexedmapv1.EntriesOutput{
			Entry: *newEntry(entry.IndexedMapEntry),
		})
		entry = entry.Next
	}

	if query.Input().Watch {
		s.mu.Lock()
		s.watchers[query.ID()] = query
		s.mu.Unlock()
		query.Watch(func(state primitive.QueryState) {
			if primitive.IsDone(state) {
				s.mu.Lock()
				delete(s.watchers, query.ID())
				s.mu.Unlock()
			}
		})
	} else {
		query.Close()
	}
}

func (s *IndexedMapStateMachine) notify(entry *LinkedMapEntryValue, event *indexedmapv1.Event) {
	for proposalID, listener := range s.streams {
		if listener.Key == "" || listener.Key == event.Key {
			proposal, ok := s.events.Proposals().Get(proposalID)
			if ok {
				proposal.Output(&indexedmapv1.EventsOutput{
					Event: *event,
				})
			} else {
				delete(s.streams, proposalID)
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, watcher := range s.watchers {
		watcher.Output(&indexedmapv1.EntriesOutput{
			Entry: *newEntry(entry.IndexedMapEntry),
		})
	}
}

func (s *IndexedMapStateMachine) scheduleTTL(entry *LinkedMapEntryValue) {
	s.cancelTTL(entry.Key)
	if entry.Value.Expire != nil {
		s.timers[entry.Key] = s.Scheduler().Schedule(*entry.Value.Expire, func() {
			// Delete the entry from the key/index maps
			delete(s.keys, entry.Key)
			delete(s.indexes, entry.Index)

			// Update links for previous and next entries
			if entry.Prev != nil {
				entry.Prev.Next = entry.Next
			} else {
				s.firstEntry = entry.Next
			}
			if entry.Next != nil {
				entry.Next.Prev = entry.Prev
			} else {
				s.lastEntry = entry.Prev
			}

			// Notify watchers of the removal
			s.notify(entry, &indexedmapv1.Event{
				Key:   entry.Key,
				Index: entry.Index,
				Event: &indexedmapv1.Event_Removed_{
					Removed: &indexedmapv1.Event_Removed{
						Value: indexedmapv1.Value{
							Value:   entry.Value.Value,
							Version: entry.Value.Version,
						},
						Expired: true,
					},
				},
			})
		})
	}
}

func (s *IndexedMapStateMachine) cancelTTL(key string) {
	ttlCancelFunc, ok := s.timers[key]
	if ok {
		ttlCancelFunc()
	}
}

func newEntry(entry *indexedmapv1.IndexedMapEntry) *indexedmapv1.Entry {
	return &indexedmapv1.Entry{
		Key:   entry.Key,
		Index: entry.Index,
		Value: newValue(entry.Value),
	}
}

func newValue(value *indexedmapv1.IndexedMapValue) *indexedmapv1.Value {
	if value == nil {
		return nil
	}
	return &indexedmapv1.Value{
		Value:   value.Value,
		Version: value.Version,
	}
}
