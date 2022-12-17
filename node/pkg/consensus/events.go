// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/lni/dragonboat/v3/raftio"
	"time"
)

func newEventListener(protocol *Protocol) *eventListener {
	return &eventListener{
		protocol: protocol,
	}
}

type eventListener struct {
	protocol *Protocol
}

func (e *eventListener) publish(event *Event) {
	e.protocol.publish(event)
}

func (e *eventListener) LeaderUpdated(info raftio.LeaderInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_LeaderUpdated{
			LeaderUpdated: &LeaderUpdatedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Term:   Term(info.Term),
				Leader: ReplicaID(info.LeaderID),
			},
		},
	})
}

func (e *eventListener) NodeHostShuttingDown() {

}

func (e *eventListener) NodeUnloaded(info raftio.NodeInfo) {

}

func (e *eventListener) NodeReady(info raftio.NodeInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_ReplicaReady{
			ReplicaReady: &ReplicaReadyEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
			},
		},
	})
}

func (e *eventListener) MembershipChanged(info raftio.NodeInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_ConfigurationChanged{
			ConfigurationChanged: &ConfigurationChangedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
			},
		},
	})
}

func (e *eventListener) ConnectionEstablished(info raftio.ConnectionInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_ConnectionEstablished{
			ConnectionEstablished: &ConnectionEstablishedEvent{
				ConnectionInfo: ConnectionInfo{
					Address:  info.Address,
					Snapshot: info.SnapshotConnection,
				},
			},
		},
	})
}

func (e *eventListener) ConnectionFailed(info raftio.ConnectionInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_ConnectionFailed{
			ConnectionFailed: &ConnectionFailedEvent{
				ConnectionInfo: ConnectionInfo{
					Address:  info.Address,
					Snapshot: info.SnapshotConnection,
				},
			},
		},
	})
}

func (e *eventListener) SendSnapshotStarted(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SendSnapshotStarted{
			SendSnapshotStarted: &SendSnapshotStartedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
				To:    ReplicaID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotCompleted(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SendSnapshotCompleted{
			SendSnapshotCompleted: &SendSnapshotCompletedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
				To:    ReplicaID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotAborted(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SendSnapshotAborted{
			SendSnapshotAborted: &SendSnapshotAbortedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
				To:    ReplicaID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SnapshotReceived(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SnapshotReceived{
			SnapshotReceived: &SnapshotReceivedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
				From:  ReplicaID(info.From),
			},
		},
	})
}

func (e *eventListener) SnapshotRecovered(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SnapshotRecovered{
			SnapshotRecovered: &SnapshotRecoveredEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
			},
		},
	})
}

func (e *eventListener) SnapshotCreated(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SnapshotCreated{
			SnapshotCreated: &SnapshotCreatedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
			},
		},
	})
}

func (e *eventListener) SnapshotCompacted(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SnapshotCompacted{
			SnapshotCompacted: &SnapshotCompactedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
			},
		},
	})
}

func (e *eventListener) LogCompacted(info raftio.EntryInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_LogCompacted{
			LogCompacted: &LogCompactedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
			},
		},
	})
}

func (e *eventListener) LogDBCompacted(info raftio.EntryInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_LogdbCompacted{
			LogdbCompacted: &LogDBCompactedEvent{
				ReplicaEvent: ReplicaEvent{
					ShardID:   ShardID(info.ClusterID),
					ReplicaID: ReplicaID(info.NodeID),
				},
				Index: Index(info.Index),
			},
		},
	})
}
