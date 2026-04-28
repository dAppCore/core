// SPDX-License-Identifier: EUPL-1.2

// Concurrent map primitive — wrapper over sync.Map.
//
// For most use cases, prefer map[K]V + core.RWMutex (type-safe, easier to
// reason about). Reach for SyncMap only in two patterns:
// (1) entries written once, read many times (caches that only grow), or
// (2) goroutines read/write/overwrite entries for disjoint key sets.
// See sync.Map's package docs for the full memory-model guarantees.
//
// Zero value is empty and ready for use. Must not be copied after first use.
//
// Usage:
//
//	var cache core.SyncMap
//	cache.Store("key", value)
//	if v, ok := cache.Load("key"); ok {
//	    use(v)
//	}

package core

import "sync"

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
