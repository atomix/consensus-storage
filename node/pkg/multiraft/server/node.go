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

func newNodeServer(node *NodeManager) multiraftv1.NodeServer {
	return &NodeServer{
		node: node,
	}
}

type NodeServer struct {
	node *NodeManager
}

func (s *NodeServer) Bootstrap(ctx context.Context, request *multiraftv1.BootstrapRequest) (*multiraftv1.BootstrapResponse, error) {
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

func (s *NodeServer) Join(ctx context.Context, request *multiraftv1.JoinRequest) (*multiraftv1.JoinResponse, error) {
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

func (s *NodeServer) Leave(ctx context.Context, request *multiraftv1.LeaveRequest) (*multiraftv1.LeaveResponse, error) {
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

/*

   MemberReadyEvent member_ready = 2;
   LeaderUpdatedEvent leader_updated = 3;
   MembershipChangedEvent membership_changed = 4;
   SendSnapshotStartedEvent send_snapshot_started = 5;
   SendSnapshotCompletedEvent send_snapshot_completed = 6;
   SendSnapshotAbortedEvent send_snapshot_aborted = 7;
   SnapshotReceivedEvent snapshot_received = 8;
   SnapshotRecoveredEvent snapshot_recovered = 9;
   SnapshotCreatedEvent snapshot_created = 10;
   SnapshotCompactedEvent snapshot_compacted = 11;
   LogCompactedEvent log_compacted = 12;
   LogDBCompactedEvent logdb_compacted = 13;
   ConnectionEstablishedEvent connection_established = 14;
   ConnectionFailedEvent connection_failed = 15;
*/
func (s *NodeServer) Watch(request *multiraftv1.WatchNodeRequest, server multiraftv1.Node_WatchServer) error {
	ch := make(chan multiraftv1.MultiRaftEvent)
	s.node.Watch(server.Context(), ch)
	for event := range ch {
		response := &multiraftv1.WatchNodeResponse{
			Timestamp: event.Timestamp,
		}
		switch t := event.Event.(type) {
		case *multiraftv1.MultiRaftEvent_MemberReady:
			response.Event = &multiraftv1.WatchNodeResponse_MemberReady{
				MemberReady: t.MemberReady,
			}
		case *multiraftv1.MultiRaftEvent_LeaderUpdated:
			response.Event = &multiraftv1.WatchNodeResponse_LeaderUpdated{
				LeaderUpdated: t.LeaderUpdated,
			}
		case *multiraftv1.MultiRaftEvent_MembershipChanged:
			response.Event = &multiraftv1.WatchNodeResponse_MembershipChanged{
				MembershipChanged: t.MembershipChanged,
			}
		case *multiraftv1.MultiRaftEvent_SendSnapshotStarted:
			response.Event = &multiraftv1.WatchNodeResponse_SendSnapshotStarted{
				SendSnapshotStarted: t.SendSnapshotStarted,
			}
		case *multiraftv1.MultiRaftEvent_SendSnapshotCompleted:
			response.Event = &multiraftv1.WatchNodeResponse_SendSnapshotCompleted{
				SendSnapshotCompleted: t.SendSnapshotCompleted,
			}
		case *multiraftv1.MultiRaftEvent_SendSnapshotAborted:
			response.Event = &multiraftv1.WatchNodeResponse_SendSnapshotAborted{
				SendSnapshotAborted: t.SendSnapshotAborted,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotReceived:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotReceived{
				SnapshotReceived: t.SnapshotReceived,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotRecovered:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotRecovered{
				SnapshotRecovered: t.SnapshotRecovered,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotCreated:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotCreated{
				SnapshotCreated: t.SnapshotCreated,
			}
		case *multiraftv1.MultiRaftEvent_SnapshotCompacted:
			response.Event = &multiraftv1.WatchNodeResponse_SnapshotCompacted{
				SnapshotCompacted: t.SnapshotCompacted,
			}
		case *multiraftv1.MultiRaftEvent_LogCompacted:
			response.Event = &multiraftv1.WatchNodeResponse_LogCompacted{
				LogCompacted: t.LogCompacted,
			}
		case *multiraftv1.MultiRaftEvent_LogdbCompacted:
			response.Event = &multiraftv1.WatchNodeResponse_LogdbCompacted{
				LogdbCompacted: t.LogdbCompacted,
			}
		case *multiraftv1.MultiRaftEvent_ConnectionEstablished:
			response.Event = &multiraftv1.WatchNodeResponse_ConnectionEstablished{
				ConnectionEstablished: t.ConnectionEstablished,
			}
		case *multiraftv1.MultiRaftEvent_ConnectionFailed:
			response.Event = &multiraftv1.WatchNodeResponse_ConnectionFailed{
				ConnectionFailed: t.ConnectionFailed,
			}
		}
		err := server.Send(response)
		if err != nil {
			return err
		}
	}
	return nil
}
