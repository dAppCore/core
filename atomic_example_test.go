package core_test

import (
	. "dappco.re/go"
)

// ExampleAtomicBool updates a boolean atomically through `AtomicBool` for shared runtime
// state. Concurrent state changes use explicit load, store, swap, and compare-and-swap
// shapes.
func ExampleAtomicBool() {
	var ready AtomicBool
	if !ready.Load() {
		Println("not ready")
	}
	ready.Store(true)
	if ready.Load() {
		Println("ready")
	}
	// Output:
	// not ready
	// ready
}

// ExampleAtomicBool_Swap swaps a value through `AtomicBool.Swap` and returns the previous
// value for shared runtime state. Concurrent state changes use explicit load, store, swap,
// and compare-and-swap shapes.
func ExampleAtomicBool_Swap() {
	var ready AtomicBool
	ready.Store(true)
	Println(ready.Swap(false))
	Println(ready.Load())
	// Output:
	// true
	// false
}

// ExampleAtomicBool_CompareAndSwap updates `AtomicBool.CompareAndSwap` only when the
// previous value matches for shared runtime state. Concurrent state changes use explicit
// load, store, swap, and compare-and-swap shapes.
func ExampleAtomicBool_CompareAndSwap() {
	var ready AtomicBool
	Println(ready.CompareAndSwap(false, true))
	Println(ready.Load())
	// Output:
	// true
	// true
}

// ExampleAtomicInt32 updates a 32-bit integer atomically through `AtomicInt32` for shared
// runtime state. Concurrent state changes use explicit load, store, swap, and
// compare-and-swap shapes.
func ExampleAtomicInt32() {
	var counter AtomicInt32
	counter.Store(10)
	Println(counter.Add(5))
	Println(counter.Swap(1))
	Println(counter.Load())
	// Output:
	// 15
	// 15
	// 1
}

// ExampleAtomicInt64 updates a 64-bit integer atomically through `AtomicInt64` for shared
// runtime state. Concurrent state changes use explicit load, store, swap, and
// compare-and-swap shapes.
func ExampleAtomicInt64() {
	var counter AtomicInt64
	counter.Add(1)
	counter.Add(1)
	counter.Add(1)
	Println(counter.Load())
	// Output:
	// 3
}

// ExampleAtomicInt64_Swap swaps a value through `AtomicInt64.Swap` and returns the
// previous value for shared runtime state. Concurrent state changes use explicit load,
// store, swap, and compare-and-swap shapes.
func ExampleAtomicInt64_Swap() {
	var counter AtomicInt64
	counter.Store(7)
	Println(counter.Swap(9))
	Println(counter.Load())
	// Output:
	// 7
	// 9
}

// ExampleAtomicInt64_CompareAndSwap updates `AtomicInt64.CompareAndSwap` only when the
// previous value matches for shared runtime state. Concurrent state changes use explicit
// load, store, swap, and compare-and-swap shapes.
func ExampleAtomicInt64_CompareAndSwap() {
	var counter AtomicInt64
	counter.Store(7)
	Println(counter.CompareAndSwap(7, 9))
	Println(counter.Load())
	// Output:
	// true
	// 9
}

// ExampleAtomicInt32_CompareAndSwap updates `AtomicInt32.CompareAndSwap` only when the
// previous value matches for shared runtime state. Concurrent state changes use explicit
// load, store, swap, and compare-and-swap shapes.
func ExampleAtomicInt32_CompareAndSwap() {
	var state AtomicInt32
	if state.CompareAndSwap(0, 1) {
		Println("claimed")
	}
	if !state.CompareAndSwap(0, 2) {
		Println("already claimed")
	}
	// Output:
	// claimed
	// already claimed
}

// ExampleAtomicUint32 updates an unsigned 32-bit integer atomically through `AtomicUint32`
// for shared runtime state. Concurrent state changes use explicit load, store, swap, and
// compare-and-swap shapes.
func ExampleAtomicUint32() {
	var counter AtomicUint32
	counter.Store(10)
	Println(counter.Add(5))
	Println(counter.Swap(1))
	Println(counter.Load())
	// Output:
	// 15
	// 15
	// 1
}

// ExampleAtomicUint32_CompareAndSwap updates `AtomicUint32.CompareAndSwap` only when the
// previous value matches for shared runtime state. Concurrent state changes use explicit
// load, store, swap, and compare-and-swap shapes.
func ExampleAtomicUint32_CompareAndSwap() {
	var counter AtomicUint32
	counter.Store(10)
	Println(counter.CompareAndSwap(10, 11))
	Println(counter.Load())
	// Output:
	// true
	// 11
}

// ExampleAtomicUint64 updates an unsigned 64-bit integer atomically through `AtomicUint64`
// for shared runtime state. Concurrent state changes use explicit load, store, swap, and
// compare-and-swap shapes.
func ExampleAtomicUint64() {
	var counter AtomicUint64
	counter.Store(10)
	Println(counter.Add(5))
	Println(counter.Swap(1))
	Println(counter.Load())
	// Output:
	// 15
	// 15
	// 1
}

// ExampleAtomicUint64_CompareAndSwap updates `AtomicUint64.CompareAndSwap` only when the
// previous value matches for shared runtime state. Concurrent state changes use explicit
// load, store, swap, and compare-and-swap shapes.
func ExampleAtomicUint64_CompareAndSwap() {
	var counter AtomicUint64
	counter.Store(10)
	Println(counter.CompareAndSwap(10, 11))
	Println(counter.Load())
	// Output:
	// true
	// 11
}

type config struct {
	name string
}

// ExampleAtomicPointer updates a typed pointer atomically through `AtomicPointer` for
// shared runtime state. Concurrent state changes use explicit load, store, swap, and
// compare-and-swap shapes.
func ExampleAtomicPointer() {
	var current AtomicPointer[config]
	current.Store(&config{name: "v1"})
	cfg := current.Load()
	Println(cfg.name)
	current.Store(&config{name: "v2"})
	Println(current.Load().name)
	// Output:
	// v1
	// v2
}

// ExampleAtomicPointer_Swap swaps a value through `AtomicPointer.Swap` and returns the
// previous value for shared runtime state. Concurrent state changes use explicit load,
// store, swap, and compare-and-swap shapes.
func ExampleAtomicPointer_Swap() {
	var current AtomicPointer[config]
	first := &config{name: "v1"}
	second := &config{name: "v2"}
	current.Store(first)

	Println(current.Swap(second).name)
	Println(current.Load().name)
	// Output:
	// v1
	// v2
}

// ExampleAtomicPointer_CompareAndSwap updates `AtomicPointer.CompareAndSwap` only when the
// previous value matches for shared runtime state. Concurrent state changes use explicit
// load, store, swap, and compare-and-swap shapes.
func ExampleAtomicPointer_CompareAndSwap() {
	var current AtomicPointer[config]
	first := &config{name: "v1"}
	second := &config{name: "v2"}
	current.Store(first)

	Println(current.CompareAndSwap(first, second))
	Println(current.Load().name)
	// Output:
	// true
	// v2
}
