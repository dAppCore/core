// SPDX-License-Identifier: EUPL-1.2

// SHA-3 operations for the Core framework.
// Provides SHA3 and SHAKE wrappers so downstream packages can use core
// primitives for common sponge-based digest operations.

package core

import (
	"crypto/sha3"
	"encoding/hex"
)

// SHA3Keccak256 returns the 32-byte digest produced by sha3.Sum256.
//
//	sum := core.SHA3Keccak256([]byte("hello"))
func SHA3Keccak256(data []byte) [32]byte {
	return sha3.Sum256(data)
}

// SHA3Keccak256Hex returns the SHA3Keccak256 digest as lowercase hexadecimal.
//
//	sum := core.SHA3Keccak256Hex([]byte("hello"))
func SHA3Keccak256Hex(data []byte) string {
	sum := SHA3Keccak256(data)
	return hex.EncodeToString(sum[:])
}

// SHA3Shake128 returns outLen bytes from SHAKE128 applied to data.
//
//	sum := core.SHA3Shake128([]byte("hello"), 32)
func SHA3Shake128(data []byte, outLen int) []byte {
	return sha3.SumSHAKE128(data, outLen)
}

// SHA3Shake256 returns outLen bytes from SHAKE256 applied to data.
//
//	sum := core.SHA3Shake256([]byte("hello"), 64)
func SHA3Shake256(data []byte, outLen int) []byte {
	return sha3.SumSHAKE256(data, outLen)
}
