// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"google.golang.org/grpc"
)

func newPrimitiveClient(id multiraftv1.PrimitiveID, session *SessionClient) *PrimitiveClient {
	return &PrimitiveClient{
		id:      id,
		session: session,
	}
}

type PrimitiveClient struct {
	id      multiraftv1.PrimitiveID
	session *SessionClient
}

func (p *PrimitiveClient) Conn() *grpc.ClientConn {
	return p.session.partition.conn
}

func (p *PrimitiveClient) Command(id multiraftv1.OperationID) *CommandContext {
	return newCommandContext(p, id)
}

func (p *PrimitiveClient) Query(id multiraftv1.OperationID) *QueryContext {
	return newQueryContext(p, id)
}

func newCommandContext(primitive *PrimitiveClient, operationID multiraftv1.OperationID) *CommandContext {
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
	primitive.session.recorder.Start(headers)
	return &CommandContext{
		session: primitive.session,
		headers: headers,
	}
}

type CommandContext struct {
	session *SessionClient
	headers *multiraftv1.CommandRequestHeaders
}

func (c *CommandContext) Headers() *multiraftv1.CommandRequestHeaders {
	return c.headers
}

func (c *CommandContext) Stream() *StreamCommandContext {
	return &StreamCommandContext{
		CommandContext: c,
	}
}

func (c *CommandContext) Receive(headers *multiraftv1.CommandResponseHeaders) bool {
	c.session.update(headers.Index)
	return true
}

func (c *CommandContext) Done() {
	c.session.recorder.End(c.headers)
}

type ServerStream[T any] interface {
	Recv() (T, error)
}

type StreamCommandContext struct {
	*CommandContext
	lastResponseSequenceNum multiraftv1.SequenceNum
}

func (c *StreamCommandContext) Open() {
	c.session.recorder.StreamOpen(c.headers)
}

func (c *StreamCommandContext) Receive(headers *multiraftv1.CommandResponseHeaders) bool {
	c.session.update(headers.Index)
	if headers.OutputSequenceNum <= c.lastResponseSequenceNum {
		return false
	}
	c.session.recorder.StreamReceive(headers)
	return true
}

func (c *StreamCommandContext) Close() {
	c.session.recorder.StreamClose(c.headers)
}

func newQueryContext(primitive *PrimitiveClient, operationID multiraftv1.OperationID) *QueryContext {
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
	return &QueryContext{
		headers: headers,
	}
}

type QueryContext struct {
	session *SessionClient
	headers *multiraftv1.QueryRequestHeaders
}

func (c *QueryContext) Headers() *multiraftv1.QueryRequestHeaders {
	return c.headers
}
