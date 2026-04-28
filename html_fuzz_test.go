// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzHTMLEscape(f *F) {
	f.Add("")
	f.Add("plain text")
	f.Add("<script>alert(\"x\")</script>")
	f.Add("&lt;already&gt;")
	f.Add("'single' & \"double\"")
	f.Add("\x00\xff")

	f.Fuzz(func(t *T, raw string) {
		escaped := HTMLEscape(raw)
		unescaped := HTMLUnescape(escaped)
		if unescaped != raw {
			t.Errorf("HTMLEscape did not round-trip raw=%q escaped=%q unescaped=%q", raw, escaped, unescaped)
		}
	})
}
