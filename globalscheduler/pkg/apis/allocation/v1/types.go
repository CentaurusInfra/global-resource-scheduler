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

const (
	GroupName           string          = "globalscheduler.com"
	Kind                string          = "Allocation"
	Version             string          = "v1"
	Plural              string          = "allocations"
	Singluar            string          = "allocation"
	ShortName           string          = "alloc"
	Name                string          = Plural + "." + GroupName
	AllocationAssigned  AllocationPhase = "Assigned"
	AllocationBound     AllocationPhase = "Bound"
	AllocationScheduled AllocationPhase = "Scheduled"
	AllocationFailed    AllocationPhase = "Failed"
)

// AllocationPhase is a label for the condition of an allocation at the current time.
type AllocationPhase string

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Allocation describes an allocation custom resource.
type Allocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AllocationSpec   `json:"spec"`
	Status            AllocationStatus `json:"status"`
}

// AllocationSpec is the spec for an allocation resource.
type AllocationSpec struct {
	ResourceGroup ResourceGroup `json:"resource_group"`
	Selector      Selector      `json:"selector"`
	Replicas      int           `json:"replicas"` // min value 1, default value 1
}

// AllocationStatus is the status for an allocation resource.
type AllocationStatus struct {
	Phase           AllocationPhase `json:"phase"`
	DistributorName string          `json:"distributor_name"`
	DispatcherName  string          `json:"dispatcher_name"`
	SchedulerName   string          `json:"scheduler_name"`
	ClusterNames    []string        `json:"cluster_names"`
}

type ResourceGroup struct {
	Name      string      `json:"name"`
	Resources []Resources `json:"resources"`
}

type Resources struct {
	Name         string   `json:"name"`
	ResourceType string   `json:"resource_type"` // vm
	Flavors      []Flavor `json:"flavors"`
	Storage      Volume   `json:"storage,omitempty"`
	NeedEip      bool     `json:"need_eip,omitempty"`
	//Count        int      `json:"count,omitempty"` // It is commented because there is a conflict with replicas
	Image           string `json:"image"`
	SecurityGroupId string `json:"security_group_id"`
	NicName         string `json:"nic_name"`
}

type Flavor struct {
	FlavorId string `json:"flavor_id,omitempty"` // default value c6.large.2
	Spot     Spot   `json:"spot,omitempty"`
}

type Spot struct {
	MaxPrice           string `json:"max_price,omitempty"`           // default value 1.5
	SpotDurationHours  int    `json:"spot_duration_hours,omitempty"` // min value 0; max value 6; default value 1
	SpotDurationCount  int    `json:"spot_duration_count,omitempty"` // default value 2
	InterruptionPolicy string `json:"interruption_policy,omitempty"` // default value immediate
}

type Volume struct {
	SATA int `json:"sata,omitempty"` // default value 40
	SAS  int `json:"sas,omitempty"`  // default value 10
	SSD  int `json:"ssd,omitempty"`  // default value 10
}

type Selector struct {
	GeoLocation GeoLocation `json:"geo_location,omitempty"`
	Regions     []Region    `json:"regions,omitempty"`
	Operator    string      `json:"operator,omitempty"` // chinatelecom, chinaunicom, and chinamobile
	Strategy    Strategy    `json:"strategy,omitempty"`
}

type GeoLocation struct {
	City     string `json:"city"`
	Province string `json:"province"`
	Area     string `json:"area"`
	Country  string `json:"country"`
}

type Region struct {
	Region           string   `json:"region"`
	AvailabilityZone []string `json:"availability_zone,omitempty"`
}

type Strategy struct {
	LocationStrategy string `json:"location_strategy,omitempty"` // centralize and discrete, default value centralize
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AllocationList is a list of Allocation resources.
type AllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Allocation `json:"items"`
}
