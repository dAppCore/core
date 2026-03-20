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
	"fmt"
	"os"
	"strings"
)

// Cli is the CLI surface for the Core command tree.
type Cli struct {
	core   *Core
	banner func(*Cli) string
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
		// No commands registered — print banner and exit
		if cl.banner != nil {
			fmt.Println(cl.banner(cl))
		}
		return Result[any]{}
	}

	// Resolve command path from args
	// "deploy to homelab" → try "deploy/to/homelab", then "deploy/to", then "deploy"
	var cmd *Command
	var remaining []string

	for i := len(clean); i > 0; i-- {
		path := strings.Join(clean[:i], "/")
		if c, ok := cl.core.commands.commands[path]; ok {
			cmd = c
			remaining = clean[i:]
			break
		}
	}

	if cmd == nil {
		// No matching command — try root-level action or print help
		if cl.banner != nil {
			fmt.Println(cl.banner(cl))
		}
		cl.PrintHelp()
		return Result[any]{}
	}

	// Build options from remaining args (flags become Options)
	opts := Options{}
	for _, arg := range remaining {
		key, val, valid := ParseFlag(arg)
		if valid {
			if val != "" {
				opts = append(opts, Option{K: key, V: val})
			} else {
				opts = append(opts, Option{K: key, V: true})
			}
		} else if !strings.HasPrefix(arg, "-") {
			opts = append(opts, Option{K: "_arg", V: arg})
		}
		// Invalid flags (e.g. -verbose, --v) are silently ignored
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
		fmt.Printf("%s commands:\n\n", name)
	} else {
		fmt.Println("Commands:\n")
	}

	cl.core.commands.mu.RLock()
	defer cl.core.commands.mu.RUnlock()

	for path, cmd := range cl.core.commands.commands {
		if cmd.hidden {
			continue
		}
		desc := cl.core.I18n().T(cmd.I18nKey())
		// If i18n returned the key itself (no translation), show path only
		if desc == cmd.I18nKey() {
			fmt.Printf("  %s\n", path)
		} else {
			fmt.Printf("  %-30s %s\n", path, desc)
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
