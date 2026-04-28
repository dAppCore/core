package core_test

import . "dappco.re/go"

func ExampleMin() {
	Println(Min(3, 7))
	// Output: 3
}

func ExampleMax() {
	Println(Max(3, 7))
	// Output: 7
}

func ExampleAbs() {
	Println(Abs(-42))
	// Output: 42
}

func ExamplePow() {
	Println(Pow(2, 3))
	// Output: 8
}

func ExampleFloor() {
	Println(Floor(3.7))
	// Output: 3
}

func ExampleCeil() {
	Println(Ceil(3.1))
	// Output: 4
}

func ExampleRound() {
	Println(Round(3.5))
	// Output: 4
}
