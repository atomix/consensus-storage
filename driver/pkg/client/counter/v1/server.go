// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	api "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"google.golang.org/grpc"
)

const Type = "Counter"
const APIVersion = "v1"

func NewServer(protocol *client.Protocol) counterv1.CounterServer {
	return &Server{
		Protocol: protocol,
	}
}

type Server struct {
	*client.Protocol
}

func (s *Server) Create(ctx context.Context, request *counterv1.CreateRequest) (*counterv1.CreateResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
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
	if err := session.CreatePrimitive(ctx, spec); err != nil {
		return nil, err
	}
	response := &counterv1.CreateResponse{}
	return response, nil
}

func (s *Server) Close(ctx context.Context, request *counterv1.CloseRequest) (*counterv1.CloseResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := session.ClosePrimitive(ctx, request.ID.Name); err != nil {
		return nil, err
	}
	response := &counterv1.CloseResponse{}
	return response, nil
}

func (s *Server) Set(ctx context.Context, request *counterv1.SetRequest) (*counterv1.SetResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := client.Command[*api.SetResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.SetResponse, error) {
		return api.NewCounterClient(conn).Set(ctx, &api.SetRequest{
			Headers: *headers,
			SetInput: api.SetInput{
				Value: request.Value,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	response := &counterv1.SetResponse{
		Value: output.Value,
	}
	return response, nil
}

func (s *Server) CompareAndSet(ctx context.Context, request *counterv1.CompareAndSetRequest) (*counterv1.CompareAndSetResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := client.Command[*api.CompareAndSetResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.CompareAndSetResponse, error) {
		return api.NewCounterClient(conn).CompareAndSet(ctx, &api.CompareAndSetRequest{
			Headers: *headers,
			CompareAndSetInput: api.CompareAndSetInput{
				Compare: request.Check,
				Update:  request.Update,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	response := &counterv1.CompareAndSetResponse{
		Value: output.Value,
	}
	return response, nil
}

func (s *Server) Get(ctx context.Context, request *counterv1.GetRequest) (*counterv1.GetResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := client.Query[*api.GetResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.GetResponse, error) {
		return api.NewCounterClient(conn).Get(ctx, &api.GetRequest{
			Headers:  *headers,
			GetInput: api.GetInput{},
		})
	})
	if err != nil {
		return nil, err
	}
	response := &counterv1.GetResponse{
		Value: output.Value,
	}
	return response, nil
}

func (s *Server) Increment(ctx context.Context, request *counterv1.IncrementRequest) (*counterv1.IncrementResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := client.Command[*api.IncrementResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.IncrementResponse, error) {
		return api.NewCounterClient(conn).Increment(ctx, &api.IncrementRequest{
			Headers: *headers,
			IncrementInput: api.IncrementInput{
				Delta: request.Delta,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	response := &counterv1.IncrementResponse{
		Value: output.Value,
	}
	return response, nil
}

func (s *Server) Decrement(ctx context.Context, request *counterv1.DecrementRequest) (*counterv1.DecrementResponse, error) {
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		return nil, err
	}
	command := client.Command[*api.DecrementResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.DecrementResponse, error) {
		return api.NewCounterClient(conn).Decrement(ctx, &api.DecrementRequest{
			Headers: *headers,
			DecrementInput: api.DecrementInput{
				Delta: request.Delta,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	response := &counterv1.DecrementResponse{
		Value: output.Value,
	}
	return response, nil
}

var _ counterv1.CounterServer = (*Server)(nil)
