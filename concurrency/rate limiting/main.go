package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int) {
	fmt.Println("started worker:", id)
	time.Sleep(time.Second)
	fmt.Println("finished worker:", id)
}

func main() {
	const maxConcurrent = 3

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			// acquire semaphore slot
			sem <- struct{}{}

			// release semaphore slot
			defer func() {
				<-sem
			}()

			worker(id)
		}(i)
	}

	wg.Wait()
}
