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
package distributor

import (
	allocv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/allocation/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	"testing"
)

type Testcase struct {
	SchedulerGeoLocation  clustercrdv1.GeolocationInfo
	AllocationGeoLocation allocv1.GeoLocation
	ExpectedResult        bool
}

func createSchedulerGeoLocation(city, province, area, country string) clustercrdv1.GeolocationInfo {
	return clustercrdv1.GeolocationInfo{city, province, area, country}
}

func createAllocationGeoLocation(city, province, area, country string) allocv1.GeoLocation {
	return allocv1.GeoLocation{city, province, area, country}
}

func TestIsAllocationGeoLocationMatched(t *testing.T) {
	testcases := []Testcase{
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("Bellevue", "WA", "NW", "US"),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("", "", "", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("New York", "NY", "NE", "US"),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("Bellevue", "WA", "NW", "CN"),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("Bellevue", "WA", "NE", "US"),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("Bellevue", "OR", "NW", "US"),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", "US"),
			createAllocationGeoLocation("Seattle", "WA", "NW", "US"),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", ""),
			createAllocationGeoLocation("Bellevue", "WA", "NW", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "", ""),
			createAllocationGeoLocation("Bellevue", "WA", "", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "", "", ""),
			createAllocationGeoLocation("Bellevue", "", "", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "NW", ""),
			createAllocationGeoLocation("Bellevue", "WA", "NE", ""),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "", ""),
			createAllocationGeoLocation("Bellevue", "OR", "", ""),
			false,
		},
		{createSchedulerGeoLocation("Bellevue", "", "", "US"),
			createAllocationGeoLocation("Bellevue", "", "", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "", "NW", ""),
			createAllocationGeoLocation("Bellevue", "", "", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "", ""),
			createAllocationGeoLocation("Bellevue", "", "", ""),
			true,
		},
		{createSchedulerGeoLocation("Bellevue", "WA", "", "US"),
			createAllocationGeoLocation("Bellevue", "WA", "", ""),
			true,
		},
	}
	for _, testcase := range testcases {
		res := isAllocationGeoLocationMatched(&testcase.SchedulerGeoLocation, testcase.AllocationGeoLocation)
		if res != testcase.ExpectedResult {
			t.Errorf("The isAllocationGeoLocationMatched test result %v is not empty as expected with geoLocations %v, %v",
				res, testcase.SchedulerGeoLocation, testcase.AllocationGeoLocation)
		}
	}
}
