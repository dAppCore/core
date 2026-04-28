package core_test

import (
	. "dappco.re/go"
)

// ExampleCore_Lock acquires a lock through `Core.Lock` for lifecycle locking. Lifecycle
// locks coordinate startup and shutdown without exposing sync primitives directly.
func ExampleCore_Lock() {
	c := New()
	lock := c.Lock("drain")
	lock.Lock()
	Println("locked")
	lock.Unlock()
	Println("unlocked")
	// Output:
	// locked
	// unlocked
}

// ExampleLock_Lock acquires a lock through `Lock.Lock` for lifecycle locking. Lifecycle
// locks coordinate startup and shutdown without exposing sync primitives directly.
func ExampleLock_Lock() {
	c := New()
	lock := c.Lock("drain")
	lock.Lock()
	Println("locked")
	lock.Unlock()
	// Output: locked
}

// ExampleLock_Unlock releases a lock through `Lock.Unlock` for lifecycle locking.
// Lifecycle locks coordinate startup and shutdown without exposing sync primitives
// directly.
func ExampleLock_Unlock() {
	c := New()
	lock := c.Lock("drain")
	lock.Lock()
	lock.Unlock()
	Println("unlocked")
	// Output: unlocked
}

// ExampleLock_RLock acquires a read lock through `Lock.RLock` for lifecycle locking.
// Lifecycle locks coordinate startup and shutdown without exposing sync primitives
// directly.
func ExampleLock_RLock() {
	c := New()
	lock := c.Lock("cache")
	lock.RLock()
	Println("read-locked")
	lock.RUnlock()
	Println("read-unlocked")
	// Output:
	// read-locked
	// read-unlocked
}

// ExampleLock_RUnlock releases a read lock through `Lock.RUnlock` for lifecycle locking.
// Lifecycle locks coordinate startup and shutdown without exposing sync primitives
// directly.
func ExampleLock_RUnlock() {
	c := New()
	lock := c.Lock("cache")
	lock.RLock()
	lock.RUnlock()
	Println("read-unlocked")
	// Output: read-unlocked
}

// ExampleLock_TryLock attempts a non-blocking write lock through `Lock.TryLock` for
// lifecycle locking. Lifecycle locks coordinate startup and shutdown without exposing sync
// primitives directly.
func ExampleLock_TryLock() {
	c := New()
	lock := c.Lock("drain")
	if lock.TryLock().OK {
		Println("acquired")
		lock.Unlock()
	}
	// Output:
	// acquired
}

// ExampleCore_LockEnable enables lifecycle locking through `Core.LockEnable` for lifecycle
// locking. Lifecycle locks coordinate startup and shutdown without exposing sync
// primitives directly.
func ExampleCore_LockEnable() {
	c := New()
	c.LockEnable()
	c.LockApply()
	Println(c.Service("late", Service{}).OK)
	// Output: false
}

// ExampleCore_LockApply applies lifecycle locking through `Core.LockApply` for lifecycle
// locking. Lifecycle locks coordinate startup and shutdown without exposing sync
// primitives directly.
func ExampleCore_LockApply() {
	c := New()
	c.LockEnable()
	c.LockApply()
	Println(c.Service("late", Service{}).OK)
	// Output: false
}

// ExampleCore_Startables lists startup hooks through `Core.Startables` for lifecycle
// locking. Lifecycle locks coordinate startup and shutdown without exposing sync
// primitives directly.
func ExampleCore_Startables() {
	c := New()
	c.Service("worker", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Startables()
	Println(len(r.Value.([]*Service)))
	// Output: 1
}

// ExampleCore_Stoppables lists shutdown hooks through `Core.Stoppables` for lifecycle
// locking. Lifecycle locks coordinate startup and shutdown without exposing sync
// primitives directly.
func ExampleCore_Stoppables() {
	c := New()
	c.Service("worker", Service{OnStop: func() Result { return Result{OK: true} }})
	r := c.Stoppables()
	Println(len(r.Value.([]*Service)))
	// Output: 1
}
