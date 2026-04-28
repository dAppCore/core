package core_test

import . "dappco.re/go"

// ExampleCore_accessors reads the grouped accessor methods through `Core` for Core
// orchestration. Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_accessors() {
	c := New(WithOption("name", "ops"))

	Println(c.Options().String("name"))
	Println(c.App().Name)
	Println(c.Data() != nil)
	Println(c.Drive() != nil)
	Println(c.Fs() != nil)
	Println(c.Config() != nil)
	Println(c.Error() != nil)
	Println(c.Log() != nil)
	Println(c.Cli() != nil)
	Println(c.IPC() != nil)
	Println(c.I18n() != nil)
	// Output:
	// ops
	// ops
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// true
}

// ExampleCore_Env reads environment access through `Core.Env` for Core orchestration. Core
// keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_Env() {
	c := New()
	Println(c.Env("OS") != "")
	// Output: true
}

// ExampleCore_Context reads the active context through `Core.Context` for Core
// orchestration. Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_Context() {
	c := New()
	Println(c.Context() != nil)
	// Output: true
}

// ExampleCore_Core returns the Core instance through `Core.Core` for Core orchestration.
// Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_Core() {
	c := New()
	Println(c.Core() == c)
	// Output: true
}

// ExampleCore_RunResult runs `Core.RunResult` through the Result-returning startup path
// for Core orchestration. Core keeps orchestration helpers reachable from one
// predictable facade.
func ExampleCore_RunResult() {
	c := New()
	Println(c.RunResult().OK)
	// Output: true
}

// ExampleCore_Run runs `Core.Run` with representative caller inputs for Core
// orchestration. Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_Run() {
	_ = New().Run
}

// ExampleCore_ACTION calls the uppercase action helper through `Core.ACTION` for Core
// orchestration. Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_ACTION() {
	c := New()
	seen := ""
	c.RegisterAction(func(_ *Core, msg Message) Result {
		seen = msg.(string)
		return Result{OK: true}
	})

	c.ACTION("started")
	Println(seen)
	// Output: started
}

// ExampleCore_QUERY calls the uppercase query helper through `Core.QUERY` for Core
// orchestration. Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_QUERY() {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{Value: Concat("query:", q.(string)), OK: true}
	})

	r := c.QUERY("status")
	Println(r.Value)
	// Output: query:status
}

// ExampleCore_QUERYALL calls every matching uppercase query through `Core.QUERYALL` for
// Core orchestration. Core keeps orchestration helpers reachable from one predictable
// facade.
func ExampleCore_QUERYALL() {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{Value: Concat("a:", q.(string)), OK: true}
	})
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{Value: Concat("b:", q.(string)), OK: true}
	})

	r := c.QUERYALL("status")
	Println(r.Value)
	// Output: [a:status b:status]
}

// ExampleCore_LogError logs an error through `Core.LogError` for Core orchestration. Core
// keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_LogError() {
	c := New()
	r := c.LogError(nil, "example", "nothing to log")
	Println(r.OK)
	// Output: true
}

// ExampleCore_LogWarn logs a warning through `Core.LogWarn` for Core orchestration. Core
// keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_LogWarn() {
	c := New()
	r := c.LogWarn(nil, "example", "nothing to warn")
	Println(r.OK)
	// Output: true
}

// ExampleCore_Must unwraps a successful Result through `Core.Must` for Core orchestration.
// Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_Must() {
	c := New()
	c.Must(nil, "example", "no panic")
	Println("ok")
	// Output: ok
}

// ExampleCore_RegistryOf retrieves a named registry through `Core.RegistryOf` for Core
// orchestration. Core keeps orchestration helpers reachable from one predictable facade.
func ExampleCore_RegistryOf() {
	c := New()
	c.Action("deploy", func(_ Context, _ Options) Result { return Result{OK: true} })
	Println(c.RegistryOf("actions").Names())
	Println(c.RegistryOf("missing").Len())
	// Output:
	// [deploy]
	// 0
}
