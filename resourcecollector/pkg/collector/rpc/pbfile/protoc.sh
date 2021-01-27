#! /bin/sh
protoc --go_out=plugins=grpc:. cluster.proto
protoc --go_out=plugins=grpc:. clusterstate.proto