package core_test

import (
	"encoding/hex"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

const (
	sha3Keccak256EmptyHex = "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	sha3Keccak256HelloHex = "3338be694f50c5f338814986cdf0686453a888b84f424d792af4b9202398f392"
	sha3Keccak256QuickHex = "69070dda01975c8c120c3aada1b282394e7f032fa9cf32f4cb2259a0897dfc04"

	sha3Shake128Empty32Hex = "7f9c2ba4e88f827d616045507605853ed73b8093f6efbc88eb1a6eacfa66ef26"
	sha3Shake256Empty64Hex = "46b9dd2b0ba88d13233b3feb743eeb243fcd52ea62b81b82b50c27646ed5762fd75dc4ddd8c0f200cb05019d67b592f6fc821c49479ab48640292eacb3b7c4be"
)

// --- SHA3 ---

func TestSHA3_Keccak256_Good(t *testing.T) {
	sum := SHA3Keccak256([]byte("hello"))

	assert.Equal(t, sha3Keccak256HelloHex, hex.EncodeToString(sum[:]))
	assert.Equal(t, sha3Keccak256HelloHex, SHA3Keccak256Hex([]byte("hello")))
}

func TestSHA3_Keccak256_Bad(t *testing.T) {
	sum := SHA3Keccak256(nil)

	assert.Equal(t, sha3Keccak256EmptyHex, hex.EncodeToString(sum[:]))
	assert.Equal(t, SHA3Keccak256(nil), SHA3Keccak256([]byte{}))
	assert.Equal(t, SHA3Keccak256Hex(nil), SHA3Keccak256Hex([]byte{}))
}

func TestSHA3_Keccak256_Ugly(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := SHA3Keccak256(data)
	sumHex := SHA3Keccak256Hex(data)

	data[0] = 't'

	assert.Equal(t, sha3Keccak256QuickHex, hex.EncodeToString(sum[:]))
	assert.Equal(t, sha3Keccak256QuickHex, sumHex)
	assert.NotEqual(t, sum, SHA3Keccak256(data))
	assert.NotEqual(t, sumHex, SHA3Keccak256Hex(data))
}

func TestSHA3_Shake_Good(t *testing.T) {
	assert.Equal(t, sha3Shake128Empty32Hex, hex.EncodeToString(SHA3Shake128(nil, 32)))
	assert.Equal(t, sha3Shake256Empty64Hex, hex.EncodeToString(SHA3Shake256(nil, 64)))
}

func TestSHA3_Shake_Bad(t *testing.T) {
	assert.Empty(t, SHA3Shake128(nil, 0))
	assert.Empty(t, SHA3Shake256([]byte{}, 0))
	assert.Equal(t, SHA3Shake128(nil, 16), SHA3Shake128([]byte{}, 16))
	assert.Equal(t, SHA3Shake256(nil, 16), SHA3Shake256([]byte{}, 16))
	assert.Panics(t, func() { SHA3Shake128(nil, -1) })
	assert.Panics(t, func() { SHA3Shake256(nil, -1) })
}

func TestSHA3_Shake_Ugly(t *testing.T) {
	data := []byte("abc")
	out128 := SHA3Shake128(data, 32)
	out256 := SHA3Shake256(data, 64)

	assert.Equal(t, SHA3Shake128(data, 16), out128[:16])
	assert.Equal(t, SHA3Shake256(data, 32), out256[:32])

	data[0] = 'z'

	assert.NotEqual(t, out128, SHA3Shake128(data, 32))
	assert.NotEqual(t, out256, SHA3Shake256(data, 64))
}
