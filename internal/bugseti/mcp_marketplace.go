// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Marketplace struct {
	Schema      string              `json:"$schema,omitempty"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Owner       MarketplaceOwner    `json:"owner"`
	Plugins     []MarketplacePlugin `json:"plugins"`
}

type MarketplaceOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type MarketplacePlugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	Category    string `json:"category"`
}

type PluginInfo struct {
	Plugin   MarketplacePlugin `json:"plugin"`
	Path     string            `json:"path"`
	Manifest map[string]any    `json:"manifest,omitempty"`
	Commands []string          `json:"commands,omitempty"`
	Skills   []string          `json:"skills,omitempty"`
}

type EthicsContext struct {
	Modal  string         `json:"modal"`
	Axioms map[string]any `json:"axioms"`
}

type marketplaceClient interface {
	ListMarketplace(ctx context.Context) ([]MarketplacePlugin, error)
	PluginInfo(ctx context.Context, name string) (*PluginInfo, error)
	EthicsCheck(ctx context.Context) (*EthicsContext, error)
	Close() error
}

type mcpMarketplaceClient struct {
	client *client.Client
}

func newMarketplaceClient(ctx context.Context) (marketplaceClient, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	command, args, err := resolveMarketplaceCommand()
	if err != nil {
		return nil, err
	}

	mcpClient, err := client.NewStdioMCPClient(command, nil, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to start marketplace MCP client: %w", err)
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "bugseti",
		Version: GetVersion(),
	}

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := mcpClient.Initialize(initCtx, initRequest); err != nil {
		_ = mcpClient.Close()
		return nil, fmt.Errorf("failed to initialize marketplace MCP client: %w", err)
	}

	return &mcpMarketplaceClient{client: mcpClient}, nil
}

func (c *mcpMarketplaceClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *mcpMarketplaceClient) ListMarketplace(ctx context.Context) ([]MarketplacePlugin, error) {
	var marketplace Marketplace
	if err := c.callToolStructured(ctx, "marketplace_list", nil, &marketplace); err != nil {
		return nil, err
	}
	return marketplace.Plugins, nil
}

func (c *mcpMarketplaceClient) PluginInfo(ctx context.Context, name string) (*PluginInfo, error) {
	var info PluginInfo
	args := map[string]any{"name": name}
	if err := c.callToolStructured(ctx, "marketplace_plugin_info", args, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *mcpMarketplaceClient) EthicsCheck(ctx context.Context) (*EthicsContext, error) {
	var ethics EthicsContext
	if err := c.callToolStructured(ctx, "ethics_check", nil, &ethics); err != nil {
		return nil, err
	}
	return &ethics, nil
}

func (c *mcpMarketplaceClient) callToolStructured(ctx context.Context, name string, args map[string]any, target any) error {
	if c == nil || c.client == nil {
		return errors.New("marketplace client is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	request := mcp.CallToolRequest{}
	request.Params.Name = name
	if args != nil {
		request.Params.Arguments = args
	}

	result, err := c.client.CallTool(ctx, request)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("marketplace tool returned no result")
	}
	if result.IsError {
		return fmt.Errorf("marketplace tool %s error: %s", name, toolResultMessage(result))
	}
	if result.StructuredContent == nil {
		return fmt.Errorf("marketplace tool %s returned no structured content", name)
	}
	payload, err := json.Marshal(result.StructuredContent)
	if err != nil {
		return fmt.Errorf("failed to encode marketplace response: %w", err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("failed to decode marketplace response: %w", err)
	}
	return nil
}

func toolResultMessage(result *mcp.CallToolResult) string {
	if result == nil {
		return "unknown error"
	}
	for _, content := range result.Content {
		switch value := content.(type) {
		case mcp.TextContent:
			if value.Text != "" {
				return value.Text
			}
		case *mcp.TextContent:
			if value != nil && value.Text != "" {
				return value.Text
			}
		}
	}
	return "unknown error"
}

func resolveMarketplaceCommand() (string, []string, error) {
	if command := strings.TrimSpace(os.Getenv("BUGSETI_MCP_COMMAND")); command != "" {
		args := strings.Fields(os.Getenv("BUGSETI_MCP_ARGS"))
		return command, args, nil
	}

	if root := strings.TrimSpace(os.Getenv("BUGSETI_MCP_ROOT")); root != "" {
		path := filepath.Join(root, "mcp")
		return "go", []string{"run", path}, nil
	}

	if root, ok := findCoreAgentRoot(); ok {
		return "go", []string{"run", filepath.Join(root, "mcp")}, nil
	}

	return "", nil, fmt.Errorf("marketplace MCP server not configured (set BUGSETI_MCP_COMMAND or BUGSETI_MCP_ROOT)")
}

func findCoreAgentRoot() (string, bool) {
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
		candidates = append(candidates, filepath.Dir(cwd))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, exeDir)
		candidates = append(candidates, filepath.Dir(exeDir))
	}

	seen := make(map[string]bool)
	for _, base := range candidates {
		base = filepath.Clean(base)
		if seen[base] {
			continue
		}
		seen[base] = true

		root := filepath.Join(base, "core-agent")
		if hasMcpDir(root) {
			return root, true
		}

		root = filepath.Join(base, "..", "core-agent")
		if hasMcpDir(root) {
			return filepath.Clean(root), true
		}
	}

	return "", false
}

func hasMcpDir(root string) bool {
	if root == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(root, "mcp", "main.go"))
	return err == nil && !info.IsDir()
}
