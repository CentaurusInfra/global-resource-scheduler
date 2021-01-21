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

package network

import (
	"context"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "EipAvailability"

// EipAvailability is a plugin that eip availability
type EipAvailability struct{}

var _ interfaces.FilterPlugin = &EipAvailability{}

// Name returns name of the plugin.
func (pl *EipAvailability) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *EipAvailability) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack, siteCacheInfo *sitecacheinfo.SiteCacheInfo) *interfaces.Status {
	var eipCount = 0
	for _, server := range stack.Resources {
		if server.NeedEip {
			eipCount += server.Count
		}
	}

	if eipCount <= 0 {
		return nil
	}

	// TODO:check quota for each eip
	return nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &EipAvailability{}, nil
}
