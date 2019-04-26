package main

import (
	"fmt"
	"google.golang.org/grpc"
	"mp3/server"
	"mp3/utils"
	"context"
)

func main() {

	serverPorts := [5]string {"5600", "5700", "5800", "5900", "6000"}
	coordPort := "6100"

	var serverConn [] server.NodeClient = make([] server.NodeClient, 5)
	for i := 0; i<5; i++ {
		ipaddr := utils.Concatenate("127.0.0.1",":",serverPorts[i])
		conn, err := grpc.Dial(ipaddr, grpc.WithInsecure())
		utils.CheckError(err)
		serverConn[i] = server.NewNodeClient(conn)
	}
	coordAddr := utils.Concatenate("127.0.0.1",":",coordPort)
	conn, error := grpc.Dial(coordAddr, grpc.WithInsecure())
	coordConn := server.NewCoordinatorClient(conn)
	utils.CheckError(error)
	transactionID, err := coordConn.OpenTransaction(context.Background(),&server.Empty{})
	utils.CheckError(err)
	fmt.Println(transactionID)


}

