// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import "github.com/atomix/runtime/sdk/pkg/logging"

type CallID interface {
	ProposalID | QueryID
}

// Call is a proposal or query call context
type Call[T CallID, I, O any] interface {
	// ID returns the execution identifier
	ID() T
	// Log returns the operation log
	Log() logging.Logger
	// Input returns the input
	Input() I
	// Output returns the output
	Output(O)
	// Error returns a failure error
	Error(error)
	// Close closes the execution
	Close()
}

func newTranscodingCall[T CallID, I1, O1, I2, O2 any](execution Call[T, I1, O1], input I2, transcoder func(O2) O1) Call[T, I2, O2] {
	return &transcodingCall[T, I1, O1, I2, O2]{
		parent:     execution,
		input:      input,
		transcoder: transcoder,
	}
}

type transcodingCall[T CallID, I1, O1, I2, O2 any] struct {
	parent     Call[T, I1, O1]
	input      I2
	transcoder func(O2) O1
}

func (e *transcodingCall[T, I1, O1, I2, O2]) ID() T {
	return e.parent.ID()
}

func (e *transcodingCall[T, I1, O1, I2, O2]) Log() logging.Logger {
	return e.parent.Log()
}

func (e *transcodingCall[T, I1, O1, I2, O2]) Input() I2 {
	return e.input
}

func (e *transcodingCall[T, I1, O1, I2, O2]) Output(output O2) {
	e.parent.Output(e.transcoder(output))
}

func (e *transcodingCall[T, I1, O1, I2, O2]) Error(err error) {
	e.parent.Error(err)
}

func (e *transcodingCall[T, I1, O1, I2, O2]) Close() {
	e.parent.Close()
}
