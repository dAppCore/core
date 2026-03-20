// SPDX-License-Identifier: EUPL-1.2

// Cli is the CLI surface layer for the Core command tree.
// It reads commands from Core's registry and wires them to terminal I/O.
//
// Run the CLI:
//
//	c := core.New(core.Options{{K: "name", V: "myapp"}})
//	c.Command("deploy", handler)
//	c.Cli().Run()
//
// The Cli resolves os.Args to a command path, parses flags,
// and calls the command's action with parsed options.
package core

import (
	"io"
	"os"
)

// Cli is the CLI surface for the Core command tree.
type Cli struct {
	core   *Core
	output io.Writer
	banner func(*Cli) string
}

// Print writes to the CLI output (defaults to os.Stdout).
//
//	c.Cli().Print("hello %s", "world")
func (cl *Cli) Print(format string, args ...any) {
	Print(cl.output, format, args...)
}

// SetOutput sets the CLI output writer.
//
//	c.Cli().SetOutput(os.Stderr)
func (cl *Cli) SetOutput(w io.Writer) {
	cl.output = w
}

// Run resolves os.Args to a command path and executes it.
//
//	c.Cli().Run()
//	c.Cli().Run("deploy", "to", "homelab")
func (cl *Cli) Run(args ...string) Result[any] {
	if len(args) == 0 {
		args = os.Args[1:]
	}

	clean := FilterArgs(args)

	if cl.core == nil || cl.core.commands == nil || len(cl.core.commands.commands) == 0 {
		if cl.banner != nil {
			cl.Print(cl.banner(cl))
		}
		return Result[any]{}
	}

	// Resolve command path from args
	var cmd *Command
	var remaining []string

	for i := len(clean); i > 0; i-- {
		path := JoinPath(clean[:i]...)
		if c, ok := cl.core.commands.commands[path]; ok {
			cmd = c
			remaining = clean[i:]
			break
		}
	}

	if cmd == nil {
		if cl.banner != nil {
			cl.Print(cl.banner(cl))
		}
		cl.PrintHelp()
		return Result[any]{}
	}

	// Build options from remaining args
	opts := Options{}
	for _, arg := range remaining {
		key, val, valid := ParseFlag(arg)
		if valid {
			if val != "" {
				opts = append(opts, Option{K: key, V: val})
			} else {
				opts = append(opts, Option{K: key, V: true})
			}
		} else if !IsFlag(arg) {
			opts = append(opts, Option{K: "_arg", V: arg})
		}
	}

	return cmd.Run(opts)
}

// PrintHelp prints available commands.
//
//	c.Cli().PrintHelp()
func (cl *Cli) PrintHelp() {
	if cl.core == nil || cl.core.commands == nil {
		return
	}

	name := ""
	if cl.core.app != nil {
		name = cl.core.app.Name
	}
	if name != "" {
		cl.Print("%s commands:", name)
	} else {
		cl.Print("Commands:")
	}

	cl.core.commands.mu.RLock()
	defer cl.core.commands.mu.RUnlock()

	for path, cmd := range cl.core.commands.commands {
		if cmd.hidden {
			continue
		}
		desc := cl.core.I18n().T(cmd.I18nKey())
		if desc == cmd.I18nKey() {
			cl.Print("  %s", path)
		} else {
			cl.Print("  %-30s %s", path, desc)
		}
	}
}

// SetBanner sets the banner function.
//
//	c.Cli().SetBanner(func(_ *core.Cli) string { return "My App v1.0" })
func (cl *Cli) SetBanner(fn func(*Cli) string) {
	cl.banner = fn
}

// Banner returns the banner string.
func (cl *Cli) Banner() string {
	if cl.banner != nil {
		return cl.banner(cl)
	}
	if cl.core != nil && cl.core.app != nil && cl.core.app.Name != "" {
		return cl.core.app.Name
	}
	return ""
}
