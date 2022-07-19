// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft/api/atomix/multiraft/map/v1"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/driver/pkg/client"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	runtimev1 "github.com/atomix/runtime/api/atomix/runtime/v1"
	"github.com/atomix/runtime/pkg/runtime"
	"google.golang.org/grpc"
)

const Type = "Map"
const APIVersion = "v1"

func NewClient(client *client.Client) mapv1.MapClient {
	return &Client{
		Client: client,
	}
}

type Client struct {
	*client.Client
}

func (c *Client) Create(ctx context.Context, request *mapv1.CreateRequest, opts ...grpc.CallOption) (*mapv1.CreateResponse, error) {
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
	response := &mapv1.CreateResponse{}
	return response, nil
}

func (c *Client) Close(ctx context.Context, request *mapv1.CloseRequest, opts ...grpc.CallOption) (*mapv1.CloseResponse, error) {
	partition := c.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := session.ClosePrimitive(ctx, request.ID.Name, opts...); err != nil {
		return nil, err
	}
	response := &mapv1.CloseResponse{}
	return response, nil
}

func (c *Client) Size(ctx context.Context, request *mapv1.SizeRequest, opts ...grpc.CallOption) (*mapv1.SizeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Put(ctx context.Context, request *mapv1.PutRequest, opts ...grpc.CallOption) (*mapv1.PutResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Insert(ctx context.Context, request *mapv1.InsertRequest, opts ...grpc.CallOption) (*mapv1.InsertResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Update(ctx context.Context, request *mapv1.UpdateRequest, opts ...grpc.CallOption) (*mapv1.UpdateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Get(ctx context.Context, request *mapv1.GetRequest, opts ...grpc.CallOption) (*mapv1.GetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Remove(ctx context.Context, request *mapv1.RemoveRequest, opts ...grpc.CallOption) (*mapv1.RemoveResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Clear(ctx context.Context, request *mapv1.ClearRequest, opts ...grpc.CallOption) (*mapv1.ClearResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Events(ctx context.Context, request *mapv1.EventsRequest, opts ...grpc.CallOption) (mapv1.Map_EventsClient, error) {
	partition := c.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(ctx, request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := primitive.Command("Events")
	defer command.Done()
	client := api.NewMapClient(primitive.Conn())
	input := &api.EventsRequest{
		Headers: *command.Headers(),
		EventsInput: api.EventsInput{
			Key:    request.Key,
			Replay: request.Replay,
		},
	}
	stream, err := client.Events(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return newEventsClient(command, stream), nil
}

func (c *Client) Entries(ctx context.Context, request *mapv1.EntriesRequest, opts ...grpc.CallOption) (mapv1.Map_EntriesClient, error) {
	//TODO implement me
	panic("implement me")
}

var _ mapv1.MapClient = (*Client)(nil)

func newEventsClient(command *client.CommandContext, stream api.Map_EventsClient) mapv1.Map_EventsClient {
	return &eventsClient{
		Map_EventsClient: stream,
		command:          command,
	}
}

type eventsClient struct {
	api.Map_EventsClient
	command *client.CommandContext
}

func (c *eventsClient) Recv() (*mapv1.EventsResponse, error) {
	response, err := c.Map_EventsClient.Recv()
	if err != nil {
		c.command.Stream().Close()
		c.command.Stream().Done()
		return nil, err
	}
	c.command.Stream().Receive(&response.Headers)
	return &mapv1.EventsResponse{
		Event: mapv1.Event{
			Type: mapv1.Event_Type(response.Event.Type),
			Entry: mapv1.Entry{
				Key: response.Event.Entry.Key,
				Value: &mapv1.Value{
					Value: response.Event.Entry.Value.Value,
					TTL:   response.Event.Entry.Value.TTL,
				},
				Timestamp: &runtimev1.Timestamp{
					Timestamp: &runtimev1.Timestamp_LogicalTimestamp{
						LogicalTimestamp: &runtimev1.LogicalTimestamp{
							Time: runtimev1.LogicalTime(response.Event.Entry.Index),
						},
					},
				},
			},
		},
	}, nil
}
