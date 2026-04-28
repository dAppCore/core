// SPDX-License-Identifier: EUPL-1.2

// Base I/O primitives for the Core framework.
//
// Re-exports stdlib io interfaces and the EOF sentinel as core types so
// consumer packages declare Reader/Writer parameters via core without
// importing io directly. Bundle size is unaffected by aliases — they
// resolve to the same underlying types — but the AX-6 import-line cost
// drops to zero for typed-parameter use.
//
// Usage:
//
//	func handle(r core.Reader, w core.Writer) error { ... }
//
//	if errors.Is(err, core.EOF) { ... }
//
//	n := core.Copy(dst, src)
//	if !n.OK { return n }
package core

import "io"

// Reader is the canonical io.Reader interface, exported as core.Reader.
type Reader = io.Reader

// Writer is the canonical io.Writer interface, exported as core.Writer.
type Writer = io.Writer

// Closer is the canonical io.Closer interface.
type Closer = io.Closer

// ReadCloser composes Reader and Closer.
type ReadCloser = io.ReadCloser

// WriteCloser composes Writer and Closer.
type WriteCloser = io.WriteCloser

// ReadWriter composes Reader and Writer.
type ReadWriter = io.ReadWriter

// ReadWriteCloser composes Reader, Writer, and Closer.
type ReadWriteCloser = io.ReadWriteCloser

// Seeker is the canonical io.Seeker interface.
type Seeker = io.Seeker

// EOF is the canonical end-of-stream sentinel error.
//
//	if errors.Is(err, core.EOF) { /* end of stream */ }
var EOF = io.EOF

// Copy copies from src to dst until EOF on src or an error occurs.
// Returns Result wrapping the number of bytes copied (int64).
//
//	r := core.Copy(dst, src)
//	if !r.OK { return r }
//	n := r.Value.(int64)
func Copy(dst Writer, src Reader) Result {
	n, err := io.Copy(dst, src)
	if err != nil {
		return Result{err, false}
	}
	return Result{n, true}
}

// CopyN copies n bytes (or until an error) from src to dst.
//
//	r := core.CopyN(dst, src, 1024)
func CopyN(dst Writer, src Reader, n int64) Result {
	written, err := io.CopyN(dst, src, n)
	if err != nil {
		return Result{err, false}
	}
	return Result{written, true}
}

// WriteString writes the contents of s to w. Returns Result wrapping
// the number of bytes written (int).
//
//	r := core.WriteString(stdout, "hello\n")
func WriteString(w Writer, s string) Result {
	n, err := io.WriteString(w, s)
	if err != nil {
		return Result{err, false}
	}
	return Result{n, true}
}
