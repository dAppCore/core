package core_test

import (
	"testing"
	"time"

	. "dappco.re/go/core"
)

// --- Mutex ---

func TestSync_Mutex_Good(t *testing.T) {
	var m Mutex
	m.Lock()
	m.Unlock()
}

func TestSync_Mutex_Bad(t *testing.T) {
	// Bad: TryLock on already-held mutex returns Result{OK: false}.
	var m Mutex
	m.Lock()
	defer m.Unlock()
	r := m.TryLock()
	AssertFalse(t, r.OK)
}

func TestSync_Mutex_Ugly(t *testing.T) {
	// Ugly: contention. Two goroutines incrementing under the same Mutex
	// must produce 1000 increments without races (-race must pass).
	var m Mutex
	count := 0
	var wg WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				m.Lock()
				count++
				m.Unlock()
			}
		}()
	}
	wg.Wait()
	AssertEqual(t, 1000, count)
}

// --- RWMutex ---

func TestSync_RWMutex_Good(t *testing.T) {
	var m RWMutex
	m.Lock()
	m.Unlock()
	m.RLock()
	m.RUnlock()
}

func TestSync_RWMutex_Bad(t *testing.T) {
	// Bad: TryLock fails when write-held.
	var m RWMutex
	m.Lock()
	defer m.Unlock()
	r := m.TryLock()
	AssertFalse(t, r.OK)
}

func TestSync_RWMutex_Ugly(t *testing.T) {
	// Ugly: many readers + occasional writer; -race must remain clean.
	var m RWMutex
	value := 0
	var wg WaitGroup
	// 5 readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.RLock()
				_ = value
				m.RUnlock()
			}
		}()
	}
	// 2 writers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				m.Lock()
				value++
				m.Unlock()
			}
		}()
	}
	wg.Wait()
	AssertEqual(t, 100, value)
}

func TestSync_RWMutex_TryRLock_Good(t *testing.T) {
	var m RWMutex
	r := m.TryRLock()
	AssertTrue(t, r.OK)
	m.RUnlock()
}

// --- Once ---

func TestSync_Once_Good(t *testing.T) {
	var o Once
	count := 0
	o.Do(func() { count++ })
	o.Do(func() { count++ })
	o.Do(func() { count++ })
	AssertEqual(t, 1, count, "Once.Do must execute the function exactly once")
}

func TestSync_Once_Bad(t *testing.T) {
	// Bad: caller passes nil. Stdlib Once panics on nil; we pass through.
	var o Once
	AssertPanics(t, func() { o.Do(nil) })
}

func TestSync_Once_Ugly(t *testing.T) {
	// Ugly: Reset between invocations re-arms the Once.
	var o Once
	count := 0
	o.Do(func() { count++ })
	o.Do(func() { count++ })
	AssertEqual(t, 1, count)
	o.Reset()
	o.Do(func() { count++ })
	o.Do(func() { count++ })
	AssertEqual(t, 2, count, "After Reset, Do must fire once more")
}

// --- WaitGroup ---

func TestSync_WaitGroup_Good(t *testing.T) {
	var wg WaitGroup
	var mu Mutex
	done := false
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		done = true
		mu.Unlock()
	}()
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	AssertTrue(t, done)
}

func TestSync_WaitGroup_Bad(t *testing.T) {
	// Bad: Done called more times than Add. Stdlib panics; we pass through.
	var wg WaitGroup
	wg.Add(1)
	wg.Done()
	AssertPanics(t, func() { wg.Done() })
}

func TestSync_WaitGroup_Ugly(t *testing.T) {
	// Ugly: many goroutines, all must complete before Wait returns.
	var wg WaitGroup
	var mu Mutex
	counter := 0
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	AssertEqual(t, 100, counter)
}
