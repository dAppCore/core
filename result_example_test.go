package core_test

import (
	. "dappco.re/go"
)

// ExampleResult_Must unwraps a successful Result through `Result.Must` for Result-based
// control flow. Success, fallback, casting, and error inspection all use the same Result
// shape.
func ExampleResult_Must() {
	v := (Result{Value: 42, OK: true}).Must()
	Println(v)
	// Output: 42
}

// ExampleResult_Or falls back from a failed Result through `Result.Or` for Result-based
// control flow. Success, fallback, casting, and error inspection all use the same Result
// shape.
func ExampleResult_Or() {
	v := (Result{OK: false}).Or("fallback")
	Println(v)
	// Output: fallback
}

// ExampleResult_Error writes or renders an error through `Result.Error` for Result-based
// control flow. Success, fallback, casting, and error inspection all use the same Result
// shape.
func ExampleResult_Error() {
	r := Result{Value: NewError("bad config"), OK: false}
	Println(r.Error())
	// Output: bad config
}

// ExampleResult_Code reads a Result error code through `Result.Code` for Result-based
// control flow. Success, fallback, casting, and error inspection all use the same Result
// shape.
func ExampleResult_Code() {
	r := Result{Value: NewCode("fs.notfound", "missing file"), OK: false}
	Println(r.Code())
	// Output: fs.notfound
}

// ExampleCast casts a Result value through `Cast` for Result-based control flow. Success,
// fallback, casting, and error inspection all use the same Result shape.
func ExampleCast() {
	r := Result{Value: "hello", OK: true}
	s, ok := Cast[string](r)
	Println(s, ok)
	// Output: hello true
}

// ExampleTry converts panic-prone work into a Result through `Try` for Result-based
// control flow. Success, fallback, casting, and error inspection all use the same Result
// shape.
func ExampleTry() {
	r := Try(func() any {
		return 42
	})
	Println(r.OK, r.Value)
	// Output: true 42
}
