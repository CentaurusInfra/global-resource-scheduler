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
)

func UpdateUnion(schedulerCopy *schedulercrdv1.Scheduler, cluster *clustercrdv1.Cluster) *schedulercrdv1.Scheduler {
	union := schedulerCopy.Spec.Union

	// IpAddress Union
	union.IpAddress = unionIpAddress(union.IpAddress, cluster.Spec.IpAddress)

	// GeoLocation Union
	union.GeoLocation = unionGeoLocation(union.GeoLocation, cluster.Spec.GeoLocation)

	// Region Union
	union.Region = unionRegion(union.Region, cluster.Spec.Region)

	// Operator Union
	union.Operator = unionOperator(union.Operator, cluster.Spec.Operator)

	// Flavors Union
	union.Flavors = unionFlavors(union.Flavors, cluster.Spec.Flavors)

	// Storage Union
	union.Storage = unionStorage(union.Storage, cluster.Spec.Storage)

	// EipCapacity Union
	union.EipCapacity = unionEipCapacity(union.EipCapacity, cluster.Spec.EipCapacity)

	// CPUCapacity Union
	union.CPUCapacity = unionCPUCapacity(union.CPUCapacity, cluster.Spec.CPUCapacity)

	// MemCapacity Union
	union.MemCapacity = unionMemCapacity(union.MemCapacity, cluster.Spec.MemCapacity)

	schedulerCopy.Spec.Union = union
	return schedulerCopy
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
	m := make(map[*clustercrdv1.OperatorInfo]int)
	for _, v := range unionOperator {
		m[v]++
	}

	times, _ := m[&operator]
	if times == 0 {
		unionOperator = append(unionOperator, &operator)
	}

	return unionOperator
}

func unionRegion(unionRegion []*clustercrdv1.RegionInfo, region clustercrdv1.RegionInfo) []*clustercrdv1.RegionInfo {
	m := make(map[*clustercrdv1.RegionInfo]int)
	for _, v := range unionRegion {
		m[v]++
	}

	times, _ := m[&region]
	if times == 0 {
		unionRegion = append(unionRegion, &region)
	}

	return unionRegion
}

func unionGeoLocation(unionGeoLocation []*clustercrdv1.GeolocationInfo, geoLocation clustercrdv1.GeolocationInfo) []*clustercrdv1.GeolocationInfo {
	m := make(map[*clustercrdv1.GeolocationInfo]int)
	for _, v := range unionGeoLocation {
		m[v]++
	}

	times, _ := m[&geoLocation]
	if times == 0 {
		unionGeoLocation = append(unionGeoLocation, &geoLocation)
	}

	return unionGeoLocation
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

func unionStorage(unionStorage []*clustercrdv1.StorageSpec, storage []clustercrdv1.StorageSpec) []*clustercrdv1.StorageSpec {
	m := make(map[*clustercrdv1.StorageSpec]int)
	for _, v := range unionStorage {
		m[v]++
	}

	for _, v := range storage {
		times, _ := m[&v]
		if times == 0 {
			unionStorage = append(unionStorage, &v)
		}
	}
	return unionStorage
}

func unionFlavors(unionFlavors []*clustercrdv1.FlavorInfo, flavors []clustercrdv1.FlavorInfo) []*clustercrdv1.FlavorInfo {
	m := make(map[*clustercrdv1.FlavorInfo]int)
	for _, v := range unionFlavors {
		m[v]++
	}

	for _, v := range flavors {
		times, _ := m[&v]
		if times == 0 {
			unionFlavors = append(unionFlavors, &v)
		}
	}
	return unionFlavors
}
