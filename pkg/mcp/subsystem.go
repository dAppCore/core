package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Subsystem registers additional MCP tools at startup.
// Implementations should be safe to call concurrently.
type Subsystem interface {
	// Name returns a human-readable identifier for logging.
	Name() string

	// RegisterTools adds tools to the MCP server during initialisation.
	RegisterTools(server *mcp.Server)
}

// SubsystemWithShutdown extends Subsystem with graceful cleanup.
type SubsystemWithShutdown interface {
	Subsystem
	Shutdown(ctx context.Context) error
}

// WithSubsystem registers a subsystem whose tools will be added
// after the built-in tools during New().
func WithSubsystem(sub Subsystem) Option {
	return func(s *Service) error {
		s.subsystems = append(s.subsystems, sub)
		return nil
	}
}
