// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	counterv1 "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft/node/pkg/primitive"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/gogo/protobuf/proto"
)

var log = logging.GetLogger()

func newServer(protocol primitive.Protocol) counterv1.CounterServer {
	return &Server{
		protocol: protocol,
	}
}

type Server struct {
	protocol primitive.Protocol
}

func (s *Server) Set(ctx context.Context, request *counterv1.SetRequest) (*counterv1.SetResponse, error) {
	log.Debugw("Set",
		logging.Stringer("SetRequest", request))
	inputBytes, err := proto.Marshal(&request.SetInput)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	outputBytes, responseHeaders, err := s.protocol.Command(ctx, inputBytes, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	var output counterv1.SetOutput
	if err := proto.Unmarshal(outputBytes, &output); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Set",
			logging.Stringer("SetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.SetResponse{
		Headers:   *responseHeaders,
		SetOutput: output,
	}
	log.Debugw("Set",
		logging.Stringer("SetResponse", response))
	return response, nil
}

func (s *Server) CompareAndSet(ctx context.Context, request *counterv1.CompareAndSetRequest) (*counterv1.CompareAndSetResponse, error) {
	log.Debugw("CompareAndSet",
		logging.Stringer("CompareAndSetRequest", request))
	inputBytes, err := proto.Marshal(&request.CompareAndSetInput)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("CompareAndSet",
			logging.Stringer("CompareAndSetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	outputBytes, responseHeaders, err := s.protocol.Command(ctx, inputBytes, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("CompareAndSet",
			logging.Stringer("CompareAndSetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	var output counterv1.CompareAndSetOutput
	if err := proto.Unmarshal(outputBytes, &output); err != nil {
		err = errors.ToProto(err)
		log.Warnw("CompareAndSet",
			logging.Stringer("CompareAndSetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.CompareAndSetResponse{
		Headers:             *responseHeaders,
		CompareAndSetOutput: output,
	}
	log.Debugw("CompareAndSet",
		logging.Stringer("CompareAndSetResponse", response))
	return response, nil
}

func (s *Server) Get(ctx context.Context, request *counterv1.GetRequest) (*counterv1.GetResponse, error) {
	log.Debugw("Get",
		logging.Stringer("GetRequest", request))
	inputBytes, err := proto.Marshal(&request.GetInput)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	outputBytes, responseHeaders, err := s.protocol.Query(ctx, inputBytes, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	var output counterv1.GetOutput
	if err := proto.Unmarshal(outputBytes, &output); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Get",
			logging.Stringer("GetRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.GetResponse{
		Headers:   *responseHeaders,
		GetOutput: output,
	}
	log.Debugw("Get",
		logging.Stringer("GetResponse", response))
	return response, nil
}

func (s *Server) Increment(ctx context.Context, request *counterv1.IncrementRequest) (*counterv1.IncrementResponse, error) {
	log.Debugw("Increment",
		logging.Stringer("IncrementRequest", request))
	inputBytes, err := proto.Marshal(&request.IncrementInput)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	outputBytes, responseHeaders, err := s.protocol.Command(ctx, inputBytes, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	var output counterv1.IncrementOutput
	if err := proto.Unmarshal(outputBytes, &output); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Increment",
			logging.Stringer("IncrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.IncrementResponse{
		Headers:         *responseHeaders,
		IncrementOutput: output,
	}
	log.Debugw("Increment",
		logging.Stringer("IncrementResponse", response))
	return response, nil
}

func (s *Server) Decrement(ctx context.Context, request *counterv1.DecrementRequest) (*counterv1.DecrementResponse, error) {
	log.Debugw("Decrement",
		logging.Stringer("DecrementRequest", request))
	inputBytes, err := proto.Marshal(&request.DecrementInput)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	outputBytes, responseHeaders, err := s.protocol.Command(ctx, inputBytes, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	var output counterv1.DecrementOutput
	if err := proto.Unmarshal(outputBytes, &output); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Decrement",
			logging.Stringer("DecrementRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &counterv1.DecrementResponse{
		Headers:         *responseHeaders,
		DecrementOutput: output,
	}
	log.Debugw("Decrement",
		logging.Stringer("DecrementResponse", response))
	return response, nil
}

var _ counterv1.CounterServer = (*Server)(nil)
