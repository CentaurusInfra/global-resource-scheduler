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

package listers

import (
	schedulernodeinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// StackFilter is a function to filter a stack. If stack passed return true else return false.
type StackFilter func(stack *types.Stack) bool

// StackLister interface represents anything that can list stacks for a scheduler.
type StackLister interface {
	// Returns the list of pods.
	List() ([]*types.Stack, error)
	// This is similar to "List()", but the returned slice does not
	// contain pods that don't pass `podFilter`.
	FilteredList(stackFilter StackFilter) ([]*types.Stack, error)
}

// NodeInfoLister interface represents anything that can list/get NodeInfo objects from node name.
type NodeInfoLister interface {
	// Returns the list of NodeInfos.
	List() ([]*schedulernodeinfo.NodeInfo, error)
	// Returns the list of NodeInfos of nodes with pods with affinity terms.
	HavePodsWithAffinityList() ([]*schedulernodeinfo.NodeInfo, error)
	// Returns the NodeInfo of the given node name.
	Get(nodeName string) (*schedulernodeinfo.NodeInfo, error)
}

// SharedLister groups scheduler-specific listers.
type SharedLister interface {
	Stacks() StackLister
	NodeInfos() NodeInfoLister
}
