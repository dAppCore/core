package core_test

import . "dappco.re/go"

// ExampleCompare orders two values through `Compare` for priority comparison. Ordering
// decisions use the core comparison wrapper rather than importing cmp directly.
func ExampleCompare() {
	Println(Compare("alpha", "bravo"))
	Println(Compare(42, 42))
	Println(Compare(9, 3))
	// Output:
	// -1
	// 0
	// 1
}
