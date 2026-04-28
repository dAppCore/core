package core_test

import (
	. "dappco.re/go"
)

const (
	sha3_256EmptyHex = "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	sha3_256HelloHex = "3338be694f50c5f338814986cdf0686453a888b84f424d792af4b9202398f392"
	sha3_256QuickHex = "69070dda01975c8c120c3aada1b282394e7f032fa9cf32f4cb2259a0897dfc04"

	keccak256EmptyHex = "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	keccak256HelloHex = "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"
	keccak256QuickHex = "4d741b6f1eb29cb2a9b9911c82f56fa8d73b04959d3d9d222895df6c0b28aa15"

	sha3Shake128Empty32Hex = "7f9c2ba4e88f827d616045507605853ed73b8093f6efbc88eb1a6eacfa66ef26"
	sha3Shake256Empty64Hex = "46b9dd2b0ba88d13233b3feb743eeb243fcd52ea62b81b82b50c27646ed5762fd75dc4ddd8c0f200cb05019d67b592f6fc821c49479ab48640292eacb3b7c4be"
)

func TestSha3_Keccak256_Good(t *T) {
	sum := Keccak256([]byte("hello"))

	AssertEqual(t, keccak256HelloHex, HexEncode(sum[:]))
}

func TestSha3_Keccak256_Bad(t *T) {
	sum := Keccak256(nil)

	AssertEqual(t, keccak256EmptyHex, HexEncode(sum[:]))
	AssertNotEqual(t, SHA3_256(nil), sum)
}

func TestSha3_Keccak256_Ugly(t *T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := Keccak256(data)
	data[0] = 't'

	AssertEqual(t, keccak256QuickHex, HexEncode(sum[:]))
	AssertNotEqual(t, sum, Keccak256(data))
}

func TestSha3_Keccak256Hex_Good(t *T) {
	AssertEqual(t, keccak256HelloHex, Keccak256Hex([]byte("hello")))
}

func TestSha3_Keccak256Hex_Bad(t *T) {
	AssertEqual(t, keccak256EmptyHex, Keccak256Hex(nil))
}

func TestSha3_Keccak256Hex_Ugly(t *T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := Keccak256Hex(data)
	data[0] = 't'

	AssertEqual(t, keccak256QuickHex, sum)
	AssertNotEqual(t, sum, Keccak256Hex(data))
}

func TestSha3_SHA3_256_Good(t *T) {
	sum := SHA3_256([]byte("hello"))

	AssertEqual(t, sha3_256HelloHex, HexEncode(sum[:]))
}

func TestSha3_SHA3_256_Bad(t *T) {
	sum := SHA3_256(nil)

	AssertEqual(t, sha3_256EmptyHex, HexEncode(sum[:]))
}

func TestSha3_SHA3_256_Ugly(t *T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := SHA3_256(data)
	data[0] = 't'

	AssertEqual(t, sha3_256QuickHex, HexEncode(sum[:]))
	AssertNotEqual(t, sum, SHA3_256(data))
}

func TestSha3_SHA3_256Hex_Good(t *T) {
	AssertEqual(t, sha3_256HelloHex, SHA3_256Hex([]byte("hello")))
}

func TestSha3_SHA3_256Hex_Bad(t *T) {
	AssertEqual(t, sha3_256EmptyHex, SHA3_256Hex(nil))
}

func TestSha3_SHA3_256Hex_Ugly(t *T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := SHA3_256Hex(data)
	data[0] = 't'

	AssertEqual(t, sha3_256QuickHex, sum)
	AssertNotEqual(t, sum, SHA3_256Hex(data))
}

func TestSha3_SHA3Shake128_Good(t *T) {
	AssertEqual(t, sha3Shake128Empty32Hex, HexEncode(SHA3Shake128(nil, 32)))
}

func TestSha3_SHA3Shake128_Bad(t *T) {
	AssertEmpty(t, SHA3Shake128(nil, 0))
	AssertPanics(t, func() { SHA3Shake128(nil, -1) })
}

func TestSha3_SHA3Shake128_Ugly(t *T) {
	out := SHA3Shake128([]byte("agent"), 32)

	AssertEqual(t, SHA3Shake128([]byte("agent"), 16), out[:16])
}

func TestSha3_SHA3Shake256_Good(t *T) {
	AssertEqual(t, sha3Shake256Empty64Hex, HexEncode(SHA3Shake256(nil, 64)))
}

func TestSha3_SHA3Shake256_Bad(t *T) {
	AssertEmpty(t, SHA3Shake256(nil, 0))
	AssertPanics(t, func() { SHA3Shake256(nil, -1) })
}

func TestSha3_SHA3Shake256_Ugly(t *T) {
	out := SHA3Shake256([]byte("agent"), 64)

	AssertEqual(t, SHA3Shake256([]byte("agent"), 32), out[:32])
}
