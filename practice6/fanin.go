package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)

	go func() {
		// Always close the channel when the goroutine exits so
		// downstream readers know the server has stopped.
		defer close(out)

		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				fmt.Printf("[%s] shutting down (ctx: %v)\n", name, ctx.Err())
				return

			case out <- fmt.Sprintf("[%s] message #%d", name, i):
				// Message sent successfully; wait a little before next one.
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	return out
}

func FanIn(ctx context.Context, channels ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup

	forward := func(ch <-chan string) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					// Source channel has been closed.
					return
				}
				select {
				case merged <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	fmt.Println("=== Problem 3: Fan-In Pattern ===")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // always release resources

	alpha := startServer(ctx, "Alpha")
	beta := startServer(ctx, "Beta")
	gamma := startServer(ctx, "Gamma")

	stream := FanIn(ctx, alpha, beta, gamma)

	for msg := range stream {
		fmt.Println(msg)
	}

	fmt.Println()
	fmt.Println("All servers stopped. Fan-In complete.")
}
