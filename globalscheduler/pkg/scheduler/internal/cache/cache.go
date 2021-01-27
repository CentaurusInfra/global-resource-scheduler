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
	"sync"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/scheduler/listers"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/sets"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/wait"
)

var (
	cleanAssumedPeriod = 1 * time.Second
	FlavorCache = &flavorCache{}
)

// New returns a Cache implementation.
// It automatically starts a go routine that manages expiration of assumed pods.
// "ttl" is how long the assumed pod will get expired.
// "stop" is the channel that would close the background goroutine.
func New(ttl time.Duration, stop <-chan struct{}) Cache {
	cache := newSchedulerCache(ttl, cleanAssumedPeriod, stop)
	cache.run()
	return cache
}

// GetFlavor get flavor from the cache
func (fc *flavorCache) GetFlavor(flavorID string, region string) (*typed.RegionFlavor, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	if region == "" {
		value := fc.FlavorMap[flavorID]
		if value == nil {
			return value, false
		}
		return value, true
	}

	// region != ""
	value := fc.RegionFlavorMap[region + "|" + flavorID]
	if value == nil {
		return value, false
	}
	return value, true
}

// UpdateFlavor update the flavor to the cache
func (fc *flavorCache) UpdateFlavorMap(regionFlavorMap map[string]*typed.RegionFlavor,
	flavorMap map[string]*typed.RegionFlavor) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.FlavorMap = flavorMap
	fc.RegionFlavorMap = regionFlavorMap
}


type flavorCache struct {
	// This mutex guards all fields within this cache struct.
	mu sync.RWMutex
	// RegionFlavorMap is a map of the region flavor id to a flavor, contains all flavors
	RegionFlavorMap map[string]*typed.RegionFlavor
	// FlavorMap is a map of the flavor id to a flavor, contains all flavors
	FlavorMap map[string]*typed.RegionFlavor
}

// siteCacheInfoListItem holds a Host pointer and acts as an item in a doubly
// linked list. When a Host is updated, it goes to the head of the list.
// The items closer to the head are the most recently updated items.
type siteCacheInfoListItem struct {
	info *schedulersitecacheinfo.SiteCacheInfo
	next *siteCacheInfoListItem
	prev *siteCacheInfoListItem
}

type schedulerCache struct {
	stop   <-chan struct{}
	ttl    time.Duration
	period time.Duration

	// This mutex guards all fields within this cache struct.
	mu sync.RWMutex
	// a set of assumed stack keys.
	// The key could further be used to get an entry in stackStates.
	assumedStacks map[string]bool
	// a map from pod key to stackState.
	stackStates    map[string]*StackState
	siteCacheInfos map[string]*siteCacheInfoListItem
	// headSiteCacheInfo points to the most recently updated Host in "siteIDs". It is the
	// head of the linked list.
	headSiteCacheInfo *siteCacheInfoListItem

	regionToSite map[string]sets.String

	siteTree *siteTree
}

type StackState struct {
	stack *types.Stack
	// Used by assumedStack to determinate expiration.
	deadline *time.Time
	// Used to block cache from expiring assumedPod if binding still runs
	bindingFinished bool
}

func newSchedulerCache(ttl, period time.Duration, stop <-chan struct{}) *schedulerCache {
	return &schedulerCache{
		ttl:    ttl,
		period: period,
		stop:   stop,

		siteCacheInfos: make(map[string]*siteCacheInfoListItem),
		siteTree:       newSiteCacheTree(nil),
		assumedStacks:  make(map[string]bool),
		stackStates:    make(map[string]*StackState),
		regionToSite:   make(map[string]sets.String),
	}
}

// newSiteCacheInfoListItem initializes a new siteCacheInfoListItem.
func newSiteCacheInfoListItem(ni *schedulersitecacheinfo.SiteCacheInfo) *siteCacheInfoListItem {
	return &siteCacheInfoListItem{
		info: ni,
	}
}

// moveSiteCacheInfoToHead moves a Host to the head of "cache.siteIDs" doubly
// linked list. The head is the most recently updated Host.
// We assume cache lock is already acquired.
func (cache *schedulerCache) moveSiteCacheInfoToHead(siteID string) {
	ni, ok := cache.siteCacheInfos[siteID]
	if !ok {
		logger.Errorf("No Host with name %v found in the cache", siteID)
		return
	}
	// if the site info list item is already at the head, we are done.
	if ni == cache.headSiteCacheInfo {
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	if cache.headSiteCacheInfo != nil {
		cache.headSiteCacheInfo.prev = ni
	}
	ni.next = cache.headSiteCacheInfo
	ni.prev = nil
	cache.headSiteCacheInfo = ni
}

// removeSiteCacheInfoFromList removes a Host from the "cache.siteIDs" doubly
// linked list.
// We assume cache lock is already acquired.
func (cache *schedulerCache) removeSiteCacheInfoFromList(siteID string) {
	ni, ok := cache.siteCacheInfos[siteID]
	if !ok {
		logger.Errorf("No site with ID %v found in the cache", siteID)
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	// if the removed item was at the head, we must update the head.
	if ni == cache.headSiteCacheInfo {
		cache.headSiteCacheInfo = ni.next
	}
	delete(cache.siteCacheInfos, siteID)
}

// Snapshot takes a snapshot of the current scheduler cache. This is used for
// debugging purposes only and shouldn't be confused with UpdateSnapshot
// function.
// This method is expensive, and should be only used in non-critical path.
func (cache *schedulerCache) Dump() *Dump {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	siteCacheInfos := make(map[string]*schedulersitecacheinfo.SiteCacheInfo, len(cache.siteCacheInfos))
	for k, v := range cache.siteCacheInfos {
		siteCacheInfos[k] = v.info.Clone()
	}

	assumedStacks := make(map[string]bool, len(cache.assumedStacks))
	for k, v := range cache.assumedStacks {
		assumedStacks[k] = v
	}

	return &Dump{
		SiteCacheInfos: siteCacheInfos,
		AssumedStacks:  assumedStacks,
	}
}

// UpdateSnapshot takes a snapshot of cached Host map. This is called at
// beginning of every scheduling cycle.
// This function tracks generation number of Host and updates only the
// entries of an existing snapshot that have changed after the snapshot was taken.
func (cache *schedulerCache) UpdateSnapshot(siteCacheInfoSnapshot *Snapshot) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Get the last generation of the snapshot.
	snapshotGeneration := siteCacheInfoSnapshot.generation

	// SiteInfoList and HavePodsWithAffinitySiteInfoList must be re-created if a siteCacheInfo was added
	// or removed from the cache.
	updateAllLists := false

	// Start from the head of the Host doubly linked list and update snapshot
	// of SiteCacheInfos updated after the last snapshot.
	for siteCacheInfo := cache.headSiteCacheInfo; siteCacheInfo != nil; siteCacheInfo = siteCacheInfo.next {
		if siteCacheInfo.info.GetGeneration() <= snapshotGeneration {
			// all the siteIDs are updated before the existing snapshot. We are done.
			break
		}

		if np := siteCacheInfo.info.GetSite(); np != nil {
			existing, ok := siteCacheInfoSnapshot.SiteCacheInfoMap[np.SiteID]
			if !ok {
				updateAllLists = true
				existing = &schedulersitecacheinfo.SiteCacheInfo{}
				siteCacheInfoSnapshot.SiteCacheInfoMap[np.SiteID] = existing
			}
			clone := siteCacheInfo.info.Clone()
			// We need to preserve the original pointer of the Host struct since it
			// is used in the SiteCacheInfoList, which we may not update.
			*existing = *clone
		}
	}
	// Update the snapshot generation with the latest Host generation.
	if cache.headSiteCacheInfo != nil {
		siteCacheInfoSnapshot.generation = cache.headSiteCacheInfo.info.GetGeneration()
	}

	if len(siteCacheInfoSnapshot.SiteCacheInfoMap) > len(cache.siteCacheInfos) {
		cache.removeDeletedSiteCacheInfosFromSnapshot(siteCacheInfoSnapshot)
		updateAllLists = true
	}

	if updateAllLists {
		cache.updateSiteCacheInfoSnapshotList(siteCacheInfoSnapshot, updateAllLists)
	}

	if len(siteCacheInfoSnapshot.SiteCacheInfoList) != cache.siteTree.numSites {
		errMsg := fmt.Sprintf("snapshot state is not consistent"+
			", length of SiteCacheInfoList=%v not equal to length of siteIDs in tree=%v "+
			", length of SiteCacheInfoMap=%v, length of siteIDs in cache=%v"+
			", trying to recover",
			len(siteCacheInfoSnapshot.SiteCacheInfoList), cache.siteTree.numSites,
			len(siteCacheInfoSnapshot.SiteCacheInfoMap), len(cache.siteCacheInfos))
		logger.Errorf(errMsg)
		// We will try to recover by re-creating the lists for the next scheduling cycle, but still return an
		// error to surface the problem, the error will likely cause a failure to the current scheduling cycle.
		cache.updateSiteCacheInfoSnapshotList(siteCacheInfoSnapshot, true)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (cache *schedulerCache) updateSiteCacheInfoSnapshotList(snapshot *Snapshot, updateAll bool) {
	if updateAll {
		// Take a snapshot of the siteIDs order in the tree
		snapshot.SiteCacheInfoList = make([]*schedulersitecacheinfo.SiteCacheInfo, 0, cache.siteTree.numSites)
		for i := 0; i < cache.siteTree.numSites; i++ {
			siteID := cache.siteTree.next()
			if n := snapshot.SiteCacheInfoMap[siteID]; n != nil {
				snapshot.SiteCacheInfoList = append(snapshot.SiteCacheInfoList, n)
			} else {
				logger.Errorf("site %q exist in siteTree but not in siteInfoMap, this should not happen.",
					siteID)
			}
		}
	}
}

// If certain siteCacheInfos were deleted after the last snapshot was taken, we should remove them from the snapshot.
func (cache *schedulerCache) removeDeletedSiteCacheInfosFromSnapshot(snapshot *Snapshot) {
	toDelete := len(snapshot.SiteCacheInfoMap) - len(cache.siteCacheInfos)
	for name := range snapshot.SiteCacheInfoMap {
		if toDelete <= 0 {
			break
		}
		if _, ok := cache.siteCacheInfos[name]; !ok {
			delete(snapshot.SiteCacheInfoMap, name)
			toDelete--
		}
	}
}

// ForgetStack removes an assumed stack from cache.
func (cache *schedulerCache) ForgetStack(stack *types.Stack) error {
	key, err := schedulersitecacheinfo.GetStackKey(stack)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	_, ok := cache.stackStates[key]
	switch {
	// Only assumed stack can be forgotten.
	case ok && cache.assumedStacks[key]:
		err := cache.removeStack(stack)
		if err != nil {
			return err
		}
		delete(cache.assumedStacks, key)
		delete(cache.stackStates, key)
	default:
		return fmt.Errorf("stack %v wasn't assumed so cannot be forgotten", key)
	}
	return nil
}

//AssumeStack assume stack to site
func (cache *schedulerCache) AssumeStack(stack *types.Stack) error {
	key, err := schedulersitecacheinfo.GetStackKey(stack)
	if err != nil {
		return err
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if _, ok := cache.stackStates[key]; ok {
		return fmt.Errorf("stack %v is in the cache, so can't be assumed", key)
	}

	cache.addStack(stack)
	ps := &StackState{
		stack: stack,
	}
	cache.stackStates[key] = ps
	cache.assumedStacks[key] = true
	return nil
}

// Assumes that lock is already acquired.
func (cache *schedulerCache) addStack(stack *types.Stack) {
	n, ok := cache.siteCacheInfos[stack.Selected.SiteID]
	if !ok {
		n = newSiteCacheInfoListItem(schedulersitecacheinfo.NewSiteCacheInfo())
		cache.siteCacheInfos[stack.Selected.SiteID] = n
	}
	n.info.AddStack(stack)
	cache.moveSiteCacheInfoToHead(stack.Selected.SiteID)
}

// Assumes that lock is already acquired.
func (cache *schedulerCache) updateStack(oldStack, newStack *types.Stack) error {
	if err := cache.removeStack(oldStack); err != nil {
		return err
	}
	cache.addStack(newStack)
	return nil
}

// Assumes that lock is already acquired.
// Removes a stack from the cached site info. When a site is removed, some pod
// deletion events might arrive later. This is not a problem, as the pods in
// the site are assumed to be removed already.
func (cache *schedulerCache) removeStack(stack *types.Stack) error {
	n, ok := cache.siteCacheInfos[stack.Selected.SiteID]
	if !ok {
		return nil
	}
	if err := n.info.RemoveStack(stack); err != nil {
		return err
	}
	cache.moveSiteCacheInfoToHead(stack.Selected.SiteID)
	return nil
}

//AddStack add stack
func (cache *schedulerCache) AddStack(stack *types.Stack) error {

	key, err := schedulersitecacheinfo.GetStackKey(stack)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.stackStates[key]
	switch {
	case ok && cache.assumedStacks[key]:
		if currState.stack.Selected != stack.Selected {
			// The stack was added to a different site than it was assumed to.
			logger.Warnf("Stack %v was assumed to be on %v but got added to %v", key, stack.Selected, currState.stack.Selected)
			// Clean this up.
			if err := cache.removeStack(currState.stack); err != nil {
				logger.Errorf("removing pod error: %v", err)
			}
			cache.AddStack(stack)
		}
		delete(cache.assumedStacks, key)
		cache.stackStates[key].deadline = nil
		cache.stackStates[key].stack = stack
	case !ok:
		// stack was expired. We should add it back.
		cache.addStack(stack)
		ps := &StackState{
			stack: stack,
		}
		cache.stackStates[key] = ps
	default:
		return fmt.Errorf("stack %v was already in added state", key)
	}
	return nil
}

//UpdateStack update stack
func (cache *schedulerCache) UpdateStack(oldStack, newStack *types.Stack) error {
	key, err := schedulersitecacheinfo.GetStackKey(oldStack)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.stackStates[key]
	switch {
	// An assumed pod won't have Update/Remove event. It needs to have Add event
	// before Update event, in which case the state would change from Assumed to Added.
	case ok && !cache.assumedStacks[key]:
		if currState.stack.Selected != newStack.Selected {
			logger.Errorf("Stack %v updated on a different site than previously added to.", key)
			logger.Errorf("Schedulercache is corrupted and can badly affect scheduling decisions")
		}
		if err := cache.updateStack(oldStack, newStack); err != nil {
			return err
		}
		currState.stack = newStack
	default:
		return fmt.Errorf("pod %v is not added to scheduler cache, so cannot be updated", key)
	}
	return nil
}

//RemoveStack remove stack
func (cache *schedulerCache) RemoveStack(stack *types.Stack) error {
	key, err := schedulersitecacheinfo.GetStackKey(stack)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.stackStates[key]
	switch {
	// An assumed stack won't have Delete/Remove event. It needs to have Add event
	// before Remove event, in which case the state would change from Assumed to Added.
	case ok && !cache.assumedStacks[key]:
		if currState.stack.Selected.SiteID != stack.Selected.SiteID {
			logger.Errorf("Stack %v was assumed to be on %v but got added to %v", key,
				stack.Selected.SiteID, currState.stack.Selected.SiteID)
			logger.Errorf("Schedulercache is corrupted and can badly affect scheduling decisions")
		}
		err := cache.removeStack(currState.stack)
		if err != nil {
			return err
		}
		delete(cache.stackStates, key)
	default:
		return fmt.Errorf("stack %v is not found in scheduler cache, so cannot be removed from it", key)
	}
	return nil
}

//IsAssumedStack is assume stack
func (cache *schedulerCache) IsAssumedStack(stack *types.Stack) (bool, error) {
	key, err := schedulersitecacheinfo.GetStackKey(stack)
	if err != nil {
		return false, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	b, found := cache.assumedStacks[key]
	if !found {
		return false, nil
	}
	return b, nil
}

// GetPod might return a pod for which its site has already been deleted from
// the main cache. This is useful to properly process pod update events.
func (cache *schedulerCache) GetStack(stack *types.Stack) (*types.Stack, error) {
	key, err := schedulersitecacheinfo.GetStackKey(stack)
	if err != nil {
		return nil, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	stackStates, ok := cache.stackStates[key]
	if !ok {
		return nil, fmt.Errorf("stack %v does not exist in scheduler cache", key)
	}

	return stackStates.stack, nil
}

func (cache *schedulerCache) AddSite(site *types.Site) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.siteCacheInfos[site.SiteID]
	if !ok {
		n = newSiteCacheInfoListItem(schedulersitecacheinfo.NewSiteCacheInfo())
		cache.siteCacheInfos[site.SiteID] = n
	}

	cache.moveSiteCacheInfoToHead(site.SiteID)

	cache.siteTree.addSite(site)
	cache.updateRegionToSite(site)
	return n.info.SetSite(site)
}

func (cache *schedulerCache) UpdateSite(oldSite, newSite *types.Site) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.siteCacheInfos[newSite.SiteID]
	if !ok {
		n = newSiteCacheInfoListItem(schedulersitecacheinfo.NewSiteCacheInfo())
		cache.siteCacheInfos[newSite.SiteID] = n
		cache.siteTree.addSite(newSite)
	}
	cache.moveSiteCacheInfoToHead(newSite.SiteID)

	cache.siteTree.updateSite(oldSite, newSite)
	cache.deleteRegionToSite(oldSite)
	cache.updateRegionToSite(newSite)
	return n.info.SetSite(newSite)
}

// RemoveSite removes a site from the cache.
// Some siteIDs might still have pods because their deletion events didn't arrive
// yet. For most intents and purposes, those pods are removed from the cache,
// having it's source of truth in the cached siteIDs.
// However, some information on pods (assumedPods, podStates) persist. These
// caches will be eventually consistent as pod deletion events arrive.
func (cache *schedulerCache) RemoveSite(site *types.Site) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	_, ok := cache.siteCacheInfos[site.SiteID]
	if !ok {
		return fmt.Errorf("site %v is not found", site.SiteID)
	}
	cache.removeSiteCacheInfoFromList(site.SiteID)
	if err := cache.siteTree.removeSite(site); err != nil {
		return err
	}

	cache.deleteRegionToSite(site)
	return nil
}

func (cache *schedulerCache) updateRegionToSite(site *types.Site) {
	_, ok := cache.regionToSite[site.Region]
	if !ok {
		cache.regionToSite[site.Region] = sets.NewString()
	}
	cache.regionToSite[site.Region].Insert(site.SiteID)
}

func (cache *schedulerCache) deleteRegionToSite(site *types.Site) {
	_, ok := cache.regionToSite[site.Region]
	if !ok {
		return
	}
	cache.regionToSite[site.Region].Delete(site.SiteID)
}

//UpdateSiteWithEipPool update eip pool
func (cache *schedulerCache) UpdateSiteWithEipPool(siteID string, eipPool *typed.EipPool) error {
	siteCacheInfo, ok := cache.siteCacheInfos[siteID]
	if !ok {
		return fmt.Errorf("site %v is not found", siteID)
	}

	err := siteCacheInfo.info.UpdateSiteWithEipPool(eipPool)
	if err != nil {
		logger.Errorf("UpdateSiteWithEipPool failed! err: %s", err)
		return err
	}

	cache.moveSiteCacheInfoToHead(siteID)

	return nil
}

//UpdateSiteWithVolumePool update volume pool
func (cache *schedulerCache) UpdateSiteWithVolumePool(siteID string, volumePool *typed.RegionVolumePool) error {
	siteCacheInfo, ok := cache.siteCacheInfos[siteID]
	if !ok {
		return fmt.Errorf("siteCacheInfo %v is not found", siteID)
	}

	err := siteCacheInfo.info.UpdateSiteWithVolumePool(volumePool)
	if err != nil {
		logger.Errorf("UpdateSiteWithEipPool failed! err: %s", err)
		return err
	}

	cache.moveSiteCacheInfoToHead(siteID)

	return nil
}

//UpdateEipPool updates eip pool info about site
func (cache *schedulerCache) UpdateEipPool(eipPool *typed.EipPool) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	siteIDs, ok := cache.regionToSite[eipPool.Region]
	if !ok {
		return nil
	}

	for siteID := range siteIDs {
		err := cache.UpdateSiteWithEipPool(siteID, eipPool)
		if err != nil {
			logger.Errorf("UpdateSiteWithEipPool failed! err: %s", err)
			continue
		}
	}

	return nil
}

//UpdateVolumePool updates volume pool info about site
func (cache *schedulerCache) UpdateVolumePool(volumePool *typed.RegionVolumePool) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	siteIDs, ok := cache.regionToSite[volumePool.Region]
	if !ok {
		return nil
	}

	for siteID := range siteIDs {
		err := cache.UpdateSiteWithVolumePool(siteID, volumePool)
		if err != nil {
			logger.Errorf("UpdateSiteWithEipPool failed! err: %s", err)
			continue
		}
	}

	return nil
}

// UpdateSiteWithResInfo update res info
func (cache *schedulerCache) UpdateSiteWithResInfo(siteID string, resInfo types.AllResInfo) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	siteCacheInfo, ok := cache.siteCacheInfos[siteID]
	if !ok {
		return nil
	}

	err := siteCacheInfo.info.UpdateSiteWithResInfo(resInfo)
	if err != nil {
		logger.Errorf("UpdateSiteWithResInfo failed! err: %s", err)
		return err
	}

	cache.moveSiteCacheInfoToHead(siteID)

	return nil
}

func (cache *schedulerCache) UpdateQos(siteID string, netMetricData *types.NetMetricDatas) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	siteCacheInfo, ok := cache.siteCacheInfos[siteID]
	if !ok {
		return nil
	}

	err := siteCacheInfo.info.UpdateQos(netMetricData)
	if err != nil {
		logger.Errorf("UpdateQos failed! err: %s", err)
		return err
	}

	cache.moveSiteCacheInfoToHead(siteID)

	return nil
}

//UpdateSiteWithVcpuMem update vcpu and mem
func (cache *schedulerCache) UpdateSiteWithRatio(region string, az string, ratios []types.AllocationRatio) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, siteCacheInfo := range cache.siteCacheInfos {
		if siteCacheInfo.info.GetSite().Region == region && siteCacheInfo.info.GetSite().AvailabilityZone == az {
			err := siteCacheInfo.info.UpdateSiteWithRatio(ratios)
			if err != nil {
				logger.Errorf("UpdateSiteWithRatio failed! err: %s", err)
				return err
			}

			cache.moveSiteCacheInfoToHead(siteCacheInfo.info.GetSite().SiteID)
			break
		}
	}

	return nil
}

//UpdateSpotResources update spot resources
func (cache *schedulerCache) UpdateSpotResources(region string, az string, spotRes map[string]types.SpotResource) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, siteCacheInfo := range cache.siteCacheInfos {
		if siteCacheInfo.info.GetSite().Region == region && siteCacheInfo.info.GetSite().AvailabilityZone == az {
			err := siteCacheInfo.info.UpdateSpotResources(spotRes)
			if err != nil {
				logger.Errorf("UpdateSiteWithRatio failed! err: %s", err)
				return err
			}
			cache.moveSiteCacheInfoToHead(siteCacheInfo.info.GetSite().SiteID)

			break
		}
	}

	return nil
}

//GetRegions get cache region info
func (cache *schedulerCache) GetRegions() map[string]types.CloudRegion {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	ret := map[string]types.CloudRegion{}
	for _, siteInfoCache := range cache.siteCacheInfos {
		region := siteInfoCache.info.GetSite().Region
		cr, ok := ret[region]
		if !ok {
			cr = types.CloudRegion{Region: region, AvailabilityZone: []string{}}
		}
		cr.AvailabilityZone = append(cr.AvailabilityZone, siteInfoCache.info.GetSite().AvailabilityZone)
		ret[region] = cr
	}

	return ret
}

//PrintString print site cache info
func (cache *schedulerCache) PrintString() {
	for _, siteCacheInfo := range cache.siteCacheInfos {
		logger.Infof("siteID: %s, info: %s", siteCacheInfo.info.GetSite().SiteID, siteCacheInfo.info.ToString())
	}
}

func (cache *schedulerCache) run() {
	go wait.Until(cache.cleanupExpiredAssumedStacks, cache.period, cache.stop)
}

func (cache *schedulerCache) cleanupExpiredAssumedStacks() {
	cache.cleanupAssumedStacks(time.Now())
}

// cleanupAssumedStacks exists for making test deterministic by taking time as input argument.
// It also reports metrics on the cache size for siteCacheInfos, stacks, and assumed stacks.
func (cache *schedulerCache) cleanupAssumedStacks(now time.Time) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// The size of assumedStacks should be small
	for key := range cache.assumedStacks {
		ps, ok := cache.stackStates[key]
		if !ok {
			logger.Fatalf("Key found in assumed set but not in podStates. Potentially a logical error.")
		}
		if !ps.bindingFinished {
			logger.Infof("Couldn't expire cache for stack %v/%v. Binding is still in progress.",
				ps.stack.UID, ps.stack.Name)
			continue
		}
		if now.After(*ps.deadline) {
			logger.Warnf("Stack %s/%s expired", ps.stack.UID, ps.stack.Name)
			if err := cache.expireStack(key, ps); err != nil {
				logger.Errorf("ExpirePod failed for %s: %v", key, err)
			}
		}
	}
}

func (cache *schedulerCache) expireStack(key string, ps *StackState) error {
	if err := cache.removeStack(ps.stack); err != nil {
		return err
	}
	delete(cache.assumedStacks, key)
	delete(cache.stackStates, key)
	return nil
}

func (cache *schedulerCache) List() ([]*types.Stack, error) {
	alwaysTrue := func(p *types.Stack) bool { return true }
	return cache.FilteredList(alwaysTrue)
}

func (cache *schedulerCache) FilteredList(stackFilter schedulerlisters.StackFilter) ([]*types.Stack, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	// stackFilter is expected to return true for most or all of the pods. We
	// can avoid expensive array growth without wasting too much memory by
	// pre-allocating capacity.
	maxSize := 0
	for _, n := range cache.siteCacheInfos {
		maxSize += len(n.info.Stacks())
	}
	stacks := make([]*types.Stack, 0, maxSize)
	for _, n := range cache.siteCacheInfos {
		for _, stack := range n.info.Stacks() {
			if stackFilter(stack) {
				stacks = append(stacks, stack)
			}
		}
	}
	return stacks, nil
}
