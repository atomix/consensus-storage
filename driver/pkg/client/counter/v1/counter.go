// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	"google.golang.org/grpc"
)

type Client struct {
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
