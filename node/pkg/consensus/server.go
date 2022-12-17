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
	group, err := s.protocol.GetConfig(request.GroupID)
	if err != nil {
		log.Warnw("GetConfig",
			logging.Stringer("GetConfigRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &GetConfigResponse{
		Group: group,
	}
	log.Debugw("GetConfig",
		logging.Stringer("GetConfigRequest", request),
		logging.Stringer("GetConfigResponse", response))
	return response, nil
}

func (s *nodeServer) BootstrapGroup(ctx context.Context, request *BootstrapGroupRequest) (*BootstrapGroupResponse, error) {
	log.Debugw("BootstrapGroup",
		logging.Stringer("BootstrapGroupRequest", request))
	if request.GroupID == 0 {
		err := errors.NewInvalid("group_id is required")
		log.Warnw("BootstrapGroup",
			logging.Stringer("BootstrapGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.MemberID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("BootstrapGroup",
			logging.Stringer("BootstrapGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.BootstrapGroup(ctx, request.GroupID, request.MemberID, request.Members...); err != nil {
		log.Warnw("BootstrapGroup",
			logging.Stringer("BootstrapGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &BootstrapGroupResponse{}
	log.Debugw("BootstrapGroup",
		logging.Stringer("BootstrapGroupRequest", request),
		logging.Stringer("BootstrapGroupResponse", response))
	return response, nil
}

func (s *nodeServer) AddMember(ctx context.Context, request *AddMemberRequest) (*AddMemberResponse, error) {
	log.Debugw("AddMember",
		logging.Stringer("AddMemberRequest", request))
	if request.GroupID == 0 {
		err := errors.NewInvalid("group_id is required")
		log.Warnw("AddMember",
			logging.Stringer("AddMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.Version == 0 {
		err := errors.NewInvalid("version is required")
		log.Warnw("AddMember",
			logging.Stringer("AddMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.AddMember(ctx, request.GroupID, request.Member, request.Version); err != nil {
		log.Warnw("AddMember",
			logging.Stringer("AddMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &AddMemberResponse{}
	log.Debugw("AddMember",
		logging.Stringer("AddMemberRequest", request),
		logging.Stringer("AddMemberResponse", response))
	return response, nil
}

func (s *nodeServer) RemoveMember(ctx context.Context, request *RemoveMemberRequest) (*RemoveMemberResponse, error) {
	log.Debugw("RemoveMember",
		logging.Stringer("RemoveMemberRequest", request))
	if request.GroupID == 0 {
		err := errors.NewInvalid("group_id is required")
		log.Warnw("RemoveMember",
			logging.Stringer("RemoveMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.MemberID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("RemoveMember",
			logging.Stringer("RemoveMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.Version == 0 {
		err := errors.NewInvalid("version is required")
		log.Warnw("RemoveMember",
			logging.Stringer("RemoveMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.RemoveMember(ctx, request.GroupID, request.MemberID, request.Version); err != nil {
		log.Warnw("RemoveMember",
			logging.Stringer("RemoveMemberRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &RemoveMemberResponse{}
	log.Debugw("RemoveMember",
		logging.Stringer("RemoveMemberRequest", request),
		logging.Stringer("RemoveMemberResponse", response))
	return response, nil
}

func (s *nodeServer) JoinGroup(ctx context.Context, request *JoinGroupRequest) (*JoinGroupResponse, error) {
	log.Debugw("JoinGroup",
		logging.Stringer("JoinGroupRequest", request))
	if request.GroupID == 0 {
		err := errors.NewInvalid("group_id is required")
		log.Warnw("JoinGroup",
			logging.Stringer("JoinGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.MemberID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("JoinGroup",
			logging.Stringer("JoinGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.JoinGroup(ctx, request.GroupID, request.MemberID); err != nil {
		log.Warnw("JoinGroup",
			logging.Stringer("JoinGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &JoinGroupResponse{}
	log.Debugw("JoinGroup",
		logging.Stringer("JoinGroupRequest", request),
		logging.Stringer("JoinGroupResponse", response))
	return response, nil
}

func (s *nodeServer) LeaveGroup(ctx context.Context, request *LeaveGroupRequest) (*LeaveGroupResponse, error) {
	log.Debugw("LeaveGroup",
		logging.Stringer("LeaveGroupRequest", request))
	if request.GroupID == 0 {
		err := errors.NewInvalid("group_id is required")
		log.Warnw("LeaveGroup",
			logging.Stringer("LeaveGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.LeaveGroup(ctx, request.GroupID); err != nil {
		log.Warnw("LeaveGroup",
			logging.Stringer("LeaveGroupRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	response := &LeaveGroupResponse{}
	log.Debugw("LeaveGroup",
		logging.Stringer("LeaveGroupRequest", request),
		logging.Stringer("LeaveGroupResponse", response))
	return response, nil
}

func (s *nodeServer) DeleteData(ctx context.Context, request *DeleteDataRequest) (*DeleteDataResponse, error) {
	log.Debugw("DeleteData",
		logging.Stringer("DeleteDataRequest", request))
	if request.GroupID == 0 {
		err := errors.NewInvalid("group_id is required")
		log.Warnw("DeleteData",
			logging.Stringer("DeleteDataRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if request.MemberID == 0 {
		err := errors.NewInvalid("member_id is required")
		log.Warnw("DeleteData",
			logging.Stringer("DeleteDataRequest", request),
			logging.Error("Error", err))
		return nil, errors.ToProto(err)
	}
	if err := s.protocol.DeleteData(ctx, request.GroupID, request.MemberID); err != nil {
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
