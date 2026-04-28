// SPDX-License-Identifier: EUPL-1.2

// Per-instance synchronisation primitives — embed in your structs to avoid
// importing "sync" directly.
//
// These are 1:1 wrappers around stdlib sync types. Methods are pass-through;
// the only purpose is to confine the "sync" import to this file so consumers
// can satisfy the banned-imports rule without losing concurrency primitives.
//
// For per-Core named coordinators (drain, service-registry, etc.) use
// c.Lock(name) — that's a different shape (registry-style, named).
//
// Usage:
//
//	type Counter struct {
//	    mu    core.Mutex      // protects value
//	    value int
//	}
//
//	func (c *Counter) Inc() {
//	    c.mu.Lock(); defer c.mu.Unlock()
//	    c.value++
//	}

package core

import "sync"

// Mutex is a mutual exclusion lock. Embed in your struct to protect internal state.
//
//	type Counter struct {
//	    mu    core.Mutex
//	    value int
//	}
//
//	func (c *Counter) Inc() {
//	    c.mu.Lock(); defer c.mu.Unlock()
//	    c.value++
//	}
type Mutex struct{ inner sync.Mutex }

// Lock acquires the mutex.
//
//	var mu core.Mutex
//	mu.Lock()
//	defer mu.Unlock()
func (m *Mutex) Lock() { m.inner.Lock() }

// Unlock releases the mutex.
//
//	var mu core.Mutex
//	mu.Lock()
//	mu.Unlock()
func (m *Mutex) Unlock() { m.inner.Unlock() }

// TryLock attempts to acquire the mutex without blocking.
// Returns Result{OK: true} on acquire, Result{OK: false} if already held.
//
//	if c.mu.TryLock().OK {
//	    defer c.mu.Unlock()
//	    // ...
//	}
func (m *Mutex) TryLock() Result {
	if m.inner.TryLock() {
		return Result{OK: true}
	}
	return Result{OK: false}
}

// RWMutex is a read-write mutex. Many readers OR one writer.
//
//	type Cache struct {
//	    mu   core.RWMutex
//	    data map[string]string
//	}
//
//	func (c *Cache) Get(k string) string {
//	    c.mu.RLock(); defer c.mu.RUnlock()
//	    return c.data[k]
//	}
//
//	func (c *Cache) Set(k, v string) {
//	    c.mu.Lock(); defer c.mu.Unlock()
//	    c.data[k] = v
//	}
type RWMutex struct{ inner sync.RWMutex }

// Lock acquires the mutex for write (exclusive).
//
//	var mu core.RWMutex
//	mu.Lock()
//	defer mu.Unlock()
func (m *RWMutex) Lock() { m.inner.Lock() }

// Unlock releases the mutex from write.
//
//	var mu core.RWMutex
//	mu.Lock()
//	mu.Unlock()
func (m *RWMutex) Unlock() { m.inner.Unlock() }

// RLock acquires the mutex for read (shared).
//
//	var mu core.RWMutex
//	mu.RLock()
//	defer mu.RUnlock()
func (m *RWMutex) RLock() { m.inner.RLock() }

// RUnlock releases the mutex from read.
//
//	var mu core.RWMutex
//	mu.RLock()
//	mu.RUnlock()
func (m *RWMutex) RUnlock() { m.inner.RUnlock() }

// TryLock attempts to acquire the write mutex without blocking.
//
//	if c.mu.TryLock().OK { defer c.mu.Unlock() }
func (m *RWMutex) TryLock() Result {
	if m.inner.TryLock() {
		return Result{OK: true}
	}
	return Result{OK: false}
}

// TryRLock attempts to acquire the read mutex without blocking.
//
//	if c.mu.TryRLock().OK { defer c.mu.RUnlock() }
func (m *RWMutex) TryRLock() Result {
	if m.inner.TryRLock() {
		return Result{OK: true}
	}
	return Result{OK: false}
}

// Once runs a function exactly once across all callers.
//
//	type Service struct {
//	    initOnce core.Once
//	    ready    bool
//	}
//
//	func (s *Service) ensure() {
//	    s.initOnce.Do(func() { s.ready = true })
//	}
type Once struct{ inner sync.Once }

// Do calls the function fn if and only if Do is being called for the first
// time for this instance of Once.
//
//	var once core.Once
//	once.Do(func() { core.Println("agent init") })
func (o *Once) Do(fn func()) { o.inner.Do(fn) }

// Reset clears the once so Do can fire again. Use for re-initialisation
// patterns where the resource is closed and re-opened.
//
//	s.closeStateStore()
//	s.initOnce.Reset()  // next Do() runs the init function again
//
// Semantics match stdlib sync.Once{} reset: the previous Once is replaced
// outright. If a Do(fn) is concurrently in flight, behaviour is undefined —
// callers must serialise Reset against any concurrent Do calls.
func (o *Once) Reset() { o.inner = sync.Once{} }

// WaitGroup waits for a collection of goroutines to finish.
//
//	var wg core.WaitGroup
//	for _, item := range items {
//	    wg.Add(1)
//	    go func(it Item) {
//	        defer wg.Done()
//	        process(it)
//	    }(item)
//	}
//	wg.Wait()
type WaitGroup struct{ inner sync.WaitGroup }

// Add adds delta, which may be negative, to the WaitGroup counter.
//
//	var wg core.WaitGroup
//	wg.Add(1)
//	go func() { defer wg.Done(); core.Println("agent done") }()
//	wg.Wait()
func (w *WaitGroup) Add(delta int) { w.inner.Add(delta) }

// Done decrements the WaitGroup counter by one.
//
//	var wg core.WaitGroup
//	wg.Add(1)
//	go func() { defer wg.Done(); core.Println("agent done") }()
//	wg.Wait()
func (w *WaitGroup) Done() { w.inner.Done() }

// Wait blocks until the WaitGroup counter is zero.
//
//	var wg core.WaitGroup
//	wg.Add(1)
//	go func() { defer wg.Done(); core.Println("agent done") }()
//	wg.Wait()
func (w *WaitGroup) Wait() { w.inner.Wait() }
