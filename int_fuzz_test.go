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

func FuzzParseInt(f *F) {
	f.Add("0", 10, 64)
	f.Add("1", 10, 0)
	f.Add("-1", 10, 64)
	f.Add("9223372036854775807", 10, 64)
	f.Add("-9223372036854775808", 10, 64)
	f.Add("101010", 2, 64)
	f.Add("zzzz", 36, 64)
	f.Add("0xff", 0, 64)
	f.Add("0777", 0, 32)
	f.Add("999999999999999999999", 10, 64)
	f.Add("", 10, 64)

	f.Fuzz(func(t *T, raw string, base int, bitSize int) {
		r := ParseInt(raw, base, bitSize)
		if r.OK {
			v := r.Value.(int64)
			formatBase := base
			parseBase := base
			if base == 0 {
				formatBase = 10
			}
			if formatBase < 2 || formatBase > 36 {
				t.Errorf("ParseInt accepted invalid base raw=%q base=%d bitSize=%d value=%d", raw, base, bitSize, v)
				return
			}

			encoded := FormatInt(v, formatBase)
			roundTrip := ParseInt(encoded, parseBase, bitSize)
			if !roundTrip.OK {
				t.Errorf("ParseInt produced int64 that FormatInt cannot parse raw=%q base=%d bitSize=%d encoded=%q value=%d", raw, base, bitSize, encoded, v)
				return
			}
			if roundTrip.Value.(int64) != v {
				t.Errorf("ParseInt round-trip mismatch raw=%q base=%d bitSize=%d encoded=%q want=%d got=%d", raw, base, bitSize, encoded, v, roundTrip.Value.(int64))
			}
		}
	})
}
