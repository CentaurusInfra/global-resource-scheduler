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

func RemoveCluster(currentClusterArray []*clustercrdv1.Cluster, cluster *clustercrdv1.Cluster) []*clustercrdv1.Cluster {
	m := make(map[*clustercrdv1.Cluster]bool)

	for _, v := range currentClusterArray {
		m[v] = true
	}

	if _, ok := m[cluster]; ok {
		delete(m, cluster)
	}

	var newClusterArray []*clustercrdv1.Cluster
	for v := range m {
		newClusterArray = append(newClusterArray, v)
	}

	return newClusterArray
}

func DeleteFromUnion(schedulerCopy *schedulercrdv1.Scheduler, cluster *clustercrdv1.Cluster) *schedulercrdv1.Scheduler {
	union := schedulerCopy.Spec.Union

	clusterArray := schedulerCopy.Spec.Cluster
	if !checkIpAddressExist(clusterArray, cluster.Spec.IpAddress) {
		// IpAddress Delete
		union.IpAddress = deleteIpAddress(union.IpAddress, cluster.Spec.IpAddress)
	}

	if !checkGeoLocationExist(clusterArray, cluster.Spec.GeoLocation) {
		// GeoLocation Delete
		union.GeoLocation = deleteGeoLocation(union.GeoLocation, cluster.Spec.GeoLocation)
	}

	if !checkRegionExist(clusterArray, cluster.Spec.Region) {
		// Region Delete
		union.Region = deleteRegion(union.Region, cluster.Spec.Region)
	}

	if !checkOperatorExist(clusterArray, cluster.Spec.Operator) {
		// Operator Delete
		union.Operator = deleteOperator(union.Operator, cluster.Spec.Operator)
	}

	if !checkEipCapacityExist(clusterArray, cluster.Spec.EipCapacity) {
		// EipCapacity Delete
		union.EipCapacity = deleteEipCapacity(union.EipCapacity, cluster.Spec.EipCapacity)
	}

	if !checkCPUCapacityExist(clusterArray, cluster.Spec.CPUCapacity) {
		// CPUCapacity Delete
		union.CPUCapacity = deleteCPUCapacity(union.CPUCapacity, cluster.Spec.CPUCapacity)
	}

	if !checkMemCapacityExist(clusterArray, cluster.Spec.MemCapacity) {
		// MemCapacity Delete
		union.MemCapacity = deleteMemCapacity(union.MemCapacity, cluster.Spec.MemCapacity)
	}

	schedulerCopy.Spec.Union = union
	return schedulerCopy
}

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

func deleteMemCapacity(unionMemCapacity []int64, memCapacity int64) []int64 {
	m := make(map[int64]bool)

	for _, v := range unionMemCapacity {
		m[v] = true
	}

	if _, ok := m[memCapacity]; ok {
		delete(m, memCapacity)
	}

	var newUnionMemCapacity []int64
	for v := range m {
		newUnionMemCapacity = append(newUnionMemCapacity, v)
	}

	return newUnionMemCapacity
}

func checkMemCapacityExist(clusterArray []*clustercrdv1.Cluster, memCapacity int64) bool {
	for _, v := range clusterArray {
		if v.Spec.MemCapacity == memCapacity {
			return true
		}
	}

	return false
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

func deleteCPUCapacity(unionCpuCapacity []int64, cpuCapacity int64) []int64 {
	m := make(map[int64]bool)

	for _, v := range unionCpuCapacity {
		m[v] = true
	}

	if _, ok := m[cpuCapacity]; ok {
		delete(m, cpuCapacity)
	}

	var newUnionCPUCapacity []int64
	for v := range m {
		newUnionCPUCapacity = append(newUnionCPUCapacity, v)
	}

	return newUnionCPUCapacity
}

func checkCPUCapacityExist(clusterArray []*clustercrdv1.Cluster, cpuCapacity int64) bool {
	for _, v := range clusterArray {
		if v.Spec.CPUCapacity == cpuCapacity {
			return true
		}
	}

	return false
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

func deleteEipCapacity(unionCapacity []int64, capacity int64) []int64 {
	m := make(map[int64]bool)

	for _, v := range unionCapacity {
		m[v] = true
	}

	if _, ok := m[capacity]; ok {
		delete(m, capacity)
	}

	var newUnionEipCapacity []int64
	for v := range m {
		newUnionEipCapacity = append(newUnionEipCapacity, v)
	}

	return newUnionEipCapacity
}

func checkEipCapacityExist(clusterArray []*clustercrdv1.Cluster, eipCapacity int64) bool {
	for _, v := range clusterArray {
		if v.Spec.EipCapacity == eipCapacity {
			return true
		}
	}

	return false
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

func deleteOperator(unionOperator []*clustercrdv1.OperatorInfo, operator clustercrdv1.OperatorInfo) []*clustercrdv1.OperatorInfo {
	m := make(map[*clustercrdv1.OperatorInfo]bool)

	for _, v := range unionOperator {
		m[v] = true
	}

	if _, ok := m[&operator]; ok {
		delete(m, &operator)
	}

	var newUnionOperator []*clustercrdv1.OperatorInfo
	for v := range m {
		newUnionOperator = append(newUnionOperator, v)
	}

	return newUnionOperator
}

func checkOperatorExist(clusterArray []*clustercrdv1.Cluster, operator clustercrdv1.OperatorInfo) bool {
	for _, v := range clusterArray {
		if reflect.DeepEqual(v.Spec.Operator, operator) {
			return true
		}
	}

	return false
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

func deleteRegion(unionRegion []*clustercrdv1.RegionInfo, region clustercrdv1.RegionInfo) []*clustercrdv1.RegionInfo {
	m := make(map[*clustercrdv1.RegionInfo]bool)

	for _, v := range unionRegion {
		m[v] = true
	}

	if _, ok := m[&region]; ok {
		delete(m, &region)
	}

	var newUnionRegion []*clustercrdv1.RegionInfo
	for v := range m {
		newUnionRegion = append(newUnionRegion, v)
	}

	return newUnionRegion
}

func checkRegionExist(clusterArray []*clustercrdv1.Cluster, region clustercrdv1.RegionInfo) bool {
	for _, v := range clusterArray {
		if reflect.DeepEqual(v.Spec.Region, region) {
			return true
		}
	}

	return false
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

func deleteGeoLocation(unionGeolocation []*clustercrdv1.GeolocationInfo, geoLocation clustercrdv1.GeolocationInfo) []*clustercrdv1.GeolocationInfo {
	m := make(map[*clustercrdv1.GeolocationInfo]bool)

	for _, v := range unionGeolocation {
		m[v] = true
	}

	if _, ok := m[&geoLocation]; ok {
		delete(m, &geoLocation)
	}

	var newUnionGeolocation []*clustercrdv1.GeolocationInfo
	for v := range m {
		newUnionGeolocation = append(newUnionGeolocation, v)
	}

	return newUnionGeolocation
}

func checkGeoLocationExist(clusterArray []*clustercrdv1.Cluster, geoLocation clustercrdv1.GeolocationInfo) bool {
	for _, v := range clusterArray {
		if reflect.DeepEqual(v.Spec.GeoLocation, geoLocation) {
			return true
		}
	}

	return false
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

func deleteIpAddress(unionIpAddress []string, ipAddress string) []string {
	m := make(map[string]bool)

	for _, v := range unionIpAddress {
		m[v] = true
	}

	if _, ok := m[ipAddress]; ok {
		delete(m, ipAddress)
	}

	var newUnionIpAddress []string
	for v := range m {
		newUnionIpAddress = append(newUnionIpAddress, v)
	}

	return newUnionIpAddress
}

func checkIpAddressExist(clusterArray []*clustercrdv1.Cluster, ipAddress string) bool {
	for _, v := range clusterArray {
		if v.Spec.IpAddress == ipAddress {
			return true
		}
	}

	return false
}
