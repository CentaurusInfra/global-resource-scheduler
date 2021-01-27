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

package rpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/common/config"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/rpc/impl"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/rpc/pbfile"

	"google.golang.org/grpc"
)

const (
	ReturnError = 0
	ReturnOK    = 1
)

func NewRpcServer() {
	lis, err := net.Listen("tcp", "127.0.0.1:"+config.RpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logger.Infof("gRPC Server started, Port: " + config.RpcPort)

	s := grpc.NewServer()
	pbfile.RegisterClusterProtocolServer(s, &impl.ClusterProtocolServer{})

	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}

func GrpcUpdateClusterStatus(node string, status int64) error {
	grpcHost := config.ClusterControllerIP
	client, ctx, conn, cancel, err := getGrpcClient(grpcHost)
	if err != nil {
		logger.Errorf("Error to make a connection to ClusterController %s", grpcHost)
		return err
	}
	defer conn.Close()
	defer cancel()
	region, az, err := ParseRegionAZ(node)
	if err != nil {
		return err
	}
	req := &pbfile.ClusterState{
		NameSpace: region,
		Name:      az,
		State:     status,
	}
	returnMessage, err := client.UpdateClusterStatus(ctx, req)
	if err != nil {
		logger.Errorf("Error to update cluster status to ClusterController, err: %s", err.Error())
		return err
	}
	if returnMessage.ReturnCode == ReturnError {
		logger.Errorf("ClusterController update cluster status err")
		return errors.New("ClusterController update cluster status err")
	}
	return nil
}

func getGrpcClient(grpcHost string) (pbfile.ResourceCollectorProtocolClient, context.Context, *grpc.ClientConn, context.CancelFunc, error) {
	logger.Infof("get gRPC client: %s", grpcHost)
	address := fmt.Sprintf("%s:%s", grpcHost, config.ClusterControllerPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, conn, nil, err
	}

	client := pbfile.NewResourceCollectorProtocolClient(conn)
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
