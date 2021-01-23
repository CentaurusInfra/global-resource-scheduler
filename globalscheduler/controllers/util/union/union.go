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

package union

import (
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
	"reflect"
)

// TBD: Change all the uniuon data structure to "map" instead of "array" so that do not need to do the loop.
func UpdateUnion(schedulerUnion schedulercrdv1.ClusterUnion, cluster *clustercrdv1.Cluster) schedulercrdv1.ClusterUnion {
	// GeoLocation Union
	schedulerUnion.GeoLocation = unionGeoLocation(schedulerUnion.GeoLocation, cluster.Spec.GeoLocation)

	// Region Union
	schedulerUnion.Region = unionRegion(schedulerUnion.Region, cluster.Spec.Region)

	// Operator Union
	schedulerUnion.Operator = unionOperator(schedulerUnion.Operator, cluster.Spec.Operator)

	// Flavors Union
	schedulerUnion.Flavors = unionFlavors(schedulerUnion.Flavors, cluster.Spec.Flavors)

	// Storage Union
	schedulerUnion.Storage = unionStorage(schedulerUnion.Storage, cluster.Spec.Storage)

	// EipCapacity Union
	schedulerUnion.EipCapacity = unionEipCapacity(schedulerUnion.EipCapacity, cluster.Spec.EipCapacity)

	// CPUCapacity Union
	schedulerUnion.CPUCapacity = unionCPUCapacity(schedulerUnion.CPUCapacity, cluster.Spec.CPUCapacity)

	// MemCapacity Union
	schedulerUnion.MemCapacity = unionMemCapacity(schedulerUnion.MemCapacity, cluster.Spec.MemCapacity)

	// ServerPrice Union
	schedulerUnion.ServerPrice = unionServerPrice(schedulerUnion.ServerPrice, cluster.Spec.ServerPrice)

	return schedulerUnion
}

func unionServerPrice(unionServerPrice []int64, serverPrice int64) []int64 {
	m := make(map[int64]int)
	for _, v := range unionServerPrice {
		m[v]++
	}

	times, _ := m[serverPrice]
	if times == 0 {
		unionServerPrice = append(unionServerPrice, serverPrice)
	}

	return unionServerPrice
}

func unionMemCapacity(unionMemCapacity []int64, memCapacity int64) []int64 {
	m := make(map[int64]int)
	for _, v := range unionMemCapacity {
		m[v]++
	}

	times, _ := m[memCapacity]
	if times == 0 {
		unionMemCapacity = append(unionMemCapacity, memCapacity)
	}

	return unionMemCapacity
}

func unionCPUCapacity(unionCPUCapacity []int64, cpuCapacity int64) []int64 {
	m := make(map[int64]int)
	for _, v := range unionCPUCapacity {
		m[v]++
	}

	times, _ := m[cpuCapacity]
	if times == 0 {
		unionCPUCapacity = append(unionCPUCapacity, cpuCapacity)
	}

	return unionCPUCapacity
}

func unionEipCapacity(unionEipCapacity []int64, eipCapacity int64) []int64 {
	m := make(map[int64]int)
	for _, v := range unionEipCapacity {
		m[v]++
	}

	times, _ := m[eipCapacity]
	if times == 0 {
		unionEipCapacity = append(unionEipCapacity, eipCapacity)
	}

	return unionEipCapacity
}

func unionOperator(unionOperator []*clustercrdv1.OperatorInfo, operator clustercrdv1.OperatorInfo) []*clustercrdv1.OperatorInfo {
	if len(unionOperator) == 0 {
		unionOperator = append(unionOperator, &operator)
		return unionOperator
	}

	for _, v := range unionOperator {
		if reflect.DeepEqual(v, &operator) {
			return unionOperator
		}
	}

	unionOperator = append(unionOperator, &operator)
	return unionOperator
}

func unionRegion(unionRegion []*clustercrdv1.RegionInfo, region clustercrdv1.RegionInfo) []*clustercrdv1.RegionInfo {
	if len(unionRegion) == 0 {
		unionRegion = append(unionRegion, &region)
		return unionRegion
	}

	for _, v := range unionRegion {
		if reflect.DeepEqual(v, &region) {
			return unionRegion
		}
	}

	unionRegion = append(unionRegion, &region)
	return unionRegion
}

func unionGeoLocation(unionGeoLocation []*clustercrdv1.GeolocationInfo, geoLocation clustercrdv1.GeolocationInfo) []*clustercrdv1.GeolocationInfo {
	if len(unionGeoLocation) == 0 {
		unionGeoLocation = append(unionGeoLocation, &geoLocation)
		return unionGeoLocation
	}

	for _, v := range unionGeoLocation {
		if reflect.DeepEqual(v, &geoLocation) {
			return unionGeoLocation
		}
	}

	unionGeoLocation = append(unionGeoLocation, &geoLocation)
	return unionGeoLocation
}

func unionStorage(unionStorage []*clustercrdv1.StorageSpec, storage []clustercrdv1.StorageSpec) []*clustercrdv1.StorageSpec {
	if len(unionStorage) == 0 {
		for _, y := range storage {
			unionStorage = append(unionStorage, &y)
		}
		return unionStorage
	}

	var result []*clustercrdv1.StorageSpec

	for _, x := range storage {
		for idx, y := range unionStorage {
			if reflect.DeepEqual(&x, y) {
				break
			} else if idx == len(unionStorage)-1 {
				res := x
				result = append(result, &res)
			}
		}
	}

	for _, z := range result {
		unionStorage = append(unionStorage, z)
	}
	return unionStorage
}

func unionFlavors(unionFlavors []*clustercrdv1.FlavorInfo, flavors []clustercrdv1.FlavorInfo) []*clustercrdv1.FlavorInfo {
	if len(unionFlavors) == 0 {
		for _, y := range flavors {
			unionFlavors = append(unionFlavors, &y)
		}
		return unionFlavors
	}

	var result []*clustercrdv1.FlavorInfo

	for _, x := range flavors {
		for idx, y := range unionFlavors {
			if reflect.DeepEqual(&x, y) {
				break
			} else if idx == len(unionFlavors)-1 {
				res := x
				result = append(result, &res)
			}
		}
	}

	for _, z := range result {
		unionFlavors = append(unionFlavors, z)
	}
	return unionFlavors
}
