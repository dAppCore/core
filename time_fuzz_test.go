// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzTimeParse(f *F) {
	f.Add(TimeRFC3339, "2026-04-28T07:00:00Z")
	f.Add(TimeRFC3339Nano, "2026-04-28T07:00:00.123456789Z")
	f.Add(TimeDateTime, "2026-04-28 07:00:00")
	f.Add(TimeDateOnly, "2026-04-28")
	f.Add(TimeTimeOnly, "07:00:00")
	f.Add(TimeKitchen, "3:04PM")
	f.Add(TimeRFC3339, "not-time")
	f.Add("", "")

	f.Fuzz(func(t *T, layout, value string) {
		r := TimeParse(layout, value)
		if r.OK {
			_ = Sprint(r.Value)
		}
	})
}

func FuzzParseDuration(f *F) {
	f.Add("1s")
	f.Add("250ms")
	f.Add("2h30m")
	f.Add("0")
	f.Add("-1s")
	f.Add("100000000000000000000h")
	f.Add("xyz")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := ParseDuration(raw)
		if r.OK {
			_ = Sprint(r.Value)
		}
	})
}
