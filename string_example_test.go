package core_test

import (
	. "dappco.re/go"
)

func ExampleHasPrefix() {
	Println(HasPrefix("--verbose", "--"))
	// Output: true
}

func ExampleHasSuffix() {
	Println(HasSuffix("report.pdf", ".pdf"))
	// Output: true
}

func ExampleTrimPrefix() {
	Println(TrimPrefix("--verbose", "--"))
	// Output: verbose
}

func ExampleTrimSuffix() {
	Println(TrimSuffix("report.pdf", ".pdf"))
	// Output: report
}

func ExampleContains() {
	Println(Contains("hello world", "world"))
	Println(Contains("hello world", "mars"))
	// Output:
	// true
	// false
}

func ExampleSplitN() {
	Println(SplitN("key=value=extra", "=", 2))
	// Output: [key value=extra]
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

func ExampleReplace() {
	Println(Replace("deploy/to/homelab", "/", "."))
	// Output: deploy.to.homelab
}

func ExampleLower() {
	Println(Lower("HELLO"))
	// Output: hello
}

func ExampleUpper() {
	Println(Upper("hello"))
	// Output: HELLO
}

func ExampleConcat() {
	Println(Concat("hello", " ", "world"))
	// Output: hello world
}

func ExampleTrim() {
	Println(Trim("  spaced  "))
	// Output: spaced
}

func ExampleRuneCount() {
	Println(RuneCount("hello"))
	Println(RuneCount("é"))
	// Output:
	// 5
	// 1
}

func ExampleNewBuilder() {
	b := NewBuilder()
	b.WriteString("hello")
	b.WriteString(" world")
	Println(b.String())
	// Output: hello world
}

func ExampleNewReader() {
	r := ReadAll(NewReader("hello"))
	Println(r.Value)
	// Output: hello
}

func ExampleSprint() {
	Println(Sprint("port=", 8080))
	// Output: port=8080
}

func ExampleSprintf() {
	Println(Sprintf("port=%d", 8080))
	// Output: port=8080
}
