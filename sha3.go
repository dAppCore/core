// SPDX-License-Identifier: EUPL-1.2

// SHA-3 operations for the Core framework.
// Provides SHA3, legacy Keccak, and SHAKE wrappers so downstream packages can
// use core primitives for common sponge-based digest operations.

package core

import (
	"crypto/sha3"
	"encoding/binary"
	"encoding/hex"
	"math/bits"
)

const keccak256Rate = 136

var keccakfRoundConstants = [24]uint64{
	0x0000000000000001,
	0x0000000000008082,
	0x800000000000808a,
	0x8000000080008000,
	0x000000000000808b,
	0x0000000080000001,
	0x8000000080008081,
	0x8000000000008009,
	0x000000000000008a,
	0x0000000000000088,
	0x0000000080008009,
	0x000000008000000a,
	0x000000008000808b,
	0x800000000000008b,
	0x8000000000008089,
	0x8000000000008003,
	0x8000000000008002,
	0x8000000000000080,
	0x000000000000800a,
	0x800000008000000a,
	0x8000000080008081,
	0x8000000000008080,
	0x0000000080000001,
	0x8000000080008008,
}

var keccakfRotationOffsets = [25]int{
	0, 1, 62, 28, 27,
	36, 44, 6, 55, 20,
	3, 10, 43, 25, 39,
	41, 45, 15, 21, 8,
	18, 2, 61, 56, 14,
}

// SHA3_256 returns the FIPS-202 SHA3-256 digest of data.
//
// SHA3-256 uses the standardized SHA-3 padding/domain separator (0x06). It is
// not pre-NIST Keccak-256, which uses legacy Keccak padding (0x01). Use
// Keccak256 for Ethereum, EVM chains, ENS, RLP signing, and other Web3
// consensus hashes.
//
//	sum := core.SHA3_256([]byte("hello"))
func SHA3_256(data []byte) [32]byte {
	return sha3.Sum256(data)
}

// SHA3_256Hex returns the FIPS-202 SHA3-256 digest as lowercase hexadecimal.
//
//	sum := core.SHA3_256Hex([]byte("hello"))
func SHA3_256Hex(data []byte) string {
	sum := SHA3_256(data)
	return hex.EncodeToString(sum[:])
}

// Keccak256 returns the legacy pre-NIST Keccak-256 digest of data.
//
// Keccak-256 uses legacy Keccak padding/domain separator (0x01). It differs
// from FIPS-202 SHA3-256, which uses standardized SHA-3 padding (0x06).
//
//	sum := core.Keccak256([]byte("hello"))
func Keccak256(data []byte) [32]byte {
	var state [25]uint64

	for len(data) >= keccak256Rate {
		for i := 0; i < keccak256Rate/8; i++ {
			state[i] ^= binary.LittleEndian.Uint64(data[i*8:])
		}
		keccakF1600(&state)
		data = data[keccak256Rate:]
	}

	var block [keccak256Rate]byte
	copy(block[:], data)
	block[len(data)] = 0x01
	block[keccak256Rate-1] ^= 0x80
	for i := 0; i < keccak256Rate/8; i++ {
		state[i] ^= binary.LittleEndian.Uint64(block[i*8:])
	}
	keccakF1600(&state)

	var sum [32]byte
	for i := 0; i < len(sum)/8; i++ {
		binary.LittleEndian.PutUint64(sum[i*8:], state[i])
	}
	return sum
}

// Keccak256Hex returns the legacy pre-NIST Keccak-256 digest as lowercase
// hexadecimal.
//
//	sum := core.Keccak256Hex([]byte("hello"))
func Keccak256Hex(data []byte) string {
	sum := Keccak256(data)
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

func keccakF1600(a *[25]uint64) {
	for _, rc := range keccakfRoundConstants {
		var c, d [5]uint64
		for x := 0; x < 5; x++ {
			c[x] = a[x] ^ a[x+5] ^ a[x+10] ^ a[x+15] ^ a[x+20]
		}
		for x := 0; x < 5; x++ {
			d[x] = c[(x+4)%5] ^ bits.RotateLeft64(c[(x+1)%5], 1)
		}
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				a[x+5*y] ^= d[x]
			}
		}

		var b [25]uint64
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				b[y+5*((2*x+3*y)%5)] = bits.RotateLeft64(
					a[x+5*y],
					keccakfRotationOffsets[x+5*y],
				)
			}
		}
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				a[x+5*y] = b[x+5*y] ^ (^b[(x+1)%5+5*y] & b[(x+2)%5+5*y])
			}
		}
		a[0] ^= rc
	}
}
