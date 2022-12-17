// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1beta2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RaftPartitionState is a state constant for RaftPartition
type RaftPartitionState string

const (
	// RaftPartitionNotReady indicates a RaftPartition is not yet ready
	RaftPartitionNotReady RaftPartitionState = "NotReady"
	// RaftPartitionReady indicates a RaftPartition is ready
	RaftPartitionReady RaftPartitionState = "Ready"
)

// RaftPartitionSpec specifies a RaftPartitionSpec configuration
type RaftPartitionSpec struct {
	RaftConfig `json:",inline"`
	Cluster    corev1.LocalObjectReference `json:"cluster"`
	Replicas   uint32                      `json:"replicas"`
	Shard      uint32                      `json:"shard"`
	Ordinal    uint32                      `json:"ordinal"`
}

// RaftPartitionStatus defines the status of a RaftPartition
type RaftPartitionStatus struct {
	State     RaftPartitionState            `json:"state,omitempty"`
	Term      *uint64                       `json:"term,omitempty"`
	Leader    *corev1.LocalObjectReference  `json:"leader,omitempty"`
	Followers []corev1.LocalObjectReference `json:"followers,omitempty"`
	Members   []RaftMemberReference         `json:"members"`
}

type RaftMemberReference struct {
	Pod     corev1.LocalObjectReference `json:"pod"`
	Ordinal uint32                      `json:"ordinal"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RaftPartition is the Schema for the RaftPartition API
// +k8s:openapi-gen=true
type RaftPartition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RaftPartitionSpec   `json:"spec,omitempty"`
	Status            RaftPartitionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RaftPartitionList contains a list of RaftPartition
type RaftPartitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the RaftPartition of items in the list
	Items []RaftPartition `json:"items"`
}
