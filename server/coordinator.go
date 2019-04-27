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
}

func (c *Coordinator)Init() {
	c.globalResources = new(ResourceMap)
	c.transactionDependency = new(DependencyMap)
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
	resourceKey := c.globalResources.ConstructKey(*req)
	if c.globalResources.Has(resourceKey) {
		origValues := c.globalResources.Get(resourceKey)
		c.transactionDependency.Set(*req.TransactionID, origValues[0])
		fmt.Println("line 50")
	}
	fmt.Println("line 51")
	if c.globalResources.TryLockAt(*req, c.abortChannel, c) {
		message := "Success"
		return &Feedback{Message:&message}, nil
	} else {
		message := "Abort"
		return &Feedback{Message:&message}, status.Errorf(codes.Aborted, "transaction aborted, found deadlock!")
	}

}

func (c*Coordinator) ReportUnlock(ctx context.Context, req *ReportUnLockParam) (*Empty, error) {
	c.globalResources.Delete(*req)
	return &Empty{}, nil
}

func (c *Coordinator) AddDependency(fromA string, toB string) bool{
	if !(c.transactionDependency.Has(fromA) && c.transactionDependency.Get(fromA) == toB) {
		c.transactionDependency.Set(fromA, toB)
		return true
	}
	return false
}


func (c *Coordinator) DeleteDependency(fromA string, toB string) {
	if c.transactionDependency.Get(fromA) == toB {
		c.transactionDependency.Delete(fromA)
	}
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
