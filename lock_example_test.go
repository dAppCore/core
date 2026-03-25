package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleCore_Lock() {
	c := New()
	lock := c.Lock("drain")
	lock.Mutex.Lock()
	fmt.Println("locked")
	lock.Mutex.Unlock()
	fmt.Println("unlocked")
	// Output:
	// locked
	// unlocked
}
