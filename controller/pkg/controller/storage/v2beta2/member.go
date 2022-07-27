// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/cenkalti/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
	"time"

	storagev2beta2 "github.com/atomix/multi-raft-storage/controller/pkg/apis/storage/v2beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func addRaftMemberController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &RaftMemberReconciler{
			client:  mgr.GetClient(),
			scheme:  mgr.GetScheme(),
			events:  mgr.GetEventRecorderFor("atomix-multi-raft-storage"),
			streams: make(map[string]func()),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-raft-member-v2beta2", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue RaftMembers that reference it
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftMember{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = controller.Watch(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		members := &storagev2beta2.RaftMemberList{}
		if err := mgr.GetClient().List(context.TODO(), members, &client.ListOptions{Namespace: object.GetNamespace()}); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, member := range members.Items {
			podName := fmt.Sprintf("%s-%d", member.Spec.Cluster, member.Spec.NodeID-1)
			if object.GetName() == podName {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: member.Namespace,
						Name:      member.Name,
					},
				})
			}
		}
		return requests
	}))
	if err != nil {
		return err
	}
	return nil
}

// RaftMemberReconciler reconciles a RaftMember object
type RaftMemberReconciler struct {
	client  client.Client
	scheme  *runtime.Scheme
	events  record.EventRecorder
	streams map[string]func()
	mu      sync.Mutex
}

// Reconcile reads that state of the cluster for a Cluster object and makes changes based on the state read
// and what is in the Cluster.Spec
func (r *RaftMemberReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile RaftMember")
	member := &storagev2beta2.RaftMember{}
	err := r.client.Get(ctx, request.NamespacedName, member)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ok, err := r.reconcileStatus(ctx, member)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *RaftMemberReconciler) reconcileStatus(ctx context.Context, member *storagev2beta2.RaftMember) (bool, error) {
	pod := &corev1.Pod{}
	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      fmt.Sprintf("%s-%d", member.Spec.Cluster, member.Spec.NodeID-1),
	}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		return false, err
	}

	var state storagev2beta2.RaftMemberState
	switch pod.Status.Phase {
	case corev1.PodPending, corev1.PodSucceeded, corev1.PodFailed:
		state = storagev2beta2.RaftMemberStopped
	case corev1.PodRunning:
		state = storagev2beta2.RaftMemberRunning
	}

	if member.Status.State != storagev2beta2.RaftMemberReady && member.Status.State != state {
		member.Status.State = state
		if err := r.client.Status().Update(ctx, member); err != nil {
			return false, err
		}
		return true, nil
	}

	err := r.startMonitoringPod(ctx, member)
	if err != nil {
		return false, err
	}
	return false, nil
}

// nolint:gocyclo
func (r *RaftMemberReconciler) startMonitoringPod(ctx context.Context, member *storagev2beta2.RaftMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.streams[member.Name]
	if ok {
		return nil
	}

	pod := &corev1.Pod{}
	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      fmt.Sprintf("%s-%d", member.Spec.Cluster, member.Spec.NodeID-1),
	}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		return err
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", pod.Status.PodIP, monitoringPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client := multiraftv1.NewNodeClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	request := &multiraftv1.WatchRequest{}
	stream, err := client.Watch(ctx, request)
	if err != nil {
		cancel()
		return err
	}

	r.streams[member.Name] = cancel
	go func() {
		defer func() {
			cancel()
			r.mu.Lock()
			delete(r.streams, member.Name)
			r.mu.Unlock()
			go func() {
				err := r.startMonitoringPod(ctx, member)
				if err != nil {
					log.Error(err)
				}
			}()
		}()

		memberName := types.NamespacedName{
			Namespace: member.Namespace,
			Name:      member.Name,
		}

		member, err := r.getMember(ctx, memberName)
		if err != nil {
			return
		}

		for {
			event, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}

			log.Infof("Received event %+v from %s", event, member.Name)

			switch e := event.Event.(type) {
			case *multiraftv1.Event_MemberReady:
				if e.MemberReady.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordPartitionReady(ctx, memberName, e.MemberReady, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_LeaderUpdated:
				if e.LeaderUpdated.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordLeaderUpdated(ctx, memberName, member.Spec.NodeID, e.LeaderUpdated, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_MembershipChanged:
				if e.MembershipChanged.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordMembershipChanged(ctx, memberName, e.MembershipChanged, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SendSnapshotStarted:
				if e.SendSnapshotStarted.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSendSnapshotStarted(ctx, memberName, e.SendSnapshotStarted, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SendSnapshotCompleted:
				if e.SendSnapshotCompleted.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSendSnapshotCompleted(ctx, memberName, e.SendSnapshotCompleted, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SendSnapshotAborted:
				if e.SendSnapshotAborted.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSendSnapshotAborted(ctx, memberName, e.SendSnapshotAborted, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SnapshotReceived:
				if e.SnapshotReceived.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSnapshotReceived(ctx, memberName, e.SnapshotReceived, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SnapshotRecovered:
				if e.SnapshotRecovered.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSnapshotRecovered(ctx, memberName, e.SnapshotRecovered, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SnapshotCreated:
				if e.SnapshotCreated.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSnapshotCreated(ctx, memberName, e.SnapshotCreated, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_SnapshotCompacted:
				if e.SnapshotCompacted.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordSnapshotCompacted(ctx, memberName, e.SnapshotCompacted, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_LogCompacted:
				if e.LogCompacted.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordLogCompacted(ctx, memberName, e.LogCompacted, metav1.NewTime(event.Timestamp))
				}
			case *multiraftv1.Event_LogdbCompacted:
				if e.LogdbCompacted.GroupID == multiraftv1.GroupID(member.Spec.GroupID) {
					r.recordLogDBCompacted(ctx, memberName, e.LogdbCompacted, metav1.NewTime(event.Timestamp))
				}
			}
		}
	}()
	return nil
}

func (r *RaftMemberReconciler) getMember(ctx context.Context, memberName types.NamespacedName) (*storagev2beta2.RaftMember, error) {
	member := &storagev2beta2.RaftMember{}
	if err := r.client.Get(ctx, memberName, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (r *RaftMemberReconciler) recordPartitionReady(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.MemberReadyEvent, timestamp metav1.Time) {
	err := r.updateMemberStatus(ctx, memberName, storagev2beta2.RaftMemberStatus{
		State:       storagev2beta2.RaftMemberReady,
		LastUpdated: &timestamp,
	})
	if err != nil {
		log.Error(err)
	}
}

func (r *RaftMemberReconciler) recordLeaderUpdated(ctx context.Context, memberName types.NamespacedName, nodeID int32, event *multiraftv1.LeaderUpdatedEvent, timestamp metav1.Time) {
	role := storagev2beta2.RaftFollower
	if event.Leader == multiraftv1.NodeID(nodeID) {
		role = storagev2beta2.RaftLeader
	}
	term := uint64(event.Term)
	var leader *uint32
	if event.Leader != 0 {
		l := uint32(event.Leader)
		leader = &l
	}
	err := r.updateMemberStatus(ctx, memberName, storagev2beta2.RaftMemberStatus{
		Role:        &role,
		Term:        &term,
		Leader:      leader,
		LastUpdated: &timestamp,
	})
	if err != nil {
		log.Error(err)
	}
}

func (r *RaftMemberReconciler) recordMembershipChanged(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.MembershipChangedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "MembershipChanged", "Membership changed")
	})
}

func (r *RaftMemberReconciler) recordSendSnapshotStarted(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SendSnapshotStartedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "SendSnapshotStared", "Started sending snapshot at index %d to %s", event.Index, event.To)
	})
}

func (r *RaftMemberReconciler) recordSendSnapshotCompleted(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SendSnapshotCompletedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "SendSnapshotCompleted", "Completed sending snapshot at index %d to %s", event.Index, event.To)
	})
}

func (r *RaftMemberReconciler) recordSendSnapshotAborted(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SendSnapshotAbortedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Warning", "SendSnapshotAborted", "Aborted sending snapshot at index %d to %s", event.Index, event.To)
	})
}

func (r *RaftMemberReconciler) recordSnapshotReceived(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SnapshotReceivedEvent, timestamp metav1.Time) {
	index := uint64(event.Index)
	err := r.updateMemberStatus(ctx, memberName, storagev2beta2.RaftMemberStatus{
		LastSnapshotIndex: &index,
		LastSnapshotTime:  &timestamp,
		LastUpdated:       &timestamp,
	})
	if err != nil {
		log.Error(err)
	}
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "SnapshotReceived", "Received snapshot at index %d from %s", event.Index, event.From)
	})
}

func (r *RaftMemberReconciler) recordSnapshotRecovered(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SnapshotRecoveredEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "SnapshotRecovered", "Recovered from snapshot at index %d", event.Index)
	})
}

func (r *RaftMemberReconciler) recordSnapshotCreated(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SnapshotCreatedEvent, timestamp metav1.Time) {
	index := uint64(event.Index)
	err := r.updateMemberStatus(ctx, memberName, storagev2beta2.RaftMemberStatus{
		LastSnapshotIndex: &index,
		LastSnapshotTime:  &timestamp,
		LastUpdated:       &timestamp,
	})
	if err != nil {
		log.Error(err)
	}
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "SnapshotCreated", "Created snapshot at index %d", event.Index)
	})
}

func (r *RaftMemberReconciler) recordSnapshotCompacted(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.SnapshotCompactedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "SnapshotCompacted", "Compacted snapshot at index %d", event.Index)
	})
}

func (r *RaftMemberReconciler) recordLogCompacted(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.LogCompactedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "LogCompacted", "Log compacted at index %d", event.Index)
	})
}

func (r *RaftMemberReconciler) recordLogDBCompacted(ctx context.Context, memberName types.NamespacedName, event *multiraftv1.LogDBCompactedEvent, timestamp metav1.Time) {
	r.recordEvent(ctx, memberName, func(member *storagev2beta2.RaftMember) {
		r.events.Eventf(member, "Normal", "LogCompacted", "Log compacted at index %d", event.Index)
		r.events.Eventf(member, "Normal", "LogDBCompacted", "LogDB compacted at index %d", event.Index)
	})
}

func (r *RaftMemberReconciler) updateMemberStatus(ctx context.Context, memberName types.NamespacedName, status storagev2beta2.RaftMemberStatus) error {
	return backoff.Retry(func() error {
		return r.tryUpdateMemberStatus(ctx, memberName, status)
	}, backoff.NewExponentialBackOff())
}

func (r *RaftMemberReconciler) tryUpdateMemberStatus(ctx context.Context, memberName types.NamespacedName, status storagev2beta2.RaftMemberStatus) error {
	member, err := r.getMember(ctx, memberName)
	if err != nil {
		return err
	}

	updated := true
	var recorders []func()
	if status.Term != nil && (member.Status.Term == nil || *status.Term > *member.Status.Term) {
		member.Status.Term = status.Term
		member.Status.Leader = nil
		recorders = append(recorders, func() {
			r.events.Eventf(member, "Normal", "TermChanged", "Term changed to %d", status.Term)
		})
		updated = true
	}
	if status.Leader != nil && (member.Status.Leader == nil || member.Status.Leader != status.Leader) {
		member.Status.Leader = status.Leader
		recorders = append(recorders, func() {
			r.events.Eventf(member, "Normal", "LeaderChanged", "Leader changed to %d", status.Leader)
		})
		updated = true
	}
	if member.Status.State != status.State {
		member.Status.State = status.State
		recorders = append(recorders, func() {
			r.events.Eventf(member, "Normal", "StateChanged", "State changed to %s", status.State)
		})
		updated = true
	}
	if status.Role != nil && (member.Status.Role == nil || *member.Status.Role != *status.Role) {
		member.Status.Role = status.Role
		recorders = append(recorders, func() {
			r.events.Eventf(member, "Normal", "RoleChanged", "Role changed to %s", status.Role)
		})
		updated = true
	}
	if status.LastUpdated != nil && (member.Status.LastUpdated == nil || status.LastUpdated.After(member.Status.LastUpdated.Time)) {
		member.Status.LastUpdated = status.LastUpdated
		updated = true
	}
	if status.LastSnapshotIndex != nil && (member.Status.LastSnapshotIndex == nil || *status.LastSnapshotIndex > *member.Status.LastSnapshotIndex) {
		member.Status.LastSnapshotIndex = status.LastSnapshotIndex
		updated = true
	}
	if status.LastSnapshotTime != nil && (member.Status.LastSnapshotTime == nil || status.LastSnapshotTime.After(member.Status.LastSnapshotTime.Time)) {
		member.Status.LastSnapshotTime = status.LastSnapshotTime
		recorders = append(recorders, func() {
			r.events.Eventf(member, "Normal", "SnapshotTaken", "Snapshot taken")
		})
		updated = true
	}

	if updated {
		if err := r.client.Status().Update(ctx, member); err != nil {
			return err
		}
		for _, recorder := range recorders {
			recorder()
		}
	}
	return nil
}

func (r *RaftMemberReconciler) recordEvent(ctx context.Context, memberName types.NamespacedName, f func(member *storagev2beta2.RaftMember)) {
	_ = backoff.Retry(func() error {
		return r.tryRecordEvent(ctx, memberName, f)
	}, backoff.NewExponentialBackOff())
}

func (r *RaftMemberReconciler) tryRecordEvent(ctx context.Context, memberName types.NamespacedName, f func(member *storagev2beta2.RaftMember)) error {
	member, err := r.getMember(ctx, memberName)
	if err != nil {
		return err
	}
	f(member)
	return nil
}

var _ reconcile.Reconciler = (*RaftMemberReconciler)(nil)
