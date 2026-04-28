package core_test

import . "dappco.re/go"

func ExampleURLParse() {
	r := URLParse("https://example.com/search?q=core")
	Println(r.Value)
	// Output: https://example.com/search?q=core
}

func ExampleURLEncode() {
	Println(URLEncode("hello world"))
	// Output: hello+world
}

func ExampleURLDecode() {
	r := URLDecode("hello+world")
	Println(r.Value)
	// Output: hello world
}

func ExampleURLPathEscape() {
	Println(URLPathEscape("deploy/to homelab"))
	// Output: deploy%2Fto%20homelab
}

func ExampleURLNormalize() {
	Println(URLNormalize("https://example.com/a b?q=hello world"))
	// Output: https://example.com/a%20b?q=hello world
}
