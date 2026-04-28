// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzHexDecode(f *F) {
	f.Add("deadbeef")
	f.Add("DEADBEEF")
	f.Add("68656c6c6f")
	f.Add("abc")
	f.Add("zz")
	f.Add("0")
	f.Add("f")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := HexDecode(raw)
		if r.OK {
			b := r.Value.([]byte)
			encoded := HexEncode(b)
			roundTrip := HexDecode(Lower(encoded))
			if !roundTrip.OK {
				t.Errorf("HexDecode produced bytes that HexEncode cannot decode raw=%q encoded=%q", raw, encoded)
			}
		}
	})
}

func FuzzBase64Decode(f *F) {
	f.Add("SGVsbG8=")
	f.Add("TWE=")
	f.Add("TQ==")
	f.Add("SGVsbG8")
	f.Add("!!!!")
	f.Add("A")
	f.Add("====")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := Base64Decode(raw)
		if r.OK {
			b := r.Value.([]byte)
			encoded := Base64Encode(b)
			roundTrip := Base64Decode(encoded)
			if !roundTrip.OK {
				t.Errorf("Base64Decode produced bytes that Base64Encode cannot decode raw=%q encoded=%q", raw, encoded)
			}
		}
	})
}

func FuzzBase64URLDecode(f *F) {
	f.Add("aGVsbG8=")
	f.Add("TWE=")
	f.Add("TQ==")
	f.Add("-_8=")
	f.Add("YW55LWJ5dGVz")
	f.Add("SGVsbG8")
	f.Add("!!!!")
	f.Add("A")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := Base64URLDecode(raw)
		if r.OK {
			b := r.Value.([]byte)
			encoded := Base64URLEncode(b)
			roundTrip := Base64URLDecode(encoded)
			if !roundTrip.OK {
				t.Errorf("Base64URLDecode produced bytes that Base64URLEncode cannot decode raw=%q encoded=%q", raw, encoded)
			}
		}
	})
}
