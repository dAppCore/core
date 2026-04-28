// SPDX-License-Identifier: EUPL-1.2

// Scanner and buffered-reader operations for the Core framework.
// Provides bufio.Scanner constructors for line-oriented reads with a
// larger line cap than the standard library default, plus the
// BufReader type alias for byte-stream reads with delimiter support
// (used by lsp.go's JSON-RPC frame parser).

package core

import "bufio"

// BufReader is the Core alias for bufio.Reader — buffered byte-stream
// reads with delimiter support (ReadString, ReadBytes, ReadLine).
//
//	r := core.NewBufReader(core.Stdin())
//	line, err := r.ReadString('\n')
type BufReader = bufio.Reader

// NewBufReader wraps r in a Core BufReader for delimiter-based reads.
//
//	r := core.NewBufReader(core.Stdin())
//	line, err := r.ReadString('\n')
func NewBufReader(r Reader) *BufReader {
	return bufio.NewReader(r)
}

const defaultLineScannerMaxSize = 1024 * 1024

// NewLineScanner returns a bufio.Scanner configured for line-oriented reads.
//
//	scanner := core.NewLineScanner(r)
func NewLineScanner(r Reader) *bufio.Scanner {
	return NewLineScannerWithSize(r, defaultLineScannerMaxSize)
}

// NewLineScannerWithSize returns a bufio.Scanner configured for line-oriented
// reads with a custom maximum line size.
//
//	scanner := core.NewLineScannerWithSize(r, 2*1024*1024)
func NewLineScannerWithSize(r Reader, max int) *bufio.Scanner {
	if max <= 0 {
		max = defaultLineScannerMaxSize
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, 0, min(max, bufio.MaxScanTokenSize)), max)
	return scanner
}
