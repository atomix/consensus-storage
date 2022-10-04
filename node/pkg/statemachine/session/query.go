// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/google/uuid"
	"sync"
	"sync/atomic"
	"time"
)

type QueryID = statemachine.QueryID

type QueryState = CallState

// Query is a read operation
type Query[I, O any] Call[QueryID, I, O]

func newSessionQuery(session *managedSession) *sessionQuery {
	return &sessionQuery{
		session: session,
	}
}

type sessionQuery struct {
	session   *managedSession
	parent    statemachine.Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]
	timestamp time.Time
	state     QueryState
	watching  atomic.Bool
	watchers  map[uuid.UUID]func(QueryState)
	mu        sync.RWMutex
	log       logging.Logger
}

func (q *sessionQuery) ID() statemachine.QueryID {
	return q.parent.ID()
}

func (q *sessionQuery) Log() logging.Logger {
	return q.log
}

func (q *sessionQuery) Session() Session {
	return q.session
}

func (q *sessionQuery) Time() time.Time {
	return q.timestamp
}

func (q *sessionQuery) State() QueryState {
	return q.state
}

func (q *sessionQuery) Watch(watcher func(QueryState)) CancelFunc {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.watchers == nil {
		q.watchers = make(map[uuid.UUID]func(QueryState))
	}
	id := uuid.New()
	q.watchers[id] = watcher
	q.watching.Store(true)
	return func() {
		q.mu.Lock()
		defer q.mu.Unlock()
		delete(q.watchers, id)
	}
}

func (q *sessionQuery) execute(parent statemachine.Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]) {
	q.state = Running
	q.parent = parent
	q.timestamp = q.session.manager.Time()
	q.log = q.session.Log().WithFields(logging.Uint64("QueryID", uint64(parent.ID())))
	switch parent.Input().Input.(type) {
	case *multiraftv1.SessionQueryInput_Query:
		query := newPrimitiveQuery(q)
		q.session.manager.sm.Query(query)
	}
}

func (q *sessionQuery) Input() *multiraftv1.SessionQueryInput {
	return q.parent.Input()
}

func (q *sessionQuery) Output(output *multiraftv1.SessionQueryOutput) {
	if q.state != Running {
		return
	}
	q.parent.Output(output)
}

func (q *sessionQuery) Error(err error) {
	if q.state != Running {
		return
	}
	q.parent.Error(err)
	q.Close()
}

func (q *sessionQuery) Cancel() {
	q.close(Canceled)
}

func (q *sessionQuery) Close() {
	q.close(Complete)
}

func (q *sessionQuery) close(phase QueryState) {
	if q.state != Running {
		return
	}
	q.state = phase
	q.parent.Close()
	if q.watching.Load() {
		q.mu.RLock()
		watchers := make([]func(QueryState), 0, len(q.watchers))
		for _, watcher := range q.watchers {
			watchers = append(watchers, watcher)
		}
		q.mu.RUnlock()
		for _, watcher := range watchers {
			watcher(phase)
		}
	}
}

var _ Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput] = (*sessionQuery)(nil)

func newPrimitiveQuery(parent *sessionQuery) *primitiveQuery {
	return &primitiveQuery{
		sessionQuery: parent,
	}
}

type primitiveQuery struct {
	*sessionQuery
}

func (p *primitiveQuery) Input() *multiraftv1.PrimitiveQueryInput {
	return p.sessionQuery.Input().GetQuery()
}

func (p *primitiveQuery) Output(output *multiraftv1.PrimitiveQueryOutput) {
	p.sessionQuery.Output(&multiraftv1.SessionQueryOutput{
		Output: &multiraftv1.SessionQueryOutput_Query{
			Query: output,
		},
	})
}

func (p *primitiveQuery) Error(err error) {
	p.sessionQuery.Output(&multiraftv1.SessionQueryOutput{
		Failure: getFailure(err),
	})
}

var _ Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput] = (*primitiveQuery)(nil)
