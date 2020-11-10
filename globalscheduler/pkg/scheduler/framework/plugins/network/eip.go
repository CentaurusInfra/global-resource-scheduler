package network

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "EipAvailability"

// EipAvailability is a plugin that eip availability
type EipAvailability struct{}

var _ interfaces.FilterPlugin = &EipAvailability{}

// Name returns name of the plugin.
func (pl *EipAvailability) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *EipAvailability) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack, nodeInfo *nodeinfo.NodeInfo) *interfaces.Status {
	var eipCount = 0
	for _, server := range stack.Resources {
		if server.NeedEip {
			eipCount += server.Count
		}
	}

	if eipCount <= 0 {
		return nil
	}

	// TODO:check quota for each eip
	eipPoolsInterface, ok := informers.InformerFac.GetInformer(informers.EIPPOOLS).GetStore().Get(nodeInfo.Node().Region)
	if !ok {
		logger.Warn(ctx, "Node (%s/%s) has no eip pools.", nodeInfo.Node().SiteID, nodeInfo.Node().Region)
		return nil
	}

	eipPool, ok := eipPoolsInterface.(typed.EipPool)
	if !ok {
		msg := fmt.Sprintf("Node (%s/%s) eipPoolConvert failed.", nodeInfo.Node().SiteID, nodeInfo.Node().Region)
		logger.Error(ctx, msg)
		return interfaces.NewStatus(interfaces.Unschedulable, msg)
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
		logger.Debug(ctx, msg)
		return interfaces.NewStatus(interfaces.Unschedulable, msg)
	}

	freeEipCount := findEipPool.Size - findEipPool.Used

	if eipCount > freeEipCount {
		msg := fmt.Sprintf("Node (%s/%s) has no enough eips, requestCount(%d), freeCount(%d).",
			nodeInfo.Node().SiteID, nodeInfo.Node().Region, freeEipCount, eipCount)
		logger.Debug(ctx, msg)
		return interfaces.NewStatus(interfaces.Unschedulable, msg)
	}

	maxCount := freeEipCount / eipCount
	interfaces.UpdateNodeSelectorState(cycleState, nodeInfo.Node().SiteID,
		map[string]interface{}{"StackMaxCount": maxCount})

	return nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &EipAvailability{}, nil
}
