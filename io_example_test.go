package core_test

import (
	. "dappco.re/go"
)

// ExampleReader declares a reader through `Reader` for streaming payloads. Stream copying,
// EOF checks, and writes avoid direct io imports in consumers.
func ExampleReader() {
	var r Reader = NewReader("hello")
	data := ReadAll(r)
	Println(data.Value)
	// Output: hello
}

// ExampleWriter declares a writer through `Writer` for streaming payloads. Stream copying,
// EOF checks, and writes avoid direct io imports in consumers.
func ExampleWriter() {
	var w Writer = NewBuffer()
	n := WriteString(w, "hello")
	Println(n.Value)
	// Output: 5
}

// ExampleEOF checks the EOF sentinel through `EOF` for streaming payloads. Stream copying,
// EOF checks, and writes avoid direct io imports in consumers.
func ExampleEOF() {
	Println(EOF != nil)
	// Output: true
}

// ExampleCopy copies a stream through `Copy` for streaming payloads. Stream copying, EOF
// checks, and writes avoid direct io imports in consumers.
func ExampleCopy() {
	dst := NewBuffer()
	r := Copy(dst, NewReader("hello"))
	Println(r.Value)
	Println(dst.String())
	// Output:
	// 5
	// hello
}

// ExampleCopyN copies a bounded stream through `CopyN` for streaming payloads. Stream
// copying, EOF checks, and writes avoid direct io imports in consumers.
func ExampleCopyN() {
	dst := NewBuffer()
	r := CopyN(dst, NewReader("hello"), 2)
	Println(r.Value)
	Println(dst.String())
	// Output:
	// 2
	// he
}

// ExampleWriteString writes text into a stream through `WriteString` for streaming
// payloads. Stream copying, EOF checks, and writes avoid direct io imports in consumers.
func ExampleWriteString() {
	dst := NewBuffer()
	r := WriteString(dst, "hello")
	Println(r.Value)
	Println(dst.String())
	// Output:
	// 5
	// hello
}

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

// ExampleBuffer declares a Buffer-typed local through the `Buffer` alias
// for in-memory payload assembly. Buffer creation stays on the core
// wrapper surface for later stream or encoding work.
func ExampleBuffer() {
	var b Buffer
	b.WriteString("ready")
	Println(b.String())
	// Output: ready
}

// ExampleNewBufferReader creates a bytes.Reader over a byte slice through
// `NewBufferReader` for HTTP body construction. Buffer creation stays on
// the core wrapper surface for later stream or encoding work.
func ExampleNewBufferReader() {
	rd := NewBufferReader([]byte("hello"))
	out := make([]byte, 5)
	rd.Read(out)
	Println(string(out))
	// Output: hello
}
