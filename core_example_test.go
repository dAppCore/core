package core_test

import (
	"context"

	. "dappco.re/go"
)

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

func ExampleCore_Env() {
	c := New()
	Println(c.Env("OS") != "")
	// Output: true
}

func ExampleCore_Context() {
	c := New()
	Println(c.Context() != nil)
	// Output: true
}

func ExampleCore_Core() {
	c := New()
	Println(c.Core() == c)
	// Output: true
}

func ExampleCore_RunE() {
	c := New()
	Println(c.RunE() == nil)
	// Output: true
}

func ExampleCore_Run() {
	_ = New().Run
}

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

func ExampleCore_QUERY() {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{Value: Concat("query:", q.(string)), OK: true}
	})

	r := c.QUERY("status")
	Println(r.Value)
	// Output: query:status
}

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

func ExampleCore_LogError() {
	c := New()
	r := c.LogError(nil, "example", "nothing to log")
	Println(r.OK)
	// Output: true
}

func ExampleCore_LogWarn() {
	c := New()
	r := c.LogWarn(nil, "example", "nothing to warn")
	Println(r.OK)
	// Output: true
}

func ExampleCore_Must() {
	c := New()
	c.Must(nil, "example", "no panic")
	Println("ok")
	// Output: ok
}

func ExampleCore_RegistryOf() {
	c := New()
	c.Action("deploy", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.RegistryOf("actions").Names())
	Println(c.RegistryOf("missing").Len())
	// Output:
	// [deploy]
	// 0
}
