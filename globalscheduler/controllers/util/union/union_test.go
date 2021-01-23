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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
	"reflect"
	"testing"
)

func TestUpdateUnionWithDifferentUnion(t *testing.T) {
	fakeSchedulerWithUnion := &schedulercrdv1.Scheduler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-scheduler",
		},
		Spec: schedulercrdv1.SchedulerSpec{
			Union: schedulercrdv1.ClusterUnion{
				GeoLocation: []*clustercrdv1.GeolocationInfo{
					{
						City: "NewYork",
					},
				},
				Region: []*clustercrdv1.RegionInfo{
					{
						Region:           "LA",
						AvailabilityZone: "C",
					},
				},
				Operator: []*clustercrdv1.OperatorInfo{
					{
						Operator: "operator-1",
					},
					{
						Operator: "operator-2",
					},
				},
				Flavors: []*clustercrdv1.FlavorInfo{
					{
						FlavorID:      "flavor-id-1",
						TotalCapacity: 100,
					},
					{
						FlavorID:      "flavor-id-2",
						TotalCapacity: 200,
					},
				},
				Storage: []*clustercrdv1.StorageSpec{
					{
						TypeID:          "storage-id-1",
						StorageCapacity: 100,
					},
					{
						TypeID:          "storage-id-2",
						StorageCapacity: 200,
					},
					{
						TypeID:          "storage-id-3",
						StorageCapacity: 300,
					},
					{
						TypeID:          "storage-id-4",
						StorageCapacity: 400,
					},
					{
						TypeID:          "storage-id-5",
						StorageCapacity: 500,
					},
				},
				EipCapacity: []int64{100},
				CPUCapacity: []int64{100},
				MemCapacity: []int64{100},
			},
		},
	}

	fakeCluster := &clustercrdv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-cluster",
		},
		Spec: clustercrdv1.ClusterSpec{
			IpAddress: "127.0.0.1",
			GeoLocation: clustercrdv1.GeolocationInfo{
				City: "Bellevue",
			},
			Region: clustercrdv1.RegionInfo{
				Region:           "Oregon",
				AvailabilityZone: "A",
			},
			Operator: clustercrdv1.OperatorInfo{
				Operator: "operator-2",
			},
			Flavors: []clustercrdv1.FlavorInfo{
				{
					FlavorID:      "flavor-id-1",
					TotalCapacity: 100,
				},
				{
					FlavorID:      "flavor-id-2",
					TotalCapacity: 200,
				},
				{
					FlavorID:      "flavor-id-3",
					TotalCapacity: 300,
				},
				{
					FlavorID:      "flavor-id-4",
					TotalCapacity: 300,
				},
				{
					FlavorID:      "flavor-id-5",
					TotalCapacity: 300,
				},
			},
			Storage: []clustercrdv1.StorageSpec{
				{
					TypeID:          "storage-id-2",
					StorageCapacity: 200,
				},
				{
					TypeID:          "storage-id-3",
					StorageCapacity: 300,
				},
				{
					TypeID:          "storage-id-4",
					StorageCapacity: 400,
				},
				{
					TypeID:          "storage-id-6",
					StorageCapacity: 600,
				},
			},
			EipCapacity: 50,
			CPUCapacity: 50,
			MemCapacity: 50,
		},
	}

	expectedRes := schedulercrdv1.ClusterUnion{
		GeoLocation: []*clustercrdv1.GeolocationInfo{
			{
				City: "NewYork",
			},
			{
				City: "Bellevue",
			},
		},
		Region: []*clustercrdv1.RegionInfo{
			{
				Region:           "LA",
				AvailabilityZone: "C",
			},
			{
				Region:           "Oregon",
				AvailabilityZone: "A",
			},
		},
		Operator: []*clustercrdv1.OperatorInfo{
			{
				Operator: "operator-1",
			},
			{
				Operator: "operator-2",
			},
		},
		Flavors: []*clustercrdv1.FlavorInfo{
			{
				FlavorID:      "flavor-id-1",
				TotalCapacity: 100,
			},
			{
				FlavorID:      "flavor-id-2",
				TotalCapacity: 200,
			},
			{
				FlavorID:      "flavor-id-3",
				TotalCapacity: 300,
			},
			{
				FlavorID:      "flavor-id-4",
				TotalCapacity: 300,
			},
			{
				FlavorID:      "flavor-id-5",
				TotalCapacity: 300,
			},
		},
		Storage: []*clustercrdv1.StorageSpec{
			{
				TypeID:          "storage-id-1",
				StorageCapacity: 100,
			},
			{
				TypeID:          "storage-id-2",
				StorageCapacity: 200,
			},
			{
				TypeID:          "storage-id-3",
				StorageCapacity: 300,
			},
			{
				TypeID:          "storage-id-4",
				StorageCapacity: 400,
			},
			{
				TypeID:          "storage-id-5",
				StorageCapacity: 500,
			},
			{
				TypeID:          "storage-id-6",
				StorageCapacity: 600,
			},
		},
		EipCapacity: []int64{100, 50},
		CPUCapacity: []int64{100, 50},
		MemCapacity: []int64{100, 50},
	}

	fakeSchedulerWithUnion.Spec.Union = UpdateUnion(fakeSchedulerWithUnion.Spec.Union, fakeCluster)
	unionRes := fakeSchedulerWithUnion.Spec.Union

	if ifEmpty(unionRes) {
		t.Fatalf("Scheduler Union is Empty: %v", unionRes)
	} else if !unionEqual(unionRes, expectedRes) {
		t.Fatalf("Scheduler Union results are different")
	}
}

func TestUpdateUnionWithDefaultUnion(t *testing.T) {
	fakeSchedulerWithUnion := &schedulercrdv1.Scheduler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-scheduler",
		},
		Spec: schedulercrdv1.SchedulerSpec{
			Union: schedulercrdv1.ClusterUnion{
				GeoLocation: []*clustercrdv1.GeolocationInfo{
					{
						City: "Bellevue",
					},
				},
				Region: []*clustercrdv1.RegionInfo{
					{
						Region:           "Oregon",
						AvailabilityZone: "A",
					},
				},
				Operator: []*clustercrdv1.OperatorInfo{
					{
						Operator: "operator",
					},
				},
				Flavors: []*clustercrdv1.FlavorInfo{
					{
						FlavorID:      "flavor-id-1",
						TotalCapacity: 100,
					},
				},
				Storage: []*clustercrdv1.StorageSpec{
					{
						TypeID:          "storage-id-1",
						StorageCapacity: 100,
					},
				},
				EipCapacity: []int64{100},
				CPUCapacity: []int64{100},
				MemCapacity: []int64{100},
			},
		},
	}

	fakeCluster := &clustercrdv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-cluster",
		},
		Spec: clustercrdv1.ClusterSpec{
			IpAddress: "127.0.0.1",
			GeoLocation: clustercrdv1.GeolocationInfo{
				City: "Bellevue",
			},
			Region: clustercrdv1.RegionInfo{
				Region:           "Oregon",
				AvailabilityZone: "A",
			},
			Operator: clustercrdv1.OperatorInfo{
				Operator: "operator",
			},
			Flavors: []clustercrdv1.FlavorInfo{
				{
					FlavorID:      "flavor-id-1",
					TotalCapacity: 100,
				},
			},
			Storage: []clustercrdv1.StorageSpec{
				{
					TypeID:          "storage-id-1",
					StorageCapacity: 100,
				},
			},
			EipCapacity: 100,
			CPUCapacity: 100,
			MemCapacity: 100,
		},
	}

	expectedRes := schedulercrdv1.ClusterUnion{
		GeoLocation: []*clustercrdv1.GeolocationInfo{
			{
				City: "Bellevue",
			},
		},
		Region: []*clustercrdv1.RegionInfo{
			{
				Region:           "Oregon",
				AvailabilityZone: "A",
			},
		},
		Operator: []*clustercrdv1.OperatorInfo{
			{
				Operator: "operator",
			},
		},
		Flavors: []*clustercrdv1.FlavorInfo{
			{
				FlavorID:      "flavor-id-1",
				TotalCapacity: 100,
			},
		},
		Storage: []*clustercrdv1.StorageSpec{
			{
				TypeID:          "storage-id-1",
				StorageCapacity: 100,
			},
		},
		EipCapacity: []int64{100},
		CPUCapacity: []int64{100},
		MemCapacity: []int64{100},
	}

	fakeSchedulerWithUnion.Spec.Union = UpdateUnion(fakeSchedulerWithUnion.Spec.Union, fakeCluster)
	unionRes := fakeSchedulerWithUnion.Spec.Union

	if ifEmpty(unionRes) {
		t.Fatalf("Scheduler Union is Empty: %v", unionRes)
	} else if !unionEqual(unionRes, expectedRes) {
		t.Fatalf("Scheduler Union results are different")
	}
}

func TestUpdateUnionWithoutDefaultUnion(t *testing.T) {
	fakeSchedulerWithoutUnion := &schedulercrdv1.Scheduler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-scheduler",
		},
	}

	fakeCluster := &clustercrdv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-cluster",
		},
		Spec: clustercrdv1.ClusterSpec{
			IpAddress: "127.0.0.1",
			GeoLocation: clustercrdv1.GeolocationInfo{
				City: "Bellevue",
			},
			Region: clustercrdv1.RegionInfo{
				Region:           "Oregon",
				AvailabilityZone: "A",
			},
			Operator: clustercrdv1.OperatorInfo{
				Operator: "operator",
			},
			Flavors: []clustercrdv1.FlavorInfo{
				{
					FlavorID:      "flavor-id-1",
					TotalCapacity: 100,
				},
			},
			Storage: []clustercrdv1.StorageSpec{
				{
					TypeID:          "storage-id-1",
					StorageCapacity: 100,
				},
			},
			EipCapacity: 100,
			CPUCapacity: 100,
			MemCapacity: 100,
		},
	}

	expectedRes := schedulercrdv1.ClusterUnion{
		GeoLocation: []*clustercrdv1.GeolocationInfo{
			{
				City: "Bellevue",
			},
		},
		Region: []*clustercrdv1.RegionInfo{
			{
				Region:           "Oregon",
				AvailabilityZone: "A",
			},
		},
		Operator: []*clustercrdv1.OperatorInfo{
			{
				Operator: "operator",
			},
		},
		Flavors: []*clustercrdv1.FlavorInfo{
			{
				FlavorID:      "flavor-id-1",
				TotalCapacity: 100,
			},
		},
		Storage: []*clustercrdv1.StorageSpec{
			{
				TypeID:          "storage-id-1",
				StorageCapacity: 100,
			},
		},
		EipCapacity: []int64{100},
		CPUCapacity: []int64{100},
		MemCapacity: []int64{100},
	}

	fakeSchedulerWithoutUnion.Spec.Union = UpdateUnion(fakeSchedulerWithoutUnion.Spec.Union, fakeCluster)
	unionRes := fakeSchedulerWithoutUnion.Spec.Union
	if ifEmpty(unionRes) {
		t.Fatalf("Scheduler Union is Empty: %v", unionRes)
	} else if !unionEqual(unionRes, expectedRes) {
		t.Fatalf("Scheduler Union results are different")
	}
}

func unionEqual(unionRes schedulercrdv1.ClusterUnion, expectedRes schedulercrdv1.ClusterUnion) bool {
	if !reflect.DeepEqual(unionRes.GeoLocation, expectedRes.GeoLocation) {
		fmt.Println("Different GeoLocation")
		return false
	}
	if !reflect.DeepEqual(unionRes.Region, expectedRes.Region) {
		fmt.Println("Different Region")
		return false
	}
	if !reflect.DeepEqual(unionRes.Operator, expectedRes.Operator) {
		fmt.Println("Different Operator")
		return false
	}
	if !reflect.DeepEqual(unionRes.Flavors, expectedRes.Flavors) {
		fmt.Println("Different Flavors")
		return false
	}
	if !reflect.DeepEqual(unionRes.Storage, expectedRes.Storage) {
		fmt.Println("Different Storage")
		return false
	}
	if !reflect.DeepEqual(unionRes.EipCapacity, expectedRes.EipCapacity) {
		fmt.Println("Different EipCapacity")
		return false
	}
	if !reflect.DeepEqual(unionRes.CPUCapacity, expectedRes.CPUCapacity) {
		fmt.Println("Different CPUCapacity")
		return false
	}
	if !reflect.DeepEqual(unionRes.MemCapacity, expectedRes.MemCapacity) {
		fmt.Println("Different MemCapacity")
		return false
	}

	return true
}

func ifEmpty(res schedulercrdv1.ClusterUnion) bool {
	return reflect.DeepEqual(res, schedulercrdv1.ClusterUnion{})
}
