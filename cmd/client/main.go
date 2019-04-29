package main

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"mp3/server"
	"mp3/utils"
	"os"
	"strconv"
	"strings"
)

func main() {

	serverPorts := [5]string {"6000", "6100", "6200", "6300", "6400"}
	coordPort := "6500"

	var serverConn [] server.NodeClient = make([] server.NodeClient, 5)
	for i := 0; i<5; i++ {
		ipaddr := utils.Concatenate("127.0.0.1",":",serverPorts[i])
		fmt.Println("Dial server ", ipaddr)
		conn, err := grpc.Dial(ipaddr, grpc.WithInsecure(), grpc.WithBlock())
		utils.CheckError(err)
		serverConn[i] = server.NewNodeClient(conn)
	}
	coordAddr := utils.Concatenate("127.0.0.1",":",coordPort)
	fmt.Println("Dial coordinator")
	conn, error := grpc.Dial(coordAddr, grpc.WithInsecure(), grpc.WithBlock())
	coordConn := server.NewCoordinatorClient(conn)
	utils.CheckError(error)
	currTransactionID, err := coordConn.OpenTransaction(context.Background(),&server.Empty{})
	utils.CheckError(err)


	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, error := reader.ReadString('\n')
		if error != nil {
			fmt.Println("Invalid input. Re-enter!")
			continue
		}
		words := strings.Fields(text)
		if len(words) < 1 {
			fmt.Println("Invalid input. Re-enter!")
			continue
		}
		cmd := words[0]

		if cmd == "COMMIT"{
			feedback, error :=coordConn.AskCommitTransaction(context.Background(),currTransactionID)
			utils.CheckError(error)
			fmt.Println(*feedback.Message)
			break
		}else if cmd == "BEGIN"{
			fmt.Println(currTransactionID)
		}else if cmd == "ABORT" {
			feedback, error :=coordConn.AskAbortTransaction(context.Background(),currTransactionID)
			utils.CheckError(error)
			fmt.Println(*feedback.Message)
			break
		}else if cmd == "SET" || cmd == "GET"{
			val := strings.Split(words[1],".")
			if len(val) == 0 {
				fmt.Println("command format: [COMMIT/ABORT/GET/SET] [Server.Obj]")
			}
			if temp, ok := strconv.Atoi(val[0]); ok==nil{
				if idx := (temp-6000)/100; idx < 5 {
					if cmd == "SET" {
						setparam := server.SetParams{}
						setparam.ObjectName = &val[1]
						setparam.TransactionID = currTransactionID.Id
						setparam.ServerIdentifier = &val[0]
						setparam.Value = &words[2]
						feedback, err := serverConn[idx].ClientSet(context.Background(),&setparam)
						//fmt.Println("idx = ",idx)
						s, _ := status.FromError(err)
						if s.Code().String() == "Aborted" {
							fmt.Println("ABORTED")
						}
						if err == nil{
							fmt.Println(*feedback.Message)
						}

					}else {
						getparam := server.GetParams{}
						getparam.TransactionID = currTransactionID.Id
						getparam.ServerIdentifier = &val[0]
						getparam.ObjectName = &val[1]
						feedback,err := serverConn[idx].ClientGet(context.Background(),&getparam)
						if err != nil{
							fmt.Println("error!", err)
						}else{
							fmt.Println(*feedback.Message)
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
	fmt.Println("End Process.")
}

