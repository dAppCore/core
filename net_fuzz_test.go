// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzParseIP(f *F) {
	f.Add("192.0.2.1")
	f.Add("2001:db8::1")
	f.Add("::1")
	f.Add("::ffff:192.0.2.1")
	f.Add("255.255.255.255")
	f.Add("garbage")
	f.Add("1.2.3.256")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		ip := ParseIP(raw)
		if ip != nil {
			if len(ip) == 0 {
				t.Errorf("ParseIP returned empty IP raw=%q", raw)
			}
			_ = ip.String()
		}
	})
}

func FuzzParseCIDR(f *F) {
	f.Add("10.0.0.0/24")
	f.Add("192.0.2.1/32")
	f.Add("0.0.0.0/0")
	f.Add("2001:db8::1/128")
	f.Add("2001:db8::/32")
	f.Add("::/0")
	f.Add("garbage")
	f.Add("1.2.3.256/24")
	f.Add("1.2.3.4/33")
	f.Add("2001:db8::1/129")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := ParseCIDR(raw)
		if r.OK {
			parts := r.Value.([]any)
			if len(parts) != 2 {
				t.Errorf("ParseCIDR OK with unexpected value shape raw=%q len=%d", raw, len(parts))
				return
			}
			if parts[0] == nil || parts[1] == nil {
				t.Errorf("ParseCIDR OK with nil part raw=%q parts=%v", raw, parts)
			}
		}
	})
}
