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

package cache

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"reflect"
	"testing"
)

func TestGetRegionFlavors(t *testing.T) {
	delimiter := constants.FlavorDelimiter
	//referred from arktos/conf/flavor_config.yaml
	flavor1 := typed.Flavor{
		ID:    "1",
		Name:  "m1.tiny",
		Vcpus: "1",
		Ram:   int64(512),
		Disk:  "1",
	}
	flavor2 := typed.Flavor{
		ID:    "2",
		Name:  "m1.small",
		Vcpus: "1",
		Ram:   int64(2048),
		Disk:  "20",
	}
	flavor3 := typed.Flavor{
		ID:    "3",
		Name:  "m1.medium",
		Vcpus: "2",
		Ram:   int64(4096),
		Disk:  "40",
	}

	regionflavor1 := &typed.RegionFlavor{
		RegionFlavorID: "SE-1" + delimiter + "1",
		Region:         "SE-1",
		Flavor:         flavor1,
	}
	regionflavor2 := &typed.RegionFlavor{
		RegionFlavorID: "SE-1" + delimiter + "2",
		Region:         "SE-1",
		Flavor:         flavor2,
	}
	regionflavor3 := &typed.RegionFlavor{
		RegionFlavorID: "SE-1" + delimiter + "3",
		Region:         "SE-1",
		Flavor:         flavor3,
	}
	regionflavor4 := &typed.RegionFlavor{
		RegionFlavorID: "NW-1" + delimiter + "1",
		Region:         "NW-1",
		Flavor:         flavor1,
	}
	regionflavor5 := &typed.RegionFlavor{
		RegionFlavorID: "NW-1" + delimiter + "2",
		Region:         "NW-1",
		Flavor:         flavor2,
	}
	regionflavor6 := &typed.RegionFlavor{
		RegionFlavorID: "NW-1" + delimiter + "3",
		Region:         "NW-1",
		Flavor:         flavor3,
	}
	flavors := make(map[string]*typed.RegionFlavor)
	flavors["1"] = &typed.RegionFlavor{
		RegionFlavorID: "1",
		Region:         "",
		Flavor:         flavor1,
	}
	flavors["2"] = &typed.RegionFlavor{
		RegionFlavorID: "2",
		Region:         "",
		Flavor:         flavor2,
	}
	flavors["3"] = &typed.RegionFlavor{
		RegionFlavorID: "3",
		Region:         "",
		Flavor:         flavor3,
	}

	regionFlavors := make(map[string]*typed.RegionFlavor)
	regionFlavors["SE-1"+delimiter+"1"] = regionflavor1
	regionFlavors["SE-1"+delimiter+"2"] = regionflavor2
	regionFlavors["SE-1"+delimiter+"3"] = regionflavor3
	regionFlavors["NW-1"+delimiter+"1"] = regionflavor4
	regionFlavors["NW-1"+delimiter+"2"] = regionflavor5
	regionFlavors["NW-1"+delimiter+"3"] = regionflavor6

	regionFlavors_SE := make(map[string]*typed.RegionFlavor)
	regionFlavors_SE["SE-1"+delimiter+"1"] = regionflavor1
	regionFlavors_SE["SE-1"+delimiter+"2"] = regionflavor2
	regionFlavors_SE["SE-1"+delimiter+"3"] = regionflavor3

	regionFlavors_NW := make(map[string]*typed.RegionFlavor)
	regionFlavors_NW["NW-1"+delimiter+"1"] = regionflavor4
	regionFlavors_NW["NW-1"+delimiter+"2"] = regionflavor5
	regionFlavors_NW["NW-1"+delimiter+"3"] = regionflavor6

	table := []struct {
		regionName      string
		flavorMap       map[string]*typed.RegionFlavor
		regionFlavorMap map[string]*typed.RegionFlavor
		expected        map[string]*typed.RegionFlavor
	}{
		{
			regionName:      "SE-1",
			flavorMap:       flavors,
			regionFlavorMap: regionFlavors,
			expected:        regionFlavors_SE,
		},
		{
			regionName:      "NW-1",
			flavorMap:       flavors,
			regionFlavorMap: regionFlavors,
			expected:        regionFlavors_NW,
		},
		{
			regionName:      "Central-2",
			flavorMap:       flavors,
			regionFlavorMap: regionFlavors,
			expected:        make(map[string]*typed.RegionFlavor),
		},
		{
			regionName:      "",
			flavorMap:       flavors,
			regionFlavorMap: regionFlavors,
			expected:        make(map[string]*typed.RegionFlavor),
		},
	}

	for _, test := range table {
		t.Run(test.regionName, func(t *testing.T) {
			snapshot := &Snapshot{
				RegionFlavorMap: test.regionFlavorMap,
				FlavorMap:       test.flavorMap,
			}
			testResult, err := snapshot.GetRegionFlavors(test.regionName)
			if err != nil {
				t.Errorf("TestGetRegionFlavors() = %v, expected = %v", err, nil)
			}
			if reflect.DeepEqual(testResult, test.expected) == false {
				t.Errorf("TestGetRegionFlavors() = %v, expected = %v", testResult, test.expected)
			}
		})
	}
}
