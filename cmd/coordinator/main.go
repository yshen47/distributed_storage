package main

import (
	"fmt"
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
	coordinator := server.Coordinator{}
	coordinator.Init()

	go dial(&coordinator)

	server.RegisterCoordinatorServer(nodeServer, &coordinator)
	err = nodeServer.Serve(lis)
	utils.CheckError(err)

	fmt.Println("End Process.")
}

func dial(coordinator *server.Coordinator){
	serverConn := make([] server.NodeClient, 5)
	serverPorts := [5]string {"5600", "5700", "5800", "5900", "6000"}
	for i := 0; i<5; i++ {
		ipaddr := utils.Concatenate("127.0.0.1",":",serverPorts[i])
		fmt.Println("Dial ", ipaddr)
		conn, err := grpc.Dial(ipaddr,  grpc.WithInsecure(), grpc.WithBlock())
		utils.CheckError(err)
		temp := server.NewNodeClient(conn)
		serverConn[i] = temp
	}
	coordinator.ServerConnection = serverConn
	fmt.Println("Successfully dialed all of the servers.")
}