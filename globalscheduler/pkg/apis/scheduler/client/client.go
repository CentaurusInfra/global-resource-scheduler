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
	schedulerv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
)

// Create post an instance of CRD into Kubernetes.
func (c *Client) Create(obj *schedulerv1.Scheduler) (*schedulerv1.Scheduler, error) {
	return c.clientset.GlobalschedulerV1().Schedulers(c.namespace).Create(obj)
}

// Update puts new instance of CRD to replace the old one.
func (c *Client) Update(obj *schedulerv1.Scheduler) (*schedulerv1.Scheduler, error) {
	return c.clientset.GlobalschedulerV1().Schedulers(c.namespace).Update(obj)
}

// Delete removes the CRD instance by given name and delete options.
func (c *Client) Delete(name string, opts *metav1.DeleteOptions) error {
	return c.clientset.GlobalschedulerV1().Schedulers(c.namespace).Delete(name, opts)
}

// Get returns a pointer to the CRD instance.
func (c *Client) Get(name string, opts metav1.GetOptions) (*schedulerv1.Scheduler, error) {
	return c.clientset.GlobalschedulerV1().Schedulers(c.namespace).Get(name, opts)
}

// List returns a list of CRD instances by given list options.
func (c *Client) List(opts metav1.ListOptions) (*schedulerv1.SchedulerList, error) {
	return c.clientset.GlobalschedulerV1().Schedulers(c.namespace).List(opts)
}
