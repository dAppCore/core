package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

// --- HTML Escape ---

func TestHTMLEscape_Good(t *testing.T) {
	AssertEqual(
		t,
		"&lt;p title=&#34;Tom &amp; Jerry&#39;s&#34;&gt;Hi&lt;/p&gt;",
		HTMLEscape(`<p title="Tom & Jerry's">Hi</p>`),
	)
}

func TestHTMLEscape_Bad(t *testing.T) {
	AssertEqual(t, "", HTMLEscape(""))
	AssertEqual(t, "Tom &amp;amp; Jerry", HTMLEscape("Tom &amp; Jerry"))
}

func TestHTMLEscape_Ugly(t *testing.T) {
	AssertEqual(t, "&#34;&amp;&#39;&lt;&gt;\x00", HTMLEscape("\"&'<>\x00"))
}

// --- HTML Unescape ---

func TestHTMLUnescape_Good(t *testing.T) {
	AssertEqual(
		t,
		`<p title="Tom & Jerry's">Hi</p>`,
		HTMLUnescape("&lt;p title=&#34;Tom &amp; Jerry&#39;s&#34;&gt;Hi&lt;/p&gt;"),
	)
}

func TestHTMLUnescape_Bad(t *testing.T) {
	AssertEqual(t, "", HTMLUnescape(""))
	AssertEqual(t, "Tom &unknown; Jerry", HTMLUnescape("Tom &unknown; Jerry"))
}

func TestHTMLUnescape_Ugly(t *testing.T) {
	AssertEqual(t, "\"&'<>\x00", HTMLUnescape("&#34;&amp;&#39;&lt;&gt;\x00"))
}
