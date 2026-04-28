package core_test

import . "dappco.re/go"

// ExampleIsLetter checks a letter through `IsLetter` for input normalisation. Character
// checks and case conversion stay behind core unicode helpers.
func ExampleIsLetter() {
	Println(IsLetter('A'))
	Println(IsLetter('7'))
	// Output:
	// true
	// false
}

// ExampleIsDigit checks a digit through `IsDigit` for input normalisation. Character
// checks and case conversion stay behind core unicode helpers.
func ExampleIsDigit() {
	Println(IsDigit('7'))
	// Output: true
}

// ExampleIsSpace checks whitespace through `IsSpace` for input normalisation. Character
// checks and case conversion stay behind core unicode helpers.
func ExampleIsSpace() {
	Println(IsSpace('\n'))
	// Output: true
}

// ExampleIsUpper checks uppercase text through `IsUpper` for input normalisation.
// Character checks and case conversion stay behind core unicode helpers.
func ExampleIsUpper() {
	Println(IsUpper('A'))
	// Output: true
}

// ExampleIsLower checks lowercase text through `IsLower` for input normalisation.
// Character checks and case conversion stay behind core unicode helpers.
func ExampleIsLower() {
	Println(IsLower('a'))
	// Output: true
}

// ExampleToUpper converts text to upper case through `ToUpper` for input normalisation.
// Character checks and case conversion stay behind core unicode helpers.
func ExampleToUpper() {
	Println(string(ToUpper('a')))
	// Output: A
}

// ExampleToLower converts text to lower case through `ToLower` for input normalisation.
// Character checks and case conversion stay behind core unicode helpers.
func ExampleToLower() {
	Println(string(ToLower('A')))
	// Output: a
}
