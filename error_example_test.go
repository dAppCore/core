package core_test

import (
	"errors"
	"fmt"

	. "dappco.re/go/core"
)

func ExampleE() {
	err := E("cache.Get", "key not found", nil)
	fmt.Println(Operation(err))
	fmt.Println(ErrorMessage(err))
	// Output:
	// cache.Get
	// key not found
}

func ExampleWrap() {
	cause := errors.New("connection refused")
	err := Wrap(cause, "database.Connect", "failed to reach host")
	fmt.Println(Operation(err))
	fmt.Println(errors.Is(err, cause))
	// Output:
	// database.Connect
	// true
}

func ExampleRoot() {
	cause := errors.New("original")
	wrapped := Wrap(cause, "op1", "first wrap")
	double := Wrap(wrapped, "op2", "second wrap")
	fmt.Println(Root(double))
	// Output: original
}
