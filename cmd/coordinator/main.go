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
	go dial()
	coordinator := server.Coordinator{}
	coordinator.Init()
	server.RegisterCoordinatorServer(nodeServer, &coordinator)
	err = nodeServer.Serve(lis)
	utils.CheckError(err)

	fmt.Println("End Process.")
}


func dial(){
	serverConn := make([] server.NodeClient, 5)
	serverPorts := [5]string {"5600", "5700", "5800", "5900", "6000"}
	for i := 0; i<5; i++ {
		ipaddr := utils.Concatenate("127.0.0.1",":",serverPorts[i])
		fmt.Println("Dial ", ipaddr)
		conn, err := grpc.Dial(ipaddr, grpc.WithBlock(), grpc.WithInsecure())
		utils.CheckError(err)
		serverConn[i] = server.NewNodeClient(conn)
	}
	fmt.Println("Successfully dialed all of the servers.")
}