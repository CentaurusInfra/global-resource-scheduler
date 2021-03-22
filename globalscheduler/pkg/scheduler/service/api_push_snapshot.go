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

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"github.com/emicklei/go-restful"
)

// Schedule get snapshot
func PushSnapshot(req *restful.Request, resp *restful.Response) {
	//ctx := req.Request.Context()
	resourceReq := new(types.SiteResourceReq)

	region := req.PathParameter("regionname")
	err := req.ReadEntity(&resourceReq)
	if err != nil {
		klog.Errorf("Failed to unmarshall allocation from request body, err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(err.Error()))
		return
	}
	klog.Infof("SiteResourceReq : %s", utils.GetJSONString(resourceReq))
	resource := resourceReq.SiteResource
	//result := types.SiteResourceRes{Result: "OK"}
	sched := scheduler.GetScheduler()
	if sched == nil {
		klog.Errorf("Scheduler is not init, please wait...")
		utils.WriteFailedJSONResponse(resp, http.StatusInternalServerError, utils.InternalServerError())
		return
	}
	result, err := sched.UpdateSiteDynamicResource(region, &resource)
	if err != nil {
		klog.Errorf("Schedule failed!, err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusInternalServerError, utils.InternalServerWithError(err.Error()))
		return
	}

	resourceResp := types.SiteResourceRes{Result: result}
	resp.WriteHeaderAndEntity(http.StatusCreated, resourceResp)
}
