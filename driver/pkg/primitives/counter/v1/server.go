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
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/stringer"
	"google.golang.org/grpc"
)

var log = logging.GetLogger()

const Service = "atomix.runtime.counter.v1.Counter"

const truncLen = 200

func NewCounterServer(protocol *client.Protocol, config api.CounterConfig) counterv1.CounterServer {
	return &multiRaftCounterServer{
		Protocol: protocol,
	}
}

type multiRaftCounterServer struct {
	*client.Protocol
}

func (s *multiRaftCounterServer) Create(ctx context.Context, request *counterv1.CreateRequest) (*counterv1.CreateResponse, error) {
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := session.CreatePrimitive(ctx, request.ID.Name, Service); err != nil {
		log.Warnw("Create",
			logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.CreateResponse{}
	log.Debugw("Create",
		logging.Stringer("CreateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CreateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftCounterServer) Close(ctx context.Context, request *counterv1.CloseRequest) (*counterv1.CloseResponse, error) {
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Close",
			logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := session.ClosePrimitive(ctx, request.ID.Name); err != nil {
		log.Warnw("Close",
			logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.CloseResponse{}
	log.Debugw("Close",
		logging.Stringer("CloseRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("CloseResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftCounterServer) Set(ctx context.Context, request *counterv1.SetRequest) (*counterv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.SetResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.SetResponse, error) {
		return api.NewCounterClient(conn).Set(ctx, &api.SetRequest{
			Headers: headers,
			SetInput: &api.SetInput{
				Value: request.Value,
			},
		})
	})
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.SetResponse{
		Value: output.Value,
	}
	log.Debugw("Set",
		logging.Stringer("SetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("SetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftCounterServer) Get(ctx context.Context, request *counterv1.GetRequest) (*counterv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Query[*api.GetResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.QueryRequestHeaders) (*api.GetResponse, error) {
		return api.NewCounterClient(conn).Get(ctx, &api.GetRequest{
			Headers:  headers,
			GetInput: &api.GetInput{},
		})
	})
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.GetResponse{
		Value: output.Value,
	}
	log.Debugw("Get",
		logging.Stringer("GetRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("GetResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftCounterServer) Increment(ctx context.Context, request *counterv1.IncrementRequest) (*counterv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.IncrementResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.IncrementResponse, error) {
		return api.NewCounterClient(conn).Increment(ctx, &api.IncrementRequest{
			Headers: headers,
			IncrementInput: &api.IncrementInput{
				Delta: request.Delta,
			},
		})
	})
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.IncrementResponse{
		Value: output.Value,
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("IncrementResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftCounterServer) Decrement(ctx context.Context, request *counterv1.DecrementRequest) (*counterv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.DecrementResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.DecrementResponse, error) {
		return api.NewCounterClient(conn).Decrement(ctx, &api.DecrementRequest{
			Headers: headers,
			DecrementInput: &api.DecrementInput{
				Delta: request.Delta,
			},
		})
	})
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.DecrementResponse{
		Value: output.Value,
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("DecrementResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

func (s *multiRaftCounterServer) Update(ctx context.Context, request *counterv1.UpdateRequest) (*counterv1.UpdateResponse, error) {
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)))
	partition := s.PartitionBy([]byte(request.ID.Name))
	session, err := partition.GetSession(ctx)
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	primitive, err := session.GetPrimitive(request.ID.Name)
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	command := client.Command[*api.UpdateResponse](primitive)
	output, err := command.Run(func(conn *grpc.ClientConn, headers *multiraftv1.CommandRequestHeaders) (*api.UpdateResponse, error) {
		return api.NewCounterClient(conn).Update(ctx, &api.UpdateRequest{
			Headers: headers,
			UpdateInput: &api.UpdateInput{
				Compare: request.Check,
				Update:  request.Update,
			},
		})
	})
	if err != nil {
		log.Warnw("Update",
			logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &counterv1.UpdateResponse{
		Value: output.Value,
	}
	log.Debugw("Update",
		logging.Stringer("UpdateRequest", stringer.Truncate(request, truncLen)),
		logging.Stringer("UpdateResponse", stringer.Truncate(response, truncLen)))
	return response, nil
}

var _ counterv1.CounterServer = (*multiRaftCounterServer)(nil)
