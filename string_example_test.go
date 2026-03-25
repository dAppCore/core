package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleContains() {
	fmt.Println(Contains("hello world", "world"))
	fmt.Println(Contains("hello world", "mars"))
	// Output:
	// true
	// false
}

func ExampleSplit() {
	parts := Split("deploy/to/homelab", "/")
	fmt.Println(parts)
	// Output: [deploy to homelab]
}

func ExampleJoin() {
	fmt.Println(Join("/", "deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

func ExampleConcat() {
	fmt.Println(Concat("hello", " ", "world"))
	// Output: hello world
}

func ExampleTrim() {
	fmt.Println(Trim("  spaced  "))
	// Output: spaced
}
