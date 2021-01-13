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

package cache

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/scheduler/listers"
	schedulernodeinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Cache collects pods' information and provides node-level aggregated information.
type Cache interface {
	schedulerlisters.StackLister

	// AssumeStack assumes a pod scheduled and aggregates the pod's information into its node.
	// The implementation also decides the policy to expire pod before being confirmed (receiving Add event).
	// After expiration, its information would be subtracted.
	AssumeStack(pod *types.Stack) error

	// AddStack either confirms a pod if it's assumed, or adds it back if it's expired.
	// If added back, the pod's information would be added again.
	AddStack(pod *types.Stack) error

	// UpdateStack removes oldStack's information and adds newStack's information.
	UpdateStack(oldStack, newStack *types.Stack) error

	// RemoveStack removes a pod. The pod's information would be subtracted from assigned node.
	RemoveStack(pod *types.Stack) error

	// GetStack returns the pod from the cache with the same namespace and the
	// same name of the specified pod.
	GetStack(pod *types.Stack) (*types.Stack, error)

	// IsAssumedStack returns true if the pod is assumed and not expired.
	IsAssumedStack(pod *types.Stack) (bool, error)

	// AddNode adds overall information about node.
	AddNode(node *types.SiteNode) error

	// UpdateNode updates overall information about node.
	UpdateNode(oldNode, newNode *types.SiteNode) error

	// RemoveNode removes overall information about node.
	RemoveNode(node *types.SiteNode) error

	//UpdateEipPool updates eip pool info about node
	UpdateEipPool(eipPool *typed.EipPool) error

	//UpdateVolumePool updates volume pool info about node
	UpdateVolumePool(volumePool *typed.RegionVolumePool) error

	// UpdateNodeWithResInfo update res info
	UpdateNodeWithResInfo(siteID string, resInfo types.AllResInfo) error

	//UpdateQos updates qos info
	UpdateQos(siteID string, netMetricData *types.NetMetricDatas) error

	//UpdateNodeWithVcpuMem update vcpu and mem
	UpdateNodeWithRatio(region string, az string, ratios []types.AllocationRatio) error

	//UpdateSpotResources update spot resources
	UpdateSpotResources(region string, az string, spotRes map[string]types.SpotResource) error

	// UpdateSnapshot updates the passed infoSnapshot to the current contents of Cache.
	// The node info contains aggregated information of pods scheduled (including assumed to be)
	// on this node.
	UpdateSnapshot(nodeSnapshot *Snapshot) error

	//GetRegions get cache region info
	GetRegions() map[string]types.CloudRegion

	//PrintString print node info
	PrintString()

	// Dump produces a dump of the current cache.
	Dump() *Dump
}

// Dump is a dump of the cache state.
type Dump struct {
	AssumedStacks map[string]bool
	Nodes         map[string]*schedulernodeinfo.NodeInfo
}
