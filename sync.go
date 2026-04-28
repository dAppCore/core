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

// Go starts fn in a new goroutine, automatically incrementing the
// counter before launch and decrementing it (via Done) after fn
// returns. Equivalent to Add(1) + go fn() + defer Done() but
// race-safe against the WaitGroup being reused.
//
//	var wg core.WaitGroup
//	wg.Go(func() { core.Println("agent done") })
//	wg.Wait()
func (w *WaitGroup) Go(fn func()) {
	w.inner.Add(1)
	go func() {
		defer w.inner.Done()
		fn()
	}()
}

// --- SyncMap: Concurrent map ---
//
// For most use cases, prefer map[K]V + core.RWMutex (type-safe, easier
// to reason about). Reach for SyncMap only in two patterns:
// (1) entries written once, read many times (caches that only grow), or
// (2) goroutines read/write/overwrite entries for disjoint key sets.
// Zero value is empty and ready for use. Must not be copied after first use.

// SyncMap is a concurrent map. Same semantics and memory model as sync.Map.
//
//	var cache core.SyncMap
//	cache.Store("config.host", "homelab.lthn.sh")
//	if value, ok := cache.Load("config.host"); ok {
//	    core.Println(value)
//	}
type SyncMap struct{ inner sync.Map }

// Load returns the value stored for key, or nil and ok=false if not present.
//
//	if v, ok := m.Load("key"); ok { /* use v */ }
func (m *SyncMap) Load(key any) (value any, ok bool) { return m.inner.Load(key) }

// Store sets the value for key.
//
//	m.Store("key", value)
func (m *SyncMap) Store(key, value any) { m.inner.Store(key, value) }

// LoadOrStore returns the existing value if present, otherwise stores and
// returns the given value. loaded is true if value was loaded, false if stored.
//
//	actual, loaded := m.LoadOrStore("key", defaultValue)
func (m *SyncMap) LoadOrStore(key, value any) (actual any, loaded bool) {
	return m.inner.LoadOrStore(key, value)
}

// LoadAndDelete deletes the value for key, returning the previous value if any.
//
//	v, loaded := m.LoadAndDelete("key")
func (m *SyncMap) LoadAndDelete(key any) (value any, loaded bool) {
	return m.inner.LoadAndDelete(key)
}

// Delete removes the value for key.
//
//	m.Delete("key")
func (m *SyncMap) Delete(key any) { m.inner.Delete(key) }

// Swap stores value for key and returns the previous value, if any.
//
//	previous, loaded := m.Swap("key", new)
func (m *SyncMap) Swap(key, value any) (previous any, loaded bool) {
	return m.inner.Swap(key, value)
}

// CompareAndSwap stores new for key only if the current value equals old.
//
//	if m.CompareAndSwap("key", old, new) { /* swapped */ }
func (m *SyncMap) CompareAndSwap(key, old, new any) (swapped bool) {
	return m.inner.CompareAndSwap(key, old, new)
}

// CompareAndDelete deletes the entry for key only if the current value equals old.
//
//	if m.CompareAndDelete("key", expected) { /* deleted */ }
func (m *SyncMap) CompareAndDelete(key, old any) (deleted bool) {
	return m.inner.CompareAndDelete(key, old)
}

// Range calls f sequentially for each key/value present. If f returns false,
// Range stops. f may be called concurrently with other operations on the map.
//
//	m.Range(func(k, v any) bool {
//	    Println(k, v)
//	    return true
//	})
func (m *SyncMap) Range(f func(key, value any) bool) { m.inner.Range(f) }

// Clear deletes all entries.
//
//	m.Clear()
func (m *SyncMap) Clear() { m.inner.Clear() }
