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
	"flag"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	clusterclient "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
)

const (
	Port             = ":50053"
	ReturnError      = 0
	ReturnOk         = 1
	StateReady       = 1
	StateDown        = 2
	StateUnreachable = 3
)

var (
	masterURL  string
	kubeconfig string
)

// ApiServer : Empty API server struct
type ResourceCollectorProtocolServer struct{}

// services - Send cluster profile
func (s *ResourceCollectorProtocolServer) UpdateClusterStatus(ctx context.Context, in *pb.ClusterState) (*pb.ReturnMessageClusterState, error) {
	klog.Infof("Received Request - UpdateClusterStatus: %v", in)
	ns := in.NameSpace
	name := in.Name
	state := in.State
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Errorf("error getting client config: %s", err.Error())
	}

	//1. cluster clientset
	clusterClientset, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("error - building global scheduler cluster client: %s", err.Error())
	}

	//2. cluster client - create a cluster api client interface for cluster v1.
	clusterClient, err := clusterclient.NewClusterClient(clusterClientset, ns)
	if err != nil {
		klog.Errorf("error - create a cluster client: %s", err.Error())
	}
	opts := metav1.GetOptions{}
	cluster, err := clusterClient.Get(name, opts)
	if err != nil || cluster == nil {
		klog.Errorf("error - update a cluster: %s", err.Error())
		return &pb.ReturnMessageClusterState{NameSpace: ns, Name: name, ReturnCode: ReturnError}, nil
	}
	switch state {
	case StateReady:
		{
			cluster.Status = "Ready"
			break
		}
	case StateDown:
		{
			cluster.Status = "Down"
			break
		}
	case StateUnreachable:
		{
			cluster.Status = "Unreachable"
			break
		}
	default:
		{
			cluster.Status = "Unknown"
			break
		}
	}
	cluster, err = clusterClient.Update(cluster)
	if err != nil {
		klog.Errorf("error - update a cluster: %s", err.Error())
		return &pb.ReturnMessageClusterState{NameSpace: ns, Name: name, ReturnCode: ReturnError}, nil

	} else {
		klog.Info("updated cluster state: %v", cluster)
	}
	return &pb.ReturnMessageClusterState{NameSpace: ns, Name: name, ReturnCode: ReturnOk}, nil
}

func main() {
	flag.Parse()
	defer klog.Flush()

	fmt.Print("Server started, Port: " + Port)
	lis, err := net.Listen("tcp", Port)
	if err != nil {
		klog.Printf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterResourceCollectorProtocolServer(s, &ResourceCollectorProtocolServer{})
	if err := s.Serve(lis); err != nil {
		klog.Printf("failed to serve: %v", err)
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
