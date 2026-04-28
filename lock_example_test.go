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
