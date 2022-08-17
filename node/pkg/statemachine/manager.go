// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"container/list"
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"sync"
	"sync/atomic"
	"time"
)

var log = logging.GetLogger()

func newStateManager(registry *PrimitiveTypeRegistry) *stateManager {
	stateMachine := &stateManager{
		pendingQueries: make(map[multiraftv1.Index]*list.List),
		scheduler:      newScheduler(),
		sequenceNum:    &atomic.Uint64{},
	}
	stateMachine.sessions = newSessionManager(registry, stateMachine)
	return stateMachine
}

type stateManager struct {
	sessions        *sessionManager
	pendingQueries  map[multiraftv1.Index]*list.List
	canceledQueries []*raftSessionQuery
	queriesMu       sync.RWMutex
	scheduler       *Scheduler
	index           multiraftv1.Index
	sequenceNum     *atomic.Uint64
	time            time.Time
}

func (s *stateManager) Snapshot(writer *snapshot.Writer) error {
	log.Infow("Persisting state to snapshot",
		logging.Uint64("Index", uint64(s.index)),
		logging.Time("Time", s.time))
	snapshot := &multiraftv1.Snapshot{
		Index:     s.index,
		Timestamp: s.time,
	}
	if err := writer.WriteMessage(snapshot); err != nil {
		return err
	}
	return s.sessions.snapshot(writer)
}

func (s *stateManager) Recover(reader *snapshot.Reader) error {
	snapshot := &multiraftv1.Snapshot{}
	if err := reader.ReadMessage(snapshot); err != nil {
		return err
	}
	log.Infow("Recovering state from snapshot",
		logging.Uint64("Index", uint64(snapshot.Index)),
		logging.Time("Time", snapshot.Timestamp))
	s.index = snapshot.Index
	s.time = snapshot.Timestamp
	return s.sessions.recover(reader)
}

func (s *stateManager) Command(input *multiraftv1.CommandInput, stream streams.WriteStream[*multiraftv1.CommandOutput]) {
	s.queriesMu.RLock()
	canceledQueries := s.canceledQueries
	s.queriesMu.RUnlock()
	if canceledQueries != nil {
		s.queriesMu.Lock()
		for _, query := range s.canceledQueries {
			query.Close()
		}
		s.canceledQueries = nil
		s.queriesMu.Unlock()
	}

	s.index++
	if input.Timestamp.After(s.time) {
		s.time = input.Timestamp
	}
	s.scheduler.runScheduledTasks(s.time)

	switch i := input.Input.(type) {
	case *multiraftv1.CommandInput_SessionCommand:
		s.sessions.updateSession(i.SessionCommand, streams.NewEncodingStream[*multiraftv1.SessionCommandOutput, *multiraftv1.CommandOutput](stream, func(value *multiraftv1.SessionCommandOutput, err error) (*multiraftv1.CommandOutput, error) {
			if err != nil {
				return nil, err
			}
			return &multiraftv1.CommandOutput{
				Index: s.index,
				Output: &multiraftv1.CommandOutput_SessionCommand{
					SessionCommand: value,
				},
			}, nil
		}))
	case *multiraftv1.CommandInput_KeepAlive:
		s.sessions.keepAlive(i.KeepAlive, streams.NewEncodingStream[*multiraftv1.KeepAliveOutput, *multiraftv1.CommandOutput](stream, func(value *multiraftv1.KeepAliveOutput, err error) (*multiraftv1.CommandOutput, error) {
			if err != nil {
				return nil, err
			}
			return &multiraftv1.CommandOutput{
				Index: s.index,
				Output: &multiraftv1.CommandOutput_KeepAlive{
					KeepAlive: value,
				},
			}, nil
		}))
	case *multiraftv1.CommandInput_OpenSession:
		s.sessions.openSession(i.OpenSession, streams.NewEncodingStream[*multiraftv1.OpenSessionOutput, *multiraftv1.CommandOutput](stream, func(value *multiraftv1.OpenSessionOutput, err error) (*multiraftv1.CommandOutput, error) {
			if err != nil {
				return nil, err
			}
			return &multiraftv1.CommandOutput{
				Index: s.index,
				Output: &multiraftv1.CommandOutput_OpenSession{
					OpenSession: value,
				},
			}, nil
		}))
	case *multiraftv1.CommandInput_CloseSession:
		s.sessions.closeSession(i.CloseSession, streams.NewEncodingStream[*multiraftv1.CloseSessionOutput, *multiraftv1.CommandOutput](stream, func(value *multiraftv1.CloseSessionOutput, err error) (*multiraftv1.CommandOutput, error) {
			if err != nil {
				return nil, err
			}
			return &multiraftv1.CommandOutput{
				Index: s.index,
				Output: &multiraftv1.CommandOutput_CloseSession{
					CloseSession: value,
				},
			}, nil
		}))
	}

	s.scheduler.runImmediateTasks()

	s.queriesMu.RLock()
	queries, ok := s.pendingQueries[s.index]
	s.queriesMu.RUnlock()
	if ok {
		s.queriesMu.Lock()
		elem := queries.Front()
		for elem != nil {
			query := elem.Value.(pendingQuery)
			log.Debugf("Dequeued QueryInput at index %d: %.250s", s.index, query.input)
			s.Query(query.ctx, query.input, query.stream)
			elem = elem.Next()
		}
		delete(s.pendingQueries, s.index)
		s.queriesMu.Unlock()
	}
}

func (s *stateManager) Query(ctx context.Context, input *multiraftv1.QueryInput, stream streams.WriteStream[*multiraftv1.QueryOutput]) {
	if input.MaxReceivedIndex > s.index {
		log.Debugf("Enqueued QueryInput at index %d: %.250s", s.index, input)
		s.queriesMu.Lock()
		queries, ok := s.pendingQueries[input.MaxReceivedIndex]
		if !ok {
			queries = list.New()
			s.pendingQueries[input.MaxReceivedIndex] = queries
		}
		queries.PushBack(pendingQuery{
			ctx:    ctx,
			input:  input,
			stream: stream,
		})
		s.queriesMu.Unlock()
	} else {
		switch i := input.Input.(type) {
		case *multiraftv1.QueryInput_SessionQuery:
			s.sessions.readSession(ctx, i.SessionQuery, streams.NewEncodingStream[*multiraftv1.SessionQueryOutput, *multiraftv1.QueryOutput](stream, func(value *multiraftv1.SessionQueryOutput, err error) (*multiraftv1.QueryOutput, error) {
				if err != nil {
					return nil, err
				}
				return &multiraftv1.QueryOutput{
					Index: s.index,
					Output: &multiraftv1.QueryOutput_SessionQuery{
						SessionQuery: value,
					},
				}, nil
			}))
		}
	}
}

type pendingQuery struct {
	ctx    context.Context
	input  *multiraftv1.QueryInput
	stream streams.WriteStream[*multiraftv1.QueryOutput]
}
