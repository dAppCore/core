package core_test

import . "dappco.re/go"

func ExampleRandomBytes() {
	token := RandomBytes(8)
	Println(len(token))
	// Output: 8
}

func ExampleRandomString() {
	token := RandomString(8)
	Println(len(token))
	// Output: 16
}

func ExampleRandomInt() {
	n := RandomInt(10, 20)
	Println(n >= 10 && n < 20)
	// Output: true
}

func ExampleRandPick() {
	choices := []string{"alpha", "bravo", "charlie"}
	Println(SliceContains(choices, RandPick(choices)))
	// Output: true
}

func ExampleRandIntn() {
	n := RandIntn(5)
	Println(n >= 0 && n < 5)
	// Output: true
}
