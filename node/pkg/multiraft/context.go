// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package multiraft

import (
	"github.com/atomix/runtime/sdk/pkg/protocol"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"sync"
	"sync/atomic"
)

// newContext returns a new protocol context
func newContext() *protocolContext {
	return &protocolContext{
		streams: make(map[protocolStreamID]streams.WriteStream[*protocol.ProposalOutput]),
	}
}

// protocolContext stores state shared by servers and the state machine
type protocolContext struct {
	streams     map[protocolStreamID]streams.WriteStream[*protocol.ProposalOutput]
	streamsMu   sync.RWMutex
	sequenceNum atomic.Uint64
}

// addStream adds a new stream
func (r *protocolContext) addStream(term Term, stream streams.WriteStream[*protocol.ProposalOutput]) SequenceNum {
	sequenceNum := SequenceNum(r.sequenceNum.Add(1))
	streamID := protocolStreamID{
		term:        term,
		sequenceNum: sequenceNum,
	}
	r.streamsMu.Lock()
	r.streams[streamID] = streams.NewCloserStream[*protocol.ProposalOutput](stream, func(s streams.WriteStream[*protocol.ProposalOutput]) {
		r.removeStream(term, sequenceNum)
	})
	r.streamsMu.Unlock()
	return sequenceNum
}

// removeStream removes a stream by ID
func (r *protocolContext) removeStream(term Term, sequenceNum SequenceNum) {
	r.streamsMu.Lock()
	defer r.streamsMu.Unlock()
	streamID := protocolStreamID{
		term:        term,
		sequenceNum: sequenceNum,
	}
	delete(r.streams, streamID)
}

// getStream gets a stream by ID
func (r *protocolContext) getStream(term Term, sequenceNum SequenceNum) streams.WriteStream[*protocol.ProposalOutput] {
	r.streamsMu.RLock()
	defer r.streamsMu.RUnlock()
	streamID := protocolStreamID{
		term:        term,
		sequenceNum: sequenceNum,
	}
	if stream, ok := r.streams[streamID]; ok {
		return stream
	}
	return streams.NewNilStream[*protocol.ProposalOutput]()
}

type protocolStreamID struct {
	term        Term
	sequenceNum SequenceNum
}

type protocolQuery struct {
	input  *protocol.QueryInput
	stream streams.WriteStream[*protocol.QueryOutput]
}
