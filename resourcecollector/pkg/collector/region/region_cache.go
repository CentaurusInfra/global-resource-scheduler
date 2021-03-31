/*Copyright 2020 Authors of Arktos.

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

package region

import (
	"fmt"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"reflect"
	"sync"
)

// Because it needs to be called in Informer, it will not be placed in the internalcache for the time being
type RegionResourceCache struct {
	mutex             sync.Mutex
	RegionResourceMap map[string]*typed.RegionResource
}

func NewRegionCache() *RegionResourceCache {
	r := &RegionResourceCache{}
	r.RegionResourceMap = make(map[string]*typed.RegionResource)
	return r
}

func (r *RegionResourceCache) AddRegionResource(region *typed.RegionResource) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[region.RegionName]; !ok {
		r.RegionResourceMap[region.RegionName] = region
	}
}

func (r *RegionResourceCache) RemoveRegionResource(regionName string) error {
	r.mutex.Lock()
	r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[regionName]; ok {
		delete(r.RegionResourceMap, regionName)
		return nil
	} else {
		return fmt.Errorf("Region %v is not found", regionName)
	}
}

func (r *RegionResourceCache) AddAvailabilityZone(regionName, availabilityZone string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[regionName]; ok {
		if r.RegionResourceMap[regionName].CpuAndMemResources == nil {
			r.RegionResourceMap[regionName].CpuAndMemResources = make(map[string]typed.CpuAndMemResource)
		}
		r.RegionResourceMap[regionName].CpuAndMemResources[availabilityZone] = typed.CpuAndMemResource{TotalVCPUs: 0, TotalMem: 0}
	}
}

func (r *RegionResourceCache) RemoveAvailabilityZone(regionName, availabilityZone string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[regionName]; ok {
		if _, ok := r.RegionResourceMap[regionName].CpuAndMemResources[availabilityZone]; ok {
			delete(r.RegionResourceMap[regionName].CpuAndMemResources, availabilityZone)
			return nil
		} else {
			return fmt.Errorf("availabilityZone %v is not found", availabilityZone)
		}
	} else {
		return fmt.Errorf("regionName %v is not found", regionName)
	}
}

func (r *RegionResourceCache) UpdateCpuAndMemResource(regionName, availabilityZone string, newCpuAndMemResource typed.CpuAndMemResource) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[regionName]; ok {
		if oldCpuAndMemResource, ok := r.RegionResourceMap[regionName].CpuAndMemResources[availabilityZone]; ok {
			if !reflect.DeepEqual(oldCpuAndMemResource, newCpuAndMemResource) {
				r.RegionResourceMap[regionName].CpuAndMemResources[availabilityZone] = newCpuAndMemResource
				return true
			}
		}
	}
	return false
}

func (r *RegionResourceCache) UpdateVolumeResource(regionName string, newVolumeResources []typed.VolumeResource) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[regionName]; ok {
		oldVolumeResources := r.RegionResourceMap[regionName].VolumeResources
		if !reflect.DeepEqual(oldVolumeResources, newVolumeResources) {
			r.RegionResourceMap[regionName].VolumeResources = newVolumeResources
			return true
		}
	}
	return false
}

func (r *RegionResourceCache) AddHostAz(regionName string, hosts []string, az string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.RegionResourceMap[regionName]; !ok {
		r.RegionResourceMap[regionName] = &typed.RegionResource{RegionName: regionName}
	}
	if r.RegionResourceMap[regionName].HostAzMap == nil {
		r.RegionResourceMap[regionName].HostAzMap = make(map[string]string)
	}
	for _, host := range hosts {
		r.RegionResourceMap[regionName].HostAzMap[host] = az
	}
}
