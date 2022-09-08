// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"google.golang.org/grpc"
)

func newPrimitiveClient(session *SessionClient, spec multiraftv1.PrimitiveSpec) *PrimitiveClient {
	return &PrimitiveClient{
		session: session,
		spec:    spec,
	}
}

type PrimitiveClient struct {
	session *SessionClient
	id      multiraftv1.PrimitiveID
	spec    multiraftv1.PrimitiveSpec
}

func (p *PrimitiveClient) open(ctx context.Context) error {
	command := Command[*multiraftv1.CreatePrimitiveResponse](p)
	response, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*multiraftv1.CreatePrimitiveResponse, error) {
		return multiraftv1.NewSessionClient(conn).CreatePrimitive(ctx, &multiraftv1.CreatePrimitiveRequest{
			Headers: headers,
			CreatePrimitiveInput: multiraftv1.CreatePrimitiveInput{
				PrimitiveSpec: p.spec,
			},
		})
	})
	if err != nil {
		return err
	}
	p.id = response.PrimitiveID
	return nil
}

func (p *PrimitiveClient) close(ctx context.Context) error {
	command := Command[*multiraftv1.ClosePrimitiveResponse](p)
	_, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*multiraftv1.ClosePrimitiveResponse, error) {
		return multiraftv1.NewSessionClient(conn).ClosePrimitive(ctx, &multiraftv1.ClosePrimitiveRequest{
			Headers: headers,
			ClosePrimitiveInput: multiraftv1.ClosePrimitiveInput{
				PrimitiveID: p.id,
			},
		})
	})
	return err
}

type CommandResponse interface {
	GetHeaders() *multiraftv1.CommandResponseHeaders
}

type QueryResponse interface {
	GetHeaders() *multiraftv1.QueryResponseHeaders
}

func Command[T CommandResponse](primitive *PrimitiveClient) *CommandContext[T] {
	headers := &multiraftv1.CommandRequestHeaders{
		OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
			PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
				SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
					PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
						PartitionID: primitive.session.partition.id,
					},
					SessionID: primitive.session.sessionID,
				},
				PrimitiveID: primitive.id,
			},
		},
		SequenceNum: primitive.session.nextRequestNum(),
	}
	return &CommandContext[T]{
		session: primitive.session,
		headers: headers,
	}
}

type CommandContext[T CommandResponse] struct {
	session *SessionClient
	headers *multiraftv1.CommandRequestHeaders
}

func (c *CommandContext[T]) Run(f func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (T, error)) (T, error) {
	c.session.recorder.Start(c.headers.SequenceNum)
	defer c.session.recorder.End(c.headers.SequenceNum)
	response, err := f(c.session.conn, c.headers)
	if err != nil {
		return response, err
	}
	headers := response.GetHeaders()
	c.session.lastIndex.Update(headers.Index)
	if headers.Status != multiraftv1.OperationResponseHeaders_OK {
		return response, getErrorFromStatus(headers.Status, headers.Message)
	}
	return response, nil
}

type CommandStream[T CommandResponse] interface {
	Recv() (T, error)
}

func StreamCommand[T CommandResponse](primitive *PrimitiveClient) *StreamCommandContext[T] {
	headers := &multiraftv1.CommandRequestHeaders{
		OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
			PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
				SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
					PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
						PartitionID: primitive.session.partition.id,
					},
					SessionID: primitive.session.sessionID,
				},
				PrimitiveID: primitive.id,
			},
		},
		SequenceNum: primitive.session.nextRequestNum(),
	}
	return &StreamCommandContext[T]{
		session: primitive.session,
		headers: headers,
	}
}

type StreamCommandContext[T CommandResponse] struct {
	session *SessionClient
	headers *multiraftv1.CommandRequestHeaders
}

func (c *StreamCommandContext[T]) Run(f func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (CommandStream[T], error)) (CommandStream[T], error) {
	c.session.recorder.Start(c.headers.SequenceNum)
	stream, err := f(c.session.conn, c.headers)
	if err != nil {
		c.session.recorder.End(c.headers.SequenceNum)
		return stream, err
	}
	c.session.recorder.StreamOpen(c.headers)
	return &CommandStreamContext[T]{
		StreamCommandContext: c,
		stream:               stream,
	}, nil
}

type CommandStreamContext[T CommandResponse] struct {
	*StreamCommandContext[T]
	stream                  CommandStream[T]
	lastResponseSequenceNum multiraftv1.SequenceNum
}

func (s *CommandStreamContext[T]) Recv() (T, error) {
	for {
		response, err := s.stream.Recv()
		if err != nil {
			s.session.recorder.StreamClose(s.headers)
			s.session.recorder.End(s.headers.SequenceNum)
			return response, err
		}
		headers := response.GetHeaders()
		s.session.lastIndex.Update(headers.Index)
		if headers.OutputSequenceNum == s.lastResponseSequenceNum+1 {
			s.lastResponseSequenceNum++
			s.session.recorder.StreamReceive(s.headers, headers)
			if headers.Status != multiraftv1.OperationResponseHeaders_OK {
				return response, getErrorFromStatus(headers.Status, headers.Message)
			}
			return response, nil
		}
	}
}

func Query[T QueryResponse](primitive *PrimitiveClient) *QueryContext[T] {
	headers := &multiraftv1.QueryRequestHeaders{
		OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
			PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
				SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
					PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
						PartitionID: primitive.session.partition.id,
					},
					SessionID: primitive.session.sessionID,
				},
				PrimitiveID: primitive.id,
			},
		},
		SequenceNum:      primitive.session.nextRequestNum(),
		MaxReceivedIndex: primitive.session.lastIndex.Get(),
	}
	return &QueryContext[T]{
		session: primitive.session,
		headers: headers,
	}
}

type QueryContext[T QueryResponse] struct {
	session *SessionClient
	headers *multiraftv1.QueryRequestHeaders
}

func (c *QueryContext[T]) Run(f func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (T, error)) (T, error) {
	c.session.recorder.Start(c.headers.SequenceNum)
	defer c.session.recorder.End(c.headers.SequenceNum)
	response, err := f(c.session.conn, c.headers)
	if err != nil {
		return response, err
	}
	headers := response.GetHeaders()
	c.session.lastIndex.Update(headers.Index)
	if headers.Status != multiraftv1.OperationResponseHeaders_OK {
		return response, getErrorFromStatus(headers.Status, headers.Message)
	}
	return response, nil
}

type QueryStream[T QueryResponse] interface {
	Recv() (T, error)
}

func StreamQuery[T QueryResponse](primitive *PrimitiveClient) *StreamQueryContext[T] {
	headers := &multiraftv1.QueryRequestHeaders{
		OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
			PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
				SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
					PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
						PartitionID: primitive.session.partition.id,
					},
					SessionID: primitive.session.sessionID,
				},
				PrimitiveID: primitive.id,
			},
		},
		SequenceNum:      primitive.session.nextRequestNum(),
		MaxReceivedIndex: primitive.session.lastIndex.Get(),
	}
	return &StreamQueryContext[T]{
		session: primitive.session,
		headers: headers,
	}
}

type StreamQueryContext[T QueryResponse] struct {
	session *SessionClient
	headers *multiraftv1.QueryRequestHeaders
}

func (c *StreamQueryContext[T]) Run(f func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (QueryStream[T], error)) (QueryStream[T], error) {
	c.session.recorder.Start(c.headers.SequenceNum)
	stream, err := f(c.session.conn, c.headers)
	if err != nil {
		c.session.recorder.End(c.headers.SequenceNum)
		return nil, err
	}
	return &QueryStreamContext[T]{
		StreamQueryContext: c,
		stream:             stream,
	}, nil
}

type QueryStreamContext[T QueryResponse] struct {
	*StreamQueryContext[T]
	stream QueryStream[T]
}

func (c *QueryStreamContext[T]) Recv() (T, error) {
	for {
		response, err := c.stream.Recv()
		if err != nil {
			c.session.recorder.End(c.headers.SequenceNum)
			return response, err
		}
		headers := response.GetHeaders()
		c.session.lastIndex.Update(headers.Index)
		if headers.Status != multiraftv1.OperationResponseHeaders_OK {
			return response, getErrorFromStatus(headers.Status, headers.Message)
		}
		return response, nil
	}
}
