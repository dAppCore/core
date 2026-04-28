package core_test

import (
	. "dappco.re/go"
)

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

func ExampleLock_Lock() {
	c := New()
	lock := c.Lock("drain")
	lock.Lock()
	Println("locked")
	lock.Unlock()
	// Output: locked
}

func ExampleLock_Unlock() {
	c := New()
	lock := c.Lock("drain")
	lock.Lock()
	lock.Unlock()
	Println("unlocked")
	// Output: unlocked
}

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

func ExampleLock_RUnlock() {
	c := New()
	lock := c.Lock("cache")
	lock.RLock()
	lock.RUnlock()
	Println("read-unlocked")
	// Output: read-unlocked
}

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

func ExampleCore_LockEnable() {
	c := New()
	c.LockEnable()
	c.LockApply()
	Println(c.Service("late", Service{}).OK)
	// Output: false
}

func ExampleCore_LockApply() {
	c := New()
	c.LockEnable()
	c.LockApply()
	Println(c.Service("late", Service{}).OK)
	// Output: false
}

func ExampleCore_Startables() {
	c := New()
	c.Service("worker", Service{OnStart: func() Result { return Result{OK: true} }})
	r := c.Startables()
	Println(len(r.Value.([]*Service)))
	// Output: 1
}

func ExampleCore_Stoppables() {
	c := New()
	c.Service("worker", Service{OnStop: func() Result { return Result{OK: true} }})
	r := c.Stoppables()
	Println(len(r.Value.([]*Service)))
	// Output: 1
}
