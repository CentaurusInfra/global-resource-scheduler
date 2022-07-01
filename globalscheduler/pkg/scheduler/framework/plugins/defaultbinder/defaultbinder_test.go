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

package defaultbinder

import (
	"context"
	"fmt"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/flavor"
	internalcache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"testing"
)

func TestBindResource(t *testing.T) {
	siteDelimeter := constants.SiteDelimiter
	flavorDelimiter := constants.FlavorDelimiter
	az := []string{"az-1", "az-2"}
	cloudRegion1 := types.CloudRegion{Region: "NW-1", AvailabilityZone: az}
	cloudRegion2 := types.CloudRegion{Region: "SE-1", AvailabilityZone: az}
	regions := []types.CloudRegion{cloudRegion1, cloudRegion2}
	selector1 := types.Selector{
		GeoLocation: types.GeoLocation{
			Country:  "US",
			Area:     "NW-1",
			Province: "Washington",
			City:     "Bellevue",
		},
		Regions:  regions,
		Operator: "globalscheduler",
		SiteID:   "NW-1" + siteDelimeter + "az-1",
	}

	selector2 := types.Selector{
		GeoLocation: types.GeoLocation{
			Country:  "US",
			Area:     "SE-1",
			Province: "Florida",
			City:     "Orlando",
		},
		Regions:  regions,
		Operator: "globalscheduler",
		SiteID:   "SE-1" + siteDelimeter + "az-1",
	}

	flavor1 := types.Flavor{FlavorID: "42"}
	flavors1 := []types.Flavor{flavor1}
	resource1 := &types.Resource{
		Name:         "vm-1",
		ResourceType: "vm",
		Storage:      nil,
		Flavors:      flavors1,
	}
	resources1 := []*types.Resource{resource1}

	flavor2 := types.Flavor{FlavorID: "2"}
	flavors2 := []types.Flavor{flavor2}
	resource2 := &types.Resource{
		Name:         "vm-2",
		ResourceType: "vm",
		Storage:      nil,
		Flavors:      flavors2,
	}
	resources2 := []*types.Resource{resource2}

	siteId := "NW-1--az-1"
	site := &types.Site{
		SiteID:           siteId,
		ClusterNamespace: "default",
		ClusterName:      "cluster1",
	}
	siteCacheInfo := &sitecacheinfo.SiteCacheInfo{
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
	siteCacheInfo.AllocatableFlavor["42"] = int64(10)
	flavor_nano := typed.Flavor{
		ID:    "42",
		Name:  "m1.nano",
		Vcpus: "1",
		Ram:   int64(128),
		Disk:  "0",
	}
	regionflavor_NW := &typed.RegionFlavor{
		RegionFlavorID: "NW-1" + flavorDelimiter + "42",
		Region:         "NW-1",
		Flavor:         flavor_nano,
	}
	flavorsMap := make(map[string]*typed.RegionFlavor)
	flavorsMap["42"] = &typed.RegionFlavor{
		RegionFlavorID: "42",
		Region:         "",
		Flavor:         flavor_nano,
	}
	regionFlavorsMap := make(map[string]*typed.RegionFlavor)
	regionFlavorsMap["NW-1"+flavorDelimiter+"42"] = regionflavor_NW

	tmpContext := context.Background()
	state := interfaces.NewCycleState()
	val1 := &flavor.FlavorAggregates{Aggregates: []flavor.FlavorAggregate{{Flavors: []interfaces.FlavorInfo{{Flavor: types.Flavor{FlavorID: "42", Spot: (*types.Spot)(nil)}, Count: 1}}, Weight: 10}}}
	val2 := &interfaces.SiteSelectorInfos{"NW-1" + siteDelimeter + "az-1": interfaces.SiteSelectorInfo{Flavors: []interfaces.FlavorInfo{{Flavor: types.Flavor{FlavorID: "42", Spot: (*types.Spot)(nil)}, Count: 1}}, VolumeType: []string(nil), StackMaxCount: 10}}
	state.Write("PreFilterFlavor", val1)
	state.Write("SiteTempSelected", val2)

	table := []struct {
		stack           *types.Stack
		flavorMap       map[string]*typed.RegionFlavor
		regionFlavorMap map[string]*typed.RegionFlavor
		expected        bool
	}{
		{
			stack: &types.Stack{
				PodName:      "pod1",
				PodNamespace: "default",
				Selector:     selector1,
				Resources:    resources1,
			},
			flavorMap:       flavorsMap,
			regionFlavorMap: regionFlavorsMap,
			expected:        true,
		},
		{
			stack: &types.Stack{
				PodName:      "pod2",
				PodNamespace: "default",
				Selector:     selector2,
				Resources:    resources1,
			},
			flavorMap:       flavorsMap,
			regionFlavorMap: regionFlavorsMap,
			expected:        false,
		},
		{
			stack: &types.Stack{
				PodName:      "pod3",
				PodNamespace: "default",
				Selector:     selector1,
				Resources:    resources2,
			},
			flavorMap:       flavorsMap,
			regionFlavorMap: regionFlavorsMap,
			expected:        false,
		},
	}
	for _, test := range table {
		testname := fmt.Sprintf("%v", test.stack.PodName)
		t.Run(testname, func(t *testing.T) {
			internalcache.FlavorCache.UpdateFlavorMap(test.regionFlavorMap, test.flavorMap)
			binder := &DefaultBinder{}
			_, siteID, flavorID, resInfo := binder.BindResource(tmpContext, state, test.stack, siteCacheInfo)
			testResult := false
			if siteID == test.stack.Selector.SiteID && flavorID == test.stack.Resources[0].Flavors[0].FlavorID && resInfo != nil {
				testResult = true
			}
			if testResult != test.expected {
				t.Errorf("TestBindResource() = %v, expected = %v", testResult, test.expected)
			}
		})
	}
}
