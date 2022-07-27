// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	"context"
	"encoding/json"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
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
	controller, err := controller.New("atomix-raft-cluster-v2beta2", mgr, options)
	if err != nil {
		return err
	}

	// Watch for changes to the storage resource and enqueue Stores that reference it
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.MultiRaftStore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSet
	err = controller.Watch(&source.Kind{Type: &storagev2beta2.RaftGroup{}}, &handler.EnqueueRequestForOwner{
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
	for nodeID := 1; nodeID <= numNodes; nodeID++ {
		nodes[nodeID] = storagev2beta2.MultiRaftNodeConfig{
			NodeID: int32(nodeID),
		}
	}

	numGroups := getNumPartitions(store)
	groups := make([]storagev2beta2.RaftGroupConfig, numGroups)
	for groupID := 0; groupID < numGroups; groupID++ {
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
		groups[groupID] = storagev2beta2.RaftGroupConfig{
			GroupID: int32(groupID),
			Members: members,
		}
	}

	return storagev2beta2.MultiRaftClusterConfig{
		Nodes:  nodes,
		Groups: groups,
	}
}

// newRaftConfigString creates a protocol configuration string for the given cluster and protocol
func newRaftConfigString(cluster *storagev2beta2.MultiRaftStore) (string, error) {
	config := multiraftv1.MultiRaftConfig{}

	electionTimeout := cluster.Spec.RaftConfig.ElectionTimeout
	if electionTimeout != nil {
		config.ElectionTimeout = &electionTimeout.Duration
	}

	heartbeatPeriod := cluster.Spec.RaftConfig.HeartbeatPeriod
	if heartbeatPeriod != nil {
		config.HeartbeatPeriod = &heartbeatPeriod.Duration
	}

	entryThreshold := cluster.Spec.RaftConfig.SnapshotEntryThreshold
	if entryThreshold != nil {
		config.SnapshotEntryThreshold = uint64(*entryThreshold)
	} else {
		config.SnapshotEntryThreshold = 10000
	}

	retainEntries := cluster.Spec.RaftConfig.CompactionRetainEntries
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

func (r *MultiRaftStoreReconciler) reconcileStatefulSet(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol stateful set")
	statefulSet := &appsv1.StatefulSet{}
	name := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}
	err := r.client.Get(ctx, name, statefulSet)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addStatefulSet(ctx, cluster)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addStatefulSet(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating raft replicas", "Name", cluster.Name, "Namespace", cluster.Namespace)

	image := getImage(cluster)
	pullPolicy := cluster.Spec.ImagePullPolicy
	if pullPolicy == "" {
		pullPolicy = corev1.PullIfNotPresent
	}

	volumes := []corev1.Volume{
		{
			Name: configVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.Name,
					},
				},
			},
		},
	}

	var volumeClaimTemplates []corev1.PersistentVolumeClaim

	dataVolumeName := dataVolume
	if cluster.Spec.VolumeClaimTemplate != nil {
		pvc := cluster.Spec.VolumeClaimTemplate
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
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    newStoreLabels(cluster),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: getStoreHeadlessServiceName(cluster),
			Replicas:    pointer.Int32Ptr(int32(getNumReplicas(cluster))),
			Selector: &metav1.LabelSelector{
				MatchLabels: newStoreLabels(cluster),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: newStoreLabels(cluster),
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
							SecurityContext: cluster.Spec.SecurityContext,
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
											MatchLabels: newStoreLabels(cluster),
										},
										Namespaces:  []string{cluster.Namespace},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					ImagePullSecrets: cluster.Spec.ImagePullSecrets,
					Volumes:          volumes,
				},
			},
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}

	if err := controllerutil.SetControllerReference(cluster, set, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, set)
}

func (r *MultiRaftStoreReconciler) reconcileService(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol service")
	service := &corev1.Service{}
	name := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}
	err := r.client.Get(ctx, name, service)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addService(ctx, cluster)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addService(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating raft service", "Name", cluster.Name, "Namespace", cluster.Namespace)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    newStoreLabels(cluster),
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
			Selector: newStoreLabels(cluster),
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, service)
}

func (r *MultiRaftStoreReconciler) reconcileHeadlessService(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Reconcile raft protocol headless service")
	service := &corev1.Service{}
	name := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      getStoreHeadlessServiceName(cluster),
	}
	err := r.client.Get(ctx, name, service)
	if err != nil && k8serrors.IsNotFound(err) {
		err = r.addHeadlessService(ctx, cluster)
	}
	return err
}

func (r *MultiRaftStoreReconciler) addHeadlessService(ctx context.Context, cluster *storagev2beta2.MultiRaftStore) error {
	log.Info("Creating headless raft service", "Name", cluster.Name, "Namespace", cluster.Namespace)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getStoreHeadlessServiceName(cluster),
			Namespace: cluster.Namespace,
			Labels:    newStoreLabels(cluster),
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
			Selector:                 newStoreLabels(cluster),
		},
	}

	if err := controllerutil.SetControllerReference(cluster, service, r.scheme); err != nil {
		return err
	}
	return r.client.Create(ctx, service)
}

func (r *MultiRaftStoreReconciler) reconcileCluster(ctx context.Context, store *storagev2beta2.MultiRaftStore) (bool, error) {
	cluster := &storagev2beta2.MultiRaftCluster{}
	nodeName := types.NamespacedName{
		Namespace: store.Namespace,
		Name:      store.Name,
	}
	if err := r.client.Get(ctx, nodeName, cluster); err != nil {
		if !k8serrors.IsNotFound(err) {
			return false, err
		}
		cluster = &storagev2beta2.MultiRaftCluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nodeName.Namespace,
				Name:      nodeName.Name,
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

var _ reconcile.Reconciler = (*MultiRaftStoreReconciler)(nil)

func getNumPartitions(cluster *storagev2beta2.MultiRaftStore) int {
	if cluster.Spec.Groups == 0 {
		return 1
	}
	return int(cluster.Spec.Groups)
}

func getNumReplicas(cluster *storagev2beta2.MultiRaftStore) int {
	if cluster.Spec.Replicas == 0 {
		return 1
	}
	return int(cluster.Spec.Replicas)
}

func getNumMembers(cluster *storagev2beta2.MultiRaftStore) int {
	if cluster.Spec.RaftConfig.QuorumSize == nil {
		return getNumReplicas(cluster)
	}
	return int(*cluster.Spec.RaftConfig.QuorumSize)
}

func getNumROMembers(cluster *storagev2beta2.MultiRaftStore) int {
	if cluster.Spec.RaftConfig.ReadReplicas == nil {
		return 0
	}
	return int(*cluster.Spec.RaftConfig.ReadReplicas)
}

// getStoreResourceName returns the given resource name for the given cluster
func getStoreResourceName(cluster *storagev2beta2.MultiRaftStore, resource string) string {
	return fmt.Sprintf("%s-%s", cluster.Name, resource)
}

// getStoreHeadlessServiceName returns the headless service name for the given cluster
func getStoreHeadlessServiceName(cluster *storagev2beta2.MultiRaftStore) string {
	return getStoreResourceName(cluster, headlessServiceSuffix)
}

// GetStoreDomain returns Kubernetes cluster domain, default to "cluster.local"
func getStoreDomain() string {
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
func getPodDNSName(cluster *storagev2beta2.MultiRaftStore, podID int) string {
	return fmt.Sprintf("%s-%d.%s.%s.svc.%s", cluster.Name, podID, getStoreHeadlessServiceName(cluster), cluster.Namespace, getStoreDomain())
}

// newStoreLabels returns the labels for the given cluster
func newStoreLabels(cluster *storagev2beta2.MultiRaftStore) map[string]string {
	labels := make(map[string]string)
	for key, value := range cluster.Labels {
		labels[key] = value
	}
	labels[appLabel] = appAtomix
	labels[clusterLabel] = cluster.Name
	return labels
}

func getImage(cluster *storagev2beta2.MultiRaftStore) string {
	if cluster.Spec.Image != "" {
		return cluster.Spec.Image
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
