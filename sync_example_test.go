package core_test

import (
	. "dappco.re/go"
)

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

func ExampleMutex_TryLock() {
	var mu Mutex
	if mu.TryLock().OK {
		Println("acquired")
		mu.Unlock()
	}
	// Output: acquired
}

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

func ExampleRWMutex_TryLock() {
	var mu RWMutex
	if mu.TryLock().OK {
		Println("write")
		mu.Unlock()
	}
	// Output: write
}

func ExampleRWMutex_TryRLock() {
	var mu RWMutex
	if mu.TryRLock().OK {
		Println("read")
		mu.RUnlock()
	}
	// Output: read
}

func ExampleOnce() {
	var o Once
	o.Do(func() { Println("ran") })
	o.Do(func() { Println("not printed") })
	// Output:
	// ran
}

func ExampleOnce_Reset() {
	var o Once
	o.Do(func() { Println("first") })
	o.Reset()
	o.Do(func() { Println("again") })
	// Output:
	// first
	// again
}

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
