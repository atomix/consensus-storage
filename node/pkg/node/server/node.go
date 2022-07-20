// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/manager"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

var log = logging.GetLogger()

func NewNodeServer(node *manager.NodeManager) multiraftv1.NodeServer {
	return &nodeServer{
		node: node,
	}
}

type nodeServer struct {
	node *manager.NodeManager
}

func (s *nodeServer) Bootstrap(ctx context.Context, request *multiraftv1.BootstrapRequest) (*multiraftv1.BootstrapResponse, error) {
	log.Debugw("Bootstrap",
		logging.Stringer("BootstrapRequest", request))
	if err := s.node.Bootstrap(request.Cluster); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Bootstrap",
			logging.Stringer("BootstrapRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.BootstrapResponse{}
	log.Debugw("Bootstrap",
		logging.Stringer("BootstrapRequest", request),
		logging.Stringer("BootstrapResponse", response))
	return response, nil
}

func (s *nodeServer) Join(ctx context.Context, request *multiraftv1.JoinRequest) (*multiraftv1.JoinResponse, error) {
	log.Debugw("Join",
		logging.Stringer("JoinRequest", request))
	if err := s.node.Join(request.Partition); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Join",
			logging.Stringer("JoinRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.JoinResponse{}
	log.Debugw("Join",
		logging.Stringer("JoinRequest", request),
		logging.Stringer("JoinResponse", response))
	return response, nil
}

func (s *nodeServer) Leave(ctx context.Context, request *multiraftv1.LeaveRequest) (*multiraftv1.LeaveResponse, error) {
	log.Debugw("Leave",
		logging.Stringer("LeaveRequest", request))
	if err := s.node.Leave(request.PartitionID); err != nil {
		err = errors.ToProto(err)
		log.Warnw("Leave",
			logging.Stringer("LeaveRequest", request),
			logging.Error("Error", err))
		return nil, err
	}
	response := &multiraftv1.LeaveResponse{}
	log.Debugw("Leave",
		logging.Stringer("LeaveRequest", request),
		logging.Stringer("LeaveResponse", response))
	return response, nil
}

func (s *nodeServer) Watch(request *multiraftv1.WatchNodeRequest, server multiraftv1.Node_WatchServer) error {
	log.Debugw("Watch",
		logging.Stringer("WatchRequest", request))
	ch := make(chan multiraftv1.NodeEvent)
	s.node.Watch(server.Context(), ch)
	for event := range ch {
		log.Debugw("Watch",
			logging.Stringer("WatchRequest", request),
			logging.Stringer("NodeEvent", &event))
		err := server.Send(&event)
		if err != nil {
			err = errors.ToProto(err)
			log.Warnw("Watch",
				logging.Stringer("WatchRequest", request),
				logging.Stringer("NodeEvent", &event),
				logging.Error("Error", err))
			return err
		}
	}
	return nil
}
