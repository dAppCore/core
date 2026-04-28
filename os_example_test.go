package core_test

import . "dappco.re/go"

func ExampleFileMode() {
	var mode FileMode = 0644
	Println((mode & ModePerm) == 0644)
	// Output: true
}

func ExampleModeDir() {
	Println(ModeDir.IsDir())
	// Output: true
}

func ExampleStdin() {
	Println(Stdin() != nil)
	// Output: true
}

func ExampleStdout() {
	Println(Stdout() != nil)
	// Output: true
}

func ExampleStderr() {
	Println(Stderr() != nil)
	// Output: true
}
