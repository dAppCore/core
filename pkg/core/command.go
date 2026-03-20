// SPDX-License-Identifier: EUPL-1.2

// Command is a DTO representing an executable operation.
// Commands don't know if they're root, child, or nested — the tree
// structure comes from composition via path-based registration.
//
// Register a command:
//
//	c.Command("deploy", func(opts core.Options) core.Result[any] {
//	    return core.Result[any]{Value: "deployed", OK: true}
//	})
//
// Register a nested command:
//
//	c.Command("deploy/to/homelab", handler)
//
// Description is an i18n key — derived from path if omitted:
//
//	"deploy"             → "cmd.deploy.description"
//	"deploy/to/homelab"  → "cmd.deploy.to.homelab.description"
package core

import (
	"strings"
	"sync"
)

// CommandAction is the function signature for command handlers.
//
//	func(opts core.Options) core.Result[any]
type CommandAction func(Options) Result[any]

// CommandLifecycle is implemented by commands that support managed lifecycle.
// Basic commands only need an action. Daemon commands implement Start/Stop/Signal
// via go-process.
type CommandLifecycle interface {
	Start(Options) Result[any]
	Stop() Result[any]
	Restart() Result[any]
	Reload() Result[any]
	Signal(string) Result[any]
}

// Command is the DTO for an executable operation.
type Command struct {
	name        string
	description string               // i18n key — derived from path if empty
	path        string               // "deploy/to/homelab"
	commands    map[string]*Command   // child commands
	action      CommandAction         // business logic
	lifecycle   CommandLifecycle      // optional — provided by go-process
	flags       Options              // declared flags
	hidden      bool
	mu          sync.RWMutex
}

// I18nKey returns the i18n key for this command's description.
//
//	cmd with path "deploy/to/homelab" → "cmd.deploy.to.homelab.description"
func (cmd *Command) I18nKey() string {
	if cmd.description != "" {
		return cmd.description
	}
	path := cmd.path
	if path == "" {
		path = cmd.name
	}
	return "cmd." + strings.ReplaceAll(path, "/", ".") + ".description"
}

// Run executes the command's action with the given options.
//
//	result := cmd.Run(core.Options{{K: "target", V: "homelab"}})
func (cmd *Command) Run(opts Options) Result[any] {
	if cmd.action == nil {
		return Result[any]{}
	}
	return cmd.action(opts)
}

// Start delegates to the lifecycle implementation if available.
func (cmd *Command) Start(opts Options) Result[any] {
	if cmd.lifecycle != nil {
		return cmd.lifecycle.Start(opts)
	}
	return cmd.Run(opts)
}

// Stop delegates to the lifecycle implementation.
func (cmd *Command) Stop() Result[any] {
	if cmd.lifecycle != nil {
		return cmd.lifecycle.Stop()
	}
	return Result[any]{}
}

// Restart delegates to the lifecycle implementation.
func (cmd *Command) Restart() Result[any] {
	if cmd.lifecycle != nil {
		return cmd.lifecycle.Restart()
	}
	return Result[any]{}
}

// Reload delegates to the lifecycle implementation.
func (cmd *Command) Reload() Result[any] {
	if cmd.lifecycle != nil {
		return cmd.lifecycle.Reload()
	}
	return Result[any]{}
}

// Signal delegates to the lifecycle implementation.
func (cmd *Command) Signal(sig string) Result[any] {
	if cmd.lifecycle != nil {
		return cmd.lifecycle.Signal(sig)
	}
	return Result[any]{}
}

// --- Command Registry (on Core) ---

// commandRegistry holds the command tree.
type commandRegistry struct {
	commands map[string]*Command
	mu       sync.RWMutex
}

// CommandHandler registers or retrieves commands on Core.
// Same pattern as Service() — zero args returns registry, one arg gets, two args registers.
//
//	c.Command("deploy", handler)           // register
//	c.Command("deploy/to/homelab", handler) // register nested
//	cmd := c.Command("deploy")              // get
func (c *Core) Command(args ...any) any {
	if c.commands == nil {
		c.commands = &commandRegistry{commands: make(map[string]*Command)}
	}

	switch len(args) {
	case 0:
		return c.commands
	case 1:
		path, _ := args[0].(string)
		c.commands.mu.RLock()
		cmd := c.commands.commands[path]
		c.commands.mu.RUnlock()
		return cmd
	default:
		path, _ := args[0].(string)
		if path == "" {
			return E("core.Command", "command path cannot be empty", nil)
		}

		c.commands.mu.Lock()
		defer c.commands.mu.Unlock()

		cmd := &Command{
			name:     pathName(path),
			path:     path,
			commands: make(map[string]*Command),
		}

		// Second arg: action function or Options
		switch v := args[1].(type) {
		case CommandAction:
			cmd.action = v
		case func(Options) Result[any]:
			cmd.action = v
		case Options:
			cmd.description = v.String("description")
			cmd.hidden = v.Bool("hidden")
		}

		// Third arg if present: Options for metadata
		if len(args) > 2 {
			if opts, ok := args[2].(Options); ok {
				cmd.description = opts.String("description")
				cmd.hidden = opts.Bool("hidden")
			}
		}

		c.commands.commands[path] = cmd

		// Build parent chain — "deploy/to/homelab" creates "deploy" and "deploy/to" if missing
		parts := strings.Split(path, "/")
		for i := len(parts) - 1; i > 0; i-- {
			parentPath := strings.Join(parts[:i], "/")
			if _, exists := c.commands.commands[parentPath]; !exists {
				c.commands.commands[parentPath] = &Command{
					name:     parts[i-1],
					path:     parentPath,
					commands: make(map[string]*Command),
				}
			}
			c.commands.commands[parentPath].commands[parts[i]] = cmd
			cmd = c.commands.commands[parentPath]
		}

		return nil
	}
}

// Commands returns all registered command paths.
//
//	paths := c.Commands()
func (c *Core) Commands() []string {
	if c.commands == nil {
		return nil
	}
	c.commands.mu.RLock()
	defer c.commands.mu.RUnlock()
	var paths []string
	for k := range c.commands.commands {
		paths = append(paths, k)
	}
	return paths
}

// pathName extracts the last segment of a path.
// "deploy/to/homelab" → "homelab"
func pathName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
