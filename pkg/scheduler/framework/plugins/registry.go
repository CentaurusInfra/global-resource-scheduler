package plugins

import (
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/exclusivenode"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/flavor"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/locationandoperator"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/network"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodeavailability"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/noderesources"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/regionandaz"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/stackaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/volume"
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
