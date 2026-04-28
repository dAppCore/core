package core_test

import . "dappco.re/go"

func ExampleAtoi() {
	r := Atoi("42")
	Println(r.Value)
	// Output: 42
}

func ExampleItoa() {
	Println(Itoa(42))
	// Output: 42
}

func ExampleFormatInt() {
	Println(FormatInt(255, 16))
	// Output: ff
}

func ExampleParseInt() {
	r := ParseInt("ff", 16, 64)
	Println(r.Value)
	// Output: 255
}
