// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"sync"
	"sync/atomic"
)

// newContext returns a new protocol context
func newContext() *protocolContext {
	return &protocolContext{
		streams:     make(map[protocolStreamID]streams.WriteStream[*multiraftv1.StateMachineProposalOutput]),
		sequenceNum: &atomic.Uint64{},
	}
}

// protocolContext stores state shared by servers and the state machine
type protocolContext struct {
	streams     map[protocolStreamID]streams.WriteStream[*multiraftv1.StateMachineProposalOutput]
	streamsMu   sync.RWMutex
	sequenceNum *atomic.Uint64
}

// addStream adds a new stream
func (r *protocolContext) addStream(term multiraftv1.Term, stream streams.WriteStream[*multiraftv1.StateMachineProposalOutput]) multiraftv1.SequenceNum {
	sequenceNum := multiraftv1.SequenceNum(r.sequenceNum.Add(1))
	streamID := protocolStreamID{
		term:        term,
		sequenceNum: sequenceNum,
	}
	r.streamsMu.Lock()
	r.streams[streamID] = streams.NewCloserStream[*multiraftv1.StateMachineProposalOutput](stream, func(s streams.WriteStream[*multiraftv1.StateMachineProposalOutput]) {
		r.removeStream(term, sequenceNum)
	})
	r.streamsMu.Unlock()
	return sequenceNum
}

// removeStream removes a stream by ID
func (r *protocolContext) removeStream(term multiraftv1.Term, sequenceNum multiraftv1.SequenceNum) {
	r.streamsMu.Lock()
	defer r.streamsMu.Unlock()
	streamID := protocolStreamID{
		term:        term,
		sequenceNum: sequenceNum,
	}
	delete(r.streams, streamID)
}

// getStream gets a stream by ID
func (r *protocolContext) getStream(term multiraftv1.Term, sequenceNum multiraftv1.SequenceNum) streams.WriteStream[*multiraftv1.StateMachineProposalOutput] {
	r.streamsMu.RLock()
	defer r.streamsMu.RUnlock()
	streamID := protocolStreamID{
		term:        term,
		sequenceNum: sequenceNum,
	}
	if stream, ok := r.streams[streamID]; ok {
		return stream
	}
	return streams.NewNilStream[*multiraftv1.StateMachineProposalOutput]()
}

type protocolStreamID struct {
	term        multiraftv1.Term
	sequenceNum multiraftv1.SequenceNum
}

type protocolQuery struct {
	input  *multiraftv1.StateMachineQueryInput
	stream streams.WriteStream[*multiraftv1.StateMachineQueryOutput]
}
