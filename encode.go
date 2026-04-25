// SPDX-License-Identifier: EUPL-1.2

// Encoding helpers for the Core framework.
// Wraps encoding/hex so consumers can use core primitives for common
// byte-string encodings.
package core

import (
	"encoding/base64"
	"encoding/hex"
)

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

// Base64Encode returns src encoded as a standard base64 string.
//
//	s := core.Base64Encode([]byte("hello"))
func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

// Base64Decode decodes a standard base64 string into bytes.
//
//	r := core.Base64Decode("aGVsbG8=")
//	if r.OK { b := r.Value.([]byte) }
func Base64Decode(s string) Result {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{b, true}
}

// Base64URLEncode returns src encoded as a URL-safe base64 string.
//
//	s := core.Base64URLEncode([]byte("hello"))
func Base64URLEncode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// Base64URLDecode decodes a URL-safe base64 string into bytes.
//
//	r := core.Base64URLDecode("aGVsbG8=")
//	if r.OK { b := r.Value.([]byte) }
func Base64URLDecode(s string) Result {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{b, true}
}
