// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	counterv1 "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft/node/pkg/multiraft"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/logging"
	"github.com/gogo/protobuf/proto"
)

var log = logging.GetLogger()

func newServer(protocol *multiraft.Protocol) counterv1.CounterServer {
	return &Server{
		protocol: protocol,
	}
}

type Server struct {
	protocol *multiraft.Protocol
}

func (s *Server) Set(ctx context.Context, request *counterv1.SetRequest) (*counterv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", request))
	input, err := proto.Marshal(request)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}

	partition, err := s.protocol.Partition(request.Headers.PartitionID)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}

	service, err := partition.GetService(ctx, request.Headers.ServiceID)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}

	output, headers, err := service.DoCommand(ctx, "Set", input)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}

	response := &counterv1.SetResponse{
		Headers: headers,
	}
	err = proto.Unmarshal(output, &response.SetOutput)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	log.Debugw("Set",
		logging.Stringer("SetResponse", response))
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
