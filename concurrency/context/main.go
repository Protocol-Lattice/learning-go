package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// context propagates cancellation signals and values across API boundaries and goroutines.
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "key", 100))
	go func() {
		cancel() // Cancel the context after some work is done
	}()
	select {
	case <-ctx.Done():
		println("Context cancelled")
		value := ctx.Value("key")
		println("Value from context:", value.(int))
	}

	ctx2, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	cancel() // cancel before deadline

	select {
	case <-ctx2.Done():
		if ctx2.Err() != context.DeadlineExceeded {
			fmt.Printf("expected DeadlineExceeded, got %v\n", ctx2.Err())
		}

	case <-time.After(200 * time.Millisecond):
		fmt.Println("context deadline did not trigger")
	}

	ctx3, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	cancel()
	select {
	case <-ctx3.Done():
		if ctx3.Err() != context.DeadlineExceeded {
			fmt.Printf("expected DeadlineExceeded, got %v\n", ctx3.Err())
		}
	}
}
