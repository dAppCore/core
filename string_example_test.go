package core_test

import (
	. "dappco.re/go"
)

// ExampleHasPrefix checks a text prefix through `HasPrefix` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleHasPrefix() {
	Println(HasPrefix("--verbose", "--"))
	// Output: true
}

// ExampleHasSuffix checks a text suffix through `HasSuffix` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleHasSuffix() {
	Println(HasSuffix("report.pdf", ".pdf"))
	// Output: true
}

// ExampleTrimPrefix removes a text prefix through `TrimPrefix` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleTrimPrefix() {
	Println(TrimPrefix("--verbose", "--"))
	// Output: verbose
}

// ExampleTrimSuffix removes a text suffix through `TrimSuffix` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleTrimSuffix() {
	Println(TrimSuffix("report.pdf", ".pdf"))
	// Output: report
}

// ExampleContains checks text membership through `Contains` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleContains() {
	Println(Contains("hello world", "world"))
	Println(Contains("hello world", "mars"))
	// Output:
	// true
	// false
}

// ExampleSplitN splits text with a limit through `SplitN` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleSplitN() {
	Println(SplitN("key=value=extra", "=", 2))
	// Output: [key value=extra]
}

// ExampleSplit splits text through `Split` for command text handling. Text predicates and
// transforms stay on the core string wrapper surface.
func ExampleSplit() {
	parts := Split("deploy/to/homelab", "/")
	Println(parts)
	// Output: [deploy to homelab]
}

// ExampleJoin joins text fields through `Join` for command text handling. Text predicates
// and transforms stay on the core string wrapper surface.
func ExampleJoin() {
	Println(Join("/", "deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

// ExampleReplace replaces text through `Replace` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleReplace() {
	Println(Replace("deploy/to/homelab", "/", "."))
	// Output: deploy.to.homelab
}

// ExampleLower normalises text to lower case through `Lower` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleLower() {
	Println(Lower("HELLO"))
	// Output: hello
}

// ExampleUpper normalises text to upper case through `Upper` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleUpper() {
	Println(Upper("hello"))
	// Output: HELLO
}

// ExampleConcat concatenates text through `Concat` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleConcat() {
	Println(Concat("hello", " ", "world"))
	// Output: hello world
}

// ExampleTrim trims surrounding text through `Trim` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleTrim() {
	Println(Trim("  spaced  "))
	// Output: spaced
}

// ExampleRuneCount counts runes through `RuneCount` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleRuneCount() {
	Println(RuneCount("hello"))
	Println(RuneCount("é"))
	// Output:
	// 5
	// 1
}

// ExampleNewBuilder builds text incrementally through `NewBuilder` for command text
// handling. Text predicates and transforms stay on the core string wrapper surface.
func ExampleNewBuilder() {
	b := NewBuilder()
	b.WriteString("hello")
	b.WriteString(" world")
	Println(b.String())
	// Output: hello world
}

// ExampleNewReader creates a text reader through `NewReader` for command text handling.
// Text predicates and transforms stay on the core string wrapper surface.
func ExampleNewReader() {
	r := ReadAll(NewReader("hello"))
	Println(r.Value)
	// Output: hello
}

// ExampleSprint formats values as text through `Sprint` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleSprint() {
	Println(Sprint("port=", 8080))
	// Output: port=8080
}

// ExampleSprintf formats templated text through `Sprintf` for command text handling. Text
// predicates and transforms stay on the core string wrapper surface.
func ExampleSprintf() {
	Println(Sprintf("port=%d", 8080))
	// Output: port=8080
}

// ExampleHTMLEscape escapes dashboard text through `HTMLEscape` for dashboard HTML text.
// UI-bound strings are escaped and unescaped without importing html directly.
func ExampleHTMLEscape() {
	Println(HTMLEscape(`<span data-x="1">Core & Go</span>`))
	// Output: &lt;span data-x=&#34;1&#34;&gt;Core &amp; Go&lt;/span&gt;
}

// ExampleHTMLUnescape unescapes dashboard text through `HTMLUnescape` for dashboard HTML
// text. UI-bound strings are escaped and unescaped without importing html directly.
func ExampleHTMLUnescape() {
	Println(HTMLUnescape("&lt;strong&gt;Core&lt;/strong&gt;"))
	// Output: <strong>Core</strong>
}

// ExampleTrimCutset trims any character class from both ends of a string
// through `TrimCutset` for command text handling. Text predicates and
// transforms stay on the core string wrapper surface.
func ExampleTrimCutset() {
	Println(TrimCutset("//path//", "/"))
	// Output: path
}

// ExampleTrimLeft trims any leading character class through `TrimLeft`
// for command text handling. Text predicates and transforms stay on the
// core string wrapper surface.
func ExampleTrimLeft() {
	Println(TrimLeft("---verbose", "-"))
	// Output: verbose
}

// ExampleTrimRight trims any trailing character class through `TrimRight`
// for command text handling. Text predicates and transforms stay on the
// core string wrapper surface.
func ExampleTrimRight() {
	Println(TrimRight("hello!!!", "!"))
	// Output: hello
}

// ExampleIndex finds a substring position through `Index` for command text
// handling. Text predicates and transforms stay on the core string
// wrapper surface.
func ExampleIndex() {
	Println(Index("key=value", "="))
	// Output: 3
}

// ExampleBuilder declares a Builder-typed local through the `Builder`
// alias for command text handling. Text predicates and transforms stay
// on the core string wrapper surface.
func ExampleBuilder() {
	var b Builder
	b.WriteString("hello")
	b.WriteString(" world")
	Println(b.String())
	// Output: hello world
}
