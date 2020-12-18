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
	"os"
	"time"

	"google.golang.org/grpc"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto"
)

const (
	address = "localhost:50052"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewClusterProtocolClient(conn)
	clusterProfile := &pb.ClusterProfile{
		ClusterNameSpace: "default",
		ClusterName:      "cluster1",
		ClusterSpec: &pb.ClusterProfile_ClusterSpecInfo{
			ClusterIpAddress: "10.0.0.1",
			GeoLocation: &pb.ClusterProfile_ClusterSpecInfo_GeoLocationInfo{
				City:     "Bellevue",
				Province: "WA",
				Area:     "west-1",
				Country:  "us",
			},
			Region: &pb.ClusterProfile_ClusterSpecInfo_RegionInfo{
				Region:           "us-west-1",
				AvailabilityZone: "us-west",
			},
			Operator: &pb.ClusterProfile_ClusterSpecInfo_OperatorInfo{
				Operator: "futurewei",
			},
			Flavor: []*pb.ClusterProfile_ClusterSpecInfo_FlavorInfo{
				{FlavorID: "Large", //1:Small, 2:Medium, 3:Large, 4:Xlarge, 5:2xLarge
					TotalCapacity: 10,
				},
			},
			Storage: []*pb.ClusterProfile_ClusterSpecInfo_StorageInfo{
				{
					TypeID:          "ssd",
					StorageCapacity: 2048,
				},
			},
			EipCapacity:   10,
			CPUCapacity:   7,
			MemCapacity:   2048,
			ServerPrice:   50,
			HomeScheduler: "scheduler-1",
		},
	}

	// Contact the server and request crd services.
	var arg string
	if len(os.Args) > 1 {
		arg = os.Args[1]
		fmt.Printf("Request %s", arg)
	} else {
		fmt.Print("Request service name is missing. example format: main.go cluster")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var errorMessage error
	switch arg {
	case "cluster":
		result, err := c.SendClusterProfile(ctx, clusterProfile)
		errorMessage = err
		fmt.Printf("return result = %v", result)
	default:
		fmt.Infof("cluster profile is not correct - %v", clusterProfile)
		errorMessage = fmt.Errorf("cluster profile is not correct - %v", clusterProfile)
	}
	if errorMessage != nil {
		fmt.Printf("could not greet: %v", errorMessage)
	}
	fmt.Printf("Returned Client")
}
