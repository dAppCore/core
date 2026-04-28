package core_test

import . "dappco.re/go"

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
