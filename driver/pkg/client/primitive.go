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
		Headers: &multiraftv1.CommandRequestHeaders{
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
	response, err := client.ClosePrimitive(ctx, request)
	if err != nil {
		return err
	}
	if response.Headers.Status != multiraftv1.OperationResponseHeaders_OK {
		return getErrorFromStatus(response.Headers.Status, response.Headers.Message)
	}
	return nil
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
			return nil, err
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
			return nil, err
		}
		headers := response.GetHeaders()
		c.session.lastIndex.Update(headers.Index)
		if headers.Status != multiraftv1.OperationResponseHeaders_OK {
			return response, getErrorFromStatus(headers.Status, headers.Message)
		}
		return response, nil
	}
}
