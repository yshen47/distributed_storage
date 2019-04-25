package main

import (
	"flag"
	"google.golang.org/grpc"
	"log"
	"mp3/server"
	"mp3/utils"
	"net"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", ":5600")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	coordinator.RegisterCoordinatorServer(grpcServer, &server.Coordinator{})
	err = grpcServer.Serve(lis)
	utils.CheckError(err)
}

