// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"github.com/atomix/runtime/pkg/stream"
	"io"
)

// StateMachine is an interface for primitive state machines
type StateMachine interface {
	// Snapshot snapshots the state machine state to the given writer
	Snapshot(writer io.Writer) error

	// Restore restores the state machine state from the given reader
	Restore(reader io.Reader) error

	// Command applies a command to the state machine
	Command(bytes []byte, stream stream.WriteStream)

	// Query applies a query to the state machine
	Query(bytes []byte, stream stream.WriteStream)
}
