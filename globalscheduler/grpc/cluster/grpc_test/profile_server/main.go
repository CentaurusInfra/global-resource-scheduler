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

package main

import (
	"context"
	"fmt"
	grpc "google.golang.org/grpc"
	"k8s.io/klog"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	"net"
)

const (
	port = ":50052"
)

// ApiServer : Empty API server struct
type ClusterProtocolServer struct{}

// services - Send cluster profile
func (s *ClusterProtocolServer) SendClusterProfile(ctx context.Context, in *pb.ClusterProfile) (*pb.ReturnMessageClusterProfile, error) {
	klog.Infof("Received Request - Create: %v", in)
	ns := in.ClusterNameSpace
	name := in.ClusterName
	return &pb.ReturnMessageClusterProfile{ClusterNameSpace: ns, ClusterName: name, ReturnCode: 1}, nil
}

func main() {
	klog.Infof("Server started, Port: " + port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		klog.Infof("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterClusterProtocolServer(s, &ClusterProtocolServer{})
	if err := s.Serve(lis); err != nil {
		klog.Infof("failed to serve: %v", err)
	}
}
