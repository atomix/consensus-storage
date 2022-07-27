// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MultiRaftClusterState is a state constant for MultiRaftCluster
type MultiRaftClusterState string

const (
	// MultiRaftClusterNotReady indicates a MultiRaftCluster is not yet ready
	MultiRaftClusterNotReady MultiRaftClusterState = "NotReady"
	// MultiRaftClusterReady indicates a MultiRaftCluster is ready
	MultiRaftClusterReady MultiRaftClusterState = "Ready"
)

// MultiRaftClusterSpec specifies a MultiRaftClusterSpec configuration
type MultiRaftClusterSpec struct {
	MultiRaftClusterConfig `json:",inline"`
}

type MultiRaftClusterConfig struct {
	Nodes  []MultiRaftNodeConfig `json:"nodes,omitempty"`
	Groups []RaftGroupConfig     `json:"groups,omitempty"`
}

// MultiRaftClusterStatus defines the status of a RaftCluster
type MultiRaftClusterStatus struct {
	State MultiRaftClusterState `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiRaftCluster is the Schema for the RaftCluster API
// +k8s:openapi-gen=true
type MultiRaftCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MultiRaftClusterSpec   `json:"spec,omitempty"`
	Status            MultiRaftClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiRaftClusterList contains a list of MultiRaftCluster
type MultiRaftClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the RaftCluster of items in the list
	Items []MultiRaftCluster `json:"items"`
}
