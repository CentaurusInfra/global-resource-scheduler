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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	"testing"
)

func TestAddingOneCluster(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername := "cluster1"
	geoLocation := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)
	cluster := newClusterWithInfo(clustername, geoLocation, region, operator, flavors, storages, eip, cpu, mem, price)
	cin.AddCluster(cluster, 0)
	verifyNode(t, *cin, []string{"US"}, []string{clustername}, []clusterv1.GeolocationInfo{geoLocation}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername}, []clusterv1.GeolocationInfo{geoLocation}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername}, []clusterv1.GeolocationInfo{geoLocation}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername}, []clusterv1.GeolocationInfo{geoLocation}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername}, []string{clustername}, []clusterv1.GeolocationInfo{geoLocation}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername], []string{}, []string{clustername}, []clusterv1.GeolocationInfo{geoLocation}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoClustersWithDifferentCities(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Seattle", Province: "WA", Area: "PAC", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	verifyNode(t, *cin, []string{"US"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue", "Seattle"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Seattle"], []string{clustername2}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Seattle"].infoMap[clustername2], []string{}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoClustersWithDifferentProvinces(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Portland", Province: "OR", Area: "PAC", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	verifyNode(t, *cin, []string{"US"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA", "OR"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["OR"], []string{"Portland"}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["OR"].infoMap["Portland"], []string{clustername2}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["OR"].infoMap["Portland"].infoMap[clustername2], []string{}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoClustersWithDifferentAreas(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Phoenix", Province: "AZ", Area: "MT", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	verifyNode(t, *cin, []string{"US"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC", "MT"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["MT"], []string{"AZ"}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["MT"].infoMap["AZ"], []string{"Phoenix"}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["MT"].infoMap["AZ"].infoMap["Phoenix"], []string{clustername2}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["MT"].infoMap["AZ"].infoMap["Phoenix"].infoMap[clustername2], []string{}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoClustersWithDifferentCountries(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Guangzhou", Province: "GD", Area: "South", Country: "CN"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	verifyNode(t, *cin, []string{"US", "CN"}, []string{clustername1, clustername2}, []clusterv1.GeolocationInfo{geoLocation1, geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["CN"], []string{"South"}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["CN"].infoMap["South"], []string{"GD"}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["CN"].infoMap["South"].infoMap["GD"], []string{"Guangzhou"}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["CN"].infoMap["South"].infoMap["GD"].infoMap["Guangzhou"], []string{clustername2}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["CN"].infoMap["South"].infoMap["GD"].infoMap["Guangzhou"].infoMap[clustername2], []string{}, []string{clustername2}, []clusterv1.GeolocationInfo{geoLocation2}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoWithDifferntCitiesDeletingOneCluster(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Seattle", Province: "WA", Area: "PAC", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	cin.DelCluster(cluster2, 0)

	verifyNode(t, *cin, []string{"US"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoWithDifferntProvincesDeletingOneCluster(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Portland", Province: "OR", Area: "PAC", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	cin.DelCluster(cluster2, 0)

	verifyNode(t, *cin, []string{"US"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoWithDifferntAreasDeletingOneCluster(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Phoenix", Province: "AZ", Area: "MT", Country: "US"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	cin.DelCluster(cluster2, 0)

	verifyNode(t, *cin, []string{"US"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func TestAddingTwoWithDifferntCountriesDeletingOneCluster(t *testing.T) {
	cin := NewClusterInfoNode()
	clustername1 := "cluster1"
	clustername2 := "cluster2"
	geoLocation1 := clusterv1.GeolocationInfo{City: "Bellevue", Province: "WA", Area: "PAC", Country: "US"}
	geoLocation2 := clusterv1.GeolocationInfo{City: "Guangzhou", Province: "GD", Area: "South", Country: "CN"}
	region := clusterv1.RegionInfo{Region: "west", AvailabilityZone: "west1"}
	operator := clusterv1.OperatorInfo{Operator: "op"}
	flavor1 := clusterv1.FlavorInfo{FlavorID: "small", TotalCapacity: 10}
	flavor2 := clusterv1.FlavorInfo{FlavorID: "medium", TotalCapacity: 20}
	flavors := []clusterv1.FlavorInfo{flavor1, flavor2}
	storage1 := clusterv1.StorageSpec{TypeID: "sata", StorageCapacity: 2000}
	storage2 := clusterv1.StorageSpec{TypeID: "sas", StorageCapacity: 1000}
	storage3 := clusterv1.StorageSpec{TypeID: "ssd", StorageCapacity: 3000}
	storages := []clusterv1.StorageSpec{storage1, storage2, storage3}
	eip := int64(3)
	cpu := int64(1)
	mem := int64(256)
	price := int64(10)

	cluster1 := newClusterWithInfo(clustername1, geoLocation1, region, operator, flavors, storages, eip, cpu, mem, price)
	cluster2 := newClusterWithInfo(clustername2, geoLocation2, region, operator, flavors, storages, eip, cpu, mem, price)

	cin.AddCluster(cluster1, 0)
	cin.AddCluster(cluster2, 0)
	cin.DelCluster(cluster2, 0)

	verifyNode(t, *cin, []string{"US"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"], []string{"PAC"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"], []string{"WA"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"], []string{"Bellevue"}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"], []string{clustername1}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
	verifyNode(t, *cin.infoMap["US"].infoMap["PAC"].infoMap["WA"].infoMap["Bellevue"].infoMap[clustername1], []string{}, []string{clustername1}, []clusterv1.GeolocationInfo{geoLocation1}, []clusterv1.RegionInfo{region},
		[]clusterv1.OperatorInfo{operator}, flavors, storages, []int64{eip}, []int64{cpu}, []int64{mem}, []int64{price})
}

func verifyNode(t *testing.T, cin ClusterInfoNode, infoMapKeys []string, clusternames []string,
	geoLocations []clusterv1.GeolocationInfo, regions []clusterv1.RegionInfo, operators []clusterv1.OperatorInfo,
	flavors []clusterv1.FlavorInfo, storages []clusterv1.StorageSpec, eips, cpus, mems, prices []int64) {
	if len(cin.infoMap) != len(infoMapKeys) {
		t.Errorf("The info map %v is not expected", cin.infoMap)
	} else if len(infoMapKeys) > 0 {
		for _, infoMapKey := range infoMapKeys {
			if _, found := cin.infoMap[infoMapKey]; !found {
				t.Errorf("The info map %v is not expected", cin.infoMap)
			}
		}
	}

	if len(cin.clusterUnionMap.GeoLocationMap) != len(geoLocations) {
		t.Errorf("The union geoLocation map %v is not expected", cin.clusterUnionMap.GeoLocationMap)
	} else {
		for _, geoLocation := range geoLocations {
			if cin.clusterUnionMap.GeoLocationMap[geoLocation] < 1 {
				t.Errorf("The union geoLocation map %v is not expected", cin.clusterUnionMap.GeoLocationMap)
			}
		}
	}
	if len(cin.clusterUnionMap.RegionMap) != len(regions) {
		t.Errorf("The union region map %v is not expected", cin.clusterUnionMap.RegionMap)
	} else {
		for _, region := range regions {
			if cin.clusterUnionMap.RegionMap[region] < 1 {
				t.Errorf("The union region map %v is not expected", cin.clusterUnionMap.RegionMap)
			}
		}
	}
	if len(cin.clusterUnionMap.OperatorMap) != len(operators) {
		t.Errorf("The union operator map %v is not expected", cin.clusterUnionMap.OperatorMap)
	} else {
		for _, operator := range operators {
			if cin.clusterUnionMap.OperatorMap[operator] < 1 {
				t.Errorf("The union operator map %v is not expected", cin.clusterUnionMap.OperatorMap)
			}
		}
	}
	if len(cin.clusterUnionMap.FlavorMap) != len(flavors) {
		t.Errorf("The union flavor map %v is not expected", cin.clusterUnionMap.FlavorMap)
	} else {
		for _, flavor := range flavors {
			if cin.clusterUnionMap.FlavorMap[flavor] < 1 {
				t.Errorf("The union flavor map %v is not expected", cin.clusterUnionMap.FlavorMap)
			}
		}
	}
	if len(cin.clusterUnionMap.StorageMap) != len(storages) {
		t.Errorf("The union storage map %v is not expected", cin.clusterUnionMap.StorageMap)
	} else {
		for _, storage := range storages {
			if cin.clusterUnionMap.StorageMap[storage] < 1 {
				t.Errorf("The union storage map %v is not expected", cin.clusterUnionMap.StorageMap)
			}
		}
	}

	if len(cin.clusterUnionMap.EipCapacityMap) != len(eips) {
		t.Errorf("The union eip capacity map %v is not expected", cin.clusterUnionMap.EipCapacityMap)
	} else {
		for _, eip := range eips {
			if cin.clusterUnionMap.EipCapacityMap[eip] < 1 {
				t.Errorf("The union eip capacity map %v is not expected", cin.clusterUnionMap.EipCapacityMap)
			}
		}
	}

	if len(cin.clusterUnionMap.CPUCapacityMap) != len(cpus) {
		t.Errorf("The union cpu capacity map %v is not expected", cin.clusterUnionMap.CPUCapacityMap)
	} else {
		for _, cpu := range cpus {
			if cin.clusterUnionMap.CPUCapacityMap[cpu] < 1 {
				t.Errorf("The union cpu capacity map %v is not expected", cin.clusterUnionMap.CPUCapacityMap)
			}
		}
	}

	if len(cin.clusterUnionMap.MemCapacityMap) != len(mems) {
		t.Errorf("The union mem capacity map %v is not expected", cin.clusterUnionMap.MemCapacityMap)
	} else {
		for _, mem := range mems {
			if cin.clusterUnionMap.MemCapacityMap[mem] < 1 {
				t.Errorf("The union mem capacity map %v is not expected", cin.clusterUnionMap.MemCapacityMap)
			}
		}
	}

	if len(cin.clusterUnionMap.ServerPriceMap) != len(prices) {
		t.Errorf("The union server price map %v is not expected", cin.clusterUnionMap.ServerPriceMap)
	} else {
		for _, price := range prices {
			if cin.clusterUnionMap.ServerPriceMap[price] < 1 {
				t.Errorf("The union server price map %v is not expected", cin.clusterUnionMap.ServerPriceMap)
			}
		}
	}

	if len(cin.clusterMap) != len(clusternames) {
		t.Errorf("The union server price map %v is not expected", cin.clusterMap)
	} else {
		for _, clustername := range clusternames {
			if cin.clusterMap[clustername] < 1 {
				t.Errorf("The union server price map %v is not expected", cin.clusterMap)
			}
		}
	}
}

func newClusterWithInfo(name string, geoLocation clusterv1.GeolocationInfo, region clusterv1.RegionInfo, operator clusterv1.OperatorInfo,
	flavors []clusterv1.FlavorInfo, storages []clusterv1.StorageSpec, eip, cpu, mem, price int64) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{Kind: clusterv1.Kind, APIVersion: clusterv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: clusterv1.ClusterSpec{
			IpAddress:     "10.0.0.1",
			GeoLocation:   geoLocation,
			Region:        region,
			Operator:      operator,
			Flavors:       flavors,
			Storage:       storages,
			EipCapacity:   eip,
			CPUCapacity:   cpu,
			MemCapacity:   mem,
			ServerPrice:   price,
			HomeScheduler: "scheduler1",
		},
	}
}
