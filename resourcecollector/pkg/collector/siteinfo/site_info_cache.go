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

package siteinfo

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"sync"
)

// Because it needs to be called in Informer, it will not be placed in the internalcache for the time being
type SiteInfoCache struct {
	mutex       sync.Mutex
	SiteInfoMap map[string]*typed.SiteInfo
}

func NewSiteInfoCache() *SiteInfoCache {
	c := &SiteInfoCache{}
	c.SiteInfoMap = make(map[string]*typed.SiteInfo)
	return c
}

func (c *SiteInfoCache) AddSite(siteInfo *typed.SiteInfo) {
	c.mutex.Lock()
	c.SiteInfoMap[siteInfo.SiteID] = siteInfo
	c.mutex.Unlock()
}

func (c *SiteInfoCache) RemoveSite(siteID string) {
	c.mutex.Lock()
	delete(c.SiteInfoMap, siteID)
	c.mutex.Unlock()
}

func (c *SiteInfoCache) GetAllSiteEndpoints() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ret := make([]string, 0, len(c.SiteInfoMap))
	for _, siteInfo := range c.SiteInfoMap {
		ret = append(ret, siteInfo.EipNetworkID)
	}

	return ret
}
