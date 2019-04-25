package main

import (
	"google.golang.org/grpc"
	"mp3/server"
	"mp3/utils"
	"net"
)

func main() {
	portNum := 6100
	lis, err := net.Listen("tcp", utils.Concatenate(":", portNum))
	utils.CheckError(err)
	nodeServer := grpc.NewServer()
	server.RegisterCoordinatorServer(nodeServer, &server.Coordinator{})
	err = nodeServer.Serve(lis)
	utils.CheckError(err)
}

