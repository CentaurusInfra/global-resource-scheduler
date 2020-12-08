package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "k8s.io/arktos/globalscheduler/grpc/cluster"
)

const (
	port = ":50051"
)

// ApiServer : Empty API server struct
type ClusterProtocolServer struct{}

// services - Send cluster profile
func (s *ClusterProtocolServer) SendClusterProfile(ctx context.Context, in *pb.ClusterProfile) (*pb.ReturnMessage, error) {
	log.Printf("Received Request - Create: %v", in)
	ns := in.ClusterNameSpace
	name := in.ClusterName
	//result := (int32)1
	return &pb.ReturnMessage{ClusterNameSpace: ns, ClusterName: name, ReturnCode: 1}, nil
}

func main() {
	fmt.Print("Server started, Port: " + port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	pb.RegisterClusterProtocolServer(s, &ClusterProtocolServer{})
	if err := s.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
	}
}
