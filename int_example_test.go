package core_test

import . "dappco.re/go"

// ExampleAtoi parses decimal text through `Atoi` for CLI argument conversion. Numeric
// parsing and formatting use core helpers for CLI-friendly values.
func ExampleAtoi() {
	r := Atoi("42")
	Println(r.Value)
	// Output: 42
}

// ExampleItoa formats an integer through `Itoa` for CLI argument conversion. Numeric
// parsing and formatting use core helpers for CLI-friendly values.
func ExampleItoa() {
	Println(Itoa(42))
	// Output: 42
}

// ExampleFormatInt formats an integer with a base through `FormatInt` for CLI argument
// conversion. Numeric parsing and formatting use core helpers for CLI-friendly values.
func ExampleFormatInt() {
	Println(FormatInt(255, 16))
	// Output: ff
}

// ExampleParseInt parses an integer with a base through `ParseInt` for CLI argument
// conversion. Numeric parsing and formatting use core helpers for CLI-friendly values.
func ExampleParseInt() {
	r := ParseInt("ff", 16, 64)
	Println(r.Value)
	// Output: 255
}
