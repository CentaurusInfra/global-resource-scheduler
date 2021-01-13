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

package defaultbinder

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name of the plugin used in the plugin registry and configurations.
const Name = "DefaultBinder"

// DefaultBinder binds pods to nodes using a k8s client.
type DefaultBinder struct {
	handle interfaces.FrameworkHandle
}

var _ interfaces.BindPlugin = &DefaultBinder{}

// New creates a DefaultBinder.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &DefaultBinder{handle: handle}, nil
}

// Name returns the name of the plugin.
func (b DefaultBinder) Name() string {
	return Name
}

// Bind binds pods to nodes using the k8s client.
func (b DefaultBinder) Bind(ctx context.Context, state *interfaces.CycleState, stack *types.Stack,
	nodeID string) *interfaces.Status {
	nodeInfo, err := b.handle.SnapshotSharedLister().NodeInfos().Get(nodeID)
	if err != nil {
		logger.Error(ctx, "get node(%s) failed! err: %s", nodeID, err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting node %q from Snapshot: %v",
			nodeID, err))
	}

	region := nodeInfo.Node().Region
	resInfo := types.AllResInfo{CpuAndMem: map[string]types.CPUAndMemory{}, Storage: map[string]float64{}}

	stack.Selected.NodeID = nodeID
	stack.Selected.Region = region
	stack.Selected.AvailabilityZone = nodeInfo.Node().AvailabilityZone

	nodeSelectedInfo, err := interfaces.GetNodeSelectorState(state, nodeID)
	if err != nil {
		logger.Error(ctx, "GetNodeSelectorState failed! err: %s", err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting node %q info failed: %v", nodeID, err))
	}

	if len(stack.Resources) != len(nodeSelectedInfo.Flavors) {
		logger.Error(ctx, "flavor count not equal to server count! err: %s", err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("nodeID(%s) flavor count not equal to "+
			"server count!", nodeID))
	}

	for i := 0; i < len(stack.Resources); i++ {
		flavorID := nodeSelectedInfo.Flavors[i].FlavorID
		stack.Resources[i].FlavorIDSelected = flavorID
		flv, ok := informers.InformerFac.GetFlavor(flavorID, region)
		if !ok {
			logger.Warn(ctx, "flavor %s not found in region(%s)", flavorID, region)
			continue
		}
		vCPUInt, err := strconv.ParseInt(flv.Vcpus, 10, 64)
		if err != nil || vCPUInt <= 0 {
			logger.Warn(ctx, "flavor %s is invalid in region(%s)", flavorID, region)
			continue
		}

		reqRes, ok := resInfo.CpuAndMem[flv.OsExtraSpecs.ResourceType]
		if !ok {
			reqRes = types.CPUAndMemory{VCPU: 0, Memory: 0}
		}
		reqRes.VCPU += vCPUInt * int64(stack.Resources[i].Count)
		reqRes.Memory += flv.Ram * int64(stack.Resources[i].Count)
		resInfo.CpuAndMem[flv.OsExtraSpecs.ResourceType] = reqRes
	}

	b.handle.Cache().UpdateNodeWithResInfo(nodeID, resInfo)

	return nil
}
