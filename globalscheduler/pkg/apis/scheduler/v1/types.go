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
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

type SchedulerStatus string

const (
	SchedulerActive SchedulerStatus = "Active"
	SchedulerDelete SchedulerStatus = "Delete"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Scheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchedulerSpec   `json:"spec"`
	Status SchedulerStatus `json:"status"`
}

type SchedulerSpec struct {
	// Location represent which geo location the scheduler is responsable for
	Location GeolocationInfo `json:"location"`

	// Tag represent the Nth scheduler object
	Tag string `json:"tag"`

	// Cluster is an array that stores the name of clusters
	Cluster []string `json:"cluster"`

	Union ClusterUnion `json:"union"`

	IpAddress string `json:"ipAddress"`

	PortNumber string `json:"portNumber"`
}

type GeolocationInfo struct {
	City     string `json:"city"`
	Province string `json:"province"`
	Area     string `json:"area"`
	Country  string `json:"country"`
}

type ClusterUnion struct {
	GeoLocation []*clustercrdv1.GeolocationInfo `json:"geolocation"`
	Region      []*clustercrdv1.RegionInfo      `json:"region"`
	Operator    []*clustercrdv1.OperatorInfo    `json:"operator"`
	Flavors     []*clustercrdv1.FlavorInfo      `json:"flavors"`
	Storage     []*clustercrdv1.StorageSpec     `json:"storage"`
	EipCapacity []int64                         `json:"eipcapacity"`
	CPUCapacity []int64                         `json:"cpucapacity"`
	MemCapacity []int64                         `json:"memcapacity"`
	ServerPrice []int64                         `json:"serverprice"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Scheduler `json:"items"`
}
