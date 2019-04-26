package server

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
	"sync"
)

// Node can be embedded to have forward compatible implementations.
type Node struct {
	Name                string
	CoordinatorDelegate CoordinatorClient
	data                map[string]string
	lockMap  			map[string]*LockTuple
	lockMapLock			sync.RWMutex
}

type LockTuple struct {
	mutex         sync.RWMutex
	transactionID string
}
// NodeServer is the server API for Node service.
type NodeServer interface {
	ClientSet(context.Context, *SetParams) (*Feedback, error)
	ClientGet(context.Context, *GetParams) (*Feedback, error)
}

func (n *Node) ClientSet(ctx context.Context, req *SetParams) (*Feedback, error) {
	if *req.ServerIdentifier != n.Name {
		return nil, status.Error(codes.InvalidArgument, "Called the wrong node server")
	}
	//TODO: add WLOCK
	n.lockMapLock.Lock()
	_, ok := n.lockMap[*req.ObjectName]
	if !ok {
		n.lockMap[*req.ObjectName] = &LockTuple{
			mutex:         sync.RWMutex{},
			transactionID: *req.TransactionID,
		}
	} else {
		n.lockMap[*req.ObjectName].transactionID = *req.TransactionID
	}
	n.lockMapLock.Unlock()

	n.WLock(*req.ObjectName)
	defer n.WUnLock(*req.ObjectName)

	n.data[*req.ObjectName] = *req.Value
	resFeedback := &Feedback{}
	result := "OK"
	resFeedback.Message = &result
	return resFeedback, nil
}

func (n *Node) ClientGet(ctx context.Context, req *GetParams) (*Feedback, error) {
	if *req.ServerIdentifier != n.Name {
		return nil, status.Error(codes.InvalidArgument, "Called the wrong node server")
	}
	n.lockMapLock.RLock()
	_, ok := n.lockMap[*req.ObjectName]
	if !ok {
		resFeedback := &Feedback{}
		var result string
		result = "NOT FOUND"
		resFeedback.Message = &result
		//TODO: tell coordinator to abort the current transaction

		n.lockMapLock.RUnlock()
		return resFeedback, status.Error(codes.Aborted, "not found")
	}
	n.lockMapLock.RUnlock()


	n.RLock(*req.ObjectName)
	defer n.RUnLock(*req.ObjectName)

	resFeedback := &Feedback{}
	val, ok := n.data[*req.ObjectName]
	resFeedback.Message = &val
	return resFeedback, nil
}

func (n *Node) RLock(objectName string) {
	fmt.Println("RLock on ", objectName)
	n.lockMap[objectName].mutex.RLock()
}

func (n *Node) RUnLock(objectName string) {
	fmt.Println("RUnLock on ", objectName)
	n.lockMap[objectName].mutex.RUnlock()
}

func (n *Node) WLock(objectName string) {
	fmt.Println("WLock on ", objectName)
	n.lockMap[objectName].mutex.Lock()
}

func (n *Node) WUnLock(objectName string) {
	fmt.Println("WUnLock on ", objectName)
	n.lockMap[objectName].mutex.Unlock()
}
