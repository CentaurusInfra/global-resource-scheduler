package typed

import (
	"encoding/json"
)

//ListOpts allows the filtering and sorting of paginated collections through the API.
type ListOpts struct {
	// Specifies the AZ name.
	AvailabilityZone string `q:"availability_zone"`
}

type Link struct {
	// Specifies the shortcut link marker name.
	Rel string `json:"rel"`

	// Provides the corresponding shortcut link.
	Href string `json:"href"`

	// Specifies the shortcut link type.
	Type string `json:"type"`
}

type SpotExtraSpecs struct {
	CondSpotBlockOperationAz     string `json:"cond:spot_block:operation:az"`
	CondSpotBlockLdh             string `json:"cond:spot_block:operation:longest_duration_hours"`
	CondSpotBlockLdc             string `json:"cond:spot_block:operation:longest_duration_count"`
	CondSpotBlockInterruptPolicy string `json:"cond:spot_block:operation:interrupt_policy"`
	CondSpotOperationAz          string `json:"cond:spot:operation:az"`
	CondSpotOperationStatus      string `json:"cond:spot:operation:status"`
}

type OsExtraSpecs struct {
	// Specifies the ECS specifications types
	PerformanceType string `json:"ecs:performancetype"`

	// Specifies the resource type.
	ResourceType string `json:"resource_type"`

	// Specifies the generation of an ECS type
	Generation string `json:"ecs:generation"`

	// Specifies a virtualization type
	VirtualizationEnvTypes string `json:"ecs:virtualization_env_types"`

	// Indicates whether the GPU is passthrough.
	PciPassthroughEnableGpu string `json:"pci_passthrough:enable_gpu"`

	// Indicates the technology used on the G1 and G2 ECSs,
	// including GPU virtualization and GPU passthrough.
	PciPassthroughGpuSpecs string `json:"pci_passthrough:gpu_specs"`

	// Indicates the model and quantity of passthrough-enabled GPUs on P1 ECSs.
	PciPassthroughAlias string `json:"pci_passthrough:alias"`

	// gpu info.wuzilin add
	InfoGPUName string `json:"info:gpu:name,omitempty"`

	CondOperationStatus string `json:"cond:operation:status"`

	CondOperationAz string `json:"cond:operation:az"`

	CondCompute string `json:"cond:compute"`

	CondImage string `json:"cond:image"`

	VifMaxNum string `json:"quota:vif_max_num"`

	PhysicsMaxRate string `json:"quota:physics_max_rate"`

	VifMultiqueueNum string `json:"quota:vif_multiqueue_num"`

	MinRate string `json:"quota:min_rate"`

	MaxRate string `json:"quota:max_rate"`

	MaxPps string `json:"quota:max_pps"`

	CPUSockets string `json:"hw:cpu_sockets"`

	NumaNodes string `json:"hw:numa_nodes"`

	CPUThreads string `json:"hw:cpu_threads"`

	MemPageSize string `json:"hw:mem_page_size"`

	ConnLimitTotal string `json:"quota:conn_limit_total"`

	CPUCores string `json:"hw:cpu_cores"`

	//add by laoyi
	SpotExtraSpecs
}

type Flavor struct {
	// Specifies the ID of ECS specifications.
	ID string `json:"id"`

	// Specifies the name of the ECS specifications.
	Name string `json:"name"`

	// Specifies the number of CPU cores in the ECS specifications.
	Vcpus string `json:"vcpus"`

	// Specifies the memory size (MB) in the ECS specifications.
	Ram int64 `json:"ram"`

	// Specifies the system disk size in the ECS specifications.
	// The value 0 indicates that the disk size is not limited.
	Disk string `json:"disk"`

	// Specifies shortcut links for ECS flavors.
	Links []Link `json:"links"`

	// Specifies extended ECS specifications.
	OsExtraSpecs OsExtraSpecs `json:"os_extra_specs"`

	// Reserved
	Swap string `json:"swap"`

	// Reserved
	FlvEphemeral int64 `json:"OS-FLV-EXT-DATA:ephemeral"`

	// Reserved
	FlvDisabled bool `json:"OS-FLV-DISABLED:disabled"`

	// Reserved
	RxtxFactor int64 `json:"rxtx_factor"`

	// Reserved
	RxtxQuota string `json:"rxtx_quota"`

	// Reserved
	RxtxCap string `json:"rxtx_cap"`

	// Reserved
	AccessIsPublic bool `json:"os-flavor-access:is_public"`
}

//RegionFlavor info
type RegionFlavor struct {
	RegionFlavorID string  `json:"region_flavor_id"`
	Region         string  `json:"region"`
	Flavor
}

//RegionFlavor List
type RegionFlavors struct {
	Flavors []RegionFlavor `json:"flavors"`
}

//RegionVolumeType List
type RegionVolumeTypes struct {
	VolumeTypes []RegionVolumeType `json:"volume_types"`
}

// Volume Type contains all the information associated with an OpenStack Volume Type.
type VolumeType struct {
	// Unique identifier for the volume type.
	ID string `json:"id"`
	// Human-readable display name for the volume type.
	Name string `json:"name"`
	// Human-readable description for the volume type.
	Description string `json:"description"`
	// Arbitrary key-value pairs defined by the user.
	ExtraSpecs map[string]string `json:"extra_specs"`
	// Whether the volume type is publicly visible.
	IsPublic bool `json:"is_public"`
	// Qos Spec ID
	QosSpecID         string `json:"qos_specs_id"`
	VolumeBackendName string `json:"volume_backend_name"`
	AvailabilityZone  string `json:"availability-zone"`
	// Availability Zone list which support this type of volume
	RESKEYAvailabilityZone string `json:"RESKEY:availability_zone"`
	// Availability Zone list which sold out
	OSVenderExtendedSoldOutAvailabilityZones string `json:"os-vender-extended:sold_out_availability_zones"`
}

type VolumeCapabilities struct {
	PoolName string `json:"pool_name"`
	ProvisionedCapacityGb float64 `json:"provisioned_capacity_gb"`
	AllocatedCapacityGb float64 `json:"allocated_capacity_gb"`
	FreeCapacityGb float64 `json:"free_capacity_gb"`
	TotalCapacityGb float64 `json:"total_capacity_gb"`
	VolumeType string `json:"volume_type"`
	MaxOverSubscriptionRatio float64 `json:"max_over_subscription_ratio"`
	ThinProvisioningSupport bool `json:"thin_provisioning_support"`
	ReservedPercentage float64 `json:"reserved_percentage"`
	// PoolModel ，0: normal，1:public test，2: us pool
	PoolModel float64 `json:"pool_model"`
}

type VolumePoolInfo struct {
	AvailabilityZone string `json:"availability_zone"`
	// Name style：cinder-volume@volume_backend_name#pool_name
	Name string `json:"name"`
	Capabilities VolumeCapabilities `json:"capabilities"`
}

type VolumePools struct {
	StoragePools []VolumePoolInfo `json:"pools"`
}

//RegionVolumeType add region info
type RegionVolumeType struct {
	VolumeType
	Region string  `json:"region"`
}

//RegionVolumePool List
type RegionVolumePools struct {
	VolumePools []RegionVolumePool `json:"volume_pools"`
}

//RegionVolumePools region with volumepool
type RegionVolumePool struct {
	VolumePools
	Region string		`json:"region"`
}

// IPCommonPools struct
type IPCommonPools struct {
	CommonPools []IPCommonPool `json:"common_pools"`
}

// IPCommonPool struct
type IPCommonPool struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Size   int    `json:"size"`
	Used   int    `json:"used"`
	Cached int    `json:"cached"`
}

type EipPools struct {
	Pools []EipPool 	`json:"eip_pools"`
}

//EipPool eip pool
type EipPool struct {
	IPCommonPools
	Region string `json:"region"`
}

func (iep EipPool) ToString() string {
	ret, err := json.Marshal(iep)
	if err != nil {
		return ""
	}

	return string(ret)
}

// NodeInfo struct
type NodeInfo struct {
	EdgeNodeID         string	`json:"edge_node_id"`
	HostName           string	`json:"host_name"`
	Region             string	`json:"region"`
	AvailabilityZone   string	`json:"availability_zone"`
	TotalDisk          int		`json:"total_disk"`
	UsedDisk           int		`json:"used_disk"`
	TotalVCPUs         int		`json:"total_vcpus"`
	UsedVCPUs          int		`json:"used_vcpus"`
	TotalMem           int		`json:"total_mem"`
	UsedMem            int		`json:"used_mem"`
	State              string	`json:"state"`
	Status             string	`json:"status"`
	ResourceType       string	`json:"resource_type"`
	CPUAllocationRatio string	`json:"cpu_allocation_ratio"`
	EcsPerformanceType string	`json:"ecs_performance_type"`
	MaxCount           int		`json:"max_count"`
}

type Operator struct {
	ID string `json:"id"`
	Name string `json:"name,omitempty"`
	I18Name string `json:"i18n_name,omitempty"`
	Sa string `json:"sa"`
}

type SiteAttribute struct {
	ID string `json:"id"`
	Key string `json:"site_attr"`
	Value string `json:"site_attr_value"`
}

type SiteBase struct {
	City string `json:"city,omitempty"`
	I18nCity string `json:"i18n_city,omitempty"`
	Province string `json:"province,omitempty"`
	I18nProvince string `json:"i18n_province,omitempty"`
	Area string `json:"area,omitempty"`
	I18nArea string `json:"i18n_area,omitempty"`
	Country string `json:"country,omitempty"`
	I18nCountry string `json:"i18n_country,omitempty"`
	Operator *Operator `json:"operator,omitempty"`
}

type Site struct {
	ID string `json:"id"`
	Name string `json:"name"`
	SiteBase
	Region string `json:"region"`
	Az string `json:"az"`
	Status string `json:"status"`
	EipNetworkID string `json:"eip_network_id"`
	EipTypeName string `json:"eip_type_name"`
	EipCidr []string `json:"eip_cidr"`
	SiteAttributes []*SiteAttribute `json:"site_attributes"`
}

type Sites struct {
	Sites []Site `json:"sites"`
}

type SiteNode struct {
	EdgeSiteID       string		`json:"edge_site_id"`
	Region           string		`json:"region"`
	AvailabilityZone string		`json:"availability_zone"`
	Operator         string		`json:"operator"`
	Status           string		`json:"status"`
	Nodes            []NodeInfo	`json:"node_infos"`
	MaxCount         int		`json:"max_count"`
}

type SiteNodes struct {
	SiteNodes []SiteNode `json:"site_nodes"`
}
