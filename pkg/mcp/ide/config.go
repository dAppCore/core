// Package ide provides an MCP subsystem that bridges the desktop IDE to
// a Laravel core-agentic backend over WebSocket.
package ide

import "time"

// Config holds connection and workspace settings for the IDE subsystem.
type Config struct {
	// LaravelWSURL is the WebSocket endpoint for the Laravel core-agentic backend.
	LaravelWSURL string

	// WorkspaceRoot is the local path used as the default workspace context.
	WorkspaceRoot string

	// ReconnectInterval controls how long to wait between reconnect attempts.
	ReconnectInterval time.Duration

	// MaxReconnectInterval caps exponential backoff for reconnection.
	MaxReconnectInterval time.Duration
}

// DefaultConfig returns sensible defaults for local development.
func DefaultConfig() Config {
	return Config{
		LaravelWSURL:         "ws://localhost:9876/ws",
		WorkspaceRoot:        ".",
		ReconnectInterval:    2 * time.Second,
		MaxReconnectInterval: 30 * time.Second,
	}
}

// Option configures the IDE subsystem.
type Option func(*Config)

// WithLaravelURL sets the Laravel WebSocket endpoint.
func WithLaravelURL(url string) Option {
	return func(c *Config) { c.LaravelWSURL = url }
}

// WithWorkspaceRoot sets the workspace root directory.
func WithWorkspaceRoot(root string) Option {
	return func(c *Config) { c.WorkspaceRoot = root }
}

// WithReconnectInterval sets the base reconnect interval.
func WithReconnectInterval(d time.Duration) Option {
	return func(c *Config) { c.ReconnectInterval = d }
}
