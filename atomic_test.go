package core_test

import (
	. "dappco.re/go"
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

func TestAtomic_AtomicBool_Load_Good(t *T) {
	var ready AtomicBool
	ready.Store(true)

	AssertTrue(t, ready.Load())
}

func TestAtomic_AtomicBool_Load_Bad(t *T) {
	var ready AtomicBool

	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicBool_Load_Ugly(t *T) {
	var ready AtomicBool
	ready.Store(true)
	ready.Store(false)

	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicBool_Store_Good(t *T) {
	var ready AtomicBool

	ready.Store(true)

	AssertTrue(t, ready.Load())
}

func TestAtomic_AtomicBool_Store_Bad(t *T) {
	var ready AtomicBool
	ready.Store(true)

	ready.Store(false)

	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicBool_Store_Ugly(t *T) {
	var ready AtomicBool

	ready.Store(false)

	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicBool_Swap_Good(t *T) {
	var ready AtomicBool

	previous := ready.Swap(true)

	AssertFalse(t, previous)
	AssertTrue(t, ready.Load())
}

func TestAtomic_AtomicBool_Swap_Bad(t *T) {
	var ready AtomicBool
	ready.Store(true)

	previous := ready.Swap(false)

	AssertTrue(t, previous)
	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicBool_Swap_Ugly(t *T) {
	var ready AtomicBool

	previous := ready.Swap(false)

	AssertFalse(t, previous)
	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicBool_CompareAndSwap_Good(t *T) {
	var ready AtomicBool

	swapped := ready.CompareAndSwap(false, true)

	AssertTrue(t, swapped)
	AssertTrue(t, ready.Load())
}

func TestAtomic_AtomicBool_CompareAndSwap_Bad(t *T) {
	var ready AtomicBool
	ready.Store(true)

	swapped := ready.CompareAndSwap(false, false)

	AssertFalse(t, swapped)
	AssertTrue(t, ready.Load())
}

func TestAtomic_AtomicBool_CompareAndSwap_Ugly(t *T) {
	var ready AtomicBool

	swapped := ready.CompareAndSwap(false, false)

	AssertTrue(t, swapped)
	AssertFalse(t, ready.Load())
}

func TestAtomic_AtomicInt32_Load_Good(t *T) {
	var depth AtomicInt32
	depth.Store(12)

	AssertEqual(t, int32(12), depth.Load())
}

func TestAtomic_AtomicInt32_Load_Bad(t *T) {
	var depth AtomicInt32

	AssertEqual(t, int32(0), depth.Load())
}

func TestAtomic_AtomicInt32_Load_Ugly(t *T) {
	var depth AtomicInt32
	depth.Store(-1)

	AssertEqual(t, int32(-1), depth.Load())
}

func TestAtomic_AtomicInt32_Store_Good(t *T) {
	var depth AtomicInt32

	depth.Store(4)

	AssertEqual(t, int32(4), depth.Load())
}

func TestAtomic_AtomicInt32_Store_Bad(t *T) {
	var depth AtomicInt32
	depth.Store(4)

	depth.Store(0)

	AssertEqual(t, int32(0), depth.Load())
}

func TestAtomic_AtomicInt32_Store_Ugly(t *T) {
	var depth AtomicInt32

	depth.Store(-2147483648)

	AssertEqual(t, int32(-2147483648), depth.Load())
}

func TestAtomic_AtomicInt32_Add_Good(t *T) {
	var depth AtomicInt32
	depth.Store(7)

	next := depth.Add(5)

	AssertEqual(t, int32(12), next)
	AssertEqual(t, int32(12), depth.Load())
}

func TestAtomic_AtomicInt32_Add_Bad(t *T) {
	var depth AtomicInt32
	depth.Store(7)

	next := depth.Add(-2)

	AssertEqual(t, int32(5), next)
	AssertEqual(t, int32(5), depth.Load())
}

func TestAtomic_AtomicInt32_Add_Ugly(t *T) {
	var depth AtomicInt32

	next := depth.Add(0)

	AssertEqual(t, int32(0), next)
	AssertEqual(t, int32(0), depth.Load())
}

func TestAtomic_AtomicInt32_Swap_Good(t *T) {
	var depth AtomicInt32
	depth.Store(3)

	previous := depth.Swap(8)

	AssertEqual(t, int32(3), previous)
	AssertEqual(t, int32(8), depth.Load())
}

func TestAtomic_AtomicInt32_Swap_Bad(t *T) {
	var depth AtomicInt32

	previous := depth.Swap(8)

	AssertEqual(t, int32(0), previous)
	AssertEqual(t, int32(8), depth.Load())
}

func TestAtomic_AtomicInt32_Swap_Ugly(t *T) {
	var depth AtomicInt32
	depth.Store(-4)

	previous := depth.Swap(-9)

	AssertEqual(t, int32(-4), previous)
	AssertEqual(t, int32(-9), depth.Load())
}

func TestAtomic_AtomicInt32_CompareAndSwap_Good(t *T) {
	var depth AtomicInt32
	depth.Store(2)

	swapped := depth.CompareAndSwap(2, 3)

	AssertTrue(t, swapped)
	AssertEqual(t, int32(3), depth.Load())
}

func TestAtomic_AtomicInt32_CompareAndSwap_Bad(t *T) {
	var depth AtomicInt32
	depth.Store(2)

	swapped := depth.CompareAndSwap(1, 3)

	AssertFalse(t, swapped)
	AssertEqual(t, int32(2), depth.Load())
}

func TestAtomic_AtomicInt32_CompareAndSwap_Ugly(t *T) {
	var depth AtomicInt32
	var wins AtomicInt32
	var wg WaitGroup

	for i := 0; i < 32; i++ {
		wg.Go(func() {
			if depth.CompareAndSwap(0, 1) {
				wins.Add(1)
			}
		})
	}
	wg.Wait()

	AssertEqual(t, int32(1), wins.Load())
	AssertEqual(t, int32(1), depth.Load())
}

func TestAtomic_AtomicInt64_Load_Good(t *T) {
	var count AtomicInt64
	count.Store(9000000000)

	AssertEqual(t, int64(9000000000), count.Load())
}

func TestAtomic_AtomicInt64_Load_Bad(t *T) {
	var count AtomicInt64

	AssertEqual(t, int64(0), count.Load())
}

func TestAtomic_AtomicInt64_Load_Ugly(t *T) {
	var count AtomicInt64
	count.Store(-9000000000)

	AssertEqual(t, int64(-9000000000), count.Load())
}

func TestAtomic_AtomicInt64_Store_Good(t *T) {
	var count AtomicInt64

	count.Store(42)

	AssertEqual(t, int64(42), count.Load())
}

func TestAtomic_AtomicInt64_Store_Bad(t *T) {
	var count AtomicInt64
	count.Store(42)

	count.Store(0)

	AssertEqual(t, int64(0), count.Load())
}

func TestAtomic_AtomicInt64_Store_Ugly(t *T) {
	var count AtomicInt64

	count.Store(-9223372036854775808)

	AssertEqual(t, int64(-9223372036854775808), count.Load())
}

func TestAtomic_AtomicInt64_Add_Good(t *T) {
	var count AtomicInt64
	count.Store(40)

	next := count.Add(2)

	AssertEqual(t, int64(42), next)
	AssertEqual(t, int64(42), count.Load())
}

func TestAtomic_AtomicInt64_Add_Bad(t *T) {
	var count AtomicInt64
	count.Store(40)

	next := count.Add(-10)

	AssertEqual(t, int64(30), next)
	AssertEqual(t, int64(30), count.Load())
}

func TestAtomic_AtomicInt64_Add_Ugly(t *T) {
	var count AtomicInt64

	next := count.Add(0)

	AssertEqual(t, int64(0), next)
	AssertEqual(t, int64(0), count.Load())
}

func TestAtomic_AtomicInt64_Swap_Good(t *T) {
	var count AtomicInt64
	count.Store(10)

	previous := count.Swap(20)

	AssertEqual(t, int64(10), previous)
	AssertEqual(t, int64(20), count.Load())
}

func TestAtomic_AtomicInt64_Swap_Bad(t *T) {
	var count AtomicInt64

	previous := count.Swap(20)

	AssertEqual(t, int64(0), previous)
	AssertEqual(t, int64(20), count.Load())
}

func TestAtomic_AtomicInt64_Swap_Ugly(t *T) {
	var count AtomicInt64
	count.Store(-10)

	previous := count.Swap(-20)

	AssertEqual(t, int64(-10), previous)
	AssertEqual(t, int64(-20), count.Load())
}

func TestAtomic_AtomicInt64_CompareAndSwap_Good(t *T) {
	var count AtomicInt64
	count.Store(2)

	swapped := count.CompareAndSwap(2, 3)

	AssertTrue(t, swapped)
	AssertEqual(t, int64(3), count.Load())
}

func TestAtomic_AtomicInt64_CompareAndSwap_Bad(t *T) {
	var count AtomicInt64
	count.Store(2)

	swapped := count.CompareAndSwap(1, 3)

	AssertFalse(t, swapped)
	AssertEqual(t, int64(2), count.Load())
}

func TestAtomic_AtomicInt64_CompareAndSwap_Ugly(t *T) {
	var count AtomicInt64
	var wins AtomicInt32
	var wg WaitGroup

	for i := 0; i < 32; i++ {
		wg.Go(func() {
			if count.CompareAndSwap(0, 1) {
				wins.Add(1)
			}
		})
	}
	wg.Wait()

	AssertEqual(t, int32(1), wins.Load())
	AssertEqual(t, int64(1), count.Load())
}

func TestAtomic_AtomicUint32_Load_Good(t *T) {
	var workers AtomicUint32
	workers.Store(4)

	AssertEqual(t, uint32(4), workers.Load())
}

func TestAtomic_AtomicUint32_Load_Bad(t *T) {
	var workers AtomicUint32

	AssertEqual(t, uint32(0), workers.Load())
}

func TestAtomic_AtomicUint32_Load_Ugly(t *T) {
	var workers AtomicUint32
	workers.Store(^uint32(0))

	AssertEqual(t, ^uint32(0), workers.Load())
}

func TestAtomic_AtomicUint32_Store_Good(t *T) {
	var workers AtomicUint32

	workers.Store(6)

	AssertEqual(t, uint32(6), workers.Load())
}

func TestAtomic_AtomicUint32_Store_Bad(t *T) {
	var workers AtomicUint32
	workers.Store(6)

	workers.Store(0)

	AssertEqual(t, uint32(0), workers.Load())
}

func TestAtomic_AtomicUint32_Store_Ugly(t *T) {
	var workers AtomicUint32

	workers.Store(^uint32(0))

	AssertEqual(t, ^uint32(0), workers.Load())
}

func TestAtomic_AtomicUint32_Add_Good(t *T) {
	var workers AtomicUint32
	workers.Store(6)

	next := workers.Add(2)

	AssertEqual(t, uint32(8), next)
	AssertEqual(t, uint32(8), workers.Load())
}

func TestAtomic_AtomicUint32_Add_Bad(t *T) {
	var workers AtomicUint32
	workers.Store(6)

	next := workers.Add(0)

	AssertEqual(t, uint32(6), next)
	AssertEqual(t, uint32(6), workers.Load())
}

func TestAtomic_AtomicUint32_Add_Ugly(t *T) {
	var workers AtomicUint32
	workers.Store(^uint32(0))

	next := workers.Add(1)

	AssertEqual(t, uint32(0), next)
	AssertEqual(t, uint32(0), workers.Load())
}

func TestAtomic_AtomicUint32_Swap_Good(t *T) {
	var workers AtomicUint32
	workers.Store(2)

	previous := workers.Swap(3)

	AssertEqual(t, uint32(2), previous)
	AssertEqual(t, uint32(3), workers.Load())
}

func TestAtomic_AtomicUint32_Swap_Bad(t *T) {
	var workers AtomicUint32

	previous := workers.Swap(3)

	AssertEqual(t, uint32(0), previous)
	AssertEqual(t, uint32(3), workers.Load())
}

func TestAtomic_AtomicUint32_Swap_Ugly(t *T) {
	var workers AtomicUint32
	workers.Store(^uint32(0))

	previous := workers.Swap(0)

	AssertEqual(t, ^uint32(0), previous)
	AssertEqual(t, uint32(0), workers.Load())
}

func TestAtomic_AtomicUint32_CompareAndSwap_Good(t *T) {
	var workers AtomicUint32
	workers.Store(2)

	swapped := workers.CompareAndSwap(2, 3)

	AssertTrue(t, swapped)
	AssertEqual(t, uint32(3), workers.Load())
}

func TestAtomic_AtomicUint32_CompareAndSwap_Bad(t *T) {
	var workers AtomicUint32
	workers.Store(2)

	swapped := workers.CompareAndSwap(1, 3)

	AssertFalse(t, swapped)
	AssertEqual(t, uint32(2), workers.Load())
}

func TestAtomic_AtomicUint32_CompareAndSwap_Ugly(t *T) {
	var workers AtomicUint32

	swapped := workers.CompareAndSwap(0, 0)

	AssertTrue(t, swapped)
	AssertEqual(t, uint32(0), workers.Load())
}

func TestAtomic_AtomicUint64_Load_Good(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(4096)

	AssertEqual(t, uint64(4096), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Load_Bad(t *T) {
	var bytesSeen AtomicUint64

	AssertEqual(t, uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Load_Ugly(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(^uint64(0))

	AssertEqual(t, ^uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Store_Good(t *T) {
	var bytesSeen AtomicUint64

	bytesSeen.Store(8192)

	AssertEqual(t, uint64(8192), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Store_Bad(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(8192)

	bytesSeen.Store(0)

	AssertEqual(t, uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Store_Ugly(t *T) {
	var bytesSeen AtomicUint64

	bytesSeen.Store(^uint64(0))

	AssertEqual(t, ^uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Add_Good(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(4096)

	next := bytesSeen.Add(4096)

	AssertEqual(t, uint64(8192), next)
	AssertEqual(t, uint64(8192), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Add_Bad(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(4096)

	next := bytesSeen.Add(0)

	AssertEqual(t, uint64(4096), next)
	AssertEqual(t, uint64(4096), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Add_Ugly(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(^uint64(0))

	next := bytesSeen.Add(1)

	AssertEqual(t, uint64(0), next)
	AssertEqual(t, uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Swap_Good(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(4096)

	previous := bytesSeen.Swap(8192)

	AssertEqual(t, uint64(4096), previous)
	AssertEqual(t, uint64(8192), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Swap_Bad(t *T) {
	var bytesSeen AtomicUint64

	previous := bytesSeen.Swap(8192)

	AssertEqual(t, uint64(0), previous)
	AssertEqual(t, uint64(8192), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_Swap_Ugly(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(^uint64(0))

	previous := bytesSeen.Swap(0)

	AssertEqual(t, ^uint64(0), previous)
	AssertEqual(t, uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_CompareAndSwap_Good(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(4096)

	swapped := bytesSeen.CompareAndSwap(4096, 8192)

	AssertTrue(t, swapped)
	AssertEqual(t, uint64(8192), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_CompareAndSwap_Bad(t *T) {
	var bytesSeen AtomicUint64
	bytesSeen.Store(4096)

	swapped := bytesSeen.CompareAndSwap(1024, 8192)

	AssertFalse(t, swapped)
	AssertEqual(t, uint64(4096), bytesSeen.Load())
}

func TestAtomic_AtomicUint64_CompareAndSwap_Ugly(t *T) {
	var bytesSeen AtomicUint64

	swapped := bytesSeen.CompareAndSwap(0, 0)

	AssertTrue(t, swapped)
	AssertEqual(t, uint64(0), bytesSeen.Load())
}

func TestAtomic_AtomicPointer_Load_Good(t *T) {
	var current AtomicPointer[pointerVal]
	value := &pointerVal{n: 7}
	current.Store(value)

	AssertSame(t, value, current.Load())
}

func TestAtomic_AtomicPointer_Load_Bad(t *T) {
	var current AtomicPointer[pointerVal]

	AssertNil(t, current.Load())
}

func TestAtomic_AtomicPointer_Load_Ugly(t *T) {
	var current AtomicPointer[pointerVal]
	current.Store(nil)

	AssertNil(t, current.Load())
}

func TestAtomic_AtomicPointer_Store_Good(t *T) {
	var current AtomicPointer[pointerVal]
	value := &pointerVal{n: 11}

	current.Store(value)

	AssertSame(t, value, current.Load())
}

func TestAtomic_AtomicPointer_Store_Bad(t *T) {
	var current AtomicPointer[pointerVal]
	value := &pointerVal{n: 11}
	current.Store(value)

	current.Store(nil)

	AssertNil(t, current.Load())
}

func TestAtomic_AtomicPointer_Store_Ugly(t *T) {
	var current AtomicPointer[pointerVal]
	empty := &pointerVal{}

	current.Store(empty)

	AssertEqual(t, 0, current.Load().n)
}

func TestAtomic_AtomicPointer_Swap_Good(t *T) {
	var current AtomicPointer[pointerVal]
	old := &pointerVal{n: 1}
	next := &pointerVal{n: 2}
	current.Store(old)

	previous := current.Swap(next)

	AssertSame(t, old, previous)
	AssertSame(t, next, current.Load())
}

func TestAtomic_AtomicPointer_Swap_Bad(t *T) {
	var current AtomicPointer[pointerVal]
	next := &pointerVal{n: 2}

	previous := current.Swap(next)

	AssertNil(t, previous)
	AssertSame(t, next, current.Load())
}

func TestAtomic_AtomicPointer_Swap_Ugly(t *T) {
	var current AtomicPointer[pointerVal]
	old := &pointerVal{n: 1}
	current.Store(old)

	previous := current.Swap(nil)

	AssertSame(t, old, previous)
	AssertNil(t, current.Load())
}

func TestAtomic_AtomicPointer_CompareAndSwap_Good(t *T) {
	var current AtomicPointer[pointerVal]
	old := &pointerVal{n: 1}
	next := &pointerVal{n: 2}
	current.Store(old)

	swapped := current.CompareAndSwap(old, next)

	AssertTrue(t, swapped)
	AssertSame(t, next, current.Load())
}

func TestAtomic_AtomicPointer_CompareAndSwap_Bad(t *T) {
	var current AtomicPointer[pointerVal]
	old := &pointerVal{n: 1}
	wrong := &pointerVal{n: 9}
	next := &pointerVal{n: 2}
	current.Store(old)

	swapped := current.CompareAndSwap(wrong, next)

	AssertFalse(t, swapped)
	AssertSame(t, old, current.Load())
}

func TestAtomic_AtomicPointer_CompareAndSwap_Ugly(t *T) {
	var current AtomicPointer[pointerVal]
	next := &pointerVal{n: 2}

	swapped := current.CompareAndSwap(nil, next)

	AssertTrue(t, swapped)
	AssertSame(t, next, current.Load())
}
