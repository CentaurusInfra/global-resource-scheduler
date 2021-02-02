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

package geoLocationDistribute

import (
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
)

type GeoLocationDistribute struct {
	CountryMap     map[string][]string
	AreaMap        map[string][]string
	ProvinceMap    map[string][]string
	CityMap        map[string][]string
	SchedulersName []string
	ClustersName   []string
}

func New() *GeoLocationDistribute {
	gld := new(GeoLocationDistribute)
	gld.CountryMap = make(map[string][]string)
	gld.AreaMap = make(map[string][]string)
	gld.ProvinceMap = make(map[string][]string)
	gld.CityMap = make(map[string][]string)
	return gld
}

func RemoveClusterName(cluster *clustercrdv1.Cluster, gld *GeoLocationDistribute) *GeoLocationDistribute {
	gld.ClustersName = clustersNameRemove(cluster.Name, gld.ClustersName)
	gld.CountryMap = countryMapRemove(cluster, gld.CountryMap)
	gld.AreaMap = areaMapRemove(cluster, gld.AreaMap)
	gld.ProvinceMap = provinceMapRemove(cluster, gld.ProvinceMap)
	gld.CityMap = cityMap(cluster, gld.CityMap)

	return gld
}

func RemoveSchedulerName(scheduler *schedulercrdv1.Scheduler, gld *GeoLocationDistribute) *GeoLocationDistribute {
	gld.SchedulersName = schedulersNameRemove(scheduler.Name, gld.SchedulersName)
	return gld
}

func schedulersNameRemove(name string, schedulersName []string) []string {
	curMap := getMap(schedulersName)
	removeIdx := curMap[name]
	res := remove(schedulersName, removeIdx)
	return res
}

func cityMap(cluster *clustercrdv1.Cluster, cityMap map[string][]string) map[string][]string {
	clusterCity := cluster.Spec.GeoLocation.City
	clustersName := cityMap[clusterCity]
	curMap := getMap(clustersName)
	removeIdx := curMap[cluster.Name]
	cityMap[clusterCity] = remove(clustersName, removeIdx)
	return cityMap
}

func provinceMapRemove(cluster *clustercrdv1.Cluster, provinceMap map[string][]string) map[string][]string {
	clusterProvince := cluster.Spec.GeoLocation.Province
	clustersName := provinceMap[clusterProvince]
	curMap := getMap(clustersName)
	removeIdx := curMap[cluster.Name]
	provinceMap[clusterProvince] = remove(clustersName, removeIdx)
	return provinceMap
}

func areaMapRemove(cluster *clustercrdv1.Cluster, areaMap map[string][]string) map[string][]string {
	clusterArea := cluster.Spec.GeoLocation.Area
	clustersName := areaMap[clusterArea]
	curMap := getMap(clustersName)
	removeIdx := curMap[cluster.Name]
	areaMap[clusterArea] = remove(clustersName, removeIdx)
	return areaMap
}

func countryMapRemove(cluster *clustercrdv1.Cluster, countryMap map[string][]string) map[string][]string {
	clusterCountry := cluster.Spec.GeoLocation.Country
	clustersName := countryMap[clusterCountry]
	curMap := getMap(clustersName)
	removeIdx := curMap[cluster.Name]
	countryMap[clusterCountry] = remove(clustersName, removeIdx)
	return countryMap
}

func clustersNameRemove(name string, clustersName []string) []string {
	curMap := getMap(clustersName)
	removeIdx := curMap[name]
	res := remove(clustersName, removeIdx)
	return res
}

func getMap(names []string) map[string]int {
	mapRes := make(map[string]int)
	for idx, val := range names {
		mapRes[val] = idx
	}
	return mapRes
}

func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
