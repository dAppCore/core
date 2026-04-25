// SPDX-License-Identifier: EUPL-1.2

// Scanner operations for the Core framework.
// Provides bufio.Scanner constructors for line-oriented reads with a larger
// line cap than the standard library default.

package core

import (
	"bufio"
	"io"
)

const defaultLineScannerMaxSize = 1024 * 1024

// NewLineScanner returns a bufio.Scanner configured for line-oriented reads.
//
//	scanner := core.NewLineScanner(r)
func NewLineScanner(r io.Reader) *bufio.Scanner {
	return NewLineScannerWithSize(r, defaultLineScannerMaxSize)
}

// NewLineScannerWithSize returns a bufio.Scanner configured for line-oriented
// reads with a custom maximum line size.
//
//	scanner := core.NewLineScannerWithSize(r, 2*1024*1024)
func NewLineScannerWithSize(r io.Reader, max int) *bufio.Scanner {
	if max <= 0 {
		max = defaultLineScannerMaxSize
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, 0, min(max, bufio.MaxScanTokenSize)), max)
	return scanner
}
