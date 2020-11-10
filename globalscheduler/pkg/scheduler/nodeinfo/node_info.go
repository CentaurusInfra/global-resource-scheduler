package nodeinfo

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/config"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

var (
	generation int64
)

// nextGeneration: Let's make sure history never forgets the name...
// Increments the generation number monotonically ensuring that generation numbers never collide.
// Collision of the generation numbers would be particularly problematic if a node was deleted and
// added back with the same name. See issue#63262.
func nextGeneration() int64 {
	return atomic.AddInt64(&generation, 1)
}

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	node *types.SiteNode

	// Total requested  of all resources on this node. This includes assumed
	// resources, which scheduler has sent for binding, but may not be scheduled yet.
	RequestedResources map[string]*types.CPUAndMemory `json:"requestedCPUAndMemory"`

	// We store TotalResources (which is Node.Status.Allocatable.*) explicitly
	// as int64, to avoid conversions and accessing map.
	TotalResources map[string]*types.CPUAndMemory `json:"allocatableCPUAndMemory"`

	// Total requested resources of all resources on this node. This includes assumed
	RequestedStorage map[string]float64 `json:"requestedStorage"`

	// We store TotalStorage  explicitly
	TotalStorage map[string]float64 `json:"allocatableStorage"`

	// We store requestedFlavor  explicitly
	// as int64, to avoid conversions and accessing map.
	RequestedFlavor map[string]int64 `json:"requestedFlavor"`

	// We store allocatableFlavor  explicitly
	// as int64, to avoid conversions and accessing map.
	AllocatableFlavor map[string]int64 `json:"allocatableFlavor"`

	// We store AllocatableSpotFlavor  explicitly
	// as int64, to avoid conversions and accessing map.
	AllocatableSpotFlavor map[string]types.SpotResource `json:"allocatableSpotFlavor"`

	stacks             []*types.Stack
	stacksWithAffinity []*types.Stack

	eipPool *typed.EipPool

	Qos map[string]float64

	// Whenever NodeInfo changes, generation is bumped.
	// This is used to avoid cloning it if the object didn't change.
	generation int64
}

// GetGeneration returns the generation on this node.
func (n *NodeInfo) GetGeneration() int64 {
	if n == nil {
		return 0
	}
	return n.generation
}

// Clone returns a copy of this node.
func (n *NodeInfo) ToString() string {
	ret := fmt.Sprintf("nodeInfo: %#v, totalResources: %#v, RequestedResources: %#v, allocatableFlavor: %#v, "+
		"allocatableSpotFlavor: %#v, allocatablStorage: %#v, requestedStorage: %#v ",
		n.node.ToString(), n.TotalResources, n.RequestedResources, n.AllocatableFlavor, n.AllocatableSpotFlavor, n.TotalStorage, n.RequestedStorage)
	if n.eipPool != nil {
		ret += fmt.Sprintf("eipPool: %#v", n.eipPool.ToString())
	}

	return ret
}

// Clone returns a copy of this node.
func (n *NodeInfo) Clone() *NodeInfo {
	clone := &NodeInfo{
		node:                  n.node.Clone(),
		RequestedResources:    make(map[string]*types.CPUAndMemory),
		TotalResources:        make(map[string]*types.CPUAndMemory),
		RequestedStorage:      make(map[string]float64),
		TotalStorage:          make(map[string]float64),
		RequestedFlavor:       make(map[string]int64),
		AllocatableFlavor:     make(map[string]int64),
		AllocatableSpotFlavor: make(map[string]types.SpotResource),
		Qos:                   make(map[string]float64),
	}

	if n.RequestedResources != nil {
		for key, value := range n.RequestedResources {
			clone.RequestedResources[key] = value.Clone()
		}
	}

	if n.TotalResources != nil {
		for key, value := range n.TotalResources {
			clone.TotalResources[key] = value.Clone()
		}
	}

	if n.RequestedStorage != nil {
		for key, value := range n.RequestedStorage {
			clone.RequestedStorage[key] = value
		}
	}

	if n.TotalStorage != nil {
		for key, value := range n.TotalStorage {
			clone.TotalStorage[key] = value
		}
	}

	if n.RequestedFlavor != nil {
		for key, value := range n.RequestedFlavor {
			clone.RequestedFlavor[key] = value
		}
	}

	if n.AllocatableFlavor != nil {
		for key, value := range n.AllocatableFlavor {
			clone.AllocatableFlavor[key] = value
		}
	}

	if n.AllocatableSpotFlavor != nil {
		for key, value := range n.AllocatableSpotFlavor {
			clone.AllocatableSpotFlavor[key] = value
		}
	}

	if n.Qos != nil {
		for key, value := range n.Qos {
			clone.Qos[key] = value
		}
	}

	return clone
}

// AddStack adds stack information to this NodeInfo.
func (n *NodeInfo) AddStack(stack *types.Stack) {
	n.stacks = append(n.stacks, stack)
	n.generation = nextGeneration()
}

// RemoveStack subtracts pod information from this NodeInfo.
func (n *NodeInfo) RemoveStack(stack *types.Stack) error {
	return nil
}

func (n *NodeInfo) getCondOperationAz(condAzStr string) map[string]string {
	var azMaps = map[string]string{}

	azArray := strings.Split(condAzStr, ",")
	for _, asStr := range azArray {
		re := regexp.MustCompile(`([\w-]+)\(([\w]+)\)`)
		matched := re.FindStringSubmatch(asStr)
		if matched == nil || len(matched) < 3 {
			continue
		}
		azMaps[matched[1]] = matched[2]
	}

	return azMaps
}

func (n *NodeInfo) getSupportFlavorsByNode() []typed.Flavor {
	var flavorSlice []typed.Flavor

	addFlavrFunc := func(flv typed.Flavor) {
		var isFind = false
		for _, one := range flavorSlice {
			if one.ID == flv.ID {
				isFind = true
				break
			}
		}

		if !isFind {
			flavorSlice = append(flavorSlice, flv)
		}
	}

	flavors := informers.InformerFac.GetInformer(informers.FLAVOR).GetStore().List()
	for _, fla := range flavors {
		regionFlv, ok := fla.(typed.RegionFlavor)
		if !ok {
			continue
		}

		for _, node := range n.Node().Nodes {
			if node.Region != regionFlv.Region {
				continue
			}
			flavorExtraSpecs := regionFlv.OsExtraSpecs
			resTypes := strings.Split(node.ResourceType, "||")
			if utils.IsContain(resTypes, flavorExtraSpecs.ResourceType) {
				flavorStatus := flavorExtraSpecs.CondOperationStatus
				azMaps := n.getCondOperationAz(flavorExtraSpecs.CondOperationAz)
				if flavorStatus == "abandon" {
					if flag, ok := azMaps[node.AvailabilityZone]; ok {
						if flag != "sellout" && flag != "abandon" {
							addFlavrFunc(regionFlv.Flavor)
						}
					}
				} else {
					if flag, ok := azMaps[node.AvailabilityZone]; ok {
						if flag != "sellout" && flag != "abandon" {
							addFlavrFunc(regionFlv.Flavor)
						}
					} else {
						addFlavrFunc(regionFlv.Flavor)
					}
				}
			}
		}
	}

	return flavorSlice
}

func (n *NodeInfo) getTotalResourceByResType(resType string) types.CPUAndMemory {
	ret := types.CPUAndMemory{}
	for tempResType, resInfo := range n.TotalResources {
		tempResTypes := strings.Split(tempResType, "||")
		if utils.IsContain(tempResTypes, resType) {
			ret.VCPU += resInfo.VCPU
			ret.Memory += resInfo.Memory
		}
	}

	return ret
}

func (n *NodeInfo) getRequestResourceByResType(resType string) types.CPUAndMemory {
	ret := types.CPUAndMemory{}
	for tempResType, resInfo := range n.RequestedResources {
		tempResTypes := strings.Split(tempResType, "||")
		if utils.IsContain(tempResTypes, resType) {
			ret.VCPU += resInfo.VCPU
			ret.Memory += resInfo.Memory
		}
	}

	return ret
}

func (n *NodeInfo) updateRequestResourceByResType(resType string, res *types.CPUAndMemory) {
	for tempResType := range n.RequestedResources {
		tempResTypes := strings.Split(tempResType, "||")
		if utils.IsContain(tempResTypes, resType) {
			delete(n.RequestedResources, tempResType)
		}
	}

	n.RequestedResources[resType] = res
}

func (n *NodeInfo) updateFlavor() {
	n.AllocatableFlavor = map[string]int64{}
	supportFlavors := n.getSupportFlavorsByNode()

	for _, flv := range supportFlavors {
		vCPUInt, err := strconv.ParseInt(flv.Vcpus, 10, 64)
		if err != nil || vCPUInt <= 0 {
			continue
		}

		if flv.Ram <= 0 {
			continue
		}

		totalRes := n.getTotalResourceByResType(flv.OsExtraSpecs.ResourceType)
		requestRes := n.getRequestResourceByResType(flv.OsExtraSpecs.ResourceType)

		if totalRes.VCPU <= 0 || totalRes.Memory <= 0 {
			continue
		}

		count := (totalRes.VCPU - requestRes.VCPU) / vCPUInt
		memCount := (totalRes.Memory - requestRes.Memory) / flv.Ram
		if count > memCount {
			count = memCount
		}

		if _, ok := n.AllocatableFlavor[flv.ID]; !ok {
			n.AllocatableFlavor[flv.ID] = 0
		}

		n.AllocatableFlavor[flv.ID] += count
	}

}

// SetNode sets the overall node information.
func (n *NodeInfo) SetNode(node *types.SiteNode) error {
	n.node = node
	n.RequestedResources = make(map[string]*types.CPUAndMemory)
	n.TotalResources = make(map[string]*types.CPUAndMemory)
	n.AllocatableFlavor = make(map[string]int64)
	n.RequestedFlavor = make(map[string]int64)
	n.AllocatableSpotFlavor = node.SpotResources

	for _, no := range node.Nodes {
		if _, ok := n.RequestedResources[no.ResourceType]; !ok {
			n.RequestedResources[no.ResourceType] = &types.CPUAndMemory{VCPU: 0, Memory: 0}
		}

		if _, ok := n.TotalResources[no.ResourceType]; !ok {
			n.TotalResources[no.ResourceType] = &types.CPUAndMemory{VCPU: 0, Memory: 0}
		}

		var allocationRatio = 1.0
		var err error
		if no.CPUAllocationRatio != "" {
			allocationRatio, err = strconv.ParseFloat(no.CPUAllocationRatio, 32)
			if err != nil {
				logger.Warnf("Format CPUAllocationRatio from string to integer failed!Error: %s.", err.Error())
			}
		}

		n.RequestedResources[no.ResourceType].VCPU += int64(no.UsedVCPUs)
		n.RequestedResources[no.ResourceType].Memory += int64(no.UsedMem)
		n.TotalResources[no.ResourceType].VCPU += int64(float64(no.TotalVCPUs) * allocationRatio)
		n.TotalResources[no.ResourceType].Memory += int64(no.TotalMem)
	}

	n.updateFlavor()

	n.generation = nextGeneration()
	return nil
}

// UpdateNodeWithEipPool update eip pool
func (n *NodeInfo) UpdateNodeWithEipPool(eipPool *typed.EipPool) error {
	n.eipPool = eipPool

	n.generation = nextGeneration()
	return nil
}

// UpdateNodeWithVolumePool update volume pool
func (n *NodeInfo) UpdateNodeWithVolumePool(volumePool *typed.RegionVolumePool) error {
	var allocatableStorage = map[string]float64{}
	var requestedStorage = map[string]float64{}
	for _, storage := range volumePool.StoragePools {
		if n.Node().AvailabilityZone == storage.AvailabilityZone {
			var voType = storage.Capabilities.VolumeType
			if _, ok := allocatableStorage[voType]; !ok {
				allocatableStorage[voType] = 0
			}

			var totalDisks = storage.Capabilities.TotalCapacityGb
			if storage.Capabilities.MaxOverSubscriptionRatio > 0 {
				totalDisks = storage.Capabilities.TotalCapacityGb * storage.Capabilities.MaxOverSubscriptionRatio
			}
			allocatableStorage[voType] += totalDisks
			requestedStorage[voType] += storage.Capabilities.ProvisionedCapacityGb
		}
	}

	n.TotalStorage = allocatableStorage
	n.RequestedStorage = requestedStorage

	n.generation = nextGeneration()
	return nil
}

// UpdateNodeWithResInfo update res info
func (n *NodeInfo) UpdateNodeWithResInfo(resInfo types.AllResInfo) error {

	for resType, res := range resInfo.CpuAndMem {
		for reqType, reqRes := range n.RequestedResources {
			resTypes := strings.Split(reqType, "||")
			if !utils.IsContain(resTypes, resType) {
				continue
			}

			reqRes.VCPU += res.VCPU
			reqRes.Memory += res.Memory
			n.RequestedResources[reqType] = reqRes
		}
	}

	for volType, used := range resInfo.Storage {
		reqVol, ok := n.RequestedStorage[volType]
		if !ok {
			reqVol = 0
		}

		reqVol += used
		n.RequestedStorage[volType] = reqVol
	}

	n.updateFlavor()

	n.generation = nextGeneration()
	return nil
}

// UpdateQos update qos
func (n *NodeInfo) UpdateQos(netMetricData *types.NetMetricDatas) error {
	var qosMap = map[string]float64{}
	for _, metric := range netMetricData.MetricDatas {
		qos := metric.Delay + config.DefaultFloat64(constants.ConfQosWeight, 0.5)*metric.Lossrate
		qosMap[metric.City] = qos
	}

	n.Qos = qosMap

	return nil
}

// UpdateNodeWithRatio update requested cpu and mem
func (n *NodeInfo) UpdateNodeWithRatio(ratios []types.AllocationRatio) error {
	var resTypeMapRatio = map[string]types.AllocationRatio{}
	for _, resAllo := range ratios {
		flavorInfo, find := informers.InformerFac.GetFlavor(resAllo.Flavor, n.Node().Region)
		if !find {
			continue
		}
		resTypeMapRatio[flavorInfo.OsExtraSpecs.ResourceType] = resAllo
	}

	for resType, value := range resTypeMapRatio {
		totalRes := n.getTotalResourceByResType(resType)

		if totalRes.VCPU <= 0 || totalRes.Memory <= 0 {
			logger.Warnf("region: %s, az: %s, resType:%s has invalid totalCpu(%d) or totalMem(%d)",
				n.Node().Region, n.Node().AvailabilityZone, resType, totalRes.VCPU, totalRes.Memory)
			continue
		}

		cpuRatio, err := strconv.ParseFloat(value.AllocationRatioByType.CoreAllocationRatio, 64)
		if err != nil {
			logger.Warnf("region: %s, az: %s, resType:%s has invalid cpuRatio(%s)",
				n.Node().Region, n.Node().AvailabilityZone, resType, value.AllocationRatioByType.CoreAllocationRatio)
			continue
		}
		memRatio, err := strconv.ParseFloat(value.AllocationRatioByType.MemAllocationRatio, 64)
		if err != nil {
			logger.Warnf("region: %s, az: %s, resType:%s has invalid memRatio(%s)",
				n.Node().Region, n.Node().AvailabilityZone, resType, value.AllocationRatioByType.MemAllocationRatio)
			continue
		}
		var usedCpu = int64(float64(totalRes.VCPU) * cpuRatio)
		var usedMem = int64(float64(totalRes.Memory) * memRatio)
		n.updateRequestResourceByResType(resType, &types.CPUAndMemory{VCPU: usedCpu, Memory: usedMem})
	}

	n.updateFlavor()
	n.generation = nextGeneration()

	return nil
}

//UpdateSpotResources update spot resources
func (n *NodeInfo) UpdateSpotResources(spotRes map[string]types.SpotResource) error {
	n.AllocatableSpotFlavor = spotRes

	n.generation = nextGeneration()
	return nil
}

// StackWithAffinity return all pods with (anti)affinity constraints on this node.
func (n *NodeInfo) StackWithAffinity() []*types.Stack {
	if n == nil {
		return nil
	}
	return n.stacksWithAffinity
}

// Stacks return all stacks scheduled (including assumed to be) on this node.
func (n *NodeInfo) Stacks() []*types.Stack {
	if n == nil {
		return nil
	}
	return n.stacks
}

// Node returns overall information about this node.
func (n *NodeInfo) Node() *types.SiteNode {
	if n == nil {
		return nil
	}
	return n.node
}

// NewNodeInfo returns a ready to use empty NodeInfo object.
// If any pods are given in arguments, their information will be aggregated in
// the returned object.
func NewNodeInfo(stacks ...*types.Stack) *NodeInfo {
	ni := &NodeInfo{
		RequestedResources: make(map[string]*types.CPUAndMemory),
		TotalResources:     make(map[string]*types.CPUAndMemory),
		AllocatableFlavor:  make(map[string]int64),
		RequestedStorage:   make(map[string]float64),
		TotalStorage:       make(map[string]float64),
	}
	for _, stack := range stacks {
		ni.AddStack(stack)
	}
	return ni
}

// GetStackKey returns the string key of a stack.
func GetStackKey(stack *types.Stack) (string, error) {
	uid := string(stack.UID)
	if len(uid) == 0 {
		return "", errors.New("Cannot get cache key for stack with empty UID")
	}
	return uid, nil
}
