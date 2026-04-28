package core_test

import . "dappco.re/go"

// ExampleMin selects the lower value through `Min` for health-check thresholds. Numeric
// helpers keep thresholds readable without importing math directly.
func ExampleMin() {
	Println(Min(3, 7))
	// Output: 3
}

// ExampleMax selects the higher value through `Max` for health-check thresholds. Numeric
// helpers keep thresholds readable without importing math directly.
func ExampleMax() {
	Println(Max(3, 7))
	// Output: 7
}

// ExampleAbs normalises a signed value through `Abs` for health-check thresholds. Numeric
// helpers keep thresholds readable without importing math directly.
func ExampleAbs() {
	Println(Abs(-42))
	// Output: 42
}

// ExamplePow raises a value to a power through `Pow` for health-check thresholds. Numeric
// helpers keep thresholds readable without importing math directly.
func ExamplePow() {
	Println(Pow(2, 3))
	// Output: 8
}

// ExampleFloor rounds a value down through `Floor` for health-check thresholds. Numeric
// helpers keep thresholds readable without importing math directly.
func ExampleFloor() {
	Println(Floor(3.7))
	// Output: 3
}

// ExampleCeil rounds a value up through `Ceil` for health-check thresholds. Numeric
// helpers keep thresholds readable without importing math directly.
func ExampleCeil() {
	Println(Ceil(3.1))
	// Output: 4
}

// ExampleRound rounds a value to the nearest integer through `Round` for health-check
// thresholds. Numeric helpers keep thresholds readable without importing math directly.
func ExampleRound() {
	Println(Round(3.5))
	// Output: 4
}
