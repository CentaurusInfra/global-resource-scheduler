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

package algorithmprovider

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/flavor"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/locationandoperator"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/queuesort"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/regionandaz"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

func getDefaultConfig() *types.Plugins {
	return &types.Plugins{
		QueueSort: &types.PluginSet{
			Enabled: []types.Plugin{
				{Name: queuesort.Name},
			},
		},
		PreFilter: &types.PluginSet{
			Enabled: []types.Plugin{
				{Name: flavor.Name},
			},
		},
		Filter: &types.PluginSet{
			Enabled: []types.Plugin{},
		},
		PreScore: &types.PluginSet{
			Enabled: []types.Plugin{},
		},
		Score: &types.PluginSet{
			Enabled: []types.Plugin{},
		},
		Bind: &types.PluginSet{
			Enabled: []types.Plugin{
				{Name: defaultbinder.Name},
			},
		},
		Strategy: &types.PluginSet{
			Enabled: []types.Plugin{
				{Name: regionandaz.Name},
				{Name: locationandoperator.Name},
			},
		},
	}
}

func GetPlugins(policy types.Policy) *types.Plugins {

	defaultPlugins := getDefaultConfig()
	for _, predicate := range policy.Predicates {
		defaultPlugins.Filter.Enabled = append(defaultPlugins.Filter.Enabled, types.Plugin{Name: predicate.Name})
	}

	for _, prioritie := range policy.Priorities {
		defaultPlugins.Score.Enabled = append(defaultPlugins.Score.Enabled, types.Plugin{Name: prioritie.Name,
			Weight: int32(prioritie.Weight)})
	}

	return defaultPlugins
}
