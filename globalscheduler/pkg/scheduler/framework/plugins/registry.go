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

package plugins

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/exclusivenode"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/flavor"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/locationandoperator"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/network"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/nodeavailability"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/noderesources"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/queuesort"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/regionandaz"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/stackaffinity"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/volume"
)

// NewRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewRegistry() interfaces.Registry {
	return interfaces.Registry{
		defaultbinder.Name:               defaultbinder.New,
		flavor.Name:                      flavor.New,
		locationandoperator.Name:         locationandoperator.New,
		network.Name:                     network.New,
		nodeavailability.Name:            nodeavailability.New,
		noderesources.FitName:            noderesources.NewFit,
		queuesort.Name:                   queuesort.New,
		regionandaz.Name:                 regionandaz.New,
		volume.Name:                      volume.New,
		noderesources.LeastAllocatedName: noderesources.NewLeastAllocated,
		stackaffinity.Name:               stackaffinity.New,
		exclusivenode.Name:               exclusivenode.New,
	}
}
