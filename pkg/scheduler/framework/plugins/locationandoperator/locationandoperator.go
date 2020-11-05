package locationandoperator

import (
	"context"
	"math"

	"k8s.io/kubernetes/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "LocationAndOperator"

// PrioritySort is a plugin that implements Priority based sorting.
type LocationAndOperator struct{}

var _ interfaces.FilterPlugin = &LocationAndOperator{}

// Name returns name of the plugin.
func (pl *LocationAndOperator) Name() string {
	return Name
}

func (pl *LocationAndOperator) locationEqual(stack *types.Stack, nodeInfo *nodeinfo.NodeInfo) (bool, *interfaces.Status) {
	if stack == nil || nodeInfo == nil {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "locationEqual has invalid args.")
	}

	country := stack.Selector.GeoLocation.Country
	if country != "" && country != nodeInfo.Node().Country {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack country not equal node county.")
	}

	area := stack.Selector.GeoLocation.Area
	if area != "" && area != nodeInfo.Node().Area {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack area not equal node area.")
	}
	province := stack.Selector.GeoLocation.Province
	if province != "" && province != nodeInfo.Node().Province {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack province not equal node province.")
	}

	city := stack.Selector.GeoLocation.City
	if city != "" && city != nodeInfo.Node().City {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack city not equal node city.")
	}

	return true, nil
}

// Filter invoked at the filter extension point.
func (pl *LocationAndOperator) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	nodeInfo *nodeinfo.NodeInfo) *interfaces.Status {

	if ok, status := pl.locationEqual(stack, nodeInfo); !ok {
		return status
	}

	if stack.Selector.Operator != "" && stack.Selector.Operator != nodeInfo.Node().Operator {
		return interfaces.NewStatus(interfaces.Unschedulable, "stack operator not equal node operator.")
	}

	return nil
}

//Strategy run strategy
func (pl *LocationAndOperator) Strategy(ctx context.Context, state *interfaces.CycleState,
	allocations *types.Allocation, nodeList interfaces.NodeScoreList) (interfaces.NodeScoreList, *interfaces.Status) {
	var count = allocations.Replicas
	var currentCount = count
	var nodeCount interfaces.NodeScoreList

	var totalCount = 0
	for _, node := range nodeList {
		selectorInfo, err := interfaces.GetNodeSelectorState(state, node.Name)
		if err != nil {
			logger.Error(ctx, "GetNodeSelectorState %s failed! err: %s", node.Name, err)
			continue
		}

		totalCount += selectorInfo.StackMaxCount
	}

	if count > totalCount {
		logger.Error(ctx, "total stack count :%d, request stack count: %d, not support!", totalCount, count)
		return nil, interfaces.NewStatus(interfaces.Unschedulable, "not find host")
	}

	if allocations.Selector.Strategy.LocationStrategy == constants.StrategyCentralize {
		for _, node := range nodeList {
			selectorInfo, err := interfaces.GetNodeSelectorState(state, node.Name)
			if err != nil {
				logger.Error(ctx, "GetNodeSelectorState %s failed! err: %s", node.Name, err)
				continue
			}

			stackCount := math.Min(float64(count), float64(selectorInfo.StackMaxCount))
			node.StackCount = int(stackCount)
			nodeCount = append(nodeCount, node)
			count -= int(stackCount)
			if count <= 0 {
				break
			}
		}
	} else {
		for _, node := range nodeList {
			selectorInfo, err := interfaces.GetNodeSelectorState(state, node.Name)
			if err != nil {
				logger.Error(ctx, "GetNodeSelectorState %s failed! err: %s", node.Name, err)
				continue
			}

			stackCount := math.Max(math.Floor(float64(count)*float64(selectorInfo.StackMaxCount)/float64(totalCount)+0.5),
				1)

			stackCount = math.Min(stackCount, float64(selectorInfo.StackMaxCount))
			node.StackCount = int(stackCount)
			nodeCount = append(nodeCount, node)

			currentCount -= int(stackCount)
			if currentCount <= 0 {
				break
			}
		}
	}

	return nodeCount, nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &LocationAndOperator{}, nil
}
