# go-api — Completion Summary

**Completed:** 21 February 2026
**Module:** `forge.lthn.ai/core/go-api`
**Status:** Phases 1–3 complete, 176 tests passing

## What Was Built

### Phase 1 — Core Framework (20 Feb 2026)

Gin-based HTTP engine with extensible middleware via `With*()` options. Key components:

- `RouteGroup` / `StreamGroup` interfaces — subsystems register their own endpoints
- `Response[T]` envelope — `OK()`, `Fail()`, `Paginated()` generics
- `Engine` — `New()`, `Register()`, `Handler()`, `Serve()` with graceful shutdown
- Bearer auth, request ID, and CORS middleware
- WebSocket endpoint wrapping a `go-ws` Hub
- Swagger UI at `/swagger/` with runtime spec serving
- `/health` endpoint always available without auth
- First integration proof in `go-ml/api/` (3 endpoints, 12 tests)

### Phase 2 — Gin Plugin Stack (20–21 Feb 2026)

17 middleware plugins added across four waves, all as drop-in `With*()` options:

| Wave | Plugins |
|------|---------|
| 1 — Gateway hardening | Authentik (OIDC + forward auth), secure headers, structured slog, timeouts, gzip, static files |
| 2 — Performance + auth | Brotli compression, in-memory response cache, server-side sessions, Casbin RBAC |
| 3 — Network + streaming | HTTP signature verification, SSE broker, reverse proxy detection, i18n locale, GraphQL |
| 4 — Observability | pprof, expvar, OpenTelemetry distributed tracing |

### Phase 3 — OpenAPI + SDK Codegen (21 Feb 2026)

Runtime spec generation (not swaggo annotations — incompatible with dynamic RouteGroups and `Response[T]` generics):

- `DescribableGroup` interface — opt-in OpenAPI metadata for route groups
- `ToolBridge` — converts MCP tool descriptors into `POST /{tool_name}` REST endpoints
- `SpecBuilder` — assembles full OpenAPI 3.1 JSON from registered groups at runtime
- Spec export to JSON and YAML (`core api spec`)
- SDK codegen wrapper for openapi-generator-cli, 11 languages (`core api sdk --lang go`)
- `go-ai` `mcp/registry.go` — generic `addToolRecorded[In,Out]` captures types in closures
- `go-ai` `mcp/bridge.go` — `BridgeToAPI()` populates ToolBridge from MCP tool registry
- CLI commands: `core api spec`, `core api sdk` (in `core/cli` dev branch)

## Key Outcomes

- **176 tests** across go-api (143), go-ai bridge (10), and CLI commands (4), all passing
- Zero internal ecosystem dependencies — subsystems import go-api, not the reverse
- Authentik (OIDC) and bearer token auth coexist; Casbin adds RBAC on top
- Four-protocol access pattern established: REST, GraphQL, WebSocket, MCP — same handlers

## Known Limitations

- Subsystem MCP tools registered via `mcp.AddTool` directly are excluded from the REST bridge (only the 10 built-in tools appear). Fix: pass `*Service` to `RegisterTools` instead of `*mcp.Server`.
- `structSchema` reflection handles flat structs only; nested structs are not recursed.
- `core api spec` currently emits a spec with only `/health`; full MCP wiring into the CLI command is pending.
