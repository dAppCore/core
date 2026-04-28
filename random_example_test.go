package core_test

import . "dappco.re/go"

// ExampleRandomBytes generates random bytes through `RandomBytes` for session nonce
// generation. Nonces, strings, and bounded integers return Results suitable for session
// work.
func ExampleRandomBytes() {
	r := RandomBytes(8)
	if r.OK {
		Println(len(r.Value.([]byte)))
	}
	// Output: 8
}

// ExampleRandomString generates random text through `RandomString` for session nonce
// generation. Nonces, strings, and bounded integers return Results suitable for session
// work.
func ExampleRandomString() {
	r := RandomString(8)
	if r.OK {
		Println(len(r.Value.(string)))
	}
	// Output: 16
}

// ExampleRandomInt generates a bounded integer through `RandomInt` for session nonce
// generation. Nonces, strings, and bounded integers return Results suitable for session
// work.
func ExampleRandomInt() {
	r := RandomInt(10, 20)
	if r.OK {
		n := r.Value.(int)
		Println(n >= 10 && n < 20)
	}
	// Output: true
}

// ExampleRandPick picks a random element through `RandPick` for session nonce generation.
// Nonces, strings, and bounded integers return Results suitable for session work.
func ExampleRandPick() {
	choices := []string{"alpha", "bravo", "charlie"}
	Println(SliceContains(choices, RandPick(choices)))
	// Output: true
}

// ExampleRandIntn generates a bounded pseudo-random integer through `RandIntn` for session
// nonce generation. Nonces, strings, and bounded integers return Results suitable for
// session work.
func ExampleRandIntn() {
	n := RandIntn(5)
	Println(n >= 0 && n < 5)
	// Output: true
}
