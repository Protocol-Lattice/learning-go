package main

import (
	"fmt"
	"sync"
)

func Producer(wg *sync.WaitGroup, ch chan<- int, i int) {
	defer wg.Done()
	for i := 0; i < 10; i++ {
		ch <- i
		fmt.Println("Produced: ", i, "by worker id: ", i)
	}
}

func Consumer(wg *sync.WaitGroup, ch <-chan int, i int) {
	defer wg.Done()

	for c := range ch {
		fmt.Println("Consumed: ", c, " by worker: ", i)
	}
}

func main() {
	wp := sync.WaitGroup{}
	wc := sync.WaitGroup{}
	ch := make(chan int, 10)
	for i := 0; i < 10; i++ {
		wp.Add(1)

		go Producer(&wp, ch, i)
	}
	for i := 0; i < 10; i++ {
		wc.Add(1)

		go Consumer(&wc, ch, i)
	}

	wp.Wait()
	close(ch)
	wc.Wait()
}
