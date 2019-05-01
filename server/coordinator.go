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
	globalResources       *ResourceMap   //serverIdentifier->objectName->transactionID
	transactionDependency *DependencyMap //key depend on value, key and value is transactionID
	ServerConnection	  []NodeClient
}

func (c *Coordinator)Init() {
	c.globalResources = new(ResourceMap)
	c.globalResources.Init()
	c.transactionDependency = new(DependencyMap)
}

func (*Coordinator) OpenTransaction(ctx context.Context, req *Empty) (*Transaction, error) {
	transactionID := utils.Concatenate(rand.Intn(1000000), int(time.Now().Unix()))
	return &Transaction{Id:&transactionID}, nil

}

func (*Coordinator) CloseTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseTransaction not implemented")
}

func (c *Coordinator) AskCommitTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	for _,elem := range c.ServerConnection {
		_,err := elem.CommitTransaction(context.Background(),req)
		utils.CheckError(err)
	}
	res := Feedback{}
	temp := "COMMITTED!"
	res.Message = &temp
	return &res,nil
}

func (c *Coordinator) AskAbortTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	for _,elem := range c.ServerConnection {
		_,err := elem.AbortTransaction(context.Background(),req)
		utils.CheckError(err)
	}
	res := Feedback{}
	temp := "ABORTED!"
	res.Message = &temp
	return &res,nil
}

func (c *Coordinator) TryLock(ctx context.Context, req *TryLockParam) (*Feedback, error) {
	fmt.Println("received new trylock request with transactionID: ", *req.TransactionID, ", lockType:", *req.LockType, ", server:", *req.ServerIdentifier, ", object:", *req.Object)
	resourceKey := c.globalResources.ConstructKey(*req)
	if c.globalResources.TryLockAt(*req, c) {
		message := "Success"
		fmt.Println("Got the mutex with param: ", *req.TransactionID)
		c.globalResources.Get(resourceKey).PrintContent()
		//time.Sleep(10 * time.Second)
		return &Feedback{Message:&message}, nil
	} else {
		message := "Abort"
		fmt.Println("Abort the mutex with param: ", *req)
		c.globalResources.Get(resourceKey).PrintContent()
		return &Feedback{Message:&message}, status.Errorf(codes.Aborted, "transaction aborted, found deadlock!")
	}
}

func (c*Coordinator) ReportUnlock(ctx context.Context, req *ReportUnLockParam) (*Empty, error) {
	resourceKey := utils.Concatenate(*req.ServerIdentifier, "_", *req.Object)
	c.globalResources.Get(resourceKey).UnlockHolder(TransactionUnit{transactionID: *req.TransactionID, lockType:*req.LockType})
	fmt.Println("Unlock with param: ", *req.TransactionID)
	c.globalResources.Get(resourceKey).PrintContent()
	return &Empty{}, nil
}

func (c *Coordinator) CheckDeadlock(param TryLockParam) bool {
	fmt.Println("Dead mutex detection: ")
	resourceKey := c.globalResources.ConstructKey(param)
	if c.globalResources.Has(resourceKey) {
		for _, owner := range c.globalResources.Get(resourceKey).lockHolders.items {
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
