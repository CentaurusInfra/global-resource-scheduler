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

package siteresources

import (
	"context"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// LeastAllocated is a score plugin that favors site with fewer allocation requested resources
// based on requested resources.
type LeastAllocated struct {
	handle interfaces.FrameworkHandle
	resourceAllocationScorer
}

var _ = interfaces.ScorePlugin(&LeastAllocated{})

// LeastAllocatedName is the name of the plugin used in the plugin registry and configurations.
const LeastAllocatedName = "SiteResourcesLeastAllocated"

// Name returns name of the plugin. It is used in logs, etc.
func (la *LeastAllocated) Name() string {
	return LeastAllocatedName
}

// Score invoked at the score extension point.
func (la *LeastAllocated) Score(ctx context.Context, state *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) (int64, *interfaces.Status) {
	// la.score favors site with fewer requested resources.
	// It calculates the percentage of memory and CPU requested by pods scheduled on the site, and
	// prioritizes based on the minimum of the average of the fraction of requested to capacity.
	//
	// Details:
	// (cpu((capacity-sum(requested))*10/capacity) + memory((capacity-sum(requested))*10/capacity))/2
	return la.score(stack, siteCacheInfo)
}

// NewLeastAllocated initializes a new plugin and returns it.
func NewLeastAllocated(h interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &LeastAllocated{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			LeastAllocatedName,
			leastResourceScorer,
			defaultRequestedRatioResources,
		},
	}, nil
}

func leastResourceScorer(requested, allocable resourceToValueMap, includeVolumes bool, requestedVolumes int,
	allocatableVolumes int) int64 {
	var score, weightSum int64
	for resource, weight := range defaultRequestedRatioResources {
		resourceScore := leastRequestedScore(requested[resource], allocable[resource])
		score += resourceScore * weight
		weightSum += weight
	}
	return score / weightSum
}

// The unused capacity is calculated on a scale of 0-10
// 0 being the lowest priority and 10 being the highest.
// The more unused resources the higher the score is.
func leastRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		return 0
	}

	return ((capacity - requested) * int64(interfaces.MaxSiteScore)) / capacity
}
