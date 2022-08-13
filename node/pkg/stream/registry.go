// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package stream

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"sync"
)

// NewRegistry returns a new stream registry
func NewRegistry() *Registry {
	return &Registry{
		streams: make(map[multiraftv1.StreamId]streams.WriteStream[*multiraftv1.CommandOutput]),
	}
}

// Registry is a registry of client streams
type Registry struct {
	streams         map[multiraftv1.StreamId]streams.WriteStream[*multiraftv1.CommandOutput]
	nextSequenceNum multiraftv1.SequenceNum
	mu              sync.RWMutex
}

// Register adds a new stream
func (r *Registry) Register(term multiraftv1.Term, stream streams.WriteStream[*multiraftv1.CommandOutput]) multiraftv1.StreamId {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextSequenceNum++
	sequenceNum := r.nextSequenceNum
	streamID := multiraftv1.StreamId{
		Term:        term,
		SequenceNum: sequenceNum,
	}
	r.streams[streamID] = streams.NewCloserStream[*multiraftv1.CommandOutput](stream, func(s streams.WriteStream[*multiraftv1.CommandOutput]) {
		r.Unregister(streamID)
	})
	return streamID
}

// Unregister removes a stream by ID
func (r *Registry) Unregister(streamID multiraftv1.StreamId) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.streams, streamID)
}

// Lookup gets a stream by ID
func (r *Registry) Lookup(streamID multiraftv1.StreamId) streams.WriteStream[*multiraftv1.CommandOutput] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if stream, ok := r.streams[streamID]; ok {
		return stream
	}
	return streams.NewNilStream[*multiraftv1.CommandOutput]()
}

type Query struct {
	Input   *multiraftv1.QueryInput
	Stream  streams.WriteStream[*multiraftv1.QueryOutput]
	Context context.Context
}
