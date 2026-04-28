package core_test

import (
	. "dappco.re/go/core"
)

func TestSyncMap_Store_Good(t *T) {
	var m SyncMap
	m.Store("k", 42)
	v, ok := m.Load("k")
	AssertTrue(t, ok)
	AssertEqual(t, 42, v)
}

func TestSyncMap_Store_Bad(t *T) {
	// Bad: Load on absent key returns nil, false.
	var m SyncMap
	v, ok := m.Load("missing")
	AssertFalse(t, ok)
	AssertNil(t, v)
}

func TestSyncMap_Store_Ugly(t *T) {
	// Ugly: overwrite via Store, then via Swap, observe the chain.
	var m SyncMap
	m.Store("k", "first")
	m.Store("k", "second")
	prev, loaded := m.Swap("k", "third")
	AssertTrue(t, loaded)
	AssertEqual(t, "second", prev)
	v, _ := m.Load("k")
	AssertEqual(t, "third", v)
}

func TestSyncMap_LoadOrStore_Good(t *T) {
	var m SyncMap
	actual, loaded := m.LoadOrStore("k", 1)
	AssertFalse(t, loaded)
	AssertEqual(t, 1, actual)
	actual, loaded = m.LoadOrStore("k", 2)
	AssertTrue(t, loaded)
	AssertEqual(t, 1, actual, "second LoadOrStore returns existing value")
}

func TestSyncMap_LoadAndDelete_Good(t *T) {
	var m SyncMap
	m.Store("k", "v")
	v, loaded := m.LoadAndDelete("k")
	AssertTrue(t, loaded)
	AssertEqual(t, "v", v)
	_, ok := m.Load("k")
	AssertFalse(t, ok)
}

func TestSyncMap_CompareAndSwap_Good(t *T) {
	var m SyncMap
	m.Store("k", 1)
	swapped := m.CompareAndSwap("k", 1, 2)
	AssertTrue(t, swapped)
	v, _ := m.Load("k")
	AssertEqual(t, 2, v)
}

func TestSyncMap_CompareAndSwap_Bad(t *T) {
	// Bad: CompareAndSwap with mismatched old returns false, no change.
	var m SyncMap
	m.Store("k", 1)
	swapped := m.CompareAndSwap("k", 99, 2)
	AssertFalse(t, swapped)
	v, _ := m.Load("k")
	AssertEqual(t, 1, v)
}

func TestSyncMap_Range_Good(t *T) {
	var m SyncMap
	m.Store("a", 1)
	m.Store("b", 2)
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	AssertEqual(t, 2, count)
}

func TestSyncMap_Range_Bad(t *T) {
	// Bad: Range on empty map fires zero callbacks.
	var m SyncMap
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	AssertEqual(t, 0, count)
}

func TestSyncMap_Range_Ugly(t *T) {
	// Ugly: Range with early termination — function returns false.
	var m SyncMap
	for i := 0; i < 10; i++ {
		m.Store(i, i*10)
	}
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return false // stop after first
	})
	AssertEqual(t, 1, count)
}

func TestSyncMap_Clear_Good(t *T) {
	var m SyncMap
	m.Store("a", 1)
	m.Store("b", 2)
	m.Clear()
	_, ok := m.Load("a")
	AssertFalse(t, ok)
}

func TestSyncMap_Concurrent_Ugly(t *T) {
	// Ugly: 100 goroutines storing disjoint keys; Range must see all.
	var m SyncMap
	var wg WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			m.Store(k, k*10)
		}(i)
	}
	wg.Wait()
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	AssertEqual(t, 100, count)
}
