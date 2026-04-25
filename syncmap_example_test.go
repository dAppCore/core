package core_test

import (
	. "dappco.re/go/core"
)

func ExampleSyncMap() {
	var m SyncMap
	m.Store("greeting", "hello")
	if v, ok := m.Load("greeting"); ok {
		Println(v)
	}
	// Output:
	// hello
}

func ExampleSyncMap_LoadOrStore() {
	var m SyncMap
	actual, loaded := m.LoadOrStore("k", "first")
	Println(actual, loaded)
	actual, loaded = m.LoadOrStore("k", "second")
	Println(actual, loaded)
	// Output:
	// first false
	// first true
}

func ExampleSyncMap_CompareAndSwap() {
	var m SyncMap
	m.Store("flag", false)
	if m.CompareAndSwap("flag", false, true) {
		Println("claimed")
	}
	// Output:
	// claimed
}
