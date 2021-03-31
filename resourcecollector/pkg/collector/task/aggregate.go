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

package task

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/cloudclient"
	"sync"
)

func BuildHostAzMap(collector internalinterfaces.ResourceCollector) {
	if collector == nil {
		klog.Error("collector need to be init correctly")
		return
	}
	regionResourceCache := collector.GetRegionResources()
	if regionResourceCache == nil || regionResourceCache.RegionResourceMap == nil {
		klog.Error("get region resource info failed")
		return
	}
	var wg sync.WaitGroup
	for _, info := range regionResourceCache.RegionResourceMap {
		cloudClient, err := cloudclient.NewClientSet(info.EipNetworkID)
		if err != nil {
			klog.Warningf("SyncResources.NewClientSet[%s] err: %s", info.EipNetworkID, err.Error())
			continue
		}
		clientComputeV2 := cloudClient.ComputeV2()
		if clientComputeV2 == nil {
			klog.Errorf("Cloud ip [%s] computeV2 client is null!", info.EipNetworkID)
			continue
		}
		wg.Add(1)
		go func(client *gophercloud.ServiceClient, regionResource *typed.RegionResource, collector internalinterfaces.ResourceCollector) {
			defer wg.Done()
			aggregatePages, err := aggregates.List(client).AllPages()
			if err != nil {
				klog.Errorf("aggregates list failed! err: %s", err.Error())

			}
			aggregates, err := aggregates.ExtractAggregates(aggregatePages)

			for _, aggregate := range aggregates {
				regionResourceCache.AddHostAz(regionResource.RegionName, aggregate.Hosts, aggregate.AvailabilityZone)
			}
			klog.Infof("The host az map has been added as %v", collector.GetRegionResources().RegionResourceMap[regionResource.RegionName])
		}(clientComputeV2, info, collector)
	}
	wg.Wait()
	klog.Info("The host az map has been synchronized")
}
