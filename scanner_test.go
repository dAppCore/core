package core_test

import (
	. "dappco.re/go"
)

func TestScanner_NewLineScanner_Good(t *T) {
	longLine := scannerRepeat("x", 70*1024)
	scanner := NewLineScanner(NewReader(Concat(longLine, "\nnext")))

	RequireTrue(t, scanner.Scan())
	AssertEqual(t, longLine, scanner.Text())
	RequireTrue(t, scanner.Scan())
	AssertEqual(t, "next", scanner.Text())
	AssertFalse(t, scanner.Scan())
	AssertNoError(t, scanner.Err())
}

func TestScanner_NewLineScanner_Bad(t *T) {
	scanner := NewLineScannerWithSize(NewReader("too long\n"), 4)

	AssertFalse(t, scanner.Scan())
	AssertError(t, scanner.Err())
}

func TestScanner_NewLineScanner_Ugly(t *T) {
	scanner := NewLineScanner(NewReader("first\r\nsecond\n\nthird"))
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	AssertEqual(t, []string{"first", "second", "", "third"}, lines)
	AssertNoError(t, scanner.Err())
}

func TestScanner_NewLineScannerWithSize_Good(t *T) {
	scanner := NewLineScannerWithSize(NewReader("agent\nready"), 64)

	RequireTrue(t, scanner.Scan())
	AssertEqual(t, "agent", scanner.Text())
	RequireTrue(t, scanner.Scan())
	AssertEqual(t, "ready", scanner.Text())
	AssertFalse(t, scanner.Scan())
	AssertNoError(t, scanner.Err())
}

func TestScanner_NewLineScannerWithSize_Bad(t *T) {
	scanner := NewLineScannerWithSize(NewReader("agent\n"), 4)

	AssertFalse(t, scanner.Scan())
	AssertError(t, scanner.Err())
}

func TestScanner_NewLineScannerWithSize_Ugly(t *T) {
	scanner := NewLineScannerWithSize(NewReader("agent"), 0)

	RequireTrue(t, scanner.Scan())
	AssertEqual(t, "agent", scanner.Text())
	AssertFalse(t, scanner.Scan())
	AssertNoError(t, scanner.Err())
}

func scannerRepeat(s string, count int) string {
	b := NewBuilder()
	for i := 0; i < count; i++ {
		b.WriteString(s)
	}
	return b.String()
}
