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

package siteresources

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	"strings"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

// resourceToWeightMap contains resource name and weight.
type resourceToWeightMap map[string]int64

// resourceToValueMap contains resource name and score.
type resourceToValueMap map[string]int64

// resourceAllocationScorer contains information to calculate resource allocation score.
type resourceAllocationScorer struct {
	Name   string
	scorer func(requested, allocable resourceToValueMap, includeVolumes bool, requestedVolumes int,
		allocatableVolumes int) int64
	resourceToWeightMap resourceToWeightMap
}

// defaultRequestedRatioResources is used to set default requestToWeight map for CPU and memory
var defaultRequestedRatioResources = resourceToWeightMap{types.ResourceMemory: 10, types.ResourceCPU: 10,
	types.ResourceStorage: 5, types.ResourceEip: 1}

// score will use `scorer` function to calculate the score.
func (r *resourceAllocationScorer) score(
	stack *types.Stack,
	siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) (int64, *interfaces.Status) {
	site := siteCacheInfo.GetSite()
	if site == nil {
		return 0, interfaces.NewStatus(interfaces.Error, "site not found")
	}
	if r.resourceToWeightMap == nil {
		return 0, interfaces.NewStatus(interfaces.Error, "resources not found")
	}
	requested := make(resourceToValueMap, len(r.resourceToWeightMap))
	allocatable := make(resourceToValueMap, len(r.resourceToWeightMap))
	for resource := range r.resourceToWeightMap {
		allocatable[resource], requested[resource] = calculateResourceAllocatableRequest(siteCacheInfo, stack, resource)
	}

	var score = r.scorer(requested, allocatable, false, 0, 0)
	logger.Infof(
		"%v -> %v: %v, map of allocatable resources %v, map of requested resources %v ,score %d,",
		stack.Name, site.SiteID, r.Name,
		allocatable, requested, score,
	)

	return score, nil
}

func calculateStackStorageRequest(stack *types.Stack) map[string]int64 {
	var ret = map[string]int64{}
	for _, server := range stack.Resources {
		for volType, size := range server.Storage {
			if _, ok := ret[volType]; ok {
				ret[volType] = 0
			}

			ret[volType] += size
		}
	}

	return ret
}

func getResourceType(flavorID string, siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) string {
	flavor, ok := cache.FlavorCache.GetFlavor(flavorID, siteCacheInfo.GetSite().Region)
	if !ok {
		for _, host := range siteCacheInfo.GetSite().Hosts {
			if siteCacheInfo.GetSite().Region != flavor.Region {
				continue
			}
			flavorExtraSpecs := flavor.OsExtraSpecs
			resTypes := strings.Split(host.ResourceType, "||")
			if utils.IsContain(resTypes, flavorExtraSpecs.ResourceType) {
				return flavorExtraSpecs.ResourceType
			}
		}
	}

	return ""
}

func calculateCpuAllocatableRequest(siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo, stack *types.Stack) (int64, int64) {
	stackInfo := calculateStackResourceRequest(stack)
	var allocatableCpu int64
	var requestedCpu int64

	for _, one := range stackInfo {
		for _, flv := range one.Flavors {
			resourceType := getResourceType(flv.FlavorID, siteCacheInfo)

			for key, value := range siteCacheInfo.TotalResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					allocatableCpu += value.VCPU
				}
			}

			for key, value := range siteCacheInfo.RequestedResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					requestedCpu += value.VCPU
				}
			}
		}
	}
	return allocatableCpu, requestedCpu
}

func calculateMemoryAllocatableRequest(siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo, stack *types.Stack) (int64, int64) {
	stackInfo := calculateStackResourceRequest(stack)
	var allocatableMem int64
	var requestedMem int64
	for _, one := range stackInfo {
		for _, flv := range one.Flavors {
			resourceType := getResourceType(flv.FlavorID, siteCacheInfo)

			for key, value := range siteCacheInfo.TotalResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					allocatableMem += value.Memory
				}
			}

			for key, value := range siteCacheInfo.RequestedResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					requestedMem += value.Memory
				}
			}
		}
	}
	return allocatableMem, requestedMem
}

func calculateStorageAllocatableRequest(siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo, stack *types.Stack) (int64, int64) {
	var allocatableStorage int64
	var requestedStroage int64
	stackStorageRequest := calculateStackStorageRequest(stack)
	for volType, size := range stackStorageRequest {
		if value, ok := siteCacheInfo.TotalStorage[volType]; ok {
			allocatableStorage += int64(value)
		}
		if value, ok := siteCacheInfo.RequestedStorage[volType]; ok {
			requestedStroage += int64(value)
		}
		requestedStroage += size
	}
	return allocatableStorage, requestedStroage
}

// calculateResourceAllocatableRequest returns resources Allocatable and Requested values
func calculateResourceAllocatableRequest(siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo, stack *types.Stack,
	resource string) (int64, int64) {

	switch resource {
	case types.ResourceCPU:
		return calculateCpuAllocatableRequest(siteCacheInfo, stack)
	case types.ResourceMemory:
		return calculateMemoryAllocatableRequest(siteCacheInfo, stack)
	case types.ResourceStorage:
		return calculateStorageAllocatableRequest(siteCacheInfo, stack)
	}

	logger.Infof("requested resource %v not considered for site score calculation", resource)

	return 0, 0
}
