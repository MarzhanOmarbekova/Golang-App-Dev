package main

import (
	"fmt"
	"sync"
)

func runSyncMap() {
	var sm sync.Map
	var wg sync.WaitGroup

	const goroutines = 100
	const key = "counter"

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			sm.Store(key, val)
		}(i)
	}

	wg.Wait()

	if v, ok := sm.Load(key); ok {
		fmt.Printf("[sync.Map]   final value of %q = %v\n", key, v)
	}
}

type safeMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func (s *safeMap) set(key string, val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = val
}

func (s *safeMap) get(key string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

func runRWMutexMap() {
	sm := &safeMap{m: make(map[string]int)}
	var wg sync.WaitGroup

	const goroutines = 100
	const key = "counter"

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			sm.set(key, val)
		}(i)
	}

	wg.Wait()

	if v, ok := sm.get(key); ok {
		fmt.Printf("[RWMutex map] final value of %q = %v\n", key, v)
	}
}

func main() {
	fmt.Println("Problem 1: Thread-Safe Map")
	fmt.Println()

	runSyncMap()
	runRWMutexMap()

	fmt.Println()
	fmt.Println("Both solutions completed with no data races.")
}
