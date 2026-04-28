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

// CommandAction is the function signature for command handlers.
//
//	func(opts core.Options) core.Result
type CommandAction func(Options) Result

// Command is the DTO for an executable operation.
// Commands are declarative — they carry enough information for multiple consumers:
//
//   - core.Cli() runs the Action
//
//   - core/cli adds rich help, completion, man pages
//
//   - go-process wraps Managed commands with lifecycle (PID, health, signals)
//
//     c.Command("serve", core.Command{
//     Action:  handler,
//     Managed: "process.daemon",  // go-process provides start/stop/restart
//     })
type Command struct {
	Name        string
	Description string        // i18n key — derived from path if empty
	Path        string        // "deploy/to/homelab"
	Action      CommandAction // business logic
	Managed     string        // "" = one-shot, "process.daemon" = managed lifecycle
	Flags       Options       // declared flags
	Hidden      bool
	commands    map[string]*Command // child commands (internal)
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
//	result := cmd.Run(core.NewOptions(core.Option{Key: "target", Value: "homelab"}))
func (cmd *Command) Run(opts Options) Result {
	if cmd.Action == nil {
		return Result{E("core.Command.Run", Concat("command \"", cmd.Path, "\" is not executable"), nil), false}
	}
	return cmd.Action(opts)
}

// IsManaged returns true if this command has a managed lifecycle.
//
//	if cmd.IsManaged() { /* go-process handles start/stop */ }
func (cmd *Command) IsManaged() bool {
	return cmd.Managed != ""
}

// --- Command Registry (on Core) ---

// CommandRegistry holds the command tree. Embeds Registry[*Command]
// for thread-safe named storage with insertion order.
type CommandRegistry struct {
	*Registry[*Command]
}

// Command gets or registers a command by path.
//
//	c.Command("deploy", Command{Action: handler})
//	r := c.Command("deploy")
func (c *Core) Command(path string, command ...Command) Result {
	if len(command) == 0 {
		return c.commands.Get(path)
	}

	if path == "" || HasPrefix(path, "/") || HasSuffix(path, "/") || Contains(path, "//") {
		return Result{E("core.Command", Concat("invalid command path: \"", path, "\""), nil), false}
	}

	// Check for duplicate executable command
	if r := c.commands.Get(path); r.OK {
		existing := r.Value.(*Command)
		if existing.Action != nil || existing.IsManaged() {
			return Result{E("core.Command", Concat("command \"", path, "\" already registered"), nil), false}
		}
	}

	cmd := &command[0]
	cmd.Name = pathName(path)
	cmd.Path = path
	if cmd.commands == nil {
		cmd.commands = make(map[string]*Command)
	}

	// Preserve existing subtree when overwriting a placeholder parent
	if r := c.commands.Get(path); r.OK {
		existing := r.Value.(*Command)
		for k, v := range existing.commands {
			if _, has := cmd.commands[k]; !has {
				cmd.commands[k] = v
			}
		}
	}

	c.commands.Set(path, cmd)

	// Build parent chain — "deploy/to/homelab" creates "deploy" and "deploy/to" if missing
	parts := Split(path, "/")
	for i := len(parts) - 1; i > 0; i-- {
		parentPath := JoinPath(parts[:i]...)
		if !c.commands.Has(parentPath) {
			c.commands.Set(parentPath, &Command{
				Name:     parts[i-1],
				Path:     parentPath,
				commands: make(map[string]*Command),
			})
		}
		parent := c.commands.Get(parentPath).Value.(*Command)
		parent.commands[parts[i]] = cmd
		cmd = parent
	}

	return Result{OK: true}
}

// Commands returns all registered command paths in registration order.
//
//	paths := c.Commands()
func (c *Core) Commands() []string {
	if c.commands == nil {
		return nil
	}
	return c.commands.Names()
}

// pathName extracts the last segment of a path.
// "deploy/to/homelab" → "homelab"
func pathName(path string) string {
	parts := Split(path, "/")
	return parts[len(parts)-1]
}
