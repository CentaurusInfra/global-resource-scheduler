/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package siteresources

import (
	"errors"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/cloudclient"
	"sync"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
)

// InformerSiteResource provides access to a shared informer and lister for Hosts.
type InformerSiteResource interface {
	Informer() cache.SharedInformer
}

type informerSiteResources struct {
	factory   internalinterfaces.SharedInformerFactory
	name      string
	key       string
	period    time.Duration
	collector internalinterfaces.ResourceCollector
}

// New initial the informerSiteResources
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration,
	collector internalinterfaces.ResourceCollector) InformerSiteResource {
	return &informerSiteResources{factory: f, name: name, key: key, period: period, collector: collector}
}

// NewSiteResourcesInformer constructs a new informer for sites.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSiteResourcesInformer(client client.Interface, reSyncPeriod time.Duration, name string, key string,
	collector internalinterfaces.ResourceCollector) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			if collector == nil {
				return nil, errors.New("collector need to be init correctly")
			}
			siteInfoCache := collector.GetSiteInfos()
			if siteInfoCache == nil || siteInfoCache.SiteInfoMap == nil {
				logger.Errorf("get site info failed")
				return nil, errors.New("get site info failed")
			}

			// result set, []typed.SiteResource
			var interfaceSlice []interface{}
			var wg sync.WaitGroup
			for siteID, info := range siteInfoCache.SiteInfoMap {
				cloudClient, err := cloudclient.NewClientSet(info.EipNetworkID)
				if err != nil {
					logger.Warnf("SiteResourcesInformer.NewClientSet[%s] err: %s", info.EipNetworkID, err.Error())
					collector.RecordSiteUnreacheable(siteID)
					continue
				}
				client := cloudClient.ComputeV2()
				if client == nil {
					logger.Errorf("Cluster[%s] computeV2 client is null!", info.EipNetworkID)
					continue
				}

				wg.Add(1)
				go func(siteID, region, az string, client *gophercloud.ServiceClient) {
					defer wg.Done()
					ret, err := getSiteResource(siteID, region, az, client)
					if err != nil {
						logger.Errorf("site[%s] list failed! err: %s", siteID, err.Error())
						return
					}
					interfaceSlice = append(interfaceSlice, ret)
				}(siteID, info.Region, info.AvailabilityZone, client)
			}
			wg.Wait()

			return interfaceSlice, nil
		}}, reSyncPeriod, name, key, nil)
}

// Get host resource information for each cluster(az) (goroutine concurrent execution)
func getSiteResource(siteID, region, az string, client *gophercloud.ServiceClient) (interface{}, error) {
	hypervisorsPages, err := hypervisors.List(client).AllPages()
	if err != nil {
		logger.Errorf("hypervisors list failed! err: %s", err.Error())
		return nil, err
	}
	hs, err := hypervisors.ExtractHypervisors(hypervisorsPages)
	if err != nil {
		logger.Errorf("hypervisors ExtractHypervisors failed! err: %s", err.Error())
		return nil, err
	}

	hosts := make([]*typed.Host, 0)
	// each host
	for _, h := range hs {
		host := &typed.Host{
			UsedVCPUs:    h.VCPUsUsed,
			UsedMem:      h.MemoryMBUsed,
			TotalVCPUs:   h.VCPUs,
			TotalMem:     h.MemoryMB,
			ResourceType: "default",
		}
		hosts = append(hosts, host)
	}
	sr := typed.SiteResource{
		SiteID:           siteID,
		Region:           region,
		AvailabilityZone: az,
		Hosts:            hosts,
	}
	return sr, nil
}

func (f *informerSiteResources) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewSiteResourcesInformer(client, resyncPeriod, name, key, f.collector)
}

func (f *informerSiteResources) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
