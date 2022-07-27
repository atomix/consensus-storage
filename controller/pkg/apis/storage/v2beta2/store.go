// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MultiRaftStoreState is a state constant for MultiRaftStore
type MultiRaftStoreState string

const (
	// MultiRaftStoreNotReady indicates a MultiRaftStore is not yet ready
	MultiRaftStoreNotReady MultiRaftStoreState = "NotReady"
	// MultiRaftStoreReady indicates a MultiRaftStore is ready
	MultiRaftStoreReady MultiRaftStoreState = "Ready"
)

// MultiRaftStoreSpec specifies a MultiRaftStore configuration
type MultiRaftStoreSpec struct {
	// Replicas is the number of raft replicas
	Replicas int32 `json:"replicas,omitempty"`

	// Groups is the number of groups
	Groups int32 `json:"groups,omitempty"`

	// Image is the image to run
	Image string `json:"image,omitempty"`

	// ImagePullPolicy is the pull policy to apply
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets is a list of secrets for pulling images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// SecurityContext is a pod security context
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// VolumeClaimTemplate is the volume claim template for Raft logs
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// Raft is the Raft protocol configuration
	RaftConfig RaftConfig `json:"raftConfig,omitempty"`
}

// MultiRaftStoreStatus defines the status of a MultiRaftStore
type MultiRaftStoreStatus struct {
	State MultiRaftStoreState `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiRaftStore is the Schema for the MultiRaftStore API
// +k8s:openapi-gen=true
type MultiRaftStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MultiRaftStoreSpec   `json:"spec,omitempty"`
	Status            MultiRaftStoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiRaftStoreList contains a list of MultiRaftStore
type MultiRaftStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the MultiRaftStore of items in the list
	Items []MultiRaftStore `json:"items"`
}
