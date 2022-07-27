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

func addRaftGroupController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &RaftGroupReconciler{
			client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			events: mgr.GetEventRecorderFor("atomix-multi-raft-storage"),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-raft-group-v2beta2", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Clusters that reference it
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftGroup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RaftMember
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftMember{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.RaftGroup{},
		IsController: true,
	})
	if err != nil {
		return err
	}
	return nil
}

// RaftGroupReconciler reconciles a RaftGroup object
type RaftGroupReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
}

func (r *RaftGroupReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile RaftGroup")
	group := &storagev2beta2.RaftGroup{}
	err := r.client.Get(ctx, request.NamespacedName, group)
	if err != nil {
		log.Error(err, "Reconcile RaftGroup")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ok, err := r.reconcileMembers(ctx, group)
	if err != nil {
		log.Error(err, "Reconcile RaftGroup")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}

	ok, err = r.reconcileStatus(ctx, group)
	if err != nil {
		log.Error(err, "Reconcile RaftGroup")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *RaftGroupReconciler) reconcileMembers(ctx context.Context, group *storagev2beta2.RaftGroup) (bool, error) {
	for _, memberConfig := range group.Spec.Members {
		if ok, err := r.reconcileMember(ctx, group, memberConfig); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *RaftGroupReconciler) reconcileMember(ctx context.Context, group *storagev2beta2.RaftGroup, config storagev2beta2.RaftMemberConfig) (bool, error) {
	member := &storagev2beta2.RaftMember{}
	memberName := types.NamespacedName{
		Namespace: group.Namespace,
		Name:      fmt.Sprintf("%s-%d", group.Name, config.NodeID),
	}
	if err := r.client.Get(ctx, memberName, member); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}

		member = &storagev2beta2.RaftMember{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: memberName.Namespace,
				Name:      memberName.Name,
				Labels:    group.Labels,
			},
			Spec: storagev2beta2.RaftMemberSpec{
				Cluster:          group.Spec.Cluster,
				GroupID:          group.Spec.GroupID,
				RaftMemberConfig: config,
			},
		}
		if err := controllerutil.SetControllerReference(group, member, r.scheme); err != nil {
			return false, err
		}
		if err := r.client.Create(ctx, member); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *RaftGroupReconciler) reconcileStatus(ctx context.Context, group *storagev2beta2.RaftGroup) (bool, error) {
	changed := false
	var recorders []func()
	memberState := storagev2beta2.RaftMemberReady
	for _, memberConfig := range group.Spec.Members {
		member := &storagev2beta2.RaftMember{}
		memberName := types.NamespacedName{
			Namespace: group.Namespace,
			Name:      fmt.Sprintf("%s-%d", group.Name, memberConfig.NodeID),
		}
		if err := r.client.Get(ctx, memberName, member); err != nil {
			return false, err
		}

		if member.Status.Term != nil && (group.Status.Term == nil || *member.Status.Term > *group.Status.Term) {
			group.Status.Term = member.Status.Term
			group.Status.Leader = nil
			recorders = append(recorders, func() {
				r.events.Eventf(group, "Normal", "TermChanged", "Term changed to %d", *group.Status.Term)
			})
			changed = true
		}

		if member.Status.Leader != nil && group.Status.Leader == nil {
			group.Status.Leader = member.Status.Leader
			recorders = append(recorders, func() {
				r.events.Eventf(group, "Normal", "LeaderChanged", "Leader changed to %s", *group.Status.Leader)
			})
			changed = true
		}

		if member.Status.State == storagev2beta2.RaftMemberNotReady {
			memberState = storagev2beta2.RaftMemberNotReady
		}
	}

	if memberState == storagev2beta2.RaftMemberReady && group.Status.State != storagev2beta2.RaftGroupReady {
		group.Status.State = storagev2beta2.RaftGroupReady
		recorders = append(recorders, func() {
			r.events.Eventf(group, "Normal", "Ready", "RaftGroup is ready")
		})
		changed = true
	}

	if memberState == storagev2beta2.RaftMemberNotReady && group.Status.State != storagev2beta2.RaftGroupNotReady {
		group.Status.State = storagev2beta2.RaftGroupNotReady
		recorders = append(recorders, func() {
			r.events.Eventf(group, "Normal", "NotReady", "RaftGroup is not ready")
		})
		changed = true
	}

	if changed {
		if err := r.client.Status().Update(ctx, group); err != nil {
			return false, err
		}
		for _, recorder := range recorders {
			recorder()
		}
		return true, nil
	}
	return false, nil
}

var _ reconcile.Reconciler = (*RaftGroupReconciler)(nil)
