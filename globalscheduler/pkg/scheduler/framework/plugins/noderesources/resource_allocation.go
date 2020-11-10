package noderesources

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	schedulernodeinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
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
	nodeInfo *schedulernodeinfo.NodeInfo) (int64, *interfaces.Status) {
	node := nodeInfo.Node()
	if node == nil {
		return 0, interfaces.NewStatus(interfaces.Error, "node not found")
	}
	if r.resourceToWeightMap == nil {
		return 0, interfaces.NewStatus(interfaces.Error, "resources not found")
	}
	requested := make(resourceToValueMap, len(r.resourceToWeightMap))
	allocatable := make(resourceToValueMap, len(r.resourceToWeightMap))
	for resource := range r.resourceToWeightMap {
		allocatable[resource], requested[resource] = calculateResourceAllocatableRequest(nodeInfo, stack, resource)
	}

	var score = r.scorer(requested, allocatable, false, 0, 0)
	logger.Infof(
		"%v -> %v: %v, map of allocatable resources %v, map of requested resources %v ,score %d,",
		stack.Name, node.SiteID, r.Name,
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

func getResourceType(flavorID string, nodeInfo *schedulernodeinfo.NodeInfo) string {
	flavors := informers.InformerFac.GetInformer(informers.FLAVOR).GetStore().List()
	for _, fla := range flavors {
		regionFlv, ok := fla.(typed.RegionFlavor)
		if !ok {
			continue
		}

		if flavorID != regionFlv.ID {
			continue
		}

		for _, node := range nodeInfo.Node().Nodes {
			if node.Region != regionFlv.Region {
				continue
			}
			flavorExtraSpecs := regionFlv.OsExtraSpecs
			resTypes := strings.Split(node.ResourceType, "||")
			if utils.IsContain(resTypes, flavorExtraSpecs.ResourceType) {
				return flavorExtraSpecs.ResourceType
			}
		}
	}

	return ""
}

func calculateCpuAllocatableRequest(nodeInfo *schedulernodeinfo.NodeInfo, stack *types.Stack) (int64, int64) {
	stackInfo := calculateStackResourceRequest(stack)
	var allocatableCpu int64
	var requestedCpu int64

	for _, one := range stackInfo {
		for _, flv := range one.Flavors {
			resourceType := getResourceType(flv.FlavorID, nodeInfo)

			for key, value := range nodeInfo.TotalResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					allocatableCpu += value.VCPU
				}
			}

			for key, value := range nodeInfo.RequestedResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					requestedCpu += value.VCPU
				}
			}
		}
	}
	return allocatableCpu, requestedCpu
}

func calculateMemoryAllocatableRequest(nodeInfo *schedulernodeinfo.NodeInfo, stack *types.Stack) (int64, int64) {
	stackInfo := calculateStackResourceRequest(stack)
	var allocatableMem int64
	var requestedMem int64
	for _, one := range stackInfo {
		for _, flv := range one.Flavors {
			resourceType := getResourceType(flv.FlavorID, nodeInfo)

			for key, value := range nodeInfo.TotalResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					allocatableMem += value.Memory
				}
			}

			for key, value := range nodeInfo.RequestedResources {
				resTypes := strings.Split(key, "||")
				if utils.IsContain(resTypes, resourceType) {
					requestedMem += value.Memory
				}
			}
		}
	}
	return allocatableMem, requestedMem
}

func calculateStorageAllocatableRequest(nodeInfo *schedulernodeinfo.NodeInfo, stack *types.Stack) (int64, int64) {
	var allocatableStorage int64
	var requestedStroage int64
	stackStorageRequest := calculateStackStorageRequest(stack)
	for volType, size := range stackStorageRequest {
		if value, ok := nodeInfo.TotalStorage[volType]; ok {
			allocatableStorage += int64(value)
		}
		if value, ok := nodeInfo.RequestedStorage[volType]; ok {
			requestedStroage += int64(value)
		}
		requestedStroage += size
	}
	return allocatableStorage, requestedStroage
}

func calculateEipAllocatableRequest(nodeInfo *schedulernodeinfo.NodeInfo, stack *types.Stack) (int64, int64) {

	var eipCount = 0
	for _, server := range stack.Resources {
		if server.NeedEip {
			eipCount++
		}
	}

	eipPoolsInterface, ok := informers.InformerFac.GetInformer(informers.EIPPOOLS).GetStore().Get(nodeInfo.Node().Region)
	if !ok {
		logger.Warnf("Node (%s/%s) has no eip pools.", nodeInfo.Node().SiteID, nodeInfo.Node().Region)
		return 0, 0
	}

	eipPool, ok := eipPoolsInterface.(typed.EipPool)
	if !ok {
		msg := fmt.Sprintf("Node (%s/%s) eipPoolConvert failed.", nodeInfo.Node().SiteID, nodeInfo.Node().Region)
		logger.Errorf(msg)
		return 0, 0
	}

	var find = false
	var findEipPool = typed.IPCommonPool{}
	for _, pool := range eipPool.CommonPools {
		if nodeInfo.Node().EipTypeName == pool.Name {
			find = true
			findEipPool = pool
			break
		}
	}

	if !find {
		msg := fmt.Sprintf("Node (%s/%s) has no eip pools.", nodeInfo.Node().SiteID, nodeInfo.Node().Region)
		logger.Infof(msg)
		return 0, 0
	}

	if eipCount <= 0 {
		return 0, 0
	}

	return int64(findEipPool.Size), int64(findEipPool.Used + eipCount)
}

// calculateResourceAllocatableRequest returns resources Allocatable and Requested values
func calculateResourceAllocatableRequest(nodeInfo *schedulernodeinfo.NodeInfo, stack *types.Stack,
	resource string) (int64, int64) {

	switch resource {
	case types.ResourceCPU:
		return calculateCpuAllocatableRequest(nodeInfo, stack)
	case types.ResourceMemory:
		return calculateMemoryAllocatableRequest(nodeInfo, stack)
	case types.ResourceStorage:
		return calculateStorageAllocatableRequest(nodeInfo, stack)
	case types.ResourceEip:
		return calculateEipAllocatableRequest(nodeInfo, stack)
	}

	logger.Infof("requested resource %v not considered for node score calculation", resource)

	return 0, 0
}
