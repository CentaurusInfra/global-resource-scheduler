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

package  state_server

import (
	"context"
	"fmt"
	"net"

	grpc "google.golang.org/grpc"
	"k8s.io/klog"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
)

const (
	Port = ":50053"
)

// ApiServer : Empty API server struct
type ResourceCollectorProtocolServer struct{}

// services - Send cluster profile
func (s *ResourceCollectorProtocolServer) UpdateClusterStatus(ctx context.Context, in *pb.ClusterState) (*pb.ReturnMessageClusterState, error) {
	klog.Infof("Received Request - Update Status: %v", in)
	ns := in.NameSpace
	name := in.Name
	state := in.State
	klog.Infof("received state: %v, %v, %v", ns, name, state)
	return &pb.ReturnMessageClusterState{NameSpace: ns, Name: name, ReturnCode: 1}, nil
}

func main() {
	klog.Infof("Server started, Port: " + Port)
	lis, err := net.Listen("tcp", Port)
	if err != nil {
		klog.Infof("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterResourceCollectorProtocolServer(s, &ResourceCollectorProtocolServer{})
	if err := s.Serve(lis); err != nil {
		klog.Infof("failed to serve: %v", err)
	}
}
