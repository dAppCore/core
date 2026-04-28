package core_test

import . "dappco.re/go"

// ExampleCore_Signal_exists checks whether signal handling has been registered in Core. OS
// signal integration is represented as action-backed Core behaviour.
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

// ExampleCore_Signal accesses signal handling through `Core.Signal` for process signal
// handling. OS signal integration is represented as action-backed Core behaviour.
func ExampleCore_Signal() {
	c := New()
	Println(c.Signal() != nil)
	// Output: true
}

// ExampleSignal_Stop stops a service through `Signal.Stop` for process signal handling. OS
// signal integration is represented as action-backed Core behaviour.
func ExampleSignal_Stop() {
	c := New()
	c.Action("signal.stop", func(_ Context, _ Options) Result {
		return Result{Value: "stopped", OK: true}
	})
	r := c.Signal().Stop()
	Println(r.Value)
	// Output: stopped
}

// ExampleSignal_Exists checks whether the signal action is registered before and after
// installation. OS signal integration is represented as action-backed Core behaviour.
func ExampleSignal_Exists() {
	c := New()
	Println(c.Signal().Exists())
	c.Action("signal.received", func(_ Context, _ Options) Result { return Result{OK: true} })
	Println(c.Signal().Exists())
	// Output:
	// false
	// true
}

// ExampleCore_Signal_subscribe registers the action that a process signal service invokes
// on receipt. OS signal integration is represented as action-backed Core behaviour.
func ExampleCore_Signal_subscribe() {
	c := New()
	c.Action("signal.received", func(_ Context, opts Options) Result {
		Println("got", opts.String("name"))
		return Result{OK: true}
	})
	// In production this fires on SIGINT/SIGTERM/SIGHUP:
	c.Action("signal.received").Run(c.Context(),
		NewOptions(Option{Key: "name", Value: "SIGINT"}))
	// Output:
	// got SIGINT
}
