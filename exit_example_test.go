package core_test

import (
	. "dappco.re/go"
)

// ExampleCore_Exit demonstrates the call shape. The osExit hook is overridden
// in tests; in production this would terminate the process.
func ExampleCore_Exit() {
	c := New()
	_ = c // c.Exit(0) terminates the process in production
	Println("ready to exit")
	// Output:
	// ready to exit
}

// ExampleCore_ExitWith demonstrates a custom shutdown timeout.
func ExampleCore_ExitWith() {
	c := New()
	_ = c // c.ExitWith(core.ExitOptions{Code: 0})
	_ = ExitOptions{Code: 0}
	Println("configured")
	// Output:
	// configured
}

// ExampleCore_ExitNow demonstrates the immediate-termination escape hatch.
func ExampleCore_ExitNow() {
	c := New()
	_ = c // c.ExitNow(2) terminates without running ServiceShutdown
	Println("emergency")
	// Output:
	// emergency
}

// ExampleExit demonstrates the package-level form for callsites without a *Core.
func ExampleExit() {
	// core.Exit(1) terminates the process in production
	Println("package-level")
	// Output:
	// package-level
}
