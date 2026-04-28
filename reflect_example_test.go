package core_test

import (
	. "dappco.re/go"
)

// ExampleTypeOf reads a type through `TypeOf` for runtime inspection. Reflection stays
// behind a narrow core surface for rare inspection code.
func ExampleTypeOf() {
	t := TypeOf(42)
	Println(t.Kind())
	// Output: int
}

// ExampleValueOf reads a value wrapper through `ValueOf` for runtime inspection.
// Reflection stays behind a narrow core surface for rare inspection code.
func ExampleValueOf() {
	v := ValueOf("hello")
	Println(v.String())
	// Output: hello
}

// ExampleDeepEqual compares nested values through `DeepEqual` for runtime inspection.
// Reflection stays behind a narrow core surface for rare inspection code.
func ExampleDeepEqual() {
	Println(DeepEqual([]int{1, 2, 3}, []int{1, 2, 3}))
	// Output: true
}

// ExampleKind reads a reflection kind through `Kind` for runtime inspection. Reflection
// stays behind a narrow core surface for rare inspection code.
func ExampleKind() {
	k := TypeOf("x").Kind()
	Println(k == KindString)
	// Output: true
}
