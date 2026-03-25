package core_test

import (

	. "dappco.re/go/core"
)

func ExampleContains() {
	Println(Contains("hello world", "world"))
	Println(Contains("hello world", "mars"))
	// Output:
	// true
	// false
}

func ExampleSplit() {
	parts := Split("deploy/to/homelab", "/")
	Println(parts)
	// Output: [deploy to homelab]
}

func ExampleJoin() {
	Println(Join("/", "deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

func ExampleConcat() {
	Println(Concat("hello", " ", "world"))
	// Output: hello world
}

func ExampleTrim() {
	Println(Trim("  spaced  "))
	// Output: spaced
}
