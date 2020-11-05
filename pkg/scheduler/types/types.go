package types

import (
	"encoding/json"

	"k8s.io/kubernetes/pkg/scheduler/client/typed"
)

const (
	// SchedulerDefaultProviderName defines the default provider names
	SchedulerDefaultProviderName = "DefaultProvider"
	// ResourceCPU cpu type
	ResourceCPU = "cpu"
	// ResourceMemory memory type
	ResourceMemory = "memory"
	// ResourceStorage storage type
	ResourceStorage = "storage"
	// ResourceEip eip type
	ResourceEip = "eip"
	// MaxNodeScore is the maximum score a Score plugin is expected to return.
	MaxNodeScore int64 = 100
	NodeTempSelectorKey = "NodeTempSelected"
)

// GeoLocation struct
type GeoLocation struct {
	// Country title
	Country string `json:"country,omitempty"`
	// Area title
	Area string `json:"area,omitempty"`
	// Province title
	Province string `json:"province,omitempty"`
	// City title
	City string `json:"city,omitempty"`
}

// CloudRegion cloud region
type CloudRegion struct {
	// Region info
	Region string `json:"region" required:"true"`
	// az info
	AvailabilityZone []string `json:"availability_zone,omitempty"`
}

// RegionAzMap  region to az
type RegionAzMap struct {
	// Region info
	Region string `json:"region" required:"true"`
	// az info
	AvailabilityZone string `json:"availability_zone,omitempty"`
}

// Strategy
type Strategy struct {
	// Location Strategy（centralize，discrete）
	LocationStrategy string `json:"location_strategy,omitempty"`
	// Region Strategy（alone）
	RegionStrategy string `json:"region_strategy,omitempty"`
}

// Spot
type Spot struct {
	MaxPrice           string `json:"max_price,omitempty"`
	SpotDurationHours  int    `json:"spot_duration_hours,omitempty"`
	SpotDurationCount  int    `json:"spot_duration_count,omitempty"`
	InterruptionPolicy string `json:"interruption_policy,omitempty"`
}

type Flavor struct {
	FlavorID string `json:"flavor_id" required:"true"`
	Spot     *Spot  `json:"spot,omitempty"`
}

// Resource is reource
type Resource struct {
	Name string `json:"name" required:"true"`
	ResourceType string `json:"resource_type" required:"true"`
	Flavors []Flavor `json:"flavors" required:"true"`
	Storage map[string]int64 `json:"storage,omitempty"`
	// NeedEip, if the value is false, no EIP is bound. If the value is true, an EIP is bound.
	NeedEip bool `json:"need_eip,omitempty"`
	// Number of ECSs to be created.
	Count int `json:"count,omitempty"`
	FlavorIDSelected string `json:"-"`
}

// Selector selector policy
type Selector struct {
	GeoLocation       GeoLocation        `json:"geo_location,omitempty"`
	Regions           []CloudRegion      `json:"regions,omitempty"`
	Operator          string             `json:"operator,omitempty"`
	NodeID            string             `json:"node_id,omitempty"`
	Strategy          Strategy           `json:"strategy,omitempty"`
	StackAffinity     []LabelRequirement `json:"stack_affinity,omitempty"`
	StackAntiAffinity []LabelRequirement `json:"stack_antiaffinity,omitempty"`
}

// Stack is resource group
type Stack struct {
	Name      string            `json:"name" required:"true"`
	Labels    map[string]string `json:"labels,omitempty"`
	Resources []*Resource       `json:"resources,omitempty"`
	Selector  Selector          `json:"-"`
	Selected  Selected          `json:"-"`
	UID       string            `json:"uid,omitempty"`
}

func (in *Stack) DeepCopy() *Stack {
	// TODO
	return in
}

// Allocation allocation object
type Allocation struct {
	ID       string   `json:"id,omitempty"`
	Stack    Stack    `json:"resource_group" required:"true"`
	Selector Selector `json:"selector" required:"true"`
	Replicas int      `json:"replicas" required:"true"`
}

// AllocationReq allocation request object
type AllocationReq struct {
	Allocation Allocation `json:"allocation" required:"true"`
}

// RespResource
type RespResource struct {
	Name string `json:"name"`
	FlavorID string `json:"flavor_id"`
	Count int `json:"count,omitempty"`
}

// Selected node
type Selected struct {
	NodeID string `json:"node_id,omitempty"`
	//Region info
	Region string `json:"region,omitempty"`
	//az info
	AvailabilityZone string `json:"availability_zone,omitempty"`
}

// RespStack
type RespStack struct {
	Name      string         `json:"name" required:"true"`
	Resources []RespResource `json:"resources,omitempty"`
	Selected  Selected       `json:"selected,omitempty"`
}

// AllocationResult allocation result object
type AllocationResult struct {
	ID    string      `json:"id"`
	Stack []RespStack `json:"resource_groups"`
}

// AllocationsResp allocation response object
type AllocationsResp struct {
	AllocationResult AllocationResult `json:"allocation"`
}

// spot resource info
type SpotResource struct {
	CpuBlockDelayResource        PreemptibleResource
	MemoryBlockDelayResource     PreemptibleResource
	CpuDelayResource             PreemptibleResource
	MemoryDelayResource          PreemptibleResource
	CpuBlockImmediateResource    PreemptibleResource
	MemoryBlockImmediateResource PreemptibleResource
	CpuImmediateResource         PreemptibleResource
	MemoryImmediateResource      PreemptibleResource
}

// SiteNode site info
type SiteNode struct {
	SiteID string `json:"site_id"`
	GeoLocation
	RegionAzMap
	Operator string `json:"operator"`
	Status   string `json:"status"`
	SiteAttribute []*typed.SiteAttribute        `json:"site_attributes"`
	EipTypeName   string                  `json:"eiptype_name"`
	SpotResources map[string]SpotResource `json:"spot_resources"`
	Nodes         []typed.NodeInfo        `json:"-"`
}

func (sn *SiteNode) Clone() *SiteNode {
	ret := &SiteNode{
		SiteID:      sn.SiteID,
		GeoLocation: sn.GeoLocation,
		RegionAzMap: sn.RegionAzMap,
		Operator:    sn.Operator,
		Status:      sn.Status,
		EipTypeName: sn.EipTypeName,
	}

	ret.Nodes = append(ret.Nodes, sn.Nodes...)
	return ret
}

func (sn *SiteNode) ToString() string {
	ret, err := json.Marshal(sn)
	if err != nil {
		return ""
	}

	return string(ret)
}

// Plugins include multiple extension points. When specified, the list of plugins for
// a particular extension point are the only ones enabled. If an extension point is
// omitted from the config, then the default set of plugins is used for that extension point.
// Enabled plugins are called in the order specified here, after default plugins. If they need to
// be invoked before default plugins, default plugins must be disabled and re-enabled here in desired order.
type Plugins struct {
	// QueueSort is a list of plugins that should be invoked when sorting pods in the scheduling queue.
	QueueSort *PluginSet

	// PreFilter is a list of plugins that should be invoked at "PreFilter" extension point of the scheduling framework.
	PreFilter *PluginSet

	// Filter is a list of plugins that should be invoked when filtering out nodes that cannot run the Pod.
	Filter *PluginSet

	// PreScore is a list of plugins that are invoked before scoring.
	PreScore *PluginSet

	// Score is a list of plugins that should be invoked when ranking nodes that have passed the filtering phase.
	Score *PluginSet

	// Reserve is a list of plugins invoked when reserving a node to run the pod.
	Reserve *PluginSet

	// Permit is a list of plugins that control binding of a Pod. These plugins can prevent or delay binding of a Pod.
	Permit *PluginSet

	// PreBind is a list of plugins that should be invoked before a pod is bound.
	PreBind *PluginSet

	// Bind is a list of plugins that should be invoked at "Bind" extension point of the scheduling framework.
	// The scheduler call these plugins in order. Scheduler skips the rest of these plugins as soon as one returns success.
	Bind *PluginSet

	// PostBind is a list of plugins that should be invoked after a pod is successfully bound.
	PostBind *PluginSet

	// Unreserve is a list of plugins invoked when a pod that was previously reserved is rejected in a later phase.
	Unreserve *PluginSet

	// Strategy is a list of plugins
	Strategy *PluginSet
}

// PluginSet specifies enabled and disabled plugins for an extension point.
// If an array is empty, missing, or nil, default plugins at that extension point will be used.
type PluginSet struct {
	// Enabled specifies plugins that should be enabled in addition to default plugins.
	// These are called after default plugins and in the same order specified here.
	Enabled []Plugin
	// Disabled specifies default plugins that should be disabled.
	// When all default plugins need to be disabled, an array containing only one "*" should be provided.
	Disabled []Plugin
}

// Plugin specifies a plugin name and its weight when applicable. Weight is used only for Score plugins.
type Plugin struct {
	// Name defines the name of plugin
	Name string
	// Weight defines the weight of plugin, only used for Score plugins.
	Weight int32
}

// LabelRequirement contains values, a key, and an operator that relates the key and values.
// The zero value of Requirement is invalid.
// Requirement implements both set based match and exact match
// Requirement should be initialized via NewRequirement constructor for creating a valid Requirement.
type LabelRequirement struct {
	Key      string `json:"key" required:"true"`
	Operator string `json:"operator" required:"true"`
	// In huge majority of cases we have at most one value here.
	// It is generally faster to operate on a single-element slice
	// than on a single-element map, so we have a slice here.
	StrValues []string `json:"values" required:"true"`
}

//CPUAndMemory cpu and memory
type CPUAndMemory struct {
	VCPU   int64 `json:"vcpu"`
	Memory int64 `json:"memory"`
}

//Clone clone
func (cm *CPUAndMemory) Clone() *CPUAndMemory {
	return &CPUAndMemory{VCPU: cm.VCPU, Memory: cm.Memory}
}

//AllResInfo res info
type AllResInfo struct {
	CpuAndMem map[string]CPUAndMemory
	Storage   map[string]float64
	eipNum    int
}

//AllocationRatioByHost is allocation ratio of host
type AllocationRatioByHost struct {
	AllocationRatio string `json:"allocation_ratio"`
	Host            string `json:"host"`
}

//AllocationRatioByHost is allocation ratio of host
type AllocationRatioByType struct {
	CoreAllocationRatio string `json:"core_allocation_ratio"`
	MemAllocationRatio  string `json:"mem_allocation_ratio"`
}

//AllocationRatio is resource allocation ratio
type AllocationRatio struct {
	AllocationRatio        string                  `json:"allocation_ratio"`
	AllocationRatioByType  AllocationRatioByType   `json:"allocation_ratio_by_type"`
	AllocationRatioByHosts []AllocationRatioByHost `json:"allocation_ratio_by_host"`
	AvailabilityZone       string                  `json:"availability_zone"`
	Flavor                 string                  `json:"flavor"`
}

type NetMetricData struct {
	Country  string  `json:"country"`
	Area     string  `json:"area"`
	Province string  `json:"province"`
	City     string  `json:"city"`
	Operator string  `json:"operator"`
	Delay    float64 `json:"delay"`
	Lossrate float64 `json:"lossrate"`
}

type NetMetricDatas struct {
	MetricDatas []NetMetricData `json:"metric_data"`
}

type PreemptibleResource struct {
	Uuid               string             `json:"uuid,omitempty"`
	Region             string             `json:"region,omitempty"`
	AvailabilityZone   string             `json:"availability_zone,omitempty"`
	AnalyzedAt         int32              `json:"analyzed_at,omitempty"`
	TargetResource     string             `json:"target_resource,omitempty"`
	ForecastAt         int32              `json:"forecast_at,omitempty"`
	FlavorType         string             `json:"flavor_type,omitempty"`
	FlavorName         string             `json:"flavor_name,omitempty"`
	Confidence         int32              `json:"confidence,omitempty"`
	InterruptionPolicy string             `json:"interruption_policy,omitempty"`
	Block              bool               `json:"block,omitempty"`
	Infinite           float32            `json:"infinite,omitempty"`
	Finite             map[string]float32 `json:"finite,omitempty"`
}

// ListSiteOpts to list site
type ListSiteOpts struct {
	//Limit query limit
	Limit string `q:"limit"`

	//Offset query begin index
	Offset string `q:"offset"`

	//id query by id
	ID string `q:"id"`

	//Area query by area
	Area string `q:"area"`

	//Province query by province
	Province string `q:"province"`

	//City query by city
	City string `q:"city"`

	//Operator query by operator
	Operator string `q:"operator"`
}
