package core_test

import . "dappco.re/go"

func ExampleNewLineScanner() {
	scanner := NewLineScanner(NewReader("alpha\nbravo\n"))
	for scanner.Scan() {
		Println(scanner.Text())
	}
	// Output:
	// alpha
	// bravo
}

func ExampleNewLineScannerWithSize() {
	scanner := NewLineScannerWithSize(NewReader("alpha\n"), 1024)
	Println(scanner.Scan())
	Println(scanner.Text())
	// Output:
	// true
	// alpha
}
