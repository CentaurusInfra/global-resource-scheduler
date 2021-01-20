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
	"reflect"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/wait"

	"github.com/deckarep/golang-set"
)

// ResourceEventHandler can handle notifications for events that happen to a
// resource. The events are informational only, so you can't return an
// error.

type ResourceEventHandler interface {
	OnList(obj []interface{})
}

// ResourceEventHandlerFuncs is an adaptor to let you easily specify as many or
// as few of the notification functions as you want while still implementing
// ResourceEventHandler.
type ResourceEventHandlerFuncs struct {
	ListFunc func(obj []interface{})
}

// OnAdd calls AddFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnList(obj []interface{}) {
	if r.ListFunc != nil {
		r.ListFunc(obj)
	}
}

// SharedInformer is the cache in order to not get the flavor, site info, volume type etc.
// from OpenStack each time
type SharedInformer interface {
	// AddEventHandler adds an event handler to the shared informer using the shared informer's resync
	// period.  Events to a single handler are delivered sequentially, but there is no coordination
	// between different handlers.
	AddEventHandler(handler ResourceEventHandler)

	// GetStore returns the Store.
	GetStore() ThreadSafeStore

	// Run starts the shared informer, which will be stopped when stopCh is closed.
	Run(stopCh <-chan struct{})
	// HasSynced returns true if the shared informer's store has synced.
	HasSynced() bool

	SyncOnce()

	SetIsCache(isCache bool)
}

type sharedInformer struct {
	store ThreadSafeStore
	// This block is tracked to handle late initialization of the controller
	lister ListerInterface

	eventHandler ResourceEventHandler

	isCache bool

	listOpts interface{}

	name string

	key string

	period time.Duration

	hasSynced bool
}

// NewEmptyInformer new an empty informer
func NewEmptyInformer() SharedInformer {
	return &sharedInformer{}
}

// NewSharedInformer creates a new instance for the listWatcher.
func NewSharedInformer(lw ListerInterface, period time.Duration, name string, key string, listOpts interface{}) SharedInformer {
	return &sharedInformer{lister: lw, period: period, name: name, key: key,
		isCache: true, listOpts: listOpts, store: NewThreadSafeStore(), hasSynced: false}
}

func (s *sharedInformer) Run(stopCh <-chan struct{}) {

	logger.Infof("Run %s Informer", s.name)

	if s.period <= 0 {
		s.period = 30 * time.Second
	}

	go wait.Until(s.SyncOnce, s.period, stopCh)
}

//SetIsCache set iscache flag
func (s *sharedInformer) SetIsCache(isCache bool) {
	s.isCache = isCache
}

//AddEventHandler add event handler
func (s *sharedInformer) AddEventHandler(handler ResourceEventHandler) {
	s.eventHandler = handler
}

func (s *sharedInformer) HasSynced() bool {
	return s.hasSynced
}

func (s *sharedInformer) GetStore() ThreadSafeStore {
	return s.store
}

func (s *sharedInformer) SyncOnce() {
	listObj, err := s.lister.List(s.listOpts)
	if err != nil {
		logger.Errorf("sharedInformer(%s) list failed! err: %s", s.name, err.Error())
		return
	}

	if s.eventHandler != nil {
		s.eventHandler.OnList(listObj)
	}

	if !s.isCache {
		s.hasSynced = true
		return
	}

	newKeysSet := mapset.NewSet()
	for _, item := range listObj {
		rVal := reflect.ValueOf(item)
		if rVal.Kind() == reflect.Ptr {
			rVal = rVal.Elem()
		}

		if !rVal.IsValid() {
			continue
		}

		val := rVal.FieldByName(s.key).String()
		s.store.Add(val, item)
		newKeysSet.Add(val)
	}

	oldKeys := s.store.ListKeys()
	oldKeysSet := mapset.NewSet()
	for _, key := range oldKeys {
		oldKeysSet.Add(key)
	}

	diffKeys := oldKeysSet.Difference(newKeysSet)
	for key := range diffKeys.Iter() {
		s.store.Delete(key.(string))
	}

	s.hasSynced = true
}
