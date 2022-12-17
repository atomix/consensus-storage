// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package consensus

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

func (s *nodeServer) GetConfig(ctx context.Context, request *GetConfigRequest) (*GetConfigResponse, error) {
	log.Debugw("GetConfig",
		logging.Stringer("GetConfigRequest", request))
	config, err := s.protocol.GetConfig(request.ShardID)
	if err != nil {
		log.Warnw("GetConfig",
			logging.Stringer("GetConfigRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &GetConfigResponse{
		Shard: config,
	}
	log.Debugw("GetConfig",
		logging.Stringer("GetConfigRequest", request),
		logging.Stringer("GetConfigResponse", response))
	return response, nil
}

func (s *nodeServer) BootstrapShard(ctx context.Context, request *BootstrapShardRequest) (*BootstrapShardResponse, error) {
	log.Debugw("BootstrapShard",
		logging.Stringer("BootstrapShardRequest", request))
	if request.ShardID == 0 {
		err := errors.NewInvalid("shard_id is required")
		log.Warnw("BootstrapShard",
			logging.Stringer("BootstrapShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.ReplicaID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("BootstrapShard",
			logging.Stringer("BootstrapShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.BootstrapShard(ctx, request.ShardID, request.ReplicaID, request.Config, request.Replicas...); err != nil {
		log.Warnw("BootstrapShard",
			logging.Stringer("BootstrapShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &BootstrapShardResponse{}
	log.Debugw("BootstrapShard",
		logging.Stringer("BootstrapShardRequest", request),
		logging.Stringer("BootstrapShardResponse", response))
	return response, nil
}

func (s *nodeServer) AddReplica(ctx context.Context, request *AddReplicaRequest) (*AddReplicaResponse, error) {
	log.Debugw("AddReplica",
		logging.Stringer("AddReplicaRequest", request))
	if request.ShardID == 0 {
		err := errors.NewInvalid("shard_id is required")
		log.Warnw("AddReplica",
			logging.Stringer("AddReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.Version == 0 {
		err := errors.NewInvalid("version is required")
		log.Warnw("AddReplica",
			logging.Stringer("AddReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.AddReplica(ctx, request.ShardID, request.Replica, request.Version); err != nil {
		log.Warnw("AddReplica",
			logging.Stringer("AddReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &AddReplicaResponse{}
	log.Debugw("AddReplica",
		logging.Stringer("AddReplicaRequest", request),
		logging.Stringer("AddReplicaResponse", response))
	return response, nil
}

func (s *nodeServer) RemoveReplica(ctx context.Context, request *RemoveReplicaRequest) (*RemoveReplicaResponse, error) {
	log.Debugw("RemoveReplica",
		logging.Stringer("RemoveReplicaRequest", request))
	if request.ShardID == 0 {
		err := errors.NewInvalid("shard_id is required")
		log.Warnw("RemoveReplica",
			logging.Stringer("RemoveReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.ReplicaID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("RemoveReplica",
			logging.Stringer("RemoveReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.Version == 0 {
		err := errors.NewInvalid("version is required")
		log.Warnw("RemoveReplica",
			logging.Stringer("RemoveReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.RemoveReplica(ctx, request.ShardID, request.ReplicaID, request.Version); err != nil {
		log.Warnw("RemoveReplica",
			logging.Stringer("RemoveReplicaRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &RemoveReplicaResponse{}
	log.Debugw("RemoveReplica",
		logging.Stringer("RemoveReplicaRequest", request),
		logging.Stringer("RemoveReplicaResponse", response))
	return response, nil
}

func (s *nodeServer) JoinShard(ctx context.Context, request *JoinShardRequest) (*JoinShardResponse, error) {
	log.Debugw("JoinShard",
		logging.Stringer("JoinShardRequest", request))
	if request.ShardID == 0 {
		err := errors.NewInvalid("shard_id is required")
		log.Warnw("JoinShard",
			logging.Stringer("JoinShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.ReplicaID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("JoinShard",
			logging.Stringer("JoinShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.JoinShard(ctx, request.ShardID, request.ReplicaID, request.Config); err != nil {
		log.Warnw("JoinShard",
			logging.Stringer("JoinShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &JoinShardResponse{}
	log.Debugw("JoinShard",
		logging.Stringer("JoinShardRequest", request),
		logging.Stringer("JoinShardResponse", response))
	return response, nil
}

func (s *nodeServer) LeaveShard(ctx context.Context, request *LeaveShardRequest) (*LeaveShardResponse, error) {
	log.Debugw("LeaveShard",
		logging.Stringer("LeaveShardRequest", request))
	if request.ShardID == 0 {
		err := errors.NewInvalid("shard_id is required")
		log.Warnw("LeaveShard",
			logging.Stringer("LeaveShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.LeaveShard(ctx, request.ShardID); err != nil {
		log.Warnw("LeaveShard",
			logging.Stringer("LeaveShardRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &LeaveShardResponse{}
	log.Debugw("LeaveShard",
		logging.Stringer("LeaveShardRequest", request),
		logging.Stringer("LeaveShardResponse", response))
	return response, nil
}

func (s *nodeServer) DeleteData(ctx context.Context, request *DeleteDataRequest) (*DeleteDataResponse, error) {
	log.Debugw("DeleteData",
		logging.Stringer("DeleteDataRequest", request))
	if request.ShardID == 0 {
		err := errors.NewInvalid("shard_id is required")
		log.Warnw("DeleteData",
			logging.Stringer("DeleteDataRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.ReplicaID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("DeleteData",
			logging.Stringer("DeleteDataRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.DeleteData(ctx, request.ShardID, request.ReplicaID); err != nil {
		log.Warnw("DeleteData",
			logging.Stringer("DeleteDataRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &DeleteDataResponse{}
	log.Debugw("DeleteData",
		logging.Stringer("DeleteDataRequest", request),
		logging.Stringer("DeleteDataResponse", response))
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
