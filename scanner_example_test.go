package core_test

import . "dappco.re/go"

// ExampleNewLineScanner scans lines through `NewLineScanner` for line-oriented logs. Line
// scanning stays on the core wrapper surface, including larger buffer cases.
func ExampleNewLineScanner() {
	scanner := NewLineScanner(NewReader("alpha\nbravo\n"))
	for scanner.Scan() {
		Println(scanner.Text())
	}
	// Output:
	// alpha
	// bravo
}

// ExampleNewLineScannerWithSize scans long lines through `NewLineScannerWithSize` for
// line-oriented logs. Line scanning stays on the core wrapper surface, including larger
// buffer cases.
func ExampleNewLineScannerWithSize() {
	scanner := NewLineScannerWithSize(NewReader("alpha\n"), 1024)
	Println(scanner.Scan())
	Println(scanner.Text())
	// Output:
	// true
	// alpha
}
