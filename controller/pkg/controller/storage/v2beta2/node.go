// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	"context"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
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

func addMultiRaftNodeController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &MultiRaftNodeReconciler{
			client:  mgr.GetClient(),
			scheme:  mgr.GetScheme(),
			events:  mgr.GetEventRecorderFor("atomix-multi-raft-storage"),
			streams: make(map[string]func()),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-multi-raft-node-v2beta2", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue MultiRaftNodes that reference it
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.MultiRaftNode{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = controller.Watch(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		nodeName := types.NamespacedName{
			Namespace: object.GetNamespace(),
			Name:      object.GetName(),
		}
		var requests []reconcile.Request
		node := &storagev2beta2.MultiRaftNode{}
		if err := mgr.GetClient().Get(context.TODO(), nodeName, node); err == nil {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: node.Namespace,
					Name:      node.Name,
				},
			})
		}
		return requests
	}))
	if err != nil {
		return err
	}
	return nil
}

// MultiRaftNodeReconciler reconciles a MultiRaftNode object
type MultiRaftNodeReconciler struct {
	client  client.Client
	scheme  *runtime.Scheme
	events  record.EventRecorder
	streams map[string]func()
	mu      sync.Mutex
}

// Reconcile reads that state of the cluster for a Cluster object and makes changes based on the state read
// and what is in the Cluster.Spec
func (r *MultiRaftNodeReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile MultiRaftNode")
	member := &storagev2beta2.MultiRaftNode{}
	err := r.client.Get(ctx, request.NamespacedName, member)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftNode")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ok, err := r.reconcileStatus(ctx, member)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftNode")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *MultiRaftNodeReconciler) reconcileStatus(ctx context.Context, node *storagev2beta2.MultiRaftNode) (bool, error) {
	podName := types.NamespacedName{
		Namespace: node.Namespace,
		Name:      fmt.Sprintf("%s-%d", node.Spec.Cluster, node.Spec.NodeID-1),
	}
	pod := &corev1.Pod{}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		return false, err
	}

	state := storagev2beta2.MultiRaftNodeReady
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			if condition.Status != corev1.ConditionTrue {
				state = storagev2beta2.MultiRaftNodeNotReady
			}
			break
		}
	}
	log.Warnf("State for %s is %s", node.Name, state)

	if state == storagev2beta2.MultiRaftNodeNotReady && node.Status.State != storagev2beta2.MultiRaftNodeNotReady {
		node.Status.State = storagev2beta2.MultiRaftNodeNotReady
		if err := r.client.Status().Update(ctx, node); err != nil {
			return false, err
		}
		r.events.Eventf(node, "Normal", "NotReady", "MultiRaftNode is not ready")
		return true, nil
	}

	if state == storagev2beta2.MultiRaftNodeReady && node.Status.State != storagev2beta2.MultiRaftNodeReady {
		node.Status.State = storagev2beta2.MultiRaftNodeReady
		if err := r.client.Status().Update(ctx, node); err != nil {
			return false, err
		}
		r.events.Eventf(node, "Normal", "Ready", "MultiRaftNode is ready")
		return true, nil
	}

	go func() {
		err := r.startMonitoringPod(ctx, node)
		if err != nil {
			log.Error(err)
		}
	}()
	return false, nil
}

// nolint:gocyclo
func (r *MultiRaftNodeReconciler) startMonitoringPod(ctx context.Context, node *storagev2beta2.MultiRaftNode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.streams[node.Name]
	if ok {
		return nil
	}

	pod := &corev1.Pod{}
	podName := types.NamespacedName{
		Namespace: node.Namespace,
		Name:      fmt.Sprintf("%s-%d", node.Spec.Cluster, node.Spec.NodeID-1),
	}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		return err
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort),
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

	r.streams[node.Name] = cancel
	go func() {
		defer func() {
			cancel()
			r.mu.Lock()
			delete(r.streams, node.Name)
			r.mu.Unlock()
			go func() {
				err := r.startMonitoringPod(ctx, node)
				if err != nil {
					log.Error(err)
				}
			}()
		}()

		nodeName := types.NamespacedName{
			Namespace: node.Namespace,
			Name:      node.Name,
		}

		for {
			event, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}

			log.Infof("Received event %+v from %s", event, node.Name)

			switch e := event.Event.(type) {
			case *multiraftv1.Event_ConnectionEstablished:
				r.recordConnectionEstablished(ctx, nodeName, e.ConnectionEstablished, metav1.NewTime(event.Timestamp))
			case *multiraftv1.Event_ConnectionFailed:
				r.recordConnectionFailed(ctx, nodeName, e.ConnectionFailed, metav1.NewTime(event.Timestamp))
			}
		}
	}()
	return nil
}

func (r *MultiRaftNodeReconciler) getNode(ctx context.Context, nodeName types.NamespacedName) (*storagev2beta2.MultiRaftNode, error) {
	node := &storagev2beta2.MultiRaftNode{}
	if err := r.client.Get(ctx, nodeName, node); err != nil {
		return nil, err
	}
	return node, nil
}

func (r *MultiRaftNodeReconciler) recordConnectionEstablished(ctx context.Context, nodeName types.NamespacedName, event *multiraftv1.ConnectionEstablishedEvent, timestamp metav1.Time) {
	node, err := r.getNode(ctx, nodeName)
	if err != nil {
		log.Error(err)
	} else {
		r.events.Eventf(node, "Normal", "ConnectionEstablished", "Connection established to %s", event.Address)
	}
}

func (r *MultiRaftNodeReconciler) recordConnectionFailed(ctx context.Context, nodeName types.NamespacedName, event *multiraftv1.ConnectionFailedEvent, timestamp metav1.Time) {
	node, err := r.getNode(ctx, nodeName)
	if err != nil {
		log.Error(err)
	} else {
		r.events.Eventf(node, "Warning", "ConnectionFailed", "Connection to %s failed", event.Address)
	}
}

var _ reconcile.Reconciler = (*MultiRaftNodeReconciler)(nil)
