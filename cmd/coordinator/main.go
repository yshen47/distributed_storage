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
	serverConn := make([] *server.NodeClient, 5)
	dial(serverConn)
	coordinator := server.Coordinator{}
	coordinator.Init(serverConn)
	server.RegisterCoordinatorServer(nodeServer, &coordinator)
	err = nodeServer.Serve(lis)
	utils.CheckError(err)

	fmt.Println("End Process.")
}


func dial(serverConn [] *server.NodeClient){

	serverPorts := [5]string {"5600", "5700", "5800", "6200", "6000"}
	for i := 0; i<5; i++ {
		ipaddr := utils.Concatenate("127.0.0.1",":",serverPorts[i])
		fmt.Println("Dial ", ipaddr)
		conn, err := grpc.Dial(ipaddr,  grpc.WithInsecure())
		utils.CheckError(err)
		temp := server.NewNodeClient(conn)
		serverConn[i] = &temp
	}
	fmt.Println("Successfully dialed all of the servers.")
}