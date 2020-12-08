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

func UpdateUnion(schedulerUnion schedulercrdv1.ClusterUnion, cluster *clustercrdv1.Cluster) schedulercrdv1.ClusterUnion {
	// IpAddress Union
	schedulerUnion.IpAddress = unionIpAddress(schedulerUnion.IpAddress, cluster.Spec.IpAddress)

	// GeoLocation Union
	schedulerUnion.GeoLocation = unionGeoLocation(schedulerUnion.GeoLocation, cluster.Spec.GeoLocation)

	// Region Union
	schedulerUnion.Region = unionRegion(schedulerUnion.Region, cluster.Spec.Region)

	// Operator Union
	schedulerUnion.Operator = unionOperator(schedulerUnion.Operator, cluster.Spec.Operator)

	//// Flavors Union
	//schedulerUnion.Flavors = unionFlavors(schedulerUnion.Flavors, cluster.Spec.Flavors)
	//
	//// Storage Union
	//schedulerUnion.Storage = unionStorage(schedulerUnion.Storage, cluster.Spec.Storage)

	// EipCapacity Union
	schedulerUnion.EipCapacity = unionEipCapacity(schedulerUnion.EipCapacity, cluster.Spec.EipCapacity)

	// CPUCapacity Union
	schedulerUnion.CPUCapacity = unionCPUCapacity(schedulerUnion.CPUCapacity, cluster.Spec.CPUCapacity)

	// MemCapacity Union
	schedulerUnion.MemCapacity = unionMemCapacity(schedulerUnion.MemCapacity, cluster.Spec.MemCapacity)

	return schedulerUnion
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
	m := make([]*clustercrdv1.OperatorInfo, len(unionOperator)+1)
	for i, v := range unionOperator {
		m[i] = v
	}

	for i, v := range m {
		if v == nil {
			m[i] = &operator
			return m
		}
		if reflect.DeepEqual(v, &operator) {
			return unionOperator
		}
	}
	m = append(m, &operator)

	return m
}

func unionRegion(unionRegion []*clustercrdv1.RegionInfo, region clustercrdv1.RegionInfo) []*clustercrdv1.RegionInfo {
	m := make([]*clustercrdv1.RegionInfo, len(unionRegion)+1)
	for i, v := range unionRegion {
		m[i] = v
	}

	for i, v := range m {
		if v == nil {
			m[i] = &region
			return m
		}
		if reflect.DeepEqual(v, &region) {
			return unionRegion
		}
	}
	m = append(m, &region)

	return m
}

func unionGeoLocation(unionGeoLocation []*clustercrdv1.GeolocationInfo, geoLocation clustercrdv1.GeolocationInfo) []*clustercrdv1.GeolocationInfo {
	m := make([]*clustercrdv1.GeolocationInfo, len(unionGeoLocation)+1)
	for i, v := range unionGeoLocation {
		m[i] = v
	}

	for i, v := range m {
		if v == nil {
			m[i] = &geoLocation
			return m
		}
		if reflect.DeepEqual(v, &geoLocation) {
			return unionGeoLocation
		}
	}
	m = append(m, &geoLocation)
	return m
}

func unionIpAddress(unionIp []string, ip string) []string {
	m := make(map[string]int)
	for _, v := range unionIp {
		m[v]++
	}

	times, _ := m[ip]
	if times == 0 {
		unionIp = append(unionIp, ip)
	}

	return unionIp
}

//
//func unionStorage(unionStorage []*clustercrdv1.StorageSpec, storage []clustercrdv1.StorageSpec) []*clustercrdv1.StorageSpec {
//	m := make(map[*clustercrdv1.StorageSpec]int)
//	for _, v := range unionStorage {
//		m[v]++
//	}
//
//	for _, v := range storage {
//		times, _ := m[&v]
//		if times == 0 {
//			unionStorage = append(unionStorage, &v)
//		}
//	}
//	return unionStorage
//}
//
//func unionFlavors(unionFlavors []*clustercrdv1.FlavorInfo, flavors []clustercrdv1.FlavorInfo) []*clustercrdv1.FlavorInfo {
//	m := make(map[*clustercrdv1.FlavorInfo]int)
//	for _, v := range unionFlavors {
//		m[v]++
//	}
//
//	for _, v := range flavors {
//		times, _ := m[&v]
//		if times == 0 {
//			unionFlavors = append(unionFlavors, &v)
//		}
//	}
//	return unionFlavors
//}
