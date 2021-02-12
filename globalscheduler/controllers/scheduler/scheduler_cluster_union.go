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

package scheduler

import (
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

type ClusterInfoNode struct {
	infoMap         map[string]*ClusterInfoNode
	clusterUnionMap CLusterUnionMap
	clusterMap      map[interface{}]int
}

type CLusterUnionMap struct {
	GeoLocationMap map[interface{}]int
	RegionMap      map[interface{}]int
	OperatorMap    map[interface{}]int
	FlavorMap      map[interface{}]int
	StorageMap     map[interface{}]int
	EipCapacityMap map[interface{}]int
	CPUCapacityMap map[interface{}]int
	MemCapacityMap map[interface{}]int
	ServerPriceMap map[interface{}]int
}

func NewClusterInfoNode() *ClusterInfoNode {
	return &ClusterInfoNode{
		infoMap: make(map[string]*ClusterInfoNode, 0),
		clusterUnionMap: CLusterUnionMap{
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
			make(map[interface{}]int, 0),
		},
		clusterMap: make(map[interface{}]int, 0),
	}
}

func (cin *ClusterInfoNode) AddCluster(cluster *clusterv1.Cluster, geoLevel int) {
	key := getGeoLocationKey(cluster, geoLevel)
	addMap(cin.clusterUnionMap.GeoLocationMap, cluster.Spec.GeoLocation)
	addMap(cin.clusterUnionMap.RegionMap, cluster.Spec.Region)
	addMap(cin.clusterUnionMap.OperatorMap, cluster.Spec.Operator)

	for _, flavor := range cluster.Spec.Flavors {
		addMap(cin.clusterUnionMap.FlavorMap, flavor)
	}
	for _, storage := range cluster.Spec.Storage {
		addMap(cin.clusterUnionMap.StorageMap, storage)
	}
	addMap(cin.clusterUnionMap.EipCapacityMap, cluster.Spec.EipCapacity)
	addMap(cin.clusterUnionMap.CPUCapacityMap, cluster.Spec.CPUCapacity)
	addMap(cin.clusterUnionMap.MemCapacityMap, cluster.Spec.MemCapacity)
	addMap(cin.clusterUnionMap.ServerPriceMap, cluster.Spec.ServerPrice)

	addMap(cin.clusterMap, cluster.ObjectMeta.Name)

	if len(key) != 0 {
		_, found := cin.infoMap[key]
		if !found {
			cin.infoMap[key] = NewClusterInfoNode()
		}
		cin.infoMap[key].AddCluster(cluster, geoLevel+1)
	}
}

func (cin *ClusterInfoNode) DelCluster(cluster *clusterv1.Cluster, geoLevel int) {
	key := getGeoLocationKey(cluster, geoLevel)
	if key == "" {
		return
	}

	_, found := cin.infoMap[key]
	if !found {
		return
	}

	delMap(cin.clusterUnionMap.GeoLocationMap, cluster.Spec.GeoLocation)
	delMap(cin.clusterUnionMap.RegionMap, cluster.Spec.Region)
	delMap(cin.clusterUnionMap.OperatorMap, cluster.Spec.Operator)

	for _, flavor := range cluster.Spec.Flavors {
		delMap(cin.clusterUnionMap.FlavorMap, flavor)
	}
	for _, storage := range cluster.Spec.Storage {
		delMap(cin.clusterUnionMap.StorageMap, storage)
	}
	delMap(cin.clusterUnionMap.EipCapacityMap, cluster.Spec.EipCapacity)
	delMap(cin.clusterUnionMap.CPUCapacityMap, cluster.Spec.CPUCapacity)
	delMap(cin.clusterUnionMap.MemCapacityMap, cluster.Spec.MemCapacity)
	delMap(cin.clusterUnionMap.ServerPriceMap, cluster.Spec.ServerPrice)

	delMap(cin.clusterMap, cluster.ObjectMeta.Name)

	cin.infoMap[key].DelCluster(cluster, geoLevel+1)
	if len(cin.infoMap[key].infoMap) == 0 {
		delete(cin.infoMap, key)
	}
}

func getGeoLocationKey(cluster *clusterv1.Cluster, geoLevel int) string {
	key := ""
	switch geoLevel {
	case 0:
		key = cluster.Spec.GeoLocation.Country
	case 1:
		key = cluster.Spec.GeoLocation.Area
	case 2:
		key = cluster.Spec.GeoLocation.Province
	case 3:
		key = cluster.Spec.GeoLocation.City
	case 4:
		key = cluster.ObjectMeta.Name
	}
	return key
}

func addMap(m map[interface{}]int, key interface{}) {
	if existedVal, ok := m[key]; ok {
		m[key] = existedVal + 1
	} else {
		m[key] = 1
	}
}

func delMap(m map[interface{}]int, key interface{}) {
	if existedVal, ok := m[key]; ok {
		if existedVal <= 1 {
			delete(m, key)
		} else if existedVal > 1 {
			m[key] = existedVal - 1
		}
	}
}
