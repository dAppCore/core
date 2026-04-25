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
