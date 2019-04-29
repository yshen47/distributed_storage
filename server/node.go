package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	uncommittedHistory 	*TransactionHistory
}

type TransactionEntry struct {
	ObjName       string // record which object was read/write in current operation, used for 2PL lock
	LockType	  string // store lock type
	transactionID string
	CurrState     map[string]string
}

func (h *TransactionEntry)initHistory(id string, objName string, locktype string){
	h.transactionID = id
	h.CurrState = make(map[string]string)
	h.ObjName = objName
	h.LockType = locktype
}


type LockTuple struct {
	mutex         	sync.RWMutex
	owners 	map[string]Owner
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
	n.uncommittedHistory = new(TransactionHistory)
}

func (n *Node) ClientSet(ctx context.Context, req *SetParams) (*Feedback, error) {
	if *req.ServerIdentifier != n.Name {
		return nil, status.Error(codes.InvalidArgument, "Called the wrong node server")
	}

	//TODO: add WLOCK
	n.lockMapLock.Lock()
	_, ok := n.lockMap[*req.ObjectName]
	if !ok {
		n.lockMap[*req.ObjectName] = new(LockTuple)
		n.lockMap[*req.ObjectName].owners = make(map[string]Owner)
	}
	_ , ok = n.lockMap[*req.ObjectName].owners[*req.TransactionID]
	n.lockMap[*req.ObjectName].owners[*req.TransactionID] = Owner{transactionID:*req.TransactionID,lockType:"W"}
	n.lockMapLock.Unlock()
	isSuccessful := n.WLock(*req.ObjectName, *req.TransactionID)
	//defer n.WUnLock(*req.ObjectName, *req.TransactionID)

	if !isSuccessful {
		resFeedback := &Feedback{}
		result := "ABORTED"
		resFeedback.Message = &result
		return resFeedback, status.Errorf(codes.Aborted, "Transaction aborted due to deadlock!")
	}

	if n.uncommittedHistory.Size() == 0 {
		h := TransactionEntry{}
		h.initHistory(*req.TransactionID,*req.ObjectName,"W")
		h.CurrState = make(map[string]string)
		h.CurrState[*req.ObjectName] = *req.Value
		n.uncommittedHistory.Append(h)
	}else{
		prevMap := n.uncommittedHistory.Get(n.uncommittedHistory.Size() - 1).CurrState
		newMap := make(map[string]string)
		for k,v := range prevMap {
			newMap[k] = v
		}
		newMap[*req.ObjectName] = *req.Value
		newEntry := TransactionEntry{}
		newEntry.initHistory(*req.TransactionID,*req.ObjectName,"W")
		newEntry.CurrState = newMap
		n.uncommittedHistory.Append(newEntry)
	}

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
		n.lockMap[*req.ObjectName] = new(LockTuple)
		n.lockMap[*req.ObjectName].owners = make(map[string]Owner)
	}
	_ , ok = n.lockMap[*req.ObjectName].owners[*req.TransactionID]
	if !ok {
		n.lockMap[*req.ObjectName].owners[*req.TransactionID] = Owner{transactionID:*req.TransactionID,lockType:"R"}
	}

	n.lockMapLock.RUnlock()


	n.RLock(*req.ObjectName, *req.TransactionID)
	//defer n.RUnLock(*req.ObjectName, *req.TransactionID)

	prevMap := n.uncommittedHistory.Get(n.uncommittedHistory.Size() - 1).CurrState
	newMap := make(map[string]string)
	for k,v := range prevMap {
		newMap[k] = v
	}
	newEntry := TransactionEntry{}
	newEntry.initHistory(*req.TransactionID,*req.ObjectName,"R")
	newEntry.CurrState = newMap
	n.uncommittedHistory.Append(newEntry)

	resFeedback := &Feedback{}
	val, ok := newMap[*req.ObjectName]
	if !ok {
		val,ok = n.data[*req.ObjectName]
		if !ok {
			return resFeedback, status.Error(codes.Aborted, "not found")
		}
	}
	resFeedback.Message = &val
	return resFeedback, nil

}


func (n *Node) CommitTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	for i := n.uncommittedHistory.Size() - 1; i >= 0; i-- {
		if n.uncommittedHistory.Get(i).transactionID == *req.Id {
			for k,v := range n.uncommittedHistory.Get(i).CurrState {
				n.data[k] = v
			}
			currEntry := n.uncommittedHistory.Delete(i)
			if currEntry.LockType == "W" {
				n.WUnLock(currEntry.ObjName,currEntry.transactionID)
			}else{
				n.RUnLock(currEntry.ObjName,currEntry.transactionID)
			}

		}
	}
	res := Feedback{}
	words := "COMMIT OK"
	res.Message = &words
	return &res,nil
}
func (n *Node) AbortTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	n.abortTransaction(*req.Id)
	res := Feedback{}
	words := "ABORTED"
	res.Message = &words
	return &res,nil
}

func (n *Node) abortTransaction(transactionID string) {
	for i := n.uncommittedHistory.Size() - 1; i >= 0; i-- {
		if n.uncommittedHistory.Get(i).transactionID == transactionID {
			currEntry := n.uncommittedHistory.Delete(i)
			if currEntry.LockType == "W" {
				n.WUnLock(currEntry.ObjName,currEntry.transactionID)
			}else{
				n.RUnLock(currEntry.ObjName,currEntry.transactionID)
			}
		}
	}
}

func (n *Node) RLock(objectName string, transactionID string) bool{
	n.lockMapLock.Lock()
	if n.lockMap[objectName].owners[transactionID].lockType == "W" {
		n.lockMapLock.Unlock()
		return true
	}
	n.lockMapLock.Unlock()

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

	n.lockMapLock.Lock()
	n.lockMap[objectName].mutex.RLock()
	n.lockMapLock.Unlock()
	return true
}

func (n *Node) RUnLock(objectName string, transactionID string) {
	n.lockMapLock.Lock()
	defer n.lockMapLock.Unlock()
	if n.lockMap[objectName].owners[transactionID].lockType == "W" {
		return
	}
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
	delete(n.lockMap[objectName].owners, transactionID)
}

func (n *Node) WLock(objectName string, transactionID string) bool{

	//fmt.Println("WLock on ", objectName)
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
	n.lockMapLock.Lock()
	n.lockMap[objectName].mutex.Lock()
	n.lockMapLock.Unlock()
	return true
}

func (n *Node) WUnLock(objectName string, transactionID string) {
	n.lockMapLock.Lock()
	defer n.lockMapLock.Unlock()
	//fmt.Println("WUnLock on ", objectName)
	reportUnlockParam := ReportUnLockParam{}
	reportUnlockParam.Object = &objectName
	reportUnlockParam.ServerIdentifier = &n.Name
	reportUnlockParam.TransactionID = &transactionID
	lockType := "W"
	reportUnlockParam.LockType = &lockType
	n.CoordinatorDelegate.ReportUnlock(context.Background(), &reportUnlockParam)
	n.lockMap[objectName].mutex.Unlock()
	delete(n.lockMap[objectName].owners, transactionID)
}
