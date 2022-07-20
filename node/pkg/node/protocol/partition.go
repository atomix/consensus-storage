// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
	streams "github.com/atomix/runtime/sdk/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/metadata"
	"sync/atomic"
)

func newPartitionProtocol(node *NodeProtocol, id multiraftv1.PartitionID) *PartitionProtocol {
	return &PartitionProtocol{
		node:    node,
		id:      id,
		streams: newStreamRegistry(),
	}
}

type PartitionProtocol struct {
	node    *NodeProtocol
	id      multiraftv1.PartitionID
	streams *streamRegistry
	leader  uint64
	term    uint64
}

func (p *PartitionProtocol) ID() multiraftv1.PartitionID {
	return p.id
}

func (p *PartitionProtocol) setLeader(term multiraftv1.Term, leader multiraftv1.NodeID) {
	atomic.StoreUint64(&p.term, uint64(term))
	atomic.StoreUint64(&p.leader, uint64(leader))
}

func (p *PartitionProtocol) getLeader() (multiraftv1.Term, multiraftv1.NodeID) {
	return multiraftv1.Term(atomic.LoadUint64(&p.term)), multiraftv1.NodeID(atomic.LoadUint64(&p.leader))
}

func (p *PartitionProtocol) Command(ctx context.Context, command *multiraftv1.CommandInput) (*multiraftv1.CommandOutput, error) {
	resultCh := make(chan streams.Result[*multiraftv1.CommandOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.commitCommand(ctx, command, streams.NewChannelStream[*multiraftv1.CommandOutput](resultCh)); err != nil {
			errCh <- err
		}
	}()

	select {
	case result, ok := <-resultCh:
		if !ok {
			err := errors.NewCanceled("stream closed")
			return nil, err
		}

		if result.Failed() {
			return nil, result.Error
		}

		return result.Value, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *PartitionProtocol) StreamCommand(ctx context.Context, input *multiraftv1.CommandInput, out streams.WriteStream[*multiraftv1.CommandOutput]) error {
	in := streams.NewBufferedStream[*multiraftv1.CommandOutput]()
	go func() {
		if err := p.commitCommand(ctx, input, in); err != nil {
			in.Error(err)
			return
		}
	}()
	go func() {
		for {
			result, ok := in.Receive()
			if !ok {
				out.Close()
				return
			}
			out.Send(result)
		}
	}()
	return nil
}

func (p *PartitionProtocol) commitCommand(ctx context.Context, input *multiraftv1.CommandInput, stream streams.WriteStream[*multiraftv1.CommandOutput]) error {
	term, leader := p.getLeader()
	if leader != p.node.id {
		return errors.NewUnavailable("not the leader")
	}

	streamID := p.streams.register(term, stream)
	defer p.streams.unregister(streamID)

	entry := &multiraftv1.RaftLogEntry{
		StreamID: streamID,
		Command:  *input,
	}

	bytes, err := proto.Marshal(entry)
	if err != nil {
		return errors.NewInternal("failed to marshal RaftLogEntry: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, clientTimeout)
	defer cancel()
	if _, err := p.node.node.SyncPropose(ctx, p.node.node.GetNoOPSession(uint64(p.id)), bytes); err != nil {
		return wrapError(err)
	}
	return nil
}

func (p *PartitionProtocol) Query(ctx context.Context, query *multiraftv1.QueryInput) (*multiraftv1.QueryOutput, error) {
	resultCh := make(chan streams.Result[*multiraftv1.QueryOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.applyQuery(ctx, query, streams.NewChannelStream[*multiraftv1.QueryOutput](resultCh)); err != nil {
			errCh <- err
		}
	}()

	select {
	case result, ok := <-resultCh:
		if !ok {
			err := errors.NewCanceled("stream closed")
			return nil, err
		}

		if result.Failed() {
			return nil, result.Error
		}

		return result.Value, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *PartitionProtocol) StreamQuery(ctx context.Context, input *multiraftv1.QueryInput, out streams.WriteStream[*multiraftv1.QueryOutput]) error {
	in := streams.NewBufferedStream[*multiraftv1.QueryOutput]()
	go func() {
		if err := p.applyQuery(ctx, input, in); err != nil {
			in.Error(err)
			return
		}
	}()
	go func() {
		for {
			result, ok := in.Receive()
			if !ok {
				out.Close()
				return
			}
			out.Send(result)
		}
	}()
	return nil
}

func (p *PartitionProtocol) applyQuery(ctx context.Context, input *multiraftv1.QueryInput, stream streams.WriteStream[*multiraftv1.QueryOutput]) error {
	query := queryContext{
		input:  input,
		stream: stream,
	}
	md, _ := metadata.FromIncomingContext(ctx)
	sync := md["Sync"] != nil
	if sync {
		ctx, cancel := context.WithTimeout(ctx, clientTimeout)
		defer cancel()
		if _, err := p.node.node.SyncRead(ctx, uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	} else {
		if _, err := p.node.node.StaleRead(uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	}
	return nil
}
