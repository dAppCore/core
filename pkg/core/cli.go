// SPDX-License-Identifier: EUPL-1.2

// CLI command framework for the Core framework.
// Based on leaanthony/clir — zero-dependency command line interface.

package core

import (
	"fmt"
	"os"
)

// CliAction represents a function called when a command is invoked.
type CliAction func() error

// CliOpts configures a Cli.
type CliOpts struct {
	Version     string
	Name        string
	Description string
}

// Cli is the CLI command framework.
type Cli struct {
	opts           *CliOpts
	rootCommand    *Command
	defaultCommand *Command
	preRunCommand  func(*Cli) error
	postRunCommand func(*Cli) error
	bannerFunction func(*Cli) string
	errorHandler   func(string, error) error
}

// defaultBannerFunction prints a banner for the application.
func defaultBannerFunction(c *Cli) string {
	version := ""
	if c.opts != nil && c.opts.Version != "" {
		version = " " + c.opts.Version
	}
	name := ""
	description := ""
	if c.opts != nil {
		name = c.opts.Name
		description = c.opts.Description
	}
	if description != "" {
		return fmt.Sprintf("%s%s - %s", name, version, description)
	}
	return fmt.Sprintf("%s%s", name, version)
}


// Command returns the root command.
func (c *Cli) Command() *Command {
	return c.rootCommand
}

// Version returns the application version string.
func (c *Cli) Version() string {
	if c.opts != nil {
		return c.opts.Version
	}
	return ""
}

// Name returns the application name.
func (c *Cli) Name() string {
	if c.opts != nil {
		return c.opts.Name
	}
	return c.rootCommand.name
}

// ShortDescription returns the application short description.
func (c *Cli) ShortDescription() string {
	if c.opts != nil {
		return c.opts.Description
	}
	return c.rootCommand.shortdescription
}

// SetBannerFunction sets the function that generates the banner string.
func (c *Cli) SetBannerFunction(fn func(*Cli) string) {
	c.bannerFunction = fn
}

// SetErrorFunction sets a custom error handler for undefined flags.
func (c *Cli) SetErrorFunction(fn func(string, error) error) {
	c.errorHandler = fn
}

// AddCommand adds a command to the application.
func (c *Cli) AddCommand(command *Command) {
	c.rootCommand.AddCommand(command)
}

// PrintBanner prints the application banner.
func (c *Cli) PrintBanner() {
	fmt.Println(c.bannerFunction(c))
	fmt.Println("")
}

// PrintHelp prints the application help.
func (c *Cli) PrintHelp() {
	c.rootCommand.PrintHelp()
}

// Run runs the application with the given arguments.
func (c *Cli) Run(args ...string) error {
	if c.preRunCommand != nil {
		if err := c.preRunCommand(c); err != nil {
			return err
		}
	}
	if len(args) == 0 {
		args = os.Args[1:]
	}
	if err := c.rootCommand.run(args); err != nil {
		return err
	}
	if c.postRunCommand != nil {
		if err := c.postRunCommand(c); err != nil {
			return err
		}
	}
	return nil
}

// DefaultCommand sets the command to run when no other commands are given.
func (c *Cli) DefaultCommand(defaultCommand *Command) *Cli {
	c.defaultCommand = defaultCommand
	return c
}

// NewChildCommand creates a new subcommand.
func (c *Cli) NewChildCommand(name string, description ...string) *Command {
	return c.rootCommand.NewChildCommand(name, description...)
}

// NewChildCommandInheritFlags creates a new subcommand that inherits parent flags.
func (c *Cli) NewChildCommandInheritFlags(name string, description ...string) *Command {
	return c.rootCommand.NewChildCommandInheritFlags(name, description...)
}

// PreRun sets a function to call before running the command.
func (c *Cli) PreRun(callback func(*Cli) error) {
	c.preRunCommand = callback
}

// PostRun sets a function to call after running the command.
func (c *Cli) PostRun(callback func(*Cli) error) {
	c.postRunCommand = callback
}

// BoolFlag adds a boolean flag to the root command.
func (c *Cli) BoolFlag(name, description string, variable *bool) *Cli {
	c.rootCommand.BoolFlag(name, description, variable)
	return c
}

// StringFlag adds a string flag to the root command.
func (c *Cli) StringFlag(name, description string, variable *string) *Cli {
	c.rootCommand.StringFlag(name, description, variable)
	return c
}

// IntFlag adds an int flag to the root command.
func (c *Cli) IntFlag(name, description string, variable *int) *Cli {
	c.rootCommand.IntFlag(name, description, variable)
	return c
}

// AddFlags adds struct-tagged flags to the root command.
func (c *Cli) AddFlags(flags any) *Cli {
	c.rootCommand.AddFlags(flags)
	return c
}

// Action defines an action for the root command.
func (c *Cli) Action(callback CliAction) *Cli {
	c.rootCommand.Action(callback)
	return c
}

// LongDescription sets the long description for the root command.
func (c *Cli) LongDescription(longdescription string) *Cli {
	c.rootCommand.LongDescription(longdescription)
	return c
}

// OtherArgs returns the non-flag arguments passed to the CLI.
func (c *Cli) OtherArgs() []string {
	return c.rootCommand.flags.Args()
}

// NewChildCommandFunction creates a subcommand from a function with struct flags.
func (c *Cli) NewChildCommandFunction(name string, description string, fn any) *Cli {
	c.rootCommand.NewChildCommandFunction(name, description, fn)
	return c
}
