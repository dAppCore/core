package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Bytes ---

func TestBytes_NewBuffer_Good(t *testing.T) {
	buf := NewBuffer([]byte("hello"))

	assert.Equal(t, "hello", buf.String())
	assert.Equal(t, 5, buf.Len())
}

func TestBytes_NewBuffer_Bad(t *testing.T) {
	buf := NewBuffer()

	assert.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len())
	assert.NoError(t, buf.WriteByte('x'))
	assert.Equal(t, "x", buf.String())
}

func TestBytes_NewBuffer_Ugly(t *testing.T) {
	src := []byte("abc")
	buf := NewBuffer(src)

	src[0] = 'z'

	assert.Equal(t, "zbc", buf.String())
}

func TestBytes_NewBufferString_Good(t *testing.T) {
	buf := NewBufferString("hello")

	assert.Equal(t, "hello", buf.String())
	assert.Equal(t, 5, buf.Len())
}

func TestBytes_NewBufferString_Bad(t *testing.T) {
	buf := NewBufferString("")

	assert.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len())
	assert.NoError(t, buf.WriteByte('x'))
	assert.Equal(t, "x", buf.String())
}

func TestBytes_NewBufferString_Ugly(t *testing.T) {
	buf := NewBufferString("a\x00b")

	assert.Equal(t, []byte{'a', 0, 'b'}, buf.Bytes())
	assert.Equal(t, 3, buf.Len())
}
