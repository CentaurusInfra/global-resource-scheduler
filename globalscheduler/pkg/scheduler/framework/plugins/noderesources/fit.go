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

package noderesources

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	schedulernodeinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/sets"
)

var _ interfaces.PreFilterPlugin = &Fit{}
var _ interfaces.FilterPlugin = &Fit{}

const (
	// FitName is the name of the plugin used in the plugin registry and configurations.
	FitName = "NodeResourcesFit"

	// preFilterStateKey is the key in CycleState to NodeResourcesFit pre-computed data.
	// Using the name of the plugin will likely help us avoid collisions with other plugins.
	preFilterStateKey = "PreFilter" + FitName
)

// Fit is a plugin that checks if a node has sufficient resources.
type Fit struct {
}

// FitArgs holds the args that are used to configure the plugin.
type FitArgs struct {
	// IgnoredResources is the list of resources that NodeResources fit filter
	// should ignore.
	IgnoredResources []string `json:"ignoredResources,omitempty"`
}

//FlavorInfos flavor infos
type FlavorInfos struct {
	Flavors []types.Flavor
	types.CPUAndMemory
	Count int64
}

func addFlavorInfo(flvInfos *FlavorInfos, flavorID string, spot *types.Spot) {
	flvInfos.Flavors = append(flvInfos.Flavors,
		types.Flavor{
			FlavorID: flavorID,
			Spot:     spot,
		})
}

// preFilterState computed at PreFilter and used at Filter.
type NodeResFilterState struct {
	FlavorInfos []FlavorInfos
}

// Clone the prefilter state.
func (s *NodeResFilterState) Clone() interfaces.StateData {
	return s
}

// Name returns name of the plugin. It is used in logs, etc.
func (f *Fit) Name() string {
	return FitName
}

func calculateStackResourceRequest(stack *types.Stack) []FlavorInfos {
	var ret = []FlavorInfos{}

	for _, server := range stack.Resources {
		flvInfos := FlavorInfos{}
		for _, flv := range server.Flavors {
			if flv.Spot != nil && flv.Spot.SpotDurationHours != 0 {
				continue
			}
			flvObj, exist := informers.InformerFac.GetFlavor(flv.FlavorID, "")
			if !exist {
				logger.Warnf("Flavor %v not found!", flv)
				continue
			}
			vCPUInt, err := strconv.Atoi(flvObj.Vcpus)
			if err != nil {
				logger.Warnf("Convert flavor Vcpus (%s) string to int failed, err: %s", flvObj.Vcpus, err.Error())
				continue
			}

			flvInfos.VCPU += int64(vCPUInt)
			flvInfos.Memory += flvObj.Ram
			flvInfos.Count = int64(server.Count)
			addFlavorInfo(&flvInfos, flvObj.ID, flv.Spot)
		}
		ret = append(ret, flvInfos)
	}

	return ret
}

// computeStackResourceRequest returns a schedulernodeinfo.Resource that covers the largest
// width in each resource dimension. Because init-containers run sequentially, we collect
// the max in each dimension iteratively. In contrast, we sum the resource vectors for
// regular containers since they run simultaneously.
//
// If Pod Overhead is specified and the feature gate is set, the resources defined for Overhead
// are added to the calculated Resource request sum
func computeStackResourceRequest(stack *types.Stack) *NodeResFilterState {
	result := &NodeResFilterState{}
	result.FlavorInfos = calculateStackResourceRequest(stack)

	return result
}

// PreFilter invoked at the prefilter extension point.
func (f *Fit) PreFilter(ctx context.Context, cycleState *interfaces.CycleState,
	stack *types.Stack) *interfaces.Status {
	cycleState.Lock()
	defer cycleState.Unlock()
	cycleState.Write(preFilterStateKey, computeStackResourceRequest(stack))
	return nil
}

func GetPreFilterState(cycleState *interfaces.CycleState) (*NodeResFilterState, error) {
	cycleState.RLock()
	defer cycleState.RUnlock()
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		// NodeResFilterState doesn't exist, likely PreFilter wasn't invoked.
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(*NodeResFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to NodeResourcesFit.preFilterState error", c)
	}
	return s, nil
}

// Filter invoked at the filter extension point.
// Checks if a node has sufficient resources, such as cpu, memory, gpu, opaque int resources etc to run a pod.
// It returns a list of insufficient resources, if empty, then the node has all the resources requested by the pod.
func (f *Fit) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	nodeInfo *schedulernodeinfo.NodeInfo) *interfaces.Status {
	s, err := GetPreFilterState(cycleState)
	if err != nil {
		return interfaces.NewStatus(interfaces.Error, err.Error())
	}

	for _, flvs := range s.FlavorInfos {
		var isMatch = false
		for _, flv := range flvs.Flavors {

			// deal with spot data
			if flv.Spot != nil && flv.Spot.MaxPrice != "" {
				// TODO: spot instance
			} else {
				if totalCount, ok := nodeInfo.AllocatableFlavor[flv.FlavorID]; ok {
					var requestedCount int64 = 0
					if value, exist := nodeInfo.RequestedFlavor[flv.FlavorID]; exist {
						requestedCount = value
					}

					if flvs.Count < totalCount-requestedCount {
						if _, exist := nodeInfo.RequestedFlavor[flv.FlavorID]; !exist {
							nodeInfo.RequestedFlavor[flv.FlavorID] = 0
						}
						nodeInfo.RequestedFlavor[flv.FlavorID] += flvs.Count
						isMatch = true
						break
					}
				}
			}
		}
		if !isMatch {
			msg := fmt.Sprintf("Node (%s-%s) do not support required flavor (%+v).",
				nodeInfo.Node().SiteID, nodeInfo.Node().Region, flvs.Flavors)
			logger.Info(ctx, msg)
			return interfaces.NewStatus(interfaces.Unschedulable, msg)
		}
	}

	return nil
}

// InsufficientResource describes what kind of resource limit is hit and caused the pod to not fit the node.
type InsufficientResource struct {
	ResourceName string
	// We explicitly have a parameter for reason to avoid formatting a message on the fly
	// for common resources, which is expensive for cluster autoscaler simulations.
	Reason    string
	Requested int64
	Used      int64
	Capacity  int64
}

// Fits checks if node have enough resources to host the pod.
func Fits(stack *types.Stack, nodeInfo *schedulernodeinfo.NodeInfo,
	ignoredExtendedResources sets.String) []InsufficientResource {
	return fitsRequest(computeStackResourceRequest(stack), nodeInfo, ignoredExtendedResources)
}

func fitsRequest(podRequest *NodeResFilterState, nodeInfo *schedulernodeinfo.NodeInfo,
	ignoredExtendedResources sets.String) []InsufficientResource {
	insufficientResources := make([]InsufficientResource, 0, 4)

	return insufficientResources
}

// NewFit initializes a new plugin and returns it.
func NewFit(_ interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	fit := &Fit{}
	return fit, nil
}
