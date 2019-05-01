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

func (ro *ResourceObject) GetNextTarget(modified bool) TransactionUnit {
	//ro.mutex.Lock()
	//defer ro.mutex.Unlock()
	holderType := ro.getHolderType()
	if holderType == "W" {
		//holder type: W  1.ID to be aborted
		if ro.abortList.Size() > 0 {
			abortID := ro.abortList.Pop("W", modified)
			return abortID
		}
		return TransactionUnit{transactionID: "RESERVEDKEY", lockType:"NA"}
	} else if holderType == "" {
		//holder type: nil   1.ID to be aborted 2. upgrade list writer 3.writer 4. reader
		if ro.abortList.Size() > 0 {
			abortID := ro.abortList.Pop("W", modified)
			return abortID
		}
		if ro.upgradeList.Size() > 0 {
			upgradeID := ro.upgradeList.Pop("W", modified)
			return upgradeID
		}
		waitingID := ro.waitingQueue.Pop("", modified)
		return waitingID
	} else if holderType == "R" {
		//holder type: R    1. ID to be aborted 2. if upgradelist != nil return nil 3. reader 4. writer
		if ro.abortList.Size() > 0 {
			abortID := ro.abortList.Pop("W", modified)
			return abortID
		}
		if ro.upgradeList.Size() > 0 {
			return TransactionUnit{transactionID: "RESERVEDKEY", lockType:"NA"}
		}
		waitingID := ro.waitingQueue.Pop("", modified)
		return waitingID
	}  else {
		fmt.Println("Lock holders have mixed types! ", holderType, ro.lockHolders.firstReaderLoc, ro.lockHolders.Size())
		ro.PrintContent()
		os.Exit(6)
	}
	return TransactionUnit{transactionID: "", lockType:"NA"}
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

func (ro *ResourceObject) AppendToWaitingQueue(unit TransactionUnit) {
	ro.mutex.Lock()
	ro.waitingQueue.Append(unit)
	ro.mutex.Unlock()
}

func (ro *ResourceObject) AppendToUpgradeList(unit TransactionUnit) {
	ro.mutex.Lock()
	ro.upgradeList.Append(unit)
	ro.mutex.Unlock()
}
func (ro *ResourceObject) UnlockHolder(unit TransactionUnit) {
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
