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

package collector

import (
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	internalcache "k8s.io/kubernetes/resourcecollector/pkg/collector/internal/cache"
)

var collector *Collector

// InitScheduler
func InitCollector(stopCh <-chan struct{}) error {
	var err error
	collector, err = NewCollector(stopCh)
	return err
}

// GetScheduler gets single scheduler instance. New Scheduler will only run once,
// if it runs failed, nil will be return.
func GetCollector() *Collector {
	if collector == nil {
		logger.Errorf("Collector need to be init correctly")
		return collector
	}

	return collector
}

type Collector struct {
	CollectorCache        internalcache.Cache
	siteCacheInfoSnapshot *internalcache.Snapshot
	SiteIPCache           *internalcache.SiteIPCache
}

func NewCollector(stopCh <-chan struct{}) (*Collector, error) {
	c := &Collector{
		CollectorCache:        internalcache.New(30*time.Second, stopCh),
		siteCacheInfoSnapshot: internalcache.NewEmptySnapshot(),
		SiteIPCache:           internalcache.NewSiteIPCache(),
	}
	return c, nil
}

// snapshot snapshots scheduler cache and node infos for all fit and priority
// functions.
func (c *Collector) snapshot() error {
	// Used for all fit and priority funcs.
	return c.Cache().UpdateSnapshot(c.siteCacheInfoSnapshot)
}

func (c *Collector) GetSnapshot() (*internalcache.Snapshot, error) {
	err := c.snapshot()
	if err != nil {
		return nil, err
	}
	return c.siteCacheInfoSnapshot, nil
}

// start resource cache informer and run
func (c *Collector) StartInformersAndRun(stopCh <-chan struct{}) {
	go func(stopCh2 <-chan struct{}) {
		// init informer
		informers.InformerFac = informers.NewSharedInformerFactory(nil, 60*time.Second)

		// init volume type informer
		volumetypeInterval := 600
		informers.InformerFac.VolumeType(informers.VOLUMETYPE, "ID",
			time.Duration(volumetypeInterval)*time.Second).Informer()

		// init site informer
		siteInfoInterval := 600
		informers.InformerFac.SiteInfo(informers.SITEINFOS, "SITEID",
			time.Duration(siteInfoInterval)*time.Second).Informer()

		// init flavor informer
		flavorInterval := 600
		informers.InformerFac.Flavor(informers.FLAVOR, "RegionFlavorID",
			time.Duration(flavorInterval)*time.Second).Informer()

		// init eip pool informer
		eipPoolInterval := 60
		eipPoolInformer := informers.InformerFac.EipPools(informers.EIPPOOLS, "Region",
			time.Duration(eipPoolInterval)*time.Second).Informer()
		eipPoolInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: updateEipPools,
			})

		// init volume pool informer
		volumePoolInterval := 60
		volumePoolInformer := informers.InformerFac.VolumePools(informers.VOLUMEPOOLS, "Region",
			time.Duration(volumePoolInterval)*time.Second).Informer()
		volumePoolInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: updateVolumePools,
			})

		// init site resource informer
		siteResourceInterval := 86400
		siteResourceInformer := informers.InformerFac.SiteResource(informers.SITERESOURCES, "SiteID",
			time.Duration(siteResourceInterval)*time.Second).Informer()
		siteResourceInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: addSitesToCache,
			})

		informers.InformerFac.Start(stopCh2)

	}(stopCh)
}

// Cache returns the cache in scheduler for test to check the data in scheduler.
func (c *Collector) Cache() internalcache.Cache {
	return c.CollectorCache
}

// update EipPools with sched cache
func updateEipPools(obj []interface{}) {
	if obj == nil {
		return
	}

	for _, eipPoolObj := range obj {
		eipPool, ok := eipPoolObj.(typed.EipPool)
		if !ok {
			logger.Warnf("convert interface to (typed.EipPool) failed.")
			continue
		}

		err := collector.Cache().UpdateEipPool(&eipPool)
		if err != nil {
			logger.Infof("UpdateEipPool failed! err: %s", err)
		}
	}
}

// update VolumePools with sched cache
func updateVolumePools(obj []interface{}) {
	if obj == nil {
		return
	}

	for _, volumePoolObj := range obj {
		volumePool, ok := volumePoolObj.(typed.RegionVolumePool)
		if !ok {
			logger.Warnf("convert interface to (typed.VolumePools) failed.")
			continue
		}

		err := collector.Cache().UpdateVolumePool(&volumePool)
		if err != nil {
			logger.Infof("updateVolumePools failed! err: %s", err)
		}
	}
}

// add site to cache
func addSitesToCache(obj []interface{}) {
	if obj == nil {
		return
	}

	siteInfos := informers.InformerFac.GetInformer(informers.SITEINFOS).GetStore().List()

	for _, sn := range obj {
		siteResource, ok := sn.(typed.SiteResource)
		if !ok {
			logger.Warnf("convert interface to (typed.SiteResource) failed.")
			continue
		}

		var isFind = false
		for _, site := range siteInfos {
			siteInfo, ok := site.(typed.SiteInfo)
			if !ok {
				continue
			}

			if siteInfo.Region == siteResource.Region && siteInfo.AvailabilityZone == siteResource.AvailabilityZone {
				info := convertToSite(siteInfo, siteResource)
				err := collector.Cache().AddSite(info)
				if err != nil {
					logger.Infof("add site to cache failed! err: %s", err)
				}

				isFind = true
				break
			}
		}

		if !isFind {
			site := &types.Site{
				SiteID: siteResource.Region + "--" + siteResource.AvailabilityZone,
				RegionAzMap: types.RegionAzMap{
					Region:           siteResource.Region,
					AvailabilityZone: siteResource.AvailabilityZone,
				},
				Status: constants.SiteStatusNormal,
			}

			site.Hosts = append(site.Hosts, siteResource.Hosts...)
			err := collector.Cache().AddSite(site)
			if err != nil {
				logger.Infof("add site to cache failed! err: %s", err)
			}
		}
	}

	collector.Cache().PrintString()
}

func convertToSite(site typed.SiteInfo, siteResource typed.SiteResource) *types.Site {
	result := &types.Site{
		SiteID: site.SiteID,
		GeoLocation: types.GeoLocation{
			Country:  site.Country,
			Area:     site.Area,
			Province: site.Province,
			City:     site.City,
		},
		RegionAzMap: types.RegionAzMap{
			Region:           site.Region,
			AvailabilityZone: site.AvailabilityZone,
		},
		Operator:      site.Operator.Name,
		EipTypeName:   site.EipTypeName,
		Status:        site.Status,
		SiteAttribute: site.SiteAttributes,
	}

	result.Hosts = append(result.Hosts, siteResource.Hosts...)
	return result
}
