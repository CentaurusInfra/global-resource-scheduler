package nodeavailability

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
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
