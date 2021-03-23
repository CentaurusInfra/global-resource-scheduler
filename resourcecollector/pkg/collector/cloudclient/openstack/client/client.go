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
	"k8s.io/klog"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

func NewComputeV2Client(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})
	if err != nil {
		klog.Errorf("NewComputeV2 err: %s", err)
		return nil, err
	}
	return client, nil
}

func NewVolumeV3Client(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	client, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})
	if err != nil {
		klog.Errorf("NewBlockStorageV3 err: %s", err)
		return nil, err
	}
	return client, nil
}
