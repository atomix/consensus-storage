// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/stream"
	"io"
)

func NewManager() *Manager {
	return &Manager{
		...
	}
}

type Manager struct {
}

func (m *Manager) Command(command *multiraftv1.CommandInput, stream stream.WriteStream) {

}

func (m *Manager) Query(query *multiraftv1.QueryInput, stream stream.WriteStream) {

}

func (m *Manager) Snapshot(writer io.Writer) error {

}

func (m *Manager) Restore(reader io.Reader) error {

}
