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

func NewPartitionServer(node *protocol.Node) multiraftv1.PartitionServer {
	return &PartitionServer{
		node: node,
	}
}

type PartitionServer struct {
	node *protocol.Node
}

func (s *PartitionServer) OpenSession(ctx context.Context, request *multiraftv1.OpenSessionRequest) (*multiraftv1.OpenSessionResponse, error) {
	log.Debugw("OpenSession",
		logging.Stringer("OpenSessionRequest", request))
	output, headers, err := s.node.OpenSession(ctx, &request.OpenSessionInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("OpenSession",
			logging.Stringer("OpenSessionRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.OpenSessionResponse{
		Headers:           *headers,
		OpenSessionOutput: *output,
	}
	log.Debugw("OpenSession",
		logging.Stringer("OpenSessionRequest", request),
		logging.Stringer("OpenSessionResponse", response))
	return response, nil
}

func (s *PartitionServer) KeepAlive(ctx context.Context, request *multiraftv1.KeepAliveRequest) (*multiraftv1.KeepAliveResponse, error) {
	log.Debugw("KeepAlive",
		logging.Stringer("KeepAliveRequest", request))
	output, headers, err := s.node.KeepAliveSession(ctx, &request.KeepAliveInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("KeepAlive",
			logging.Stringer("KeepAliveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.KeepAliveResponse{
		Headers:         *headers,
		KeepAliveOutput: *output,
	}
	log.Debugw("KeepAlive",
		logging.Stringer("KeepAliveRequest", request),
		logging.Stringer("KeepAliveResponse", response))
	return response, nil
}

func (s *PartitionServer) CloseSession(ctx context.Context, request *multiraftv1.CloseSessionRequest) (*multiraftv1.CloseSessionResponse, error) {
	log.Debugw("CloseSession",
		logging.Stringer("CloseSessionRequest", request))
	output, headers, err := s.node.CloseSession(ctx, &request.CloseSessionInput, &request.Headers)
	if err != nil {
		err = errors.ToProto(err)
		log.Warnw("CloseSession",
			logging.Stringer("CloseSessionRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.CloseSessionResponse{
		Headers:            *headers,
		CloseSessionOutput: *output,
	}
	log.Debugw("CloseSession",
		logging.Stringer("CloseSessionRequest", request),
		logging.Stringer("CloseSessionResponse", response))
	return response, nil
}

func (s *PartitionServer) Watch(request *multiraftv1.WatchPartitionRequest, server multiraftv1.Partition_WatchServer) error {
	log.Debugw("Watch",
		logging.Stringer("WatchPartitionRequest", request))
	ch := make(chan multiraftv1.PartitionEvent)
	partition, ok := s.node.Partition(request.PartitionID)
	if !ok {
		err := errors.NewUnavailable("partition %d not found", request.PartitionID)
		err = errors.ToProto(err)
		log.Warnw("Watch",
			logging.Stringer("WatchPartitionRequest", request),
			logging.Error("Error", err))
		return err
	}
	partition.Watch(server.Context(), ch)
	for event := range ch {
		log.Debugw("Watch",
			logging.Stringer("WatchPartitionRequest", request),
			logging.Stringer("PartitionEvent", &event))
		err := server.Send(&event)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Watch",
				logging.Stringer("WatchPartitionRequest", request),
				logging.Stringer("PartitionEvent", &event),
				logging.Error("Error", err))
			return err
		}
	}
	return nil
}
