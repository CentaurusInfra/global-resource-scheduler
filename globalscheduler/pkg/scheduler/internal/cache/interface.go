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
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Cache collects pods' information and provides site-level aggregated information.
type Cache interface {
	schedulerlisters.StackLister

	// AssumeStack assumes a stack scheduled and aggregates the pod's information into its site.
	// The implementation also decides the policy to expire pod before being confirmed (receiving Add event).
	// After expiration, its information would be subtracted.
	AssumeStack(stack *types.Stack) error

	// ForgetStack removes an assumed stack from cache.
	ForgetStack(stack *types.Stack) error

	// AddStack either confirms a stack if it's assumed, or adds it back if it's expired.
	// If added back, the pod's information would be added again.
	AddStack(stack *types.Stack) error

	// UpdateStack removes oldStack's information and adds newStack's information.
	UpdateStack(oldStack, newStack *types.Stack) error

	// RemoveStack removes a stack. The stack's information would be subtracted from assigned site.
	RemoveStack(stack *types.Stack) error

	// GetStack returns the pod from the cache with the same namespace and the
	// same name of the specified pod.
	GetStack(stack *types.Stack) (*types.Stack, error)

	// IsAssumedStack returns true if the stack is assumed and not expired.
	IsAssumedStack(stack *types.Stack) (bool, error)

	// AddSite adds overall information about site.
	AddSite(site *types.Site) error

	// UpdateSite updates overall information about site.
	UpdateSite(oldSite, newSite *types.Site) error

	// RemoveSite removes overall information about site.
	RemoveSite(site *types.Site) error

	//UpdateEipPool updates eip pool info about site
	UpdateEipPool(eipPool *typed.EipPool) error

	//UpdateVolumePool updates volume pool info about site
	UpdateVolumePool(volumePool *typed.RegionVolumePool) error

	// UpdateSiteWithResInfo update res info
	UpdateSiteWithResInfo(siteID string, resInfo types.AllResInfo) error

	//UpdateQos updates qos info
	UpdateQos(siteID string, netMetricData *types.NetMetricDatas) error

	//UpdateSiteWithVcpuMem update vcpu and mem
	UpdateSiteWithRatio(region string, az string, ratios []types.AllocationRatio) error

	//UpdateSpotResources update spot resources
	UpdateSpotResources(region string, az string, spotRes map[string]types.SpotResource) error

	// UpdateSnapshot updates the passed infoSnapshot to the current contents of Cache.
	// The site info contains aggregated information of pods scheduled (including assumed to be)
	// on this site.
	UpdateSnapshot(snapshot *Snapshot) error

	//GetRegions get cache region info
	GetRegions() map[string]types.CloudRegion

	//PrintString print site cache info
	PrintString()

	// Dump produces a dump of the current cache.
	Dump() *Dump
}

// Dump is a dump of the cache state.
type Dump struct {
	AssumedStacks  map[string]bool
	SiteCacheInfos map[string]*schedulersitecacheinfo.SiteCacheInfo
}
