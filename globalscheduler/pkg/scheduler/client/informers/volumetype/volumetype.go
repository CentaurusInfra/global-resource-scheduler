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

package volumetype

import (
	"errors"
	"k8s.io/klog"
	"time"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumetypes"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
)

// InformerVolumeType provides access to a shared informer and lister for VolumeType.
type InformerVolumeType interface {
	Informer() cache.SharedInformer
}

type informerVolumeType struct {
	factory   internalinterfaces.SharedInformerFactory
	name      string
	key       string
	period    time.Duration
	collector internalinterfaces.ResourceCollector
}

// New initial the informerVolumeType
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration,
	collector internalinterfaces.ResourceCollector) InformerVolumeType {
	return &informerVolumeType{factory: f, name: name, key: key, period: period, collector: collector}
}

// NewVolumeTypeInformer constructs a new informer for volume type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVolumeTypeInformer(client client.Interface, resyncPeriod time.Duration, name string, key string,
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
					klog.Warningf("VolumeType.NewClientSet[%s] err: %s", info.EipNetworkID, err.Error())
					continue
				}
				client := cloudClient.VolumeV3()

				allPages, err := volumetypes.List(client, volumetypes.ListOpts{}).AllPages()
				if err != nil {
					klog.Errorf("volumetypes list failed! err: %s", err.Error())
					return nil, err
				}
				volumeTypes, err := volumetypes.ExtractVolumeTypes(allPages)
				if err != nil {
					klog.Errorf("volumetypes ExtractVolumeTypes failed! err: %s", err.Error())
					return nil, err
				}
				for _, volType := range volumeTypes {
					regionVolType := typed.RegionVolumeType{
						VolumeType: typed.VolumeType{
							ID:   volType.ID,
							Name: volType.Name,
						},
						Region: info.Region,
					}
					if volType.ExtraSpecs != nil {
						if backendName, ok := volType.ExtraSpecs["volume_backend_name"]; ok {
							regionVolType.VolumeType.VolumeBackendName = backendName
						}
					}
					interfaceSlice = append(interfaceSlice, regionVolType)
				}
			}
			return interfaceSlice, nil
		}}, resyncPeriod, name, key, nil)
}

func (f *informerVolumeType) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewVolumeTypeInformer(client, resyncPeriod, name, key, f.collector)
}

func (f *informerVolumeType) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
