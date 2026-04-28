package core_test

import (
	. "dappco.re/go/core"
)

// --- Bytes ---

func TestBytes_NewBuffer_Good(t *T) {
	buf := NewBuffer([]byte("hello"))

	AssertEqual(t, "hello", buf.String())
	AssertEqual(t, 5, buf.Len())
}

func TestBytes_NewBuffer_Bad(t *T) {
	buf := NewBuffer()

	AssertNotNil(t, buf)
	AssertEqual(t, 0, buf.Len())
	AssertNoError(t, buf.WriteByte('x'))
	AssertEqual(t, "x", buf.String())
}

func TestBytes_NewBuffer_Ugly(t *T) {
	src := []byte("abc")
	buf := NewBuffer(src)

	src[0] = 'z'

	AssertEqual(t, "zbc", buf.String())
}

func TestBytes_NewBufferString_Good(t *T) {
	buf := NewBufferString("hello")

	AssertEqual(t, "hello", buf.String())
	AssertEqual(t, 5, buf.Len())
}

func TestBytes_NewBufferString_Bad(t *T) {
	buf := NewBufferString("")

	AssertNotNil(t, buf)
	AssertEqual(t, 0, buf.Len())
	AssertNoError(t, buf.WriteByte('x'))
	AssertEqual(t, "x", buf.String())
}

func TestBytes_NewBufferString_Ugly(t *T) {
	buf := NewBufferString("a\x00b")

	AssertEqual(t, []byte{'a', 0, 'b'}, buf.Bytes())
	AssertEqual(t, 3, buf.Len())
}
