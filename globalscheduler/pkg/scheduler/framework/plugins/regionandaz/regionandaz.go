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

package regionandaz

import (
	"context"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/sets"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "regionandaz"

// RegionAndAz is a plugin that implements Priority based sorting.
type RegionAndAz struct {
	handle interfaces.FrameworkHandle
}

//RegionMap region mapping
type RegionMap struct {
	count    int
	nodeList interfaces.NodeScoreList
}

var _ interfaces.FilterPlugin = &RegionAndAz{}
var _ interfaces.StrategyPlugin = &RegionAndAz{}

// Name returns name of the plugin.
func (pl *RegionAndAz) Name() string {
	return Name
}

func (pl *RegionAndAz) regionEqual(region types.CloudRegion, nodeInfo *nodeinfo.NodeInfo) bool {
	if nodeInfo == nil {
		return false
	}

	if region.Region != "" && region.Region != nodeInfo.Node().Region {
		return false
	}

	if region.AvailabilityZone != nil && len(region.AvailabilityZone) > 0 {
		azSets := sets.NewString(region.AvailabilityZone...)
		if !azSets.Has(nodeInfo.Node().AvailabilityZone) {
			return false
		}
	}

	return true
}

// Filter invoked at the filter extension point.
func (pl *RegionAndAz) Filter(ctx context.Context, cycleState *interfaces.CycleState,
	stack *types.Stack, nodeInfo *nodeinfo.NodeInfo) *interfaces.Status {

	if len(stack.Selector.Regions) <= 0 {
		return nil
	}

	var match = false
	for _, region := range stack.Selector.Regions {
		if pl.regionEqual(region, nodeInfo) {
			match = true
			break
		}
	}

	if !match {
		return interfaces.NewStatus(interfaces.Unschedulable, "stack region not equal node region.")
	}

	return nil
}

//Strategy run strategy
func (pl *RegionAndAz) Strategy(ctx context.Context, state *interfaces.CycleState,
	allocations *types.Allocation, nodeList interfaces.NodeScoreList) (interfaces.NodeScoreList, *interfaces.Status) {

	if allocations.Selector.Strategy.RegionStrategy != constants.StrategyRegionAlone {
		return nodeList, nil
	}

	var regionMap = map[string]RegionMap{}
	for _, node := range nodeList {
		selectorInfo, err := interfaces.GetNodeSelectorState(state, node.Name)
		if err != nil {
			logger.Error(ctx, "GetNodeSelectorState %s failed! err: %s", node.Name, err)
			continue
		}

		nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(node.Name)
		if err != nil {
			logger.Error(ctx, "get node info %s failed! err: %s", node.Name, err)
			continue
		}

		if _, ok := regionMap[nodeInfo.Node().Region]; !ok {
			regionMap[nodeInfo.Node().Region] = RegionMap{count: 0, nodeList: interfaces.NodeScoreList{}}
		}

		tempRegion := regionMap[nodeInfo.Node().Region]
		tempRegion.count += selectorInfo.StackMaxCount
		tempRegion.nodeList = append(tempRegion.nodeList, node)
		regionMap[nodeInfo.Node().Region] = tempRegion
	}

	var finalRegion string
	var maxCount int
	for region, regionInfo := range regionMap {
		if regionInfo.count < allocations.Replicas {
			continue
		}

		if regionInfo.count > maxCount {
			finalRegion = region
			maxCount = regionInfo.count
		}
	}

	if finalRegion == "" {
		logger.Error(ctx, "we need (%d) stack, now support region-count(%#v)", allocations.Replicas, regionMap)
		return nil, interfaces.NewStatus(interfaces.Unschedulable, "region Capability cannot meet the needs")
	}

	return regionMap[finalRegion].nodeList, nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &RegionAndAz{handle: handle}, nil
}
