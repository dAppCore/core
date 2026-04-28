package core_test

import . "dappco.re/go"

func ExampleReader() {
	var r Reader = NewReader("hello")
	data := ReadAll(r)
	Println(data.Value)
	// Output: hello
}

func ExampleWriter() {
	var w Writer = NewBuffer()
	n := WriteString(w, "hello")
	Println(n.Value)
	// Output: 5
}

func ExampleEOF() {
	Println(EOF != nil)
	// Output: true
}

func ExampleCopy() {
	dst := NewBuffer()
	r := Copy(dst, NewReader("hello"))
	Println(r.Value)
	Println(dst.String())
	// Output:
	// 5
	// hello
}

func ExampleCopyN() {
	dst := NewBuffer()
	r := CopyN(dst, NewReader("hello"), 2)
	Println(r.Value)
	Println(dst.String())
	// Output:
	// 2
	// he
}

func ExampleWriteString() {
	dst := NewBuffer()
	r := WriteString(dst, "hello")
	Println(r.Value)
	Println(dst.String())
	// Output:
	// 5
	// hello
}
