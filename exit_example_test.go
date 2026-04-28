package core_test

import (
	. "dappco.re/go"
)

// ExampleCore_Exit documents the normal Core exit call without terminating the example
// process. Production exit paths are documented without terminating the example test
// process.
func ExampleCore_Exit() {
	c := New()
	_ = c // c.Exit(0) terminates the process in production
	Println("ready to exit")
	// Output:
	// ready to exit
}

// ExampleCore_ExitWith configures exit code and shutdown timing without terminating the
// example process. Production exit paths are documented without terminating the example
// test process.
func ExampleCore_ExitWith() {
	c := New()
	_ = c // c.ExitWith(core.ExitOptions{Code: 0})
	_ = ExitOptions{Code: 0}
	Println("configured")
	// Output:
	// configured
}

// ExampleCore_ExitNow documents immediate termination without running it during the
// example process. Production exit paths are documented without terminating the example
// test process.
func ExampleCore_ExitNow() {
	c := New()
	_ = c // c.ExitNow(2) terminates without running ServiceShutdown
	Println("emergency")
	// Output:
	// emergency
}

// ExampleExit documents the package-level exit helper for call sites without a Core value.
// Production exit paths are documented without terminating the example test process.
func ExampleExit() {
	// core.Exit(1) terminates the process in production
	Println("package-level")
	// Output:
	// package-level
}
