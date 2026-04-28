package core_test

import (
	. "dappco.re/go/core"
)

// --- AtomicBool ---

func TestAtomic_Bool_Good(t *T) {
	var a AtomicBool
	AssertFalse(t, a.Load())
	a.Store(true)
	AssertTrue(t, a.Load())
}

func TestAtomic_Bool_Bad(t *T) {
	// Bad: CompareAndSwap with wrong old returns false, no change.
	var a AtomicBool
	a.Store(true)
	swapped := a.CompareAndSwap(false, false)
	AssertFalse(t, swapped)
	AssertTrue(t, a.Load(), "CAS with wrong old must not mutate")
}

func TestAtomic_Bool_Ugly(t *T) {
	// Ugly: 100 goroutines racing CompareAndSwap to claim a one-shot flag.
	// Exactly one must win.
	var a AtomicBool
	var wins AtomicInt32
	var wg WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if a.CompareAndSwap(false, true) {
				wins.Add(1)
			}
		}()
	}
	wg.Wait()
	AssertEqual(t, int32(1), wins.Load(),
		"exactly one goroutine must win the CAS race")
}

// --- AtomicInt32 ---

func TestAtomic_Int32_Good(t *T) {
	var a AtomicInt32
	a.Store(5)
	AssertEqual(t, int32(5), a.Load())
	got := a.Add(3)
	AssertEqual(t, int32(8), got)
}

func TestAtomic_Int32_Bad(t *T) {
	// Bad: Swap returns previous value, not new.
	var a AtomicInt32
	a.Store(10)
	prev := a.Swap(20)
	AssertEqual(t, int32(10), prev)
	AssertEqual(t, int32(20), a.Load())
}

func TestAtomic_Int32_Ugly(t *T) {
	// Ugly: 1000 concurrent Adds. Final value must be exact (race-free).
	var a AtomicInt32
	var wg WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Add(1)
		}()
	}
	wg.Wait()
	AssertEqual(t, int32(1000), a.Load())
}

// --- AtomicInt64 ---

func TestAtomic_Int64_Good(t *T) {
	var a AtomicInt64
	a.Store(1 << 40)
	AssertEqual(t, int64(1<<40), a.Load())
}

func TestAtomic_Int64_Bad(t *T) {
	var a AtomicInt64
	a.Store(100)
	swapped := a.CompareAndSwap(99, 200)
	AssertFalse(t, swapped)
	AssertEqual(t, int64(100), a.Load())
}

func TestAtomic_Int64_Ugly(t *T) {
	var a AtomicInt64
	var wg WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Add(1)
		}()
	}
	wg.Wait()
	AssertEqual(t, int64(1000), a.Load())
}

// --- AtomicUint32 ---

func TestAtomic_Uint32_Good(t *T) {
	var a AtomicUint32
	a.Store(7)
	AssertEqual(t, uint32(7), a.Load())
	a.Add(3)
	AssertEqual(t, uint32(10), a.Load())
}

func TestAtomic_Uint32_Bad(t *T) {
	var a AtomicUint32
	a.Store(5)
	swapped := a.CompareAndSwap(99, 10)
	AssertFalse(t, swapped)
	AssertEqual(t, uint32(5), a.Load())
}

func TestAtomic_Uint32_Ugly(t *T) {
	var a AtomicUint32
	var wg WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Add(2)
		}()
	}
	wg.Wait()
	AssertEqual(t, uint32(1000), a.Load())
}

// --- AtomicUint64 ---

func TestAtomic_Uint64_Good(t *T) {
	var a AtomicUint64
	a.Store(1 << 50)
	AssertEqual(t, uint64(1<<50), a.Load())
}

func TestAtomic_Uint64_Bad(t *T) {
	var a AtomicUint64
	a.Store(100)
	prev := a.Swap(200)
	AssertEqual(t, uint64(100), prev)
}

func TestAtomic_Uint64_Ugly(t *T) {
	var a AtomicUint64
	var wg WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Add(1)
		}()
	}
	wg.Wait()
	AssertEqual(t, uint64(1000), a.Load())
}

// --- AtomicPointer ---

type pointerVal struct {
	n int
}

func TestAtomic_Pointer_Good(t *T) {
	var a AtomicPointer[pointerVal]
	AssertNil(t, a.Load())
	v := &pointerVal{n: 42}
	a.Store(v)
	AssertEqual(t, 42, a.Load().n)
}

func TestAtomic_Pointer_Bad(t *T) {
	// Bad: Swap returns nil if no prior value.
	var a AtomicPointer[pointerVal]
	prev := a.Swap(&pointerVal{n: 1})
	AssertNil(t, prev)
}

func TestAtomic_Pointer_Ugly(t *T) {
	// Ugly: 100 goroutines racing Store; Load at the end returns one of them.
	var a AtomicPointer[pointerVal]
	var wg WaitGroup
	pvs := make([]*pointerVal, 100)
	for i := 0; i < 100; i++ {
		pvs[i] = &pointerVal{n: i}
		wg.Add(1)
		go func(pv *pointerVal) {
			defer wg.Done()
			a.Store(pv)
		}(pvs[i])
	}
	wg.Wait()
	final := a.Load()
	AssertNotNil(t, final, "after 100 stores, Load must return non-nil")
	AssertGreaterOrEqual(t, final.n, 0)
	AssertLess(t, final.n, 100)
}

func TestAtomic_Pointer_CompareAndSwap_Good(t *T) {
	var a AtomicPointer[pointerVal]
	old := &pointerVal{n: 1}
	new := &pointerVal{n: 2}
	a.Store(old)
	swapped := a.CompareAndSwap(old, new)
	AssertTrue(t, swapped)
	AssertEqual(t, 2, a.Load().n)
}
