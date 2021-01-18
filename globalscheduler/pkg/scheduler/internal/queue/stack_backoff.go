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

package queue

import (
	"sync"
	"time"

	ktypes "k8s.io/apimachinery/pkg/types"
)

// StackBackoffMap is a structure that stores backoff related information for stacks
type StackBackoffMap struct {
	// lock for performing actions on this StackBackoffMap
	lock sync.RWMutex
	// initial backoff duration
	initialDuration time.Duration
	// maximal backoff duration
	maxDuration time.Duration
	// map for stack -> number of attempts for this stack
	stackAttempts map[ktypes.NamespacedName]int
	// map for stack -> lastUpdateTime stack of this stack
	stackLastUpdateTime map[ktypes.NamespacedName]time.Time
}

// NewStackBackoffMap creates a StackBackoffMap with initial duration and max duration.
func NewStackBackoffMap(initialDuration, maxDuration time.Duration) *StackBackoffMap {
	return &StackBackoffMap{
		initialDuration:     initialDuration,
		maxDuration:         maxDuration,
		stackAttempts:       make(map[ktypes.NamespacedName]int),
		stackLastUpdateTime: make(map[ktypes.NamespacedName]time.Time),
	}
}

// GetBackoffTime returns the time that nsStack completes backoff
func (pbm *StackBackoffMap) GetBackoffTime(nsStack ktypes.NamespacedName) (time.Time, bool) {
	pbm.lock.RLock()
	defer pbm.lock.RUnlock()
	if _, found := pbm.stackAttempts[nsStack]; found == false {
		return time.Time{}, false
	}
	lastUpdateTime := pbm.stackLastUpdateTime[nsStack]
	backoffDuration := pbm.calculateBackoffDuration(nsStack)
	backoffTime := lastUpdateTime.Add(backoffDuration)
	return backoffTime, true
}

// calculateBackoffDuration is a helper function for calculating the backoffDuration
// based on the number of attempts the stack has made.
func (pbm *StackBackoffMap) calculateBackoffDuration(nsStack ktypes.NamespacedName) time.Duration {
	backoffDuration := pbm.initialDuration
	if _, found := pbm.stackAttempts[nsStack]; found {
		for i := 1; i < pbm.stackAttempts[nsStack]; i++ {
			backoffDuration = backoffDuration * 2
			if backoffDuration > pbm.maxDuration {
				return pbm.maxDuration
			}
		}
	}
	return backoffDuration
}

// clearStackBackoff removes all tracking information for nsStack.
// Lock is supposed to be acquired by caller.
func (pbm *StackBackoffMap) clearStackBackoff(nsStack ktypes.NamespacedName) {
	delete(pbm.stackAttempts, nsStack)
	delete(pbm.stackLastUpdateTime, nsStack)
}

// ClearStackBackoff is the thread safe version of clearStackBackoff
func (pbm *StackBackoffMap) ClearStackBackoff(nsStack ktypes.NamespacedName) {
	pbm.lock.Lock()
	pbm.clearStackBackoff(nsStack)
	pbm.lock.Unlock()
}

// CleanupStacksCompletesBackingoff execute garbage collection on the stack backoff,
// i.e, it will remove a stack from the StackBackoffMap if
// lastUpdateTime + maxBackoffDuration is before the current timestamp
func (pbm *StackBackoffMap) CleanupStacksCompletesBackingoff() {
	pbm.lock.Lock()
	defer pbm.lock.Unlock()
	for stack, value := range pbm.stackLastUpdateTime {
		if value.Add(pbm.maxDuration).Before(time.Now()) {
			pbm.clearStackBackoff(stack)
		}
	}
}

// BackoffStack updates the lastUpdateTime for an nsStack,
// and increases its numberOfAttempts by 1
func (pbm *StackBackoffMap) BackoffStack(nsStack ktypes.NamespacedName) {
	pbm.lock.Lock()
	pbm.stackLastUpdateTime[nsStack] = time.Now()
	pbm.stackAttempts[nsStack]++
	pbm.lock.Unlock()
}
