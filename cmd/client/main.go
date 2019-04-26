package main

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mp3/server"
	"mp3/utils"
	"os"
	"strconv"
	"strings"
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

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		words := strings.Fields(text)
		cmd := words[0]
		val := strings.Split(words[1],".")
		if cmd == "COMMIT"{
			coordConn.CommitTransaction(context.Background(),&server.Empty{})
		}else if cmd == "ABORT" {
			coordConn.AbortTransaction(context.Background(),&server.Empty{})
		}else if cmd == "SET" || cmd == "GET"{
			if len(val) == 0 {
				fmt.Println("command format: [COMMIT/ABORT/GET/SET] [Server.Obj]")
			}
			if temp, ok := strconv.Atoi(val[0]); ok==nil{
				if idx := (temp-5600)/100; idx < 5 {
					if cmd == "SET" {
						setparam := server.SetParams{}
						setparam.ObjectName = &val[1]
						setparam.ServerIdentifier = &val[0]
						setparam.Value = &words[2]
						feedback,err := serverConn[idx].ClientSet(context.Background(),&setparam)
						fmt.Println("idx = ",idx)
						if err != nil{
							fmt.Println("error!", err)
						}else{
							fmt.Println(feedback.Message)
						}

					}else {
						getparam := server.GetParams{}
						getparam.ServerIdentifier = &val[0]
						getparam.ObjectName = &val[1]
						feedback,err := serverConn[idx].ClientGet(context.Background(),&getparam)
						if err != nil{
							fmt.Println("error!", err)
						}else{
							fmt.Println(feedback.Message)
						}

					}
				}
			}else{
				fmt.Println("command format: [COMMIT/ABORT/GET/SET] [Server.Obj]")
			}
		}else {
			fmt.Println("command format: [COMMIT/ABORT/GET/SET] [Server.Obj]")
		}
	}
}

