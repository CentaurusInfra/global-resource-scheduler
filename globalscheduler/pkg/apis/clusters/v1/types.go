package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// These const variables are used in our custom controller.
const (
	GroupName string = "globalscheduler.com"
	Kind      string = "Cluster"
	Version   string = "v1"
	Plural    string = "clusters"
	Singluar  string = "cluster"
	ShortName string = "cluster"
	Name      string = Plural + "." + GroupName
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
