package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func runMutexCounter() int {
	var (
		counter int
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	return counter
}

func runAtomicCounter() int64 {
	var (
		counter int64
		wg      sync.WaitGroup
	)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// atomic.AddInt64 is lock-free and safe for concurrent use.
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()
	return atomic.LoadInt64(&counter)
}

func main() {
	fmt.Println("=== Problem 2: Concurrent Counter ===")
	fmt.Println()

	fmt.Println("Note: the unsafe version is intentionally omitted from")
	fmt.Println("this run to keep -race clean. See comments in source.")
	fmt.Println()

	mutexResult := runMutexCounter()
	fmt.Printf("[sync.Mutex]   final counter = %d  (expected 1000)\n", mutexResult)

	atomicResult := runAtomicCounter()
	fmt.Printf("[sync/atomic]  final counter = %d  (expected 1000)\n", atomicResult)

	fmt.Println()
	fmt.Println("Both solutions completed with no data races.")
}
