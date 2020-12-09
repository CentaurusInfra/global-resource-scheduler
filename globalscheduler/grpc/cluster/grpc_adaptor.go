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

package cluster

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/klog"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

type EventType string

const (
	port = "50052"

	ReturnError = 0
	ReturnOK    = 1
)

// GrpcCreateCluster is to invoking grpc func of SendClusterProfile
func GrpcSendClusterProfile(grpcHost string, cluster *clusterv1.Cluster) *pb.ReturnMessage {
	client, ctx, conn, cancel, err := getGrpcClient(grpcHost)
	if err != nil {
		return getReturnMessageFromError(cluster.ObjectMeta.Namespace, cluster.ObjectMeta.Name, &err)
	}
	defer conn.Close()
	defer cancel()
	returnMessage, err := client.SendClusterProfile(ctx, ConvertClusterToClusterProfile(cluster))
	if err != nil {
		return getReturnMessageFromError(cluster.ObjectMeta.Namespace, cluster.ObjectMeta.Name, &err)
	}
	return returnMessage
}

//Convert cluster to grpc cluster message format
func ConvertClusterToClusterProfile(cluster *clusterv1.Cluster) *pb.ClusterProfile {
	var flavors []*pb.ClusterProfile_ClusterSpecInfo_FlavorInfo
	if cluster.Spec.Flavors == nil {
		flavors = nil
	} else {
		for i := 0; i < len(cluster.Spec.Flavors); i++ {
			flavors[i].FlavorID = cluster.Spec.Flavors[i].FlavorID
			flavors[i].TotalCapacity = cluster.Spec.Flavors[i].TotalCapacity
		}
	}
	var storages []*pb.ClusterProfile_ClusterSpecInfo_StorageInfo
	if cluster.Spec.Storage == nil {
		storages = nil
	} else {
		for i := 0; i < len(cluster.Spec.Storage); i++ {
			storages[i].TypeID = cluster.Spec.Storage[i].TypeID
			storages[i].StorageCapacity = cluster.Spec.Storage[i].StorageCapacity
		}
	}

	clusterProfile := &pb.ClusterProfile{
		ClusterNameSpace: cluster.ObjectMeta.Namespace,
		ClusterName:      cluster.ObjectMeta.Name,
		ClusterSpec: &pb.ClusterProfile_ClusterSpecInfo{
			ClusterIpAddress: cluster.Spec.IpAddress,
			GeoLocation: &pb.ClusterProfile_ClusterSpecInfo_GeoLocationInfo{
				City:     cluster.Spec.GeoLocation.City,
				Province: cluster.Spec.GeoLocation.Province,
				Area:     cluster.Spec.GeoLocation.Area,
				Country:  cluster.Spec.GeoLocation.Country,
			},
			Region: &pb.ClusterProfile_ClusterSpecInfo_RegionInfo{
				Region:           cluster.Spec.Region.Region,
				AvailabilityZone: cluster.Spec.Region.AvailabilityZone[0],
			},
			Operator: &pb.ClusterProfile_ClusterSpecInfo_OperatorInfo{
				Operator: cluster.Spec.Operator.Operator,
			},
			Flavor:        flavors,
			Storage:       storages,
			EipCapacity:   cluster.Spec.EipCapacity,
			CPUCapacity:   cluster.Spec.CPUCapacity,
			MemCapacity:   cluster.Spec.MemCapacity,
			ServerPrice:   cluster.Spec.ServerPrice,
			HomeScheduler: cluster.Spec.HomeScheduler,
		},
	}
	klog.Infof("Cluster controller is sending cluster profile info to ResourceCollector %v", clusterProfile)
	return clusterProfile
}

func getGrpcClient(grpcHost string) (pb.ClusterProtocolClient, context.Context, *grpc.ClientConn, context.CancelFunc, error) {
	address := fmt.Sprintf("%s:%s", grpcHost, port)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, conn, nil, err
	}

	client := pb.NewClusterProtocolClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	return client, ctx, conn, cancel, nil
}

func getReturnMessageFromError(ns, name string, err *error) *pb.ReturnMessage {
	return &pb.ReturnMessage{
		ClusterNameSpace: ns,
		ClusterName:      name,
		ReturnCode:       ReturnError,
		Message:          fmt.Sprintf("Grpc call failed: %s", (*err).Error()),
	}
}
