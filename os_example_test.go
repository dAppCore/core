package core_test

import . "dappco.re/go"

// ExampleFileMode uses file mode aliases through `FileMode` for process IO access. File
// modes and standard streams are exposed through core aliases.
func ExampleFileMode() {
	var mode FileMode = 0644
	Println((mode & ModePerm) == 0644)
	// Output: true
}

// ExampleModeDir checks directory mode aliases through `ModeDir` for process IO access.
// File modes and standard streams are exposed through core aliases.
func ExampleModeDir() {
	Println(ModeDir.IsDir())
	// Output: true
}

// ExampleStdin reads standard input through `Stdin` for process IO access. File modes and
// standard streams are exposed through core aliases.
func ExampleStdin() {
	Println(Stdin() != nil)
	// Output: true
}

// ExampleStdout reads standard output through `Stdout` for process IO access. File modes
// and standard streams are exposed through core aliases.
func ExampleStdout() {
	Println(Stdout() != nil)
	// Output: true
}

// ExampleStderr reads standard error through `Stderr` for process IO access. File modes
// and standard streams are exposed through core aliases.
func ExampleStderr() {
	Println(Stderr() != nil)
	// Output: true
}
