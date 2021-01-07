/*
Copyright 2017 The Kubernetes Authors.
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

// This file contains structures that implement scheduling queue types.
// Scheduling queues hold Stacks waiting to be scheduled. This file implements a
// priority queue which has two sub queues. One sub-queue holds Stacks that are
// being considered for scheduling. This is called activeQ. Another queue holds
// Stacks that are already tried and are determined to be unschedulable. The latter
// is called unschedulableQ.

package queue

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/clock"
)

var (
	queueClosed = "scheduling queue is closed"
)

// If the stack stays in unschedulableQ longer than the unschedulableQTimeInterval,
// the stack will be moved from unschedulableQ to activeQ.
const unschedulableQTimeInterval = 60 * time.Second

// SchedulingQueue is an interface for a queue to store stacks waiting to be scheduled.
// The interface follows a pattern similar to cache.FIFO and cache.Heap and
// makes it easy to use those data structures as a SchedulingQueue.
type SchedulingQueue interface {
	Add(stack *types.Stack) error
	AddIfNotPresent(stack *types.Stack) error
	// AddUnschedulableIfNotPresent adds an unschedulable stack back to scheduling queue.
	// The stackSchedulingCycle represents the current scheduling cycle number which can be
	// returned by calling SchedulingCycle().
	AddUnschedulableIfNotPresent(stack *types.Stack, stackSchedulingCycle int64) error
	// SchedulingCycle returns the current number of scheduling cycle which is
	// cached by scheduling queue. Normally, incrementing this number whenever
	// a stack is popped (e.g. called Pop()) is enough.
	SchedulingCycle() int64
	// Pop removes the head of the queue and returns it. It blocks if the
	// queue is empty and waits until a new item is added to the queue.
	Pop() (*types.Stack, error)
	Update(oldStack, newStack *types.Stack) error
	Delete(stack *types.Stack) error
	MoveAllToActiveQueue()
	NominatedStacksForSite(siteID string) []*types.Stack
	PendingStacks() []*types.Stack
	// Close closes the SchedulingQueue so that the goroutine which is
	// waiting to pop items can exit gracefully.
	Close()
	// UpdateNominatedStackForSite adds the given stack to the nominated stack map or
	// updates it if it already exists.
	UpdateNominatedStackForSite(stack *types.Stack, siteID string)
	// DeleteNominatedStackIfExists deletes nominatedStack from internal cache
	DeleteNominatedStackIfExists(stack *types.Stack)
	// NumUnschedulableStacks returns the number of unschedulable stacks exist in the SchedulingQueue.
	NumUnschedulableStacks() int
}

// NewSchedulingQueue initializes a priority queue as a new scheduling queue.
func NewSchedulingQueue(stop <-chan struct{}, fwk framework.Framework) SchedulingQueue {
	return NewPriorityQueue(stop, fwk)
}

// NominatedSiteID returns nominated site name of a Stack.
func NominatedSiteID(stack *types.Stack) string {
	// TODO: original return stack.Status.NominatedSiteID
	return ""
}

// PriorityQueue implements a scheduling queue.
// The head of PriorityQueue is the highest priority pending stack. This structure
// has three sub queues. One sub-queue holds stacks that are being considered for
// scheduling. This is called activeQ and is a Heap. Another queue holds
// stacks that are already tried and are determined to be unschedulable. The latter
// is called unschedulableQ. The third queue holds stacks that are moved from
// unschedulable queues and will be moved to active queue when backoff are completed.
type PriorityQueue struct {
	stop  <-chan struct{}
	clock clock.Clock
	// stackBackoff tracks backoff for stacks attempting to be rescheduled
	stackBackoff *StackBackoffMap

	lock sync.RWMutex
	cond sync.Cond

	// activeQ is heap structure that scheduler actively looks at to find stacks to
	// schedule. Head of heap is the highest priority stack.
	activeQ *utils.Heap
	// stackBackoffQ is a heap ordered by backoff expiry. Stacks which have completed backoff
	// are popped from this heap before the scheduler looks at activeQ
	stackBackoffQ *utils.Heap
	// unschedulableQ holds stacks that have been tried and determined unschedulable.
	unschedulableQ *UnschedulableStacksMap
	// nominatedStacks is a structures that stores stacks which are nominated to run
	// on site.
	nominatedStacks *nominatedStackMap
	// schedulingCycle represents sequence number of scheduling cycle and is incremented
	// when a stack is popped.
	schedulingCycle int64
	// moveRequestCycle caches the sequence number of scheduling cycle when we
	// received a move request. Unscheduable stacks in and before this scheduling
	// cycle will be put back to activeQueue if we were trying to schedule them
	// when we received move request.
	moveRequestCycle int64

	// closed indicates that the queue is closed.
	// It is mainly used to let Pop() exit its control loop while waiting for an item.
	closed bool
}

// Making sure that PriorityQueue implements SchedulingQueue.
var _ = SchedulingQueue(&PriorityQueue{})

// newStackInfoNoTimestamp builds a StackInfo object without timestamp.
func newStackInfoNoTimestamp(stack *types.Stack) *framework.StackInfo {
	return &framework.StackInfo{
		Stack: stack,
	}
}

// activeQComp is the function used by the activeQ heap algorithm to sort stacks.
// It sorts stacks based on their priority. When priorities are equal, it uses
// StackInfo.timestamp.
func activeQComp(stackInfo1, stackInfo2 interface{}) bool {
	pInfo1 := stackInfo1.(*framework.StackInfo)
	pInfo2 := stackInfo2.(*framework.StackInfo)
	prio1 := utils.GetStackPriority(pInfo1.Stack)
	prio2 := utils.GetStackPriority(pInfo2.Stack)
	return (prio1 > prio2) || (prio1 == prio2 && pInfo1.Timestamp.Before(pInfo2.Timestamp))
}

// NewPriorityQueue creates a PriorityQueue object.
func NewPriorityQueue(stop <-chan struct{}, fwk framework.Framework) *PriorityQueue {
	return NewPriorityQueueWithClock(stop, clock.RealClock{}, fwk)
}

// NewPriorityQueueWithClock creates a PriorityQueue which uses the passed clock for time.
func NewPriorityQueueWithClock(stop <-chan struct{}, clock clock.Clock, fwk framework.Framework) *PriorityQueue {
	comp := activeQComp
	if fwk != nil {
		if queueSortFunc := fwk.QueueSortFunc(); queueSortFunc != nil {
			comp = func(stackInfo1, stackInfo2 interface{}) bool {
				pInfo1 := stackInfo1.(*framework.StackInfo)
				pInfo2 := stackInfo2.(*framework.StackInfo)

				return queueSortFunc(pInfo1, pInfo2)
			}
		}
	}

	pq := &PriorityQueue{
		clock:            clock,
		stop:             stop,
		stackBackoff:     NewStackBackoffMap(1*time.Second, 10*time.Second),
		activeQ:          utils.NewHeap(stackInfoKeyFunc, comp),
		unschedulableQ:   newUnschedulableStacksMap(),
		nominatedStacks:  newNominatedStackMap(),
		moveRequestCycle: -1,
	}
	pq.cond.L = &pq.lock
	pq.stackBackoffQ = utils.NewHeap(stackInfoKeyFunc, pq.stacksCompareBackoffCompleted)

	pq.run()

	return pq
}

// run starts the goroutine to pump from stackBackoffQ to activeQ
func (p *PriorityQueue) run() {
	go wait.Until(p.flushBackoffQCompleted, 1.0*time.Second, p.stop)
	go wait.Until(p.flushUnschedulableQLeftover, 30*time.Second, p.stop)
}

// Add adds a stack to the active queue. It should be called only when a new stack
// is added so there is no chance the stack is already in active/unschedulable/backoff queues
func (p *PriorityQueue) Add(stack *types.Stack) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	pInfo := p.newStackInfo(stack)
	if err := p.activeQ.Add(pInfo); err != nil {
		klog.Errorf("Error adding stack %v/%v/%v to the scheduling queue: %v", stack.Tenant, stack.Namespace, stack.Name, err)
		return err
	}
	if p.unschedulableQ.get(stack) != nil {
		klog.Errorf("Error: stack %v/%v/%v is already in the unschedulable queue.", stack.Tenant, stack.Namespace, stack.Name)
		p.unschedulableQ.delete(stack)
	}
	// Delete stack from backoffQ if it is backing off
	if err := p.stackBackoffQ.Delete(pInfo); err == nil {
		klog.Errorf("Error: stack %v/%v/%v is already in the stackBackoff queue.", stack.Tenant, stack.Namespace, stack.Name)
	}
	p.nominatedStacks.add(stack, "")
	p.cond.Broadcast()

	return nil
}

// AddIfNotPresent adds a stack to the active queue if it is not present in any of
// the queues. If it is present in any, it doesn't do any thing.
func (p *PriorityQueue) AddIfNotPresent(stack *types.Stack) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.unschedulableQ.get(stack) != nil {
		return nil
	}

	pInfo := p.newStackInfo(stack)
	if _, exists, _ := p.activeQ.Get(pInfo); exists {
		return nil
	}
	if _, exists, _ := p.stackBackoffQ.Get(pInfo); exists {
		return nil
	}
	err := p.activeQ.Add(pInfo)
	if err != nil {
		klog.Errorf("Error adding stack %v/%v/%v to the scheduling queue: %v", stack.Tenant, stack.Namespace, stack.Name, err)
	} else {
		p.nominatedStacks.add(stack, "")
		p.cond.Broadcast()
	}
	return err
}

// nsNameForStack returns a namespacedname for a stack
func nsNameForStack(stack *types.Stack) ktypes.NamespacedName {
	return ktypes.NamespacedName{
		Tenant:    stack.Tenant,
		Namespace: stack.Namespace,
		Name:      stack.Name,
	}
}

// clearStackBackoff clears all backoff state for a stack (resets expiry)
func (p *PriorityQueue) clearStackBackoff(stack *types.Stack) {
	p.stackBackoff.ClearStackBackoff(nsNameForStack(stack))
}

// isStackBackingOff returns true if a stack is still waiting for its backoff timer.
// If this returns true, the stack should not be re-tried.
func (p *PriorityQueue) isStackBackingOff(stack *types.Stack) bool {
	boTime, exists := p.stackBackoff.GetBackoffTime(nsNameForStack(stack))
	if !exists {
		return false
	}
	return boTime.After(p.clock.Now())
}

// backoffStack checks if stack is currently undergoing backoff. If it is not it updates the backoff
// timeout otherwise it does nothing.
func (p *PriorityQueue) backoffStack(stack *types.Stack) {
	p.stackBackoff.CleanupStacksCompletesBackingoff()

	stackID := nsNameForStack(stack)
	boTime, found := p.stackBackoff.GetBackoffTime(stackID)
	if !found || boTime.Before(p.clock.Now()) {
		p.stackBackoff.BackoffStack(stackID)
	}
}

// SchedulingCycle returns current scheduling cycle.
func (p *PriorityQueue) SchedulingCycle() int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.schedulingCycle
}

// AddUnschedulableIfNotPresent inserts a stack that cannot be scheduled into
// the queue, unless it is already in the queue. Normally, PriorityQueue puts
// unschedulable stacks in `unschedulableQ`. But if there has been a recent move
// request, then the stack is put in `stackBackoffQ`.
func (p *PriorityQueue) AddUnschedulableIfNotPresent(stack *types.Stack, stackSchedulingCycle int64) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.unschedulableQ.get(stack) != nil {
		return fmt.Errorf("stack is already present in unschedulableQ")
	}

	pInfo := p.newStackInfo(stack)
	if _, exists, _ := p.activeQ.Get(pInfo); exists {
		return fmt.Errorf("stack is already present in the activeQ")
	}
	if _, exists, _ := p.stackBackoffQ.Get(pInfo); exists {
		return fmt.Errorf("stack is already present in the backoffQ")
	}

	// Every unschedulable stack is subject to backoff timers.
	p.backoffStack(stack)

	// If a move request has been received, move it to the BackoffQ, otherwise move
	// it to unschedulableQ.
	if p.moveRequestCycle >= stackSchedulingCycle {
		if err := p.stackBackoffQ.Add(pInfo); err != nil {
			return fmt.Errorf("error adding stack %v to the backoff queue: %v", stack.Name, err)
		}
	} else {
		p.unschedulableQ.addOrUpdate(pInfo)
	}

	p.nominatedStacks.add(stack, "")
	return nil

}

// flushBackoffQCompleted Moves all stacks from backoffQ which have completed backoff in to activeQ
func (p *PriorityQueue) flushBackoffQCompleted() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		rawStackInfo := p.stackBackoffQ.Peek()
		if rawStackInfo == nil {
			return
		}
		stack := rawStackInfo.(*framework.StackInfo).Stack
		boTime, found := p.stackBackoff.GetBackoffTime(nsNameForStack(stack))
		if !found {
			klog.Errorf("Unable to find backoff value for stack %v in backoffQ", nsNameForStack(stack))
			p.stackBackoffQ.Pop()
			p.activeQ.Add(rawStackInfo)
			defer p.cond.Broadcast()
			continue
		}

		if boTime.After(p.clock.Now()) {
			return
		}
		_, err := p.stackBackoffQ.Pop()
		if err != nil {
			klog.Errorf("Unable to pop stack %v from backoffQ despite backoff completion.", nsNameForStack(stack))
			return
		}
		p.activeQ.Add(rawStackInfo)
		defer p.cond.Broadcast()
	}
}

// flushUnschedulableQLeftover moves stack which stays in unschedulableQ longer than the durationStayUnschedulableQ
// to activeQ.
func (p *PriorityQueue) flushUnschedulableQLeftover() {
	p.lock.Lock()
	defer p.lock.Unlock()

	var stacksToMove []*framework.StackInfo
	currentTime := p.clock.Now()
	for _, pInfo := range p.unschedulableQ.stackInfoMap {
		lastScheduleTime := pInfo.Timestamp
		if currentTime.Sub(lastScheduleTime) > unschedulableQTimeInterval {
			stacksToMove = append(stacksToMove, pInfo)
		}
	}

	if len(stacksToMove) > 0 {
		p.moveStacksToActiveQueue(stacksToMove)
	}
}

// Pop removes the head of the active queue and returns it. It blocks if the
// activeQ is empty and waits until a new item is added to the queue. It
// increments scheduling cycle when a stack is popped.
func (p *PriorityQueue) Pop() (*types.Stack, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for p.activeQ.Len() == 0 {
		// When the queue is empty, invocation of Pop() is blocked until new item is enqueued.
		// When Close() is called, the p.closed is set and the condition is broadcast,
		// which causes this loop to continue and return from the Pop().
		if p.closed {
			return nil, fmt.Errorf(queueClosed)
		}
		p.cond.Wait()
	}
	obj, err := p.activeQ.Pop()
	if err != nil {
		return nil, err
	}
	pInfo := obj.(*framework.StackInfo)
	p.schedulingCycle++
	return pInfo.Stack, err
}

// isStackUpdated checks if the stack is updated in a way that it may have become
// schedulable. It drops status of the stack and compares it with old version.
func isStackUpdated(oldStack, newStack *types.Stack) bool {
	strip := func(stack *types.Stack) *types.Stack {
		p := stack.DeepCopy()
		return p
	}
	return !reflect.DeepEqual(strip(oldStack), strip(newStack))
}

// Update updates a stack in the active or backoff queue if present. Otherwise, it removes
// the item from the unschedulable queue if stack is updated in a way that it may
// become schedulable and adds the updated one to the active queue.
// If stack is not present in any of the queues, it is added to the active queue.
func (p *PriorityQueue) Update(oldStack, newStack *types.Stack) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if oldStack != nil {
		oldStackInfo := newStackInfoNoTimestamp(oldStack)
		// If the stack is already in the active queue, just update it there.
		if oldStackInfo, exists, _ := p.activeQ.Get(oldStackInfo); exists {
			p.nominatedStacks.update(oldStack, newStack)
			newStackInfo := newStackInfoNoTimestamp(newStack)
			newStackInfo.Timestamp = oldStackInfo.(*framework.StackInfo).Timestamp
			err := p.activeQ.Update(newStackInfo)
			return err
		}

		// If the stack is in the backoff queue, update it there.
		if oldStackInfo, exists, _ := p.stackBackoffQ.Get(oldStackInfo); exists {
			p.nominatedStacks.update(oldStack, newStack)
			p.stackBackoffQ.Delete(newStackInfoNoTimestamp(oldStack))
			newStackInfo := newStackInfoNoTimestamp(newStack)
			newStackInfo.Timestamp = oldStackInfo.(*framework.StackInfo).Timestamp
			err := p.activeQ.Add(newStackInfo)
			if err == nil {
				p.cond.Broadcast()
			}
			return err
		}
	}

	// If the stack is in the unschedulable queue, updating it may make it schedulable.
	if usStackInfo := p.unschedulableQ.get(newStack); usStackInfo != nil {
		p.nominatedStacks.update(oldStack, newStack)
		newStackInfo := newStackInfoNoTimestamp(newStack)
		newStackInfo.Timestamp = usStackInfo.Timestamp
		if isStackUpdated(oldStack, newStack) {
			// If the stack is updated reset backoff
			p.clearStackBackoff(newStack)
			p.unschedulableQ.delete(usStackInfo.Stack)
			err := p.activeQ.Add(newStackInfo)
			if err == nil {
				p.cond.Broadcast()
			}
			return err
		}
		// Stack is already in unschedulable queue and hasnt updated, no need to backoff again
		p.unschedulableQ.addOrUpdate(newStackInfo)
		return nil
	}
	// If stack is not in any of the queues, we put it in the active queue.
	err := p.activeQ.Add(p.newStackInfo(newStack))
	if err == nil {
		p.nominatedStacks.add(newStack, "")
		p.cond.Broadcast()
	}
	return err
}

// Delete deletes the item from either of the two queues. It assumes the stack is
// only in one queue.
func (p *PriorityQueue) Delete(stack *types.Stack) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.nominatedStacks.delete(stack)
	err := p.activeQ.Delete(newStackInfoNoTimestamp(stack))
	if err != nil { // The item was probably not found in the activeQ.
		p.clearStackBackoff(stack)
		p.stackBackoffQ.Delete(newStackInfoNoTimestamp(stack))
		p.unschedulableQ.delete(stack)
	}
	return nil
}

// MoveAllToActiveQueue moves all stacks from unschedulableQ to activeQ. This
// function adds all stacks and then signals the condition variable to ensure that
// if Pop() is waiting for an item, it receives it after all the stacks are in the
// queue and the head is the highest priority stack.
func (p *PriorityQueue) MoveAllToActiveQueue() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, pInfo := range p.unschedulableQ.stackInfoMap {
		stack := pInfo.Stack
		if p.isStackBackingOff(stack) {
			if err := p.stackBackoffQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding stack %v to the backoff queue: %v", stack.Name, err)
			}
		} else {
			if err := p.activeQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding stack %v to the scheduling queue: %v", stack.Name, err)
			}
		}
	}
	p.unschedulableQ.clear()
	p.moveRequestCycle = p.schedulingCycle
	p.cond.Broadcast()
}

// NOTE: this function assumes lock has been acquired in caller
func (p *PriorityQueue) moveStacksToActiveQueue(stackInfoList []*framework.StackInfo) {
	for _, pInfo := range stackInfoList {
		stack := pInfo.Stack
		if p.isStackBackingOff(stack) {
			if err := p.stackBackoffQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding stack %v to the backoff queue: %v", stack.Name, err)
			}
		} else {
			if err := p.activeQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding stack %v to the scheduling queue: %v", stack.Name, err)
			}
		}
		p.unschedulableQ.delete(stack)
	}
	p.moveRequestCycle = p.schedulingCycle
	p.cond.Broadcast()
}

// NominatedStacksForSite returns stacks that are nominated to run on the given site,
// but they are waiting for other stacks to be removed from the site before they
// can be actually scheduled.
func (p *PriorityQueue) NominatedStacksForSite(siteID string) []*types.Stack {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.nominatedStacks.stacksForSite(siteID)
}

// PendingStacks returns all the pending stacks in the queue. This function is
// used for debugging purposes in the scheduler cache dumper and comparer.
func (p *PriorityQueue) PendingStacks() []*types.Stack {
	p.lock.RLock()
	defer p.lock.RUnlock()
	result := []*types.Stack{}
	for _, pInfo := range p.activeQ.List() {
		result = append(result, pInfo.(*framework.StackInfo).Stack)
	}
	for _, pInfo := range p.stackBackoffQ.List() {
		result = append(result, pInfo.(*framework.StackInfo).Stack)
	}
	for _, pInfo := range p.unschedulableQ.stackInfoMap {
		result = append(result, pInfo.Stack)
	}
	return result
}

// Close closes the priority queue.
func (p *PriorityQueue) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.closed = true
	p.cond.Broadcast()
}

// DeleteNominatedStackIfExists deletes stack nominatedStacks.
func (p *PriorityQueue) DeleteNominatedStackIfExists(stack *types.Stack) {
	p.lock.Lock()
	p.nominatedStacks.delete(stack)
	p.lock.Unlock()
}

// UpdateNominatedStackForSite adds a stack to the nominated stacks of the given site.
// This is called during the preemption process after a site is nominated to run
// the stack. We update the structure before sending a request to update the stack
// object to avoid races with the following scheduling cycles.
func (p *PriorityQueue) UpdateNominatedStackForSite(stack *types.Stack, siteID string) {
	p.lock.Lock()
	p.nominatedStacks.add(stack, siteID)
	p.lock.Unlock()
}

func (p *PriorityQueue) stacksCompareBackoffCompleted(stackInfo1, stackInfo2 interface{}) bool {
	pInfo1 := stackInfo1.(*framework.StackInfo)
	pInfo2 := stackInfo2.(*framework.StackInfo)
	bo1, _ := p.stackBackoff.GetBackoffTime(nsNameForStack(pInfo1.Stack))
	bo2, _ := p.stackBackoff.GetBackoffTime(nsNameForStack(pInfo2.Stack))
	return bo1.Before(bo2)
}

// NumUnschedulableStacks returns the number of unschedulable stacks exist in the SchedulingQueue.
func (p *PriorityQueue) NumUnschedulableStacks() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.unschedulableQ.stackInfoMap)
}

// newStackInfo builds a StackInfo object.
func (p *PriorityQueue) newStackInfo(stack *types.Stack) *framework.StackInfo {
	if p.clock == nil {
		return &framework.StackInfo{
			Stack: stack,
		}
	}

	return &framework.StackInfo{
		Stack:     stack,
		Timestamp: p.clock.Now(),
	}
}

// UnschedulableStacksMap holds stacks that cannot be scheduled. This data structure
// is used to implement unschedulableQ.
type UnschedulableStacksMap struct {
	// stackInfoMap is a map key by a stack's full-name and the value is a pointer to the StackInfo.
	stackInfoMap map[string]*framework.StackInfo
	keyFunc      func(*types.Stack) string
}

// Add adds a stack to the unschedulable stackInfoMap.
func (u *UnschedulableStacksMap) addOrUpdate(pInfo *framework.StackInfo) {
	stackID := u.keyFunc(pInfo.Stack)
	u.stackInfoMap[stackID] = pInfo
}

// Delete deletes a stack from the unschedulable stackInfoMap.
func (u *UnschedulableStacksMap) delete(stack *types.Stack) {
	stackID := u.keyFunc(stack)
	delete(u.stackInfoMap, stackID)
}

// Get returns the StackInfo if a stack with the same key as the key of the given "stack"
// is found in the map. It returns nil otherwise.
func (u *UnschedulableStacksMap) get(stack *types.Stack) *framework.StackInfo {
	stackKey := u.keyFunc(stack)
	if pInfo, exists := u.stackInfoMap[stackKey]; exists {
		return pInfo
	}
	return nil
}

// Clear removes all the entries from the unschedulable stackInfoMap.
func (u *UnschedulableStacksMap) clear() {
	u.stackInfoMap = make(map[string]*framework.StackInfo)
}

// newUnschedulableStacksMap initializes a new object of UnschedulableStacksMap.
func newUnschedulableStacksMap() *UnschedulableStacksMap {
	return &UnschedulableStacksMap{
		stackInfoMap: make(map[string]*framework.StackInfo),
		keyFunc:      utils.GetStackFullName,
	}
}

// nominatedStackMap is a structure that stores stacks nominated to run on site.
// It exists because nominatedSiteName of stack objects stored in the structure
// may be different than what scheduler has here. We should be able to find stacks
// by their UID and update/delete them.
type nominatedStackMap struct {
	// nominatedStacks is a map keyed by a site name and the value is a list of
	// stacks which are nominated to run on the site. These are stacks which can be in
	// the activeQ or unschedulableQ.
	nominatedStacks map[string][]*types.Stack
	// nominatedStackToSite is map keyed by a Stack UID to the site name where it is
	// nominated.
	nominatedStackToSite map[string]string
}

func (npm *nominatedStackMap) add(p *types.Stack, siteID string) {
	// always delete the stack if it already exist, to ensure we never store more than
	// one instance of the stack.
	npm.delete(p)

	nnn := siteID
	if len(nnn) == 0 {
		nnn = NominatedSiteID(p)
		if len(nnn) == 0 {
			return
		}
	}
	npm.nominatedStackToSite[p.UID] = nnn
	for _, np := range npm.nominatedStacks[nnn] {
		if np.UID == p.UID {
			klog.V(4).Infof("Stack %v/%v/%v already exists in the nominated map!", p.Tenant, p.Namespace, p.Name)
			return
		}
	}
	npm.nominatedStacks[nnn] = append(npm.nominatedStacks[nnn], p)
}

func (npm *nominatedStackMap) delete(p *types.Stack) {
	nnn, ok := npm.nominatedStackToSite[p.UID]
	if !ok {
		return
	}
	for i, np := range npm.nominatedStacks[nnn] {
		if np.UID == p.UID {
			npm.nominatedStacks[nnn] = append(npm.nominatedStacks[nnn][:i], npm.nominatedStacks[nnn][i+1:]...)
			if len(npm.nominatedStacks[nnn]) == 0 {
				delete(npm.nominatedStacks, nnn)
			}
			break
		}
	}
	delete(npm.nominatedStackToSite, p.UID)
}

func (npm *nominatedStackMap) update(oldStack, newStack *types.Stack) {
	// In some cases, an Update event with no "NominatedSite" present is received right
	// after a site("NominatedSite") is reserved for this stack in memory.
	// In this case, we need to keep reserving the NominatedSite when updating the stack pointer.
	siteID := ""
	// We won't fall into below `if` block if the Update event represents:
	// (1) NominatedSite info is added
	// (2) NominatedSite info is updated
	// (3) NominatedSite info is removed
	if NominatedSiteID(oldStack) == "" && NominatedSiteID(newStack) == "" {
		if nnn, ok := npm.nominatedStackToSite[oldStack.UID]; ok {
			// This is the only case we should continue reserving the NominatedSite
			siteID = nnn
		}
	}
	// We update irrespective of the nominatedSiteName changed or not, to ensure
	// that stack pointer is updated.
	npm.delete(oldStack)
	npm.add(newStack, siteID)
}

func (npm *nominatedStackMap) stacksForSite(siteID string) []*types.Stack {
	if list, ok := npm.nominatedStacks[siteID]; ok {
		return list
	}
	return nil
}

func newNominatedStackMap() *nominatedStackMap {
	return &nominatedStackMap{
		nominatedStacks:      make(map[string][]*types.Stack),
		nominatedStackToSite: make(map[string]string),
	}
}

// MakeNextStackFunc returns a function to retrieve the next stack from a given
// scheduling queue
func MakeNextStackFunc(queue SchedulingQueue) func() *types.Stack {
	return func() *types.Stack {
		stack, err := queue.Pop()
		if err == nil {
			klog.V(4).Infof("About to try and schedule stack %v/%v/%v", stack.Tenant, stack.Namespace, stack.Name)
			return stack
		}
		klog.Errorf("Error while retrieving next stack from scheduling queue: %v", err)
		return nil
	}
}

func stackInfoKeyFunc(obj interface{}) (string, error) {
	stack := obj.(*framework.StackInfo).Stack
	// return tenant + namespace + name as keyFunc
	return stack.Tenant + "/" + stack.Namespace + "/" + stack.Name, nil
}
