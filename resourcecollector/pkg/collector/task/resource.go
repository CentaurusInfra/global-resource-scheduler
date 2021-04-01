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
	"bytes"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/schedulerstats"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/cloudclient"
	"net/http"
	"sync"
)

func SyncResources(collector internalinterfaces.ResourceCollector) {
	if collector == nil {
		klog.Error("collector need to be init correctly")
		return
	}
	regionResourceCache := collector.GetRegionResources()
	if regionResourceCache == nil || regionResourceCache.RegionResourceMap == nil {
		klog.Error("get region resource info failed")
		return
	}
	regionResources := make([]types.RegionResource, 0)
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

		clientVolumeV3 := cloudClient.VolumeV3()
		if clientVolumeV3 == nil {
			klog.Errorf("Cloud ip [%s] VolumeV3 client is null!", info.EipNetworkID)
			continue
		}
		var cpuAndMemResources []types.AzCpuMem
		var volumeResources []types.Volume
		var wg sync.WaitGroup
		wg.Add(1)
		go func(client *gophercloud.ServiceClient, regionResource *typed.RegionResource, collector internalinterfaces.ResourceCollector) {
			defer wg.Done()
			cpuAndMemResources, err = getRegionCpuAndMemResources(client, regionResource, collector)
			if err != nil {
				klog.Errorf("region cpu and mem resource [%s] list failed! err: %s", regionResource.RegionName, err.Error())
				return
			}
		}(clientComputeV2, info, collector)
		wg.Add(1)
		go func(client *gophercloud.ServiceClient, regionResource *typed.RegionResource, collector internalinterfaces.ResourceCollector) {
			defer wg.Done()
			volumeResources, err = getRegionVolumeResources(client, regionResource, collector)
			if err != nil {
				klog.Errorf("region volume resource [%s] list failed! err: %s", regionResource.RegionName, err.Error())
				return
			}
		}(clientVolumeV3, info, collector)
		wg.Wait()
		if len(cpuAndMemResources) > 0 || len(volumeResources) > 0 {
			regionResource := types.RegionResource{RegionName: info.RegionName}
			if len(cpuAndMemResources) > 0 {
				regionResource.CPUMemResources = cpuAndMemResources
			}
			if len(volumeResources) > 0 {
				regionResource.VolumeResources = volumeResources
			}
			regionResources = append(regionResources, regionResource)
		}
	}
	if len(regionResources) > 0 {
		for _, scheduler := range collector.GetSchedulers() {
			resourceReq, err := json.Marshal(types.RegionResourcesReq{regionResources})
			if err != nil {
				klog.Errorf("Failed to convert the updated region resources %v to requests with error %v", regionResources, err)
				return
			}
			schedulerEndpoint := scheduler.Spec.IpAddress + ":" + scheduler.Spec.PortNumber + constants.RegionResourcesURL
			req, _ := http.NewRequest("POST", schedulerEndpoint, bytes.NewBuffer(resourceReq))
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				klog.Errorf("Failed to send the updated region resources %v to scheduler %s with error", regionResources, schedulerEndpoint, err)
			}

			if err != nil {
				klog.Errorf("Failed to send the updated region resources %v to scheduler %s with error", regionResources, schedulerEndpoint, err)
				return
			}
		}
	}
}

// Get cpu and memory resources information for each region
func getRegionCpuAndMemResources(client *gophercloud.ServiceClient, regionResource *typed.RegionResource, collector internalinterfaces.ResourceCollector) ([]types.AzCpuMem, error) {
	res := make([]types.AzCpuMem, 0)

	hypervisorsPages, err := hypervisors.List(client).AllPages()
	if err != nil {
		klog.Errorf("hypervisors list failed! err: %s", err.Error())
		return res, err
	}
	hs, err := hypervisors.ExtractHypervisors(hypervisorsPages)

	if err != nil {
		klog.Errorf("hypervisors ExtractHypervisors failed! err: %s", err.Error())
		return res, err
	}

	// If there is  az configured, we will get  the host resources per az from the az host mapping and assign them to different az
	if regionResource.HostAzMap != nil && len(regionResource.HostAzMap) > 0 {
		azCpuMemMap := make(map[string]*typed.CpuAndMemResource, 0)
		for _, h := range hs {
			if az, ok := regionResource.HostAzMap[h.HypervisorHostname]; ok {
				if _, existed := azCpuMemMap[az]; !existed {
					azCpuMemMap[az] = &typed.CpuAndMemResource{0, 0}
				}
				azCpuMemMap[az].TotalMem = azCpuMemMap[az].TotalMem + h.MemoryMB
				azCpuMemMap[az].TotalVCPUs = azCpuMemMap[az].TotalVCPUs + h.VCPUs
			}
		}
		for az, cmr := range azCpuMemMap {
			if collector.GetRegionResources().UpdateCpuAndMemResource(regionResource.RegionName, az, *cmr) {
				klog.Infof("The CpuAndMemResource for region[%s] az[%s] has been updated to %v", regionResource.RegionName, az, cmr)
				res = append(res, types.AzCpuMem{AvailabilityZone: az, CpuCapacity: int64(cmr.TotalVCPUs), MemCapacity: int64(cmr.TotalMem)})
			}
		}
	}
	// If there is no host az mapping, the system will assign default values for testing
	defaultCmd := typed.CpuAndMemResource{TotalVCPUs: 3, TotalMem: 512}
	for az, cmd := range regionResource.CpuAndMemResources {
		if cmd.TotalMem == 0 && cmd.TotalVCPUs == 0 {
			collector.GetRegionResources().UpdateCpuAndMemResource(regionResource.RegionName, az, defaultCmd)
			res = append(res, types.AzCpuMem{AvailabilityZone: az, CpuCapacity: int64(defaultCmd.TotalVCPUs), MemCapacity: int64(defaultCmd.TotalMem)})
		}
	}
	return res, nil
}

// Get volume resource information for each region
func getRegionVolumeResources(client *gophercloud.ServiceClient, regionResource *typed.RegionResource, collector internalinterfaces.ResourceCollector) ([]types.Volume, error) {
	res := make([]types.Volume, 0)
	volumeResources := make([]typed.VolumeResource, 0)

	if regionResource.HostAzMap == nil || len(regionResource.HostAzMap) == 0 {
		// If there is no host az mapping, the system will assign default values for testing
		volumeResources = []typed.VolumeResource{{VolumeType: "lvmdriver-1", TotalCapacityGb: 22}}
	} else {
		listOpts := schedulerstats.ListOpts{
			Detail: true,
		}
		allPages, err := schedulerstats.List(client, listOpts).AllPages()
		if err != nil {
			klog.Errorf("schedulerstats list failed! err: %s", err.Error())
			return res, err
		}
		allStats, err := schedulerstats.ExtractStoragePools(allPages)
		if err != nil {
			klog.Errorf("schedulerstats ExtractStoragePools failed! err: %s", err.Error())
			return res, err
		}

		for _, stat := range allStats {
			volumeResource := typed.VolumeResource{
				VolumeType:      stat.Capabilities.VolumeBackendName,
				TotalCapacityGb: stat.Capabilities.TotalCapacityGB,
			}
			volumeResources = append(volumeResources, volumeResource)
		}
	}

	if collector.GetRegionResources().UpdateVolumeResource(regionResource.RegionName, volumeResources) {
		klog.Infof("The volumeResources for region[%s]  has been updated to %v", regionResource.RegionName, volumeResources)
		for _, volumeResource := range volumeResources {
			res = append(res, types.Volume{TypeId: volumeResource.VolumeType, StorageCapacity: volumeResource.TotalCapacityGb})
		}
	}

	return res, nil
}
