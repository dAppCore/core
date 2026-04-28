// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzAtoi(f *F) {
	f.Add("0")
	f.Add("1")
	f.Add("-1")
	f.Add("9223372036854775807")
	f.Add("999999999999999999999")
	f.Add("abc")
	f.Add("")
	f.Add("1e10")

	f.Fuzz(func(t *T, raw string) {
		r := Atoi(raw)
		if r.OK {
			n := r.Value.(int)
			roundTrip := Atoi(Itoa(n))
			if !roundTrip.OK {
				t.Errorf("Atoi produced int that Itoa cannot parse raw=%q value=%d", raw, n)
			}
		}
	})
}
