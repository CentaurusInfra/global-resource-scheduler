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
	schedulernodeinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/sets"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/wait"
)

var (
	cleanAssumedPeriod = 1 * time.Second
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

// nodeInfoListItem holds a NodeInfo pointer and acts as an item in a doubly
// linked list. When a NodeInfo is updated, it goes to the head of the list.
// The items closer to the head are the most recently updated items.
type nodeInfoListItem struct {
	info *schedulernodeinfo.NodeInfo
	next *nodeInfoListItem
	prev *nodeInfoListItem
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
	stackStates map[string]*StackState
	nodes       map[string]*nodeInfoListItem
	// headNode points to the most recently updated NodeInfo in "nodes". It is the
	// head of the linked list.
	headNode *nodeInfoListItem

	regionToNode map[string]sets.String

	nodeTree *nodeTree
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

		nodes:         make(map[string]*nodeInfoListItem),
		nodeTree:      newNodeTree(nil),
		assumedStacks: make(map[string]bool),
		stackStates:   make(map[string]*StackState),
		regionToNode:  make(map[string]sets.String),
	}
}

// newNodeInfoListItem initializes a new nodeInfoListItem.
func newNodeInfoListItem(ni *schedulernodeinfo.NodeInfo) *nodeInfoListItem {
	return &nodeInfoListItem{
		info: ni,
	}
}

// moveNodeInfoToHead moves a NodeInfo to the head of "cache.nodes" doubly
// linked list. The head is the most recently updated NodeInfo.
// We assume cache lock is already acquired.
func (cache *schedulerCache) moveNodeInfoToHead(nodeID string) {
	ni, ok := cache.nodes[nodeID]
	if !ok {
		logger.Errorf("No NodeInfo with name %v found in the cache", nodeID)
		return
	}
	// if the node info list item is already at the head, we are done.
	if ni == cache.headNode {
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	if cache.headNode != nil {
		cache.headNode.prev = ni
	}
	ni.next = cache.headNode
	ni.prev = nil
	cache.headNode = ni
}

// removeNodeInfoFromList removes a NodeInfo from the "cache.nodes" doubly
// linked list.
// We assume cache lock is already acquired.
func (cache *schedulerCache) removeNodeInfoFromList(nodeID string) {
	ni, ok := cache.nodes[nodeID]
	if !ok {
		logger.Errorf("No NodeInfo with name %v found in the cache", nodeID)
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	// if the removed item was at the head, we must update the head.
	if ni == cache.headNode {
		cache.headNode = ni.next
	}
	delete(cache.nodes, nodeID)
}

// Snapshot takes a snapshot of the current scheduler cache. This is used for
// debugging purposes only and shouldn't be confused with UpdateSnapshot
// function.
// This method is expensive, and should be only used in non-critical path.
func (cache *schedulerCache) Dump() *Dump {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	nodes := make(map[string]*schedulernodeinfo.NodeInfo, len(cache.nodes))
	for k, v := range cache.nodes {
		nodes[k] = v.info.Clone()
	}

	assumedStacks := make(map[string]bool, len(cache.assumedStacks))
	for k, v := range cache.assumedStacks {
		assumedStacks[k] = v
	}

	return &Dump{
		Nodes:         nodes,
		AssumedStacks: assumedStacks,
	}
}

// UpdateSnapshot takes a snapshot of cached NodeInfo map. This is called at
// beginning of every scheduling cycle.
// This function tracks generation number of NodeInfo and updates only the
// entries of an existing snapshot that have changed after the snapshot was taken.
func (cache *schedulerCache) UpdateSnapshot(nodeSnapshot *Snapshot) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Get the last generation of the snapshot.
	snapshotGeneration := nodeSnapshot.generation

	// NodeInfoList and HavePodsWithAffinityNodeInfoList must be re-created if a node was added
	// or removed from the cache.
	updateAllLists := false

	// Start from the head of the NodeInfo doubly linked list and update snapshot
	// of NodeInfos updated after the last snapshot.
	for node := cache.headNode; node != nil; node = node.next {
		if node.info.GetGeneration() <= snapshotGeneration {
			// all the nodes are updated before the existing snapshot. We are done.
			break
		}

		if np := node.info.Node(); np != nil {
			existing, ok := nodeSnapshot.nodeInfoMap[np.SiteID]
			if !ok {
				updateAllLists = true
				existing = &schedulernodeinfo.NodeInfo{}
				nodeSnapshot.nodeInfoMap[np.SiteID] = existing
			}
			clone := node.info.Clone()
			// We need to preserve the original pointer of the NodeInfo struct since it
			// is used in the NodeInfoList, which we may not update.
			*existing = *clone
		}
	}
	// Update the snapshot generation with the latest NodeInfo generation.
	if cache.headNode != nil {
		nodeSnapshot.generation = cache.headNode.info.GetGeneration()
	}

	if len(nodeSnapshot.nodeInfoMap) > len(cache.nodes) {
		cache.removeDeletedNodesFromSnapshot(nodeSnapshot)
		updateAllLists = true
	}

	if updateAllLists {
		cache.updateNodeInfoSnapshotList(nodeSnapshot, updateAllLists)
	}

	if len(nodeSnapshot.nodeInfoList) != cache.nodeTree.numNodes {
		errMsg := fmt.Sprintf("snapshot state is not consistent"+
			", length of NodeInfoList=%v not equal to length of nodes in tree=%v "+
			", length of NodeInfoMap=%v, length of nodes in cache=%v"+
			", trying to recover",
			len(nodeSnapshot.nodeInfoList), cache.nodeTree.numNodes,
			len(nodeSnapshot.nodeInfoMap), len(cache.nodes))
		logger.Errorf(errMsg)
		// We will try to recover by re-creating the lists for the next scheduling cycle, but still return an
		// error to surface the problem, the error will likely cause a failure to the current scheduling cycle.
		cache.updateNodeInfoSnapshotList(nodeSnapshot, true)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (cache *schedulerCache) updateNodeInfoSnapshotList(snapshot *Snapshot, updateAll bool) {
	if updateAll {
		// Take a snapshot of the nodes order in the tree
		snapshot.nodeInfoList = make([]*schedulernodeinfo.NodeInfo, 0, cache.nodeTree.numNodes)
		for i := 0; i < cache.nodeTree.numNodes; i++ {
			nodeName := cache.nodeTree.next()
			if n := snapshot.nodeInfoMap[nodeName]; n != nil {
				snapshot.nodeInfoList = append(snapshot.nodeInfoList, n)
			} else {
				logger.Errorf("node %q exist in nodeTree but not in NodeInfoMap, this should not happen.",
					nodeName)
			}
		}
	}
}

// If certain nodes were deleted after the last snapshot was taken, we should remove them from the snapshot.
func (cache *schedulerCache) removeDeletedNodesFromSnapshot(snapshot *Snapshot) {
	toDelete := len(snapshot.nodeInfoMap) - len(cache.nodes)
	for name := range snapshot.nodeInfoMap {
		if toDelete <= 0 {
			break
		}
		if _, ok := cache.nodes[name]; !ok {
			delete(snapshot.nodeInfoMap, name)
			toDelete--
		}
	}
}

//AssumeStack assume stack to node
func (cache *schedulerCache) AssumeStack(stack *types.Stack) error {
	key, err := schedulernodeinfo.GetStackKey(stack)
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
	n, ok := cache.nodes[stack.Selected.NodeID]
	if !ok {
		n = newNodeInfoListItem(schedulernodeinfo.NewNodeInfo())
		cache.nodes[stack.Selected.NodeID] = n
	}
	n.info.AddStack(stack)
	cache.moveNodeInfoToHead(stack.Selected.NodeID)
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
// Removes a stack from the cached node info. When a node is removed, some pod
// deletion events might arrive later. This is not a problem, as the pods in
// the node are assumed to be removed already.
func (cache *schedulerCache) removeStack(stack *types.Stack) error {
	n, ok := cache.nodes[stack.Selected.NodeID]
	if !ok {
		return nil
	}
	if err := n.info.RemoveStack(stack); err != nil {
		return err
	}
	cache.moveNodeInfoToHead(stack.Selected.NodeID)
	return nil
}

//AddStack add stack
func (cache *schedulerCache) AddStack(stack *types.Stack) error {

	key, err := schedulernodeinfo.GetStackKey(stack)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.stackStates[key]
	switch {
	case ok && cache.assumedStacks[key]:
		if currState.stack.Selected != stack.Selected {
			// The stack was added to a different node than it was assumed to.
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
	key, err := schedulernodeinfo.GetStackKey(oldStack)
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
			logger.Errorf("Stack %v updated on a different node than previously added to.", key)
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
	key, err := schedulernodeinfo.GetStackKey(stack)
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
		if currState.stack.Selected.NodeID != stack.Selected.NodeID {
			logger.Errorf("Stack %v was assumed to be on %v but got added to %v", key,
				stack.Selected.NodeID, currState.stack.Selected.NodeID)
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
	key, err := schedulernodeinfo.GetStackKey(stack)
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

// GetPod might return a pod for which its node has already been deleted from
// the main cache. This is useful to properly process pod update events.
func (cache *schedulerCache) GetStack(stack *types.Stack) (*types.Stack, error) {
	key, err := schedulernodeinfo.GetStackKey(stack)
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

func (cache *schedulerCache) AddNode(node *types.SiteNode) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.nodes[node.SiteID]
	if !ok {
		n = newNodeInfoListItem(schedulernodeinfo.NewNodeInfo())
		cache.nodes[node.SiteID] = n
	}

	cache.moveNodeInfoToHead(node.SiteID)

	cache.nodeTree.addNode(node)
	cache.updateRegionToNode(node)
	return n.info.SetNode(node)
}

func (cache *schedulerCache) UpdateNode(oldNode, newNode *types.SiteNode) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.nodes[newNode.SiteID]
	if !ok {
		n = newNodeInfoListItem(schedulernodeinfo.NewNodeInfo())
		cache.nodes[newNode.SiteID] = n
		cache.nodeTree.addNode(newNode)
	}
	cache.moveNodeInfoToHead(newNode.SiteID)

	cache.nodeTree.updateNode(oldNode, newNode)
	cache.deleteRegionToNode(oldNode)
	cache.updateRegionToNode(newNode)
	return n.info.SetNode(newNode)
}

// RemoveNode removes a node from the cache.
// Some nodes might still have pods because their deletion events didn't arrive
// yet. For most intents and purposes, those pods are removed from the cache,
// having it's source of truth in the cached nodes.
// However, some information on pods (assumedPods, podStates) persist. These
// caches will be eventually consistent as pod deletion events arrive.
func (cache *schedulerCache) RemoveNode(node *types.SiteNode) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	_, ok := cache.nodes[node.SiteID]
	if !ok {
		return fmt.Errorf("node %v is not found", node.SiteID)
	}
	cache.removeNodeInfoFromList(node.SiteID)
	if err := cache.nodeTree.removeNode(node); err != nil {
		return err
	}

	cache.deleteRegionToNode(node)
	return nil
}

func (cache *schedulerCache) updateRegionToNode(newNode *types.SiteNode) {
	_, ok := cache.regionToNode[newNode.Region]
	if !ok {
		cache.regionToNode[newNode.Region] = sets.NewString()
	}
	cache.regionToNode[newNode.Region].Insert(newNode.SiteID)
}

func (cache *schedulerCache) deleteRegionToNode(newNode *types.SiteNode) {
	_, ok := cache.regionToNode[newNode.Region]
	if !ok {
		return
	}
	cache.regionToNode[newNode.Region].Delete(newNode.SiteID)
}

//UpdateNodeWithEipPool update eip pool
func (cache *schedulerCache) UpdateNodeWithEipPool(nodeID string, eipPool *typed.EipPool) error {
	node, ok := cache.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node %v is not found", nodeID)
	}

	err := node.info.UpdateNodeWithEipPool(eipPool)
	if err != nil {
		logger.Errorf("UpdateNodeWithEipPool failed! err: %s", err)
		return err
	}

	cache.moveNodeInfoToHead(nodeID)

	return nil
}

//UpdateNodeWithVolumePool update volume pool
func (cache *schedulerCache) UpdateNodeWithVolumePool(nodeID string, volumePool *typed.RegionVolumePool) error {
	node, ok := cache.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node %v is not found", nodeID)
	}

	err := node.info.UpdateNodeWithVolumePool(volumePool)
	if err != nil {
		logger.Errorf("UpdateNodeWithEipPool failed! err: %s", err)
		return err
	}

	cache.moveNodeInfoToHead(nodeID)

	return nil
}

//UpdateEipPool updates eip pool info about node
func (cache *schedulerCache) UpdateEipPool(eipPool *typed.EipPool) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	nodes, ok := cache.regionToNode[eipPool.Region]
	if !ok {
		return nil
	}

	for node := range nodes {
		err := cache.UpdateNodeWithEipPool(node, eipPool)
		if err != nil {
			logger.Errorf("UpdateNodeWithEipPool failed! err: %s", err)
			continue
		}
	}

	return nil
}

//UpdateVolumePool updates volume pool info about node
func (cache *schedulerCache) UpdateVolumePool(volumePool *typed.RegionVolumePool) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	nodes, ok := cache.regionToNode[volumePool.Region]
	if !ok {
		return nil
	}

	for node := range nodes {
		err := cache.UpdateNodeWithVolumePool(node, volumePool)
		if err != nil {
			logger.Errorf("UpdateNodeWithEipPool failed! err: %s", err)
			continue
		}
	}

	return nil
}

// UpdateNodeWithResInfo update res info
func (cache *schedulerCache) UpdateNodeWithResInfo(siteID string, resInfo types.AllResInfo) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, ok := cache.nodes[siteID]
	if !ok {
		return nil
	}

	err := node.info.UpdateNodeWithResInfo(resInfo)
	if err != nil {
		logger.Errorf("UpdateNodeWithResInfo failed! err: %s", err)
		return err
	}

	cache.moveNodeInfoToHead(siteID)

	return nil
}

func (cache *schedulerCache) UpdateQos(siteID string, netMetricData *types.NetMetricDatas) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	node, ok := cache.nodes[siteID]
	if !ok {
		return nil
	}

	err := node.info.UpdateQos(netMetricData)
	if err != nil {
		logger.Errorf("UpdateQos failed! err: %s", err)
		return err
	}

	cache.moveNodeInfoToHead(siteID)

	return nil
}

//UpdateNodeWithVcpuMem update vcpu and mem
func (cache *schedulerCache) UpdateNodeWithRatio(region string, az string, ratios []types.AllocationRatio) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, node := range cache.nodes {
		if node.info.Node().Region == region && node.info.Node().AvailabilityZone == az {
			err := node.info.UpdateNodeWithRatio(ratios)
			if err != nil {
				logger.Errorf("UpdateNodeWithRatio failed! err: %s", err)
				return err
			}

			cache.moveNodeInfoToHead(node.info.Node().SiteID)
			break
		}
	}

	return nil
}

//UpdateSpotResources update spot resources
func (cache *schedulerCache) UpdateSpotResources(region string, az string, spotRes map[string]types.SpotResource) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, node := range cache.nodes {
		if node.info.Node().Region == region && node.info.Node().AvailabilityZone == az {
			err := node.info.UpdateSpotResources(spotRes)
			if err != nil {
				logger.Errorf("UpdateNodeWithRatio failed! err: %s", err)
				return err
			}
			cache.moveNodeInfoToHead(node.info.Node().SiteID)

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
	for _, node := range cache.nodes {
		region := node.info.Node().Region
		cr, ok := ret[region]
		if !ok {
			cr = types.CloudRegion{Region: region, AvailabilityZone: []string{}}
		}
		cr.AvailabilityZone = append(cr.AvailabilityZone, node.info.Node().AvailabilityZone)
		ret[region] = cr
	}

	return ret
}

//PrintString print node info
func (cache *schedulerCache) PrintString() {
	for _, node := range cache.nodes {
		logger.Infof("siteID: %s, info: %s", node.info.Node().SiteID, node.info.ToString())
	}
}

func (cache *schedulerCache) run() {
	go wait.Until(cache.cleanupExpiredAssumedStacks, cache.period, cache.stop)
}

func (cache *schedulerCache) cleanupExpiredAssumedStacks() {
	cache.cleanupAssumedStacks(time.Now())
}

// cleanupAssumedStacks exists for making test deterministic by taking time as input argument.
// It also reports metrics on the cache size for nodes, stacks, and assumed stacks.
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
	for _, n := range cache.nodes {
		maxSize += len(n.info.Stacks())
	}
	stacks := make([]*types.Stack, 0, maxSize)
	for _, n := range cache.nodes {
		for _, stack := range n.info.Stacks() {
			if stackFilter(stack) {
				stacks = append(stacks, stack)
			}
		}
	}
	return stacks, nil
}
