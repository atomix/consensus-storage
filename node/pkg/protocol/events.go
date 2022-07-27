// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/lni/dragonboat/v3/raftio"
	"time"
)

func newEventListener(node *Node) *eventListener {
	return &eventListener{
		node: node,
	}
}

type eventListener struct {
	node *Node
}

func (e *eventListener) publish(event *multiraftv1.Event) {
	e.node.publish(event)
}

func (e *eventListener) LeaderUpdated(info raftio.LeaderInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_LeaderUpdated{
			LeaderUpdated: &multiraftv1.LeaderUpdatedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Term:    multiraftv1.Term(info.Term),
				Leader:  multiraftv1.NodeID(info.LeaderID),
			},
		},
	})
}

func (e *eventListener) NodeHostShuttingDown() {

}

func (e *eventListener) NodeUnloaded(info raftio.NodeInfo) {

}

func (e *eventListener) NodeReady(info raftio.NodeInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_MemberReady{
			MemberReady: &multiraftv1.MemberReadyEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				NodeID:  multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) MembershipChanged(info raftio.NodeInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_MembershipChanged{
			MembershipChanged: &multiraftv1.MembershipChangedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				NodeID:  multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) ConnectionEstablished(info raftio.ConnectionInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_ConnectionEstablished{
			ConnectionEstablished: &multiraftv1.ConnectionEstablishedEvent{
				ConnectionInfo: multiraftv1.ConnectionInfo{
					Address:  info.Address,
					Snapshot: info.SnapshotConnection,
				},
			},
		},
	})
}

func (e *eventListener) ConnectionFailed(info raftio.ConnectionInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_ConnectionFailed{
			ConnectionFailed: &multiraftv1.ConnectionFailedEvent{
				ConnectionInfo: multiraftv1.ConnectionInfo{
					Address:  info.Address,
					Snapshot: info.SnapshotConnection,
				},
			},
		},
	})
}

func (e *eventListener) SendSnapshotStarted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SendSnapshotStarted{
			SendSnapshotStarted: &multiraftv1.SendSnapshotStartedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
				To:      multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotCompleted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SendSnapshotCompleted{
			SendSnapshotCompleted: &multiraftv1.SendSnapshotCompletedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
				To:      multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotAborted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SendSnapshotAborted{
			SendSnapshotAborted: &multiraftv1.SendSnapshotAbortedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
				To:      multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SnapshotReceived(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SnapshotReceived{
			SnapshotReceived: &multiraftv1.SnapshotReceivedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
				From:    multiraftv1.NodeID(info.From),
			},
		},
	})
}

func (e *eventListener) SnapshotRecovered(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SnapshotRecovered{
			SnapshotRecovered: &multiraftv1.SnapshotRecoveredEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) SnapshotCreated(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SnapshotCreated{
			SnapshotCreated: &multiraftv1.SnapshotCreatedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) SnapshotCompacted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_SnapshotCompacted{
			SnapshotCompacted: &multiraftv1.SnapshotCompactedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) LogCompacted(info raftio.EntryInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_LogCompacted{
			LogCompacted: &multiraftv1.LogCompactedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) LogDBCompacted(info raftio.EntryInfo) {
	e.publish(&multiraftv1.Event{
		Timestamp: time.Now(),
		Event: &multiraftv1.Event_LogdbCompacted{
			LogdbCompacted: &multiraftv1.LogDBCompactedEvent{
				GroupID: multiraftv1.GroupID(info.ClusterID),
				Index:   multiraftv1.Index(info.Index),
			},
		},
	})
}
