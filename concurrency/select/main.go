package main

import (
	"context"
	"fmt"
	"sync"
)

func main() {
	ch := make(chan int, 10)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			ch <- i
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	values := make([]int, 10)

loop:
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				break loop
			}

			values[v] = v

		case <-ctx.Done():
			break loop
		}
	}

	for _, val := range values {
		fmt.Println(val)
	}
}
