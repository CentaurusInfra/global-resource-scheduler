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

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

// siteTree is a tree-like data structure that holds site names in each zone. Zone names are
// keys to "SiteTree.tree" and values of "SiteTree.tree" are arrays of site names.
// SiteTree is NOT thread-safe, any concurrent updates/reads from it must be synchronized by the caller.
// It is used only by schedulerCache, and should stay as such.
type siteTree struct {
	tree      map[string]*siteArray // a map from zone (region-zone) to an array of siteIDs in the zone.
	zones     []string              // a list of all the zones in the tree (keys)
	zoneIndex int
	numSites  int
}

// siteArray is a struct that has site that are in a zone.
// We use a slice (as opposed to a set/map) to store the site cache because iterating over the site cache is
// a lot more frequent than searching them by name.
type siteArray struct {
	siteIDs   []string
	lastIndex int
}

func (na *siteArray) next() (siteID string, exhausted bool) {
	if len(na.siteIDs) == 0 {
		logger.Errorf("The siteArray is empty. It should have been deleted from SiteTree.")
		return "", false
	}
	if na.lastIndex >= len(na.siteIDs) {
		return "", true
	}
	siteID = na.siteIDs[na.lastIndex]
	na.lastIndex++
	return siteID, false
}

// newSiteCacheTree creates a SiteCacheTree from siteIDs.
func newSiteCacheTree(sites []*types.Site) *siteTree {
	nt := &siteTree{
		tree: make(map[string]*siteArray),
	}
	for _, n := range sites {
		nt.addSite(n)
	}
	return nt
}

// addSite adds a site and its corresponding zone to the tree. If the zone already exists, the site
// is added to the array of siteIDs in that zone.
func (nt *siteTree) addSite(site *types.Site) {
	zone := utils.GetZoneKey(site)
	if na, ok := nt.tree[zone]; ok {
		for _, siteID := range na.siteIDs {
			if siteID == site.SiteID {
				return
			}
		}
		na.siteIDs = append(na.siteIDs, site.SiteID)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = &siteArray{siteIDs: []string{site.SiteID}, lastIndex: 0}
	}
	logger.Infof("Added site %q in group %q to SiteTree", site.SiteID, zone)
	nt.numSites++
}

// removeSite removes a site from the SiteTree.
func (nt *siteTree) removeSite(n *types.Site) error {
	zone := utils.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for i, siteID := range na.siteIDs {
			if siteID == n.SiteID {
				na.siteIDs = append(na.siteIDs[:i], na.siteIDs[i+1:]...)
				if len(na.siteIDs) == 0 {
					nt.removeZone(zone)
				}
				logger.Infof("Removed site %q in group %q from SiteTree", n.SiteID, zone)
				nt.numSites--
				return nil
			}
		}
	}
	logger.Errorf("Site %q in group %q was not found", n.SiteID, zone)
	return fmt.Errorf("site %q in group %q was not found", n.SiteID, zone)
}

// removeZone removes a zone from tree.
// This function must be called while writer locks are hold.
func (nt *siteTree) removeZone(zone string) {
	delete(nt.tree, zone)
	for i, z := range nt.zones {
		if z == zone {
			nt.zones = append(nt.zones[:i], nt.zones[i+1:]...)
			return
		}
	}
}

// updateSite updates a site in the SiteTree.
func (nt *siteTree) updateSite(old, new *types.Site) {
	var oldZone string
	if old != nil {
		oldZone = utils.GetZoneKey(old)
	}
	newZone := utils.GetZoneKey(new)
	// If the zone ID of the site has not changed, we don't need to do anything. Name of the site
	// cannot be changed in an update.
	if oldZone == newZone {
		return
	}
	err := nt.removeSite(old) // No error checking. We ignore whether the old site exists or not.
	if err != nil {
		logger.Errorf("removeSite failed! err: %s", err)
	}
	nt.addSite(new)
}

func (nt *siteTree) resetExhausted() {
	for _, na := range nt.tree {
		na.lastIndex = 0
	}
	nt.zoneIndex = 0
}

// next returns the name of the next site. SiteTree iterates over zones and in each zone iterates
// over siteIDs in a round robin fashion.
func (nt *siteTree) next() string {
	if len(nt.zones) == 0 {
		return ""
	}
	numExhaustedZones := 0
	for {
		if nt.zoneIndex >= len(nt.zones) {
			nt.zoneIndex = 0
		}
		zone := nt.zones[nt.zoneIndex]
		nt.zoneIndex++
		// We do not check the exhausted zones before calling next() on the zone. This ensures
		// that if more siteIDs are added to a zone after it is exhausted, we iterate over the new siteIDs.
		siteID, exhausted := nt.tree[zone].next()
		if exhausted {
			numExhaustedZones++
			if numExhaustedZones >= len(nt.zones) { // all zones are exhausted. we should reset.
				nt.resetExhausted()
			}
		} else {
			return siteID
		}
	}
}
