// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/google/uuid"
	"time"
)

type QueryID = session.QueryID

type QueryState = session.QueryState

// Query is a read operation
type Query[I, O any] session.Query[I, O]

func newPrimitiveQuery[I, O any](session *primitiveSession[I, O]) *primitiveQuery[I, O] {
	return &primitiveQuery[I, O]{
		session: session,
	}
}

type primitiveQuery[I, O any] struct {
	parent   session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]
	session  *primitiveSession[I, O]
	input    I
	state    CallState
	watchers map[uuid.UUID]func(CallState)
	cancel   session.CancelFunc
	log      logging.Logger
}

func (q *primitiveQuery[I, O]) ID() QueryID {
	return q.parent.ID()
}

func (q *primitiveQuery[I, O]) Log() logging.Logger {
	return q.log
}

func (q *primitiveQuery[I, O]) Time() time.Time {
	return q.parent.Time()
}

func (q *primitiveQuery[I, O]) State() QueryState {
	return q.state
}

func (q *primitiveQuery[I, O]) Watch(watcher func(QueryState)) CancelFunc {
	if q.state != Running {
		watcher(q.state)
		return func() {}
	}
	if q.watchers == nil {
		q.watchers = make(map[uuid.UUID]func(QueryState))
	}
	id := uuid.New()
	q.watchers[id] = watcher
	return func() {
		delete(q.watchers, id)
	}
}

func (q *primitiveQuery[I, O]) Session() Session {
	return q.session
}

func (q *primitiveQuery[I, O]) init(parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) error {
	input, err := q.session.primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		return err
	}

	q.parent = parent
	q.input = input
	q.state = Running
	q.log = q.session.Log().WithFields(logging.Uint64("QueryID", uint64(parent.ID())))
	q.cancel = parent.Watch(func(state session.QueryState) {
		if q.state != Running {
			return
		}
		switch state {
		case session.Complete:
			q.destroy(Complete)
		case session.Canceled:
			q.destroy(Canceled)
		}
	})
	return nil
}

func (q *primitiveQuery[I, O]) execute(parent session.Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]) {
	if q.state != Pending {
		return
	}

	if err := q.init(parent); err != nil {
		q.Log().Errorw("Failed decoding proposal", logging.Error("Error", err))
		parent.Error(errors.NewInternal("failed decoding proposal: %s", err.Error()))
		parent.Close()
	} else {
		q.session.primitive.sm.Query(q)
		if q.state == Running {
			q.session.registerQuery(q)
		}
	}
}

func (q *primitiveQuery[I, O]) Input() I {
	return q.input
}

func (q *primitiveQuery[I, O]) Output(output O) {
	if q.state != Running {
		return
	}
	payload, err := q.session.primitive.codec.EncodeOutput(output)
	if err != nil {
		q.Log().Errorw("Failed encoding proposal", logging.Error("Error", err))
		q.parent.Error(errors.NewInternal("failed encoding proposal: %s", err.Error()))
		q.parent.Close()
	} else {
		q.parent.Output(&multiraftv1.PrimitiveQueryOutput{
			Payload: payload,
		})
	}
}

func (q *primitiveQuery[I, O]) Error(err error) {
	if q.state != Running {
		return
	}
	q.parent.Error(err)
	q.parent.Close()
}

func (q *primitiveQuery[I, O]) Cancel() {
	if q.state != Running {
		return
	}
	q.parent.Cancel()
}

func (q *primitiveQuery[I, O]) Close() {
	if q.state != Running {
		return
	}
	q.parent.Close()
}

func (q *primitiveQuery[I, O]) destroy(state QueryState) {
	q.cancel()
	q.state = state
	q.session.unregisterQuery(q.ID())
	if q.watchers != nil {
		for _, watcher := range q.watchers {
			watcher(state)
		}
	}
}

var _ Query[any, any] = (*primitiveQuery[any, any])(nil)
