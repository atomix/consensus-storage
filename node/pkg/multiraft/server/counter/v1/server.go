// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	counterv1 "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft/node/pkg/multiraft/server"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/logging"
)

var log = logging.GetLogger()

func newServer(node *server.NodeService) counterv1.CounterServer {
	return &Server{
		node: node,
	}
}

type Server struct {
	node *server.NodeService
}

func (s *Server) Set(ctx context.Context, request *counterv1.SetRequest) (*counterv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", request))
	output, headers, err := server.ExecuteCommand[*counterv1.SetInput, *counterv1.SetOutput](s.node)(ctx, "Set", &request.SetInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.SetResponse{
		Headers:   *headers,
		SetOutput: *output,
	}
	log.Debugw("Set",
		logging.Stringer("SetResponse", response))
	return response, nil
}

func (s *Server) CompareAndSet(ctx context.Context, request *counterv1.CompareAndSetRequest) (*counterv1.CompareAndSetResponse, error) {
	log.Debugw("CompareAndSet",
		logging.Stringer("CompareAndSetRequest", request))
	output, headers, err := server.ExecuteCommand[*counterv1.CompareAndSetInput, *counterv1.CompareAndSetOutput](s.node)(ctx, "CompareAndSet", &request.CompareAndSetInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("CompareAndSet",
			logging.Stringer("CompareAndSetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.CompareAndSetResponse{
		Headers:             *headers,
		CompareAndSetOutput: *output,
	}
	log.Debugw("CompareAndSet",
		logging.Stringer("CompareAndSetResponse", response))
	return response, nil
}

func (s *Server) Get(ctx context.Context, request *counterv1.GetRequest) (*counterv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", request))
	output, headers, err := server.ExecuteQuery[*counterv1.GetInput, *counterv1.GetOutput](s.node)(ctx, "Get", &request.GetInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.GetResponse{
		Headers:   *headers,
		GetOutput: *output,
	}
	log.Debugw("Get",
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *Server) Increment(ctx context.Context, request *counterv1.IncrementRequest) (*counterv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request))
	output, headers, err := server.ExecuteCommand[*counterv1.IncrementInput, *counterv1.IncrementOutput](s.node)(ctx, "Increment", &request.IncrementInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.IncrementResponse{
		Headers:         *headers,
		IncrementOutput: *output,
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementResponse", response))
	return response, nil
}

func (s *Server) Decrement(ctx context.Context, request *counterv1.DecrementRequest) (*counterv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request))
	output, headers, err := server.ExecuteCommand[*counterv1.DecrementInput, *counterv1.DecrementOutput](s.node)(ctx, "Decrement", &request.DecrementInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.DecrementResponse{
		Headers:         *headers,
		DecrementOutput: *output,
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementResponse", response))
	return response, nil
}

var _ counterv1.CounterServer = (*Server)(nil)
