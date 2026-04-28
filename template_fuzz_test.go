// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzParseTemplate(f *F) {
	f.Add("")
	f.Add("plain text")
	f.Add("hello {{.Name}}")
	f.Add("{{")
	f.Add("{{define \"x\"}}x{{end}}{{template \"x\" .}}")
	f.Add("{{range .Items}}{{.}}{{end}}")
	f.Add("{{if .OK}}yes{{else}}no{{end}}")

	f.Fuzz(func(t *T, text string) {
		r := ParseTemplate("fuzz", text)
		if r.OK {
			tmpl := r.Value.(*Template)
			w := NewBufferString("")
			_ = ExecuteTemplate(tmpl, w, nil)
		}
	})
}
