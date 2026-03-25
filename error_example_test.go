package core_test

import (

	. "dappco.re/go/core"
)

func ExampleE() {
	err := E("cache.Get", "key not found", nil)
	Println(Operation(err))
	Println(ErrorMessage(err))
	// Output:
	// cache.Get
	// key not found
}

func ExampleWrap() {
	cause := NewError("connection refused")
	err := Wrap(cause, "database.Connect", "failed to reach host")
	Println(Operation(err))
	Println(Is(err, cause))
	// Output:
	// database.Connect
	// true
}

func ExampleRoot() {
	cause := NewError("original")
	wrapped := Wrap(cause, "op1", "first wrap")
	double := Wrap(wrapped, "op2", "second wrap")
	Println(Root(double))
	// Output: original
}
