package core_test

import . "dappco.re/go"

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

func ExampleCore_QueryAll() {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "api", OK: true} })
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "worker", OK: true} })

	Println(c.QueryAll("services").Value)
	// Output: [api worker]
}

func ExampleCore_RegisterQuery() {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "registered", OK: true} })
	Println(c.Query("anything").Value)
	// Output: registered
}

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

func ExampleIpc() {
	c := New()
	Println(c.IPC() != nil)
	// Output: true
}
