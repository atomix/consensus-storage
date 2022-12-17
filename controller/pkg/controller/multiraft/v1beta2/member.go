// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1beta2

import (
	"context"
	"fmt"
	"github.com/atomix/consensus-storage/node/pkg/consensus"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
	"time"

	multiraftv1beta2 "github.com/atomix/consensus-storage/controller/pkg/apis/multiraft/v1beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func addRaftMemberController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &RaftMemberReconciler{
			client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			events: mgr.GetEventRecorderFor("atomix-consensus-storage"),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-raft-member", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Stores that reference it
	err = controller.Watch(&source.Kind{Type: &multiraftv1beta2.RaftMember{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pod
	err = controller.Watch(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		members := &multiraftv1beta2.RaftMemberList{}
		if err := mgr.GetClient().List(context.Background(), members, &client.ListOptions{Namespace: object.GetNamespace()}); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, member := range members.Items {
			if member.Spec.Pod.Name == object.GetName() {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: object.GetNamespace(),
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
	client client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
	mu     sync.RWMutex
}

// Reconcile reads that state of the cluster for a Store object and makes changes based on the state read
// and what is in the Store.Spec
func (r *RaftMemberReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile RaftMember")
	member := &multiraftv1beta2.RaftMember{}
	err := r.client.Get(ctx, request.NamespacedName, member)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if member.DeletionTimestamp != nil {
		if err := r.reconcileDelete(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if err := r.reconcileCreate(ctx, member); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *RaftMemberReconciler) reconcileCreate(ctx context.Context, member *multiraftv1beta2.RaftMember) error {
	if !hasFinalizer(member, raftMemberKey) {
		log.Infof("Adding '%s' finalizer to RaftMember %s/%s", raftMemberKey, member.Namespace, member.Name)
		addFinalizer(member, raftMemberKey)
		if err := r.client.Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		return nil
	}

	storeName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Annotations[multiRaftStoreKey],
	}
	store := &multiraftv1beta2.MultiRaftStore{}
	if err := r.client.Get(ctx, storeName, store); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return err
	}

	clusterName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Spec.Cluster.Name,
	}
	cluster := &multiraftv1beta2.MultiRaftCluster{}
	if err := r.client.Get(ctx, clusterName, cluster); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return err
	}

	partitionName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      fmt.Sprintf("%s-%s", storeName.Name, member.Annotations[raftPartitionKey]),
	}
	partition := &multiraftv1beta2.RaftPartition{}
	if err := r.client.Get(ctx, partitionName, partition); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return err
	}

	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Spec.Pod.Name,
	}
	pod := &corev1.Pod{}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return err
	}

	if ok, err := r.addMember(ctx, store, cluster, partition, pod, member); err != nil {
		return err
	} else if ok {
		return nil
	}
	return nil
}

func (r *RaftMemberReconciler) reconcileDelete(ctx context.Context, member *multiraftv1beta2.RaftMember) error {
	if !hasFinalizer(member, raftMemberKey) {
		return nil
	}

	storeName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Annotations[multiRaftStoreKey],
	}
	store := &multiraftv1beta2.MultiRaftStore{}
	if err := r.client.Get(ctx, storeName, store); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		log.Infof("Removing '%s' finalizer from RaftMember %s/%s", raftMemberKey, member.Namespace, member.Name)
		removeFinalizer(member, raftMemberKey)
		if err := r.client.Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		return nil
	}

	clusterName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Spec.Cluster.Name,
	}
	cluster := &multiraftv1beta2.MultiRaftCluster{}
	if err := r.client.Get(ctx, clusterName, cluster); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		log.Infof("Removing '%s' finalizer from RaftMember %s/%s", raftMemberKey, member.Namespace, member.Name)
		removeFinalizer(member, raftMemberKey)
		if err := r.client.Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		return nil
	}

	partitionName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      fmt.Sprintf("%s-%s", storeName.Name, member.Annotations[raftPartitionKey]),
	}
	partition := &multiraftv1beta2.RaftPartition{}
	if err := r.client.Get(ctx, partitionName, partition); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		log.Infof("Removing '%s' finalizer from RaftMember %s/%s", raftMemberKey, member.Namespace, member.Name)
		removeFinalizer(member, raftMemberKey)
		if err := r.client.Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		return nil
	}

	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Spec.Pod.Name,
	}
	pod := &corev1.Pod{}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		log.Infof("Removing '%s' finalizer from RaftMember %s/%s", raftMemberKey, member.Namespace, member.Name)
		removeFinalizer(member, raftMemberKey)
		if err := r.client.Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return err
		}
		return nil
	}

	if ok, err := r.removeMember(ctx, store, cluster, partition, pod, member); err != nil {
		return err
	} else if ok {
		return nil
	}

	log.Infof("Removing '%s' finalizer from RaftMember %s/%s", raftMemberKey, member.Namespace, member.Name)
	removeFinalizer(member, raftMemberKey)
	if err := r.client.Update(ctx, member); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return err
	}
	return nil
}

func (r *RaftMemberReconciler) addMember(ctx context.Context, store *multiraftv1beta2.MultiRaftStore, cluster *multiraftv1beta2.MultiRaftCluster, partition *multiraftv1beta2.RaftPartition, pod *corev1.Pod, member *multiraftv1beta2.RaftMember) (bool, error) {
	if member.Status.PodRef == nil || member.Status.PodRef.UID != pod.UID {
		log.Infof("Pod UID has changed for RaftMember %s/%s; reverting RaftMember status", member.Namespace, member.Name)
		member.Status.PodRef = &corev1.ObjectReference{
			APIVersion: pod.APIVersion,
			Kind:       pod.Kind,
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			UID:        pod.UID,
		}
		member.Status.Version = nil
		if err := r.client.Status().Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return false, err
		}
		return true, nil
	}

	var containerVersion int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == nodeContainerName {
			containerVersion = containerStatus.RestartCount + 1
			break
		}
	}

	if member.Status.Version == nil || containerVersion > *member.Status.Version {
		if member.Status.State != multiraftv1beta2.RaftMemberNotReady {
			member.Status.State = multiraftv1beta2.RaftMemberNotReady
			if err := r.client.Status().Update(ctx, member); err != nil {
				log.Error(err, "Reconcile RaftMember")
				return false, err
			}
			r.events.Eventf(member, "Normal", "StateChanged", "State changed to %s", member.Status.State)
			return true, nil
		}

		switch member.Spec.BootstrapPolicy {
		case multiraftv1beta2.RaftBootstrap:
			members := make([]consensus.MemberConfig, 0, len(member.Spec.Config.Peers))
			for _, peer := range member.Spec.Config.Peers {
				host := fmt.Sprintf("%s.%s.%s.svc.%s", peer.Pod.Name, getHeadlessServiceName(member.Spec.Cluster.Name), member.Namespace, getClusterDomain())
				members = append(members, consensus.MemberConfig{
					MemberID: consensus.MemberID(peer.RaftNodeID),
					Host:     host,
					Port:     protocolPort,
					Role:     consensus.MemberRole_MEMBER,
				})
			}

			address := fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort)
			conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Error(err, "Reconcile RaftMember")
				return false, err
			}
			defer conn.Close()

			// Bootstrap the member with the initial configuration
			log.Infof("Boostrapping RaftMember %s/%s in RaftPartition %s", member.Namespace, member.Name, partition.Name)
			client := consensus.NewNodeClient(conn)
			request := &consensus.BootstrapGroupRequest{
				GroupID:  consensus.GroupID(member.Spec.ShardID),
				MemberID: consensus.MemberID(member.Spec.MemberID),
				Members:  members,
			}
			if _, err := client.BootstrapGroup(ctx, request); err != nil {
				err = errors.FromProto(err)
				if !errors.IsAlreadyExists(err) {
					r.events.Eventf(store, "Warning", "BootstrapShardFailed", "Failed to bootstrap node %d on shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
					r.events.Eventf(cluster, "Warning", "BootstrapShardFailed", "Failed to bootstrap node %d on shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
					r.events.Eventf(partition, "Warning", "BootstrapShardFailed", "Failed to bootstrap node %d on shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
					r.events.Eventf(pod, "Warning", "BootstrapShardFailed", "Failed to bootstrap node %d on shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
					r.events.Eventf(member, "Warning", "BootstrapShardFailed", "Failed to bootstrap node on shard %d: %s", member.Spec.ShardID, err.Error())
					return false, err
				}
			} else {
				r.events.Eventf(store, "Normal", "BootstrappedShard", "Bootstrapped node %d on shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
				r.events.Eventf(cluster, "Normal", "BootstrappedShard", "Bootstrapped node %d on shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
				r.events.Eventf(partition, "Normal", "BootstrappedShard", "Bootstrapped node %d on shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
				r.events.Eventf(pod, "Normal", "BootstrappedShard", "Bootstrapped node %d on shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
				r.events.Eventf(member, "Normal", "BootstrappedShard", "Bootstrapped node on shard %d", member.Spec.ShardID)
			}
		case multiraftv1beta2.RaftJoin:
			// Loop through the list of peers and attempt to add the member to the Raft group until successful
			for _, peer := range member.Spec.Config.Peers {
				if ok, err := r.tryAddMember(ctx, store, cluster, partition, member, peer); err != nil {
					return false, err
				} else if ok {
					address := fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort)
					conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
					if err != nil {
						log.Error(err, "Reconcile RaftMember")
						return false, err
					}

					// Bootstrap the member by joining it to the cluster
					log.Infof("Joining RaftMember %s/%s to RaftPartition %s", member.Namespace, member.Name, partition.Name)
					client := consensus.NewNodeClient(conn)
					request := &consensus.JoinGroupRequest{
						GroupID:  consensus.GroupID(member.Spec.ShardID),
						MemberID: consensus.MemberID(member.Spec.MemberID),
					}
					if _, err := client.JoinGroup(ctx, request); err != nil {
						log.Error(err, "Reconcile RaftMember")
						err = errors.FromProto(err)
						if !errors.IsAlreadyExists(err) {
							r.events.Eventf(store, "Warning", "JoinShardFailed", "Failed to join node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
							r.events.Eventf(cluster, "Warning", "JoinShardFailed", "Failed to join node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
							r.events.Eventf(partition, "Warning", "JoinShardFailed", "Failed to join node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
							r.events.Eventf(pod, "Warning", "JoinShardFailed", "Failed to join node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
							r.events.Eventf(member, "Warning", "JoinShardFailed", "Failed to join node to shard %d: %s", member.Spec.ShardID, err.Error())
							_ = conn.Close()
							return false, err
						}
					} else {
						r.events.Eventf(store, "Normal", "JoinedShard", "Joined node %d to shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
						r.events.Eventf(cluster, "Normal", "JoinedShard", "Joined node %d to shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
						r.events.Eventf(partition, "Normal", "JoinedShard", "Joined node %d to shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
						r.events.Eventf(pod, "Normal", "JoinedShard", "Joined node %d to shard %d", member.Spec.RaftNodeID, member.Spec.ShardID)
						r.events.Eventf(member, "Normal", "JoinedShard", "Joined node to shard %d", member.Spec.ShardID)
					}
					_ = conn.Close()
					break
				}
			}
		}

		member.Status.Version = &containerVersion
		if err := r.client.Status().Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *RaftMemberReconciler) tryAddMember(ctx context.Context, store *multiraftv1beta2.MultiRaftStore, cluster *multiraftv1beta2.MultiRaftCluster, partition *multiraftv1beta2.RaftPartition, member *multiraftv1beta2.RaftMember, peer multiraftv1beta2.RaftMemberReference) (bool, error) {
	pod := &corev1.Pod{}
	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      peer.Pod.Name,
	}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		return false, err
	}

	address := fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		return false, err
	}
	defer conn.Close()

	client := consensus.NewNodeClient(conn)
	getConfigRequest := &consensus.GetConfigRequest{
		GroupID: consensus.GroupID(member.Spec.ShardID),
	}
	getConfigResponse, err := client.GetConfig(ctx, getConfigRequest)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		return false, err
	}

	log.Infof("Adding RaftMember %s/%s to RaftPartition %s via %s", member.Namespace, member.Name, partition.Name, peer.Pod.Name)
	addMemberRequest := &consensus.AddMemberRequest{
		GroupID: consensus.GroupID(member.Spec.ShardID),
		Member: consensus.MemberConfig{
			MemberID: consensus.MemberID(member.Spec.MemberID),
			Host:     fmt.Sprintf("%s.%s.%s.svc.%s", member.Spec.Pod.Name, getHeadlessServiceName(member.Spec.Cluster.Name), member.Namespace, getClusterDomain()),
			Port:     protocolPort,
		},
		Version: getConfigResponse.Group.Version,
	}
	_, err = client.AddMember(ctx, addMemberRequest)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		r.events.Eventf(store, "Warning", "AddMemberFailed", "Failed to add node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(cluster, "Warning", "AddMemberFailed", "Failed to add node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(partition, "Warning", "AddMemberFailed", "Failed to add node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(pod, "Warning", "AddMemberFailed", "Failed to add node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(member, "Warning", "AddMemberFailed", "Failed to add node to shard %d: %s", member.Spec.ShardID, err.Error())
		return false, err
	}
	r.events.Eventf(store, "Normal", "AddedMember", "Added node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(cluster, "Normal", "AddedMember", "Added node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(partition, "Normal", "AddedMember", "Added node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(pod, "Normal", "AddedMember", "Added node %d to shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(member, "Normal", "AddedMember", "Added node to shard %d: %s", member.Spec.ShardID)
	return true, nil
}

func (r *RaftMemberReconciler) removeMember(ctx context.Context, store *multiraftv1beta2.MultiRaftStore, cluster *multiraftv1beta2.MultiRaftCluster, partition *multiraftv1beta2.RaftPartition, pod *corev1.Pod, member *multiraftv1beta2.RaftMember) (bool, error) {
	if member.Status.State != multiraftv1beta2.RaftMemberNotReady {
		member.Status.State = multiraftv1beta2.RaftMemberNotReady
		if err := r.client.Status().Update(ctx, member); err != nil {
			log.Error(err, "Reconcile RaftMember")
			return false, err
		}
		r.events.Eventf(member, "Normal", "StateChanged", "State changed to %s", member.Status.State)
		return true, nil
	}

	address := fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		return false, err
	}
	defer conn.Close()

	// Shutdown the group member.
	log.Infof("Terminating RaftMember %s/%s in RaftPartition %s", member.Namespace, member.Name, partition.Name)
	client := consensus.NewNodeClient(conn)
	request := &consensus.LeaveGroupRequest{
		GroupID: consensus.GroupID(member.Spec.ShardID),
	}
	if _, err := client.LeaveGroup(ctx, request); err != nil {
		log.Error(err, "Reconcile RaftMember")
	}

	// Loop through the list of peers and attempt to remove the member from the Raft group until successful
	for _, peer := range member.Spec.Config.Peers {
		if ok, err := r.tryRemoveMember(ctx, store, cluster, partition, member, peer); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *RaftMemberReconciler) tryRemoveMember(ctx context.Context, store *multiraftv1beta2.MultiRaftStore, cluster *multiraftv1beta2.MultiRaftCluster, partition *multiraftv1beta2.RaftPartition, member *multiraftv1beta2.RaftMember, peer multiraftv1beta2.RaftMemberReference) (bool, error) {
	pod := &corev1.Pod{}
	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      peer.Pod.Name,
	}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		log.Error(err, "Reconcile RaftMember")
		return false, err
	}

	address := fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		return false, err
	}
	defer conn.Close()

	client := consensus.NewNodeClient(conn)
	getConfigRequest := &consensus.GetConfigRequest{
		GroupID: consensus.GroupID(member.Spec.ShardID),
	}
	getConfigResponse, err := client.GetConfig(ctx, getConfigRequest)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		return false, err
	}

	log.Infof("Removing RaftMember %s/%s from RaftPartition %s via %s", member.Namespace, member.Name, partition.Name, peer.Pod.Name)
	removeMemberRequest := &consensus.RemoveMemberRequest{
		GroupID:  consensus.GroupID(member.Spec.ShardID),
		MemberID: consensus.MemberID(member.Spec.MemberID),
		Version:  getConfigResponse.Group.Version,
	}
	_, err = client.RemoveMember(ctx, removeMemberRequest)
	if err != nil {
		log.Error(err, "Reconcile RaftMember")
		r.events.Eventf(store, "Warning", "RemoveMemberFailed", "Failed to remove node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(cluster, "Warning", "RemoveMemberFailed", "Failed to remove node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(partition, "Warning", "RemoveMemberFailed", "Failed to remove node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(pod, "Warning", "RemoveMemberFailed", "Failed to remove node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID, err.Error())
		r.events.Eventf(member, "Warning", "RemoveMemberFailed", "Failed to remove node from shard %d: %s", member.Spec.ShardID, err.Error())
		return false, err
	}
	r.events.Eventf(store, "Normal", "RemovedMember", "Removed node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(cluster, "Normal", "RemovedMember", "Removed node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(partition, "Normal", "RemovedMember", "Removed node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(pod, "Normal", "RemovedMember", "Removed node %d from shard %d: %s", member.Spec.RaftNodeID, member.Spec.ShardID)
	r.events.Eventf(member, "Normal", "RemovedMember", "Removed node from shard %d: %s", member.Spec.ShardID)
	return true, nil
}

var _ reconcile.Reconciler = (*RaftMemberReconciler)(nil)
