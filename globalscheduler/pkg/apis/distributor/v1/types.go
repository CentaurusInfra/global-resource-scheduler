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

// DistributorSpec defines the desired state of Distributor
type DistributorSpec struct {
	Policy DistributorPolicy `json:"policy,omitempty"`
}

// DistributorPolicy describes how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type DistributorPolicy string

const (
	// GeoLocation  allows distributors to assign schedulers to pods based on geoLocation
	GeoLocation DistributorPolicy = "GeoLocation"
)

// DistributorStatus defines the observed state of Distributor
type DistributorStatus struct {
	// Information when was the last time the job was successfully scheduled.
	// +optional
	Active *bool `json:"active,omitempty"`
	// Information when was the last time the job was successfully scheduled.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Distributor is the Schema for the distributors API
type Distributor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DistributorSpec   `json:"spec,omitempty"`
	Status DistributorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DistributorList is a list of Distributor resources.
type DistributorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Distributor `json:"items"`
}
