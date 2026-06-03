package main

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"golang.org/x/sync/semaphore"
)

func main() {
	ctx := context.Background()

	maxWorkers := runtime.GOMAXPROCS(0)
	sem := semaphore.NewWeighted(int64(maxWorkers))

	out := make([]int, 32)

	for i := range out {
		// Acquire 1 token before starting a goroutine.
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Fatalf("failed to acquire semaphore: %v", err)
		}

		go func(i int) {
			defer sem.Release(1)

			out[i] = i + 1
		}(i)
	}

	// Acquire all tokens to wait until all goroutines release their tokens.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.Fatalf("failed to acquire semaphore while waiting: %v", err)
	}

	fmt.Println(out)
}
