// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MultiRaftNodeState is a state constant for MultiRaftNode
type MultiRaftNodeState string

const (
	// MultiRaftNodeNotReady indicates a MultiRaftNode is not yet ready
	MultiRaftNodeNotReady MultiRaftNodeState = "NotReady"
	// MultiRaftNodeReady indicates a MultiRaftNode is ready
	MultiRaftNodeReady MultiRaftNodeState = "Ready"
)

// MultiRaftNodeSpec specifies a MultiRaftNodeSpec configuration
type MultiRaftNodeSpec struct {
	MultiRaftNodeConfig `json:",inline"`
	Cluster             string `json:"cluster"`
}

type MultiRaftNodeConfig struct {
	NodeID int32 `json:"nodeID"`
}

// MultiRaftNodeStatus defines the status of a RaftCluster
type MultiRaftNodeStatus struct {
	State MultiRaftNodeState `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiRaftNode is the Schema for the RaftCluster API
// +k8s:openapi-gen=true
type MultiRaftNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MultiRaftNodeSpec   `json:"spec,omitempty"`
	Status            MultiRaftNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiRaftNodeList contains a list of MultiRaftNode
type MultiRaftNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the RaftCluster of items in the list
	Items []MultiRaftNode `json:"items"`
}
