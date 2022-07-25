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

func (e *eventListener) publishNode(event *multiraftv1.NodeEvent) {
	e.node.publish(event)
}

func (e *eventListener) publishPartition(event *multiraftv1.PartitionEvent) {
	if partition, ok := e.node.Partition(event.PartitionID); ok {
		partition.publish(event)
	}
}

func (e *eventListener) LeaderUpdated(info raftio.LeaderInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_LeaderUpdated{
			LeaderUpdated: &multiraftv1.LeaderUpdatedEvent{
				Term:   multiraftv1.Term(info.Term),
				Leader: multiraftv1.NodeID(info.LeaderID),
			},
		},
	})
}

func (e *eventListener) NodeHostShuttingDown() {

}

func (e *eventListener) NodeUnloaded(info raftio.NodeInfo) {

}

func (e *eventListener) NodeReady(info raftio.NodeInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_MemberReady{
			MemberReady: &multiraftv1.MemberReadyEvent{
				NodeID: multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) MembershipChanged(info raftio.NodeInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_MembershipChanged{
			MembershipChanged: &multiraftv1.MembershipChangedEvent{
				NodeID: multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) ConnectionEstablished(info raftio.ConnectionInfo) {
	e.publishNode(&multiraftv1.NodeEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.NodeEvent_ConnectionEstablished{
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
	e.publishNode(&multiraftv1.NodeEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.NodeEvent_ConnectionFailed{
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
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SendSnapshotStarted{
			SendSnapshotStarted: &multiraftv1.SendSnapshotStartedEvent{
				Index: multiraftv1.Index(info.Index),
				To:    multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotCompleted(info raftio.SnapshotInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SendSnapshotCompleted{
			SendSnapshotCompleted: &multiraftv1.SendSnapshotCompletedEvent{
				Index: multiraftv1.Index(info.Index),
				To:    multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotAborted(info raftio.SnapshotInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SendSnapshotAborted{
			SendSnapshotAborted: &multiraftv1.SendSnapshotAbortedEvent{
				Index: multiraftv1.Index(info.Index),
				To:    multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SnapshotReceived(info raftio.SnapshotInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SnapshotReceived{
			SnapshotReceived: &multiraftv1.SnapshotReceivedEvent{
				Index: multiraftv1.Index(info.Index),
				From:  multiraftv1.NodeID(info.From),
			},
		},
	})
}

func (e *eventListener) SnapshotRecovered(info raftio.SnapshotInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SnapshotRecovered{
			SnapshotRecovered: &multiraftv1.SnapshotRecoveredEvent{
				Index: multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) SnapshotCreated(info raftio.SnapshotInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SnapshotCreated{
			SnapshotCreated: &multiraftv1.SnapshotCreatedEvent{
				Index: multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) SnapshotCompacted(info raftio.SnapshotInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_SnapshotCompacted{
			SnapshotCompacted: &multiraftv1.SnapshotCompactedEvent{
				Index: multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) LogCompacted(info raftio.EntryInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_LogCompacted{
			LogCompacted: &multiraftv1.LogCompactedEvent{
				Index: multiraftv1.Index(info.Index),
			},
		},
	})
}

func (e *eventListener) LogDBCompacted(info raftio.EntryInfo) {
	e.publishPartition(&multiraftv1.PartitionEvent{
		Timestamp:   time.Now(),
		PartitionID: multiraftv1.PartitionID(info.ClusterID),
		Event: &multiraftv1.PartitionEvent_LogdbCompacted{
			LogdbCompacted: &multiraftv1.LogDBCompactedEvent{
				Index: multiraftv1.Index(info.Index),
			},
		},
	})
}
