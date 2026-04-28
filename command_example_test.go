package core_test

import (
	. "dappco.re/go"
)

// ExampleCommand_I18nKey builds an i18n key through `Command.I18nKey` for managed CLI
// command dispatch. Command registration and lookup use managed metadata before execution.
func ExampleCommand_I18nKey() {
	cmd := &Command{Path: "deploy/to/homelab"}
	Println(cmd.I18nKey())
	// Output: cmd.deploy.to.homelab.description
}

// ExampleCommand_Run runs `Command.Run` with representative caller inputs for managed CLI
// command dispatch. Command registration and lookup use managed metadata before execution.
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

// ExampleCommand_IsManaged checks managed command state through `Command.IsManaged` for
// managed CLI command dispatch. Command registration and lookup use managed metadata
// before execution.
func ExampleCommand_IsManaged() {
	cmd := &Command{Managed: "process.daemon"}
	Println(cmd.IsManaged())
	// Output: true
}

// ExampleCore_Command_register registers a value through `Core.Command` for managed CLI
// command dispatch. Command registration and lookup use managed metadata before execution.
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

// ExampleCore_Command_get retrieves a value through `Core.Command` for managed CLI command
// dispatch. Command registration and lookup use managed metadata before execution.
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

// ExampleCore_Command_managed registers managed command metadata through `Core.Command`
// for managed CLI command dispatch. Command registration and lookup use managed metadata
// before execution.
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

// ExampleCore_Commands lists command names through `Core.Commands` for managed CLI command
// dispatch. Command registration and lookup use managed metadata before execution.
func ExampleCore_Commands() {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("test", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	Println(c.Commands())
	// Output: [deploy test]
}
