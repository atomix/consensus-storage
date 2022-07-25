// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"container/list"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/runtime/sdk/pkg/logging"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"sync"
	"time"
)

var log = logging.GetLogger()

func newStateManager(registry *PrimitiveTypeRegistry) *stateManager {
	stateMachine := &stateManager{
		queries:   make(map[multiraftv1.Index]*list.List),
		scheduler: newScheduler(),
	}
	stateMachine.sessions = newSessionManager(registry, stateMachine)
	return stateMachine
}

type stateManager struct {
	sessions  *sessionManager
	queries   map[multiraftv1.Index]*list.List
	queriesMu sync.RWMutex
	scheduler *Scheduler
	index     multiraftv1.Index
	time      time.Time
}

func (s *stateManager) Snapshot(writer *snapshot.Writer) error {
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
	s.index = snapshot.Index
	s.time = snapshot.Timestamp
	return s.sessions.recover(reader)
}

func (s *stateManager) Command(input *multiraftv1.CommandInput, stream streams.WriteStream[*multiraftv1.CommandOutput]) {
	s.index++
	if input.Timestamp.After(s.time) {
		s.time = input.Timestamp
	}
	s.scheduler.runScheduledTasks(s.time)

	switch i := input.Input.(type) {
	case *multiraftv1.CommandInput_SessionCommand:
		s.sessions.commandSession(i.SessionCommand, streams.NewEncodingStream[*multiraftv1.SessionCommandOutput, *multiraftv1.CommandOutput](stream, func(value *multiraftv1.SessionCommandOutput, err error) (*multiraftv1.CommandOutput, error) {
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
	queries, ok := s.queries[s.index]
	s.queriesMu.RUnlock()
	if ok {
		s.queriesMu.Lock()
		elem := queries.Front()
		for elem != nil {
			query := elem.Value.(pendingQuery)
			log.Debugf("Dequeued QueryInput at index %d: %.250s", s.index, query.input)
			s.Query(query.input, query.stream)
			elem = elem.Next()
		}
		delete(s.queries, s.index)
		s.queriesMu.Unlock()
	}
}

func (s *stateManager) Query(input *multiraftv1.QueryInput, stream streams.WriteStream[*multiraftv1.QueryOutput]) {
	if input.MaxReceivedIndex > s.index {
		log.Debugf("Enqueued QueryInput at index %d: %.250s", s.index, input)
		s.queriesMu.Lock()
		queries, ok := s.queries[input.MaxReceivedIndex]
		if !ok {
			queries = list.New()
			s.queries[input.MaxReceivedIndex] = queries
		}
		queries.PushBack(pendingQuery{
			input:  input,
			stream: stream,
		})
		s.queriesMu.Unlock()
	} else {
		switch i := input.Input.(type) {
		case *multiraftv1.QueryInput_SessionQuery:
			s.sessions.querySession(i.SessionQuery, streams.NewEncodingStream[*multiraftv1.SessionQueryOutput, *multiraftv1.QueryOutput](stream, func(value *multiraftv1.SessionQueryOutput, err error) (*multiraftv1.QueryOutput, error) {
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
	input  *multiraftv1.QueryInput
	stream streams.WriteStream[*multiraftv1.QueryOutput]
}
