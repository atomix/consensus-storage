// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/snapshot"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
	dbstatemachine "github.com/lni/dragonboat/v3/statemachine"
	"io"
)

func NewStateMachine(partition *PartitionProtocol) dbstatemachine.IStateMachine {
	return &stateMachine{
		state:   statemachine.NewStateMachine(partition.node.registry),
		streams: partition.streams,
	}
}

type stateMachine struct {
	state   statemachine.StateMachine
	streams *streamRegistry
}

func (s *stateMachine) Update(bytes []byte) (dbstatemachine.Result, error) {
	logEntry := &multiraftv1.RaftLogEntry{}
	if err := proto.Unmarshal(bytes, logEntry); err != nil {
		return dbstatemachine.Result{}, err
	}

	stream := s.streams.lookup(logEntry.StreamID)
	s.state.Command(&logEntry.Command, stream)
	return dbstatemachine.Result{}, nil
}

func (s *stateMachine) Lookup(value interface{}) (interface{}, error) {
	query := value.(queryContext)
	s.state.Query(query.input, query.stream)
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

type queryContext struct {
	input  *multiraftv1.QueryInput
	stream streams.WriteStream[*multiraftv1.QueryOutput]
}
