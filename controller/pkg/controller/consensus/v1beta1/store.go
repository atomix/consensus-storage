// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"context"
	multiraftv1beta2 "github.com/atomix/consensus-storage/controller/pkg/apis/multiraft/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	consensusv1beta1 "github.com/atomix/consensus-storage/controller/pkg/apis/consensus/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func addConsensusStoreController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &MultiRaftStoreReconciler{
			client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			events: mgr.GetEventRecorderFor("atomix-consensus-storage"),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-consensus-store", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Stores that reference it
	err = controller.Watch(&source.Kind{Type: &consensusv1beta1.ConsensusStore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource MultiRaftStore
	err = controller.Watch(&source.Kind{Type: &multiraftv1beta2.MultiRaftStore{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &consensusv1beta1.ConsensusStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource MultiRaftCluster
	err = controller.Watch(&source.Kind{Type: &multiraftv1beta2.MultiRaftCluster{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &consensusv1beta1.ConsensusStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}
	return nil
}

// MultiRaftStoreReconciler reconciles a MultiRaftCluster object
type MultiRaftStoreReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
}

// Reconcile reads that state of the cluster for a Store object and makes changes based on the state read
// and what is in the Store.Spec
func (r *MultiRaftStoreReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile ConsensusStore")
	store := &consensusv1beta1.ConsensusStore{}
	err := r.client.Get(ctx, request.NamespacedName, store)
	if err != nil {
		log.Error(err, "Reconcile ConsensusStore")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	multiRaftCluster := &multiraftv1beta2.MultiRaftCluster{}
	if err := r.client.Get(ctx, request.NamespacedName, multiRaftCluster); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "Reconcile ConsensusStore")
			return reconcile.Result{}, err
		}

		multiRaftCluster = &multiraftv1beta2.MultiRaftCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:        store.Name,
				Namespace:   store.Namespace,
				Labels:      store.Labels,
				Annotations: store.Annotations,
			},
			Spec: multiraftv1beta2.MultiRaftClusterSpec{
				Replicas:            uint32(store.Spec.Replicas),
				Image:               store.Spec.Image,
				ImagePullPolicy:     store.Spec.ImagePullPolicy,
				ImagePullSecrets:    store.Spec.ImagePullSecrets,
				SecurityContext:     store.Spec.SecurityContext,
				VolumeClaimTemplate: store.Spec.VolumeClaimTemplate,
				Config: multiraftv1beta2.MultiRaftClusterConfig{
					Server: multiraftv1beta2.MultiRaftServerConfig{
						ReadBufferSize:       store.Spec.Config.Server.ReadBufferSize,
						WriteBufferSize:      store.Spec.Config.Server.WriteBufferSize,
						MaxRecvMsgSize:       store.Spec.Config.Server.MaxRecvMsgSize,
						MaxSendMsgSize:       store.Spec.Config.Server.MaxSendMsgSize,
						MaxConcurrentStreams: store.Spec.Config.Server.MaxConcurrentStreams,
						NumStreamWorkers:     store.Spec.Config.Server.NumStreamWorkers,
					},
				},
			},
		}

		if store.Spec.Config.Logging.Loggers != nil {
			multiRaftCluster.Spec.Config.Logging.Loggers = make(map[string]multiraftv1beta2.LoggerConfig)
			for name, config := range store.Spec.Config.Logging.Loggers {
				loggerConfig := multiraftv1beta2.LoggerConfig{
					Level: config.Level,
				}
				if config.Output != nil {
					loggerConfig.Output = make(map[string]multiraftv1beta2.OutputConfig)
					for outputName, outputConfig := range config.Output {
						loggerConfig.Output[outputName] = multiraftv1beta2.OutputConfig{
							Sink:  outputConfig.Sink,
							Level: outputConfig.Level,
						}
					}
				}
				multiRaftCluster.Spec.Config.Logging.Loggers[name] = loggerConfig
			}
		}
		if err := r.client.Create(ctx, multiRaftCluster); err != nil && !k8serrors.IsAlreadyExists(err) {
			log.Error(err, "Reconcile ConsensusStore")
			return reconcile.Result{}, err
		}
	}

	multiRaftStore := &multiraftv1beta2.MultiRaftStore{}
	if err := r.client.Get(ctx, request.NamespacedName, multiRaftStore); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "Reconcile ConsensusStore")
			return reconcile.Result{}, err
		}

		multiRaftStore = &multiraftv1beta2.MultiRaftStore{
			ObjectMeta: metav1.ObjectMeta{
				Name:        store.Name,
				Namespace:   store.Namespace,
				Labels:      store.Labels,
				Annotations: store.Annotations,
			},
			Spec: multiraftv1beta2.MultiRaftStoreSpec{
				Cluster: corev1.LocalObjectReference{
					Name: store.Name,
				},
				Partitions: uint32(store.Spec.Groups),
			},
		}
		if err := r.client.Create(ctx, multiRaftStore); err != nil && !k8serrors.IsAlreadyExists(err) {
			log.Error(err, "Reconcile ConsensusStore")
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

var _ reconcile.Reconciler = (*MultiRaftStoreReconciler)(nil)
