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

package nodeavailability

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "NodeAvailability"

// NodeAvailability is a plugin that node availability
type NodeAvailability struct{}

var _ interfaces.FilterPlugin = &NodeAvailability{}

// Name returns name of the plugin.
func (pl *NodeAvailability) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *NodeAvailability) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	nodeInfo *nodeinfo.NodeInfo) *interfaces.Status {

	if nodeInfo.Node().Status == constants.SiteStatusOffline || nodeInfo.Node().Status == constants.SiteStatusSellout {
		msg := fmt.Sprintf("Node(%s) status is %s, not available!", nodeInfo.Node().SiteID, nodeInfo.Node().Status)
		logger.Debugf(msg)
		return interfaces.NewStatus(interfaces.Unschedulable, msg)
	}

	if stack.Selector.NodeID != "" && stack.Selector.NodeID != nodeInfo.Node().SiteID {
		msg := fmt.Sprintf("Node(%s) not suitable!", nodeInfo.Node().SiteID)
		logger.Debugf(msg)
		return interfaces.NewStatus(interfaces.Unschedulable, msg)
	}

	return nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &NodeAvailability{}, nil
}
