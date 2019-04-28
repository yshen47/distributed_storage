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
	abortChannel          chan string    //transactionID in the channel
	globalResources       *ResourceMap   //serverIdentifier->objectName->transactionID
	transactionDependency *DependencyMap //key depend on value, key and value is transactionID
	serverConnection	  []NodeClient
}

func (c *Coordinator)Init(serverConn []*NodeClient) {
	c.globalResources = new(ResourceMap)
	c.globalResources.Init()
	c.transactionDependency = new(DependencyMap)
	c.abortChannel = make(chan string)
	c.serverConnection = make([]NodeClient,5)
	for i := 0; i< len(serverConn); i++{
		c.serverConnection = append(c.serverConnection, *serverConn[i])
	}
}

func (*Coordinator) OpenTransaction(ctx context.Context, req *Empty) (*Transaction, error) {
	transactionID := utils.Concatenate(rand.Intn(1000000), int(time.Now().Unix()))
	return &Transaction{Id:&transactionID}, nil

}

func (*Coordinator) CloseTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseTransaction not implemented")
}

func (c *Coordinator) AskCommitTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskCommitTransaction not implemented")
}

func (*Coordinator) AskAbortTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskAbortTransaction not implemented")
}

func (c *Coordinator) TryLock(ctx context.Context, req *TryLockParam) (*Feedback, error) {
	fmt.Println("received new trylock request with transactionID: ", *req.TransactionID, ", server:", *req.ServerIdentifier, ", object:", *req.Object)
	resourceKey := c.globalResources.ConstructKey(*req)
	//if c.globalResources.Has(resourceKey) {
	//	origValues := c.globalResources.Get(resourceKey)
	//	c.transactionDependency.Set(*req.TransactionID, origValues[0])
	//}
	if c.globalResources.TryLockAt(*req, c.abortChannel, c) {
		message := "Success"
		fmt.Println("Got the lock with param: ", *req.TransactionID)
		fmt.Println(c.globalResources.Get(resourceKey).owners)
		fmt.Println("=================")
		//time.Sleep(10 * time.Second)
		return &Feedback{Message:&message}, nil
	} else {
		message := "Abort"
		fmt.Println("Abort the lock with param: ", *req)
		return &Feedback{Message:&message}, status.Errorf(codes.Aborted, "transaction aborted, found deadlock!")
	}

}

func (c*Coordinator) ReportUnlock(ctx context.Context, req *ReportUnLockParam) (*Empty, error) {
	c.globalResources.Delete(*req)
	resourceKey := utils.Concatenate(*req.ServerIdentifier, "_", *req.Object)
	if *req.LockType == "W" {
		c.globalResources.Get(resourceKey).lock.Unlock()
	} else {
		c.globalResources.Get(resourceKey).lock.RUnlock()
	}

	fmt.Println("Unlock with param: ", *req.TransactionID)
	fmt.Println(c.globalResources.Get(resourceKey).owners)
	fmt.Println("=================")
	return &Empty{}, nil
}

func (c *Coordinator) CheckDeadlock(param TryLockParam) bool {
	fmt.Println("Dead lock detection: ")
	resourceKey := c.globalResources.ConstructKey(param)
	if c.globalResources.Has(resourceKey) {
		for _, owner := range c.globalResources.Get(resourceKey).owners {
			if !(owner.lockType == "R" && *param.LockType == "R") {
				if c.checkDeadlockHelper(*param.TransactionID, owner.transactionID) {
					return true
				}
			}
		}
	}
	return false
}

func (c *Coordinator) checkDeadlockHelper(targetID string, currID string) bool {
	fmt.Println("Current ID:", currID)
	for _, nextID := range c.transactionDependency.Get(currID) {
		fmt.Println("Next ID: ", nextID)
		if nextID == targetID {
			return true
		}
		if c.checkDeadlockHelper(targetID, nextID) {
			return true
		}
	}
	return false
}
