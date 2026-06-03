package main

import (
	"fmt"
	"sync"
	"time"
)

func job(id int) {
	fmt.Println("start:", id)
	time.Sleep(time.Second)
	fmt.Println("done:", id)
}

func main() {
	const maxConcurrent = 3

	sem := make(chan struct{}, maxConcurrent)
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		<-ticker.C // rate limit start speed

		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			job(id)
		}(i)
	}

	wg.Wait()
}
