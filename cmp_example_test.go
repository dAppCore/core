package core_test

import . "dappco.re/go"

func ExampleCompare() {
	Println(Compare("alpha", "bravo"))
	Println(Compare(42, 42))
	Println(Compare(9, 3))
	// Output:
	// -1
	// 0
	// 1
}
