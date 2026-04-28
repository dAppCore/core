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
//
//	var ready core.AtomicBool
//	ready.Store(true)
//	if ready.Load() { core.Println("agent ready") }
type AtomicBool struct{ inner atomic.Bool }

// Load returns the current value.
//
//	var ready core.AtomicBool
//	ready.Store(true)
//	current := ready.Load()
//	core.Println(current)
func (a *AtomicBool) Load() bool { return a.inner.Load() }

// Store sets the value.
//
//	var ready core.AtomicBool
//	ready.Store(true)
func (a *AtomicBool) Store(val bool) { a.inner.Store(val) }

// Swap stores new and returns the previous value.
//
//	var ready core.AtomicBool
//	previous := ready.Swap(true)
//	core.Println(previous)
func (a *AtomicBool) Swap(new bool) bool { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
//
//	if a.CompareAndSwap(false, true) { /* claimed */ }
func (a *AtomicBool) CompareAndSwap(old, new bool) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicInt32 is a race-free atomic int32. Zero value is 0.
//
//	var queueDepth core.AtomicInt32
//	queueDepth.Store(3)
//	core.Println(queueDepth.Load())
type AtomicInt32 struct{ inner atomic.Int32 }

// Load returns the current value.
//
//	var queueDepth core.AtomicInt32
//	queueDepth.Store(3)
//	depth := queueDepth.Load()
//	core.Println(depth)
func (a *AtomicInt32) Load() int32 { return a.inner.Load() }

// Store sets the value.
//
//	var queueDepth core.AtomicInt32
//	queueDepth.Store(3)
func (a *AtomicInt32) Store(val int32) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
//
//	new := a.Add(1)
func (a *AtomicInt32) Add(delta int32) int32 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
//
//	var queueDepth core.AtomicInt32
//	previous := queueDepth.Swap(5)
//	core.Println(previous)
func (a *AtomicInt32) Swap(new int32) int32 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
//
//	var queueDepth core.AtomicInt32
//	if queueDepth.CompareAndSwap(0, 1) {
//	    core.Println("worker claimed queue")
//	}
func (a *AtomicInt32) CompareAndSwap(old, new int32) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicInt64 is a race-free atomic int64. Zero value is 0.
//
//	var taskID core.AtomicInt64
//	next := taskID.Add(1)
//	core.Println(next)
type AtomicInt64 struct{ inner atomic.Int64 }

// Load returns the current value.
//
//	var taskID core.AtomicInt64
//	taskID.Store(42)
//	current := taskID.Load()
//	core.Println(current)
func (a *AtomicInt64) Load() int64 { return a.inner.Load() }

// Store sets the value.
//
//	var taskID core.AtomicInt64
//	taskID.Store(42)
func (a *AtomicInt64) Store(val int64) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
//
//	var taskID core.AtomicInt64
//	next := taskID.Add(1)
//	core.Println(next)
func (a *AtomicInt64) Add(delta int64) int64 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
//
//	var taskID core.AtomicInt64
//	previous := taskID.Swap(100)
//	core.Println(previous)
func (a *AtomicInt64) Swap(new int64) int64 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
//
//	var taskID core.AtomicInt64
//	if taskID.CompareAndSwap(0, 1) {
//	    core.Println("first task")
//	}
func (a *AtomicInt64) CompareAndSwap(old, new int64) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicUint32 is a race-free atomic uint32. Zero value is 0.
//
//	var workers core.AtomicUint32
//	workers.Add(1)
//	core.Println(workers.Load())
type AtomicUint32 struct{ inner atomic.Uint32 }

// Load returns the current value.
//
//	var workers core.AtomicUint32
//	workers.Store(4)
//	core.Println(workers.Load())
func (a *AtomicUint32) Load() uint32 { return a.inner.Load() }

// Store sets the value.
//
//	var workers core.AtomicUint32
//	workers.Store(4)
func (a *AtomicUint32) Store(val uint32) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
//
//	var workers core.AtomicUint32
//	workers.Add(1)
func (a *AtomicUint32) Add(delta uint32) uint32 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
//
//	var workers core.AtomicUint32
//	previous := workers.Swap(8)
//	core.Println(previous)
func (a *AtomicUint32) Swap(new uint32) uint32 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
//
//	var workers core.AtomicUint32
//	if workers.CompareAndSwap(0, 4) {
//	    core.Println("pool opened")
//	}
func (a *AtomicUint32) CompareAndSwap(old, new uint32) bool {
	return a.inner.CompareAndSwap(old, new)
}

// AtomicUint64 is a race-free atomic uint64. Zero value is 0.
//
//	var bytesSeen core.AtomicUint64
//	bytesSeen.Add(4096)
//	core.Println(bytesSeen.Load())
type AtomicUint64 struct{ inner atomic.Uint64 }

// Load returns the current value.
//
//	var bytesSeen core.AtomicUint64
//	bytesSeen.Store(4096)
//	core.Println(bytesSeen.Load())
func (a *AtomicUint64) Load() uint64 { return a.inner.Load() }

// Store sets the value.
//
//	var bytesSeen core.AtomicUint64
//	bytesSeen.Store(4096)
func (a *AtomicUint64) Store(val uint64) { a.inner.Store(val) }

// Add atomically adds delta and returns the new value.
//
//	var bytesSeen core.AtomicUint64
//	bytesSeen.Add(4096)
func (a *AtomicUint64) Add(delta uint64) uint64 { return a.inner.Add(delta) }

// Swap stores new and returns the previous value.
//
//	var bytesSeen core.AtomicUint64
//	previous := bytesSeen.Swap(8192)
//	core.Println(previous)
func (a *AtomicUint64) Swap(new uint64) uint64 { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current value equals old.
//
//	var bytesSeen core.AtomicUint64
//	if bytesSeen.CompareAndSwap(0, 4096) {
//	    core.Println("first payload")
//	}
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
//
//	var current core.AtomicPointer[core.ConfigOptions]
//	cfg := core.ConfigOptions{Settings: map[string]any{"config.host": "homelab.lthn.sh"}}
//	current.Store(&cfg)
//	loaded := current.Load()
//	_ = loaded
func (a *AtomicPointer[T]) Load() *T { return a.inner.Load() }

// Store sets the pointer.
//
//	var current core.AtomicPointer[core.ConfigOptions]
//	cfg := core.ConfigOptions{Settings: map[string]any{"config.host": "homelab.lthn.sh"}}
//	current.Store(&cfg)
func (a *AtomicPointer[T]) Store(val *T) { a.inner.Store(val) }

// Swap stores new and returns the previous pointer.
//
//	var current core.AtomicPointer[core.ConfigOptions]
//	next := core.ConfigOptions{Settings: map[string]any{"config.host": "homelab.lthn.sh"}}
//	previous := current.Swap(&next)
//	_ = previous
func (a *AtomicPointer[T]) Swap(new *T) *T { return a.inner.Swap(new) }

// CompareAndSwap stores new only if the current pointer equals old.
//
//	var current core.AtomicPointer[core.ConfigOptions]
//	old := core.ConfigOptions{}
//	next := core.ConfigOptions{Settings: map[string]any{"config.host": "homelab.lthn.sh"}}
//	current.Store(&old)
//	_ = current.CompareAndSwap(&old, &next)
func (a *AtomicPointer[T]) CompareAndSwap(old, new *T) bool {
	return a.inner.CompareAndSwap(old, new)
}
