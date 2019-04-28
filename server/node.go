package server

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
	"mp3/utils"
	"sync"
)

// Node can be embedded to have forward compatible implementations.
type Node struct {
	Name                string
	CoordinatorDelegate CoordinatorClient
	data                map[string]string // this is committed state
	lockMap  			map[string]*LockTuple
	lockMapLock			sync.RWMutex
	uncommittedHistory 	[]TransactionHistory
}

type TransactionHistory struct {
	objName				string // record which object was read/write in current operation, used for 2PL lock
	transactionID		string
	CurrState			map[string]string
}

func (h *TransactionHistory)initHistory(id string){
	h.transactionID = id
	h.CurrState = make(map[string]string)
}


type LockTuple struct {
	mutex         sync.RWMutex
	transactionID string
}
// NodeServer is the server API for Node service.
type NodeServer interface {
	ClientSet(context.Context, *SetParams) (*Feedback, error)
	ClientGet(context.Context, *GetParams) (*Feedback, error)
	CommitTransaction(context.Context, *Transaction) (*Feedback, error)
	AbortTransaction(context.Context, *Transaction) (*Feedback, error)
}

func (n *Node) Init() {
	n.lockMap = make(map[string]*LockTuple)
	n.data = make(map[string]string)
	n.uncommittedHistory = make([]TransactionHistory,0)
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
	isSuccessful := n.WLock(*req.ObjectName, *req.TransactionID)
	//defer n.WUnLock(*req.ObjectName, *req.TransactionID)

	if !isSuccessful {
		resFeedback := &Feedback{}
		result := "ABORTED"
		resFeedback.Message = &result
		return resFeedback, status.Errorf(codes.Aborted, "Transaction aborted due to deadlock!")
	}

	if len(n.uncommittedHistory) == 0 {
		currentState := TransactionHistory{}
		currentState.initHistory(*req.TransactionID)
		currentState.CurrState[*req.ObjectName] = *req.Value
		currentState.objName = *req.ObjectName
		n.uncommittedHistory = append(n.uncommittedHistory,currentState)
	}else {
		var prevMap map[string]string
		for i :=len(n.uncommittedHistory) -1; i >=0; i-- { // find the most updated table with my transaction ID
			if n.uncommittedHistory[i].transactionID == *req.TransactionID {
				prevMap = n.uncommittedHistory[i].CurrState
			}
		}
		newMap := make(map[string]string)
		for k,v := range prevMap {
			newMap[k] = v
		}
		newMap[*req.ObjectName] = *req.Value
		newEntry := TransactionHistory{}
		newEntry.CurrState = newMap
		newEntry.transactionID = *req.TransactionID
		newEntry.objName = *req.ObjectName
		n.uncommittedHistory = append(n.uncommittedHistory, newEntry)
	}

	resFeedback := &Feedback{}
	result := "OK"
	resFeedback.Message = &result
	fmt.Println("About to return 67")
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


	n.RLock(*req.ObjectName, *req.TransactionID)
	//defer n.RUnLock(*req.ObjectName, *req.TransactionID)

	var prevMap map[string]string
	for i :=len(n.uncommittedHistory) -1; i >=0; i-- { // find the most updated table with my transaction ID
		if n.uncommittedHistory[i].transactionID == *req.TransactionID {
			prevMap = n.uncommittedHistory[i].CurrState
		}
	}
	newMap := make(map[string]string)
	for k,v := range prevMap {
		newMap[k] = v
	}
	newEntry := TransactionHistory{}
	newEntry.CurrState = newMap
	newEntry.transactionID = *req.TransactionID
	newEntry.objName = *req.ObjectName
	n.uncommittedHistory = append(n.uncommittedHistory, newEntry)


	resFeedback := &Feedback{}
	val, ok := n.data[*req.ObjectName]
	resFeedback.Message = &val
	return resFeedback, nil
}


func (n *Node) CommitTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	fmt.Println("here")
	for i := len(n.uncommittedHistory) - 1; i >= 0; i-- {
		if n.uncommittedHistory[i].transactionID == *req.Id {
			for k,v := range n.uncommittedHistory[i].CurrState {
				n.data[k] = v
			}
			n.uncommittedHistory = nil
			res := Feedback{}
			words := "COMMIT OK"
			res.Message = &words
			return &res,nil
		}
	}
	return nil, status.Errorf(codes.Unknown, "commit fails")
}
func (n *Node) AbortTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	n.abortTransaction(*req.Id)
	return nil, status.Errorf(codes.Unimplemented, "method AbortTransaction not implemented")
}

func (n *Node) abortTransaction(transactionID string) {
	n.uncommittedHistory = nil
}

func (n *Node) RLock(objectName string, transactionID string) bool{
	fmt.Println("RLock on ", objectName)
	tryLockParam := TryLockParam{}
	tryLockParam.TransactionID = &transactionID
	tryLockParam.ServerIdentifier = &n.Name
	tryLockParam.Object = &objectName
	lockType := "R"
	tryLockParam.LockType = &lockType
	_, err := n.CoordinatorDelegate.TryLock(context.Background(), &tryLockParam)
	s, _ := status.FromError(err)
	if s.Code().String() == "Aborted" {
		fmt.Println(s.Message())
		n.abortTransaction(transactionID)
		return false
	}
	n.lockMap[objectName].mutex.RLock()
	return true
}

func (n *Node) RUnLock(objectName string, transactionID string) {
	fmt.Println("RUnLock on ", objectName)
	reportUnlockParam := ReportUnLockParam{}
	reportUnlockParam.Object = &objectName
	reportUnlockParam.ServerIdentifier = &n.Name
	reportUnlockParam.TransactionID = &transactionID
	lockType := "R"
	reportUnlockParam.LockType = &lockType
	_, err := n.CoordinatorDelegate.ReportUnlock(context.Background(), &reportUnlockParam)
	utils.CheckError(err)
	n.lockMap[objectName].mutex.RUnlock()
}

func (n *Node) WLock(objectName string, transactionID string) bool{
	fmt.Println("WLock on ", objectName)
	tryLockParam := TryLockParam{}
	tryLockParam.TransactionID = &transactionID
	tryLockParam.ServerIdentifier = &n.Name
	lockType := "W"
	tryLockParam.LockType = &lockType
	tryLockParam.Object = &objectName
	_, err := n.CoordinatorDelegate.TryLock(context.Background(), &tryLockParam)
	s, _ := status.FromError(err)
	if s.Code().String() == "Aborted" {
		fmt.Println(s.Message())
		n.abortTransaction(transactionID)
		return false
	}

	n.lockMap[objectName].mutex.Lock()
	return true
}

func (n *Node) WUnLock(objectName string, transactionID string) {
	fmt.Println("WUnLock on ", objectName)
	reportUnlockParam := ReportUnLockParam{}
	reportUnlockParam.Object = &objectName
	reportUnlockParam.ServerIdentifier = &n.Name
	reportUnlockParam.TransactionID = &transactionID
	lockType := "W"
	reportUnlockParam.LockType = &lockType
	n.CoordinatorDelegate.ReportUnlock(context.Background(), &reportUnlockParam)
	n.lockMap[objectName].mutex.Unlock()
}
