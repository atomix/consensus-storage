// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/stream"
)

func newPrimitiveClient(session *SessionClient) *PrimitiveClient {
	return &PrimitiveClient{
		session: session,
	}
}

type PrimitiveClient struct {
	session *SessionClient
}

func (p *PrimitiveClient) Command(id multiraftv1.OperationID) *CommandClient {
	return newCommandClient(p, id)
}

func (p *PrimitiveClient) Query(id multiraftv1.OperationID) *QueryClient {
	return newQueryClient(p, id)
}

func newOperationClient(primitive *PrimitiveClient, id multiraftv1.OperationID) *OperationClient {
	return &OperationClient{
		PrimitiveClient: primitive,
		id:              id,
	}
}

type OperationClient struct {
	*PrimitiveClient
	id multiraftv1.OperationID
}

func (o *OperationClient) ID() multiraftv1.OperationID {
	return o.id
}

func newCommandClient(primitive *PrimitiveClient, id multiraftv1.OperationID) *CommandClient {
	return &CommandClient{
		OperationClient: newOperationClient(primitive, id),
	}
}

type CommandClient struct {
	*OperationClient
}

func (c *CommandClient) Execute(ctx context.Context, input []byte) ([]byte, error) {

}

func (c *CommandClient) ExecuteStream(ctx context.Context, input []byte, stream stream.WriteStream) error {

}

func newQueryClient(primitive *PrimitiveClient, id multiraftv1.OperationID) *QueryClient {
	return &QueryClient{
		OperationClient: newOperationClient(primitive, id),
	}
}

type QueryClient struct {
	*OperationClient
}

func (c *QueryClient) Execute(ctx context.Context, input []byte) ([]byte, error) {

}

func (c *QueryClient) ExecuteStream(ctx context.Context, input []byte, stream stream.WriteStream) error {

}
