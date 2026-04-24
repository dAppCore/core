package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestSyncMap_Store_Good(t *testing.T) {
	var m SyncMap
	m.Store("k", 42)
	v, ok := m.Load("k")
	assert.True(t, ok)
	assert.Equal(t, 42, v)
}

func TestSyncMap_Store_Bad(t *testing.T) {
	// Bad: Load on absent key returns nil, false.
	var m SyncMap
	v, ok := m.Load("missing")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestSyncMap_Store_Ugly(t *testing.T) {
	// Ugly: overwrite via Store, then via Swap, observe the chain.
	var m SyncMap
	m.Store("k", "first")
	m.Store("k", "second")
	prev, loaded := m.Swap("k", "third")
	assert.True(t, loaded)
	assert.Equal(t, "second", prev)
	v, _ := m.Load("k")
	assert.Equal(t, "third", v)
}

func TestSyncMap_LoadOrStore_Good(t *testing.T) {
	var m SyncMap
	actual, loaded := m.LoadOrStore("k", 1)
	assert.False(t, loaded)
	assert.Equal(t, 1, actual)
	actual, loaded = m.LoadOrStore("k", 2)
	assert.True(t, loaded)
	assert.Equal(t, 1, actual, "second LoadOrStore returns existing value")
}

func TestSyncMap_LoadAndDelete_Good(t *testing.T) {
	var m SyncMap
	m.Store("k", "v")
	v, loaded := m.LoadAndDelete("k")
	assert.True(t, loaded)
	assert.Equal(t, "v", v)
	_, ok := m.Load("k")
	assert.False(t, ok)
}

func TestSyncMap_CompareAndSwap_Good(t *testing.T) {
	var m SyncMap
	m.Store("k", 1)
	swapped := m.CompareAndSwap("k", 1, 2)
	assert.True(t, swapped)
	v, _ := m.Load("k")
	assert.Equal(t, 2, v)
}

func TestSyncMap_CompareAndSwap_Bad(t *testing.T) {
	// Bad: CompareAndSwap with mismatched old returns false, no change.
	var m SyncMap
	m.Store("k", 1)
	swapped := m.CompareAndSwap("k", 99, 2)
	assert.False(t, swapped)
	v, _ := m.Load("k")
	assert.Equal(t, 1, v)
}

func TestSyncMap_Range_Good(t *testing.T) {
	var m SyncMap
	m.Store("a", 1)
	m.Store("b", 2)
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	assert.Equal(t, 2, count)
}

func TestSyncMap_Range_Bad(t *testing.T) {
	// Bad: Range on empty map fires zero callbacks.
	var m SyncMap
	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count)
}

func TestSyncMap_Range_Ugly(t *testing.T) {
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
	assert.Equal(t, 1, count)
}

func TestSyncMap_Clear_Good(t *testing.T) {
	var m SyncMap
	m.Store("a", 1)
	m.Store("b", 2)
	m.Clear()
	_, ok := m.Load("a")
	assert.False(t, ok)
}

func TestSyncMap_Concurrent_Ugly(t *testing.T) {
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
	assert.Equal(t, 100, count)
}
