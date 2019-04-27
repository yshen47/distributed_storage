package concurrency_data_structure

// cat generic_ccmap.go | genny gen "Key=string Value=*blockchain.Transaction" > [targetName].go
import (
"sync"
"github.com/cheekybits/genny/generic"
)

// Key the key of the ccmap
type Key generic.Type

// TransactionMap the content of the ccmap
type Value generic.Type

// ValueDictionary the set of Items
type ValueMap struct {
	items map[Key]Value
	lock  sync.RWMutex
}

// Set adds a new item to the ccmap
func (d *ValueMap) Set(k Key, v Value) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.items == nil {
		d.items = make(map[Key]Value)
	}
	d.items[k] = v
}

// Delete removes a value from the ccmap, given its key
func (d *ValueMap) Delete(k Key) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	_, ok := d.items[k]
	if ok {
		delete(d.items, k)
	}
	return ok
}

// Has returns true if the key exists in the ccmap
func (d *ValueMap) Has(k Key) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	_, ok := d.items[k]
	return ok
}

// Get returns the value associated with the key
func (d *ValueMap) Get(k Key) Value {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.items[k]
}

// Clear removes all the items from the ccmap
func (d *ValueMap) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.items = make(map[Key]Value)
}

// Size returns the amount of elements in the ccmap
func (d *ValueMap) Size() int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return len(d.items)
}

// Keys returns a slice of all the keys present
func (d *ValueMap) GetKys() []Key {
	d.lock.RLock()
	defer d.lock.RUnlock()
	keys := []Key{}
	for i := range d.items {
		keys = append(keys, i)
	}
	return keys
}

// Values returns a slice of all the values present
func (d *ValueMap) GetVals() []Value {
	d.lock.RLock()
	defer d.lock.RUnlock()
	values := []Value{}
	for i := range d.items {
		values = append(values, d.items[i])
	}
	return values
}
