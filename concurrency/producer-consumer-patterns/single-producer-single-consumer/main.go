package main

import (
	"fmt"
	"sync"
)

func Producer(ch chan<- int) {
	ch <- 10
	fmt.Println("Produced: ", 10)
	defer close(ch)
}

func Consumer(wg *sync.WaitGroup, ch <-chan int) {
	defer wg.Done()

	for c := range ch {
		fmt.Println("Consumed: ", c)
	}
}

func main() {
	wg := sync.WaitGroup{}
	ch := make(chan int)
	wg.Add(1)
	go Producer(ch)
	go Consumer(&wg, ch)

	wg.Wait()
}
