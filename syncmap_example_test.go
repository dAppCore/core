package core_test

import (
	. "dappco.re/go"
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

func ExampleSyncMap_Load() {
	var m SyncMap
	m.Store("greeting", "hello")
	value, ok := m.Load("greeting")
	Println(value)
	Println(ok)
	// Output:
	// hello
	// true
}

func ExampleSyncMap_Store() {
	var m SyncMap
	m.Store("greeting", "hello")
	value, _ := m.Load("greeting")
	Println(value)
	// Output: hello
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

func ExampleSyncMap_LoadAndDelete() {
	var m SyncMap
	m.Store("greeting", "hello")
	value, loaded := m.LoadAndDelete("greeting")
	_, exists := m.Load("greeting")
	Println(value, loaded)
	Println(exists)
	// Output:
	// hello true
	// false
}

func ExampleSyncMap_Delete() {
	var m SyncMap
	m.Store("greeting", "hello")
	m.Delete("greeting")
	_, ok := m.Load("greeting")
	Println(ok)
	// Output: false
}

func ExampleSyncMap_Swap() {
	var m SyncMap
	m.Store("greeting", "hello")
	previous, loaded := m.Swap("greeting", "hi")
	current, _ := m.Load("greeting")
	Println(previous, loaded)
	Println(current)
	// Output:
	// hello true
	// hi
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

func ExampleSyncMap_CompareAndDelete() {
	var m SyncMap
	m.Store("flag", true)
	Println(m.CompareAndDelete("flag", true))
	_, ok := m.Load("flag")
	Println(ok)
	// Output:
	// true
	// false
}

func ExampleSyncMap_Range() {
	var m SyncMap
	m.Store("b", 2)
	m.Store("a", 1)

	var keys []string
	m.Range(func(key, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	SliceSort(keys)
	Println(keys)
	// Output: [a b]
}

func ExampleSyncMap_Clear() {
	var m SyncMap
	m.Store("greeting", "hello")
	m.Clear()
	_, ok := m.Load("greeting")
	Println(ok)
	// Output: false
}
