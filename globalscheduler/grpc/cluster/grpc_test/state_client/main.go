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

// Package main implements a client .
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"k8s.io/klog"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
	"os"
	"time"
)

const (
	Address = "localhost:50053"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		klog.Infof("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewResourceCollectorProtocolClient(conn)
	clusterState := &pb.ClusterState{
		NameSpace: "default",
		Name:      "cluster2",
		State:     1,
	}

	// Contact the server and request crd services.
	var arg string
	if len(os.Args) > 1 {
		arg = os.Args[1]
		klog.Infof("Request %s", arg)
	} else {
		klog.Infof("Request service name is missing. example format: main.go cluster")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var errorMessage error
	switch arg {
	case "state":
		result, err := c.UpdateClusterStatus(ctx, clusterState)
		errorMessage = err
		klog.Infof("return result = %v", result)
	default:
		klog.Infof("cluster state is not correct - %v", clusterState)
		errorMessage = fmt.Errorf("cluster state is not correct - %v", clusterState)
	}
	if errorMessage != nil {
		klog.Infof("could not greet: %v", errorMessage)
	}
	klog.Infof("Returned ClusterState Client")
}
