package main

import (
"fmt"
"time"
)

func main() {
	c := make(chan int) // an unbuffered channel
	go func(ch chan<- int, x int) {
		time.Sleep(time.Second)
		// <-ch    // fails to compile.
		// Block until the result is received.
		ch <- x*x // 9 is sent
	}(c, 3)

	go func(ch <-chan int) {
		// Block here until 9 is sent.
		n := <-ch
		fmt.Println("1L: ", n) // 9
		// ch <- 123   // fails to compile
		time.Sleep(time.Second)
	}(c)

	go func(ch <-chan int) {
		// Block here until 9 is sent.
		n := <-ch
		fmt.Println("2L: ",n) // 9
		// ch <- 123   // fails to compile
		time.Sleep(time.Second)
	}(c)

	fmt.Println("bye")
	for {
		time.Sleep(10*time.Second)
	}
}
