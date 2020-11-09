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
	clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

// Client is an API client to help perform CRUD for CRD instances.
type ClusterClient struct {
	clientset *clientset.Clientset
	namespace string
	plural    string
}

// GetNamespace returns the namespace the client talks to.
func (c *ClusterClient) GetNamespace() string {
	return c.namespace
}

// GetPlural returns the plural the client is managing.
func (c *ClusterClient) GetPlural() string {
	return c.plural
}

func NewClusterClient(clusterClientset *clientset.Clientset, namespace string) (*ClusterClient, error) {
	c := &ClusterClient{
		clientset: clusterClientset,
		namespace: namespace,
		plural:    clusterv1.Plural,
	}
	return c, nil
}
