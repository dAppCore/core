package core

import (
	"fmt"
	"os"
)

// Cli - The main application object.
type CliApp struct {
	version        string
	rootCommand    *Command
	defaultCommand *Command
	preRunCommand  func(*CliApp) error
	postRunCommand func(*CliApp) error
	bannerFunction func(*CliApp) string
	errorHandler   func(string, error) error
}

// Version - Get the Application version string.
func (c *CliApp) Version() string {
	return c.version
}

// Name - Get the Application Name
func (c *CliApp) Name() string {
	return c.rootCommand.name
}

// ShortDescription - Get the Application short description.
func (c *CliApp) ShortDescription() string {
	return c.rootCommand.shortdescription
}

// SetBannerFunction - Set the function that is called
// to get the banner string.
func (c *CliApp) SetBannerFunction(fn func(*CliApp) string) {
	c.bannerFunction = fn
}

// SetErrorFunction - Set custom error message when undefined
// flags are used by the user. First argument is a string containing
// the command path used. Second argument is the undefined flag error.
func (c *CliApp) SetErrorFunction(fn func(string, error) error) {
	c.errorHandler = fn
}

// AddCommand - Adds a command to the application.
func (c *CliApp) AddCommand(command *Command) {
	c.rootCommand.AddCommand(command)
}

// PrintBanner - Prints the application banner!
func (c *CliApp) PrintBanner() {
	fmt.Println(c.bannerFunction(c))
	fmt.Println("")
}

// PrintHelp - Prints the application's help.
func (c *CliApp) PrintHelp() {
	c.rootCommand.PrintHelp()
}

// Run - Runs the application with the given arguments.
func (c *CliApp) Run(args ...string) error {
	if c.preRunCommand != nil {
		err := c.preRunCommand(c)
		if err != nil {
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
		err := c.postRunCommand(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// DefaultCommand - Sets the given command as the command to run when
// no other commands given.
func (c *CliApp) DefaultCommand(defaultCommand *Command) *CliApp {
	c.defaultCommand = defaultCommand
	return c
}

// NewSubCommand - Creates a new SubCommand for the application.
func (c *CliApp) NewSubCommand(name, description string) *Command {
	return c.rootCommand.NewSubCommand(name, description)
}

// NewSubCommandInheritFlags - Creates a new SubCommand for the application, inherit flags from parent Command
func (c *CliApp) NewSubCommandInheritFlags(name, description string) *Command {
	return c.rootCommand.NewSubCommandInheritFlags(name, description)
}

// PreRun - Calls the given function before running the specific command.
func (c *CliApp) PreRun(callback func(*CliApp) error) {
	c.preRunCommand = callback
}

// PostRun - Calls the given function after running the specific command.
func (c *CliApp) PostRun(callback func(*CliApp) error) {
	c.postRunCommand = callback
}

// BoolFlag - Adds a boolean flag to the root command.
func (c *CliApp) BoolFlag(name, description string, variable *bool) *CliApp {
	c.rootCommand.BoolFlag(name, description, variable)
	return c
}

// StringFlag - Adds a string flag to the root command.
func (c *CliApp) StringFlag(name, description string, variable *string) *CliApp {
	c.rootCommand.StringFlag(name, description, variable)
	return c
}

// IntFlag - Adds an int flag to the root command.
func (c *CliApp) IntFlag(name, description string, variable *int) *CliApp {
	c.rootCommand.IntFlag(name, description, variable)
	return c
}

func (c *CliApp) AddFlags(flags interface{}) *CliApp {
	c.rootCommand.AddFlags(flags)
	return c
}

// Action - Define an action from this command.
func (c *CliApp) Action(callback CliAction) *CliApp {
	c.rootCommand.Action(callback)
	return c
}

// LongDescription - Sets the long description for the command.
func (c *CliApp) LongDescription(longdescription string) *CliApp {
	c.rootCommand.LongDescription(longdescription)
	return c
}

// OtherArgs - Returns the non-flag arguments passed to the cli.
// NOTE: This should only be called within the context of an action.
func (c *CliApp) OtherArgs() []string {
	return c.rootCommand.flags.Args()
}

func (c *CliApp) NewSubCommandFunction(name string, description string, test interface{}) *CliApp {
	c.rootCommand.NewSubCommandFunction(name, description, test)
	return c
}
