package core_test

import . "dappco.re/go"

func ExampleNewBuffer() {
	buf := NewBuffer([]byte("hello"))
	Println(buf.String())

	empty := NewBuffer()
	Println(empty.Len())
	// Output:
	// hello
	// 0
}

func ExampleNewBufferString() {
	buf := NewBufferString("hello world")
	Println(buf.String())
	Println(buf.Len())
	// Output:
	// hello world
	// 11
}
