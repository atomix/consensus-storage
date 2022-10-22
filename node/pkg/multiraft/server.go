// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package multiraft

import (
	"context"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

func NewNodeServer(protocol *Protocol) NodeServer {
	return &nodeServer{
		protocol: protocol,
	}
}

type nodeServer struct {
	protocol *Protocol
}

func (s *nodeServer) Bootstrap(ctx context.Context, request *BootstrapRequest) (*BootstrapResponse, error) {
	log.Debugw("Bootstrap",
		logging.Stringer("BootstrapRequest", request))
	if err := s.protocol.Bootstrap(request.Group); err != nil {
		log.Warnw("Bootstrap",
			logging.Stringer("BootstrapRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &BootstrapResponse{}
	log.Debugw("Bootstrap",
		logging.Stringer("BootstrapRequest", request),
		logging.Stringer("BootstrapResponse", response))
	return response, nil
}

func (s *nodeServer) Join(ctx context.Context, request *JoinRequest) (*JoinResponse, error) {
	log.Debugw("Join",
		logging.Stringer("JoinRequest", request))
	if err := s.protocol.Join(request.Group); err != nil {
		log.Warnw("Join",
			logging.Stringer("JoinRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &JoinResponse{}
	log.Debugw("Join",
		logging.Stringer("JoinRequest", request),
		logging.Stringer("JoinResponse", response))
	return response, nil
}

func (s *nodeServer) Leave(ctx context.Context, request *LeaveRequest) (*LeaveResponse, error) {
	log.Debugw("Leave",
		logging.Stringer("LeaveRequest", request))
	if err := s.protocol.Leave(request.GroupID); err != nil {
		log.Warnw("Leave",
			logging.Stringer("LeaveRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &LeaveResponse{}
	log.Debugw("Leave",
		logging.Stringer("LeaveRequest", request),
		logging.Stringer("LeaveResponse", response))
	return response, nil
}

func (s *nodeServer) Watch(request *WatchRequest, server Node_WatchServer) error {
	log.Debugw("Watch",
		logging.Stringer("WatchRequest", request))
	ch := make(chan Event, 100)
	go s.protocol.Watch(server.Context(), ch)
	for event := range ch {
		log.Debugw("Watch",
			logging.Stringer("WatchRequest", request),
			logging.Stringer("NodeEvent", &event))
		err := server.Send(&event)
		if err != nil {
			log.Warnw("Watch",
				logging.Stringer("WatchRequest", request),
				logging.Stringer("NodeEvent", &event),
				logging.Error("Error", err))
			return errors.ToProto(err)
		}
	}
	return nil
}
