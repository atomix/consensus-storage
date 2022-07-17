// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"github.com/atomix/multi-raft/driver/pkg/client"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	"google.golang.org/grpc"
)

func NewClient(client *client.Client) counterv1.CounterClient {
	return &Client{
		Client: client,
	}
}

type Client struct {
	*client.Client
}

func (c *Client) Create(ctx context.Context, request *counterv1.CreateRequest, opts ...grpc.CallOption) (*counterv1.CreateResponse, error) {
	partition := c.PartitionBy([]byte(request.ID.Name))
}

func (c *Client) Close(ctx context.Context, request *counterv1.CloseRequest, opts ...grpc.CallOption) (*counterv1.CloseResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Set(ctx context.Context, request *counterv1.SetRequest, opts ...grpc.CallOption) (*counterv1.SetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) CompareAndSet(ctx context.Context, request *counterv1.CompareAndSetRequest, opts ...grpc.CallOption) (*counterv1.CompareAndSetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Get(ctx context.Context, request *counterv1.GetRequest, opts ...grpc.CallOption) (*counterv1.GetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Increment(ctx context.Context, request *counterv1.IncrementRequest, opts ...grpc.CallOption) (*counterv1.IncrementResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Decrement(ctx context.Context, request *counterv1.DecrementRequest, opts ...grpc.CallOption) (*counterv1.DecrementResponse, error) {
	//TODO implement me
	panic("implement me")
}

var _ counterv1.CounterClient = (*Client)
