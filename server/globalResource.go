// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package server

import (
	"mp3/utils"
	"sync"
)

// cat generic_ccmap.go | genny gen "Key=string Value=*blockchain.Transaction" > [targetName].go
type CoordinatorDelegate interface{
	AddDependency(fromA string, toB string) bool
	DeleteDependency(fromA string, toB string)
	CheckDeadlock(initTransactionID string) bool
}

// StringDictionary the set of Items
type ResourceMap struct {
	items map[string]*ResourceObject //key: serverIdentifier + "_" + objectName, value: [transactionID + "_" + lockType]
	lock  sync.RWMutex
}

type transactionUnit struct {
	transactionID 	string
	lockType		string
}

func (d *ResourceMap) Init() {
	d.items = make(map[string]*ResourceObject)
}

// Set adds a new item to the ccmap
func (d *ResourceMap) Set(param TryLockParam) {
	d.lock.Lock()
	defer d.lock.Unlock()
	resourceKey := d.ConstructKey(param)
	if d.items[resourceKey] == nil {
		d.items[resourceKey] = new(ResourceObject)

	}
	d.items[resourceKey].owners = append(d.items[resourceKey].owners, transactionUnit{transactionID: *param.TransactionID, lockType:*param.LockType})
}

// Delete removes a value from the ccmap, given its key
func (d *ResourceMap) Delete(param ReportUnLockParam) bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	resourceKey := utils.Concatenate(*param.ServerIdentifier, "_", *param.Object)

	for i, owner := range d.items[resourceKey].owners {
		if owner.transactionID == *param.TransactionID {
			d.items[resourceKey].owners = append(d.items[resourceKey].owners[:i], d.items[resourceKey].owners[i+1:]...)
			return true
		}
	}
	return false
}

// Has returns true if the key exists in the ccmap
func (d *ResourceMap) Has(k string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	_, ok := d.items[k]
	return ok
}

// Get returns the value associated with the key
func (d *ResourceMap) Get(k string) *ResourceObject {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.items[k]
}

// Clear removes all the items from the ccmap
func (d *ResourceMap) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.items = make(map[string]*ResourceObject)
}

// Size returns the amount of elements in the ccmap
func (d *ResourceMap) Size() int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return len(d.items)
}

func (d *ResourceMap) ConstructKey(param TryLockParam) string {
	return utils.Concatenate(*param.ServerIdentifier, "_", *param.Object)
}

func (d *ResourceMap) TryLockAt(param TryLockParam, coordinator *Coordinator) bool {


	resourceKey := d.ConstructKey(param)
	hangingLockType := *param.LockType
	if coordinator.CheckDeadlock(param) {
		return false
	}

	if d.Has(resourceKey) {
		for _, owner := range d.Get(resourceKey).owners {
			if !(owner.lockType == "R" && *param.LockType == "R") {
				coordinator.transactionDependency.Set(*param.TransactionID, owner.transactionID)
			}
		}
	}

	if d.items[resourceKey] == nil {
		d.items[resourceKey] = new(ResourceObject)

	}
	if hangingLockType == "R" {
		d.Get(resourceKey).mutex.RLock()
	} else {
		d.Get(resourceKey).mutex.Lock()
	}
	m.lock
	for param.TransactionID != GettargetID {
		sync.Cond.Wait()
	}
	m.unlock

	if param.TransactionID in abortList {
		return false

	}
	d.Set(param)
	coordinator.transactionDependency.Delete(*param.TransactionID)
	return true
}