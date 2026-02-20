# go-api Design — REST Framework + OpenAPI SDK Generation

**Date:** 2026-02-20
**Author:** Virgil
**Status:** Approved
**Module:** `forge.lthn.ai/core/go-api`

## Problem

The Core Go ecosystem exposes 42+ tools via MCP (JSON-RPC), which is ideal for AI agents but inaccessible to regular HTTP clients, frontend applications, and third-party integrators. There is no REST API layer, no OpenAPI specification, and no generated SDKs.

Both external customers (Host UK products) and Lethean network peers need programmatic access to the same services.

## Solution

A new `go-api` package that provides:

1. **Gin-based REST framework** with standard middleware
2. **RouteGroup interface** that subsystems implement to register their own endpoints
3. **WebSocket integration** via existing go-ws Hub for real-time streaming
4. **OpenAPI 3.1 spec generation** via swaggo annotations
5. **SDK generation pipeline** targeting Python, TypeScript, Go client, and more

## Architecture

### Three-Protocol Access

Same backend services, three client protocols:

```
                    ┌─── REST (go-api)     POST /v1/ml/generate → JSON
                    │
Client ────────────┼─── WebSocket (go-ws)  subscribe ml.generate → streaming
                    │
                    └─── MCP (go-ai)       ml_generate → JSON-RPC
```

### Dependency Graph

```
go-api (Gin engine + middleware + OpenAPI)
    ↑ imported by (each registers its own routes)
    ├── go-ai/api/       → /v1/file/*, /v1/process/*, /v1/metrics/*
    ├── go-ml/api/       → /v1/ml/*
    ├── go-rag/api/      → /v1/rag/*
    ├── go-agentic/api/  → /v1/tasks/*
    ├── go-help/api/     → /v1/help/*
    └── go-ws/api/       → /ws (WebSocket upgrade)
```

go-api has zero internal ecosystem dependencies. Subsystems import go-api, not the other way round.

### Subsystem Opt-In

Not every MCP tool becomes a REST endpoint. Each subsystem decides what to expose via a separate `RegisterAPI()` method, independent of MCP's `RegisterTools()`. A subsystem with 15 MCP tools might expose 5 REST endpoints.

## Package Structure

```
forge.lthn.ai/core/go-api
├── api.go            # Engine struct, New(), Serve(), Shutdown()
├── middleware.go      # Auth, CORS, rate limiting, request logging, recovery
├── options.go         # WithAddr, WithAuth, WithCORS, WithRateLimit, etc.
├── group.go           # RouteGroup interface + registration
├── response.go        # Envelope type, error responses, pagination
├── docs/              # Generated swagger docs (swaggo output)
├── sdk/               # SDK generation tooling / Makefile targets
└── go.mod             # forge.lthn.ai/core/go-api
```

## Core Interface

```go
// RouteGroup registers API routes onto a Gin router group.
// Subsystems implement this to expose their endpoints.
type RouteGroup interface {
    // Name returns the route group identifier (e.g. "ml", "rag", "tasks")
    Name() string
    // BasePath returns the URL prefix (e.g. "/v1/ml")
    BasePath() string
    // RegisterRoutes adds handlers to the provided router group
    RegisterRoutes(rg *gin.RouterGroup)
}

// StreamGroup optionally declares WebSocket channels a subsystem publishes to.
type StreamGroup interface {
    Channels() []string
}
```

### Subsystem Example (go-ml)

```go
// In go-ml/api/routes.go
package api

type Routes struct {
    service *ml.Service
}

func NewRoutes(svc *ml.Service) *Routes {
    return &Routes{service: svc}
}

func (r *Routes) Name() string     { return "ml" }
func (r *Routes) BasePath() string { return "/v1/ml" }

func (r *Routes) RegisterRoutes(rg *gin.RouterGroup) {
    rg.POST("/generate", r.Generate)
    rg.POST("/score", r.Score)
    rg.GET("/backends", r.Backends)
    rg.GET("/status", r.Status)
}

func (r *Routes) Channels() []string {
    return []string{"ml.generate", "ml.status"}
}

// @Summary Generate text via ML backend
// @Tags ml
// @Accept json
// @Produce json
// @Param input body MLGenerateInput true "Generation parameters"
// @Success 200 {object} Response[MLGenerateOutput]
// @Router /v1/ml/generate [post]
func (r *Routes) Generate(c *gin.Context) {
    var input MLGenerateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, api.Fail("invalid_input", err.Error()))
        return
    }
    result, err := r.service.Generate(c.Request.Context(), input.Backend, input.Prompt, ml.GenOpts{
        Temperature: input.Temperature,
        MaxTokens:   input.MaxTokens,
        Model:       input.Model,
    })
    if err != nil {
        c.JSON(500, api.Fail("ml.generate_failed", err.Error()))
        return
    }
    c.JSON(200, api.OK(MLGenerateOutput{
        Response: result,
        Backend:  input.Backend,
        Model:    input.Model,
    }))
}
```

### Engine Wiring (in core CLI)

```go
engine := api.New(
    api.WithAddr(":8080"),
    api.WithCORS("*"),
    api.WithAuth(api.BearerToken(cfg.APIKey)),
    api.WithRateLimit(100, time.Minute),
    api.WithWSHub(wsHub),
)

engine.Register(mlapi.NewRoutes(mlService))
engine.Register(ragapi.NewRoutes(ragService))
engine.Register(agenticapi.NewRoutes(agenticService))

engine.Serve(ctx)  // Blocks until context cancelled
```

## Response Envelope

All endpoints return a consistent envelope:

```go
type Response[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data,omitempty"`
    Error   *Error `json:"error,omitempty"`
    Meta    *Meta  `json:"meta,omitempty"`
}

type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

type Meta struct {
    RequestID string `json:"request_id"`
    Duration  string `json:"duration"`
    Page      int    `json:"page,omitempty"`
    PerPage   int    `json:"per_page,omitempty"`
    Total     int    `json:"total,omitempty"`
}
```

Helper functions:

```go
func OK[T any](data T) Response[T]
func Fail(code, message string) Response[any]
func Paginated[T any](data T, page, perPage, total int) Response[T]
```

## Middleware Stack

```go
api.New(
    api.WithAddr(":8080"),
    api.WithCORS(api.CORSConfig{...}),         // gin-contrib/cors
    api.WithAuth(api.BearerToken("...")),       // Phase 1: simple bearer token
    api.WithRateLimit(100, time.Minute),        // Per-IP sliding window
    api.WithRequestID(),                        // X-Request-ID header generation
    api.WithRecovery(),                         // Panic recovery → 500 response
    api.WithLogger(slog.Default()),             // Structured request logging
)
```

Auth evolution path: bearer token → API keys → JWT → OAuth2/OIDC. Middleware slot stays the same.

## WebSocket Integration

go-api wraps the existing go-ws Hub as a first-class transport:

```go
// Automatic registration:
// GET /ws → WebSocket upgrade (go-ws Hub)

// Client subscribes:   {"type":"subscribe","channel":"ml.generate"}
// Events arrive:       {"type":"event","channel":"ml.generate","data":{...}}
// Client unsubscribes: {"type":"unsubscribe","channel":"ml.generate"}
```

Subsystems implementing `StreamGroup` declare which channels they publish to. This metadata feeds into the OpenAPI spec as documentation.

## OpenAPI + SDK Generation

### Spec Generation (swaggo)

```bash
# Generate from swaggo annotations
swag init -g api.go -o docs/ --parseDependency --parseInternal

# Output: docs/swagger.json, docs/swagger.yaml
```

### Swagger UI

```go
// Built-in at GET /swagger/*any
// Served via gin-swagger middleware
```

### SDK Generation

```bash
# Via openapi-generator-cli
openapi-generator-cli generate -i docs/swagger.json -g python -o sdk/python
openapi-generator-cli generate -i docs/swagger.json -g typescript-axios -o sdk/typescript
openapi-generator-cli generate -i docs/swagger.json -g go -o sdk/go-client
openapi-generator-cli generate -i docs/swagger.json -g java -o sdk/java
openapi-generator-cli generate -i docs/swagger.json -g csharp -o sdk/csharp
```

### CLI Commands

```bash
core api serve                    # Start REST + WebSocket server
core api spec                     # Emit OpenAPI JSON to stdout
core api spec --format yaml       # YAML variant
core api sdk --lang python        # Generate Python SDK to sdk/python/
core api sdk --lang typescript    # Generate TypeScript SDK
```

## Dependencies

| Package | Purpose | Size |
|---------|---------|------|
| `github.com/gin-gonic/gin` | HTTP framework | ~15K LOC |
| `github.com/swaggo/swag` | OpenAPI annotation parser | ~8K LOC |
| `github.com/swaggo/gin-swagger` | Swagger UI middleware | ~500 LOC |
| `github.com/gin-contrib/cors` | CORS middleware | ~300 LOC |
| `forge.lthn.ai/core/go-ws` | WebSocket Hub (existing) | ~1.5K LOC |

## Estimated Size

| Component | LOC |
|-----------|-----|
| Engine + options | ~200 |
| Middleware | ~150 |
| Response envelope | ~80 |
| RouteGroup interface | ~30 |
| WebSocket integration | ~60 |
| Tests | ~300 |
| **Total go-api** | **~820** |

Each subsystem's `api/` package adds ~100-200 LOC per route group.

## Non-Goals (Phase 1)

- GraphQL endpoint
- gRPC gateway
- Automatic MCP-to-REST adapter (can be added later)
- OAuth2 server (use external provider)
- API versioning beyond /v1/ prefix

## Success Criteria

1. `core api serve` starts a Gin server with registered subsystem routes
2. `core api spec` emits valid OpenAPI 3.1 JSON
3. Generated Python and TypeScript SDKs can call endpoints successfully
4. WebSocket subscriptions work alongside REST
5. Swagger UI accessible at `/swagger/`
6. All endpoints return consistent Response envelope
7. Bearer token auth protects all routes
