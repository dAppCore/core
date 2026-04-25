// SPDX-License-Identifier: EUPL-1.2

// Byte buffer operations for the Core framework.
// Provides bytes.Buffer constructors so downstream packages can avoid
// importing "bytes" directly for common buffer creation.

package core

import "bytes"

// NewBuffer returns a bytes.Buffer initialised with b.
// With no input, it returns an empty bytes.Buffer.
//
//	buf := core.NewBuffer([]byte("hello"))
//	empty := core.NewBuffer()
func NewBuffer(b ...[]byte) *bytes.Buffer {
	if len(b) == 0 {
		return &bytes.Buffer{}
	}
	return bytes.NewBuffer(b[0])
}

// NewBufferString returns a bytes.Buffer initialised with s.
//
//	buf := core.NewBufferString("hello")
func NewBufferString(s string) *bytes.Buffer {
	return bytes.NewBufferString(s)
}
