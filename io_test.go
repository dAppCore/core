// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

type ioFailingReader struct{}

func (ioFailingReader) Read([]byte) (int, error) {
	return 0, AnError
}

type ioFailingWriter struct{}

func (ioFailingWriter) Write([]byte) (int, error) {
	return 0, AnError
}

type trackingReadCloser struct {
	reader Reader
	closed bool
}

func (r *trackingReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}

func TestIo_EOF_Good(t *T) {
	wrapped := ErrorJoin(EOF)

	AssertTrue(t, Is(wrapped, EOF))
}

func TestIo_Copy_Good(t *T) {
	src := NewReader("agent ready")
	dst := NewBuffer()
	r := Copy(dst, src)

	AssertTrue(t, r.OK)
	AssertEqual(t, int64(11), r.Value.(int64))
	AssertEqual(t, "agent ready", dst.String())
}

func TestIo_Copy_Bad(t *T) {
	r := Copy(NewBuffer(), ioFailingReader{})

	AssertFalse(t, r.OK)
	AssertEqual(t, AnError, r.Value)
}

func TestIo_Copy_Ugly(t *T) {
	r := Copy(NewBuffer(), NewReader(""))

	AssertTrue(t, r.OK)
	AssertEqual(t, int64(0), r.Value.(int64))
}

func TestIo_CopyN_Good(t *T) {
	src := NewReader("agent ready")
	dst := NewBuffer()
	r := CopyN(dst, src, 5)

	AssertTrue(t, r.OK)
	AssertEqual(t, int64(5), r.Value.(int64))
	AssertEqual(t, "agent", dst.String())
}

func TestIo_CopyN_Bad(t *T) {
	r := CopyN(NewBuffer(), NewReader("abc"), 100)

	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), EOF)
}

func TestIo_CopyN_Ugly(t *T) {
	r := CopyN(NewBuffer(), NewReader("agent"), 0)

	AssertTrue(t, r.OK)
	AssertEqual(t, int64(0), r.Value.(int64))
}

func TestIo_ReadAll_Good(t *T) {
	r := ReadAll(NewReader("agent ready"))

	AssertTrue(t, r.OK)
	AssertEqual(t, "agent ready", r.Value)
}

func TestIo_ReadAll_Bad(t *T) {
	r := ReadAll("not a reader")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "not a reader")
}

func TestIo_ReadAll_Ugly(t *T) {
	reader := &trackingReadCloser{reader: NewReader("")}
	r := ReadAll(reader)

	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value)
	AssertTrue(t, reader.closed)
}

func TestIo_WriteString_Good(t *T) {
	dst := NewBuffer()
	r := WriteString(dst, "agent\n")

	AssertTrue(t, r.OK)
	AssertEqual(t, 6, r.Value.(int))
	AssertEqual(t, "agent\n", dst.String())
}

func TestIo_WriteString_Bad(t *T) {
	r := WriteString(ioFailingWriter{}, "agent")

	AssertFalse(t, r.OK)
	AssertEqual(t, AnError, r.Value)
}

func TestIo_WriteString_Ugly(t *T) {
	dst := NewBuffer()
	r := WriteString(dst, "")

	AssertTrue(t, r.OK)
	AssertEqual(t, 0, r.Value.(int))
	AssertEqual(t, "", dst.String())
}

func TestIo_Reader_Good_AcceptsCoreReader(t *T) {
	var r Reader = NewReader("ok")
	buf := make([]byte, 2)
	n, err := r.Read(buf)

	AssertNoError(t, err)
	AssertEqual(t, 2, n)
	AssertEqual(t, "ok", string(buf))
}

func TestIo_Writer_Good_AcceptsCoreBuffer(t *T) {
	var w Writer = NewBuffer()
	n, err := w.Write([]byte("hi"))

	AssertNoError(t, err)
	AssertEqual(t, 2, n)
}

// --- Bytes ---

func TestIo_NewBuffer_Good(t *T) {
	buf := NewBuffer([]byte("hello"))

	AssertEqual(t, "hello", buf.String())
	AssertEqual(t, 5, buf.Len())
}

func TestIo_NewBuffer_Bad(t *T) {
	buf := NewBuffer()

	AssertNotNil(t, buf)
	AssertEqual(t, 0, buf.Len())
	AssertNoError(t, buf.WriteByte('x'))
	AssertEqual(t, "x", buf.String())
}

func TestIo_NewBuffer_Ugly(t *T) {
	src := []byte("abc")
	buf := NewBuffer(src)

	src[0] = 'z'

	AssertEqual(t, "zbc", buf.String())
}

func TestIo_NewBufferString_Good(t *T) {
	buf := NewBufferString("hello")

	AssertEqual(t, "hello", buf.String())
	AssertEqual(t, 5, buf.Len())
}

func TestIo_NewBufferString_Bad(t *T) {
	buf := NewBufferString("")

	AssertNotNil(t, buf)
	AssertEqual(t, 0, buf.Len())
	AssertNoError(t, buf.WriteByte('x'))
	AssertEqual(t, "x", buf.String())
}

func TestIo_NewBufferString_Ugly(t *T) {
	buf := NewBufferString("a\x00b")

	AssertEqual(t, []byte{'a', 0, 'b'}, buf.Bytes())
	AssertEqual(t, 3, buf.Len())
}
