// Package clir provides a simple API for creating command line apps
package core

import (
	"fmt"
)

// defaultBannerFunction prints a banner for the application.
// If version is a blank string, it is ignored.
func defaultBannerFunction(c *CliApp) string {
	version := ""
	if len(c.Version()) > 0 {
		version = " " + c.Version()
	}
	return fmt.Sprintf("%s%s - %s", c.Name(), version, c.ShortDescription())
}

// NewCli - Creates a new Cli application object
func NewCliApp(name, description, version string) *CliApp {
	result := &CliApp{
		version:        version,
		bannerFunction: defaultBannerFunction,
	}
	result.rootCommand = NewCommand(name, description)
	result.rootCommand.setApp(result)
	result.rootCommand.setParentCommandPath("")
	return result
}
