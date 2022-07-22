// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"google.golang.org/grpc"
)

func newPrimitiveClient(session *SessionClient, id multiraftv1.PrimitiveID) *PrimitiveClient {
	return &PrimitiveClient{
		session: session,
		id:      id,
	}
}

type PrimitiveClient struct {
	session *SessionClient
	id      multiraftv1.PrimitiveID
}

func (p *PrimitiveClient) close(ctx context.Context) error {
	request := &multiraftv1.ClosePrimitiveRequest{
		Headers: multiraftv1.CommandRequestHeaders{
			OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
				PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
					SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
						PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
							PartitionID: p.session.partition.id,
						},
						SessionID: p.session.sessionID,
					},
					PrimitiveID: p.id,
				},
			},
			SequenceNum: p.session.nextRequestNum(),
		},
		ClosePrimitiveInput: multiraftv1.ClosePrimitiveInput{
			PrimitiveID: p.id,
		},
	}
	client := multiraftv1.NewSessionClient(p.session.partition.conn)
	_, err := client.ClosePrimitive(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

type CommandResponse interface {
	GetHeaders() multiraftv1.CommandResponseHeaders
}

type QueryResponse interface {
	GetHeaders() multiraftv1.QueryResponseHeaders
}

func Command[T CommandResponse](primitive *PrimitiveClient, operationID multiraftv1.OperationID) *CommandContext[T] {
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
			OperationID: operationID,
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
	c.session.recorder.Start(c.headers)
	defer c.session.recorder.End(c.headers)
	response, err := f(c.session.conn, c.headers)
	if err != nil {
		return response, err
	}
	c.session.lastIndex.Update(response.GetHeaders().Index)
	return response, nil
}

func StreamCommand[T any, U CommandResponse](primitive *PrimitiveClient, operationID multiraftv1.OperationID) *StreamCommandContext[T, U] {
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
			OperationID: operationID,
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
	c.session.recorder.Start(c.headers)
	client, err := f(c.session.conn, c.headers)
	if err != nil {
		c.session.recorder.End(c.headers)
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
			c.session.recorder.End(c.headers)
			return response, err
		}
		headers := response.GetHeaders()
		c.session.lastIndex.Update(headers.Index)
		if headers.OutputSequenceNum == c.lastResponseSequenceNum+1 {
			c.session.recorder.StreamReceive(&headers)
			return response, nil
		}
	}
}

func Query[T QueryResponse](primitive *PrimitiveClient, operationID multiraftv1.OperationID) *QueryContext[T] {
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
			OperationID: operationID,
		},
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
	response, err := f(c.session.conn, c.headers)
	if err != nil {
		return response, err
	}
	c.session.lastIndex.Update(response.GetHeaders().Index)
	return response, nil
}

func StreamQuery[T any, U QueryResponse](primitive *PrimitiveClient, operationID multiraftv1.OperationID) *StreamQueryContext[T, U] {
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
			OperationID: operationID,
		},
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
	return f(c.session.conn, c.headers)
}

func (c *StreamQueryContext[T, U]) Recv(f func() (U, error)) (U, error) {
	response, err := f()
	if err != nil {
		return response, err
	}
	c.session.lastIndex.Update(response.GetHeaders().Index)
	return response, nil
}
