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
	lis, err := net.Listen("tcp", utils.Concatenate(":", portNum))
	utils.CheckError(err)
	nodeServer := grpc.NewServer()
	server.RegisterNodeServer(nodeServer, &server.Node{portNum, make(map[string]string)})

	err = nodeServer.Serve(lis)
	utils.CheckError(err)
}

