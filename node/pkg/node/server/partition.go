// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/logging"
	"github.com/atomix/runtime/pkg/stream"
	"github.com/gogo/protobuf/proto"
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

func (p *Partition) executeCommand(ctx context.Context, command *multiraftv1.CommandInput) (*multiraftv1.CommandOutput, error) {
	resultCh := make(chan stream.Result[*multiraftv1.CommandOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.commitCommand(ctx, command, stream.NewChannelStream[*multiraftv1.CommandOutput](resultCh)); err != nil {
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

func (p *Partition) executeStreamCommand(ctx context.Context, command *multiraftv1.CommandInput, ch chan<- multiraftv1.CommandOutput) error {
	resultCh := make(chan stream.Result[*multiraftv1.CommandOutput])
	errCh := make(chan error)

	stream := stream.NewBufferedStream[*multiraftv1.CommandOutput]()
	go func() {
		defer close(resultCh)
		for {
			result, ok := stream.Receive()
			if !ok {
				return
			}
			resultCh <- result
		}
	}()

	go func() {
		if err := p.commitCommand(ctx, command, stream); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				return nil
			}

			if result.Failed() {
				return result.Error
			}

			ch <- *result.Value
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Partition) commitCommand(ctx context.Context, input *multiraftv1.CommandInput, stream stream.WriteStream[*multiraftv1.CommandOutput]) error {
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

func (p *Partition) executeQuery(ctx context.Context, query *multiraftv1.QueryInput, sync bool) (*multiraftv1.QueryOutput, error) {
	resultCh := make(chan stream.Result[*multiraftv1.QueryOutput], 1)
	errCh := make(chan error, 1)
	go func() {
		if err := p.applyQuery(ctx, query, stream.NewChannelStream[*multiraftv1.QueryOutput](resultCh), sync); err != nil {
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

func (p *Partition) executeStreamQuery(ctx context.Context, query *multiraftv1.QueryInput, ch chan<- multiraftv1.QueryOutput, sync bool) error {
	resultCh := make(chan stream.Result[*multiraftv1.QueryOutput])
	errCh := make(chan error)

	stream := stream.NewBufferedStream[*multiraftv1.QueryOutput]()
	go func() {
		defer close(resultCh)
		for {
			result, ok := stream.Receive()
			if !ok {
				return
			}
			resultCh <- result
		}
	}()

	go func() {
		if err := p.applyQuery(ctx, query, stream, sync); err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case result, ok := <-resultCh:
			if !ok {
				return nil
			}

			if result.Failed() {
				return result.Error
			}

			ch <- *result.Value
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Partition) applyQuery(ctx context.Context, input *multiraftv1.QueryInput, stream stream.WriteStream[*multiraftv1.QueryOutput], sync bool) error {
	query := queryContext{
		input:  input,
		stream: stream,
	}
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
