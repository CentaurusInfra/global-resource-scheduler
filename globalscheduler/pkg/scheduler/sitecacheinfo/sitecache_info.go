/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package sitecacheinfo

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
// Collision of the generation numbers would be particularly problematic if a Site was deleted and
// added back with the same name. See issue#63262.
func nextGeneration() int64 {
	return atomic.AddInt64(&generation, 1)
}

// SiteCacheInfo is Site level aggregated information.
type SiteCacheInfo struct {
	Site *types.Site `json:"site"`

	// Total requested  of all resources on this Site. This includes assumed
	// resources, which scheduler has sent for binding, but may not be scheduled yet.
	RequestedResources map[string]*types.CPUAndMemory `json:"requestedCPUAndMemory"`

	// We store TotalResources (which is Site.Status.Allocatable.*) explicitly
	// as int64, to avoid conversions and accessing map.
	TotalResources map[string]*types.CPUAndMemory `json:"allocatableCPUAndMemory"`

	// Total requested resources of all resources on this Site. This includes assumed
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

	// Whenever SiteCacheInfo changes, generation is bumped.
	// This is used to avoid cloning it if the object didn't change.
	generation int64
}

// GetGeneration returns the generation on this Site.
func (n *SiteCacheInfo) GetGeneration() int64 {
	if n == nil {
		return 0
	}
	return n.generation
}

// Clone returns a copy of this Site.
func (n *SiteCacheInfo) ToString() string {
	ret := fmt.Sprintf("siteInfo: %#v, totalResources: %#v, RequestedResources: %#v, allocatableFlavor: %#v, "+
		"allocatableSpotFlavor: %#v, allocatablStorage: %#v, requestedStorage: %#v ",
		n.Site.ToString(), n.TotalResources, n.RequestedResources, n.AllocatableFlavor, n.AllocatableSpotFlavor, n.TotalStorage, n.RequestedStorage)
	if n.eipPool != nil {
		ret += fmt.Sprintf("eipPool: %#v", n.eipPool.ToString())
	}

	return ret
}

// Clone returns a copy of this Site.
func (n *SiteCacheInfo) Clone() *SiteCacheInfo {
	clone := &SiteCacheInfo{
		Site:                  n.Site.Clone(),
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

// AddStack adds stack information to this SiteCacheInfo.
func (n *SiteCacheInfo) AddStack(stack *types.Stack) {
	n.stacks = append(n.stacks, stack)
	n.generation = nextGeneration()
}

// RemoveStack subtracts pod information from this SiteCacheInfo.
func (n *SiteCacheInfo) RemoveStack(stack *types.Stack) error {
	return nil
}

func (n *SiteCacheInfo) getCondOperationAz(condAzStr string) map[string]string {
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

func (n *SiteCacheInfo) getSupportFlavorsBySite() []typed.Flavor {
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

		for _, host := range n.GetSite().Hosts {
			if n.GetSite().Region != regionFlv.Region {
				continue
			}
			flavorExtraSpecs := regionFlv.OsExtraSpecs
			resTypes := strings.Split(host.ResourceType, "||")
			if utils.IsContain(resTypes, flavorExtraSpecs.ResourceType) {
				flavorStatus := flavorExtraSpecs.CondOperationStatus
				azMaps := n.getCondOperationAz(flavorExtraSpecs.CondOperationAz)
				if flavorStatus == "abandon" {
					if flag, ok := azMaps[n.GetSite().AvailabilityZone]; ok {
						if flag != "sellout" && flag != "abandon" {
							addFlavrFunc(regionFlv.Flavor)
						}
					}
				} else {
					if flag, ok := azMaps[n.GetSite().AvailabilityZone]; ok {
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

func (n *SiteCacheInfo) getTotalResourceByResType(resType string) types.CPUAndMemory {
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

func (n *SiteCacheInfo) getRequestResourceByResType(resType string) types.CPUAndMemory {
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

func (n *SiteCacheInfo) updateRequestResourceByResType(resType string, res *types.CPUAndMemory) {
	for tempResType := range n.RequestedResources {
		tempResTypes := strings.Split(tempResType, "||")
		if utils.IsContain(tempResTypes, resType) {
			delete(n.RequestedResources, tempResType)
		}
	}

	n.RequestedResources[resType] = res
}

func (n *SiteCacheInfo) updateFlavor() {
	n.AllocatableFlavor = map[string]int64{}
	supportFlavors := n.getSupportFlavorsBySite()

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

// SetSite sets the overall Site information.
func (n *SiteCacheInfo) SetSite(site *types.Site) error {
	n.Site = site
	n.RequestedResources = make(map[string]*types.CPUAndMemory)
	n.TotalResources = make(map[string]*types.CPUAndMemory)
	n.AllocatableFlavor = make(map[string]int64)
	n.RequestedFlavor = make(map[string]int64)
	n.AllocatableSpotFlavor = site.SpotResources

	for _, no := range site.Hosts {
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

// UpdateSiteWithEipPool update eip pool
func (n *SiteCacheInfo) UpdateSiteWithEipPool(eipPool *typed.EipPool) error {
	n.eipPool = eipPool

	n.generation = nextGeneration()
	return nil
}

// UpdateSiteWithVolumePool update volume pool
func (n *SiteCacheInfo) UpdateSiteWithVolumePool(volumePool *typed.RegionVolumePool) error {
	var allocatableStorage = map[string]float64{}
	var requestedStorage = map[string]float64{}
	for _, storage := range volumePool.StoragePools {
		if n.GetSite().AvailabilityZone == storage.AvailabilityZone {
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

// UpdateSiteWithResInfo update res info
func (n *SiteCacheInfo) UpdateSiteWithResInfo(resInfo types.AllResInfo) error {
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
func (n *SiteCacheInfo) UpdateQos(netMetricData *types.NetMetricDatas) error {
	var qosMap = map[string]float64{}
	for _, metric := range netMetricData.MetricDatas {
		qos := metric.Delay + config.DefaultFloat64(constants.ConfQosWeight, 0.5)*metric.Lossrate
		qosMap[metric.City] = qos
	}

	n.Qos = qosMap

	return nil
}

// UpdateSiteWithRatio update requested cpu and mem
func (n *SiteCacheInfo) UpdateSiteWithRatio(ratios []types.AllocationRatio) error {
	var resTypeMapRatio = map[string]types.AllocationRatio{}
	for _, resAllo := range ratios {
		flavorInfo, find := informers.InformerFac.GetFlavor(resAllo.Flavor, n.GetSite().Region)
		if !find {
			continue
		}
		resTypeMapRatio[flavorInfo.OsExtraSpecs.ResourceType] = resAllo
	}

	for resType, value := range resTypeMapRatio {
		totalRes := n.getTotalResourceByResType(resType)

		if totalRes.VCPU <= 0 || totalRes.Memory <= 0 {
			logger.Warnf("region: %s, az: %s, resType:%s has invalid totalCpu(%d) or totalMem(%d)",
				n.GetSite().Region, n.GetSite().AvailabilityZone, resType, totalRes.VCPU, totalRes.Memory)
			continue
		}

		cpuRatio, err := strconv.ParseFloat(value.AllocationRatioByType.CoreAllocationRatio, 64)
		if err != nil {
			logger.Warnf("region: %s, az: %s, resType:%s has invalid cpuRatio(%s)",
				n.GetSite().Region, n.GetSite().AvailabilityZone, resType, value.AllocationRatioByType.CoreAllocationRatio)
			continue
		}
		memRatio, err := strconv.ParseFloat(value.AllocationRatioByType.MemAllocationRatio, 64)
		if err != nil {
			logger.Warnf("region: %s, az: %s, resType:%s has invalid memRatio(%s)",
				n.GetSite().Region, n.GetSite().AvailabilityZone, resType, value.AllocationRatioByType.MemAllocationRatio)
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
func (n *SiteCacheInfo) UpdateSpotResources(spotRes map[string]types.SpotResource) error {
	n.AllocatableSpotFlavor = spotRes

	n.generation = nextGeneration()
	return nil
}

// StackWithAffinity return all pods with (anti)affinity constraints on this Site.
func (n *SiteCacheInfo) StackWithAffinity() []*types.Stack {
	if n == nil {
		return nil
	}
	return n.stacksWithAffinity
}

// Stacks return all stacks scheduled (including assumed to be) on this Site.
func (n *SiteCacheInfo) Stacks() []*types.Stack {
	if n == nil {
		return nil
	}
	return n.stacks
}

// Site returns overall information about this Site.
func (n *SiteCacheInfo) GetSite() *types.Site {
	if n == nil {
		return nil
	}
	return n.Site
}

// NewSiteCacheInfo returns a ready to use empty SiteCacheInfo object.
// If any pods are given in arguments, their information will be aggregated in
// the returned object.
func NewSiteCacheInfo(stacks ...*types.Stack) *SiteCacheInfo {
	ni := &SiteCacheInfo{
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
