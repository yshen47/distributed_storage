package main

import (
	"fmt"
	"sync"
)

func main() {
	mutex := sync.RWMutex{}
	fmt.Println("10")
	mutex.RLock()
	fmt.Println("20")
	mutex.Lock()
	fmt.Print("30")
}
