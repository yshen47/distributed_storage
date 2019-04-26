package main

import (
	"fmt"
	"google.golang.org/grpc"
	"mp3/server"
	"mp3/utils"
	"net"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Print("Usage: go run main.go [port] \n")
		return
	}
	portNum, err := strconv.Atoi(os.Args[1])

	coordAddr := utils.Concatenate("127.0.0.1",":","6100")
	conn, error := grpc.Dial(coordAddr, grpc.WithInsecure())
	coordConn := server.NewCoordinatorClient(conn)
	utils.CheckError(error)

	node := server.Node{}
	node.Init()
	node.Name = strconv.Itoa(portNum)
	node.CoordinatorDelegate = coordConn

	lis, err := net.Listen("tcp", utils.Concatenate(":", portNum))
	utils.CheckError(err)
	nodeServer := grpc.NewServer()
	server.RegisterNodeServer(nodeServer, &node)
	err = nodeServer.Serve(lis)
	utils.CheckError(err)






}

