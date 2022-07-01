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

package sitecacheinfo

import (
	"fmt"
	"testing"
)

func TestUpdateFlavorCount(t *testing.T) {
	table := []struct {
		allocatableFlavor map[string]int64
		requestedFlavor   map[string]int64
		deduct            bool
		expected          int64
	}{
		{
			allocatableFlavor: map[string]int64{"42": int64(1)},
			requestedFlavor:   map[string]int64{"42": int64(1)},
			deduct:            true,
			expected:          0,
		},
		{
			allocatableFlavor: map[string]int64{"42": int64(1)},
			requestedFlavor:   map[string]int64{"42": int64(1)},
			deduct:            false,
			expected:          2,
		},
	}

	for _, test := range table {
		testname := fmt.Sprintf("%v", test.deduct)
		t.Run(testname, func(t *testing.T) {
			siteCacheInfo := &SiteCacheInfo{
				AllocatableFlavor: test.allocatableFlavor,
				RequestedFlavor:   test.requestedFlavor,
			}
			siteCacheInfo.updateFlavorCount(test.deduct)
			testResult := siteCacheInfo.AllocatableFlavor["42"]
			if testResult != test.expected {
				t.Errorf("TestUpdateFlavorCount() = %v, expected = %v", testResult, test.expected)
			}
		})
	}
}
