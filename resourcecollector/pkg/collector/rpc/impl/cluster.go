package impl

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
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
	region := in.ClusterNameSpace
	az := in.ClusterName
	regionAZ := fmt.Sprintf("%s|%s", region, az)
	ip := in.ClusterSpec.ClusterIpAddress
	if ip == "" {
		return nil, errors.New("cluster ip is not set")
	}

	switch in.Status {
	case ClusterStatusCreated:
		logger.Infof("grpc.GrpcSendClusterProfile created- regionAZ[%s], IP[%s]", regionAZ, ip)
	case ClusterStatusDeleted:
		logger.Infof("grpc.GrpcSendClusterProfile deleted- regionAZ[%s], IP[%s]", regionAZ, ip)
	default:
		logger.Infof("grpc.GrpcSendClusterProfile status error[%s]- regionAZ[%s], IP[%s]", in.Status, regionAZ, ip)
		return nil, errors.New("status error")
	}

	resp := &pbfile.ReturnMessageClusterProfile{
		ClusterNameSpace: region,
		ClusterName:      az,
		ReturnCode:       pbfile.ReturnMessageClusterProfile_OK,
		Message:          "ok",
	}
	return resp, nil
}
