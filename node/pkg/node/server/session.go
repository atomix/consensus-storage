// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

func NewSessionServer(node *protocol.Node) multiraftv1.SessionServer {
	return &SessionServer{
		node: node,
	}
}

type SessionServer struct {
	node *protocol.Node
}

func (s *SessionServer) CreatePrimitive(ctx context.Context, request *multiraftv1.CreatePrimitiveRequest) (*multiraftv1.CreatePrimitiveResponse, error) {
	log.Debugw("CreatePrimitive",
		logging.Stringer("CreatePrimitiveRequest", request))
	output, headers, err := s.node.CreatePrimitive(ctx, &request.CreatePrimitiveInput, &request.Headers)
	if err != nil {
		log.Warnw("CreatePrimitive",
			logging.Stringer("CreatePrimitiveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
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
	output, headers, err := s.node.ClosePrimitive(ctx, &request.ClosePrimitiveInput, &request.Headers)
	if err != nil {
		log.Warnw("ClosePrimitive",
			logging.Stringer("ClosePrimitiveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
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
