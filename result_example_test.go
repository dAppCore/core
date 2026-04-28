package core_test

import (
	. "dappco.re/go"
)

func ExampleResult_Must() {
	v := (Result{Value: 42, OK: true}).Must()
	Println(v)
	// Output: 42
}

func ExampleResult_Or() {
	v := (Result{OK: false}).Or("fallback")
	Println(v)
	// Output: fallback
}

func ExampleResult_Error() {
	r := Result{Value: NewError("bad config"), OK: false}
	Println(r.Error())
	// Output: bad config
}

func ExampleCast() {
	r := Result{Value: "hello", OK: true}
	s, ok := Cast[string](r)
	Println(s, ok)
	// Output: hello true
}

func ExampleTry() {
	r := Try(func() any {
		return 42
	})
	Println(r.OK, r.Value)
	// Output: true 42
}
