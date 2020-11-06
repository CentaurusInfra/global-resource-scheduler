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

package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

// Create post an instance of CRD into Kubernetes.
func (cc *ClusterClient) Create(obj *clusterv1.Cluster) (*clusterv1.Cluster, error) {
	return cc.clientset.GlobalschedulerV1().Clusters(cc.namespace).Create(obj)
}

// Update puts new instance of CRD to replace the old one.
func (cc *ClusterClient) Update(obj *clusterv1.Cluster) (*clusterv1.Cluster, error) {
	return cc.clientset.GlobalschedulerV1().Clusters(cc.namespace).Update(obj)
}

// Delete removes the CRD instance by given name and delete options.
func (cc *ClusterClient) Delete(name string, opts *metav1.DeleteOptions) error {
	return cc.clientset.GlobalschedulerV1().Clusters(cc.namespace).Delete(name, opts)
}

// Get returns a pointer to the CRD instance.
func (cc *ClusterClient) Get(name string, opts metav1.GetOptions) (*clusterv1.Cluster, error) {
	return cc.clientset.GlobalschedulerV1().Clusters(cc.namespace).Get(name, opts)
}

// List returns a list of CRD instances by given list options.
func (cc *ClusterClient) List(opts metav1.ListOptions) (*clusterv1.ClusterList, error) {
	return cc.clientset.GlobalschedulerV1().Clusters(cc.namespace).List(opts)
}
