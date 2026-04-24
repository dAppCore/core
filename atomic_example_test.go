package core_test

import (
	. "dappco.re/go/core"
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

func ExampleAtomicInt64() {
	var counter AtomicInt64
	counter.Add(1)
	counter.Add(1)
	counter.Add(1)
	Println(counter.Load())
	// Output:
	// 3
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
