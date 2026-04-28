package core_test

import . "dappco.re/go"

func ExampleHTMLEscape() {
	Println(HTMLEscape(`<span data-x="1">Core & Go</span>`))
	// Output: &lt;span data-x=&#34;1&#34;&gt;Core &amp; Go&lt;/span&gt;
}

func ExampleHTMLUnescape() {
	Println(HTMLUnescape("&lt;strong&gt;Core&lt;/strong&gt;"))
	// Output: <strong>Core</strong>
}
