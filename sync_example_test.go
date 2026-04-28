package core_test

import (
	. "dappco.re/go"
)

// ExampleMutex uses a mutual exclusion lock through `Mutex` for concurrent service
// coordination. Concurrency helpers mirror the stdlib shapes while keeping ownership in
// core.
func ExampleMutex() {
	var mu Mutex
	mu.Lock()
	Println("locked")
	mu.Unlock()
	Println("unlocked")
	// Output:
	// locked
	// unlocked
}

// ExampleMutex_TryLock attempts a non-blocking write lock through `Mutex.TryLock` for
// concurrent service coordination. Concurrency helpers mirror the stdlib shapes while
// keeping ownership in core.
func ExampleMutex_TryLock() {
	var mu Mutex
	if mu.TryLock().OK {
		Println("acquired")
		mu.Unlock()
	}
	// Output: acquired
}

// ExampleRWMutex uses a read-write lock through `RWMutex` for concurrent service
// coordination. Concurrency helpers mirror the stdlib shapes while keeping ownership in
// core.
func ExampleRWMutex() {
	var mu RWMutex
	mu.RLock()
	Println("read-locked")
	mu.RUnlock()
	mu.Lock()
	Println("write-locked")
	mu.Unlock()
	// Output:
	// read-locked
	// write-locked
}

// ExampleRWMutex_TryLock attempts a non-blocking write lock through `RWMutex.TryLock` for
// concurrent service coordination. Concurrency helpers mirror the stdlib shapes while
// keeping ownership in core.
func ExampleRWMutex_TryLock() {
	var mu RWMutex
	if mu.TryLock().OK {
		Println("write")
		mu.Unlock()
	}
	// Output: write
}

// ExampleRWMutex_TryRLock attempts a non-blocking read lock through `RWMutex.TryRLock` for
// concurrent service coordination. Concurrency helpers mirror the stdlib shapes while
// keeping ownership in core.
func ExampleRWMutex_TryRLock() {
	var mu RWMutex
	if mu.TryRLock().OK {
		Println("read")
		mu.RUnlock()
	}
	// Output: read
}

// ExampleOnce runs work once through `Once` for concurrent service coordination.
// Concurrency helpers mirror the stdlib shapes while keeping ownership in core.
func ExampleOnce() {
	var o Once
	o.Do(func() { Println("ran") })
	o.Do(func() { Println("not printed") })
	// Output:
	// ran
}

// ExampleOnce_Reset resets one-shot state through `Once.Reset` for concurrent service
// coordination. Concurrency helpers mirror the stdlib shapes while keeping ownership in
// core.
func ExampleOnce_Reset() {
	var o Once
	o.Do(func() { Println("first") })
	o.Reset()
	o.Do(func() { Println("again") })
	// Output:
	// first
	// again
}

// ExampleWaitGroup waits for concurrent work through `WaitGroup` for concurrent service
// coordination. Concurrency helpers mirror the stdlib shapes while keeping ownership in
// core.
func ExampleWaitGroup() {
	var wg WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Println("worker done")
	}()
	wg.Wait()
	// Output:
	// worker done
}
