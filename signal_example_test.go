package core_test

import (
	"context"

	. "dappco.re/go"
)

// ExampleCore_Signal_exists shows the registration check.
func ExampleCore_Signal_exists() {
	c := New()
	if c.Signal().Exists() {
		Println("signal handling available")
	} else {
		Println("no signal service registered")
	}
	// Output:
	// no signal service registered
}

// ExampleCore_Signal_subscribe shows the action-subscription pattern. In
// production the go-process service registers signal.received and broadcasts
// on each OS signal; here we register a stub action to demonstrate the surface.
func ExampleCore_Signal_subscribe() {
	c := New()
	c.Action("signal.received", func(_ context.Context, opts Options) Result {
		Println("got", opts.String("name"))
		return Result{OK: true}
	})
	// In production this fires on SIGINT/SIGTERM/SIGHUP:
	c.Action("signal.received").Run(c.Context(),
		NewOptions(Option{Key: "name", Value: "SIGINT"}))
	// Output:
	// got SIGINT
}
