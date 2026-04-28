package core_test

import . "dappco.re/go"

// ExampleNewBuffer creates an empty buffer through `NewBuffer` for in-memory payload
// assembly. Buffer creation stays on the core wrapper surface for later stream or encoding
// work.
func ExampleNewBuffer() {
	buf := NewBuffer([]byte("hello"))
	Println(buf.String())

	empty := NewBuffer()
	Println(empty.Len())
	// Output:
	// hello
	// 0
}

// ExampleNewBufferString creates a buffer from existing text through `NewBufferString` for
// in-memory payload assembly. Buffer creation stays on the core wrapper surface for later
// stream or encoding work.
func ExampleNewBufferString() {
	buf := NewBufferString("hello world")
	Println(buf.String())
	Println(buf.Len())
	// Output:
	// hello world
	// 11
}
