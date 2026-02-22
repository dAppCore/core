# MCP Integration — Completion Summary

**Completed:** 2026-02-05
**Plan:** `docs/plans/2026-02-05-mcp-integration.md`

## What Was Built

### RAG Tools (`pkg/mcp/tools_rag.go`)
Three MCP tools added to the existing `pkg/mcp` server:
- `rag_query` — semantic search against Qdrant vector DB
- `rag_ingest` — ingest a file or directory into a named collection
- `rag_collections` — list available Qdrant collections (with optional stats)

### Metrics Tools (`pkg/mcp/tools_metrics.go`)
Two MCP tools for agent activity tracking:
- `metrics_record` — write a typed event (agent_id, repo, arbitrary data) to JSONL storage
- `metrics_query` — query events with aggregation by type, repo, and agent; supports human-friendly duration strings (7d, 24h)

Also added `parseDuration()` helper for "Nd"/"Nh"/"Nm" duration strings.

### `core mcp serve` Command (`internal/cmd/mcpcmd/cmd_mcp.go`)
New CLI sub-command registered via `cli.WithCommands()` (not `init()`).
- Runs `pkg/mcp` server over stdio by default
- TCP mode via `MCP_ADDR=:9000` environment variable
- `--workspace` flag to restrict file operations to a directory

Registered in the full CLI variant. i18n strings added for all user-facing text.

### Plugin Configuration
`.mcp.json` created for the `agentic-flows` Claude Code plugin, pointing to `core mcp serve`. Exposes all 15 tools to Claude Code agents via the `core-cli` MCP server name.

## Key Outcomes

- `core mcp serve` is the single entry point for all MCP tooling (file ops, RAG, metrics, language detection, process management, WebSocket, webview/CDP)
- MCP command moved to `go-ai/cmd/mcpcmd/` in final form; the plan's `internal/cmd/mcpcmd/` path reflects the pre-extraction location
- Registration pattern updated from `init()` + `RegisterCommands()` to `cli.WithCommands()` lifecycle hooks
- Services required at runtime: Qdrant (localhost:6333), Ollama with nomic-embed-text (localhost:11434)
