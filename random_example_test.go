package core_test

import . "dappco.re/go"

func ExampleRandomBytes() {
	r := RandomBytes(8)
	if r.OK {
		Println(len(r.Value.([]byte)))
	}
	// Output: 8
}

func ExampleRandomString() {
	r := RandomString(8)
	if r.OK {
		Println(len(r.Value.(string)))
	}
	// Output: 16
}

func ExampleRandomInt() {
	r := RandomInt(10, 20)
	if r.OK {
		n := r.Value.(int)
		Println(n >= 10 && n < 20)
	}
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
