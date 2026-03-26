package core_test

import (

	. "dappco.re/go/core"
)

func ExampleCore_Lock() {
	c := New()
	lock := c.Lock("drain")
	lock.Mutex.Lock()
	Println("locked")
	lock.Mutex.Unlock()
	Println("unlocked")
	// Output:
	// locked
	// unlocked
}
