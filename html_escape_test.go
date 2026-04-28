package core_test

import (
	. "dappco.re/go"
)

func TestHtml_escape_HTMLEscape_Good(t *T) {
	AssertEqual(
		t,
		"&lt;p title=&#34;agent &amp; dispatch&#34;&gt;ready&lt;/p&gt;",
		HTMLEscape(`<p title="agent & dispatch">ready</p>`),
	)
}

func TestHtml_escape_HTMLEscape_Bad(t *T) {
	AssertEqual(t, "", HTMLEscape(""))
}

func TestHtml_escape_HTMLEscape_Ugly(t *T) {
	AssertEqual(t, "&#34;&amp;&#39;&lt;&gt;\x00", HTMLEscape("\"&'<>\x00"))
}

func TestHtml_escape_HTMLUnescape_Good(t *T) {
	AssertEqual(
		t,
		`<p title="agent & dispatch">ready</p>`,
		HTMLUnescape("&lt;p title=&#34;agent &amp; dispatch&#34;&gt;ready&lt;/p&gt;"),
	)
}

func TestHtml_escape_HTMLUnescape_Bad(t *T) {
	AssertEqual(t, "", HTMLUnescape(""))
}

func TestHtml_escape_HTMLUnescape_Ugly(t *T) {
	AssertEqual(t, "agent &unknown; dispatch", HTMLUnescape("agent &unknown; dispatch"))
}
