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

package queuesort

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "PrioritySort"

func GetStackPriority(stack *types.Stack) int {
	return 1
}

// PrioritySort is a plugin that implements Priority based sorting.
type PrioritySort struct{}

var _ interfaces.QueueSortPlugin = &PrioritySort{}

// Name returns name of the plugin.
func (pl *PrioritySort) Name() string {
	return Name
}

// Less is the function used by the activeQ heap algorithm to sort pods.
// It sorts pods based on their priority. When priorities are equal, it uses
// PodInfo.timestamp.
func (pl *PrioritySort) Less(sInfo1, sInfo2 *interfaces.StackInfo) bool {
	p1 := GetStackPriority(sInfo1.Stack)
	p2 := GetStackPriority(sInfo2.Stack)
	return (p1 > p2) || (p1 == p2 && sInfo1.Timestamp.Before(sInfo2.Timestamp))
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &PrioritySort{}, nil
}
