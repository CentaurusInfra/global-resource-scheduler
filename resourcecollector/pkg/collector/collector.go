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
	"errors"
	"sync"
	"time"

	"k8s.io/klog"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/cloudclient"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/common/config"
	internalcache "k8s.io/kubernetes/resourcecollector/pkg/collector/internal/cache"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/rpcclient"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/siteinfo"
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
func GetCollector() (*Collector, error) {
	if collector == nil {
		klog.Error("collector need to be init correctly")
		err := errors.New("collector need to be init correctly")
		return nil, err
	}

	return collector, nil
}

type Collector struct {
	ResourceCache         internalcache.Cache
	siteCacheInfoSnapshot *internalcache.Snapshot
	SiteInfoCache         *siteinfo.SiteInfoCache
	SchedulerInfoCache    []*schedulercrdv1.Scheduler

	clientSetCache *cloudclient.ClientSetCache

	mutex           sync.Mutex
	unreachableNum  map[string]int
	unreachableChan chan string
}

func NewCollector(stopCh <-chan struct{}) (*Collector, error) {
	c := &Collector{
		ResourceCache:         internalcache.New(30*time.Second, stopCh),
		siteCacheInfoSnapshot: internalcache.NewEmptySnapshot(),
		SiteInfoCache:         siteinfo.NewSiteInfoCache(),
		SchedulerInfoCache:    make([]*schedulercrdv1.Scheduler, 0),

		clientSetCache: cloudclient.NewClientSetCache(),

		unreachableNum:  make(map[string]int),
		unreachableChan: make(chan string, 3),
	}
	return c, nil
}

func (c *Collector) RefreshClientSetCache(stopCh <-chan struct{}) {
	for {
		select {
		case <-time.Tick(time.Minute * time.Duration(config.GlobalConf.RefreshOpenStackTokenInterval)):
			endpoints := c.SiteInfoCache.GetAllSiteEndpoints()
			c.clientSetCache.RefreshClientSets(endpoints)
		case <-stopCh:
			klog.Info("stop Refresh ClientSet")
			break
		}
	}
}

func (c *Collector) RecordSiteUnreacheable(siteID, clusterNamespace, clusterName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.unreachableNum[siteID]++
	klog.Infof("RecordSiteUnreacheable site[%s] unreachableNum[%d]", siteID, c.unreachableNum[siteID])
	if c.unreachableNum[siteID] == config.GlobalConf.MaxUnreachableNum {
		delete(c.unreachableNum, siteID)
		go c.notifySiteUnreachable(siteID, clusterNamespace, clusterName)
	}
}

func (c *Collector) notifySiteUnreachable(siteID, clusterNamespace, clusterName string) {
	klog.Infof("site[%s] is unreachable, send grpc to cluster-controller", siteID)
	err := rpcclient.GrpcUpdateClusterStatus(clusterNamespace, clusterName, rpcclient.StateUnreachable)
	if err != nil {
		klog.Errorf("send grpc to cluster-controller err: %s", err.Error())
		return
	}
	klog.Infof("update site[%s] state unreachable success", siteID)
}

// snapshot snapshots scheduler cache and node infos for all fit and priority
// functions.
func (c *Collector) snapshot() error {
	// Used for all fit and priority funcs.
	return c.Cache().UpdateSnapshot(c.siteCacheInfoSnapshot)
}

// Cache returns the cache in scheduler for test to check the data in scheduler.
func (c *Collector) Cache() internalcache.Cache {
	return c.ResourceCache
}

func (c *Collector) GetSiteInfos() *siteinfo.SiteInfoCache {
	return c.SiteInfoCache
}

func (c *Collector) GetClientSet(siteEndpoint string) (*cloudclient.ClientSet, error) {
	return c.clientSetCache.GetClientSet(siteEndpoint)
}

func (c *Collector) GetSnapshot() (*internalcache.Snapshot, error) {
	err := c.snapshot()
	if err != nil {
		return nil, err
	}
	return c.siteCacheInfoSnapshot, nil
}

func (c *Collector) GetSchedulerInfoCache() []*schedulercrdv1.Scheduler {
	return c.SchedulerInfoCache
}

// start resource cache informer and run
func (c *Collector) StartInformersAndRun(stopCh <-chan struct{}) {
	go func(stopCh2 <-chan struct{}) {
		// init informer
		informers.InformerFac = informers.NewSharedInformerFactory(nil, 60*time.Second)

		// init flavor informer
		flavorInterval := config.GlobalConf.FlavorInterval
		informers.InformerFac.Flavor(informers.FLAVOR, "RegionFlavorID",
			time.Duration(flavorInterval)*time.Second, c).Informer()

		// init site resource informer
		siteResourceInterval := config.GlobalConf.SiteResourceInterval
		siteResourceInformer := informers.InformerFac.SiteResource(informers.SITERESOURCES, "SiteID",
			time.Duration(siteResourceInterval)*time.Second, c).Informer()
		siteResourceInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: addSitesToCache,
			})

		// init volume pool informer
		volumePoolInterval := config.GlobalConf.VolumePoolInterval
		volumePoolInformer := informers.InformerFac.VolumePools(informers.VOLUMEPOOLS, "Region",
			time.Duration(volumePoolInterval)*time.Second, c).Informer()
		volumePoolInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: updateVolumePools,
			})

		// init volume type informer
		volumeTypeInterval := config.GlobalConf.VolumeTypeInterval
		informers.InformerFac.VolumeType(informers.VOLUMETYPE, "ID",
			time.Duration(volumeTypeInterval)*time.Second, c).Informer()

		// todo init eip pool informer
		//eipPoolInterval := config.GlobalConf.EipPoolInterval
		//eipPoolInformer := informers.InformerFac.EipPools(informers.EIPPOOLS, "Region",
		//	time.Duration(eipPoolInterval)*time.Second).Informer()
		//eipPoolInformer.AddEventHandler(
		//	cache.ResourceEventHandlerFuncs{
		//		ListFunc: updateEipPools,
		//	})

		informers.InformerFac.Start(stopCh2)

	}(stopCh)
}

// update EipPools with sched cache
func updateEipPools(obj []interface{}) {
	if obj == nil {
		return
	}

	for _, eipPoolObj := range obj {
		eipPool, ok := eipPoolObj.(typed.EipPool)
		if !ok {
			klog.Warning("convert interface to (typed.EipPool) failed.")
			continue
		}

		err := collector.Cache().UpdateEipPool(&eipPool)
		if err != nil {
			klog.Infof("UpdateEipPool failed! err: %s", err)
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
			klog.Warning("convert interface to (typed.VolumePools) failed.")
			continue
		}

		err := collector.Cache().UpdateVolumePool(&volumePool)
		if err != nil {
			klog.Infof("updateVolumePools failed! err: %s", err)
		}
	}
}

// add site to cache
func addSitesToCache(objs []interface{}) {
	if objs == nil {
		return
	}

	col, err := GetCollector()
	if err != nil {
		klog.Errorf("GetCollector err: %s", err.Error())
		return
	}
	//siteInfos := informers.InformerFac.GetInformer(informers.SITEINFOS).GetStore().List()
	siteInfos := col.SiteInfoCache.SiteInfoMap

	// Iterate the site information collected by the SiteResources Informer
	for _, obj := range objs {
		siteResource, ok := obj.(typed.SiteResource)
		if !ok {
			klog.Warning("convert interface to (typed.SiteResource) failed.")
			continue
		}

		// Check to see if this site exists in SiteInfo
		var isFind = false
		for _, siteInfo := range siteInfos {
			if siteInfo.SiteID == siteResource.SiteID {
				// Integrate site static information and resource information
				info := convertToSite(siteInfo, siteResource)
				err := collector.Cache().AddSite(info)
				if err != nil {
					klog.Infof("add site to cache failed! err: %s", err)
				}

				isFind = true
				break
			}
		}

		if !isFind {
			klog.Warningf("siteResource.SiteID[%s] is not in siteInfo, Not add to the cache", siteResource.SiteID)
		}
	}

	collector.Cache().PrintString()
}

// Integrate site static information and resource information
func convertToSite(siteInfo *typed.SiteInfo, siteResource typed.SiteResource) *types.Site {
	result := &types.Site{
		SiteID:           siteInfo.SiteID,
		ClusterName:      siteInfo.ClusterName,
		ClusterNamespace: siteInfo.ClusterNamespace,
		GeoLocation: types.GeoLocation{
			Country:  siteInfo.Country,
			Area:     siteInfo.Area,
			Province: siteInfo.Province,
			City:     siteInfo.City,
		},
		RegionAzMap: types.RegionAzMap{
			Region:           siteInfo.Region,
			AvailabilityZone: siteInfo.AvailabilityZone,
		},
		Operator:      siteInfo.Operator.Name,
		EipTypeName:   siteInfo.EipTypeName,
		Status:        siteInfo.Status,
		SiteAttribute: siteInfo.SiteAttributes,
	}

	result.Hosts = append(result.Hosts, siteResource.Hosts...)
	return result
}
