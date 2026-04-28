package core_test

import (
	. "dappco.re/go"
)

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

func ExampleAtomicBool_Swap() {
	var ready AtomicBool
	ready.Store(true)
	Println(ready.Swap(false))
	Println(ready.Load())
	// Output:
	// true
	// false
}

func ExampleAtomicBool_CompareAndSwap() {
	var ready AtomicBool
	Println(ready.CompareAndSwap(false, true))
	Println(ready.Load())
	// Output:
	// true
	// true
}

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

func ExampleAtomicInt64() {
	var counter AtomicInt64
	counter.Add(1)
	counter.Add(1)
	counter.Add(1)
	Println(counter.Load())
	// Output:
	// 3
}

func ExampleAtomicInt64_Swap() {
	var counter AtomicInt64
	counter.Store(7)
	Println(counter.Swap(9))
	Println(counter.Load())
	// Output:
	// 7
	// 9
}

func ExampleAtomicInt64_CompareAndSwap() {
	var counter AtomicInt64
	counter.Store(7)
	Println(counter.CompareAndSwap(7, 9))
	Println(counter.Load())
	// Output:
	// true
	// 9
}

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

func ExampleAtomicUint32_CompareAndSwap() {
	var counter AtomicUint32
	counter.Store(10)
	Println(counter.CompareAndSwap(10, 11))
	Println(counter.Load())
	// Output:
	// true
	// 11
}

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
