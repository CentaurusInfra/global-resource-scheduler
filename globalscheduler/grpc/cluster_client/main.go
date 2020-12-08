// Package main implements a client .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"k8s.io/klog"
	pb "k8s.io/kubernetes/globalscheduler/grpc/cluster"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
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
				{FlavorID: 5, //1:Small, 2:Medium, 3:Large, 4:Xlarge, 5:2xLarge
					TotalCapacity: 10,
				},
			},
			Storage: []*pb.ClusterProfile_ClusterSpecInfo_StorageInfo{
				{
					TypeID:          pb.ClusterProfile_ClusterSpecInfo_SSD,
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
		klog.Infof("cluster profile is not correct - %v", clusterProfile)
		errorMessage = fmt.Errorf("cluster profile is not correct - %v", clusterProfile)
	}
	if errorMessage != nil {
		log.Printf("could not greet: %v", errorMessage)
	}

	log.Printf("Returned Client")
}
