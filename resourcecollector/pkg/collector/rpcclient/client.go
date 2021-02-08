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

package rpcclient

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/klog"
	"strings"
	"time"

	"google.golang.org/grpc"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/common/config"
)

const (
	ReturnError      = 0
	ReturnOK         = 1
	StateReady       = 1
	StateDown        = 2
	StateUnreachable = 3
)

func GrpcUpdateClusterStatus(siteID string, state int64) error {
	grpcHost := config.GlobalConf.ClusterControllerIP
	client, ctx, conn, cancel, err := getGrpcClient(grpcHost)
	if err != nil {
		klog.Errorf("Error to make a connection to ClusterController %s", grpcHost)
		return err
	}
	defer conn.Close()
	defer cancel()
	region, az, err := ParseRegionAZ(siteID)
	if err != nil {
		return err
	}
	req := &pb.ClusterState{
		NameSpace: region,
		Name:      az,
		State:     state,
	}
	returnMessage, err := client.UpdateClusterStatus(ctx, req)
	if err != nil {
		klog.Errorf("Error to update cluster status to ClusterController, err: %s", err.Error())
		return err
	}
	if returnMessage.ReturnCode == ReturnError {
		klog.Errorf("ClusterController update cluster status err")
		return errors.New("ClusterController update cluster status err")
	}
	return nil
}

func getGrpcClient(grpcHost string) (pb.ResourceCollectorProtocolClient, context.Context, *grpc.ClientConn, context.CancelFunc, error) {
	klog.Infof("get gRPC client: %s", grpcHost)
	address := fmt.Sprintf("%s:%s", grpcHost, config.GlobalConf.ClusterControllerPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, conn, nil, err
	}

	client := pb.NewResourceCollectorProtocolClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	return client, ctx, conn, cancel, nil
}

func ParseRegionAZ(regionAZ string) (region string, az string, err error) {
	strs := strings.Split(regionAZ, "|")
	if len(strs) != 2 {
		return "", "", errors.New("invalid node format")
	}
	return strs[0], strs[1], nil
}
