package core_test

import . "dappco.re/go"

// ExampleCore_Query runs or declares a query through `Core.Query` for Core message
// dispatch. Actions and queries share one Core dispatch pattern for messages.
func ExampleCore_Query() {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		if q == "status" {
			return Result{Value: "ready", OK: true}
		}
		return Result{}
	})

	Println(c.Query("status").Value)
	// Output: ready
}

// ExampleCore_QueryAll runs all matching queries through `Core.QueryAll` for Core message
// dispatch. Actions and queries share one Core dispatch pattern for messages.
func ExampleCore_QueryAll() {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "api", OK: true} })
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "worker", OK: true} })

	Println(c.QueryAll("services").Value)
	// Output: [api worker]
}

// ExampleCore_RegisterQuery registers a query through `Core.RegisterQuery` for Core
// message dispatch. Actions and queries share one Core dispatch pattern for messages.
func ExampleCore_RegisterQuery() {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "registered", OK: true} })
	Println(c.Query("anything").Value)
	// Output: registered
}

// ExampleCore_RegisterAction registers an action through `Core.RegisterAction` for Core
// message dispatch. Actions and queries share one Core dispatch pattern for messages.
func ExampleCore_RegisterAction() {
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

// ExampleCore_RegisterActions registers a group of actions through `Core.RegisterActions`
// for Core message dispatch. Actions and queries share one Core dispatch pattern for
// messages.
func ExampleCore_RegisterActions() {
	c := New()
	var seen []string
	c.RegisterActions(
		func(_ *Core, msg Message) Result {
			seen = append(seen, Concat("a:", msg.(string)))
			return Result{OK: true}
		},
		func(_ *Core, msg Message) Result {
			seen = append(seen, Concat("b:", msg.(string)))
			return Result{OK: true}
		},
	)

	c.ACTION("started")
	Println(seen)
	// Output: [a:started b:started]
}

// ExampleIpc declares IPC registration through `Ipc` for Core message dispatch. Actions
// and queries share one Core dispatch pattern for messages.
func ExampleIpc() {
	c := New()
	Println(c.IPC() != nil)
	// Output: true
}
