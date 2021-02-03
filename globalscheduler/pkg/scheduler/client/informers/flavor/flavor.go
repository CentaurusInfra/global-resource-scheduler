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

package flavor

import (
	"errors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/cloudclient"
	"strconv"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
)

// InformerFlavor provides access to a shared informer and lister for Flavors.
type InformerFlavor interface {
	Informer() cache.SharedInformer
}

type informerFlavor struct {
	factory   internalinterfaces.SharedInformerFactory
	name      string
	key       string
	period    time.Duration
	collector internalinterfaces.ResourceCollector
}

// New initial the informerFlavor
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration,
	collector internalinterfaces.ResourceCollector) InformerFlavor {
	return &informerFlavor{factory: f, name: name, key: key, period: period, collector: collector}
}

// NewFlavorInformer constructs a new informer for flavor.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFlavorInformer(client client.Interface, resyncPeriod time.Duration, name string, key string,
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

			var interfaceSlice []interface{}
			// todo MultiExec
			for _, info := range siteInfoCache.SiteInfoMap {
				cloudClient, err := cloudclient.NewClientSet(info.EipNetworkID)
				if err != nil {
					logger.Warnf("FlavorInformer.NewClientSet[%s] err: %s", info.EipNetworkID, err.Error())
					continue
				}
				client := cloudClient.ComputeV2()

				flasPages, err := flavors.ListDetail(client, flavors.ListOpts{}).AllPages()
				if err != nil {
					logger.Errorf("flavor list failed! err: %s", err.Error())
					return nil, err
				}
				flas, err := flavors.ExtractFlavors(flasPages)
				if err != nil {
					logger.Errorf("flavor ExtractFlavors failed! err: %s", err.Error())
					return nil, err
				}
				for _, flavor := range flas {
					flv := typed.Flavor{
						ID:    flavor.Name, // eg: "m1.small"
						Name:  flavor.Name, // eg: "m1.small"
						Vcpus: strconv.Itoa(flavor.VCPUs),
						Ram:   int64(flavor.RAM),
						OsExtraSpecs: typed.OsExtraSpecs{
							ResourceType: "default",
						},
					}
					regionFlv := typed.RegionFlavor{
						RegionFlavorID: info.Region + "|" + flavor.Name,
						Region:         info.Region,
						Flavor:         flv,
					}
					interfaceSlice = append(interfaceSlice, regionFlv)
				}
			}
			return interfaceSlice, nil
		}}, resyncPeriod, name, key, typed.ListOpts{})
}

func (f *informerFlavor) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewFlavorInformer(client, resyncPeriod, name, key, f.collector)
}

func (f *informerFlavor) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
