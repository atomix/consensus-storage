// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"google.golang.org/grpc"
	"sync/atomic"
)

func newPrimitiveClient(spec multiraftv1.PrimitiveSpec, session *SessionClient) *PrimitiveClient {
	return &PrimitiveClient{
		spec:    spec,
		session: session,
	}
}

type PrimitiveClient struct {
	id      uint64
	spec    multiraftv1.PrimitiveSpec
	session *SessionClient
}

func (p *PrimitiveClient) Conn() *grpc.ClientConn {
	return p.session.partition.conn
}

func (p *PrimitiveClient) create(ctx context.Context) error {
	request := &multiraftv1.CreatePrimitiveRequest{
		Headers: multiraftv1.CommandRequestHeaders{
			OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
				PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
					SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
						PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
							PartitionID: p.session.partition.id,
						},
						SessionID: p.session.sessionID,
					},
				},
			},
			SequenceNum: p.session.nextRequestNum(),
		},
		CreatePrimitiveInput: multiraftv1.CreatePrimitiveInput{
			PrimitiveSpec: p.spec,
		},
	}
	client := multiraftv1.NewSessionClient(p.session.partition.conn)
	response, err := client.CreatePrimitive(ctx, request)
	if err != nil {
		return err
	}
	atomic.StoreUint64(&p.id, uint64(response.PrimitiveID))
	return nil
}

func (p *PrimitiveClient) Command(id multiraftv1.OperationID) *CommandContext {
	return newCommandContext(p, id)
}

func (p *PrimitiveClient) Query(id multiraftv1.OperationID) *QueryContext {
	return newQueryContext(p, id)
}

func (p *PrimitiveClient) close(ctx context.Context) error {
	primitiveID := atomic.LoadUint64(&p.id)
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
					PrimitiveID: multiraftv1.PrimitiveID(primitiveID),
				},
			},
			SequenceNum: p.session.nextRequestNum(),
		},
		ClosePrimitiveInput: multiraftv1.ClosePrimitiveInput{
			PrimitiveID: multiraftv1.PrimitiveID(primitiveID),
		},
	}
	client := multiraftv1.NewSessionClient(p.session.partition.conn)
	_, err := client.ClosePrimitive(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func newCommandContext(primitive *PrimitiveClient, operationID multiraftv1.OperationID) *CommandContext {
	primitiveID := atomic.LoadUint64(&primitive.id)
	headers := &multiraftv1.CommandRequestHeaders{
		OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
			PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
				SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
					PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
						PartitionID: primitive.session.partition.id,
					},
					SessionID: primitive.session.sessionID,
				},
				PrimitiveID: multiraftv1.PrimitiveID(primitiveID),
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
	primitiveID := atomic.LoadUint64(&primitive.id)
	headers := &multiraftv1.QueryRequestHeaders{
		OperationRequestHeaders: multiraftv1.OperationRequestHeaders{
			PrimitiveRequestHeaders: multiraftv1.PrimitiveRequestHeaders{
				SessionRequestHeaders: multiraftv1.SessionRequestHeaders{
					PartitionRequestHeaders: multiraftv1.PartitionRequestHeaders{
						PartitionID: primitive.session.partition.id,
					},
					SessionID: primitive.session.sessionID,
				},
				PrimitiveID: multiraftv1.PrimitiveID(primitiveID),
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
