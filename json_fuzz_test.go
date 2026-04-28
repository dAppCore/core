// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzJSONUnmarshal(f *F) {
	f.Add(`{"k":"v"}`)
	f.Add(`[]`)
	f.Add(`42`)
	f.Add(`{"k":}`)
	f.Add(`{"a":{"b":[1,2]}}`)
	f.Add(`[[[[[[[[[[0]]]]]]]]]]`)
	f.Add(`"\u2603"`)
	f.Add(`1e100000000000000000000`)
	f.Add(``)

	f.Fuzz(func(t *T, raw string) {
		var target any
		r := JSONUnmarshal([]byte(raw), &target)
		if r.OK {
			if roundTrip := JSONMarshal(target); !roundTrip.OK {
				t.Errorf("JSONUnmarshal produced unmarshalable value raw=%q value=%v", raw, target)
			}
		}
	})
}

func FuzzJSONUnmarshalString(f *F) {
	f.Add(`{"k":"v"}`)
	f.Add(`[]`)
	f.Add(`42`)
	f.Add(`{"k":}`)
	f.Add(`{"a":{"b":[1,2]}}`)
	f.Add(`[[[[[[[[[[0]]]]]]]]]]`)
	f.Add(`"\u2603"`)
	f.Add(`1e100000000000000000000`)
	f.Add(``)

	f.Fuzz(func(t *T, raw string) {
		var target any
		r := JSONUnmarshalString(raw, &target)
		if r.OK {
			if roundTrip := JSONMarshal(target); !roundTrip.OK {
				t.Errorf("JSONUnmarshalString produced unmarshalable value raw=%q value=%v", raw, target)
			}
		}
	})
}
