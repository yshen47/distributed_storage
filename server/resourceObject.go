package server

import (
	"fmt"
	"os"
	"sync"
)

type ResourceObject struct {
	mutex        sync.Mutex
	cond         sync.Cond
	abortList    *TransactionUnitList
	upgradeList  *TransactionUnitList
	waitingQueue *TransactionUnitList
	lockHolders  *TransactionUnitList
}

func (ro *ResourceObject)Init() {
	ro.abortList = new(TransactionUnitList)
	ro.upgradeList = new(TransactionUnitList)
	ro.waitingQueue = new(TransactionUnitList) //order: writer first then reader
	ro.lockHolders = new(TransactionUnitList)
	ro.cond = *sync.NewCond(&ro.mutex)
}

func (ro *ResourceObject) GetNextTarget(modified bool) string {
	//ro.mutex.Lock()
	//defer ro.mutex.Unlock()
	if ro.getHolderType() == "W" || ro.getHolderType() == "" {
		//holder type: W or nil   1.ID to be aborted 2. upgrade list writer 3.writer 4. reader
		if ro.abortList.Size() > 0 {
			abortID := ro.abortList.Pop("W", modified)
			return abortID.transactionID
		}
		if ro.upgradeList.Size() > 0 {
			upgradeID := ro.upgradeList.Pop("W", modified)
			return upgradeID.transactionID
		}
		waitingID := ro.waitingQueue.Pop("", modified)
		return waitingID.transactionID

	} else if ro.getHolderType() == "R" {
		//holder type: R    1. ID to be aborted 2. if upgradelist != nil return nil 3. reader 4. writer
		if ro.abortList.Size() > 0 {
			abortID := ro.abortList.Pop("W", modified)
			return abortID.transactionID
		}
		if ro.upgradeList.Size() > 0 {
			return "RESERVEDKEY"
		}
		waitingID := ro.waitingQueue.Pop("", modified)
		return waitingID.transactionID
	}  else {
		fmt.Println("Lock holders have mixed types!")
		os.Exit(6)
	}
	return ""
}

func (ro *ResourceObject) getHolderType() string {
	if ro.lockHolders.firstReaderLoc == 0 {
		if ro.lockHolders.Size() == 0 {
			return ""
		} else {
			return "R"
		}
	} else if ro.lockHolders.firstReaderLoc == ro.lockHolders.Size() {
		return "W"
	} else {
		return "NA"
	}
}

func (ro *ResourceObject) AppendToWaitingQueue(unit transactionUnit) {
	ro.mutex.Lock()
	ro.waitingQueue.Append(unit)
	ro.mutex.Unlock()
}
func (ro *ResourceObject) UnlockHolder(unit transactionUnit) {
	ro.mutex.Lock()
	if !ro.lockHolders.Remove(unit) {
		fmt.Println("unit doesn't exist in holders!")
		ro.mutex.Unlock()
		return
	}
	ro.cond.Broadcast()
	ro.mutex.Unlock()
}

func (ro *ResourceObject) PrintContent() {
	ro.waitingQueue.PrintContent("waitingQueue:")
	ro.lockHolders.PrintContent("lockholders:")
	ro.abortList.PrintContent("abortList:")
	ro.upgradeList.PrintContent("upgradeList:")
}
