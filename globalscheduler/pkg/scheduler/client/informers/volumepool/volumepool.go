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

package volumepool

import (
	"errors"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/schedulerstats"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// InformerVolumePool provides access to a shared informer and lister for eip available.
type InformerVolumePool interface {
	Informer() cache.SharedInformer
}

type informerVolumePool struct {
	factory   internalinterfaces.SharedInformerFactory
	name      string
	key       string
	period    time.Duration
	collector internalinterfaces.ResourceCollector
}

// New initial the informerSite
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration,
	collector internalinterfaces.ResourceCollector) InformerVolumePool {
	return &informerVolumePool{factory: f, name: name, key: key, period: period, collector: collector}
}

// NewVolumePoolInformer constructs a new informer for volumepool.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVolumePoolInformer(client client.Interface, resyncPeriod time.Duration, name string, key string,
	collector internalinterfaces.ResourceCollector) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			if collector == nil {
				return nil, errors.New("collector need to be init correctly")
			}
			siteInfoCache := collector.GetSiteInfos()
			if siteInfoCache == nil || siteInfoCache.SiteInfoMap == nil {
				klog.Errorf("get site info failed")
				return nil, errors.New("get site info failed")
			}

			var interfaceSlice []interface{}
			// todo MultiExec
			for _, info := range siteInfoCache.SiteInfoMap {
				cloudClient, err := collector.GetClientSet(info.EipNetworkID)
				if err != nil {
					klog.Warningf("VolumePool.NewClientSet[%s] err: %s", info.EipNetworkID, err.Error())
					continue
				}
				client := cloudClient.VolumeV3()

				listOpts := schedulerstats.ListOpts{
					Detail: true,
				}
				allPages, err := schedulerstats.List(client, listOpts).AllPages()
				if err != nil {
					klog.Errorf("schedulerstats list failed! err: %s", err.Error())
					return nil, err
				}
				allStats, err := schedulerstats.ExtractStoragePools(allPages)
				if err != nil {
					klog.Errorf("schedulerstats ExtractStoragePools failed! err: %s", err.Error())
					return nil, err
				}
				volPools := typed.VolumePools{}
				for _, stat := range allStats {
					cap := typed.VolumeCapabilities{
						VolumeType:               stat.Capabilities.VolumeBackendName,
						FreeCapacityGb:           stat.Capabilities.FreeCapacityGB,
						TotalCapacityGb:          stat.Capabilities.TotalCapacityGB,
						ProvisionedCapacityGb:    stat.Capabilities.ProvisionedCapacityGB,
						MaxOverSubscriptionRatio: 1,
					}
					volPoolInfo := typed.VolumePoolInfo{
						AvailabilityZone: info.AvailabilityZone,
						Name:             stat.Name,
						Capabilities:     cap,
					}
					volPools.StoragePools = append(volPools.StoragePools, volPoolInfo)
				}
				interfaceSlice = append(interfaceSlice, typed.RegionVolumePool{
					VolumePools: volPools,
					Region:      info.Region,
				})
			}
			return interfaceSlice, nil
		}}, resyncPeriod, name, key, types.ListSiteOpts{})
}

func (f *informerVolumePool) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewVolumePoolInformer(client, resyncPeriod, name, key, f.collector)
}

func (f *informerVolumePool) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
