// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

type QueryID uint64

// Query is a read operation
type Query[I, O any] interface {
	Call[QueryID, I, O]
}

func newTranscodingQuery[I1, O1, I2, O2 any](proposal Query[I1, O1], input I2, transcoder func(O2) O1) Query[I2, O2] {
	return &transcodingQuery[I1, O1, I2, O2]{
		Call: newTranscodingCall[QueryID, I1, O1, I2, O2](proposal, input, transcoder),
	}
}

type transcodingQuery[I1, O1, I2, O2 any] struct {
	Call[QueryID, I2, O2]
}
