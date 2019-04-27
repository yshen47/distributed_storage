package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"mp3/utils"
	"time"
)

// Coordinator can be embedded to have forward compatible implementations.
type Coordinator struct {
	abortChannel chan string //transactionID in the channel
	globalResources	*ResourceMap //serverIdentifier->objectName->transactionID
}

func (c *Coordinator)Init() {
	c.globalResources = new(ResourceMap)
	c.abortChannel = make(chan string)
}

func (*Coordinator) OpenTransaction(ctx context.Context, req *Empty) (*Transaction, error) {
	transactionID := utils.Concatenate(rand.Intn(1000000), int(time.Now().Unix()))
	return &Transaction{Id:&transactionID}, nil

}

func (*Coordinator) CloseTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseTransaction not implemented")
}

func (*Coordinator) AskCommitTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskCommitTransaction not implemented")
}

func (*Coordinator) AskAbortTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskAbortTransaction not implemented")
}

func (c *Coordinator) TryLock(ctx context.Context, req *TryLockParam) (*Feedback, error) {
	fmt.Println("received new trylock request with param: ", *req)
	c.globalResources.TryLockAt(*req, c.abortChannel)
	return nil, status.Errorf(codes.Unimplemented, "method TryLock not implemented")
}

func (*Coordinator) ReportUnlock(ctx context.Context, req *ReportUnLockParam) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportUnlock not implemented")
}

//func (c *Coordinator) checkDeadlock() bool {
//
//}