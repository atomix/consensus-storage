// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft/api/atomix/multiraft/map/v1"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/driver/pkg/client"
	"github.com/atomix/multi-raft/driver/pkg/util/async"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	runtimev1 "github.com/atomix/runtime/api/atomix/runtime/v1"
	"github.com/atomix/runtime/pkg/runtime"
	"io"
)

const Type = "Map"
const APIVersion = "v1"

func NewServer(protocol *client.Protocol) mapv1.MapServer {
	return &Server{
		Protocol: protocol,
	}
}

type Server struct {
	*client.Protocol
}

func (s *Server) Create(ctx context.Context, request *mapv1.CreateRequest) (*mapv1.CreateResponse, error) {
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return err
		}
		spec := multiraftv1.PrimitiveSpec{
			Type: multiraftv1.PrimitiveType{
				Name:       Type,
				ApiVersion: APIVersion,
			},
			Namespace: runtime.GetNamespace(),
			Name:      request.ID.Name,
		}
		return session.CreatePrimitive(ctx, spec)
	})
	if err != nil {
		return nil, err
	}
	response := &mapv1.CreateResponse{}
	return response, nil
}

func (s *Server) Close(ctx context.Context, request *mapv1.CloseRequest) (*mapv1.CloseResponse, error) {
	partitions := s.Partitions()
	err := async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return err
		}
		return session.ClosePrimitive(ctx, request.ID.Name)
	})
	if err != nil {
		return nil, err
	}
	response := &mapv1.CloseResponse{}
	return response, nil
}

func (s *Server) Size(ctx context.Context, request *mapv1.SizeRequest) (*mapv1.SizeResponse, error) {
	partitions := s.Partitions()
	sizes, err := async.ExecuteAsync[int](len(partitions), func(i int) (int, error) {
		partition := partitions[i]
		session, err := partition.GetSession(ctx)
		if err != nil {
			return 0, err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			return 0, err
		}
		query := primitive.Query("Size")
		client := api.NewMapClient(primitive.Conn())
		input := &api.SizeRequest{
			Headers:   *query.Headers(),
			SizeInput: api.SizeInput{},
		}
		output, err := client.Size(ctx, input)
		if err != nil {
			return 0, err
		}
		return int(output.Size_), nil
	})
	if err != nil {
		return nil, err
	}
	var size int
	for _, s := range sizes {
		size += s
	}
	response := &mapv1.SizeResponse{
		Size_: uint32(size),
	}
	return response, nil
}

func (s *Server) Put(ctx context.Context, request *mapv1.PutRequest) (*mapv1.PutResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Insert(ctx context.Context, request *mapv1.InsertRequest) (*mapv1.InsertResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Update(ctx context.Context, request *mapv1.UpdateRequest) (*mapv1.UpdateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Get(ctx context.Context, request *mapv1.GetRequest) (*mapv1.GetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Remove(ctx context.Context, request *mapv1.RemoveRequest) (*mapv1.RemoveResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Clear(ctx context.Context, request *mapv1.ClearRequest) (*mapv1.ClearResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Events(request *mapv1.EventsRequest, server mapv1.Map_EventsServer) error {
	partitions := s.Partitions()
	return async.IterAsync(len(partitions), func(i int) error {
		partition := partitions[i]
		session, err := partition.GetSession(server.Context())
		if err != nil {
			return err
		}
		primitive, err := session.GetPrimitive(request.ID.Name)
		if err != nil {
			return err
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
		stream, err := client.Events(server.Context(), input)
		if err != nil {
			return err
		}
		for {
			output, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			response := &mapv1.EventsResponse{
				Event: mapv1.Event{
					Type: mapv1.Event_Type(output.Event.Type),
					Entry: mapv1.Entry{
						Key: output.Event.Entry.Key,
						Value: &mapv1.Value{
							Value: output.Event.Entry.Value.Value,
							TTL:   output.Event.Entry.Value.TTL,
						},
						Timestamp: &runtimev1.Timestamp{
							Timestamp: &runtimev1.Timestamp_LogicalTimestamp{
								LogicalTimestamp: &runtimev1.LogicalTimestamp{
									Time: runtimev1.LogicalTime(output.Event.Entry.Index),
								},
							},
						},
					},
				},
			}
			if err := server.Send(response); err != nil {
				return err
			}
		}
	})
}

func (s *Server) Entries(request *mapv1.EntriesRequest, server mapv1.Map_EntriesServer) error {
	//TODO implement me
	panic("implement me")
}

var _ mapv1.MapServer = (*Server)(nil)
