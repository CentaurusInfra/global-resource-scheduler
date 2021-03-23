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

package cloudclient

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/klog"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"

	"k8s.io/kubernetes/resourcecollector/pkg/collector/cloudclient/openstack/client"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/common/config"
)

type ClientSet struct {
	computeV2 *gophercloud.ServiceClient
	volumeV3  *gophercloud.ServiceClient
}

func NewClientSet(siteEndpoint string) (*ClientSet, error) {
	cs := &ClientSet{}

	// 1. ops
	scop := gophercloud.AuthScope{
		ProjectName: config.GlobalConf.OpenStackScopProjectName,
		DomainID:    config.GlobalConf.OpenStackScopDomainID,
	}
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: fmt.Sprintf("http://%s/identity", siteEndpoint),
		Username:         config.GlobalConf.OpenStackUsername,
		Password:         config.GlobalConf.OpenStackPassword,
		DomainID:         config.GlobalConf.OpenStackDomainID,
		Scope:            &scop,
	}

	// 2. provider
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		klog.Errorf("AuthenticatedClient wrong: %s", err.Error())
		return nil, err
	}

	// 3. init client
	cs.computeV2, err = client.NewComputeV2Client(provider)
	if err != nil {
		klog.Errorf("NewComputeV2Client wrong: %s", err.Error())
		return nil, err
	}
	cs.volumeV3, err = client.NewVolumeV3Client(provider)
	if err != nil {
		klog.Errorf("NewVolumeV3Client wrong: %s", err.Error())
		return nil, err
	}

	//===edit===
	re, err := regexp.Compile("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}")
	if err != nil {
		klog.Errorf("some wrong: %s", err.Error())
		return nil, err
	}
	intranetIP := string(re.Find([]byte(cs.computeV2.Endpoint)))
	publicIP := string(re.Find([]byte(siteEndpoint)))
	if publicIP != "" {
		cs.computeV2.Endpoint = strings.Replace(cs.computeV2.Endpoint, intranetIP, publicIP, 1)
		cs.volumeV3.Endpoint = strings.Replace(cs.volumeV3.Endpoint, intranetIP, publicIP, 1)
	}

	return cs, nil
}

func (c *ClientSet) ComputeV2() *gophercloud.ServiceClient {
	return c.computeV2
}

func (c *ClientSet) VolumeV3() *gophercloud.ServiceClient {
	return c.volumeV3
}
