/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DispatcherRange defines the upper and the lower bound of pod clustername for dispatchers to process
type DispatcherRange struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// DispatcherSpec defines the desired state of dispatcher
type DispatcherSpec struct {
	// Cluster is an array that stores the name of clusters
	ClusterRange DispatcherRange `json:"clusterRange,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Dispatcher is the Schema for the Dispatchers API
type Dispatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DispatcherSpec `json:"spec,omitempty"`
	Status string         `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DispatcherList is a list of Dispatcher resources.
type DispatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Dispatcher `json:"items"`
}
