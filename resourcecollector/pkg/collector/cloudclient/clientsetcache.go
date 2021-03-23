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

package cloudclient

import (
	"sync"

	"k8s.io/klog"
)

type ClientSetCache struct {
	mutex        sync.Mutex
	clientSetMap map[string]*ClientSet
}

func NewClientSetCache() *ClientSetCache {
	c := &ClientSetCache{}
	c.clientSetMap = make(map[string]*ClientSet)
	return c
}

func (c *ClientSetCache) GetClientSet(siteEndpoint string) (*ClientSet, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if clientSet, ok := c.clientSetMap[siteEndpoint]; ok {
		return clientSet, nil
	}
	clientSet, err := NewClientSet(siteEndpoint)
	if err != nil {
		klog.Errorf("NewClientSet[%s] err: %s", siteEndpoint, err.Error())
		return nil, err
	}
	c.clientSetMap[siteEndpoint] = clientSet
	return clientSet, nil
}

func (c *ClientSetCache) RefreshClientSets(siteEndpoints []string) {
	newMap := make(map[string]*ClientSet)
	for _, siteEndpoint := range siteEndpoints {
		clientSet, err := NewClientSet(siteEndpoint)
		if err != nil {
			klog.Errorf("NewClientSet[%s] err: %s", siteEndpoint, err.Error())
			continue
		}
		newMap[siteEndpoint] = clientSet
	}

	c.mutex.Lock()
	c.clientSetMap = newMap
	c.mutex.Unlock()

	klog.Infof("RefreshClientSets success, len of clientSetMap is %d", len(c.clientSetMap))
}
