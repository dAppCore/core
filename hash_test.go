package core_test

import (
	"encoding/hex"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

const (
	sha256EmptyHex = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sha256HelloHex = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	sha256QuickHex = "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"
	hmacSHA256Hex  = "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	hmacSHA512Hex  = "b42af09057bac1e2d41708e48a902e09b5ff7f12ab428a4fe86653c73dd248fb82f948a549f7b791a5b41915ee4d1ec3935357e4e2317250d0372afa2ebeeb3a"
	hkdfSHA256Hex  = "f6d2fcc47cb939deafe3853a1e641a27e6924aff7a63d09cb04ccfffbe4776ef"
	hkdfSHA512Hex  = "90e269f053d383c4b2070be93238adf358f3d67bd7b17ca3de95f10a50a8385e"
)

// --- Hash ---

func TestHash_SHA256_Good(t *testing.T) {
	assert.Equal(t, digestFromHex(t, sha256HelloHex), SHA256([]byte("hello")))
}

func TestHash_SHA256_Bad(t *testing.T) {
	assert.Equal(t, digestFromHex(t, sha256EmptyHex), SHA256(nil))
	assert.Equal(t, SHA256(nil), SHA256([]byte{}))
}

func TestHash_SHA256_Ugly(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := SHA256(data)

	data[0] = 't'

	assert.Equal(t, digestFromHex(t, sha256QuickHex), sum)
	assert.NotEqual(t, sum, SHA256(data))
}

func TestHash_SHA256Hex_Good(t *testing.T) {
	assert.Equal(t, sha256HelloHex, SHA256Hex([]byte("hello")))
}

func TestHash_SHA256Hex_Bad(t *testing.T) {
	assert.Equal(t, sha256EmptyHex, SHA256Hex(nil))
	assert.Equal(t, SHA256Hex(nil), SHA256Hex([]byte{}))
}

func TestHash_SHA256Hex_Ugly(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	sum := SHA256Hex(data)

	data[0] = 't'

	assert.Equal(t, sha256QuickHex, sum)
	assert.NotEqual(t, sum, SHA256Hex(data))
}

func TestHash_SHA256String_Good(t *testing.T) {
	assert.Equal(t, SHA256([]byte("hello")), SHA256String("hello"))
}

func TestHash_SHA256String_Bad(t *testing.T) {
	assert.Equal(t, SHA256([]byte{}), SHA256String(""))
}

func TestHash_SHA256String_Ugly(t *testing.T) {
	s := "line 1\nline 2\t\x00"

	assert.Equal(t, SHA256([]byte(s)), SHA256String(s))
}

func TestHash_SHA256HexString_Good(t *testing.T) {
	assert.Equal(t, SHA256Hex([]byte("hello")), SHA256HexString("hello"))
}

func TestHash_SHA256HexString_Bad(t *testing.T) {
	assert.Equal(t, SHA256Hex([]byte{}), SHA256HexString(""))
}

func TestHash_SHA256HexString_Ugly(t *testing.T) {
	s := "line 1\nline 2\t\x00"

	assert.Equal(t, SHA256Hex([]byte(s)), SHA256HexString(s))
}

func TestHash_HMAC_Good(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")

	assert.Equal(t, hmacSHA256Hex, HexEncode(HMAC("sha256", []byte("key"), data)))
}

func TestHash_HMAC_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = HMAC("md5", []byte("key"), []byte("data"))
	})
}

func TestHash_HMAC_Ugly(t *testing.T) {
	key := []byte("key")
	data := []byte("The quick brown fox jumps over the lazy dog")
	digest := HMAC("sha512", key, data)

	key[0] = 'K'
	data[0] = 't'

	assert.Equal(t, hmacSHA512Hex, HexEncode(digest))
	assert.NotEqual(t, digest, HMAC("sha512", key, data))
}

func TestHash_HKDF_Good(t *testing.T) {
	key := HKDF("sha256", []byte("secret"), []byte("salt"), []byte("info"), 32)

	assert.Equal(t, hkdfSHA256Hex, HexEncode(key))
}

func TestHash_HKDF_Bad(t *testing.T) {
	assert.Panics(t, func() {
		_ = HKDF("md5", []byte("secret"), nil, nil, 16)
	})
}

func TestHash_HKDF_Ugly(t *testing.T) {
	key := HKDF("sha512", []byte("secret"), []byte("salt"), []byte("info"), 32)
	empty := HKDF("sha256", []byte("secret"), []byte("salt"), []byte("info"), 0)

	assert.Equal(t, hkdfSHA512Hex, HexEncode(key))
	assert.Equal(t, []byte{}, empty)
}

func digestFromHex(t *testing.T, want string) [32]byte {
	t.Helper()

	b, err := hex.DecodeString(want)
	if err != nil {
		t.Fatalf("invalid SHA-256 fixture: %v", err)
	}
	if len(b) != 32 {
		t.Fatalf("invalid SHA-256 fixture length: %d", len(b))
	}

	var digest [32]byte
	copy(digest[:], b)
	return digest
}
