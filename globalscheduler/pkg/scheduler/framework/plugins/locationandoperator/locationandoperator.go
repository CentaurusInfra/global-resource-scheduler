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

package locationandoperator

import (
	"context"
	"math"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
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

func (pl *LocationAndOperator) locationEqual(stack *types.Stack, siteCacheInfo *sitecacheinfo.SiteCacheInfo) (bool, *interfaces.Status) {
	if stack == nil || siteCacheInfo == nil {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "locationEqual has invalid args.")
	}

	country := stack.Selector.GeoLocation.Country
	if country != "" && country != siteCacheInfo.GetSite().Country {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack country not equal site county.")
	}

	area := stack.Selector.GeoLocation.Area
	if area != "" && area != siteCacheInfo.GetSite().Area {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack area not equal site area.")
	}
	province := stack.Selector.GeoLocation.Province
	if province != "" && province != siteCacheInfo.GetSite().Province {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack province not equal site province.")
	}

	city := stack.Selector.GeoLocation.City
	if city != "" && city != siteCacheInfo.GetSite().City {
		return false, interfaces.NewStatus(interfaces.Unschedulable, "stack city not equal site city.")
	}

	return true, nil
}

// Filter invoked at the filter extension point.
func (pl *LocationAndOperator) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) *interfaces.Status {

	if ok, status := pl.locationEqual(stack, siteCacheInfo); !ok {
		return status
	}

	if stack.Selector.Operator != "" && stack.Selector.Operator != siteCacheInfo.GetSite().Operator {
		return interfaces.NewStatus(interfaces.Unschedulable, "stack operator not equal site operator.")
	}

	return nil
}

//Strategy run strategy
func (pl *LocationAndOperator) Strategy(ctx context.Context, state *interfaces.CycleState,
	allocations *types.Allocation, siteScoreList interfaces.SiteScoreList) (interfaces.SiteScoreList, *interfaces.Status) {
	var count = allocations.Replicas
	var currentCount = count
	var siteCount interfaces.SiteScoreList

	var totalCount = 0
	for _, siteScore := range siteScoreList {
		selectorInfo, err := interfaces.GetSiteSelectorState(state, siteScore.SiteID)
		if err != nil {
			logger.Error(ctx, "GetSiteSelectorState %s failed! err: %s", siteScore.SiteID, err)
			continue
		}

		totalCount += selectorInfo.StackMaxCount
	}

	if count > totalCount {
		logger.Error(ctx, "total stack count :%d, request stack count: %d, not support!", totalCount, count)
		return nil, interfaces.NewStatus(interfaces.Unschedulable, "not find host")
	}

	if allocations.Selector.Strategy.LocationStrategy == constants.StrategyCentralize {
		for _, siteScore := range siteScoreList {
			selectorInfo, err := interfaces.GetSiteSelectorState(state, siteScore.SiteID)
			if err != nil {
				logger.Error(ctx, "GetSiteSelectorState %s failed! err: %s", siteScore.SiteID, err)
				continue
			}

			stackCount := math.Min(float64(count), float64(selectorInfo.StackMaxCount))
			siteScore.StackCount = int(stackCount)
			siteCount = append(siteCount, siteScore)
			count -= int(stackCount)
			if count <= 0 {
				break
			}
		}
	} else {
		for _, siteScore := range siteScoreList {
			selectorInfo, err := interfaces.GetSiteSelectorState(state, siteScore.SiteID)
			if err != nil {
				logger.Error(ctx, "GetSiteSelectorState %s failed! err: %s", siteScore.SiteID, err)
				continue
			}

			stackCount := math.Max(math.Floor(float64(count)*float64(selectorInfo.StackMaxCount)/float64(totalCount)+0.5),
				1)

			stackCount = math.Min(stackCount, float64(selectorInfo.StackMaxCount))
			siteScore.StackCount = int(stackCount)
			siteCount = append(siteCount, siteScore)

			currentCount -= int(stackCount)
			if currentCount <= 0 {
				break
			}
		}
	}

	return siteCount, nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &LocationAndOperator{}, nil
}
