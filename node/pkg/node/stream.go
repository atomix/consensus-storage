// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/stream"
	"sync"
)

// newStreamRegistry returns a new stream registry
func newStreamRegistry() *streamRegistry {
	return &streamRegistry{
		streams: make(map[multiraftv1.StreamId]stream.WriteStream[*multiraftv1.CommandOutput]),
	}
}

// streamRegistry is a registry of client streams
type streamRegistry struct {
	streams         map[multiraftv1.StreamId]stream.WriteStream[*multiraftv1.CommandOutput]
	nextSequenceNum multiraftv1.SequenceNum
	mu              sync.RWMutex
}

// register adds a new stream
func (r *streamRegistry) register(term multiraftv1.Term, stream stream.WriteStream[*multiraftv1.CommandOutput]) multiraftv1.StreamId {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextSequenceNum++
	sequenceNum := r.nextSequenceNum
	streamID := multiraftv1.StreamId{
		Term:        term,
		SequenceNum: sequenceNum,
	}
	r.streams[streamID] = stream
	return streamID
}

// unregister removes a stream by ID
func (r *streamRegistry) unregister(streamID multiraftv1.StreamId) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.streams, streamID)
}

// lookup gets a stream by ID
func (r *streamRegistry) lookup(streamID multiraftv1.StreamId) stream.WriteStream[*multiraftv1.CommandOutput] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if stream, ok := r.streams[streamID]; ok {
		return stream
	}
	return stream.NewNilStream[*multiraftv1.CommandOutput]()
}
