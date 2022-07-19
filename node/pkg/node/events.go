// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/logging"
	"github.com/lni/dragonboat/v3/raftio"
	"sync"
	"time"
)

func newEventListener() *eventListener {
	return &eventListener{
		listeners: make(map[int]chan<- multiraftv1.MultiRaftEvent),
	}
}

type eventListener struct {
	listeners  map[int]chan<- multiraftv1.MultiRaftEvent
	listenerID int
	mu         sync.RWMutex
}

func (e *eventListener) listen(ctx context.Context, ch chan<- multiraftv1.MultiRaftEvent) {
	e.listenerID++
	id := e.listenerID
	e.mu.Lock()
	e.listeners[id] = ch
	e.mu.Unlock()

	go func() {
		<-ctx.Done()
		e.mu.Lock()
		close(ch)
		delete(e.listeners, id)
		e.mu.Unlock()
	}()
}

func (e *eventListener) publish(event *multiraftv1.MultiRaftEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	log.Infow("Publish MultiRaftEvent",
		logging.Stringer("MultiRaftEvent", event))
	for _, listener := range e.listeners {
		listener <- *event
	}
}

func (e *eventListener) LeaderUpdated(info raftio.LeaderInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_LeaderUpdated{
			LeaderUpdated: &multiraftv1.LeaderUpdatedEvent{
				LeaderEvent: multiraftv1.LeaderEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Term:   multiraftv1.Term(info.Term),
					Leader: multiraftv1.NodeID(info.LeaderID),
				},
			},
		},
	})
}

func (e *eventListener) NodeHostShuttingDown() {

}

func (e *eventListener) NodeUnloaded(info raftio.NodeInfo) {

}

func (e *eventListener) NodeReady(info raftio.NodeInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_MemberReady{
			MemberReady: &multiraftv1.MemberReadyEvent{
				PartitionEvent: multiraftv1.PartitionEvent{
					PartitionID: multiraftv1.PartitionID(info.ClusterID),
				},
			},
		},
	})
}

func (e *eventListener) MembershipChanged(info raftio.NodeInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_MembershipChanged{
			MembershipChanged: &multiraftv1.MembershipChangedEvent{
				PartitionEvent: multiraftv1.PartitionEvent{
					PartitionID: multiraftv1.PartitionID(info.ClusterID),
				},
			},
		},
	})
}

func (e *eventListener) ConnectionEstablished(info raftio.ConnectionInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_ConnectionEstablished{
			ConnectionEstablished: &multiraftv1.ConnectionEstablishedEvent{
				ConnectionEvent: multiraftv1.ConnectionEvent{
					Address:  info.Address,
					Snapshot: info.SnapshotConnection,
				},
			},
		},
	})
}

func (e *eventListener) ConnectionFailed(info raftio.ConnectionInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_ConnectionFailed{
			ConnectionFailed: &multiraftv1.ConnectionFailedEvent{
				ConnectionEvent: multiraftv1.ConnectionEvent{
					Address:  info.Address,
					Snapshot: info.SnapshotConnection,
				},
			},
		},
	})
}

func (e *eventListener) SendSnapshotStarted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SendSnapshotStarted{
			SendSnapshotStarted: &multiraftv1.SendSnapshotStartedEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
				To: multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotCompleted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SendSnapshotCompleted{
			SendSnapshotCompleted: &multiraftv1.SendSnapshotCompletedEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
				To: multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SendSnapshotAborted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SendSnapshotAborted{
			SendSnapshotAborted: &multiraftv1.SendSnapshotAbortedEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
				To: multiraftv1.NodeID(info.NodeID),
			},
		},
	})
}

func (e *eventListener) SnapshotReceived(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SnapshotReceived{
			SnapshotReceived: &multiraftv1.SnapshotReceivedEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
				From: multiraftv1.NodeID(info.From),
			},
		},
	})
}

func (e *eventListener) SnapshotRecovered(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SnapshotRecovered{
			SnapshotRecovered: &multiraftv1.SnapshotRecoveredEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
			},
		},
	})
}

func (e *eventListener) SnapshotCreated(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SnapshotCreated{
			SnapshotCreated: &multiraftv1.SnapshotCreatedEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
			},
		},
	})
}

func (e *eventListener) SnapshotCompacted(info raftio.SnapshotInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_SnapshotCompacted{
			SnapshotCompacted: &multiraftv1.SnapshotCompactedEvent{
				SnapshotEvent: multiraftv1.SnapshotEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
			},
		},
	})
}

func (e *eventListener) LogCompacted(info raftio.EntryInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_LogCompacted{
			LogCompacted: &multiraftv1.LogCompactedEvent{
				LogEvent: multiraftv1.LogEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
			},
		},
	})
}

func (e *eventListener) LogDBCompacted(info raftio.EntryInfo) {
	e.publish(&multiraftv1.MultiRaftEvent{
		Timestamp: time.Now(),
		Event: &multiraftv1.MultiRaftEvent_LogdbCompacted{
			LogdbCompacted: &multiraftv1.LogDBCompactedEvent{
				LogEvent: multiraftv1.LogEvent{
					PartitionEvent: multiraftv1.PartitionEvent{
						PartitionID: multiraftv1.PartitionID(info.ClusterID),
					},
					Index: multiraftv1.Index(info.Index),
				},
			},
		},
	})
}
