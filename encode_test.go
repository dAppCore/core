package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- HexEncode ---

func TestEncode_HexEncode_Good(t *testing.T) {
	assert.Equal(t, "68656c6c6f", HexEncode([]byte("hello")))
}

func TestEncode_HexEncode_Bad(t *testing.T) {
	assert.Equal(t, "", HexEncode(nil))
	assert.Equal(t, HexEncode(nil), HexEncode([]byte{}))
}

func TestEncode_HexEncode_Ugly(t *testing.T) {
	src := []byte{0x00, 0x0f, 0x10, 0xff}
	encoded := HexEncode(src)

	src[0] = 0xff

	assert.Equal(t, "000f10ff", encoded)
	assert.NotEqual(t, encoded, HexEncode(src))
}

// --- HexDecode ---

func TestEncode_HexDecode_Good(t *testing.T) {
	r := HexDecode("68656c6c6f")
	assert.True(t, r.OK)
	assert.Equal(t, []byte("hello"), r.Value)
}

func TestEncode_HexDecode_Bad(t *testing.T) {
	r := HexDecode("not-hex")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestEncode_HexDecode_Ugly(t *testing.T) {
	r := HexDecode("abc")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

// --- Base64Encode ---

func TestEncode_Base64Encode_Good(t *testing.T) {
	assert.Equal(t, "aGVsbG8=", Base64Encode([]byte("hello")))
}

func TestEncode_Base64Encode_Bad(t *testing.T) {
	assert.Equal(t, "", Base64Encode(nil))
	assert.Equal(t, Base64Encode(nil), Base64Encode([]byte{}))
}

func TestEncode_Base64Encode_Ugly(t *testing.T) {
	src := []byte{0xfb, 0xff, 0xff}
	encoded := Base64Encode(src)

	src[0] = 0x00

	assert.Equal(t, "+///", encoded)
	assert.NotEqual(t, encoded, Base64Encode(src))
}

// --- Base64Decode ---

func TestEncode_Base64Decode_Good(t *testing.T) {
	r := Base64Decode("aGVsbG8=")
	assert.True(t, r.OK)
	assert.Equal(t, []byte("hello"), r.Value)
}

func TestEncode_Base64Decode_Bad(t *testing.T) {
	r := Base64Decode("not-base64")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestEncode_Base64Decode_Ugly(t *testing.T) {
	r := Base64Decode("aGVsbG8===")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

// --- Base64URLEncode ---

func TestEncode_Base64URLEncode_Good(t *testing.T) {
	assert.Equal(t, "aGVsbG8=", Base64URLEncode([]byte("hello")))
}

func TestEncode_Base64URLEncode_Bad(t *testing.T) {
	assert.Equal(t, "", Base64URLEncode(nil))
	assert.Equal(t, Base64URLEncode(nil), Base64URLEncode([]byte{}))
}

func TestEncode_Base64URLEncode_Ugly(t *testing.T) {
	src := []byte{0xfb, 0xff, 0xff}
	encoded := Base64URLEncode(src)

	src[0] = 0x00

	assert.Equal(t, "-___", encoded)
	assert.NotEqual(t, encoded, Base64URLEncode(src))
}

// --- Base64URLDecode ---

func TestEncode_Base64URLDecode_Good(t *testing.T) {
	r := Base64URLDecode("aGVsbG8=")
	assert.True(t, r.OK)
	assert.Equal(t, []byte("hello"), r.Value)
}

func TestEncode_Base64URLDecode_Bad(t *testing.T) {
	r := Base64URLDecode("not+url")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestEncode_Base64URLDecode_Ugly(t *testing.T) {
	r := Base64URLDecode("aGVsbG8===")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}
