package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func doSomethingUnreliable() error {
	if rand.Intn(10) < 7 {
		fmt.Println("Operation failed, retrying...")
		return errors.New("temporary failure")
	}
	fmt.Println("Operation succeeded!")
	return nil
}

// Example 1: Naive infinite loop (anti-pattern, shown for education)
func example1NaiveLoop() {
	fmt.Println("\n  Example 1: Naive Infinite Loop (Anti-Pattern)  ")
	var err error
	attempts := 0
	for {
		err = doSomethingUnreliable()
		attempts++
		if err == nil {
			break
		}
		if attempts >= 3 {
			fmt.Println("Stopping naive loop after 3 attempts for demo purposes")
			break
		}
	}
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
	}
}

// Example 2: Fixed delay
func example2FixedDelay() {
	fmt.Println("\n  Example 2: Fixed Delay  ")
	var err error
	const maxRetries = 5
	const delay = 300 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
		fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, delay)
		if attempt < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	if err != nil {
		fmt.Printf("Failed after %d attempts: %v\n", maxRetries, err)
	} else {
		fmt.Println("Succeeded within retry limit.")
	}
}

// Example 3: Exponential backoff
func example3ExponentialBackoff() {
	fmt.Println("\n  Example 3: Exponential Backoff  ")
	var err error
	const maxRetries = 5
	baseDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
		if attempt == maxRetries-1 {
			break
		}
		backoffTime := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoffTime > maxDelay {
			backoffTime = maxDelay
		}
		fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, backoffTime)
		time.Sleep(backoffTime)
	}
	if err != nil {
		fmt.Printf("Failed after %d attempts: %v\n", maxRetries, err)
	} else {
		fmt.Println("Succeeded!")
	}
}

// Example 4: Exponential backoff + Full Jitter
func example4FullJitter() {
	fmt.Println("\n  Example 4: Exponential Backoff + Full Jitter  ")
	var err error
	const maxRetries = 5
	baseDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
		if attempt == maxRetries-1 {
			break
		}
		backoffTime := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoffTime > maxDelay {
			backoffTime = maxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoffTime)))
		fmt.Printf("Attempt %d failed, waiting ~%v (backoff %v + jitter) before next retry...\n", attempt+1, jitter, backoffTime)
		time.Sleep(jitter)
	}
	if err != nil {
		fmt.Printf("Failed after %d attempts: %v\n", maxRetries, err)
	} else {
		fmt.Println("Succeeded!")
	}
}

// Example 5: Context + cancellation
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func retryWithContext(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var err error
	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if attempt == cfg.MaxRetries-1 {
			return err
		}
		backoff := cfg.BaseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > cfg.MaxDelay {
			backoff = cfg.MaxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoff)))
		fmt.Printf("Attempt %d failed, waiting ~%v (max backoff: %v)...\n", attempt+1, jitter, backoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jitter):
		}
	}
	return err
}

func example5Context() {
	fmt.Println("\n  Example 5: Context + Cancellation  ")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cfg := RetryConfig{
		MaxRetries: 10,
		BaseDelay:  200 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	}

	err := retryWithContext(ctx, cfg, doSomethingUnreliable)
	if err != nil {
		fmt.Printf("Final error: %v\n", err)
	} else {
		fmt.Println("Succeeded with context!")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	example1NaiveLoop()
	example2FixedDelay()
	example3ExponentialBackoff()
	example4FullJitter()
	example5Context()
}
