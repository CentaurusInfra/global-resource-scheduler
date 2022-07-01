/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package scheduler

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	internalcache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	fakecache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache/fake"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"testing"
)

func TestSkipStackUpdate(t *testing.T) {
	table := []struct {
		name               string
		stack              *types.Stack
		isAssumedStackFunc func(*types.Stack) bool
		getStackFunc       func(*types.Stack) *types.Stack
		expected           bool
	}{
		{
			name:               "Non-assumed stack",
			stack:              &types.Stack{},
			isAssumedStackFunc: func(*types.Stack) bool { return false },
			expected:           false,
		},
		{
			name: "Assumed stack with same stack",
			stack: &types.Stack{
				PodName: "stack01",
			},
			isAssumedStackFunc: func(*types.Stack) bool { return true },
			getStackFunc: func(*types.Stack) *types.Stack {
				return &types.Stack{
					PodName: "stack01",
				}
			},
			expected: true,
		},
		{
			name: "Assumed stack with different stack",
			stack: &types.Stack{
				PodName: "stack01",
			},
			isAssumedStackFunc: func(*types.Stack) bool { return true },
			getStackFunc: func(*types.Stack) *types.Stack {
				return &types.Stack{
					PodName: "stack02",
				}
			},
			expected: false,
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			sched := &Scheduler{
				SchedulerCache: &fakecache.Cache{
					IsAssumedStackFunc: test.isAssumedStackFunc,
					GetStackFunc:       test.getStackFunc,
				},
			}
			got := sched.skipStackUpdate(test.stack)
			if got != test.expected {
				t.Errorf("TestSkipStackUpdate() = %t, expected = %t", got, false)
			}
		})
	}
}

func TestWithdrawResource(t *testing.T) {
	siteId := "SW-1||az-1"
	cm := types.CPUAndMemory{
		VCPU:   1,
		Memory: 128,
	}
	cpuAndMemMap := make(map[string]types.CPUAndMemory)
	cpuAndMemMap["default"] = cm
	storageMap := make(map[string]float64)
	storageMap["ssd"] = float64(128)
	allRes := types.AllResInfo{
		CpuAndMem: cpuAndMemMap,
		Storage:   storageMap,
	}

	podSiteResource := &PodSiteResource{
		PodName:  "pod1",
		SiteID:   siteId,
		Flavor:   "42",
		Resource: allRes,
	}
	podSiteResourceMap := make(map[string]*PodSiteResource)
	podSiteResourceMap["pod1"] = podSiteResource

	region := "SW-1"
	id := "42"
	name := "m1.nano"
	vcpus := "1"
	ram := int64(128)
	disk := "0"
	regionFlavorID := region + "--" + id
	flavor := &typed.RegionFlavor{
		RegionFlavorID: regionFlavorID,
		Region:         region,
		Flavor: typed.Flavor{
			ID:    id,
			Name:  name,
			Vcpus: vcpus,
			Ram:   ram,
			Disk:  disk,
		},
	}
	site := &types.Site{
		SiteID:           siteId,
		ClusterNamespace: "default",
		ClusterName:      "cluster1",
	}

	siteCacheInfo := &schedulersitecacheinfo.SiteCacheInfo{
		Site:                  site,
		RequestedResources:    make(map[string]*types.CPUAndMemory),
		TotalResources:        make(map[string]*types.CPUAndMemory),
		RequestedStorage:      make(map[string]float64),
		TotalStorage:          make(map[string]float64),
		RequestedFlavor:       make(map[string]int64),
		AllocatableFlavor:     make(map[string]int64),
		AllocatableSpotFlavor: make(map[string]types.SpotResource),
		Qos:                   make(map[string]float64),
	}
	SiteCacheInfoMap := make(map[string]*schedulersitecacheinfo.SiteCacheInfo)
	RegionFlavorMap := make(map[string]*typed.RegionFlavor)
	FlavorMap := make(map[string]*typed.RegionFlavor)
	FlavorMap[id] = flavor
	RegionFlavorMap[regionFlavorID] = flavor
	SiteCacheInfoMap[siteId] = siteCacheInfo
	table := []struct {
		podName          string
		siteCacheInfoMap map[string]*schedulersitecacheinfo.SiteCacheInfo
		regionFlavorMap  map[string]*typed.RegionFlavor
		flavorMap        map[string]*typed.RegionFlavor
		expected         bool
	}{
		{
			podName:          "pod1",
			siteCacheInfoMap: SiteCacheInfoMap,
			regionFlavorMap:  RegionFlavorMap,
			flavorMap:        FlavorMap,
			expected:         true,
		},
	}
	for _, test := range table {
		t.Run(test.podName, func(t *testing.T) {
			sched := &Scheduler{
				PodSiteResourceMap: podSiteResourceMap,
				siteCacheInfoSnapshot: &internalcache.Snapshot{
					SiteCacheInfoMap: test.siteCacheInfoMap,
					RegionFlavorMap:  test.regionFlavorMap,
					FlavorMap:        test.flavorMap,
				},
			}
			var nAllocatableFlavorBefore int64
			nAllocatableFlavorBefore = int64(0)
			siteInfoBefore, ok := sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteId]
			if ok {
				nAllocatableFlavorBefore = siteInfoBefore.AllocatableFlavor[id]
			}
			err := sched.withdrawResource(test.podName)
			if err != nil {
				t.Errorf("TestWithdrawResource() = %t, expected = %t", err, nil)
			}
			siteInfoAfter := sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteId]
			nAllocatableFlavorAfter := siteInfoAfter.AllocatableFlavor[id]
			if testResult := nAllocatableFlavorAfter >= nAllocatableFlavorBefore; testResult != test.expected {
				t.Errorf("TestWithdrawResource() = %v, expected = %v", testResult, test.expected)
			}
		})
	}
}
