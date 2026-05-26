package main

import (
	"fmt"
	"sync"
)

func main() {
	mutexExample()
	rwMutexExample()
}

func mutexExample() {
	mu := sync.Mutex{}
	counter := 0
	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mu.Lock()
			defer mu.Unlock()

			counter++
		}()
	}

	wg.Wait()

	fmt.Println("Mutex counter value:", counter)
}

func rwMutexExample() {
	rwMu := sync.RWMutex{}
	counter := 0
	wg := sync.WaitGroup{}

	// Writers: only one writer can update counter at a time.
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			rwMu.Lock()
			counter++
			rwMu.Unlock()
		}()
	}

	// Readers: many readers can read counter at the same time.
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			rwMu.RLock()
			fmt.Printf("Reader %d sees counter: %d\n", id, counter)
			rwMu.RUnlock()
		}(i)
	}

	wg.Wait()

	rwMu.RLock()
	fmt.Println("RWMutex final counter value:", counter)
	rwMu.RUnlock()
}
