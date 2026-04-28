// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"bytes"
	"errors"
	"strings"

	. "dappco.re/go/core"
)

func TestIo_EOF_Good(t *T) {
	// EOF sentinel must satisfy errors.Is for canonical end-of-stream check.
	wrapped := errors.Join(EOF)
	AssertTrue(t, errors.Is(wrapped, EOF))
}

func TestIo_Copy_Good(t *T) {
	src := strings.NewReader("hello world")
	var dst bytes.Buffer
	r := Copy(&dst, src)
	AssertTrue(t, r.OK)
	AssertEqual(t, int64(11), r.Value.(int64))
	AssertEqual(t, "hello world", dst.String())
}

func TestIo_CopyN_Good(t *T) {
	src := strings.NewReader("hello world")
	var dst bytes.Buffer
	r := CopyN(&dst, src, 5)
	AssertTrue(t, r.OK)
	AssertEqual(t, int64(5), r.Value.(int64))
	AssertEqual(t, "hello", dst.String())
}

func TestIo_CopyN_Bad_ShortSource(t *T) {
	src := strings.NewReader("abc")
	var dst bytes.Buffer
	r := CopyN(&dst, src, 100)
	AssertFalse(t, r.OK)
}

func TestIo_WriteString_Good(t *T) {
	var dst bytes.Buffer
	r := WriteString(&dst, "hello\n")
	AssertTrue(t, r.OK)
	AssertEqual(t, 6, r.Value.(int))
	AssertEqual(t, "hello\n", dst.String())
}

func TestIo_Reader_Good_AcceptsBytesBuffer(t *T) {
	// Type alias must accept any io.Reader implementation as core.Reader.
	var r Reader = strings.NewReader("ok")
	buf := make([]byte, 2)
	n, _ := r.Read(buf)
	AssertEqual(t, 2, n)
	AssertEqual(t, "ok", string(buf))
}

func TestIo_Writer_Good_AcceptsBytesBuffer(t *T) {
	var w Writer = &bytes.Buffer{}
	n, _ := w.Write([]byte("hi"))
	AssertEqual(t, 2, n)
}
