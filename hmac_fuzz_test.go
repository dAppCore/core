// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzHMAC(f *F) {
	f.Add("sha256", []byte("key"), []byte("data"))
	f.Add("sha512", []byte("secret"), []byte(""))
	f.Add("", []byte(""), []byte(""))
	f.Add("codex-fuzz-unsupported-algo", []byte("key"), []byte("data"))
	f.Add("SHA256", []byte{0, 1, 2}, []byte{3, 4, 5})

	f.Fuzz(func(t *T, algo string, key []byte, data []byte) {
		r := HMAC(algo, key, data)

		unsupported := HMAC("codex-fuzz-unsupported-algo", key, data)
		if unsupported.OK {
			t.Errorf("HMAC accepted unsupported algorithm")
		}

		if r.OK {
			mac := r.Value.([]byte)
			again := HMAC(algo, key, data)
			if !again.OK {
				t.Errorf("HMAC result cannot be reproduced algo=%q", algo)
				return
			}

			againMAC := again.Value.([]byte)
			if len(mac) != len(againMAC) {
				t.Errorf("HMAC length changed algo=%q first=%d second=%d", algo, len(mac), len(againMAC))
				return
			}

			for i := range mac {
				if mac[i] != againMAC[i] {
					t.Errorf("HMAC is not deterministic algo=%q index=%d", algo, i)
					return
				}
			}
		}
	})
}
