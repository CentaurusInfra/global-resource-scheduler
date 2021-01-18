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
	"testing"

	fakecache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache/fake"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
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
				Name: "stack01",
			},
			isAssumedStackFunc: func(*types.Stack) bool { return true },
			getStackFunc: func(*types.Stack) *types.Stack {
				return &types.Stack{
					Name: "stack01",
				}
			},
			expected: true,
		},
		{
			name: "Assumed stack with different stack",
			stack: &types.Stack{
				Name: "stack01",
			},
			isAssumedStackFunc: func(*types.Stack) bool { return true },
			getStackFunc: func(*types.Stack) *types.Stack {
				return &types.Stack{
					Name: "stack02",
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
