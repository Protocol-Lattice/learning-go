package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func job(ctx context.Context, id int) error {
	select {
	case <-time.After(time.Second):
		fmt.Println("done:", id)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx := context.Background()

	g, ctx := errgroup.WithContext(ctx)

	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent)

	for i := 1; i <= 10; i++ {
		i := i

		g.Go(func() error {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return ctx.Err()
			}

			return job(ctx, i)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("all jobs completed")
}
