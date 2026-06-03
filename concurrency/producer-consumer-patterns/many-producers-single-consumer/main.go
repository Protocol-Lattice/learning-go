package main

import (
	"fmt"
	"sync"
)

func Producer(wg *sync.WaitGroup, ch chan<- int, i int) {
	for value := 0; value < 10; value++ {
		ch <- value
		fmt.Println("Produced: ", value, " by worker: ", i)
	}
	defer wg.Done()
}

func Consumer(wg *sync.WaitGroup, ch <-chan int) {
	defer wg.Done()

	for c := range ch {
		fmt.Println("Consumed: ", c)
	}
}

func main() {
	wp := sync.WaitGroup{}
	wc := sync.WaitGroup{}
	ch := make(chan int)
	wc.Add(1)
	for i := 0; i < 10; i++ {
		wp.Add(1)
		go Producer(&wp, ch, i)

	}
	go Consumer(&wc, ch)

	wp.Wait()
	close(ch)
	wc.Wait()
}
