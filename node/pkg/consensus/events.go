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
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
				Term:   Term(info.Term),
				Leader: MemberID(info.LeaderID),
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
		Event: &Event_MemberReady{
			MemberReady: &MemberReadyEvent{
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
			},
		},
	})
}

func (e *eventListener) MembershipChanged(info raftio.NodeInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_MembershipChanged{
			MembershipChanged: &MembershipChangedEvent{
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
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
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
				Index: Index(info.Index),
				To:    MemberID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotCompleted(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SendSnapshotCompleted{
			SendSnapshotCompleted: &SendSnapshotCompletedEvent{
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
				Index: Index(info.Index),
				To:    MemberID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotAborted(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SendSnapshotAborted{
			SendSnapshotAborted: &SendSnapshotAbortedEvent{
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
				Index: Index(info.Index),
				To:    MemberID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SnapshotReceived(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SnapshotReceived{
			SnapshotReceived: &SnapshotReceivedEvent{
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
				Index: Index(info.Index),
				From:  MemberID(info.From),
			},
		},
	})
}

func (e *eventListener) SnapshotRecovered(info raftio.SnapshotInfo) {
	e.publish(&Event{
		Timestamp: time.Now(),
		Event: &Event_SnapshotRecovered{
			SnapshotRecovered: &SnapshotRecoveredEvent{
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
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
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
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
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
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
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
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
				MemberEvent: MemberEvent{
					GroupID:  GroupID(info.ClusterID),
					MemberID: MemberID(info.NodeID),
				},
				Index: Index(info.Index),
			},
		},
	})
}
