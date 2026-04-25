// SPDX-License-Identifier: EUPL-1.2

// Encoding helpers for the Core framework.
// Wraps encoding/hex so consumers can use core primitives for common
// byte-string encodings.
package core

import "encoding/hex"

// HexEncode returns src encoded as a lowercase hexadecimal string.
//
//	s := core.HexEncode([]byte("hello"))
func HexEncode(src []byte) string {
	return hex.EncodeToString(src)
}

// HexDecode decodes a hexadecimal string into bytes.
//
//	r := core.HexDecode("68656c6c6f")
//	if r.OK { b := r.Value.([]byte) }
func HexDecode(s string) Result {
	b, err := hex.DecodeString(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{b, true}
}
