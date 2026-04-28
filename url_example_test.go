package core_test

import . "dappco.re/go"

// ExampleURLParse parses an endpoint through `URLParse` for remote endpoint handling.
// Endpoint parsing, escaping, and normalisation use core URL wrappers.
func ExampleURLParse() {
	r := URLParse("https://example.com/search?q=core")
	Println(r.Value)
	// Output: https://example.com/search?q=core
}

// ExampleURLEncode encodes query text through `URLEncode` for remote endpoint handling.
// Endpoint parsing, escaping, and normalisation use core URL wrappers.
func ExampleURLEncode() {
	Println(URLEncode("hello world"))
	// Output: hello+world
}

// ExampleURLDecode decodes query text through `URLDecode` for remote endpoint handling.
// Endpoint parsing, escaping, and normalisation use core URL wrappers.
func ExampleURLDecode() {
	r := URLDecode("hello+world")
	Println(r.Value)
	// Output: hello world
}

// ExampleURLPathEscape escapes a path segment through `URLPathEscape` for remote endpoint
// handling. Endpoint parsing, escaping, and normalisation use core URL wrappers.
func ExampleURLPathEscape() {
	Println(URLPathEscape("deploy/to homelab"))
	// Output: deploy%2Fto%20homelab
}

// ExampleURLNormalize normalises an endpoint through `URLNormalize` for remote endpoint
// handling. Endpoint parsing, escaping, and normalisation use core URL wrappers.
func ExampleURLNormalize() {
	Println(URLNormalize("https://example.com/a b?q=hello world"))
	// Output: https://example.com/a%20b?q=hello world
}
