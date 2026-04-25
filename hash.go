// SPDX-License-Identifier: EUPL-1.2

// Hash operations for the Core framework.
// Provides crypto/sha256 wrappers so downstream packages can use core
// primitives for common digest operations.

package core

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 returns the SHA-256 digest of data.
//
//	sum := core.SHA256([]byte("hello"))
func SHA256(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// SHA256Hex returns the SHA-256 digest of data as lowercase hexadecimal.
//
//	sum := core.SHA256Hex([]byte("hello"))
func SHA256Hex(data []byte) string {
	sum := SHA256(data)
	return hex.EncodeToString(sum[:])
}

// SHA256String returns the SHA-256 digest of s.
//
//	sum := core.SHA256String("hello")
func SHA256String(s string) [32]byte {
	return SHA256([]byte(s))
}

// SHA256HexString returns the SHA-256 digest of s as lowercase hexadecimal.
//
//	sum := core.SHA256HexString("hello")
func SHA256HexString(s string) string {
	return SHA256Hex([]byte(s))
}
