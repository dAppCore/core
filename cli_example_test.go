package core_test

import (
	. "dappco.re/go"
)

// ExampleAssertCLI demonstrates the canonical CLI test shape: build a
// Core, register a process.run handler (typically dappco.re/go-process
// in production code, an in-test fake here), then dispatch a CLITest
// case that asserts on stdout substring and OK status.
func ExampleAssertCLI() {
	c := New()
	c.Action("process.run", func(ctx Context, opts Options) Result {
		return Result{Value: "go version go1.26.0\n", OK: true}
	})

	tc := CLITest{
		Cmd:      "go",
		Args:     []string{"version"},
		WantOK:   true,
		Contains: "go1.26",
	}
	r := c.Process().Run(Background(), tc.Cmd, tc.Args...)
	Println(r.OK)
	Println(Contains(r.Value.(string), tc.Contains))
	// Output:
	// true
	// true
}
