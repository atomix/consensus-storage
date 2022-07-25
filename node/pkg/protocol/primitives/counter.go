// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitives

import (
	"context"
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
)

var log = logging.GetLogger()

func RegisterCounterServer(server *grpc.Server, node *protocol.Node) {
	counterv1.RegisterCounterServer(server, newCounterServer(node))
}

var counterCodec = protocol.NewCodec[*counterv1.CounterInput, *counterv1.CounterOutput](
	func(input *counterv1.CounterInput) ([]byte, error) {
		return proto.Marshal(input)
	},
	func(bytes []byte) (*counterv1.CounterOutput, error) {
		output := &counterv1.CounterOutput{}
		if err := proto.Unmarshal(bytes, output); err != nil {
			return nil, err
		}
		return output, nil
	})

func newCounterServer(node *protocol.Node) counterv1.CounterServer {
	return &Server{
		protocol: protocol.NewProtocol[*counterv1.CounterInput, *counterv1.CounterOutput](node, counterCodec),
	}
}

type Server struct {
	protocol protocol.Protocol[*counterv1.CounterInput, *counterv1.CounterOutput]
}

func (s *Server) Set(ctx context.Context, request *counterv1.SetRequest) (*counterv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", request))
	input := &counterv1.CounterInput{
		Input: &counterv1.CounterInput_Set{
			Set: &request.SetInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, &request.Headers)
	if err != nil {
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.SetResponse{
		Headers:   *headers,
		SetOutput: *output.GetSet(),
	}
	log.Debugw("Set",
		logging.Stringer("SetResponse", response))
	return response, nil
}

func (s *Server) CompareAndSet(ctx context.Context, request *counterv1.CompareAndSetRequest) (*counterv1.CompareAndSetResponse, error) {
	log.Debugw("CompareAndSet",
		logging.Stringer("CompareAndSetRequest", request))
	input := &counterv1.CounterInput{
		Input: &counterv1.CounterInput_CompareAndSet{
			CompareAndSet: &request.CompareAndSetInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, &request.Headers)
	if err != nil {
		log.Warnw("CompareAndSet",
			logging.Stringer("CompareAndSetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.CompareAndSetResponse{
		Headers:             *headers,
		CompareAndSetOutput: *output.GetCompareAndSet(),
	}
	log.Debugw("CompareAndSet",
		logging.Stringer("CompareAndSetResponse", response))
	return response, nil
}

func (s *Server) Get(ctx context.Context, request *counterv1.GetRequest) (*counterv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", request))
	input := &counterv1.CounterInput{
		Input: &counterv1.CounterInput_Get{
			Get: &request.GetInput,
		},
	}
	output, headers, err := s.protocol.Query(ctx, input, &request.Headers)
	if err != nil {
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.GetResponse{
		Headers:   *headers,
		GetOutput: *output.GetGet(),
	}
	log.Debugw("Get",
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *Server) Increment(ctx context.Context, request *counterv1.IncrementRequest) (*counterv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request))
	input := &counterv1.CounterInput{
		Input: &counterv1.CounterInput_Increment{
			Increment: &request.IncrementInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, &request.Headers)
	if err != nil {
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.IncrementResponse{
		Headers:         *headers,
		IncrementOutput: *output.GetIncrement(),
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementResponse", response))
	return response, nil
}

func (s *Server) Decrement(ctx context.Context, request *counterv1.DecrementRequest) (*counterv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request))
	input := &counterv1.CounterInput{
		Input: &counterv1.CounterInput_Decrement{
			Decrement: &request.DecrementInput,
		},
	}
	output, headers, err := s.protocol.Command(ctx, input, &request.Headers)
	if err != nil {
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.DecrementResponse{
		Headers:         *headers,
		DecrementOutput: *output.GetDecrement(),
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementResponse", response))
	return response, nil
}

var _ counterv1.CounterServer = (*Server)(nil)
