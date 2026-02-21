# go-api Design — HTTP Gateway + OpenAPI SDK Generation

**Date:** 2026-02-20
**Author:** Virgil
**Status:** Phase 1 + Phase 2 + Phase 3 Complete (176 tests in go-api)
**Module:** `forge.lthn.ai/core/go-api`

## Problem

The Core Go ecosystem exposes 42+ tools via MCP (JSON-RPC), which is ideal for AI agents but inaccessible to regular HTTP clients, frontend applications, and third-party integrators. There is no unified HTTP gateway, no OpenAPI specification, and no generated SDKs.

Both external customers (Host UK products) and Lethean network peers need programmatic access to the same services. The gateway also serves web routes, static assets, and streaming endpoints — not just REST APIs.

## Solution

A `go-api` package that acts as the central HTTP gateway:

1. **Gin-based HTTP gateway** with extensible middleware via gin-contrib plugins
2. **RouteGroup interface** that subsystems implement to register their own endpoints (API, web, or both)
3. **WebSocket + SSE integration** for real-time streaming
4. **OpenAPI 3.1 spec generation** via runtime SpecBuilder (not swaggo annotations)
5. **SDK generation pipeline** targeting 11 languages via openapi-generator-cli

## Architecture

### Four-Protocol Access

Same backend services, four client protocols:

```
                    ┌─── REST (go-api)      POST /v1/ml/generate → JSON
                    │
                    ├─── GraphQL (gqlgen)   mutation { mlGenerate(...) { response } }
Client ────────────┤
                    ├─── WebSocket (go-ws)  subscribe ml.generate → streaming
                    │
                    └─── MCP (go-ai)        ml_generate → JSON-RPC
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

Auth evolution path: bearer token → API keys → Authentik (OIDC/forward auth). Middleware slot stays the same.

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

### Runtime Spec Generation (SpecBuilder)

swaggo annotations were rejected because routes are dynamic via RouteGroup, Response[T] generics break swaggo, and MCP tools already carry JSON Schema at runtime. Instead, a `SpecBuilder` constructs the full OpenAPI 3.1 spec from registered RouteGroups at runtime.

```go
// Groups that implement DescribableGroup contribute endpoint metadata
type DescribableGroup interface {
    RouteGroup
    Describe() []RouteDescription
}

// SpecBuilder assembles the spec from all groups
builder := &api.SpecBuilder{Title: "Core API", Description: "...", Version: "1.0.0"}
spec, _ := builder.Build(engine.Groups())
```

### MCP-to-REST Bridge (ToolBridge)

The `ToolBridge` converts MCP tool descriptors into REST POST endpoints and implements both `RouteGroup` and `DescribableGroup`. Each tool becomes `POST /{tool_name}`. Generic types are captured at MCP registration time via closures, enabling JSON unmarshalling to the correct input type at request time.

```go
bridge := api.NewToolBridge("/v1/tools")
mcp.BridgeToAPI(mcpService, bridge)  // Populates bridge from MCP tool registry
engine.Register(bridge)              // Registers REST endpoints + OpenAPI metadata
```

### Swagger UI

```go
// Built-in at GET /swagger/*any
// SpecBuilder output served via gin-swagger, cached via sync.Once
api.New(api.WithSwagger("Core API", "...", "1.0.0"))
```

### SDK Generation

```bash
# Via openapi-generator-cli (11 languages supported)
core api sdk --lang go                          # Generate Go SDK
core api sdk --lang typescript-fetch,python     # Multiple languages
core api sdk --lang rust --output ./sdk/        # Custom output dir
```

### CLI Commands

```bash
core api spec                     # Emit OpenAPI JSON to stdout
core api spec --format yaml       # YAML variant
core api spec --output spec.json  # Write to file
core api sdk --lang python        # Generate Python SDK
core api sdk --lang go,rust       # Multiple SDKs
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP framework |
| `github.com/swaggo/gin-swagger` | Swagger UI middleware |
| `github.com/gin-contrib/cors` | CORS middleware |
| `github.com/gin-contrib/secure` | Security headers |
| `github.com/gin-contrib/sessions` | Server-side sessions |
| `github.com/gin-contrib/authz` | Casbin authorisation |
| `github.com/gin-contrib/httpsign` | HTTP signature verification |
| `github.com/gin-contrib/slog` | Structured request logging |
| `github.com/gin-contrib/timeout` | Per-request timeouts |
| `github.com/gin-contrib/gzip` | Gzip compression |
| `github.com/gin-contrib/static` | Static file serving |
| `github.com/gin-contrib/pprof` | Runtime profiling |
| `github.com/gin-contrib/expvar` | Runtime metrics |
| `github.com/gin-contrib/location/v2` | Reverse proxy detection |
| `github.com/99designs/gqlgen` | GraphQL endpoint |
| `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` | Distributed tracing |
| `gopkg.in/yaml.v3` | YAML spec export |
| `forge.lthn.ai/core/go-ws` | WebSocket Hub (existing) |

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

## Phase 1 — Implemented (20 Feb 2026)

**Commit:** `17ae945` on Forge (`core/go-api`)

| Component | Status | Tests |
|-----------|--------|-------|
| Response envelope (OK, Fail, Paginated) | Done | 9 |
| RouteGroup + StreamGroup interfaces | Done | 4 |
| Engine (New, Register, Handler, Serve) | Done | 9 |
| Bearer auth middleware | Done | 3 |
| Request ID middleware | Done | 2 |
| CORS middleware (gin-contrib/cors) | Done | 3 |
| WebSocket endpoint | Done | 3 |
| Swagger UI (gin-swagger) | Done | 2 |
| Health endpoint | Done | 1 |
| **Total** | **~840 LOC** | **36** |

**Integration proof:** go-ml/api/ registers 3 endpoints with 12 tests (`0c23858`).

## Phase 2 Wave 1 — Implemented (20 Feb 2026)

**Commits:** `6bb7195..daae6f7` on Forge (`core/go-api`)

| Component | Option | Dependency | Tests |
|-----------|--------|------------|-------|
| Authentik (forward auth + OIDC) | `WithAuthentik()` | `go-oidc/v3`, `oauth2` | 14 |
| Security headers (HSTS, CSP, etc.) | `WithSecure()` | `gin-contrib/secure` | 8 |
| Structured request logging | `WithSlog()` | `gin-contrib/slog` | 6 |
| Per-request timeouts | `WithTimeout()` | `gin-contrib/timeout` | 5 |
| Gzip compression | `WithGzip()` | `gin-contrib/gzip` | 5 |
| Static file serving | `WithStatic()` | `gin-contrib/static` | 5 |
| **Wave 1 Total** | | | **43** |

**Cumulative:** 76 tests (36 Phase 1 + 43 Wave 1 - 3 shared), all passing.

## Phase 2 Wave 2 — Implemented (20 Feb 2026)

**Commits:** `64a8b16..67dcc83` on Forge (`core/go-api`)

| Component | Option | Dependency | Tests | Notes |
|-----------|--------|------------|-------|-------|
| Brotli compression | `WithBrotli()` | `andybalholm/brotli` | 5 | Custom middleware; `gin-contrib/brotli` is empty stub |
| Response caching | `WithCache()` | none (in-memory) | 5 | Custom middleware; `gin-contrib/cache` is per-handler, not global |
| Server-side sessions | `WithSessions()` | `gin-contrib/sessions` | 5 | Cookie store, configurable name + secret |
| Casbin authorisation | `WithAuthz()` | `gin-contrib/authz`, `casbin/v2` | 5 | Subject via Basic Auth; RBAC policy model |
| **Wave 2 Total** | | | **20** | |

**Cumulative:** 102 passing tests (2 integration skipped), all green.

## Phase 2 Wave 3 — Implemented (20 Feb 2026)

**Commits:** `7b3f99e..d517fa2` on Forge (`core/go-api`)

| Component | Option | Dependency | Tests | Notes |
|-----------|--------|------------|-------|-------|
| HTTP signature verification | `WithHTTPSign()` | `gin-contrib/httpsign` | 5 | HMAC-SHA256; extensible via httpsign.Option |
| Server-Sent Events | `WithSSE()` | none (custom SSEBroker) | 6 | Channel filtering, multi-client broadcast, GET /events |
| Reverse proxy detection | `WithLocation()` | `gin-contrib/location/v2` | 5 | X-Forwarded-Host/Proto parsing |
| Locale detection | `WithI18n()` | `golang.org/x/text/language` | 5 | Accept-Language parsing, message lookup, GetLocale/GetMessage |
| GraphQL endpoint | `WithGraphQL()` | `99designs/gqlgen` | 5 | /graphql + optional /graphql/playground |
| **Wave 3 Total** | | | **26** | |

**Cumulative:** 128 passing tests (2 integration skipped), all green.

## Phase 2 Wave 4 — Implemented (21 Feb 2026)

**Commits:** `32b3680..8ba1716` on Forge (`core/go-api`)

| Component | Option | Dependency | Tests | Notes |
|-----------|--------|------------|-------|-------|
| Runtime profiling | `WithPprof()` | `gin-contrib/pprof` | 5 | /debug/pprof/* endpoints, flag-based mount |
| Runtime metrics | `WithExpvar()` | `gin-contrib/expvar` | 5 | /debug/vars endpoint, flag-based mount |
| Distributed tracing | `WithTracing()` | `otelgin` + OpenTelemetry SDK | 5 | W3C traceparent propagation, span attributes |
| **Wave 4 Total** | | | **15** | |

**Cumulative:** 143 passing tests (2 integration skipped), all green.

**Phase 2 complete.** All 4 waves implemented. Every planned plugin has a `With*()` option and tests.

## Phase 3 — OpenAPI Spec Generation + SDK Codegen (21 Feb 2026)

**Architecture:** Runtime OpenAPI generation via SpecBuilder (NOT swaggo annotations). Routes are dynamic via RouteGroup, Response[T] generics break swaggo, and MCP tools carry JSON Schema at runtime. A `ToolBridge` converts tool descriptors into RouteGroup + OpenAPI metadata. A `SpecBuilder` constructs the full OpenAPI 3.1 spec. SDK codegen wraps `openapi-generator-cli`.

### Wave 1: go-api (Tasks 1-5)

**Commits:** `465bd60..1910aec` on Forge (`core/go-api`)

| Component | File | Tests | Notes |
|-----------|------|-------|-------|
| DescribableGroup interface | `group.go` | 5 | Opt-in OpenAPI metadata for RouteGroups |
| ToolBridge | `bridge.go` | 6 | Tool descriptors → POST endpoints + DescribableGroup |
| SpecBuilder | `openapi.go` | 6 | OpenAPI 3.1 JSON with Response[T] envelope wrapping |
| Swagger refactor | `swagger.go` | 5 | Replaced hardcoded empty spec with SpecBuilder |
| Spec export | `export.go` | 5 | JSON + YAML export to file/writer |
| SDK codegen | `codegen.go` | 5 | 11-language wrapper for openapi-generator-cli |
| **Wave 1 Total** | | **32** | |

### Wave 2: go-ai MCP bridge (Tasks 6-7)

**Commits:** `2107eda..c37e1cf` on Forge (`core/go-ai`)

| Component | File | Tests | Notes |
|-----------|------|-------|-------|
| Tool registry | `mcp/registry.go` | 5 | Generic `addToolRecorded[In,Out]` captures types in closures |
| BridgeToAPI | `mcp/bridge.go` | 5 | MCP tools → go-api ToolBridge, 10MB body limit, error classification |
| **Wave 2 Total** | | **10** | |

### Wave 3: CLI commands (Tasks 8-9)

**Commit:** `d6eec4d` on Forge (`core/cli` dev branch)

| Component | File | Tests | Notes |
|-----------|------|-------|-------|
| `core api spec` | `cmd/api/cmd_spec.go` | 2 | JSON/YAML export, --output/--format flags |
| `core api sdk` | `cmd/api/cmd_sdk.go` | 2 | --lang (required), --output, --spec, --package flags |
| **Wave 3 Total** | | **4** | |

**Cumulative go-api:** 176 passing tests. **Phase 3 complete.**

### Known Limitations

- **Subsystem tools excluded from bridge:** Subsystems call `mcp.AddTool` directly, bypassing `addToolRecorded`. Only the 10 built-in MCP tools appear in the REST bridge. Future: pass `*Service` to `RegisterTools` instead of `*mcp.Server`.
- **Flat schema only:** `structSchema` reflection handles flat structs but does not recurse into nested structs. Adequate for current tool inputs.
- **CLI spec produces empty bridge:** `core api spec` currently generates a spec with only `/health`. Full MCP integration requires wiring the MCP service into the CLI command.

## Phase 2 — Gin Plugin Roadmap (Complete)

All plugins drop in as `With*()` options on the Engine. No architecture changes needed.

### Security & Auth

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~**Authentik**~~ | ~~`WithAuthentik()`~~ | ~~OIDC + forward auth integration.~~ | ~~**Done**~~ |
| ~~gin-contrib/secure~~ | ~~`WithSecure()`~~ | ~~Security headers: HSTS, X-Frame-Options, X-Content-Type-Options, CSP.~~ | ~~**Done**~~ |
| ~~gin-contrib/sessions~~ | ~~`WithSessions()`~~ | ~~Server-side sessions (cookie store). Web session management alongside Authentik tokens.~~ | ~~**Done**~~ |
| ~~gin-contrib/authz~~ | ~~`WithAuthz()`~~ | ~~Casbin-based authorisation. Policy-driven access control via RBAC.~~ | ~~**Done**~~ |
| ~~gin-contrib/httpsign~~ | ~~`WithHTTPSign()`~~ | ~~HTTP signature verification. HMAC-SHA256 with extensible options.~~ | ~~**Done**~~ |

### Performance & Reliability

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~gin-contrib/cache~~ | ~~`WithCache()`~~ | ~~Response caching (in-memory). GET response caching with TTL, lazy eviction.~~ | ~~**Done**~~ |
| ~~gin-contrib/timeout~~ | ~~`WithTimeout()`~~ | ~~Per-request timeouts.~~ | ~~**Done**~~ |
| ~~gin-contrib/gzip~~ | ~~`WithGzip()`~~ | ~~Gzip response compression.~~ | ~~**Done**~~ |
| ~~gin-contrib/brotli~~ | ~~`WithBrotli()`~~ | ~~Brotli compression via `andybalholm/brotli`. Custom middleware (gin-contrib stub empty).~~ | ~~**Done**~~ |

### Observability

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~gin-contrib/slog~~ | ~~`WithSlog()`~~ | ~~Structured request logging via slog.~~ | ~~**Done**~~ |
| ~~gin-contrib/pprof~~ | ~~`WithPprof()`~~ | ~~Runtime profiling endpoints at /debug/pprof/. Flag-based mount.~~ | ~~**Done**~~ |
| ~~gin-contrib/expvar~~ | ~~`WithExpvar()`~~ | ~~Go runtime metrics at /debug/vars. Flag-based mount.~~ | ~~**Done**~~ |
| ~~otelgin~~ | ~~`WithTracing()`~~ | ~~OpenTelemetry distributed tracing. W3C traceparent propagation.~~ | ~~**Done**~~ |

### Content & Streaming

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~gin-contrib/static~~ | ~~`WithStatic()`~~ | ~~Serve static files.~~ | ~~**Done**~~ |
| ~~gin-contrib/sse~~ | ~~`WithSSE()`~~ | ~~Server-Sent Events. Custom SSEBroker with channel filtering, GET /events.~~ | ~~**Done**~~ |
| ~~gin-contrib/location~~ | ~~`WithLocation()`~~ | ~~Auto-detect scheme/host from X-Forwarded-* headers.~~ | ~~**Done**~~ |

### Query Layer

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~99designs/gqlgen~~ | ~~`WithGraphQL()`~~ | ~~GraphQL endpoint at `/graphql` + optional playground. Accepts gqlgen ExecutableSchema.~~ | ~~**Done**~~ |

The GraphQL schema can be generated from the same Go Input/Output structs that define the REST endpoints. gqlgen produces an `http.Handler` that mounts directly on Gin. Subsystems opt-in via:

```go
// Subsystems that want GraphQL implement this alongside RouteGroup
type ResolverGroup interface {
    // RegisterResolvers adds query/mutation resolvers to the GraphQL schema
    RegisterResolvers(schema *graphql.Schema)
}
```

This means a subsystem like go-ml exposes:
- **REST:** `POST /v1/ml/generate` (existing)
- **GraphQL:** `mutation { mlGenerate(prompt: "...", backend: "mlx") { response, model } }` (same handler)
- **MCP:** `ml_generate` tool (existing)

Four protocols, one set of handlers.

### Ecosystem Integration

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~gin-contrib/i18n~~ | ~~`WithI18n()`~~ | ~~Locale detection via Accept-Language. Custom middleware using `golang.org/x/text/language`.~~ | ~~**Done**~~ |
| [gin-contrib/graceful](https://github.com/gin-contrib/graceful) | — | Already implemented in Engine.Serve(). Could swap to this for more robust lifecycle management if needed. | — |
| [gin-contrib/requestid](https://github.com/gin-contrib/requestid) | — | Already implemented. Theirs uses UUID, ours uses hex. Could swap for standards compliance. | — |

### Implementation Order

**Wave 1 (gateway hardening):** ~~Authentik, secure, slog, timeout, gzip, static~~ **DONE** (20 Feb 2026)
**Wave 2 (performance + auth):** ~~cache, sessions, authz, brotli~~ **DONE** (20 Feb 2026)
**Wave 3 (network + streaming):** ~~httpsign, sse, location, i18n, gqlgen~~ **DONE** (20 Feb 2026)
**Wave 4 (observability):** ~~pprof, expvar, tracing~~ **DONE** (21 Feb 2026)

Each wave adds `With*()` options + tests. No breaking changes — existing code continues to work without any new options enabled.

## Authentik Integration

[Authentik](https://goauthentik.io/) is the identity provider and edge auth proxy. It handles user registration, login, MFA, social auth, SAML, and OIDC — so go-api doesn't have to.

### Two Integration Modes

**1. Forward Auth (web traffic)**

Traefik sits in front of go-api. For web routes, Traefik's `forwardAuth` middleware checks with Authentik before passing the request through. Authentik handles login flows, session cookies, and consent. go-api receives pre-authenticated requests with identity headers.

```
Browser → Traefik → Authentik (forward auth) → go-api
                         ↓
                    Login page (if unauthenticated)
```

go-api reads trusted headers set by Authentik:
```
X-Authentik-Username: alice
X-Authentik-Groups: admins,developers
X-Authentik-Email: alice@example.com
X-Authentik-Uid: <uuid>
X-Authentik-Jwt: <signed token>
```

**2. OIDC Token Validation (API traffic)**

API clients (SDKs, CLI tools, network peers) authenticate directly with Authentik's OAuth2 token endpoint, then send the JWT to go-api. go-api validates the JWT using Authentik's OIDC discovery endpoint (`.well-known/openid-configuration`).

```
SDK client → Authentik (token endpoint) → receives JWT
SDK client → go-api (Authorization: Bearer <jwt>) → validates via OIDC
```

### Implementation in go-api

```go
engine := api.New(
    api.WithAuthentik(api.AuthentikConfig{
        Issuer:       "https://auth.lthn.ai/application/o/core-api/",
        ClientID:     "core-api",
        TrustedProxy: true,  // Trust X-Authentik-* headers from Traefik
    }),
)
```

`WithAuthentik()` adds middleware that:
1. Checks for `X-Authentik-Jwt` header (forward auth mode) — validates signature, extracts claims
2. Falls back to `Authorization: Bearer <jwt>` header (direct OIDC mode) — validates via JWKS
3. Populates `c.Set("user", AuthentikUser{...})` in the Gin context for handlers to use
4. Skips /health, /swagger, and any public paths

```go
// In any handler:
func (r *Routes) ListItems(c *gin.Context) {
    user := api.GetUser(c)  // Returns *AuthentikUser or nil
    if user == nil {
        c.JSON(401, api.Fail("unauthorised", "Authentication required"))
        return
    }
    // user.Username, user.Groups, user.Email, user.UID available
}
```

### Auth Layers

```
Authentik (identity)     → WHO is this? (user, groups, email)
    ↓
go-api middleware         → IS their token valid? (JWT verification)
    ↓
Casbin authz (optional)  → CAN they do this? (role → endpoint policies)
    ↓
Handler                   → DOES this (business logic)
```

Phase 1 bearer auth continues to work alongside Authentik — useful for service-to-service tokens, CI/CD, and development. `WithBearerAuth` and `WithAuthentik` can coexist.

### Authentik Deployment

Authentik runs as a Docker service alongside go-api, fronted by Traefik:
- **auth.lthn.ai** — Authentik UI + OIDC endpoints (production)
- **auth.leth.in** — Authentik for devnet/testnet
- Traefik routes `/outpost.goauthentik.io/` to Authentik's embedded outpost for forward auth

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/coreos/go-oidc/v3` | OIDC discovery + JWT validation |
| `golang.org/x/oauth2` | OAuth2 token exchange (for server-side flows) |

Both are standard Go libraries with no heavy dependencies.

## Non-Goals

- gRPC gateway
- Built-in user registration/login (Authentik handles this)
- API versioning beyond /v1/ prefix

## Success Criteria

### Phase 1 (Done)

1. ~~`core api serve` starts a Gin server with registered subsystem routes~~
2. ~~WebSocket subscriptions work alongside REST~~
3. ~~Swagger UI accessible at `/swagger/`~~
4. ~~All endpoints return consistent Response envelope~~
5. ~~Bearer token auth protects all routes~~
6. ~~First subsystem integration (go-ml/api/) proves the pattern~~

### Phase 2 (Done)

7. ~~Security headers, compression, and caching active in production~~
8. ~~Session-based auth alongside bearer tokens~~
9. ~~HTTP signature verification for Lethean network peers~~
10. ~~Static file serving for docs site and SDK downloads~~
11. ~~GraphQL endpoint at `/graphql` with playground~~

### Phase 3 (Done)

12. ~~`core api spec` emits valid OpenAPI 3.1 JSON via runtime SpecBuilder~~
13. ~~`core api sdk` generates SDKs for 11 languages via openapi-generator-cli~~
14. ~~MCP tools bridged to REST endpoints via ToolBridge + BridgeToAPI~~
15. ~~OpenAPI spec includes Response[T] envelope wrapping~~
16. ~~Spec export to file in JSON and YAML formats~~
