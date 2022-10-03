// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"time"
)

type CallState int

const (
	Pending CallState = iota
	Running
	Complete
	Canceled
)

// Call is a proposal or query call
type Call[T statemachine.CallID, I, O any] interface {
	statemachine.Call[T, I, O]
	// Time returns the state machine time at the time of the call
	Time() time.Time
	// Session returns the call session
	Session() Session
	// State returns the call state
	State() CallState
	// Watch watches the call state for changes
	Watch(watcher func(CallState)) CancelFunc
	// Cancel cancels the call
	Cancel()
}
