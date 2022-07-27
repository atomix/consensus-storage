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
	apiPort               = 5678
	protocolPortName      = "raft"
	protocolPort          = 5679
	probePort             = 5679
	defaultImageEnv       = "DEFAULT_NODE_V2BETA1_IMAGE"
	defaultImage          = "atomix/atomix-raft-storage-node:latest"
	headlessServiceSuffix = "hs"
	appLabel              = "app"
	clusterLabel          = "cluster"
	appAtomix             = "atomix"
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

const monitoringPort = 5000

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

	// Watch for changes to secondary resource MultiRaftCluster
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.MultiRaftCluster{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &storagev2beta2.MultiRaftStore{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RaftGroup
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftGroup{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		group := object.(*storagev2beta2.RaftGroup)
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Namespace: group.Namespace,
					Name:      group.Spec.Cluster,
				},
			},
		}
	}))
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

	err = r.reconcileConfigMap(ctx, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileStatefulSet(ctx, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileService(ctx, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileHeadlessService(ctx, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	ok, err := r.reconcileCluster(ctx, store)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}

	ok, err = r.reconcileStore(ctx, store)
	if err != nil {
		log.Error(err, "Reconcile MultiRaftStore")
		return reconcile.Result{}, err
	} else if ok {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *MultiRaftStoreReconciler) reconcileConfigMap(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol config map")
	cm := &corev1.ConfigMap{}
	name := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}
	err := r.client.Get(ctx, name, cm)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addConfigMap(ctx, cluster)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addConfigMap(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating raft ConfigMap", "Name", cluster.Name, "Namespace", cluster.Namespace)
	raftConfig, err := newRaftConfigString(cluster)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    newStoreLabels(cluster),
		},
		Data: map[string]string{
			raftConfigFile: raftConfig,
		},
	}

	if err := controllerutil.SetControllerReference(cluster, cm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, cm)
}

// getClusterConfig creates a node configuration string for the given cluster
func getClusterConfig(store *storagev2beta2.MultiRaftStore) storagev2beta2.MultiRaftClusterConfig {
	numNodes := getNumReplicas(store)
	nodes := make([]storagev2beta2.MultiRaftNodeConfig, numNodes)
	for i := 0; i < numNodes; i++ {
		nodeID := i + 1
		nodes[i] = storagev2beta2.MultiRaftNodeConfig{
			NodeID: int32(nodeID),
		}
	}

	numGroups := getNumPartitions(store)
	groups := make([]storagev2beta2.RaftGroupConfig, numGroups)
	for i := 0; i < numGroups; i++ {
		groupID := i + 1
		numMembers := getNumMembers(store)
		numROMembers := getNumROMembers(store)
		members := make([]storagev2beta2.RaftMemberConfig, numMembers+numROMembers)
		for j := 0; j < numMembers; j++ {
			members[j] = storagev2beta2.RaftMemberConfig{
				NodeID: int32((((numMembers+numROMembers)*groupID)+j)%numNodes) + 1,
				Type:   storagev2beta2.RaftVotingMember,
			}
		}
		for j := numMembers; j < numMembers+numROMembers; j++ {
			members[j] = storagev2beta2.RaftMemberConfig{
				NodeID: int32((((numMembers+numROMembers)*groupID)+j)%numNodes) + 1,
				Type:   storagev2beta2.RaftObserver,
			}
		}
		groups[i] = storagev2beta2.RaftGroupConfig{
			GroupID: int32(groupID),
			Members: members,
		}
	}

	return storagev2beta2.MultiRaftClusterConfig{
		Nodes:  nodes,
		Groups: groups,
	}
}

// getDriverConfig creates a driver configuration for the given cluster
func (r *MultiRaftStoreReconciler) getDriverConfig(ctx context.Context, store *storagev2beta2.MultiRaftStore) (*multiraftv1.DriverConfig, error) {
	numPartitions := getNumPartitions(store)
	partitions := make([]multiraftv1.PartitionConfig, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitionID := i + 1
		partition := multiraftv1.PartitionConfig{
			PartitionID: multiraftv1.PartitionID(partitionID),
		}
		groupName := types.NamespacedName{
			Namespace: store.Namespace,
			Name:      fmt.Sprintf("%s-%d", store.Name, partitionID),
		}
		group := &storagev2beta2.RaftGroup{}
		if err := r.client.Get(ctx, groupName, group); err == nil {
			if group.Status.Leader != nil {
				partition.Leader = fmt.Sprintf(getPodDNSName(store.Namespace, store.Name, int(*group.Status.Leader)-1), apiPort)
			}
			for _, member := range group.Spec.Members {
				if group.Status.Leader == nil || member.NodeID != *group.Status.Leader {
					partition.Followers = append(partition.Followers, fmt.Sprintf(getPodDNSName(store.Namespace, store.Name, int(member.NodeID-1)), apiPort))
				}
			}
		} else if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		partitions[i] = partition
	}
	return &multiraftv1.DriverConfig{
		Partitions: partitions,
	}, nil
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
			Name:      store.Name,
			Namespace: store.Namespace,
			Labels:    newStoreLabels(store),
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
					Labels: newStoreLabels(store),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "raft",
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
								`set -ex
[[ ` + "`hostname`" + ` =~ -([0-9]+)$ ]] || exit 1
ordinal=${BASH_REMATCH[1]}
atomix-multi-raft-node $ordinal`,
							},
							Args: []string{
								"--config",
								fmt.Sprintf("%s/%s", configPath, raftConfigFile),
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"stat", "/tmp/atomix-ready"},
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
			Name:      store.Name,
			Namespace: store.Namespace,
			Labels:    newStoreLabels(store),
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

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getStoreHeadlessServiceName(store.Name),
			Namespace: store.Namespace,
			Labels:    newStoreLabels(store),
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
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

func (r *MultiRaftStoreReconciler) reconcileCluster(ctx context.Context, store *storagev2beta2.MultiRaftStore) (bool, error) {
	cluster := &storagev2beta2.MultiRaftCluster{}
	clusterName := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      store.Name,
	}
	if err := r.client.Get(ctx, clusterName, cluster); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}
		cluster = &storagev2beta2.MultiRaftCluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: clusterName.Namespace,
				Name:      clusterName.Name,
				Labels:    store.Labels,
			},
			Spec: storagev2beta2.MultiRaftClusterSpec{
				MultiRaftClusterConfig: getClusterConfig(store),
			},
		}
		if err := controllerutil.SetControllerReference(store, cluster, r.scheme); err != nil {
			return false, err
		}
		if err := r.client.Create(ctx, cluster); err != nil {
			return false, err
		}
		return true, nil
	}

	switch cluster.Status.State {
	case storagev2beta2.MultiRaftClusterReady:
		if store.Status.State != storagev2beta2.MultiRaftStoreReady {
			store.Status.State = storagev2beta2.MultiRaftStoreReady
			if err := r.client.Status().Update(ctx, store); err != nil {
				return false, err
			}
			return true, nil
		}
	case storagev2beta2.MultiRaftClusterNotReady:
		if store.Status.State != storagev2beta2.MultiRaftStoreNotReady {
			store.Status.State = storagev2beta2.MultiRaftStoreNotReady
			if err := r.client.Status().Update(ctx, store); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (r *MultiRaftStoreReconciler) reconcileStore(ctx context.Context, multiRaftStore *storagev2beta2.MultiRaftStore) (bool, error) {
	store := &v1beta1.Store{}
	storeName := types.NamespacedName{
		Namespace: multiRaftStore.Namespace,
		Name:      multiRaftStore.Name,
	}
	if err := r.client.Get(ctx, storeName, store); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}

		config, err := r.getDriverConfig(ctx, multiRaftStore)
		if err != nil {
			return false, err
		}

		marshaler := &jsonpb.Marshaler{}
		configString, err := marshaler.MarshalToString(config)
		if err != nil {
			return false, err
		}

		store = &v1beta1.Store{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: storeName.Namespace,
				Name:      storeName.Name,
				Labels:    multiRaftStore.Labels,
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
		if err := controllerutil.SetControllerReference(multiRaftStore, store, r.scheme); err != nil {
			return false, err
		}
		if err := r.client.Create(ctx, store); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

var _ reconcile.Reconciler = (*MultiRaftStoreReconciler)(nil)

func getNumPartitions(store *storagev2beta2.MultiRaftStore) int {
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
	if store.Spec.RaftConfig.QuorumSize == nil {
		return getNumReplicas(store)
	}
	return int(*store.Spec.RaftConfig.QuorumSize)
}

func getNumROMembers(store *storagev2beta2.MultiRaftStore) int {
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
func getPodDNSName(namespace string, cluster string, podID int) string {
	return fmt.Sprintf("%s-%d.%s.%s.svc.%s", cluster, podID, getStoreHeadlessServiceName(cluster), namespace, getClusterDomain())
}

// newStoreLabels returns the labels for the given cluster
func newStoreLabels(store *storagev2beta2.MultiRaftStore) map[string]string {
	labels := make(map[string]string)
	for key, value := range store.Labels {
		labels[key] = value
	}
	labels[appLabel] = appAtomix
	labels[clusterLabel] = store.Name
	return labels
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
