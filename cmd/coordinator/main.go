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
	// connect to server
	serverPorts := [5]string {"5600", "5700", "5800", "5900", "6000"}
	var serverConn [] server.NodeClient = make([] server.NodeClient, 5)
	for i := 0; i<5; i++ {
		ipaddr := utils.Concatenate("127.0.0.1",":",serverPorts[i])
		conn, err := grpc.Dial(ipaddr, grpc.WithInsecure())
		utils.CheckError(err)
		serverConn[i] = server.NewNodeClient(conn)
	}
}

