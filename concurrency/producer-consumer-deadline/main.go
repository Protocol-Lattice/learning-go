package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

func producer(ctx context.Context, ch chan<- int, wg *sync.WaitGroup, index int) {
	defer wg.Done()
	for i := 0; i < 10; i++ {
		// Symulujemy pracę producenta.
		select {
		case <-ctx.Done():
			fmt.Printf("Producer %d stopped: %v\n", index, ctx.Err())
			return
		case <-time.After(20 * time.Millisecond):
		}

		value := index*10 + i

		select {
		case <-ctx.Done():
			fmt.Printf("Producer %d stopped: %v\n", index, ctx.Err())
			return
		case ch <- value:
			fmt.Println("Sent:", value)
		}
	}
}

func consumer(ch <-chan int, wg *sync.WaitGroup, mu *sync.Mutex, values *[]int) {
	defer wg.Done()

	for v := range ch {
		mu.Lock()
		*values = append(*values, v)
		mu.Unlock()
	}
}
func main() {
	ch := make(chan int, 10)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wp := sync.WaitGroup{}
	values := []int{}
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go producer(ctx, ch, &wg, i)
	}
	for i := 0; i < 10; i++ {
		wp.Add(1)

		go consumer(ch, &wp, &mu, &values)
	}
	wg.Wait()
	close(ch)
	wp.Wait()
	sort.Ints(values)

	for _, v := range values {
		fmt.Println("Received: ", v)
	}

	fmt.Println("Done")
}
