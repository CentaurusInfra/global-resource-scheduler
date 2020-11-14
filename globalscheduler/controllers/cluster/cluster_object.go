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

package cluster

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

// Create a cluster object.
func (c *ClusterController) CreateObject() error {
	object := &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster1",
			Namespace: corev1.NamespaceDefault,
		},
		Spec: clusterv1.ClusterSpec{
			IpAddress: "10.0.0.1",
			GeoLocation: clusterv1.GeolocationInfo{
				City:     "Bellevue",
				Province: "Washington",
				Area:     "West",
				Country:  "US",
			},
			Region: clusterv1.RegionInfo{
				Region:           "us-west",
				AvailabilityZone: []string{"us-west-1", "us-west-2"},
			},
			Operator: clusterv1.OperatorInfo{
				Operator: "globalscheduler",
			},
			Flavors: []clusterv1.FlavorInfo{
				{FlavorID: "small", TotalCapacity: 10},
				{FlavorID: "medium", TotalCapacity: 20},
			},
			Storage: []clusterv1.StorageSpec{
				{TypeID: "sata", StorageCapacity: 2000},
				{TypeID: "sas", StorageCapacity: 1000},
				{TypeID: "ssd", StorageCapacity: 3000},
			},
			EipCapacity:   3,
			CPUCapacity:   8,
			MemCapacity:   256,
			ServerPrice:   10,
			HomeScheduler: "scheduler1",
		},
	}

	_, err := c.clusterclientset.GlobalschedulerV1().Clusters(corev1.NamespaceDefault).Create(object)
	errorMessage := err
	if err != nil {
		klog.Fatalf("could not create: %v", errorMessage)
	}
	klog.Infof("Created a cluster: %s", object.Name)
	return err
}
