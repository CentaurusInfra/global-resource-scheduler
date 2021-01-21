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

package volume

import (
	"context"
	"fmt"
	"math"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "Volume"

// VolumeType is a plugin that filter volume type

type Volume struct {
	handle interfaces.FrameworkHandle
}

var _ interfaces.FilterPlugin = &Volume{}
var _ interfaces.ScorePlugin = &Volume{}

// Name returns name of the plugin.
func (pl *Volume) Name() string {
	return Name
}

func calculateStackStorageRequest(stack *types.Stack) map[string]float64 {
	var ret = map[string]float64{}
	for _, server := range stack.Resources {
		for volType, size := range server.Storage {
			if _, ok := ret[volType]; ok {
				ret[volType] = 0
			}

			count := server.Count
			ret[volType] += float64(count) * float64(size)
		}
	}

	return ret
}

// Filter invoked at the filter extension point.
func (pl *Volume) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) *interfaces.Status {

	var maxCount float64 = math.MaxFloat64
	stackStorageRequest := calculateStackStorageRequest(stack)

	for volType, size := range stackStorageRequest {
		var allocatableSize float64
		var requestedSize float64
		var ok bool
		if allocatableSize, ok = siteCacheInfo.TotalStorage[volType]; !ok {
			msg := fmt.Sprintf("Site (%s) do not support required volume(%s).Support volume(%v)",
				siteCacheInfo.GetSite().SiteID, volType, siteCacheInfo.TotalStorage)
			logger.Debug(ctx, msg)
			return interfaces.NewStatus(interfaces.Unschedulable, msg)
		}

		if requestedSize, ok = siteCacheInfo.RequestedStorage[volType]; !ok {
			requestedSize = 0
		}

		if allocatableSize < requestedSize+size {
			msg := fmt.Sprintf("Site (%s) do not support required volume(%s-%f).Support volume(%v)",
				siteCacheInfo.GetSite().SiteID, volType, size, siteCacheInfo.TotalStorage)
			logger.Debug(ctx, msg)
			return interfaces.NewStatus(interfaces.Unschedulable, msg)
		}

		if size > 0 {
			maxCount = math.Min(maxCount, float64((allocatableSize-requestedSize)/size))
		}

		maxCount = math.Min(maxCount, float64((allocatableSize-requestedSize)/size))
	}
	interfaces.UpdateSiteSelectorState(cycleState, siteCacheInfo.GetSite().SiteID,
		map[string]interface{}{"StackMaxCount": maxCount})
	return nil
}

// Score invoked at the score extension point.
func (pl *Volume) Score(ctx context.Context, state *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) (int64, *interfaces.Status) {
	var requestedTotalSize float64 = 0
	var allocatableTotalSize float64 = 0

	for _, value := range siteCacheInfo.RequestedStorage {
		requestedTotalSize += value
	}

	for _, value := range siteCacheInfo.TotalStorage {
		allocatableTotalSize += value
	}

	if allocatableTotalSize <= 0 {
		return 0, nil
	}

	score := ((allocatableTotalSize - requestedTotalSize) * float64(interfaces.MaxSiteScore)) / allocatableTotalSize
	return int64(score), nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &Volume{handle: handle}, nil
}
