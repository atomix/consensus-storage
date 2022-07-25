// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/stream"
	"github.com/gogo/protobuf/proto"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"io"
)

func NewStateMachine(streams *stream.Registry, primitives *PrimitiveTypeRegistry) dbstatemachine.IStateMachine {
	return &stateMachine{
		state:   newStateManager(primitives),
		streams: streams,
	}
}

type stateMachine struct {
	state   *stateManager
	streams *stream.Registry
}

func (s *stateMachine) Update(bytes []byte) (dbstatemachine.Result, error) {
	logEntry := &multiraftv1.RaftLogEntry{}
	if err := proto.Unmarshal(bytes, logEntry); err != nil {
		return dbstatemachine.Result{}, err
	}

	stream := s.streams.Lookup(logEntry.StreamID)
	s.state.Command(&logEntry.Command, stream)
	return dbstatemachine.Result{}, nil
}

func (s *stateMachine) Lookup(value interface{}) (interface{}, error) {
	query := value.(*stream.Query)
	s.state.Query(query.Input, query.Stream)
	return nil, nil
}

func (s *stateMachine) SaveSnapshot(writer io.Writer, collection dbstatemachine.ISnapshotFileCollection, i <-chan struct{}) error {
	return s.state.Snapshot(snapshot.NewWriter(writer))
}

func (s *stateMachine) RecoverFromSnapshot(reader io.Reader, files []dbstatemachine.SnapshotFile, i <-chan struct{}) error {
	return s.state.Recover(snapshot.NewReader(reader))
}

func (s *stateMachine) Close() error {
	return nil
}
