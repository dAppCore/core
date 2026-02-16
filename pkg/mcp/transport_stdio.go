package mcp

import (
	"context"

	"forge.lthn.ai/core/go/pkg/log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServeStdio starts the MCP server over stdin/stdout.
// This is the default transport for CLI integrations.
func (s *Service) ServeStdio(ctx context.Context) error {
	s.logger.Info("MCP Stdio server starting", "user", log.Username())
	return s.server.Run(ctx, &mcp.StdioTransport{})
}
