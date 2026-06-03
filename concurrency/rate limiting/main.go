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
	maxConcurrent := 4
	sem := make(chan struct{}, maxConcurrent)
	ticker := time.NewTicker(200 * time.Millisecond)
	var wg sync.WaitGroup

	defer ticker.Stop()
	for i := 0; i < 36; i++ {
		<-ticker.C
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			job(id)
		}(i)
	}
}
