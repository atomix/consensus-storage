// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1beta2

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	multiraftv1beta2 "github.com/atomix/consensus-storage/controller/pkg/apis/multiraft/v1beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func addRaftPartitionController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &RaftPartitionReconciler{
			client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			events: mgr.GetEventRecorderFor("atomix-consensus-storage"),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-raft-partition", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Stores that reference it
	err = controller.Watch(&source.Kind{Type: &multiraftv1beta2.RaftPartition{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RaftMember
	err = controller.Watch(&source.Kind{Type: &multiraftv1beta2.RaftMember{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &multiraftv1beta2.RaftPartition{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource MultiRaftCluster
	err = controller.Watch(&source.Kind{Type: &multiraftv1beta2.MultiRaftCluster{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		partitions := &multiraftv1beta2.RaftPartitionList{}
		if err := mgr.GetClient().List(context.Background(), partitions, &client.ListOptions{Namespace: object.GetNamespace()}); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, partition := range partitions.Items {
			if partition.Spec.Cluster.Name == object.GetName() {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: object.GetNamespace(),
						Name:      partition.Name,
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

// RaftPartitionReconciler reconciles a RaftPartition object
type RaftPartitionReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
}

// Reconcile reads that state of the cluster for a Store object and makes changes based on the state read
// and what is in the Store.Spec
func (r *RaftPartitionReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile RaftPartition")
	partition := &multiraftv1beta2.RaftPartition{}
	if err := r.client.Get(ctx, request.NamespacedName, partition); err != nil {
		log.Error(err, "Reconcile RaftPartition")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	cluster := &multiraftv1beta2.MultiRaftCluster{}
	clusterName := types.NamespacedName{
		Namespace: partition.Namespace,
		Name:      partition.Spec.Cluster.Name,
	}
	if err := r.client.Get(ctx, clusterName, cluster); err != nil {
		log.Error(err, "Reconcile RaftPartition")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if ok, err := r.reconcileMembers(ctx, cluster, partition); err != nil {
		log.Error(err, "Reconcile RaftPartition")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}

	if ok, err := r.reconcileStatus(ctx, partition); err != nil {
		log.Error(err, "Reconcile RaftPartition")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *RaftPartitionReconciler) reconcileMembers(ctx context.Context, cluster *multiraftv1beta2.MultiRaftCluster, partition *multiraftv1beta2.RaftPartition) (bool, error) {
	for ordinal := 0; ordinal < int(partition.Spec.Replicas); ordinal++ {
		if ok, err := r.reconcileMember(ctx, cluster, partition, ordinal); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *RaftPartitionReconciler) reconcileMember(ctx context.Context, cluster *multiraftv1beta2.MultiRaftCluster, partition *multiraftv1beta2.RaftPartition, ordinal int) (bool, error) {
	memberName := types.NamespacedName{
		Namespace: partition.Namespace,
		Name:      fmt.Sprintf("%s-%d", partition.Name, ordinal),
	}
	member := &multiraftv1beta2.RaftMember{}
	if err := r.client.Get(ctx, memberName, member); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}

		member = &multiraftv1beta2.RaftMember{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   memberName.Namespace,
				Name:        memberName.Name,
				Labels:      newMemberLabels(cluster, partition, ordinal),
				Annotations: newMemberAnnotations(cluster, partition, ordinal),
			},
			Spec: multiraftv1beta2.RaftMemberSpec{
				Pod: corev1.LocalObjectReference{
					Name: getMemberPodName(cluster, partition, ordinal),
				},
				Type: multiraftv1beta2.RaftVoter,
			},
		}
		if err := controllerutil.SetControllerReference(partition, member, r.scheme); err != nil {
			return false, err
		}
		if err := r.client.Create(ctx, member); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *RaftPartitionReconciler) reconcileStatus(ctx context.Context, partition *multiraftv1beta2.RaftPartition) (bool, error) {
	state := multiraftv1beta2.RaftPartitionReady
	for ordinal := 1; ordinal <= int(partition.Spec.Replicas); ordinal++ {
		memberName := types.NamespacedName{
			Namespace: partition.Namespace,
			Name:      fmt.Sprintf("%s-%d", partition.Name, ordinal),
		}
		member := &multiraftv1beta2.RaftMember{}
		if err := r.client.Get(ctx, memberName, member); err != nil {
			return false, err
		}
		if member.Status.State == multiraftv1beta2.RaftMemberNotReady {
			state = multiraftv1beta2.RaftPartitionNotReady
		}
	}

	if partition.Status.State != state {
		partition.Status.State = state
		if err := r.client.Status().Update(ctx, partition); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

var _ reconcile.Reconciler = (*RaftPartitionReconciler)(nil)
