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
	"k8s.io/kubernetes/globalscheduler/pkg/apis/cluster"
)

// These const variables are used in our custom controller.
const (
	Kind      string = "Cluster"
	Plural    string = "clusters"
	Singluar  string = "cluster"
	ShortName string = "cluster"
	Name      string = Plural + "." + cluster.GroupName
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Cluster describes a Cluster custom resource.
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec `json:"spec"`
	Status            string      `json:"status"`
}

// ClusterSpec is the spec for a Cluster resource.
//This is where you would put your custom resource data
type ClusterSpec struct {
	IpAdrress     string          `json:"ipaddress"`
	GeoLocation   GeolocationInfo `json:"geolocation"`
	Region        RegionInfo      `json:"region"`
	Operator      OperatorInfo    `json:"operator"`
	Flavors       []FlavorInfo    `json:"flavors"`
	Storage       []StorageSpec   `json:"storage"`
	EipCapacity   int64           `json:"eipcapacity"`
	CPUCapacity   int64           `json:"cpucapacity"`
	MemCapacity   int64           `json:"memcapacity"`
	ServerPrice   int64           `json:"serverprice"`
	HomeScheduler string          `json:"homescheduler"`
}
type FlavorInfo struct {
	FlavorID      string `json:"flavorid"`
	TotalCapacity int64  `json:"totalcapacity"`
}
type StorageSpec struct {
	TypeID          string `json:"typeid"` //(sata, sas, ssd)
	StorageCapacity int64  `json:"storagecapacity"`
}
type GeolocationInfo struct {
	City     string `json:"city"`
	Province string `json:"province"`
	Area     string `json:"area"`
	Country  string `json:"country"`
}
type RegionInfo struct {
	Region           string `json:"region"`
	AvailabilityZone string `json:"availabilityzone"`
}
type OperatorInfo struct {
	Operator string `json:"operator"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClusterList is a list of Cluster resources.
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Cluster `json:"items"`
}
