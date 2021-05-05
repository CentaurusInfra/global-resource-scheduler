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

package defaultbinder

import (
	"context"
	"fmt"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"strconv"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// Name of the plugin used in the plugin registry and configurations.
const Name = "DefaultBinder"

// DefaultBinder binds pods to site using a k8s client.
type DefaultBinder struct {
	handle interfaces.FrameworkHandle
}

var _ interfaces.BindPlugin = &DefaultBinder{}

// New creates a DefaultBinder.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &DefaultBinder{handle: handle}, nil
}

// Name returns the name of the plugin.
func (b DefaultBinder) Name() string {
	return Name
}

// Bind binds pods to site using the k8s client.
func (b DefaultBinder) Bind(ctx context.Context, state *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) *interfaces.Status {
	region := siteCacheInfo.GetSite().RegionAzMap.Region

	//eipNum : private data
	resInfo := types.AllResInfo{CpuAndMem: map[string]types.CPUAndMemory{}, Storage: map[string]float64{}}
	siteID := siteCacheInfo.Site.SiteID

	stack.Selected.SiteID = siteID
	stack.Selected.Region = region
	stack.Selected.AvailabilityZone = siteCacheInfo.GetSite().RegionAzMap.AvailabilityZone
	stack.Selected.ClusterName = siteCacheInfo.Site.ClusterName
	stack.Selected.ClusterNamespace = siteCacheInfo.Site.ClusterNamespace

	//siteSelectedInfo is type of SiteSelectorInfo at cycle_state.go
	siteSelectedInfo, err := interfaces.GetSiteSelectorState(state, siteID)
	if err != nil {
		klog.Errorf("GetSiteSelectorState failed! err: %s", err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting site %q info failed: %v", siteID, err))
	}
	klog.Errorf("GetSiteSelectorState: %v", siteSelectedInfo)
	if len(stack.Resources) != len(siteSelectedInfo.Flavors) {
		klog.Errorf("flavor count not equal to server count! err: %s", err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("siteID(%s) flavor count not equal to "+
			"server count!", siteID))
	}

	//i, stackresource := range boundStack.Resources
	for i := 0; i < len(stack.Resources); i++ {
		//CPU & Mem
		flavorID := siteSelectedInfo.Flavors[i].FlavorID
		stack.Resources[i].FlavorIDSelected = flavorID
		flv, ok := cache.FlavorCache.GetFlavor(flavorID, region)
		if !ok {
			klog.Warningf("flavor %s not found in region(%s)", flavorID, region)
			continue
		}
		vCPUInt, err := strconv.ParseInt(flv.Vcpus, 10, 64)
		if err != nil || vCPUInt <= 0 {
			klog.Warningf("flavor %s is invalid in region(%s)", flavorID, region)
			continue
		}
		reqRes, ok := resInfo.CpuAndMem[flv.OsExtraSpecs.ResourceType]
		if !ok {
			reqRes = types.CPUAndMemory{VCPU: 0, Memory: 0}
		}
		reqRes.VCPU += vCPUInt * int64(stack.Resources[i].Count)
		reqRes.Memory += flv.Ram * int64(stack.Resources[i].Count)

		//put them all to resInfo
		resInfo.CpuAndMem[flv.OsExtraSpecs.ResourceType] = reqRes
	}
	klog.Infof("Bind - Resource : %v", resInfo)
	b.handle.Cache().UpdateSiteWithResInfo(siteID, resInfo)

	klog.Infof("Resource Deduction Before: %v", siteCacheInfo)
	snapshotSiteCacheInfo, err := b.handle.SnapshotSharedLister().SiteCacheInfos().Get(siteID)
	if err != nil {
		klog.Errorf("snapshotSiteCacheInfo of site %s not found error: ", siteID, err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting site %q info failed: %v", siteID, err))
	}
	klog.Infof("Resource Deduction - snapshotSiteCacheInfo: %v, %v", snapshotSiteCacheInfo.Site, snapshotSiteCacheInfo)
	snapshotSiteCacheInfo.DeductSiteResInfo(resInfo)
	klog.Infof("Resource Deduction After: %v, %v", siteCacheInfo.Site, siteCacheInfo)

	//siteCacheInfo.DeductSiteResInfo(resInfo)
	//klog.Infof("Resource Deduction After: %v", siteCacheInfo)
	return nil
}

// Bind binds pods to site using the k8s client.
/*func (b DefaultBinder) Bind(ctx context.Context, state *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) *interfaces.Status {
	region := siteCacheInfo.GetSite().RegionAzMap.Region
	resInfo := types.AllResInfo{CpuAndMem: map[string]types.CPUAndMemory{}, Storage: map[string]float64{}}
	siteID := siteCacheInfo.Site.SiteID

	stack.Selected.SiteID = siteID
	stack.Selected.Region = region
	stack.Selected.AvailabilityZone = siteCacheInfo.GetSite().RegionAzMap.AvailabilityZone
	stack.Selected.ClusterName = siteCacheInfo.Site.ClusterName
	stack.Selected.ClusterNamespace = siteCacheInfo.Site.ClusterNamespace

	siteSelectedInfo, err := interfaces.GetSiteSelectorState(state, siteID)
	if err != nil {
		klog.Errorf("GetSiteSelectorState failed! err: %s", err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting site %q info failed: %v", siteID, err))
	}

	if len(stack.Resources) != len(siteSelectedInfo.Flavors) {
		klog.Errorf("flavor count not equal to server count! err: %s", err)
		return interfaces.NewStatus(interfaces.Error, fmt.Sprintf("siteID(%s) flavor count not equal to "+
			"server count!", siteID))
	}

	for i := 0; i < len(stack.Resources); i++ {
		flavorID := siteSelectedInfo.Flavors[i].FlavorID
		stack.Resources[i].FlavorIDSelected = flavorID
		flv, ok := cache.FlavorCache.GetFlavor(flavorID, region)
		if !ok {
			klog.Warningf("flavor %s not found in region(%s)", flavorID, region)
			continue
		}
		vCPUInt, err := strconv.ParseInt(flv.Vcpus, 10, 64)
		if err != nil || vCPUInt <= 0 {
			klog.Warningf("flavor %s is invalid in region(%s)", flavorID, region)
			continue
		}

		reqRes, ok := resInfo.CpuAndMem[flv.OsExtraSpecs.ResourceType]
		if !ok {
			reqRes = types.CPUAndMemory{VCPU: 0, Memory: 0}
		}
		reqRes.VCPU += vCPUInt * int64(stack.Resources[i].Count)
		reqRes.Memory += flv.Ram * int64(stack.Resources[i].Count)
		resInfo.CpuAndMem[flv.OsExtraSpecs.ResourceType] = reqRes
	}

	b.handle.Cache().UpdateSiteWithResInfo(siteID, resInfo)

	return nil
}*/
