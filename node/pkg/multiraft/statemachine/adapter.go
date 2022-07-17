// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	streamregistry "github.com/atomix/multi-raft/node/pkg/multiraft/stream"
	"github.com/atomix/runtime/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"github.com/lni/dragonboat/v3/statemachine"
	"io"
)

func newManagerAdapter(manager *Manager, streams *streamregistry.Registry) statemachine.IStateMachine {
	return &stateMachineAdapter{
		manager: manager,
		streams: streams,
	}
}

type stateMachineAdapter struct {
	manager *Manager
	streams *streamregistry.Registry
}

func (s *stateMachineAdapter) Update(bytes []byte) (statemachine.Result, error) {
	logEntry := &multiraftv1.RaftLogEntry{}
	if err := proto.Unmarshal(bytes, logEntry); err != nil {
		return statemachine.Result{}, err
	}

	stream := s.streams.Lookup(logEntry.StreamID)
	s.manager.Command(&logEntry.Command, stream)
	return statemachine.Result{}, nil
}

func (s *stateMachineAdapter) Lookup(value interface{}) (interface{}, error) {
	query := value.(queryContext)
	s.manager.Query(query.input, query.stream)
	return nil, nil
}

func (s *stateMachineAdapter) SaveSnapshot(writer io.Writer, collection statemachine.ISnapshotFileCollection, i <-chan struct{}) error {
	return s.manager.Snapshot(writer)
}

func (s *stateMachineAdapter) RecoverFromSnapshot(reader io.Reader, files []statemachine.SnapshotFile, i <-chan struct{}) error {
	return s.manager.Restore(reader)
}

func (s *stateMachineAdapter) Close() error {
	return nil
}

type queryContext struct {
	input  *multiraftv1.QueryInput
	stream stream.WriteStream
}
