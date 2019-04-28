package concurrency_data_structure
//cat generic_ccmap.go | genny gen "Key=string Value=*blockchain.Transaction" > [targetName].go
import (
	"github.com/cheekybits/genny/generic"
	"sync"
)

type Value generic.Type

// ValueList the set of Items
type ValueList struct {
	items [] Value
	lock  sync.RWMutex
}

// Set adds a new item to the tail of the list
func (d *ValueList) Append(v Value) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.items == nil {
		d.items =  make([]Value, 1)
	}
	d.items = append(d.items, v)
}

// GetTransactionToCommit front
func (d *ValueList) Pop(n int) []Value {
	d.lock.Lock()
	defer d.lock.Unlock()
	var res [] Value
	if n < len(d.items) {
		res = d.items[:n]
		d.items = d.items[n:]
	} else {
		res = d.items
		d.items = make([]Value, 1)
	}
	return res
}

func (d *ValueList) Size() int {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return len(d.items)
}