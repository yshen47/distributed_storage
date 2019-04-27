// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package server

import (
	"mp3/utils"
	"strings"
	"fmt"
	"sync"
	"time"
)

// cat generic_ccmap.go | genny gen "String=string String=*blockchain.Transaction" > [targetName].go

// StringDictionary the set of Items
type ResourceMap struct {
	items map[string]string //key: serverIdentifier + "_" + objectName, value: transactionID + "_" + lockType
	lock  sync.RWMutex
}

// Set adds a new item to the ccmap
func (d *ResourceMap) Set(param TryLockParam) {
	d.lock.Lock()
	defer d.lock.Unlock()
	resourceKey := d.ConstructKey(param)
	resourceVal := utils.Concatenate(param.TransactionID, "_", param.LockType)
	if d.items == nil {
		d.items = make(map[string]string)
	}
	d.items[resourceKey] = resourceVal
}

// Delete removes a value from the ccmap, given its key
func (d *ResourceMap) Delete(k string) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	_, ok := d.items[k]
	if ok {
		delete(d.items, k)
	}
	return ok
}

// Has returns true if the key exists in the ccmap
func (d *ResourceMap) Has(k string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	_, ok := d.items[k]
	return ok
}

// Get returns the value associated with the key
func (d *ResourceMap) Get(k string) []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return strings.Split(d.items[k], "_")
}

// Clear removes all the items from the ccmap
func (d *ResourceMap) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.items = make(map[string]string)
}

// Size returns the amount of elements in the ccmap
func (d *ResourceMap) Size() int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return len(d.items)
}

// Strings returns a slice of all the keys present
func (d *ResourceMap) GetKys() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	keys := []string{}
	for i := range d.items {
		keys = append(keys, i)
	}
	return keys
}

//// Strings returns a slice of all the values present
//func (d *ResourceMap) GetVals() []string {
//	d.lock.RLock()
//	defer d.lock.RUnlock()
//	values := []string{}
//	for i := range d.items {
//		values = append(values, d.items[i])
//	}
//	return values
//}

func (d *ResourceMap) ConstructKey(param TryLockParam) string {
	return utils.Concatenate(param.ServerIdentifier, "_", param.Object)
}

func (d *ResourceMap) TryLockAt(param TryLockParam, abortChannel chan string) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	resourceKey := d.ConstructKey(param)
	hangingLockType := *param.LockType
	for {
		continueFlag := true
		select {
		case transactionID := <- abortChannel:
			if transactionID == *param.TransactionID {
				fmt.Println("Abort ", transactionID)
				return
			}
		default:
			resource, ok := d.items[resourceKey]
			if ok {
				lockType := resource[-1]
				if lockType == 'R' && hangingLockType == "R" {
					continueFlag = false
				}
			} else {
				continueFlag = false
			}
		}
		if !continueFlag {
			break
		}
		time.Sleep(100*time.Millisecond)
	}
	d.Set(param)
}
