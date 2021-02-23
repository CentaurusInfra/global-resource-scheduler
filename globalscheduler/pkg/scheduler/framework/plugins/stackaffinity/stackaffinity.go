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

package stackaffinity

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/labels"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
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

func getAffinityScore(ctx context.Context, stack *types.Stack, siteCacheInfo *sitecacheinfo.SiteCacheInfo) (int64, *interfaces.Status) {
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

	if len(siteCacheInfo.Stacks()) == 0 || selector.Empty() {
		return 0, nil
	}

	var score int64
	for _, existingStack := range siteCacheInfo.Stacks() {
		if selector.Matches(labels.Set(existingStack.Labels)) {
			score += types.MaxSiteScore
			break
		}
	}

	return score, nil
}

func getAntiAffinityScore(ctx context.Context, stack *types.Stack, siteCacheInfo *sitecacheinfo.SiteCacheInfo) (int64, *interfaces.Status) {
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

	if len(siteCacheInfo.Stacks()) == 0 || selector.Empty() {
		return types.MaxSiteScore, nil
	}

	for _, existingStack := range siteCacheInfo.Stacks() {
		if selector.Matches(labels.Set(existingStack.Labels)) {
			return 0, nil
		}
	}

	return types.MaxSiteScore, nil
}

// Score invoked at the score extension point.
func (pl *StackAffinity) Score(ctx context.Context, state *interfaces.CycleState,
	stack *types.Stack, siteCacheInfo *sitecacheinfo.SiteCacheInfo) (int64, *interfaces.Status) {
	site := siteCacheInfo.GetSite()
	if site == nil {
		return 0, interfaces.NewStatus(interfaces.Error, fmt.Sprintf("site not found"))
	}

	affinityScore, status := getAffinityScore(ctx, stack, siteCacheInfo)
	if status != nil {
		logger.Error(ctx, "getAffinityScore failed! err: %s", state)
		return 0, status
	}

	AntiaffinityScore, status := getAntiAffinityScore(ctx, stack, siteCacheInfo)
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
