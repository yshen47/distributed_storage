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
	fmt.Println("received new trylock request with param: ", *req)
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
		time.Sleep(10 * time.Second)
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

func (c *Coordinator) AddDependency(fromA string, toB string) bool{
	if !(c.transactionDependency.Has(fromA) && c.transactionDependency.Get(fromA) == toB) {
		fmt.Println("New Dependency: ", fromA, " depends on ", toB)
		c.transactionDependency.Set(fromA, toB)
		return true
	}
	return false
}

func (c *Coordinator) DeleteDependency(fromA string) {
	fmt.Println("Delete dependency: " ,fromA, " on ", c.transactionDependency.Get(fromA))
	c.transactionDependency.Delete(fromA)
	fmt.Println(c.transactionDependency.items)
}

func (c *Coordinator) CheckDeadlock(initTransactionID string) bool {
	i := 0
	curr := initTransactionID
	for i < c.transactionDependency.Size() + 1 {
		if c.transactionDependency.Has(curr) {
			next := c.transactionDependency.Get(curr)
			if next == initTransactionID {
				return true
			}
			curr = next
		} else {
			return false
		}
		i += 1
	}
	return false
}
