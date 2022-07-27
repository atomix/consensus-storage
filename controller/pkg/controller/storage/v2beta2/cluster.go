// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	"context"
	"fmt"
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

	storagev2beta2 "github.com/atomix/multi-raft-storage/controller/pkg/apis/storage/v2beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func addMultiRaftClusterController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &MultiRaftClusterReconciler{
			client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			events: mgr.GetEventRecorderFor("atomix-multi-raft-storage"),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-multi-raft-cluster-v2beta2", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Clusters that reference it
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.MultiRaftCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource MultiRaftNode
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.MultiRaftNode{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftCluster{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RaftGroup
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftGroup{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftCluster{},
		IsController: true,
	})
	if err != nil {
		return err
	}
	return nil
}

// MultiRaftClusterReconciler reconciles a MultiRaftCluster object
type MultiRaftClusterReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
}

func (r *MultiRaftClusterReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile MultiRaftCluster")
	cluster := &storagev2beta2.MultiRaftCluster{}
	err := r.client.Get(ctx, request.NamespacedName, cluster)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftCluster")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ok, err := r.reconcileNodes(ctx, cluster)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftCluster")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}

	ok, err = r.reconcileGroups(ctx, cluster)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftCluster")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}

	ok, err = r.reconcileStatus(ctx, cluster)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftCluster")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *MultiRaftClusterReconciler) reconcileNodes(ctx context.Context, cluster *storagev2beta2.MultiRaftCluster) (bool, error) {
	for _, nodeConfig := range cluster.Spec.Nodes {
		if ok, err := r.reconcileNode(ctx, cluster, nodeConfig); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *MultiRaftClusterReconciler) reconcileNode(ctx context.Context, cluster *storagev2beta2.MultiRaftCluster, config storagev2beta2.MultiRaftNodeConfig) (bool, error) {
	node := &storagev2beta2.MultiRaftNode{}
	nodeName := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      fmt.Sprintf("%s-%d", cluster.Name, config.NodeID),
	}
	if err := r.client.Get(ctx, nodeName, node); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}

		node = &storagev2beta2.MultiRaftNode{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nodeName.Namespace,
				Name:      nodeName.Name,
				Labels:    cluster.Labels,
			},
			Spec: storagev2beta2.MultiRaftNodeSpec{
				Cluster:             cluster.Name,
				MultiRaftNodeConfig: config,
			},
		}
		if err := controllerutil.SetControllerReference(cluster, node, r.scheme); err != nil {
			return false, err
		}
		if err := r.client.Create(ctx, node); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *MultiRaftClusterReconciler) reconcileGroups(ctx context.Context, cluster *storagev2beta2.MultiRaftCluster) (bool, error) {
	for _, groupConfig := range cluster.Spec.Groups {
		if ok, err := r.reconcileGroup(ctx, cluster, groupConfig); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *MultiRaftClusterReconciler) reconcileGroup(ctx context.Context, cluster *storagev2beta2.MultiRaftCluster, config storagev2beta2.RaftGroupConfig) (bool, error) {
	group := &storagev2beta2.RaftGroup{}
	groupName := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      fmt.Sprintf("%s-%d", cluster.Name, config.GroupID),
	}
	if err := r.client.Get(ctx, groupName, group); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}

		group = &storagev2beta2.RaftGroup{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: groupName.Namespace,
				Name:      groupName.Name,
				Labels:    cluster.Labels,
			},
			Spec: storagev2beta2.RaftGroupSpec{
				Cluster:         cluster.Name,
				RaftGroupConfig: config,
			},
		}
		if err := controllerutil.SetControllerReference(cluster, group, r.scheme); err != nil {
			return false, err
		}
		if err := r.client.Create(ctx, group); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *MultiRaftClusterReconciler) reconcileStatus(ctx context.Context, cluster *storagev2beta2.MultiRaftCluster) (bool, error) {
	for _, nodeConfig := range cluster.Spec.Nodes {
		node := &storagev2beta2.MultiRaftNode{}
		nodeName := types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-%d", cluster.Name, nodeConfig.NodeID),
		}
		if err := r.client.Get(ctx, nodeName, node); err != nil {
			return false, err
		}
		if node.Status.State == storagev2beta2.MultiRaftNodeNotReady {
			if cluster.Status.State != storagev2beta2.MultiRaftClusterNotReady {
				cluster.Status.State = storagev2beta2.MultiRaftClusterNotReady
				if err := r.client.Status().Update(ctx, cluster); err != nil {
					return false, err
				}
				r.events.Eventf(cluster, "Normal", "NotReady", "MultiRaftCluster is not ready")
				return true, nil
			}
			return false, nil
		}
	}

	for _, groupConfig := range cluster.Spec.Groups {
		group := &storagev2beta2.RaftGroup{}
		groupName := types.NamespacedName{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-%d", cluster.Name, groupConfig.GroupID),
		}
		if err := r.client.Get(ctx, groupName, group); err != nil {
			return false, err
		}
		if group.Status.State == storagev2beta2.RaftGroupNotReady {
			if cluster.Status.State != storagev2beta2.MultiRaftClusterNotReady {
				cluster.Status.State = storagev2beta2.MultiRaftClusterNotReady
				if err := r.client.Status().Update(ctx, cluster); err != nil {
					return false, err
				}
				r.events.Eventf(cluster, "Normal", "NotReady", "MultiRaftCluster is not ready")
				return true, nil
			}
			return false, nil
		}
	}

	if cluster.Status.State != storagev2beta2.MultiRaftClusterReady {
		cluster.Status.State = storagev2beta2.MultiRaftClusterReady
		if err := r.client.Status().Update(ctx, cluster); err != nil {
			return false, err
		}
		r.events.Eventf(cluster, "Normal", "Ready", "MultiRaftCluster is ready")
		return true, nil
	}
	return false, nil
}

var _ reconcile.Reconciler = (*MultiRaftClusterReconciler)(nil)
