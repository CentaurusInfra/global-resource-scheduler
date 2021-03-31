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

package service

import (
	"k8s.io/klog"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

// Push updated region resources
func PushRegionResources(req *restful.Request, resp *restful.Response) {
	resourcesReq := new(types.RegionResourcesReq)
	err := req.ReadEntity(&resourcesReq)
	if err != nil {
		klog.Errorf("Failed to unmarshall region resources from request body, err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(err.Error()))
		return
	}
	klog.Infof("RegionResourceReq: %s", utils.GetJSONString(resourcesReq))
	resources := resourcesReq.RegionResources
	sched := scheduler.GetScheduler()
	if sched == nil {
		klog.Errorf("Scheduler is not initialized, please wait...")
		utils.WriteFailedJSONResponse(resp, http.StatusInternalServerError, utils.InternalServerError())
		return
	}
	for _, resource := range resources {
		err = sched.UpdateSiteDynamicResource(resource.RegionName, &types.SiteResource{CPUMemResources: resource.CPUMemResources, VolumeResources: resource.VolumeResources})
		if err != nil {
			klog.Errorf("Schedule  to update site dynamic resource for region %s with  err %v", resource.RegionName, err)
			utils.WriteFailedJSONResponse(resp, http.StatusInternalServerError, utils.InternalServerWithError(err.Error()))
			return
		}
	}
	resourceResp := types.SiteResourceRes{Result: "ok"}
	resp.WriteHeaderAndEntity(http.StatusCreated, resourceResp)
}
