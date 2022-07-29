// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	"context"
	"encoding/json"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/controller/pkg/apis/atomix/v1beta1"
	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/pointer"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"

	storagev2beta2 "github.com/atomix/multi-raft-storage/controller/pkg/apis/storage/v2beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	apiPort                  = 5678
	protocolPort             = 5679
	probePort                = 5679
	defaultImageEnv          = "DEFAULT_NODE_IMAGE"
	defaultImage             = "atomix/multi-raft-node:latest"
	headlessServiceSuffix    = "hs"
	appLabel                 = "app"
	storeLabel               = "store"
	appAtomix                = "atomix"
	nodeContainerName        = "atomix-multi-raft-node"
	multiRaftStoreAnnotation = "multiraft.storage.atomix.io/store"
	multiRaftStoreFinalizer  = "storage.atomix.io/multi-raft"
)

const (
	configPath     = "/etc/atomix"
	raftConfigFile = "raft.json"
	dataPath       = "/var/lib/atomix"
)

const (
	configVolume = "config"
	dataVolume   = "data"
)

const (
	driverName    = "MultiRaft"
	driverVersion = "v1beta1"
)

const clusterDomainEnv = "CLUSTER_DOMAIN"

func addMultiRaftStoreController(mgr manager.Manager) error {
	options := controller.Options{
		Reconciler: &MultiRaftStoreReconciler{
			client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			events: mgr.GetEventRecorderFor("atomix-multi-raft-storage"),
		},
		RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond*10, time.Second*5),
	}

	// Create a new controller
	controller, err := controller.New("atomix-multi-raft-store-v2beta2", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Stores that reference it
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.MultiRaftStore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSet
	err = controller.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RaftGroup
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftGroup{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RaftMember
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftMember{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Store
	err = controller.Watch(&source.Kind{Type: &v1beta1.Store{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pod
	err = controller.Watch(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		store, ok := object.GetAnnotations()[multiRaftStoreAnnotation]
		if !ok {
			return nil
		}
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Namespace: object.GetNamespace(),
					Name:      store,
				},
			},
		}
	}))
	if err != nil {
		return err
	}
	return nil
}

// MultiRaftStoreReconciler reconciles a MultiRaftStore object
type MultiRaftStoreReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	events record.EventRecorder
}

// Reconcile reads that state of the cluster for a Store object and makes changes based on the state read
// and what is in the Store.Spec
func (r *MultiRaftStoreReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile MultiRaftStore")
	store := &storagev2beta2.MultiRaftStore{}
	err := r.client.Get(ctx, request.NamespacedName, store)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if err := r.reconcileConfigMap(ctx, store); err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	}

	if err := r.reconcileStatefulSet(ctx, store); err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	}

	if err := r.reconcileService(ctx, store); err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	}

	if err := r.reconcileHeadlessService(ctx, store); err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	}

	if ok, err := r.reconcileGroups(ctx, store); err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}

	if err := r.reconcileStore(ctx, store); err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *MultiRaftStoreReconciler) reconcileConfigMap(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol config map")
	cm := &corev1.ConfigMap{}
	name := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      store.Name,
	}
	err := r.client.Get(ctx, name, cm)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addConfigMap(ctx, store)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addConfigMap(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating raft ConfigMap", "Name", store.Name, "Namespace", store.Namespace)
	raftConfig, err := newRaftConfigString(store)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        store.Name,
			Namespace:   store.Namespace,
			Labels:      newStoreLabels(store),
			Annotations: newStoreAnnotations(store),
		},
		Data: map[string]string{
			raftConfigFile: raftConfig,
		},
	}

	if err := controllerutil.SetControllerReference(store, cm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, cm)
}

// newRaftConfigString creates a protocol configuration string for the given store and protocol
func newRaftConfigString(store *storagev2beta2.MultiRaftStore) (string, error) {
	config := multiraftv1.MultiRaftConfig{}

	electionTimeout := store.Spec.RaftConfig.ElectionTimeout
	if electionTimeout != nil {
		config.ElectionTimeout = &electionTimeout.Duration
	}

	heartbeatPeriod := store.Spec.RaftConfig.HeartbeatPeriod
	if heartbeatPeriod != nil {
		config.HeartbeatPeriod = &heartbeatPeriod.Duration
	}

	entryThreshold := store.Spec.RaftConfig.SnapshotEntryThreshold
	if entryThreshold != nil {
		config.SnapshotEntryThreshold = uint64(*entryThreshold)
	} else {
		config.SnapshotEntryThreshold = 10000
	}

	retainEntries := store.Spec.RaftConfig.CompactionRetainEntries
	if retainEntries != nil {
		config.CompactionRetainEntries = uint64(*retainEntries)
	} else {
		config.CompactionRetainEntries = 1000
	}

	bytes, err := json.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *MultiRaftStoreReconciler) reconcileStatefulSet(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol stateful set")
	statefulSet := &appsv1.StatefulSet{}
	name := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      store.Name,
	}
	err := r.client.Get(ctx, name, statefulSet)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addStatefulSet(ctx, store)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addStatefulSet(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating raft replicas", "Name", store.Name, "Namespace", store.Namespace)

	image := getImage(store)
	pullPolicy := store.Spec.ImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}

	volumes := []corev1.Volume{
		{
			Name: configVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: store.Name,
					},
				},
			},
		},
	}

	var volumeClaimTemplates []corev1.PersistentVolumeClaim

	dataVolumeName := dataVolume
	if store.Spec.VolumeClaimTemplate != nil {
		pvc := store.Spec.VolumeClaimTemplate
		if pvc.Name == "" {
			pvc.Name = dataVolume
		} else {
			dataVolumeName = pvc.Name
		}
		volumeClaimTemplates = append(volumeClaimTemplates, *pvc)
	} else {
		volumes = append(volumes, corev1.Volume{
			Name: dataVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	set := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        store.Name,
			Namespace:   store.Namespace,
			Labels:      newStoreLabels(store),
			Annotations: newStoreAnnotations(store),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: getStoreHeadlessServiceName(store.Name),
			Replicas:    pointer.Int32Ptr(int32(getNumReplicas(store))),
			Selector: &metav1.LabelSelector{
				MatchLabels: newStoreLabels(store),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      newStoreLabels(store),
					Annotations: newStoreAnnotations(store),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            nodeContainerName,
							Image:           image,
							ImagePullPolicy: pullPolicy,
							Ports: []corev1.ContainerPort{
								{
									Name:          "api",
									ContainerPort: apiPort,
								},
								{
									Name:          "protocol",
									ContainerPort: protocolPort,
								},
							},
							Command: []string{
								"bash",
								"-c",
								fmt.Sprintf(`set -ex
[[ `+"`hostname`"+` =~ -([0-9]+)$ ]] || exit 1
ordinal=${BASH_REMATCH[1]}
atomix-multi-raft-node --config %s/%s --api-port %d --raft-host %s-$ordinal.%s.%s.svc.%s --raft-port %d`,
									configPath, raftConfigFile, apiPort, store.Name, getStoreHeadlessServiceName(store.Name), store.Namespace, getClusterDomain(), protocolPort),
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{Type: intstr.Int, IntVal: probePort},
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      10,
								FailureThreshold:    12,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{Type: intstr.Int, IntVal: probePort},
									},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      10,
							},
							SecurityContext: store.Spec.SecurityContext,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      dataVolumeName,
									MountPath: dataPath,
								},
								{
									Name:      configVolume,
									MountPath: configPath,
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 1,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: newStoreLabels(store),
										},
										Namespaces:  []string{store.Namespace},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					ImagePullSecrets: store.Spec.ImagePullSecrets,
					Volumes:          volumes,
				},
			},
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}

	if err := controllerutil.SetControllerReference(store, set, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, set)
}

func (r *MultiRaftStoreReconciler) reconcileService(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol service")
	service := &corev1.Service{}
	name := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      store.Name,
	}
	err := r.client.Get(ctx, name, service)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addService(ctx, store)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addService(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating raft service", "Name", store.Name, "Namespace", store.Namespace)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        store.Name,
			Namespace:   store.Namespace,
			Labels:      newStoreLabels(store),
			Annotations: newStoreAnnotations(store),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "api",
					Port: apiPort,
				},
				{
					Name: "protocol",
					Port: protocolPort,
				},
			},
			Selector: newStoreLabels(store),
		},
	}

	if err := controllerutil.SetControllerReference(store, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, service)
}

func (r *MultiRaftStoreReconciler) reconcileHeadlessService(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol headless service")
	service := &corev1.Service{}
	name := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      getStoreHeadlessServiceName(store.Name),
	}
	err := r.client.Get(ctx, name, service)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addHeadlessService(ctx, store)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addHeadlessService(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating headless raft service", "Name", store.Name, "Namespace", store.Namespace)

	annotations := newStoreAnnotations(store)
	annotations["service.alpha.kubernetes.io/tolerate-unready-endpoints"] = "true"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        getStoreHeadlessServiceName(store.Name),
			Namespace:   store.Namespace,
			Labels:      newStoreLabels(store),
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "api",
					Port: apiPort,
				},
				{
					Name: "protocol",
					Port: protocolPort,
				},
			},
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
			Selector:                 newStoreLabels(store),
		},
	}

	if err := controllerutil.SetControllerReference(store, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, service)
}

func (r *MultiRaftStoreReconciler) reconcileGroups(ctx context.Context, store *storagev2beta2.MultiRaftStore) (bool, error) {
	numGroups := getNumGroups(store)
	state := storagev2beta2.MultiRaftStoreReady
	for groupID := 1; groupID <= numGroups; groupID++ {
		if group, updated, err := r.reconcileGroup(ctx, store, groupID); err != nil {
			return false, err
		} else if updated {
			return true, nil
		} else if group.Status.State == storagev2beta2.RaftGroupNotReady {
			state = storagev2beta2.MultiRaftStoreNotReady
		}
	}

	if store.Status.State != state {
		store.Status.State = state
		if err := r.client.Status().Update(ctx, store); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *MultiRaftStoreReconciler) reconcileGroup(ctx context.Context, store *storagev2beta2.MultiRaftStore, groupID int) (*storagev2beta2.RaftGroup, bool, error) {
	groupName := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      fmt.Sprintf("%s-%d", store.Name, groupID),
	}
	group := &storagev2beta2.RaftGroup{}
	if err := r.client.Get(ctx, groupName, group); err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, false, err
		}

		group = &storagev2beta2.RaftGroup{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: groupName.Namespace,
				Name:      groupName.Name,
				Labels:    store.Labels,
			},
			Spec: storagev2beta2.RaftGroupSpec{
				RaftConfig: store.Spec.RaftConfig,
			},
		}
		if err := controllerutil.SetControllerReference(store, group, r.scheme); err != nil {
			return nil, false, err
		}
		if err := r.client.Create(ctx, group); err != nil {
			return nil, false, err
		}
		return group, true, nil
	}

	if ok, err := r.reconcileMembers(ctx, store, group, groupID); err != nil {
		return group, false, err
	} else if ok {
		return group, true, nil
	}
	return group, false, nil
}

func (r *MultiRaftStoreReconciler) reconcileMembers(ctx context.Context, store *storagev2beta2.MultiRaftStore, group *storagev2beta2.RaftGroup, groupID int) (bool, error) {
	state := storagev2beta2.RaftGroupReady
	for memberID := 1; memberID <= getNumMembers(store); memberID++ {
		if member, ok, err := r.reconcileMember(ctx, store, group, groupID, memberID); err != nil {
			return false, err
		} else if ok {
			return true, nil
		} else if member.Status.State == storagev2beta2.RaftMemberNotReady {
			state = storagev2beta2.RaftGroupNotReady
		}
	}

	if group.Status.State != state {
		group.Status.State = state
		if err := r.client.Status().Update(ctx, group); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *MultiRaftStoreReconciler) reconcileMember(ctx context.Context, store *storagev2beta2.MultiRaftStore, group *storagev2beta2.RaftGroup, groupID int, memberID int) (*storagev2beta2.RaftMember, bool, error) {
	memberName := types.NamespacedName{
		Namespace: group.Namespace,
		Name:      fmt.Sprintf("%s-%d", group.Name, memberID),
	}
	member := &storagev2beta2.RaftMember{}
	if err := r.client.Get(ctx, memberName, member); err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, false, err
		}

		var memberType storagev2beta2.RaftMemberType
		if memberID <= getNumVotingMembers(store) {
			memberType = storagev2beta2.RaftVotingMember
		} else if memberID <= getNumMembers(store) {
			memberType = storagev2beta2.RaftObserver
		} else {
			memberType = storagev2beta2.RaftWitness
		}

		member = &storagev2beta2.RaftMember{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: memberName.Namespace,
				Name:      memberName.Name,
				Labels:    member.Labels,
			},
			Spec: storagev2beta2.RaftMemberSpec{
				Pod: corev1.LocalObjectReference{
					Name: getPodName(store, groupID, memberID),
				},
				Type: memberType,
			},
		}
		if err := controllerutil.SetControllerReference(store, member, r.scheme); err != nil {
			return nil, false, err
		}
		if err := controllerutil.SetOwnerReference(group, member, r.scheme); err != nil {
			return nil, false, err
		}
		if err := r.client.Create(ctx, member); err != nil {
			return nil, false, err
		}
		return member, true, nil
	}

	podName := types.NamespacedName{
		Namespace: member.Namespace,
		Name:      member.Spec.Pod.Name,
	}
	pod := &corev1.Pod{}
	if err := r.client.Get(ctx, podName, pod); err != nil {
		return nil, false, err
	}

	var containerVersion int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == nodeContainerName {
			containerVersion = containerStatus.RestartCount + 1
			break
		}
	}

	if member.Status.Version == nil || containerVersion > *member.Status.Version {
		if member.Status.State != storagev2beta2.RaftMemberNotReady {
			member.Status.State = storagev2beta2.RaftMemberNotReady
			if err := r.client.Status().Update(ctx, member); err != nil {
				return nil, false, err
			}
			r.events.Eventf(member, "Normal", "StateChanged", "State changed to %s", member.Status.State)
			return member, true, nil
		}

		address := fmt.Sprintf("%s:%d", pod.Status.PodIP, apiPort)
		conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, false, err
		}
		defer conn.Close()

		members := make([]multiraftv1.MemberConfig, getNumMembers(store))
		for i := 0; i < getNumMembers(store); i++ {
			members[i] = multiraftv1.MemberConfig{
				MemberID: multiraftv1.MemberID(i + 1),
				Host:     getPodDNSName(store.Namespace, store.Name, getPodName(store, groupID, i+1)),
				Port:     protocolPort,
			}
		}

		var role multiraftv1.MemberRole
		switch member.Spec.Type {
		case storagev2beta2.RaftVotingMember:
			role = multiraftv1.MemberRole_MEMBER
		case storagev2beta2.RaftObserver:
			role = multiraftv1.MemberRole_OBSERVER
		case storagev2beta2.RaftWitness:
			role = multiraftv1.MemberRole_WITNESS
		}

		client := multiraftv1.NewNodeClient(conn)
		request := &multiraftv1.BootstrapRequest{
			Group: multiraftv1.GroupConfig{
				GroupID:  multiraftv1.GroupID(groupID),
				MemberID: multiraftv1.MemberID(memberID),
				Role:     role,
				Members:  members,
			},
		}
		if _, err := client.Bootstrap(ctx, request); err != nil {
			return nil, false, err
		}

		member.Status.Version = &containerVersion
		if err := r.client.Status().Update(ctx, member); err != nil {
			return nil, false, err
		}
		return member, true, nil
	}
	return member, false, nil
}

func (r *MultiRaftStoreReconciler) reconcileStore(ctx context.Context, store *storagev2beta2.MultiRaftStore) error {
	partitions, err := r.getPartitions(ctx, store)
	if err != nil {
		return err
	}

	if store.Status.Partitions != nil && isPartitionsEqual(store.Status.Partitions, partitions) {
		return nil
	}

	config := getConfig(partitions)
	marshaler := &jsonpb.Marshaler{}
	configString, err := marshaler.MarshalToString(&config)
	if err != nil {
		return err
	}

	atomixStore := &v1beta1.Store{}
	atomixStoreName := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      store.Name,
	}
	if err := r.client.Get(ctx, atomixStoreName, atomixStore); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}

		atomixStore = &v1beta1.Store{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: atomixStoreName.Namespace,
				Name:      atomixStoreName.Name,
				Labels:    store.Labels,
			},
			Spec: v1beta1.StoreSpec{
				Driver: v1beta1.Driver{
					Name:    driverName,
					Version: driverVersion,
				},
				Config: runtime.RawExtension{
					Raw: []byte(configString),
				},
			},
		}
		if err := controllerutil.SetControllerReference(store, atomixStore, r.scheme); err != nil {
			return err
		}
		if err := r.client.Create(ctx, atomixStore); err != nil {
			return err
		}
		return nil
	}

	atomixStore.Spec.Config = runtime.RawExtension{
		Raw: []byte(configString),
	}
	if err := r.client.Update(ctx, atomixStore); err != nil {
		return err
	}

	store.Status.Partitions = partitions
	if err := r.client.Status().Update(ctx, store); err != nil {
		return err
	}
	return nil
}

func isPartitionsEqual(partitions1, partitions2 []storagev2beta2.RaftPartitionStatus) bool {
	if len(partitions1) != len(partitions2) {
		return false
	}
	for i := 0; i < len(partitions1); i++ {
		if !isPartitionEqual(partitions1[i], partitions2[i]) {
			return false
		}
	}
	return true
}

func isPartitionEqual(partition1, partition2 storagev2beta2.RaftPartitionStatus) bool {
	if partition1.Leader == nil && partition2.Leader != nil {
		return false
	}
	if partition1.Leader != nil && partition2.Leader == nil {
		return false
	}
	if partition1.Leader != nil && partition2.Leader != nil && *partition1.Leader != *partition2.Leader {
		return false
	}
	if len(partition1.Followers) != len(partition2.Followers) {
		return false
	}
	for _, follower1 := range partition1.Followers {
		match := false
		for _, follower2 := range partition2.Followers {
			if follower1 == follower2 {
				match = true
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func getConfig(partitions []storagev2beta2.RaftPartitionStatus) multiraftv1.DriverConfig {
	var config multiraftv1.DriverConfig
	for _, partition := range partitions {
		var leader string
		if partition.Leader != nil {
			leader = *partition.Leader
		}
		config.Partitions = append(config.Partitions, multiraftv1.PartitionConfig{
			PartitionID: multiraftv1.PartitionID(partition.PartitionID),
			Leader:      leader,
			Followers:   partition.Followers,
		})
	}
	return config
}

func (r *MultiRaftStoreReconciler) getPartitions(ctx context.Context, store *storagev2beta2.MultiRaftStore) ([]storagev2beta2.RaftPartitionStatus, error) {
	numGroups := getNumGroups(store)
	partitions := make([]storagev2beta2.RaftPartitionStatus, 0, numGroups)
	for groupID := 1; groupID <= numGroups; groupID++ {
		groupName := types.NamespacedName{
			Namespace: store.Namespace,
			Name:      fmt.Sprintf("%s-%d", store.Name, groupID),
		}
		group := &storagev2beta2.RaftGroup{}
		if err := r.client.Get(ctx, groupName, group); err != nil {
			return nil, err
		}

		partition := storagev2beta2.RaftPartitionStatus{
			PartitionID: int32(groupID),
		}
		if group.Status.Leader != nil {
			memberName := types.NamespacedName{
				Namespace: group.Namespace,
				Name:      group.Status.Leader.Name,
			}
			member := &storagev2beta2.RaftMember{}
			if err := r.client.Get(ctx, memberName, member); err != nil {
				return nil, err
			}
			address := fmt.Sprintf("%s:%d", getPodDNSName(store.Namespace, store.Name, member.Spec.Pod.Name), apiPort)
			partition.Leader = &address
		}

		for _, follower := range group.Status.Followers {
			memberName := types.NamespacedName{
				Namespace: group.Namespace,
				Name:      follower.Name,
			}
			member := &storagev2beta2.RaftMember{}
			if err := r.client.Get(ctx, memberName, member); err != nil {
				return nil, err
			}
			address := fmt.Sprintf("%s:%d", getPodDNSName(store.Namespace, store.Name, member.Spec.Pod.Name), apiPort)
			partition.Followers = append(partition.Followers, address)
		}
		partitions = append(partitions, partition)
	}
	return partitions, nil
}

var _ reconcile.Reconciler = (*MultiRaftStoreReconciler)(nil)

func getNumGroups(store *storagev2beta2.MultiRaftStore) int {
	if store.Spec.Groups == 0 {
		return 1
	}
	return int(store.Spec.Groups)
}

func getNumReplicas(store *storagev2beta2.MultiRaftStore) int {
	if store.Spec.Replicas == 0 {
		return 1
	}
	return int(store.Spec.Replicas)
}

func getNumMembers(store *storagev2beta2.MultiRaftStore) int {
	return getNumVotingMembers(store) + getNumNonVotingMembers(store)
}

func getNumVotingMembers(store *storagev2beta2.MultiRaftStore) int {
	if store.Spec.RaftConfig.QuorumSize == nil {
		return getNumReplicas(store)
	}
	return int(*store.Spec.RaftConfig.QuorumSize)
}

func getNumNonVotingMembers(store *storagev2beta2.MultiRaftStore) int {
	if store.Spec.RaftConfig.ReadReplicas == nil {
		return 0
	}
	return int(*store.Spec.RaftConfig.ReadReplicas)
}

// getStoreResourceName returns the given resource name for the given store
func getStoreResourceName(name string, resource string) string {
	return fmt.Sprintf("%s-%s", name, resource)
}

// getStoreHeadlessServiceName returns the headless service name for the given store
func getStoreHeadlessServiceName(cluster string) string {
	return getStoreResourceName(cluster, headlessServiceSuffix)
}

// getClusterDomain returns Kubernetes cluster domain, default to "cluster.local"
func getClusterDomain() string {
	clusterDomain := os.Getenv(clusterDomainEnv)
	if clusterDomain == "" {
		apiSvc := "kubernetes.default.svc"
		cname, err := net.LookupCNAME(apiSvc)
		if err != nil {
			return "cluster.local"
		}
		clusterDomain = strings.TrimSuffix(strings.TrimPrefix(cname, apiSvc+"."), ".")
	}
	return clusterDomain
}

// getPodDNSName returns the fully qualified DNS name for the given pod ID
func getPodDNSName(namespace string, store string, name string) string {
	return fmt.Sprintf("%s.%s.%s.svc.%s", name, getStoreHeadlessServiceName(store), namespace, getClusterDomain())
}

func getPodName(store *storagev2beta2.MultiRaftStore, groupID int, memberID int) string {
	podOrdinal := ((getNumMembers(store) * groupID) + (memberID - 1)) % getNumReplicas(store)
	return fmt.Sprintf("%s-%d", store.Name, podOrdinal)
}

// newStoreLabels returns the labels for the given cluster
func newStoreLabels(store *storagev2beta2.MultiRaftStore) map[string]string {
	labels := make(map[string]string)
	for key, value := range store.Labels {
		labels[key] = value
	}
	labels[appLabel] = appAtomix
	labels[storeLabel] = store.Name
	return labels
}

func newStoreAnnotations(store *storagev2beta2.MultiRaftStore) map[string]string {
	annotations := make(map[string]string)
	for key, value := range store.Annotations {
		annotations[key] = value
	}
	annotations[multiRaftStoreAnnotation] = store.Name
	return annotations
}

func getImage(store *storagev2beta2.MultiRaftStore) string {
	if store.Spec.Image != "" {
		return store.Spec.Image
	}
	return getDefaultImage()
}

func getDefaultImage() string {
	image := os.Getenv(defaultImageEnv)
	if image == "" {
		image = defaultImage
	}
	return image
}
