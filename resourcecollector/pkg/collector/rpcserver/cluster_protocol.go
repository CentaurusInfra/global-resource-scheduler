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

package rpcserver

import (
	"context"
	"errors"
	"fmt"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/resourcecollector/pkg/collector"
)

const (
	ClusterStatusCreated = "Created"
	ClusterStatusDeleted = "Deleted"
)

type ClusterProtocolServer struct{}

// services - Send cluster profile
func (s *ClusterProtocolServer) SendClusterProfile(ctx context.Context,
	in *pb.ClusterProfile) (*pb.ReturnMessageClusterProfile, error) {
	logger.Infof("Received RPC Request")
	ns := in.ClusterNameSpace
	name := in.ClusterName
	siteID := fmt.Sprintf("%s|%s", ns, name)
	if in.ClusterSpec == nil || in.ClusterSpec.GeoLocation == nil || in.ClusterSpec.Operator == nil {
		err := errors.New("ClusterSpec information is incomplete")
		return getReturnMessageFromError(ns, name, &err), err
	}
	ip := in.ClusterSpec.ClusterIpAddress
	if ip == "" {
		err := errors.New("cluster ip is not set")
		return getReturnMessageFromError(ns, name, &err), err
	}

	col, err := collector.GetCollector()
	if err != nil {
		logger.Error(ctx, "get new collector failed, err: %s", err.Error())
		err := errors.New("server internal error")
		return getReturnMessageFromError(ns, name, &err), err
	}
	switch in.ClusterStatus {
	case ClusterStatusCreated:
		logger.Infof("grpc.GrpcSendClusterProfile created- siteID[%s], IP[%s]", siteID, ip)
		siteInfo := &typed.SiteInfo{
			SiteID:           siteID,
			Name:             siteID,
			Region:           ns,
			AvailabilityZone: name,
			Status:           ClusterStatusCreated,
			City:             in.ClusterSpec.GeoLocation.City,
			Province:         in.ClusterSpec.GeoLocation.Province,
			Area:             in.ClusterSpec.GeoLocation.Area,
			Country:          in.ClusterSpec.GeoLocation.Country,
			Operator:         &typed.Operator{Name: in.ClusterSpec.Operator.Operator},
			EipNetworkID:     ip,
			EipTypeName:      "not known",
			EipCidr:          nil,
			SiteAttributes:   nil,
		}
		col.SiteInfoCache.AddSite(siteInfo)
		//informers.InformerFac.SyncOnSiteChange()
	case ClusterStatusDeleted:
		logger.Infof("grpc.GrpcSendClusterProfile deleted- siteID[%s], IP[%s]", siteID, ip)
		col.SiteInfoCache.RemoveSite(siteID)
		err := col.ResourceCache.RemoveSite(siteID)
		if err != nil {
			logger.Errorf("col.ResourceCache.RemoveSite err: %s", err.Error())
		}
		//informers.InformerFac.SyncOnSiteChange()
	default:
		logger.Infof("grpc.GrpcSendClusterProfile status error[%s]- siteID[%s], IP[%s]", in.ClusterStatus, siteID, ip)
		err := errors.New("status error")
		return getReturnMessageFromError(ns, name, &err), err
	}

	resp := &pb.ReturnMessageClusterProfile{
		ClusterNameSpace: ns,
		ClusterName:      name,
		ReturnCode:       pb.ReturnMessageClusterProfile_OK,
		Message:          "ok",
	}
	return resp, nil
}

func getReturnMessageFromError(ns, name string, err *error) *pb.ReturnMessageClusterProfile {
	return &pb.ReturnMessageClusterProfile{
		ClusterNameSpace: ns,
		ClusterName:      name,
		ReturnCode:       pb.ReturnMessageClusterProfile_Error,
		Message:          fmt.Sprintf("Grpc call failed: %s", (*err).Error()),
	}
}
