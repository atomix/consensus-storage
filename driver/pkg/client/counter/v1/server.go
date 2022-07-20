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
	command := primitive.Command("Set")
	defer command.Done()
	client := api.NewCounterClient(primitive.Conn())
	input := &api.SetRequest{
		Headers: *command.Headers(),
		SetInput: api.SetInput{
			Value: request.Value,
		},
	}
	output, err := client.Set(ctx, input)
	if err != nil {
		return nil, err
	}
	command.Receive(&output.Headers)
	response := &counterv1.SetResponse{
		Value: output.Value,
	}
	return response, nil
}

func (s *Server) CompareAndSet(ctx context.Context, request *counterv1.CompareAndSetRequest) (*counterv1.CompareAndSetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Get(ctx context.Context, request *counterv1.GetRequest) (*counterv1.GetResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Increment(ctx context.Context, request *counterv1.IncrementRequest) (*counterv1.IncrementResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Decrement(ctx context.Context, request *counterv1.DecrementRequest) (*counterv1.DecrementResponse, error) {
	//TODO implement me
	panic("implement me")
}

var _ counterv1.CounterServer = (*Server)(nil)
