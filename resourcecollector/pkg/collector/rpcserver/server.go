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
	"google.golang.org/grpc"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/common/config"
	"net"
)

func NewRpcServer() {
	lis, err := net.Listen("tcp", "127.0.0.1:"+config.GlobalConf.RpcPort)
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}
	logger.Infof("gRPC Server started, Port: " + config.GlobalConf.RpcPort)

	s := grpc.NewServer()
	pb.RegisterClusterProtocolServer(s, &ClusterProtocolServer{})

	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}
