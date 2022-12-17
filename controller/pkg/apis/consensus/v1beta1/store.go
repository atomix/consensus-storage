// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConsensusStoreSpec specifies a ConsensusStore configuration
type ConsensusStoreSpec struct {
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

	// Config is the consensus store configuration
	Config ConsensusStoreConfig `json:"config,omitempty"`
}

type ConsensusStoreConfig struct {
	// Server is the consensus server configuration
	Server ServerConfig `json:"server,omitempty"`

	// Raft is the Raft protocol configuration
	Raft RaftConfig `json:"raft,omitempty"`

	// Logging is the store logging configuration
	Logging LoggingConfig `json:"logging,omitempty"`
}

type ServerConfig struct {
	ReadBufferSize       *int               `json:"readBufferSize"`
	WriteBufferSize      *int               `json:"writeBufferSize"`
	MaxRecvMsgSize       *resource.Quantity `json:"maxRecvMsgSize"`
	MaxSendMsgSize       *resource.Quantity `json:"maxSendMsgSize"`
	NumStreamWorkers     *uint32            `json:"numStreamWorkers"`
	MaxConcurrentStreams *uint32            `json:"maxConcurrentStreams"`
}

// ConsensusStoreState is a state constant for ConsensusStore
type ConsensusStoreState string

const (
	// ConsensusStoreNotReady indicates a ConsensusStore is not yet ready
	ConsensusStoreNotReady ConsensusStoreState = "NotReady"
	// ConsensusStoreReady indicates a ConsensusStore is ready
	ConsensusStoreReady ConsensusStoreState = "Ready"
)

type ConsensusStoreStatus struct {
	State ConsensusStoreState `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConsensusStore is the Schema for the ConsensusStore API
// +k8s:openapi-gen=true
type ConsensusStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ConsensusStoreSpec   `json:"spec,omitempty"`
	Status            ConsensusStoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConsensusStoreList contains a list of ConsensusStore
type ConsensusStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the ConsensusStore of items in the list
	Items []ConsensusStore `json:"items"`
}
