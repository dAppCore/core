package core_test

import (

	. "dappco.re/go/core"
)

func ExampleCore_Command_register() {
	c := New()
	c.Command("deploy/to/homelab", Command{
		Description: "Deploy to homelab",
		Action: func(opts Options) Result {
			return Result{Value: "deployed", OK: true}
		},
	})

	Println(c.Command("deploy/to/homelab").OK)
	// Output: true
}

func ExampleCore_Command_managed() {
	c := New()
	c.Command("serve", Command{
		Action:  func(_ Options) Result { return Result{OK: true} },
		Managed: "process.daemon",
	})

	cmd := c.Command("serve").Value.(*Command)
	Println(cmd.IsManaged())
	// Output: true
}

func ExampleCore_Commands() {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("test", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	Println(c.Commands())
	// Output: [deploy test]
}
