package core_test

import (
	. "dappco.re/go"
)

func ExampleCommand_I18nKey() {
	cmd := &Command{Path: "deploy/to/homelab"}
	Println(cmd.I18nKey())
	// Output: cmd.deploy.to.homelab.description
}

func ExampleCommand_Run() {
	cmd := &Command{
		Action: func(opts Options) Result {
			return Result{Value: opts.String("target"), OK: true}
		},
	}
	r := cmd.Run(NewOptions(Option{Key: "target", Value: "homelab"}))
	Println(r.Value)
	// Output: homelab
}

func ExampleCommand_IsManaged() {
	cmd := &Command{Managed: "process.daemon"}
	Println(cmd.IsManaged())
	// Output: true
}

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

func ExampleCore_Command_get() {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	r := c.Command("deploy")
	cmd := r.Value.(*Command)
	Println(cmd.Name)
	Println(cmd.Path)
	// Output:
	// deploy
	// deploy
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
