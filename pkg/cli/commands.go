// Package cli provides the CLI runtime and utilities.
package cli

import (
	"context"

	"forge.lthn.ai/core/go/pkg/framework"
	"github.com/spf13/cobra"
)

// WithCommands creates a framework Option that registers a command group.
// The register function receives the root command during service startup,
// allowing commands to participate in the Core lifecycle.
//
//	cli.Main(
//	    cli.WithCommands("config", config.AddConfigCommands),
//	    cli.WithCommands("doctor", doctor.AddDoctorCommands),
//	)
func WithCommands(name string, register func(root *Command)) framework.Option {
	return framework.WithName("cmd."+name, func(c *framework.Core) (any, error) {
		return &commandService{core: c, register: register}, nil
	})
}

type commandService struct {
	core     *framework.Core
	register func(root *Command)
}

func (s *commandService) OnStartup(_ context.Context) error {
	if root, ok := s.core.App.(*cobra.Command); ok {
		s.register(root)
	}
	return nil
}
