package core_test

import . "dappco.re/go"

func ExampleHexEncode() {
	Println(HexEncode([]byte("hello")))
	// Output: 68656c6c6f
}

func ExampleHexDecode() {
	r := HexDecode("68656c6c6f")
	Println(string(r.Value.([]byte)))
	// Output: hello
}

func ExampleBase64Encode() {
	Println(Base64Encode([]byte("hello")))
	// Output: aGVsbG8=
}

func ExampleBase64Decode() {
	r := Base64Decode("aGVsbG8=")
	Println(string(r.Value.([]byte)))
	// Output: hello
}

func ExampleBase64URLEncode() {
	Println(Base64URLEncode([]byte("hello?")))
	// Output: aGVsbG8_
}

func ExampleBase64URLDecode() {
	r := Base64URLDecode("aGVsbG8_")
	Println(string(r.Value.([]byte)))
	// Output: hello?
}
