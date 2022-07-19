// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/logging"
	streams "github.com/atomix/runtime/pkg/stream"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/metadata"
	"sync/atomic"
)

func newPartition(protocol *Protocol, id multiraftv1.PartitionID) *Partition {
	return &Partition{
		protocol: protocol,
		id:       id,
		streams:  newStreamRegistry(),
	}
}

type Partition struct {
	protocol *Protocol
	id       multiraftv1.PartitionID
	streams  *streamRegistry
	leader   uint64
	term     uint64
}

func (p *Partition) ID() multiraftv1.PartitionID {
	return p.id
}

func (p *Partition) setLeader(term multiraftv1.Term, leader multiraftv1.NodeID) {
	atomic.StoreUint64(&p.term, uint64(term))
	atomic.StoreUint64(&p.leader, uint64(leader))
}

func (p *Partition) getLeader() (multiraftv1.Term, multiraftv1.NodeID) {
	return multiraftv1.Term(atomic.LoadUint64(&p.term)), multiraftv1.NodeID(atomic.LoadUint64(&p.leader))
}

func (p *Partition) Command(ctx context.Context, command *multiraftv1.CommandInput) (*multiraftv1.CommandOutput, error) {
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

func (p *Partition) StreamCommand(ctx context.Context, input *multiraftv1.CommandInput, out streams.WriteStream[*multiraftv1.CommandOutput]) error {
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

func (p *Partition) commitCommand(ctx context.Context, input *multiraftv1.CommandInput, stream streams.WriteStream[*multiraftv1.CommandOutput]) error {
	term, leader := p.getLeader()
	if leader != p.protocol.id {
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
	if _, err := p.protocol.node.host.SyncPropose(ctx, p.protocol.node.host.GetNoOPSession(uint64(p.id)), bytes); err != nil {
		return wrapError(err)
	}
	return nil
}

func (p *Partition) Query(ctx context.Context, query *multiraftv1.QueryInput) (*multiraftv1.QueryOutput, error) {
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

func (p *Partition) StreamQuery(ctx context.Context, input *multiraftv1.QueryInput, out streams.WriteStream[*multiraftv1.QueryOutput]) error {
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

func (p *Partition) applyQuery(ctx context.Context, input *multiraftv1.QueryInput, stream streams.WriteStream[*multiraftv1.QueryOutput]) error {
	query := queryContext{
		input:  input,
		stream: stream,
	}
	md, _ := metadata.FromIncomingContext(ctx)
	sync := md["Sync"] != nil
	if sync {
		ctx, cancel := context.WithTimeout(ctx, clientTimeout)
		defer cancel()
		if _, err := p.protocol.node.host.SyncRead(ctx, uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	} else {
		if _, err := p.protocol.node.host.StaleRead(uint64(p.id), query); err != nil {
			return wrapError(err)
		}
	}
	return nil
}

func newPartitionServer(protocol *Protocol) multiraftv1.PartitionServer {
	return &PartitionServer{
		protocol: protocol,
	}
}

type PartitionServer struct {
	protocol *Protocol
}

func (s *PartitionServer) OpenSession(ctx context.Context, request *multiraftv1.OpenSessionRequest) (*multiraftv1.OpenSessionResponse, error) {
	log.Debugw("OpenSession",
		logging.Stringer("OpenSessionRequest", request))
	output, headers, err := s.protocol.OpenSession(ctx, &request.OpenSessionInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("OpenSession",
			logging.Stringer("OpenSessionRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.OpenSessionResponse{
		Headers:           *headers,
		OpenSessionOutput: *output,
	}
	log.Debugw("OpenSession",
		logging.Stringer("OpenSessionRequest", request),
		logging.Stringer("OpenSessionResponse", response))
	return response, nil
}

func (s *PartitionServer) KeepAlive(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
	log.Debugw("KeepAlive",
		logging.Stringer("KeepAliveRequest", request))
	output, headers, err := s.protocol.KeepAliveSession(ctx, &request.KeepAliveInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("KeepAlive",
			logging.Stringer("KeepAliveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.KeepAliveResponse{
		Headers:         *headers,
		KeepAliveOutput: *output,
	}
	log.Debugw("KeepAlive",
		logging.Stringer("KeepAliveRequest", request),
		logging.Stringer("KeepAliveResponse", response))
	return response, nil
}

func (s *PartitionServer) CloseSession(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
	log.Debugw("CloseSession",
		logging.Stringer("CloseSessionRequest", request))
	output, headers, err := s.protocol.CloseSession(ctx, &request.CloseSessionInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("CloseSession",
			logging.Stringer("CloseSessionRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.CloseSessionResponse{
		Headers:            *headers,
		CloseSessionOutput: *output,
	}
	log.Debugw("CloseSession",
		logging.Stringer("CloseSessionRequest", request),
		logging.Stringer("CloseSessionResponse", response))
	return response, nil
}

func (s *PartitionServer) Watch(request *multiraftv1.WatchPartitionRequest, server multiraftv1.Partition_WatchServer) error {
	//TODO implement me
	panic("implement me")
}
