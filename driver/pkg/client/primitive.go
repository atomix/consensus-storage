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

func StreamCommand[T any, U CommandResponse](primitive *PrimitiveClient) *StreamCommandContext[T, U] {
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
	return &StreamCommandContext[T, U]{
		session: primitive.session,
		headers: headers,
	}
}

type StreamCommandContext[T any, U CommandResponse] struct {
	session                 *SessionClient
	headers                 *multiraftv1.CommandRequestHeaders
	lastResponseSequenceNum multiraftv1.SequenceNum
}

func (c *StreamCommandContext[T, U]) Open(f func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (T, error)) (T, error) {
	c.session.recorder.Start(c.headers.SequenceNum)
	client, err := f(c.session.conn, c.headers)
	if err != nil {
		c.session.recorder.End(c.headers.SequenceNum)
		return client, err
	}
	c.session.recorder.StreamOpen(c.headers)
	return client, nil
}

func (c *StreamCommandContext[T, U]) Recv(f func() (U, error)) (U, error) {
	for {
		response, err := f()
		if err != nil {
			c.session.recorder.StreamClose(c.headers)
			c.session.recorder.End(c.headers.SequenceNum)
			return response, err
		}
		headers := response.GetHeaders()
		c.session.lastIndex.Update(headers.Index)
		if headers.OutputSequenceNum == c.lastResponseSequenceNum+1 {
			c.lastResponseSequenceNum++
			c.session.recorder.StreamReceive(c.headers, headers)
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

func StreamQuery[T any, U QueryResponse](primitive *PrimitiveClient) *StreamQueryContext[T, U] {
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
	return &StreamQueryContext[T, U]{
		session: primitive.session,
		headers: headers,
	}
}

type StreamQueryContext[T any, U QueryResponse] struct {
	session *SessionClient
	headers *multiraftv1.QueryRequestHeaders
}

func (c *StreamQueryContext[T, U]) Open(f func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (T, error)) (T, error) {
	c.session.recorder.Start(c.headers.SequenceNum)
	client, err := f(c.session.conn, c.headers)
	if err != nil {
		c.session.recorder.End(c.headers.SequenceNum)
		return client, err
	}
	return client, nil
}

func (c *StreamQueryContext[T, U]) Recv(f func() (U, error)) (U, error) {
	for {
		response, err := f()
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
