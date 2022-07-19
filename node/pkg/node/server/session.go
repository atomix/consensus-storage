// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/errors"
	"github.com/atomix/runtime/pkg/logging"
)

func newSessionServer(protocol *Protocol) multiraftv1.SessionServer {
	return &SessionServer{
		protocol: protocol,
	}
}

type SessionServer struct {
	protocol *Protocol
}

func (s *SessionServer) CreatePrimitive(ctx context.Context, request *multiraftv1.CreatePrimitiveRequest) (*multiraftv1.CreatePrimitiveResponse, error) {
	log.Debugw("CreatePrimitive",
		logging.Stringer("CreatePrimitiveRequest", request))
	output, headers, err := s.protocol.CreatePrimitive(ctx, &request.CreatePrimitiveInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("CreatePrimitive",
			logging.Stringer("CreatePrimitiveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.CreatePrimitiveResponse{
		Headers:               *headers,
		CreatePrimitiveOutput: *output,
	}
	log.Debugw("CreatePrimitive",
		logging.Stringer("CreatePrimitiveRequest", request),
		logging.Stringer("CreatePrimitiveResponse", response))
	return response, nil
}

func (s *SessionServer) ClosePrimitive(ctx context.Context, request *multiraftv1.ClosePrimitiveRequest) (*multiraftv1.ClosePrimitiveResponse, error) {
	log.Debugw("ClosePrimitive",
		logging.Stringer("ClosePrimitiveRequest", request))
	output, headers, err := s.protocol.ClosePrimitive(ctx, &request.ClosePrimitiveInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("ClosePrimitive",
			logging.Stringer("ClosePrimitiveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.ClosePrimitiveResponse{
		Headers:              *headers,
		ClosePrimitiveOutput: *output,
	}
	log.Debugw("ClosePrimitive",
		logging.Stringer("ClosePrimitiveRequest", request),
		logging.Stringer("ClosePrimitiveResponse", response))
	return response, nil
}
