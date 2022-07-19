// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"google.golang.org/grpc"
	"sync"
	"sync/atomic"
	"time"
)

func newSessionClient(partition *PartitionClient) *SessionClient {
	session := &SessionClient{
		partition: partition,
	}
	session.recorder = &Recorder{
		session: session,
	}
	return session
}

type SessionClient struct {
	partition    *PartitionClient
	Timeout      time.Duration
	sessionID    multiraftv1.SessionID
	lastIndex    *sessionIndex
	requestNum   *sessionRequestID
	requestCh    chan sessionRequestEvent
	primitives   map[string]*PrimitiveClient
	primitivesMu sync.RWMutex
	recorder     *Recorder
}

func (s *SessionClient) CreatePrimitive(ctx context.Context, spec multiraftv1.PrimitiveSpec, opts ...grpc.CallOption) error {

}

func (s *SessionClient) GetPrimitive(ctx context.Context, name string) (*PrimitiveClient, error) {

}

func (s *SessionClient) ClosePrimitive(ctx context.Context, name string, opts ...grpc.CallOption) error {

}

func (s *SessionClient) nextRequestNum() multiraftv1.SequenceNum {
	return s.requestNum.Next()
}

func (s *SessionClient) update(index multiraftv1.Index) {
	s.lastIndex.Update(index)
}

type Recorder struct {
	session *SessionClient
}

func (r *Recorder) record(eventType sessionRequestEventType, sequenceNum multiraftv1.SequenceNum) {
	r.session.requestCh <- sessionRequestEvent{
		eventType:  eventType,
		requestNum: sequenceNum,
	}
}

func (r *Recorder) Start(headers *multiraftv1.CommandRequestHeaders) {
	r.record(sessionRequestEventStart, headers.SequenceNum)
}

func (r *Recorder) StreamOpen(headers *multiraftv1.CommandRequestHeaders) {
	r.record(sessionStreamEventOpen, headers.SequenceNum)
}

func (r *Recorder) StreamReceive(headers *multiraftv1.CommandResponseHeaders) {
	r.record(sessionStreamEventReceive, headers.OutputSequenceNum)
}

func (r *Recorder) StreamClose(headers *multiraftv1.CommandRequestHeaders) {
	r.record(sessionStreamEventClose, headers.SequenceNum)
}

func (r *Recorder) End(headers *multiraftv1.CommandRequestHeaders) {
	r.record(sessionRequestEventEnd, headers.SequenceNum)
}

type sessionIndex struct {
	value uint64
}

func (i *sessionIndex) Update(index multiraftv1.Index) {
	update := uint64(index)
	for {
		current := atomic.LoadUint64(&i.value)
		if current < update {
			updated := atomic.CompareAndSwapUint64(&i.value, current, update)
			if updated {
				break
			}
		} else {
			break
		}
	}
}

func (i *sessionIndex) Get() multiraftv1.Index {
	value := atomic.LoadUint64(&i.value)
	return multiraftv1.Index(value)
}

type sessionRequestID struct {
	value uint64
}

func (i *sessionRequestID) Next() multiraftv1.SequenceNum {
	value := atomic.AddUint64(&i.value, 1)
	return multiraftv1.SequenceNum(value)
}

type sessionRequestEventType int

const (
	sessionRequestEventStart sessionRequestEventType = iota
	sessionRequestEventEnd
	sessionStreamEventOpen
	sessionStreamEventReceive
	sessionStreamEventClose
	sessionStreamEventAck
)

type sessionRequestEvent struct {
	requestNum  multiraftv1.SequenceNum
	responseNum multiraftv1.SequenceNum
	eventType   sessionRequestEventType
}

type sessionResponseStream struct {
	currentResponseNum multiraftv1.SequenceNum
	ackedResponseNum   multiraftv1.SequenceNum
}
