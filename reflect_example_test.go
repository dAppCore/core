package core_test

import (
	. "dappco.re/go"
)

func ExampleTypeOf() {
	t := TypeOf(42)
	Println(t.Kind())
	// Output: int
}

func ExampleValueOf() {
	v := ValueOf("hello")
	Println(v.String())
	// Output: hello
}

func ExampleDeepEqual() {
	Println(DeepEqual([]int{1, 2, 3}, []int{1, 2, 3}))
	// Output: true
}

func ExampleKind() {
	k := TypeOf("x").Kind()
	Println(k == KindString)
	// Output: true
}
