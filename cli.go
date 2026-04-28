// SPDX-License-Identifier: EUPL-1.2

// Cli is the CLI surface layer for the Core command tree.
//
//	c := core.New(core.WithOption("name", "myapp")).Value.(*Core)
//	c.Command("deploy", core.Command{Action: handler})
//	c.Cli().Run()
package core

// CliOptions holds configuration for the Cli service.
//
//	c := core.New()
//	runtime := core.NewServiceRuntime(c, core.CliOptions{})
//	_ = runtime.Options()
type CliOptions struct{}

// Cli is the CLI surface for the Core command tree.
//
//	c := core.New(core.WithOption("name", "homelab"))
//	cli := c.Cli()
//	cli.Print("%s ready", c.App().Name)
type Cli struct {
	*ServiceRuntime[CliOptions]
	output Writer
	banner func(*Cli) string
}

// Register creates a Cli service factory for core.WithService.
//
//	core.New(core.WithService(core.CliRegister))
func CliRegister(c *Core) Result {
	cl := &Cli{output: Stdout()}
	cl.ServiceRuntime = NewServiceRuntime[CliOptions](c, CliOptions{})
	return c.RegisterService("cli", cl)
}

// Print writes to the CLI output (defaults to core.Stdout()).
//
//	c.Cli().Print("hello %s", "world")
func (cl *Cli) Print(format string, args ...any) {
	Print(cl.output, format, args...)
}

// SetOutput sets the CLI output writer.
//
//	c.Cli().SetOutput(core.Stderr())
func (cl *Cli) SetOutput(w Writer) {
	cl.output = w
}

// Run resolves core.Args() to a command path and executes it.
//
//	c.Cli().Run()
//	c.Cli().Run("deploy", "to", "homelab")
func (cl *Cli) Run(args ...string) Result {
	if len(args) == 0 {
		args = Args()[1:]
	}

	clean := FilterArgs(args)
	c := cl.Core()

	if c == nil || c.commands == nil {
		if cl.banner != nil {
			cl.Print(cl.banner(cl))
		}
		return Result{}
	}

	if c.commands.Len() == 0 {
		if cl.banner != nil {
			cl.Print(cl.banner(cl))
		}
		return Result{}
	}

	// Resolve command path from args
	var cmd *Command
	var remaining []string

	for i := len(clean); i > 0; i-- {
		path := JoinPath(clean[:i]...)
		if r := c.commands.Get(path); r.OK {
			cmd = r.Value.(*Command)
			remaining = clean[i:]
			break
		}
	}

	if cmd == nil {
		if cl.banner != nil {
			cl.Print(cl.banner(cl))
		}
		cl.PrintHelp()
		return Result{}
	}

	// Build options from remaining args
	opts := NewOptions()
	for _, arg := range remaining {
		key, val, valid := ParseFlag(arg)
		if valid {
			if Contains(arg, "=") {
				opts.Set(key, val)
			} else {
				opts.Set(key, true)
			}
		} else if !IsFlag(arg) {
			opts.Set("_arg", arg)
		}
	}

	if cmd.Action != nil {
		return cmd.Run(opts)
	}
	return Result{E("core.Cli.Run", Concat("command \"", cmd.Path, "\" is not executable"), nil), false}
}

// PrintHelp prints available commands.
//
//	c.Cli().PrintHelp()
func (cl *Cli) PrintHelp() {
	c := cl.Core()
	if c == nil || c.commands == nil {
		return
	}

	name := ""
	if c.app != nil {
		name = c.app.Name
	}
	if name != "" {
		cl.Print("%s commands:", name)
	} else {
		cl.Print("Commands:")
	}

	c.commands.Each(func(path string, cmd *Command) {
		if cmd.Hidden || (cmd.Action == nil && !cmd.IsManaged()) {
			return
		}
		tr := c.I18n().Translate(cmd.I18nKey())
		desc, _ := tr.Value.(string)
		if desc == "" || desc == cmd.I18nKey() {
			cl.Print("  %s", path)
		} else {
			cl.Print("  %-30s %s", path, desc)
		}
	})
}

// SetBanner sets the banner function.
//
//	c.Cli().SetBanner(func(_ *core.Cli) string { return "My App v1.0" })
func (cl *Cli) SetBanner(fn func(*Cli) string) {
	cl.banner = fn
}

// Banner returns the banner string.
//
//	c := core.New(core.WithOption("name", "homelab"))
//	banner := c.Cli().Banner()
//	core.Println(banner)
func (cl *Cli) Banner() string {
	if cl.banner != nil {
		return cl.banner(cl)
	}
	c := cl.Core()
	if c != nil && c.app != nil && c.app.Name != "" {
		return c.app.Name
	}
	return ""
}
