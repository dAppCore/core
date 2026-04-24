// SPDX-License-Identifier: EUPL-1.2

// Typed atomic primitives — wrappers over sync/atomic typed types.
//
// Each type is a value-type wrapper; zero value is the natural zero
// (false / 0 / nil). Must not be copied after first use.
//
// Mirrors Go 1.19+ stdlib typed atomics one-for-one. Method names and
// semantics are pass-through.
//
// Usage:
//
//	var ready core.AtomicBool
//	ready.Store(true)
//	if ready.Load() { /* signal received */ }
//
//	var counter core.AtomicInt64
//	counter.Add(1)
//	value := counter.Load()
//
//	var current core.AtomicPointer[Config]
//	current.Store(&Config{...})
//	cfg := current.Load()

package core

import "sync/atomic"

// AtomicBool is a race-free atomic boolean. Zero value is false.
type AtomicBool struct{ inner atomic.Bool }

// Load returns the current value.
func (a *AtomicBool) Load() bool { return a.inner.Load() }

// Store sets the value.
func (a *AtomicBool) Store(val bool) { a.inner.Store(val) }

// Swap stores new and returns the previous value.
func (a *AtomicBool) Swap(new bool) bool { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
//
//	if a.CompareAndSwap(false, true) { /* claimed */ }
func (a *AtomicBool) CompareAndSwap(old, new bool) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicInt32 is a race-free atomic int32. Zero value is 0.
type AtomicInt32 struct{ inner atomic.Int32 }

// Load returns the current value.
func (a *AtomicInt32) Load() int32 { return a.inner.Load() }

// Store sets the value.
func (a *AtomicInt32) Store(val int32) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
//
//	new := a.Add(1)
func (a *AtomicInt32) Add(delta int32) int32 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
func (a *AtomicInt32) Swap(new int32) int32 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
func (a *AtomicInt32) CompareAndSwap(old, new int32) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicInt64 is a race-free atomic int64. Zero value is 0.
type AtomicInt64 struct{ inner atomic.Int64 }

// Load returns the current value.
func (a *AtomicInt64) Load() int64 { return a.inner.Load() }

// Store sets the value.
func (a *AtomicInt64) Store(val int64) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
func (a *AtomicInt64) Add(delta int64) int64 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
func (a *AtomicInt64) Swap(new int64) int64 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
func (a *AtomicInt64) CompareAndSwap(old, new int64) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicUint32 is a race-free atomic uint32. Zero value is 0.
type AtomicUint32 struct{ inner atomic.Uint32 }

// Load returns the current value.
func (a *AtomicUint32) Load() uint32 { return a.inner.Load() }

// Store sets the value.
func (a *AtomicUint32) Store(val uint32) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
func (a *AtomicUint32) Add(delta uint32) uint32 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
func (a *AtomicUint32) Swap(new uint32) uint32 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
func (a *AtomicUint32) CompareAndSwap(old, new uint32) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicUint64 is a race-free atomic uint64. Zero value is 0.
type AtomicUint64 struct{ inner atomic.Uint64 }

// Load returns the current value.
func (a *AtomicUint64) Load() uint64 { return a.inner.Load() }

// Store sets the value.
func (a *AtomicUint64) Store(val uint64) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
func (a *AtomicUint64) Add(delta uint64) uint64 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
func (a *AtomicUint64) Swap(new uint64) uint64 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
func (a *AtomicUint64) CompareAndSwap(old, new uint64) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicPointer is a race-free atomic *T. Zero value is nil.
//
//	var current core.AtomicPointer[Config]
//	current.Store(&Config{...})
//	cfg := current.Load()
type AtomicPointer[T any] struct{ inner atomic.Pointer[T] }

// Load returns the current pointer.
func (a *AtomicPointer[T]) Load() *T { return a.inner.Load() }

// Store sets the pointer.
func (a *AtomicPointer[T]) Store(val *T) { a.inner.Store(val) }

// Swap stores new and returns the previous pointer.
func (a *AtomicPointer[T]) Swap(new *T) *T { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current pointer equals old.
func (a *AtomicPointer[T]) CompareAndSwap(old, new *T) bool {
	return a.inner.CompareAndSwap(old, new)
}
