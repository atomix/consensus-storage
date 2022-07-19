// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/driver/pkg/client"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	"github.com/atomix/runtime/pkg/runtime"
	"google.golang.org/grpc"
)

const Type = "Counter"
const APIVersion = "v1"

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
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	spec := multiraftv1.PrimitiveSpec{
		Type: multiraftv1.PrimitiveType{
			Name:       Type,
			ApiVersion: APIVersion,
		},
		Namespace: runtime.GetNamespace(),
		Name:      request.ID.Name,
	}
	if err := session.CreatePrimitive(ctx, spec, opts...); err != nil {
		return nil, err
	}
	response := &counterv1.CreateResponse{}
	return response, nil
}

func (c *Client) Close(ctx context.Context, request *counterv1.CloseRequest, opts ...grpc.CallOption) (*counterv1.CloseResponse, error) {
	partition := c.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := session.ClosePrimitive(ctx, request.ID.Name, opts...); err != nil {
		return nil, err
	}
	response := &counterv1.CloseResponse{}
	return response, nil
}

func (c *Client) Set(ctx context.Context, request *counterv1.SetRequest, opts ...grpc.CallOption) (*counterv1.SetResponse, error) {
	partition := c.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(ctx, request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := primitive.Command("Set")
	defer command.Done()
	client := api.NewCounterClient(primitive.Conn())
	input := &api.SetRequest{
		Headers: *command.Headers(),
		SetInput: api.SetInput{
			Value: request.Value,
		},
	}
	output, err := client.Set(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	command.Receive(&output.Headers)
	response := &counterv1.SetResponse{
		Value: output.Value,
	}
	return response, nil
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

var _ counterv1.CounterClient = (*Client)(nil)
