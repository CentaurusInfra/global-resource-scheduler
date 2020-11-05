package algorithmprovider

import (
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/flavor"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/locationandoperator"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/regionandaz"
	"k8s.io/kubernetes/pkg/scheduler/types"
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
