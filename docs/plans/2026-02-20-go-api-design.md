# go-api Design — HTTP Gateway + OpenAPI SDK Generation

**Date:** 2026-02-20
**Author:** Virgil
**Status:** Phase 1 Implemented, Phase 2 Planned
**Module:** `forge.lthn.ai/core/go-api`

## Problem

The Core Go ecosystem exposes 42+ tools via MCP (JSON-RPC), which is ideal for AI agents but inaccessible to regular HTTP clients, frontend applications, and third-party integrators. There is no unified HTTP gateway, no OpenAPI specification, and no generated SDKs.

Both external customers (Host UK products) and Lethean network peers need programmatic access to the same services. The gateway also serves web routes, static assets, and streaming endpoints — not just REST APIs.

## Solution

A `go-api` package that acts as the central HTTP gateway:

1. **Gin-based HTTP gateway** with extensible middleware via gin-contrib plugins
2. **RouteGroup interface** that subsystems implement to register their own endpoints (API, web, or both)
3. **WebSocket + SSE integration** for real-time streaming
4. **OpenAPI 3.1 spec generation** via swaggo annotations
5. **SDK generation pipeline** targeting Python, TypeScript, Go client, and more

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

## Phase 2 — Remaining Gin Plugin Roadmap

All plugins drop in as `With*()` options on the Engine. No architecture changes needed.

### Security & Auth

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~**Authentik**~~ | ~~`WithAuthentik()`~~ | ~~OIDC + forward auth integration.~~ | ~~**Done**~~ |
| ~~gin-contrib/secure~~ | ~~`WithSecure()`~~ | ~~Security headers: HSTS, X-Frame-Options, X-Content-Type-Options, CSP.~~ | ~~**Done**~~ |
| [gin-contrib/sessions](https://github.com/gin-contrib/sessions) | `WithSessions()` | Server-side sessions (cookie, Redis, memcached). For web session management alongside Authentik tokens. | High |
| [gin-contrib/authz](https://github.com/gin-contrib/authz) | `WithAuthz()` | Casbin-based authorisation. Policy-driven access control — define who can call which endpoints. Complements Authentik's identity with fine-grained permissions. | Medium |
| [gin-contrib/httpsign](https://github.com/gin-contrib/httpsign) | `WithHTTPSign()` | HTTP signature verification. Maps to UEPS Ed25519 consent tokens for Lethean network peer authentication. | Medium |

### Performance & Reliability

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| [gin-contrib/cache](https://github.com/gin-contrib/cache) | `WithCache()` | Response caching (in-memory, Redis). Huge for gateway performance — cache GET responses, invalidate on writes. | High |
| ~~gin-contrib/timeout~~ | ~~`WithTimeout()`~~ | ~~Per-request timeouts.~~ | ~~**Done**~~ |
| ~~gin-contrib/gzip~~ | ~~`WithGzip()`~~ | ~~Gzip response compression.~~ | ~~**Done**~~ |
| [gin-contrib/brotli](https://github.com/gin-contrib/brotli) | `WithBrotli()` | Brotli compression. Better ratios than gzip for text/HTML. Can use alongside gzip with content negotiation. | Medium |

### Observability

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~gin-contrib/slog~~ | ~~`WithSlog()`~~ | ~~Structured request logging via slog.~~ | ~~**Done**~~ |
| [gin-contrib/pprof](https://github.com/gin-contrib/pprof) | `WithPprof()` | Runtime profiling endpoints at /debug/pprof/. Dev/staging only, gate behind auth. | Low |
| [gin-contrib/expvar](https://github.com/gin-contrib/expvar) | `WithExpvar()` | Go runtime metrics (goroutines, memstats). Dev/debug only. | Low |
| [opengintracing](https://github.com/gin-contrib/opengintracing) | `WithTracing()` | OpenTelemetry distributed tracing. Needed when multiple services form a request chain. | Low |

### Content & Streaming

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| ~~gin-contrib/static~~ | ~~`WithStatic()`~~ | ~~Serve static files.~~ | ~~**Done**~~ |
| [gin-contrib/sse](https://github.com/gin-contrib/sse) | `WithSSE()` | Server-Sent Events. One-way streaming alternative to WebSocket — ideal for ML generation progress, live logs, event feeds. | Medium |
| [gin-contrib/location](https://github.com/gin-contrib/location) | `WithLocation()` | Auto-detect scheme/host from request headers (X-Forwarded-*). Needed behind reverse proxy. | Medium |

### Query Layer

| Plugin | Option | Purpose | Priority |
|--------|--------|---------|----------|
| [99designs/gqlgen](https://github.com/99designs/gqlgen) | `WithGraphQL()` | GraphQL endpoint at `/graphql` + playground at `/graphql/playground`. Code-first: generates resolvers from Go types. Same backend services as REST — subsystems implement a `ResolverGroup` interface alongside `RouteGroup`. Clients choose REST or GraphQL per preference. | Medium |

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
| [gin-contrib/i18n](https://github.com/gin-contrib/i18n) | `WithI18n()` | Localised API responses. Bridge to go-i18n grammar engine for localised error messages — unique differentiator. | Medium |
| [gin-contrib/graceful](https://github.com/gin-contrib/graceful) | — | Already implemented in Engine.Serve(). Could swap to this for more robust lifecycle management if needed. | — |
| [gin-contrib/requestid](https://github.com/gin-contrib/requestid) | — | Already implemented. Theirs uses UUID, ours uses hex. Could swap for standards compliance. | — |

### Implementation Order

**Wave 1 (gateway hardening):** ~~Authentik, secure, slog, timeout, gzip, static~~ **DONE** (20 Feb 2026)
**Wave 2 (performance + auth):** cache, sessions, authz, brotli
**Wave 3 (network + streaming):** httpsign, sse, location, i18n, gqlgen
**Wave 4 (observability):** pprof, expvar, opengintracing

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
- Automatic MCP-to-REST adapter (can be added later)
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

### Phase 2

7. `core api spec` emits valid OpenAPI 3.1 JSON from swaggo annotations
8. Generated Python and TypeScript SDKs can call endpoints successfully
9. Security headers, compression, and caching active in production
10. Session-based auth alongside bearer tokens
11. HTTP signature verification for Lethean network peers
12. Static file serving for docs site and SDK downloads
13. GraphQL endpoint at `/graphql` with playground, subsystem resolvers via ResolverGroup interface
