package stackaffinity

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/pkg/scheduler/labels"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "StackAffinity"

// VolumeType is a plugin that filter volume type

type StackAffinity struct{ handle interfaces.FrameworkHandle }

var _ interfaces.ScorePlugin = &StackAffinity{}

// Name returns name of the plugin.
func (pl *StackAffinity) Name() string {
	return Name
}

func getAffinityScore(ctx context.Context, stack *types.Stack, nodeInfo *nodeinfo.NodeInfo) (int64, *interfaces.Status) {
	selector := labels.NewSelector()
	if len(stack.Selector.StackAffinity) > 0 {
		for _, aff := range stack.Selector.StackAffinity {
			r, err := labels.NewRequirement(aff.Key, labels.Operator(aff.Operator), aff.StrValues)
			if err != nil {
				logger.Error(ctx, "NewRequirement requirement(%s) failed! err : %s", aff, err)
				return 0, interfaces.NewStatus(interfaces.Error, fmt.Sprintf("requirement failed."))
			}
			selector = selector.Add(*r)
		}
	}

	if len(nodeInfo.Stacks()) == 0 || selector.Empty() {
		return 0, nil
	}

	var score int64
	for _, existingStack := range nodeInfo.Stacks() {
		if selector.Matches(labels.Set(existingStack.Labels)) {
			score += types.MaxNodeScore
			break
		}
	}

	return score, nil
}

func getAntiAffinityScore(ctx context.Context, stack *types.Stack, nodeInfo *nodeinfo.NodeInfo) (int64, *interfaces.Status) {
	selector := labels.NewSelector()
	if len(stack.Selector.StackAntiAffinity) > 0 {
		for _, aff := range stack.Selector.StackAntiAffinity {
			r, err := labels.NewRequirement(aff.Key, labels.Operator(aff.Operator), aff.StrValues)
			if err != nil {
				logger.Error(ctx, "NewRequirement requirement(%s) failed! err : %s", aff, err)
				return 0, interfaces.NewStatus(interfaces.Error, fmt.Sprintf("requirement failed."))
			}
			selector = selector.Add(*r)
		}
	}

	if len(nodeInfo.Stacks()) == 0 || selector.Empty() {
		return types.MaxNodeScore, nil
	}

	for _, existingStack := range nodeInfo.Stacks() {
		if selector.Matches(labels.Set(existingStack.Labels)) {
			return 0, nil
		}
	}

	return types.MaxNodeScore, nil
}

// Score invoked at the score extension point.
func (pl *StackAffinity) Score(ctx context.Context, state *interfaces.CycleState,
	stack *types.Stack, nodeID string) (int64, *interfaces.Status) {
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeID)
	if err != nil {
		return 0, interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting node %s from Snapshot: %v",
			nodeInfo.Node().SiteID, err))
	}

	node := nodeInfo.Node()
	if node == nil {
		return 0, interfaces.NewStatus(interfaces.Error, fmt.Sprintf("node not found"))
	}

	affinityScore, status := getAffinityScore(ctx, stack, nodeInfo)
	if status != nil {
		logger.Error(ctx, "getAffinityScore failed! err: %s", state)
		return 0, status
	}

	AntiaffinityScore, status := getAntiAffinityScore(ctx, stack, nodeInfo)
	if status != nil {
		logger.Error(ctx, "getAffinityScore failed! err: %s", state)
		return 0, status
	}

	return (affinityScore + AntiaffinityScore) / 2, nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &StackAffinity{
		handle: handle,
	}, nil
}
