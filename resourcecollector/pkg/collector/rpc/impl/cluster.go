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

package impl

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/resourcecollector/pkg/collector"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/rpc/pbfile"
)

const (
	ClusterStatusCreated = "Created"
	ClusterStatusDeleted = "Deleted"
)

type ClusterProtocolServer struct{}

// services - Send cluster profile
func (s *ClusterProtocolServer) SendClusterProfile(ctx context.Context,
	in *pbfile.ClusterProfile) (*pbfile.ReturnMessageClusterProfile, error) {
	logger.Infof("Received RPC Request")
	ns := in.ClusterNameSpace
	name := in.ClusterName
	siteID := fmt.Sprintf("%s|%s", ns, name)
	ip := in.ClusterSpec.ClusterIpAddress
	if ip == "" {
		err := errors.New("cluster ip is not set")
		return getReturnMessageFromError(ns, name, &err), err
	}

	col := collector.GetCollector()
	if col == nil {
		logger.Error(ctx, "get new collector failed")
		err := errors.New("cluster ip is not set")
		return getReturnMessageFromError(ns, name, &err), err
	}
	switch in.Status {
	case ClusterStatusCreated:
		logger.Infof("grpc.GrpcSendClusterProfile created- siteID[%s], IP[%s]", siteID, ip)
		col.SiteIPCache.AddSiteIP(siteID, ip)
	case ClusterStatusDeleted:
		logger.Infof("grpc.GrpcSendClusterProfile deleted- siteID[%s], IP[%s]", siteID, ip)
		col.SiteIPCache.RemoveSiteIP(siteID)
	default:
		logger.Infof("grpc.GrpcSendClusterProfile status error[%s]- siteID[%s], IP[%s]", in.Status, siteID, ip)
		err := errors.New("status error")
		return getReturnMessageFromError(ns, name, &err), err
	}

	resp := &pbfile.ReturnMessageClusterProfile{
		ClusterNameSpace: ns,
		ClusterName:      name,
		ReturnCode:       pbfile.ReturnMessageClusterProfile_OK,
		Message:          "ok",
	}
	return resp, nil
}

func getReturnMessageFromError(ns, name string, err *error) *pbfile.ReturnMessageClusterProfile {
	return &pbfile.ReturnMessageClusterProfile{
		ClusterNameSpace: ns,
		ClusterName:      name,
		ReturnCode:       pbfile.ReturnMessageClusterProfile_Error,
		Message:          fmt.Sprintf("Grpc call failed: %s", (*err).Error()),
	}
}
