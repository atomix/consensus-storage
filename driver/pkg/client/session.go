// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/bits-and-blooms/bloom/v3"
	"google.golang.org/grpc"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const chanBufSize = 1000

// The false positive rate for request/response filters
const fpRate float64 = 0.05

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
	requestNum   *sessionRequestNum
	requestCh    chan sessionRequestEvent
	primitives   map[string]*PrimitiveClient
	primitivesMu sync.RWMutex
	recorder     *Recorder
}

func (s *SessionClient) CreatePrimitive(ctx context.Context, spec multiraftv1.PrimitiveSpec) error {
	s.primitivesMu.Lock()
	defer s.primitivesMu.Unlock()
	primitive, ok := s.primitives[spec.Name]
	if ok {
		return nil
	}
	primitive = newPrimitiveClient(spec, s)
	if err := primitive.create(ctx); err != nil {
		return err
	}
	s.primitives[primitive.spec.Name] = primitive
	return nil
}

func (s *SessionClient) GetPrimitive(name string) (*PrimitiveClient, error) {
	s.primitivesMu.RLock()
	defer s.primitivesMu.RUnlock()
	primitive, ok := s.primitives[name]
	if !ok {
		return nil, errors.NewUnavailable("primitive not found")
	}
	return primitive, nil
}

func (s *SessionClient) ClosePrimitive(ctx context.Context, name string, opts ...grpc.CallOption) error {
	s.primitivesMu.Lock()
	defer s.primitivesMu.Unlock()
	primitive, ok := s.primitives[name]
	if !ok {
		return nil
	}
	if err := primitive.close(ctx); err != nil {
		return err
	}
	delete(s.primitives, name)
	return nil
}

func (s *SessionClient) nextRequestNum() multiraftv1.SequenceNum {
	return s.requestNum.Next()
}

func (s *SessionClient) update(index multiraftv1.Index) {
	s.lastIndex.Update(index)
}

func (s *SessionClient) open(ctx context.Context) error {
	request := &multiraftv1.OpenSessionRequest{
		Headers: multiraftv1.PartitionRequestHeaders{
			PartitionID: s.partition.id,
		},
		OpenSessionInput: multiraftv1.OpenSessionInput{
			Timeout: s.Timeout,
		},
	}

	client := multiraftv1.NewPartitionClient(s.partition.conn)
	response, err := client.OpenSession(ctx, request)
	if err != nil {
		return errors.FromProto(err)
	}

	s.sessionID = response.SessionID

	s.lastIndex = &sessionIndex{}
	s.lastIndex.Update(response.Headers.Index)

	s.requestNum = &sessionRequestNum{}

	s.requestCh = make(chan sessionRequestEvent, chanBufSize)
	go func() {
		ticker := time.NewTicker(s.Timeout / 4)
		var requestNum multiraftv1.SequenceNum
		requests := make(map[multiraftv1.SequenceNum]bool)
		responseStreams := make(map[multiraftv1.SequenceNum]*sessionResponseStream)
		for {
			select {
			case requestEvent := <-s.requestCh:
				switch requestEvent.eventType {
				case sessionRequestEventStart:
					for requestNum < requestEvent.requestNum {
						requestNum++
						requests[requestNum] = true
						log.Debugf("Started request %d", requestNum)
					}
				case sessionRequestEventEnd:
					if requests[requestEvent.requestNum] {
						delete(requests, requestEvent.requestNum)
						log.Debugf("Finished request %d", requestEvent.requestNum)
					}
				case sessionStreamEventOpen:
					responseStreams[requestNum] = &sessionResponseStream{}
					log.Debugf("Opened request %d response stream", requestNum)
				case sessionStreamEventReceive:
					responseStream, ok := responseStreams[requestEvent.requestNum]
					if ok {
						if requestEvent.responseNum == responseStream.currentResponseNum+1 {
							responseStream.currentResponseNum++
							log.Debugf("Received request %d stream response %d", requestEvent.requestNum, requestEvent.responseNum)
						}
					}
				case sessionStreamEventClose:
					delete(responseStreams, requestEvent.requestNum)
					log.Debugf("Closed request %d response stream", requestEvent.requestNum)
				case sessionStreamEventAck:
					responseStream, ok := responseStreams[requestEvent.requestNum]
					if ok {
						if requestEvent.responseNum > responseStream.ackedResponseNum {
							responseStream.ackedResponseNum = requestEvent.responseNum
							log.Debugf("Acked request %d stream responses up to %d", requestEvent.requestNum, requestEvent.responseNum)
						}
					}
				}
			case <-ticker.C:
				openRequests := bloom.NewWithEstimates(uint(len(requests)), fpRate)
				completeResponses := make(map[multiraftv1.SequenceNum]multiraftv1.SequenceNum)
				for requestNum := range requests {
					requestBytes := make([]byte, 8)
					binary.BigEndian.PutUint64(requestBytes, uint64(requestNum))
					openRequests.Add(requestBytes)
				}
				for requestNum, responseStream := range responseStreams {
					if responseStream.currentResponseNum > 1 && responseStream.currentResponseNum > responseStream.ackedResponseNum {
						completeResponses[requestNum] = responseStream.currentResponseNum
					}
				}
				go func(lastRequestNum multiraftv1.SequenceNum) {
					err := s.keepAliveSessions(context.Background(), lastRequestNum, openRequests, completeResponses)
					if err != nil {
						log.Error(err)
					} else {
						for requestNum, responseNum := range completeResponses {
							s.requestCh <- sessionRequestEvent{
								eventType:   sessionStreamEventAck,
								requestNum:  requestNum,
								responseNum: responseNum,
							}
						}
					}
				}(requestNum)
			}
		}
	}()
	return nil
}

func (s *SessionClient) keepAliveSessions(ctx context.Context, lastRequestNum multiraftv1.SequenceNum, openRequests *bloom.BloomFilter, completeResponses map[multiraftv1.SequenceNum]multiraftv1.SequenceNum) error {
	openRequestsBytes, err := json.Marshal(openRequests)
	if err != nil {
		return err
	}

	request := &multiraftv1.KeepAliveRequest{
		Headers: multiraftv1.PartitionRequestHeaders{
			PartitionID: s.partition.id,
		},
		KeepAliveInput: multiraftv1.KeepAliveInput{
			SessionID:              s.sessionID,
			LastInputSequenceNum:   lastRequestNum,
			InputFilter:            openRequestsBytes,
			LastOutputSequenceNums: completeResponses,
		},
	}

	client := multiraftv1.NewPartitionClient(s.partition.conn)
	response, err := client.KeepAlive(ctx, request)
	if err != nil {
		err = errors.FromProto(err)
		if errors.IsFault(err) {
			log.Error("Detected potential data loss: ", err)
			log.Infof("Exiting process...")
			os.Exit(errors.Code(err))
		}
		return errors.NewInternal(err.Error())
	}
	s.lastIndex.Update(response.Headers.Index)
	return nil
}

func (s *SessionClient) close(ctx context.Context) error {
	request := &multiraftv1.CloseSessionRequest{
		Headers: multiraftv1.PartitionRequestHeaders{
			PartitionID: s.partition.id,
		},
		CloseSessionInput: multiraftv1.CloseSessionInput{
			SessionID: s.sessionID,
		},
	}

	client := multiraftv1.NewPartitionClient(s.partition.conn)
	_, err := client.CloseSession(ctx, request)
	if err != nil {
		return errors.FromProto(err)
	}
	return nil
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

type sessionRequestNum struct {
	value uint64
}

func (i *sessionRequestNum) Next() multiraftv1.SequenceNum {
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
