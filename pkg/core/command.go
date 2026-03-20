// SPDX-License-Identifier: EUPL-1.2

// Command is a DTO representing an executable operation.
// Commands don't know if they're root, child, or nested — the tree
// structure comes from composition via path-based registration.
//
// Register a command:
//
//	c.Command("deploy", func(opts core.Options) core.Result {
//	    return core.Result{"deployed", true}
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
	"sync"
)

// CommandAction is the function signature for command handlers.
//
//	func(opts core.Options) core.Result
type CommandAction func(Options) Result

// CommandLifecycle is implemented by commands that support managed lifecycle.
// Basic commands only need an action. Daemon commands implement Start/Stop/Signal
// via go-process.
type CommandLifecycle interface {
	Start(Options) Result
	Stop() Result
	Restart() Result
	Reload() Result
	Signal(string) Result
}

// Command is the DTO for an executable operation.
type Command struct {
	Name        string
	Description string           // i18n key — derived from path if empty
	Path        string           // "deploy/to/homelab"
	Action      CommandAction    // business logic
	Lifecycle   CommandLifecycle // optional — provided by go-process
	Flags       Options          // declared flags
	Hidden      bool
	commands    map[string]*Command // child commands (internal)
	mu          sync.RWMutex
}

// I18nKey returns the i18n key for this command's description.
//
//	cmd with path "deploy/to/homelab" → "cmd.deploy.to.homelab.description"
func (cmd *Command) I18nKey() string {
	if cmd.Description != "" {
		return cmd.Description
	}
	path := cmd.Path
	if path == "" {
		path = cmd.Name
	}
	return Concat("cmd.", Replace(path, "/", "."), ".description")
}

// Run executes the command's action with the given options.
//
//	result := cmd.Run(core.Options{{Key: "target", Value: "homelab"}})
func (cmd *Command) Run(opts Options) Result {
	if cmd.Action == nil {
		return Result{}
	}
	return cmd.Action(opts)
}

// Start delegates to the lifecycle implementation if available.
func (cmd *Command) Start(opts Options) Result {
	if cmd.Lifecycle != nil {
		return cmd.Lifecycle.Start(opts)
	}
	return cmd.Run(opts)
}

// Stop delegates to the lifecycle implementation.
func (cmd *Command) Stop() Result {
	if cmd.Lifecycle != nil {
		return cmd.Lifecycle.Stop()
	}
	return Result{}
}

// Restart delegates to the lifecycle implementation.
func (cmd *Command) Restart() Result {
	if cmd.Lifecycle != nil {
		return cmd.Lifecycle.Restart()
	}
	return Result{}
}

// Reload delegates to the lifecycle implementation.
func (cmd *Command) Reload() Result {
	if cmd.Lifecycle != nil {
		return cmd.Lifecycle.Reload()
	}
	return Result{}
}

// Signal delegates to the lifecycle implementation.
func (cmd *Command) Signal(sig string) Result {
	if cmd.Lifecycle != nil {
		return cmd.Lifecycle.Signal(sig)
	}
	return Result{}
}

// --- Command Registry (on Core) ---

// commandRegistry holds the command tree.
type commandRegistry struct {
	commands map[string]*Command
	mu       sync.RWMutex
}

// Command gets or registers a command by path.
//
//	c.Command("deploy", Command{Action: handler})
//	r := c.Command("deploy")
func (c *Core) Command(path string, command ...Command) Result {
	if c.commands == nil {
		c.commands = &commandRegistry{commands: make(map[string]*Command)}
	}

	if len(command) == 0 {
		c.commands.mu.RLock()
		cmd, ok := c.commands.commands[path]
		c.commands.mu.RUnlock()
		return Result{cmd, ok}
	}

	if path == "" {
		return Result{E("core.Command", "command path cannot be empty", nil), false}
	}

	c.commands.mu.Lock()
	defer c.commands.mu.Unlock()

	cmd := &command[0]
	cmd.Name = pathName(path)
	cmd.Path = path
	if cmd.commands == nil {
		cmd.commands = make(map[string]*Command)
	}

	c.commands.commands[path] = cmd

	// Build parent chain — "deploy/to/homelab" creates "deploy" and "deploy/to" if missing
	parts := Split(path, "/")
	for i := len(parts) - 1; i > 0; i-- {
		parentPath := JoinPath(parts[:i]...)
		if _, exists := c.commands.commands[parentPath]; !exists {
			c.commands.commands[parentPath] = &Command{
				Name:     parts[i-1],
				Path:     parentPath,
				commands: make(map[string]*Command),
			}
		}
		c.commands.commands[parentPath].commands[parts[i]] = cmd
		cmd = c.commands.commands[parentPath]
	}

	return Result{OK: true}
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
	parts := Split(path, "/")
	return parts[len(parts)-1]
}
