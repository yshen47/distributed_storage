package main

import (
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"fmt"
)

func main() {
	serverPorts := [5]string {"5600", "5700", "5800", "5900", "6000"}
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRouteGuideServer(grpcServer, &routeGuideServer{})
	... // determine whether to use TLS
	grpcServer.Serve(lis)
}

