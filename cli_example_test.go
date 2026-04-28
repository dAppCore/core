package core_test

import . "dappco.re/go"

// ExampleCliRegister registers CLI commands through `CliRegister` for a dAppCore
// command-line tool. Command output, banners, and help text are routed through the CLI
// abstraction.
func ExampleCliRegister() {
	_ = CliRegister
}

// ExampleCli_Print writes text through `Cli.Print` for a dAppCore command-line tool.
// Command output, banners, and help text are routed through the CLI abstraction.
func ExampleCli_Print() {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)

	c.Cli().Print("hello %s", "codex")
	Println(TrimSuffix(buf.String(), "\n"))
	// Output:
	// hello codex
}

// ExampleCli_SetOutput redirects output through `Cli.SetOutput` for a dAppCore
// command-line tool. Command output, banners, and help text are routed through the CLI
// abstraction.
func ExampleCli_SetOutput() {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().Print("ready")

	Println(Contains(buf.String(), "ready"))
	// Output: true
}

// ExampleCli_Run runs `Cli.Run` with representative caller inputs for a dAppCore
// command-line tool. Command output, banners, and help text are routed through the CLI
// abstraction.
func ExampleCli_Run() {
	c := New()
	c.Command("deploy", Command{
		Action: func(opts Options) Result {
			return Result{Value: opts.String("target"), OK: true}
		},
	})

	r := c.Cli().Run("deploy", "--target=homelab")
	Println(r.Value)
	// Output: homelab
}

// ExampleCli_PrintHelp prints help text through `Cli.PrintHelp` for a dAppCore
// command-line tool. Command output, banners, and help text are routed through the CLI
// abstraction.
func ExampleCli_PrintHelp() {
	c := New(WithOption("name", "ops"))
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	c.Cli().PrintHelp()
	Println(Contains(buf.String(), "ops commands:"))
	Println(Contains(buf.String(), "deploy"))
	// Output:
	// true
	// true
}

// ExampleCli_SetBanner sets a banner through `Cli.SetBanner` for a dAppCore command-line
// tool. Command output, banners, and help text are routed through the CLI abstraction.
func ExampleCli_SetBanner() {
	c := New()
	c.Cli().SetBanner(func(_ *Cli) string { return "ops v1" })
	Println(c.Cli().Banner())
	// Output: ops v1
}

// ExampleCli_Banner reads a banner through `Cli.Banner` for a dAppCore command-line tool.
// Command output, banners, and help text are routed through the CLI abstraction.
func ExampleCli_Banner() {
	c := New(WithOption("name", "ops"))
	Println(c.Cli().Banner())
	// Output: ops
}

// ExampleCliOptions declares CLI options through `CliOptions` for a dAppCore command-line
// tool. Command output, banners, and help text are routed through the CLI abstraction.
func ExampleCliOptions() {
	opts := CliOptions{}
	Println(Sprint(opts))
	// Output: {}
}
