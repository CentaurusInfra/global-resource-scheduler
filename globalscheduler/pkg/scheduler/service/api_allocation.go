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
	"fmt"
	"k8s.io/klog"
	"net/http"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"

	"github.com/emicklei/go-restful"
	"github.com/satori/go.uuid"
)

// Scheduling request
func Allocations(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	allocationReq := new(types.AllocationReq)

	err := req.ReadEntity(&allocationReq)
	if err != nil {
		klog.Errorf("Failed to unmarshall allocation from request body, err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(err.Error()))
		return
	}
	klog.Infof("ReqAllocation : %s", utils.GetJSONString(allocationReq))

	allocation := allocationReq.Allocation
	if allocation.ID == "" {
		allocation.ID = uuid.NewV4().String()
	}

	err = CheckIDOrName(allocation)
	if err != nil {
		klog.Errorf("name/id of allocation is invalid, err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest,
			utils.RequestBodyParamInvalid("name/id of allocation is invalid"))
		return
	}

	if allocation.Replicas <= 0 {
		errMsg := fmt.Sprintf("allocation.Replicas(%d) is invalid", allocation.Replicas)
		klog.Error(errMsg)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(errMsg))
		return
	}

	// CheckStrategy
	err = CheckStrategy(allocation)
	if err != nil {
		klog.Errorf("CheckStrategy failed! err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(err.Error()))
		return
	}

	// Check whether the flavor is valid.
	err = CheckFlavor(allocation)
	if err != nil {
		klog.Errorf("CheckFlavor failed! err : %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(err.Error()))
		return
	}

	// Check whether the storage is valid.
	err = CheckStorage(allocation)
	if err != nil {
		klog.Errorf("CheckStorage failed! err : %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusBadRequest, utils.RequestBodyParamInvalid(err.Error()))
		return
	}

	allocationResult := types.AllocationResult{ID: allocation.ID, Stack: []types.RespStack{}}
	sched := scheduler.GetScheduler()
	if sched == nil {
		klog.Errorf("Scheduler is not init, please wait...")
		utils.WriteFailedJSONResponse(resp, http.StatusInternalServerError, utils.InternalServerError())
		return
	}
	result, err := sched.Schedule2(ctx, &allocation)
	if err != nil {
		klog.Errorf("Schedule failed!, err: %s", err)
		utils.WriteFailedJSONResponse(resp, http.StatusInternalServerError, utils.InternalServerWithError(err.Error()))
		return
	}

	for _, newStack := range result.Stacks {
		respStack := types.RespStack{Name: newStack.Name, Selected: newStack.Selected}
		for _, server := range newStack.Resources {
			respServer := types.RespResource{Name: server.Name,
				FlavorID: server.FlavorIDSelected, Count: server.Count}
			respStack.Resources = append(respStack.Resources, respServer)
		}

		allocationResult.Stack = append(allocationResult.Stack, respStack)
	}

	allocationResp := types.AllocationsResp{AllocationResult: allocationResult}
	resp.WriteHeaderAndEntity(http.StatusCreated, allocationResp)
}
