// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzURLParse(f *F) {
	f.Add("https://example.com/path")
	f.Add("http://localhost:8080")
	f.Add("ftp://files.example.com/file.txt")
	f.Add("https://user:pass@example.com/private?q=1")
	f.Add("https://例え.テスト/パス?q=値")
	f.Add("example.com/path")
	f.Add("://missing-scheme")
	f.Add("http://%zz")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := URLParse(raw)
		if r.OK {
			if r.Value == nil {
				t.Errorf("URLParse OK with nil value raw=%q", raw)
			}
			_ = URLNormalize(raw)
		}
	})
}

func FuzzURLDecode(f *F) {
	f.Add("hello+world")
	f.Add("a%2Fb%3Fc%3D1")
	f.Add("%E2%9C%93")
	f.Add("%")
	f.Add("%zz")
	f.Add("%u1234")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := URLDecode(raw)
		if r.OK {
			decoded := r.Value.(string)
			_ = URLPathEscape(decoded)
		}
	})
}

func FuzzURLPathEscape(f *F) {
	f.Add("a/b")
	f.Add("hello world")
	f.Add("a+b")
	f.Add("%")
	f.Add("ümlaut/路径")
	f.Add("\x00")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		escaped := URLPathEscape(raw)
		if Contains(escaped, " ") {
			t.Errorf("URLPathEscape left literal space raw=%q escaped=%q", raw, escaped)
		}
		_ = URLDecode(escaped)
	})
}

func FuzzURLNormalize(f *F) {
	f.Add("https://example.com/path")
	f.Add("http://localhost:8080")
	f.Add("ftp://files.example.com/file.txt")
	f.Add("https://user:pass@example.com/private?q=1")
	f.Add("https://例え.テスト/パス?q=値")
	f.Add("example.com/path")
	f.Add("http://example.com/a b")
	f.Add("http://[::1")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		normalized := URLNormalize(raw)
		if normalized == "" {
			return
		}
		if r := URLParse(normalized); !r.OK {
			t.Errorf("URLNormalize produced unparsable URL raw=%q normalized=%q", raw, normalized)
		}
	})
}
