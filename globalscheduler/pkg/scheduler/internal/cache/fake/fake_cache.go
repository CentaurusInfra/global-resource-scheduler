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

package fake

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	internalcache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/scheduler/listers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

var _ internalcache.Cache = &Cache{}

// Cache is used for testing
type Cache struct {
	AssumeFunc         func(*types.Stack)
	ForgetFunc         func(*types.Stack)
	IsAssumedStackFunc func(*types.Stack) bool
	GetStackFunc       func(*types.Stack) *types.Stack
}

func (c *Cache) List() ([]*types.Stack, error) {
	return nil, nil
}

func (c *Cache) FilteredList(stackFilter schedulerlisters.StackFilter) ([]*types.Stack, error) {
	return nil, nil
}

func (c *Cache) AddSite(site *types.Site) error {
	return nil
}

func (c *Cache) UpdateSite(oldSite, newSite *types.Site) error {
	return nil
}

func (c *Cache) RemoveSite(site *types.Site) error {
	return nil
}

func (c *Cache) UpdateEipPool(eipPool *typed.EipPool) error {
	return nil
}

func (c *Cache) UpdateVolumePool(volumePool *typed.RegionVolumePool) error {
	return nil
}

func (c *Cache) UpdateSiteWithResInfo(siteID string, resInfo types.AllResInfo) error {
	return nil
}

func (c *Cache) UpdateQos(siteID string, netMetricData *types.NetMetricDatas) error {
	return nil
}

func (c *Cache) UpdateSiteWithRatio(region string, az string, ratios []types.AllocationRatio) error {
	return nil
}

func (c *Cache) UpdateSpotResources(region string, az string, spotRes map[string]types.SpotResource) error {
	return nil
}

func (c *Cache) UpdateSnapshot(siteCacheInfoSnapshot *internalcache.Snapshot) error {
	return nil
}

func (c *Cache) GetRegions() map[string]types.CloudRegion {
	return nil
}

func (c *Cache) PrintString() {
	return
}

func (c *Cache) Dump() *internalcache.Dump {
	return nil
}

// AssumeStack is a fake method for testing.
func (c *Cache) AssumeStack(pod *types.Stack) error {
	c.AssumeFunc(pod)
	return nil
}

// FinishBinding is a fake method for testing.
func (c *Cache) FinishBinding(pod *types.Stack) error { return nil }

// ForgetStack is a fake method for testing.
func (c *Cache) ForgetStack(pod *types.Stack) error {
	c.ForgetFunc(pod)
	return nil
}

// AddStack is a fake method for testing.
func (c *Cache) AddStack(pod *types.Stack) error { return nil }

// UpdateStack is a fake method for testing.
func (c *Cache) UpdateStack(oldStack, newStack *types.Stack) error { return nil }

// RemoveStack is a fake method for testing.
func (c *Cache) RemoveStack(pod *types.Stack) error { return nil }

// IsAssumedStack is a fake method for testing.
func (c *Cache) IsAssumedStack(stack *types.Stack) (bool, error) {
	return c.IsAssumedStackFunc(stack), nil
}

// GetStack is a fake method for testing.
func (c *Cache) GetStack(stack *types.Stack) (*types.Stack, error) {
	return c.GetStackFunc(stack), nil
}

// Snapshot is a fake method for testing
func (c *Cache) Snapshot() *internalcache.Snapshot {
	return &internalcache.Snapshot{}
}
