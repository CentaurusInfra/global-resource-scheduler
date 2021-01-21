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
	"fmt"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"

	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/scheduler/listers"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Snapshot is a snapshot of cache SiteCacheInfo and SiteTree order. The scheduler takes a
// snapshot at the beginning of each scheduling cycle and uses it for its operations in that cycle.
type Snapshot struct {
	// SiteCacheInfoMap a map of site name to a snapshot of its SiteCacheInfo.
	SiteCacheInfoMap map[string]*schedulersitecacheinfo.SiteCacheInfo				`json:"siteCacheInfoMap"`
	// SiteCacheInfoList is the list of siteIDs as ordered in the cache's siteTree.
	SiteCacheInfoList []*schedulersitecacheinfo.SiteCacheInfo						`json:"siteCacheInfoList"`
	// HavePodsWithAffinitySiteCacheInfoList is the list of siteIDs with at least one pod declaring affinity terms.
	HavePodsWithAffinitySiteCacheInfoList []*schedulersitecacheinfo.SiteCacheInfo	`json:"havePodsWithAffinitySiteCacheInfoList"`
	// RegionFlavorMap is a map of the region flavor id to a flavor, contains all flavors
	RegionFlavorMap map[string]*typed.RegionFlavor									`json:"regionFlavorMap"`
	// FlavorMap is a map of the flavor id to a flavor, contains all flavors
	FlavorMap map[string]*typed.RegionFlavor										`json:"flavorMap"`
	generation                            int64
}

var _ schedulerlisters.SharedLister = &Snapshot{}

// NewEmptySnapshot initializes a Snapshot struct and returns it.
func NewEmptySnapshot() *Snapshot {
	return &Snapshot{
		SiteCacheInfoMap: make(map[string]*schedulersitecacheinfo.SiteCacheInfo),
	}
}

// NewSnapshot initializes a Snapshot struct and returns it.
func NewSnapshot(stacks []*types.Stack, sites []*types.Site) *Snapshot {
	siteCacheInfoMap := createSiteInfoCacheMap(stacks, sites)
	siteCacheInfoList := make([]*schedulersitecacheinfo.SiteCacheInfo, 0, len(siteCacheInfoMap))
	havePodsWithAffinitySiteCacheInfoList := make([]*schedulersitecacheinfo.SiteCacheInfo, 0, len(siteCacheInfoMap))
	for _, v := range siteCacheInfoMap {
		siteCacheInfoList = append(siteCacheInfoList, v)
		if len(v.StackWithAffinity()) > 0 {
			havePodsWithAffinitySiteCacheInfoList = append(havePodsWithAffinitySiteCacheInfoList, v)
		}
	}

	s := NewEmptySnapshot()
	s.SiteCacheInfoMap = siteCacheInfoMap
	s.SiteCacheInfoList = siteCacheInfoList
	s.HavePodsWithAffinitySiteCacheInfoList = havePodsWithAffinitySiteCacheInfoList

	return s
}

// createSiteInfoCacheMap obtains a list of pods and pivots that list into a map
// where the keys are site names and the values are the aggregated information
// for that site.
func createSiteInfoCacheMap(stacks []*types.Stack, sites []*types.Site) map[string]*schedulersitecacheinfo.SiteCacheInfo {
	siteIDToInfo := make(map[string]*schedulersitecacheinfo.SiteCacheInfo)
	for _, stack := range stacks {
		siteID := stack.Selected.SiteID
		if _, ok := siteIDToInfo[siteID]; !ok {
			siteIDToInfo[siteID] = schedulersitecacheinfo.NewSiteCacheInfo()
		}
		siteIDToInfo[siteID].AddStack(stack)
	}

	for _, site := range sites {
		if _, ok := siteIDToInfo[site.SiteID]; !ok {
			siteIDToInfo[site.SiteID] = schedulersitecacheinfo.NewSiteCacheInfo()
		}
		siteCacheInfo := siteIDToInfo[site.SiteID]
		siteCacheInfo.SetSite(site)
	}
	return siteIDToInfo
}

// Stacks returns a StackLister
func (s *Snapshot) Stacks() schedulerlisters.StackLister {
	return stackLister(s.SiteCacheInfoList)
}

// SiteCacheInfos returns a SiteCacheInfoLister.
func (s *Snapshot) SiteCacheInfos() schedulerlisters.SiteCacheInfoLister {
	return s
}

// NumSites returns the number of siteIDs in the snapshot.
func (s *Snapshot) NumSites() int {
	return len(s.SiteCacheInfoList)
}

type stackLister []*schedulersitecacheinfo.SiteCacheInfo

// List returns the list of stacks in the snapshot.
func (p stackLister) List() ([]*types.Stack, error) {
	alwaysTrue := func(*types.Stack) bool { return true }
	return p.FilteredList(alwaysTrue)
}

// FilteredList returns a filtered list of stacks in the snapshot.
func (p stackLister) FilteredList(filter schedulerlisters.StackFilter) ([]*types.Stack, error) {
	// stackFilter is expected to return true for most or all of the stacks. We
	// can avoid expensive array growth without wasting too much memory by
	// pre-allocating capacity.
	maxSize := 0
	for _, n := range p {
		maxSize += len(n.Stacks())
	}
	stacks := make([]*types.Stack, 0, maxSize)
	for _, n := range p {
		for _, stack := range n.Stacks() {
			if filter(stack) {
				stacks = append(stacks, stack)
			}
		}
	}
	return stacks, nil
}

// List returns the list of siteIDs in the snapshot.
func (s *Snapshot) List() ([]*schedulersitecacheinfo.SiteCacheInfo, error) {
	return s.SiteCacheInfoList, nil
}

// HavePodsWithAffinityList returns the list of siteIDs with at least one pods with inter-pod affinity
func (s *Snapshot) HavePodsWithAffinityList() ([]*schedulersitecacheinfo.SiteCacheInfo, error) {
	return s.HavePodsWithAffinitySiteCacheInfoList, nil
}

// Get returns the SiteCacheInfo of the given site ID.
func (s *Snapshot) Get(siteID string) (*schedulersitecacheinfo.SiteCacheInfo, error) {
	if v, ok := s.SiteCacheInfoMap[siteID]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("sitecacheinfo not found for site ID %q", siteID)
}
