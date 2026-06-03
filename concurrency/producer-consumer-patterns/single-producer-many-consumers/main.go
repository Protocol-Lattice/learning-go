package main

import (
	"fmt"
	"sync"
)

func Producer(ch chan<- int) {
	for i := 0; i < 10; i++ {
		ch <- i
		fmt.Println("Produced: ", i)
	}
	defer close(ch)
}

func Consumer(wg *sync.WaitGroup, ch <-chan int, i int) {
	defer wg.Done()

	for c := range ch {
		fmt.Println("Consumed: ", c, " by worker: ", i)
	}
}

func main() {
	wg := sync.WaitGroup{}
	ch := make(chan int, 10)
	go Producer(ch)
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go Consumer(&wg, ch, i)
	}

	wg.Wait()
}
