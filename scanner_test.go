package core_test

import (
	"bufio"
	"strings"
	"testing"

	. "dappco.re/go/core"
)

// --- Scanner ---

func TestScanner_NewLineScanner_Good(t *testing.T) {
	longLine := strings.Repeat("x", bufio.MaxScanTokenSize+1)
	scanner := NewLineScanner(strings.NewReader(longLine + "\nnext"))

	RequireTrue(t, scanner.Scan())
	AssertEqual(t, longLine, scanner.Text())

	RequireTrue(t, scanner.Scan())
	AssertEqual(t, "next", scanner.Text())

	AssertFalse(t, scanner.Scan())
	AssertNoError(t, scanner.Err())
}

func TestScanner_NewLineScanner_Bad(t *testing.T) {
	scanner := NewLineScannerWithSize(strings.NewReader("too long\n"), 4)

	AssertFalse(t, scanner.Scan())
	AssertError(t, scanner.Err())
}

func TestScanner_NewLineScanner_Ugly(t *testing.T) {
	scanner := NewLineScanner(strings.NewReader("first\r\nsecond\n\nthird"))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	AssertEqual(t, []string{"first", "second", "", "third"}, lines)
	AssertNoError(t, scanner.Err())
}
