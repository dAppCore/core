package core_test

import . "dappco.re/go"

func ExampleIsLetter() {
	Println(IsLetter('A'))
	Println(IsLetter('7'))
	// Output:
	// true
	// false
}

func ExampleIsDigit() {
	Println(IsDigit('7'))
	// Output: true
}

func ExampleIsSpace() {
	Println(IsSpace('\n'))
	// Output: true
}

func ExampleIsUpper() {
	Println(IsUpper('A'))
	// Output: true
}

func ExampleIsLower() {
	Println(IsLower('a'))
	// Output: true
}

func ExampleToUpper() {
	Println(string(ToUpper('a')))
	// Output: A
}

func ExampleToLower() {
	Println(string(ToLower('A')))
	// Output: a
}
