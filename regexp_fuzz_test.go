// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzRegex(f *F) {
	f.Add(`\d+`)
	f.Add(`^.*$`)
	f.Add(`(\w+)=(\d+)`)
	f.Add(`(a+)+$`)
	f.Add(`[`)
	f.Add(`(`)
	f.Add(`\C`)
	f.Add(``)

	f.Fuzz(func(t *T, pattern string) {
		r := Regex(pattern)
		if r.OK {
			rx := r.Value.(*Regexp)
			_ = rx.String()
			_ = rx.MatchString(pattern)
			_ = rx.FindString(pattern)
			_ = rx.FindAllString(pattern, -1)
			_ = rx.FindStringSubmatch(pattern)
			_ = rx.ReplaceAllString(pattern, "x")
			_ = rx.Split(pattern, -1)
		}
	})
}
