package core_test

import . "dappco.re/go"

func ExampleCliRegister() {
	_ = CliRegister
}

func ExampleCli_Print() {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)

	c.Cli().Print("hello %s", "codex")
	Println(TrimSuffix(buf.String(), "\n"))
	// Output:
	// hello codex
}

func ExampleCli_SetOutput() {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().Print("ready")

	Println(Contains(buf.String(), "ready"))
	// Output: true
}

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

func ExampleCli_SetBanner() {
	c := New()
	c.Cli().SetBanner(func(_ *Cli) string { return "ops v1" })
	Println(c.Cli().Banner())
	// Output: ops v1
}

func ExampleCli_Banner() {
	c := New(WithOption("name", "ops"))
	Println(c.Cli().Banner())
	// Output: ops
}

func ExampleCliOptions() {
	opts := CliOptions{}
	Println(Sprint(opts))
	// Output: {}
}
