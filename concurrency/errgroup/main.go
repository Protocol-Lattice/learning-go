package main

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

func main() {
	var g errgroup.Group

	for i := 0; i < 5; i++ {
		i := i

		g.Go(func() error {
			fmt.Println("worker:", i)

			if i == 3 {
				return fmt.Errorf("worker %d failed", i)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("all workers finished successfully")
}
