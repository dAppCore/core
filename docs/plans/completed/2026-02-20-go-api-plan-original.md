# go-api Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build `forge.lthn.ai/core/go-api`, a Gin-based REST framework with OpenAPI generation that subsystems plug into via a RouteGroup interface.

**Architecture:** go-api provides the HTTP engine, middleware stack, response envelope, and OpenAPI tooling. Each ecosystem package (go-ml, go-rag, etc.) imports go-api and registers its own route group. WebSocket support via go-ws Hub runs alongside REST.

**Tech Stack:** Go 1.25, Gin, swaggo/swag, gin-swagger, gin-contrib/cors, go-ws

**Design doc:** `docs/plans/2026-02-20-go-api-design.md`

**Repo location:** `/Users/snider/Code/go-api` (module: `forge.lthn.ai/core/go-api`)

**Licence:** EUPL-1.2

**Convention:** UK English in comments and user-facing strings. Test naming: `_Good`, `_Bad`, `_Ugly`.

---

### Task 1: Scaffold Repository

**Files:**
- Create: `/Users/snider/Code/go-api/go.mod`
- Create: `/Users/snider/Code/go-api/response.go`
- Create: `/Users/snider/Code/go-api/response_test.go`
- Create: `/Users/snider/Code/go-api/LICENCE`

**Step 1: Create repo and go.mod**

```bash
mkdir -p /Users/snider/Code/go-api
cd /Users/snider/Code/go-api
git init
```

Create `go.mod`:
```
module forge.lthn.ai/core/go-api

go 1.25.5

require github.com/gin-gonic/gin v1.10.0
```

Then run:
```bash
go mod tidy
```

**Step 2: Create LICENCE file**

Copy the EUPL-1.2 licence text. Use the same LICENCE file as other ecosystem repos:
```bash
cp /Users/snider/Code/go-ai/LICENCE /Users/snider/Code/go-api/LICENCE
```

**Step 3: Commit scaffold**

```bash
git add go.mod go.sum LICENCE
git commit -m "chore: scaffold go-api module with Gin dependency"
```

---

### Task 2: Response Envelope (TDD)

**Files:**
- Create: `/Users/snider/Code/go-api/response.go`
- Create: `/Users/snider/Code/go-api/response_test.go`

**Step 1: Write the failing tests**

Create `response_test.go`:
```go
package api_test

import (
	"encoding/json"
	"testing"

	api "forge.lthn.ai/core/go-api"
)

func TestOK_Good(t *testing.T) {
	type Payload struct {
		Name string `json:"name"`
	}
	resp := api.OK(Payload{Name: "test"})

	if !resp.Success {
		t.Fatal("expected Success to be true")
	}
	if resp.Data.Name != "test" {
		t.Fatalf("expected Data.Name = test, got %s", resp.Data.Name)
	}
	if resp.Error != nil {
		t.Fatal("expected Error to be nil")
	}
}

func TestFail_Good(t *testing.T) {
	resp := api.Fail("not_found", "Resource not found")

	if resp.Success {
		t.Fatal("expected Success to be false")
	}
	if resp.Error == nil {
		t.Fatal("expected Error to be non-nil")
	}
	if resp.Error.Code != "not_found" {
		t.Fatalf("expected Code = not_found, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "Resource not found" {
		t.Fatalf("expected Message = Resource not found, got %s", resp.Error.Message)
	}
}

func TestFailWithDetails_Good(t *testing.T) {
	details := map[string]string{"field": "email"}
	resp := api.FailWithDetails("validation_error", "Invalid input", details)

	if resp.Error.Details == nil {
		t.Fatal("expected Details to be non-nil")
	}
}

func TestPaginated_Good(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := api.Paginated(items, 1, 10, 42)

	if !resp.Success {
		t.Fatal("expected Success to be true")
	}
	if resp.Meta == nil {
		t.Fatal("expected Meta to be non-nil")
	}
	if resp.Meta.Page != 1 {
		t.Fatalf("expected Page = 1, got %d", resp.Meta.Page)
	}
	if resp.Meta.PerPage != 10 {
		t.Fatalf("expected PerPage = 10, got %d", resp.Meta.PerPage)
	}
	if resp.Meta.Total != 42 {
		t.Fatalf("expected Total = 42, got %d", resp.Meta.Total)
	}
}

func TestOK_JSON_Good(t *testing.T) {
	resp := api.OK("hello")
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if raw["success"] != true {
		t.Fatal("expected success = true in JSON")
	}
	if raw["data"] != "hello" {
		t.Fatalf("expected data = hello, got %v", raw["data"])
	}
	// error and meta should be omitted
	if _, ok := raw["error"]; ok {
		t.Fatal("expected error to be omitted from JSON")
	}
	if _, ok := raw["meta"]; ok {
		t.Fatal("expected meta to be omitted from JSON")
	}
}

func TestFail_JSON_Good(t *testing.T) {
	resp := api.Fail("err", "msg")
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if raw["success"] != false {
		t.Fatal("expected success = false in JSON")
	}
	// data should be omitted
	if _, ok := raw["data"]; ok {
		t.Fatal("expected data to be omitted from JSON")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: Compilation errors — `api.OK`, `api.Fail`, etc. not defined.

**Step 3: Implement response.go**

Create `response.go`:
```go
// Package api provides a Gin-based REST framework with OpenAPI generation.
// Subsystems implement RouteGroup to register their own endpoints.
package api

// Response is the standard envelope for all API responses.
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Error describes a failed API request.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Meta carries pagination and request metadata.
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Total     int    `json:"total,omitempty"`
}

// OK returns a successful response wrapping data.
func OK[T any](data T) Response[T] {
	return Response[T]{Success: true, Data: data}
}

// Fail returns an error response with code and message.
func Fail(code, message string) Response[any] {
	return Response[any]{
		Success: false,
		Error:   &Error{Code: code, Message: message},
	}
}

// FailWithDetails returns an error response with additional detail payload.
func FailWithDetails(code, message string, details any) Response[any] {
	return Response[any]{
		Success: false,
		Error:   &Error{Code: code, Message: message, Details: details},
	}
}

// Paginated returns a successful response with pagination metadata.
func Paginated[T any](data T, page, perPage, total int) Response[T] {
	return Response[T]{
		Success: true,
		Data:    data,
		Meta:    &Meta{Page: page, PerPage: perPage, Total: total},
	}
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: All 6 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/go-api
git add response.go response_test.go
git commit -m "feat: add response envelope with OK, Fail, Paginated helpers"
```

---

### Task 3: RouteGroup Interface

**Files:**
- Create: `/Users/snider/Code/go-api/group.go`
- Create: `/Users/snider/Code/go-api/group_test.go`

**Step 1: Write the failing test**

Create `group_test.go`:
```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	api "forge.lthn.ai/core/go-api"
	"github.com/gin-gonic/gin"
)

// stubGroup is a minimal RouteGroup for testing.
type stubGroup struct{}

func (s *stubGroup) Name() string     { return "stub" }
func (s *stubGroup) BasePath() string { return "/v1/stub" }

func (s *stubGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(200, api.OK("pong"))
	})
}

// stubStreamGroup implements both RouteGroup and StreamGroup.
type stubStreamGroup struct {
	stubGroup
}

func (s *stubStreamGroup) Channels() []string {
	return []string{"stub.events", "stub.updates"}
}

func TestRouteGroup_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := gin.New()
	group := &stubGroup{}

	rg := g.Group(group.BasePath())
	group.RegisterRoutes(rg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/stub/ping", nil)
	g.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestStreamGroup_Good(t *testing.T) {
	group := &stubStreamGroup{}

	// Verify it satisfies StreamGroup
	var sg api.StreamGroup = group
	channels := sg.Channels()

	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if channels[0] != "stub.events" {
		t.Fatalf("expected stub.events, got %s", channels[0])
	}
}

func TestRouteGroupName_Good(t *testing.T) {
	group := &stubGroup{}

	var rg api.RouteGroup = group
	if rg.Name() != "stub" {
		t.Fatalf("expected name stub, got %s", rg.Name())
	}
	if rg.BasePath() != "/v1/stub" {
		t.Fatalf("expected basepath /v1/stub, got %s", rg.BasePath())
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: Compilation errors — `api.RouteGroup`, `api.StreamGroup` not defined.

**Step 3: Implement group.go**

Create `group.go`:
```go
package api

import "github.com/gin-gonic/gin"

// RouteGroup registers API routes onto a Gin router group.
// Subsystems implement this to expose their REST endpoints.
type RouteGroup interface {
	// Name returns the route group identifier (e.g. "ml", "rag", "tasks").
	Name() string
	// BasePath returns the URL prefix (e.g. "/v1/ml").
	BasePath() string
	// RegisterRoutes adds handlers to the provided router group.
	RegisterRoutes(rg *gin.RouterGroup)
}

// StreamGroup optionally declares WebSocket channels a subsystem publishes to.
// Subsystems implementing both RouteGroup and StreamGroup expose both REST
// endpoints and real-time event channels.
type StreamGroup interface {
	// Channels returns the WebSocket channel names this group publishes to.
	Channels() []string
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: All tests PASS (previous 6 + new 3).

**Step 5: Commit**

```bash
cd /Users/snider/Code/go-api
git add group.go group_test.go
git commit -m "feat: add RouteGroup and StreamGroup interfaces"
```

---

### Task 4: Engine + Options (TDD)

**Files:**
- Create: `/Users/snider/Code/go-api/api.go`
- Create: `/Users/snider/Code/go-api/options.go`
- Create: `/Users/snider/Code/go-api/api_test.go`

**Step 1: Write the failing tests**

Create `api_test.go`:
```go
package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "forge.lthn.ai/core/go-api"
	"github.com/gin-gonic/gin"
)

func TestNew_Good(t *testing.T) {
	engine, err := api.New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestNewWithAddr_Good(t *testing.T) {
	engine, err := api.New(api.WithAddr(":9090"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if engine.Addr() != ":9090" {
		t.Fatalf("expected addr :9090, got %s", engine.Addr())
	}
}

func TestDefaultAddr_Good(t *testing.T) {
	engine, _ := api.New()
	if engine.Addr() != ":8080" {
		t.Fatalf("expected default addr :8080, got %s", engine.Addr())
	}
}

func TestRegister_Good(t *testing.T) {
	engine, _ := api.New()
	group := &stubGroup{}

	engine.Register(group)

	if len(engine.Groups()) != 1 {
		t.Fatalf("expected 1 group, got %d", len(engine.Groups()))
	}
	if engine.Groups()[0].Name() != "stub" {
		t.Fatalf("expected group name stub, got %s", engine.Groups()[0].Name())
	}
}

func TestRegisterMultiple_Good(t *testing.T) {
	engine, _ := api.New()
	engine.Register(&stubGroup{})
	engine.Register(&stubStreamGroup{})

	if len(engine.Groups()) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(engine.Groups()))
	}
}

func TestHandler_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New()
	engine.Register(&stubGroup{})

	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/stub/ping", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Fatal("expected success = true")
	}
}

func TestHealthEndpoint_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New()
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestServeAndShutdown_Good(t *testing.T) {
	engine, _ := api.New(api.WithAddr(":0"))
	engine.Register(&stubGroup{})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- engine.Serve(ctx)
	}()

	// Wait for context cancellation to trigger shutdown
	<-ctx.Done()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed && err != context.DeadlineExceeded {
			t.Fatalf("Serve() returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() did not return after context cancellation")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: Compilation errors — `api.New`, `api.WithAddr`, `api.Engine` not defined.

**Step 3: Implement options.go**

Create `options.go`:
```go
package api

// Option configures the Engine.
type Option func(*Engine) error

// WithAddr sets the listen address (default ":8080").
func WithAddr(addr string) Option {
	return func(e *Engine) error {
		e.addr = addr
		return nil
	}
}
```

**Step 4: Implement api.go**

Create `api.go`:
```go
package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Engine is the central REST API server.
// Register RouteGroups to add endpoints, then call Serve to start.
type Engine struct {
	gin    *gin.Engine
	addr   string
	groups []RouteGroup
	logger *slog.Logger
	built  bool
}

// New creates an Engine with the given options.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		addr:   ":8080",
		logger: slog.Default(),
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return e, nil
}

// Addr returns the configured listen address.
func (e *Engine) Addr() string {
	return e.addr
}

// Groups returns all registered route groups.
func (e *Engine) Groups() []RouteGroup {
	return e.groups
}

// Register adds a RouteGroup to the engine.
// Routes are mounted when Handler() or Serve() is called.
func (e *Engine) Register(group RouteGroup) {
	e.groups = append(e.groups, group)
	e.built = false
}

// build constructs the Gin engine with all registered groups.
func (e *Engine) build() {
	if e.built && e.gin != nil {
		return
	}

	e.gin = gin.New()
	e.gin.Use(gin.Recovery())

	// Health endpoint
	e.gin.GET("/health", func(c *gin.Context) {
		c.JSON(200, OK("healthy"))
	})

	// Mount each route group
	for _, group := range e.groups {
		rg := e.gin.Group(group.BasePath())
		group.RegisterRoutes(rg)
		e.logger.Info("registered route group", "name", group.Name(), "path", group.BasePath())
	}

	e.built = true
}

// Handler returns the http.Handler for testing or custom server usage.
func (e *Engine) Handler() http.Handler {
	e.build()
	return e.gin
}

// Serve starts the HTTP server and blocks until the context is cancelled.
// Performs graceful shutdown on context cancellation.
func (e *Engine) Serve(ctx context.Context) error {
	e.build()

	srv := &http.Server{
		Addr:    e.addr,
		Handler: e.gin,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5_000_000_000) // 5s
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	if err, ok := <-errCh; ok {
		return err
	}

	return nil
}
```

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1
```

Expected: All tests PASS.

**Step 6: Commit**

```bash
cd /Users/snider/Code/go-api
git add api.go options.go api_test.go
git commit -m "feat: add Engine with Register, Handler, Serve, and graceful shutdown"
```

---

### Task 5: Middleware (TDD)

**Files:**
- Create: `/Users/snider/Code/go-api/middleware.go`
- Create: `/Users/snider/Code/go-api/middleware_test.go`
- Modify: `/Users/snider/Code/go-api/options.go` — add middleware options

**Step 1: Write the failing tests**

Create `middleware_test.go`:
```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	api "forge.lthn.ai/core/go-api"
	"github.com/gin-gonic/gin"
)

func TestBearerAuth_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithBearerAuth("secret-token"))
	engine.Register(&stubGroup{})
	handler := engine.Handler()

	// Request without token → 401
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/stub/ping", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}

	// Request with correct token → 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/stub/ping", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 with correct token, got %d", w.Code)
	}
}

func TestBearerAuth_Bad(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithBearerAuth("secret-token"))
	engine.Register(&stubGroup{})
	handler := engine.Handler()

	// Wrong token → 401
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/stub/ping", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	handler.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 with wrong token, got %d", w.Code)
	}
}

func TestHealthBypassesAuth_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithBearerAuth("secret-token"))
	handler := engine.Handler()

	// Health endpoint should not require auth
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for /health without auth, got %d", w.Code)
	}
}

func TestRequestID_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithRequestID())
	engine.Register(&stubGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/stub/ping", nil)
	handler.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRequestIDPreserved_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithRequestID())
	engine.Register(&stubGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/stub/ping", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	handler.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-ID")
	if rid != "my-custom-id" {
		t.Fatalf("expected X-Request-ID = my-custom-id, got %s", rid)
	}
}

func TestCORS_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithCORS("https://example.com"))
	engine.Register(&stubGroup{})
	handler := engine.Handler()

	// Preflight request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/v1/stub/ping", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Fatalf("expected CORS origin https://example.com, got %s", origin)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: Compilation errors — `WithBearerAuth`, `WithRequestID`, `WithCORS` not defined.

**Step 3: Implement middleware.go**

Create `middleware.go`:
```go
package api

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
)

// bearerAuthMiddleware validates Bearer tokens.
// Skips paths listed in skip (e.g. /health, /swagger).
func bearerAuthMiddleware(token string, skip []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, s := range skip {
			if strings.HasPrefix(path, s) {
				c.Next()
				return
			}
		}

		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(401, Fail("unauthorised", "Missing Authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != token {
			c.JSON(401, Fail("unauthorised", "Invalid bearer token"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// requestIDMiddleware sets X-Request-ID on every response.
// If the client sends one, it is preserved; otherwise a random ID is generated.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			b := make([]byte, 16)
			rand.Read(b)
			rid = hex.EncodeToString(b)
		}
		c.Header("X-Request-ID", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}
```

**Step 4: Add middleware options to options.go**

Append to `options.go`:
```go
import "github.com/gin-contrib/cors"

// WithBearerAuth adds bearer token authentication middleware.
// The /health and /swagger paths are excluded from authentication.
func WithBearerAuth(token string) Option {
	return func(e *Engine) error {
		e.middlewares = append(e.middlewares, bearerAuthMiddleware(token, []string{"/health", "/swagger"}))
		return nil
	}
}

// WithRequestID adds a middleware that sets X-Request-ID on every response.
func WithRequestID() Option {
	return func(e *Engine) error {
		e.middlewares = append(e.middlewares, requestIDMiddleware())
		return nil
	}
}

// WithCORS configures Cross-Origin Resource Sharing.
// Pass "*" to allow all origins, or specific origins.
func WithCORS(allowOrigins ...string) Option {
	return func(e *Engine) error {
		config := cors.DefaultConfig()
		if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
			config.AllowAllOrigins = true
		} else {
			config.AllowOrigins = allowOrigins
		}
		config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
		config.AllowHeaders = []string{"Authorization", "Content-Type", "X-Request-ID"}
		e.middlewares = append(e.middlewares, cors.New(config))
		return nil
	}
}
```

Update `Engine` struct in `api.go` to include `middlewares []gin.HandlerFunc` field, and apply them in `build()`:
```go
// Add to Engine struct:
middlewares []gin.HandlerFunc

// In build(), after gin.New() and gin.Recovery(), before health endpoint:
for _, mw := range e.middlewares {
    e.gin.Use(mw)
}
```

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1
```

Expected: All tests PASS.

**Step 6: Commit**

```bash
cd /Users/snider/Code/go-api
git add middleware.go middleware_test.go options.go api.go
git commit -m "feat: add bearer auth, request ID, and CORS middleware"
```

---

### Task 6: WebSocket Integration (TDD)

**Files:**
- Create: `/Users/snider/Code/go-api/websocket.go`
- Create: `/Users/snider/Code/go-api/websocket_test.go`
- Modify: `/Users/snider/Code/go-api/options.go` — add WithWSHub
- Modify: `/Users/snider/Code/go-api/api.go` — mount /ws route

**Step 1: Write the failing test**

Create `websocket_test.go`:
```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "forge.lthn.ai/core/go-api"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestWSEndpoint_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithWSHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	})))

	srv := httptest.NewServer(engine.Handler())
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(msg) != "hello" {
		t.Fatalf("expected hello, got %s", string(msg))
	}
}

func TestNoWSHandler_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New()
	handler := engine.Handler()

	// /ws should 404 when no handler configured
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404 without WS handler, got %d", w.Code)
	}
}

func TestChannelListing_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New()
	engine.Register(&stubStreamGroup{})

	channels := engine.Channels()
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: Compilation errors.

**Step 3: Implement websocket.go + option + engine changes**

Create `websocket.go`:
```go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// wrapWSHandler adapts a standard http.Handler to a Gin handler for the /ws route.
func wrapWSHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
```

Add to `options.go`:
```go
// WithWSHandler registers a WebSocket handler at GET /ws.
// Typically this wraps a go-ws Hub.Handler().
func WithWSHandler(h http.Handler) Option {
	return func(e *Engine) error {
		e.wsHandler = h
		return nil
	}
}
```

Add to `Engine` struct in `api.go`:
```go
wsHandler http.Handler
```

Add to `build()` after mounting route groups:
```go
// WebSocket endpoint
if e.wsHandler != nil {
    e.gin.GET("/ws", wrapWSHandler(e.wsHandler))
}
```

Add `Channels()` method to `Engine`:
```go
// Channels returns all WebSocket channel names from registered StreamGroups.
func (e *Engine) Channels() []string {
	var channels []string
	for _, g := range e.groups {
		if sg, ok := g.(StreamGroup); ok {
			channels = append(channels, sg.Channels()...)
		}
	}
	return channels
}
```

**Step 4: Run go mod tidy to pick up gorilla/websocket**

```bash
cd /Users/snider/Code/go-api
go mod tidy
```

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1
```

Expected: All tests PASS.

**Step 6: Commit**

```bash
cd /Users/snider/Code/go-api
git add websocket.go websocket_test.go options.go api.go go.mod go.sum
git commit -m "feat: add WebSocket endpoint and channel listing from StreamGroups"
```

---

### Task 7: Swagger/OpenAPI Integration

**Files:**
- Create: `/Users/snider/Code/go-api/swagger.go`
- Create: `/Users/snider/Code/go-api/swagger_test.go`
- Modify: `/Users/snider/Code/go-api/options.go` — add WithSwagger
- Modify: `/Users/snider/Code/go-api/api.go` — mount swagger routes

**Step 1: Write the failing test**

Create `swagger_test.go`:
```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	api "forge.lthn.ai/core/go-api"
	"github.com/gin-gonic/gin"
)

func TestSwaggerEndpoint_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithSwagger("Core API", "REST API for the Lethean ecosystem", "0.1.0"))
	engine.Register(&stubGroup{})
	handler := engine.Handler()

	// Swagger JSON endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/doc.json", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for swagger doc.json, got %d", w.Code)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty swagger doc")
	}
}

func TestSwaggerDisabledByDefault_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New()
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/doc.json", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404 when swagger disabled, got %d", w.Code)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v
```

Expected: Compilation errors.

**Step 3: Implement swagger.go + option**

Create `swagger.go`:
```go
package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

// swaggerSpec holds a minimal OpenAPI spec for runtime serving.
type swaggerSpec struct {
	title       string
	description string
	version     string
}

func (s *swaggerSpec) ReadDoc() string {
	// Minimal OpenAPI 3.0 document — swaggo generates the full one at build time.
	// This serves as the runtime fallback and base template.
	return `{
	"swagger": "2.0",
	"info": {
		"title": "` + s.title + `",
		"description": "` + s.description + `",
		"version": "` + s.version + `"
	},
	"basePath": "/",
	"paths": {}
}`
}

// registerSwagger mounts the swagger UI and doc.json endpoint.
func registerSwagger(g *gin.Engine, title, description, version string) {
	spec := &swaggerSpec{title: title, description: description, version: version}
	swag.Register(swag.Name, spec)

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

Add to `options.go`:
```go
// WithSwagger enables the Swagger UI at /swagger/.
func WithSwagger(title, description, version string) Option {
	return func(e *Engine) error {
		e.swaggerTitle = title
		e.swaggerDesc = description
		e.swaggerVersion = version
		e.swaggerEnabled = true
		return nil
	}
}
```

Add fields to `Engine` struct:
```go
swaggerEnabled bool
swaggerTitle   string
swaggerDesc    string
swaggerVersion string
```

Add to `build()` after WebSocket:
```go
// Swagger UI
if e.swaggerEnabled {
    registerSwagger(e.gin, e.swaggerTitle, e.swaggerDesc, e.swaggerVersion)
}
```

**Step 4: Run go mod tidy**

```bash
cd /Users/snider/Code/go-api
go get github.com/swaggo/gin-swagger github.com/swaggo/files github.com/swaggo/swag
go mod tidy
```

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1
```

Expected: All tests PASS.

**Step 6: Commit**

```bash
cd /Users/snider/Code/go-api
git add swagger.go swagger_test.go options.go api.go go.mod go.sum
git commit -m "feat: add Swagger UI endpoint with runtime spec serving"
```

---

### Task 8: CLAUDE.md + README.md

**Files:**
- Create: `/Users/snider/Code/go-api/CLAUDE.md`
- Create: `/Users/snider/Code/go-api/README.md`

**Step 1: Write CLAUDE.md**

```markdown
# CLAUDE.md

This file provides guidance to Claude Code when working with the go-api repository.

## Project Overview

**go-api** is the REST framework for the Lethean Go ecosystem. It provides a Gin-based HTTP engine with middleware, response envelopes, WebSocket integration, and OpenAPI generation. Subsystems implement the `RouteGroup` interface to register their own endpoints.

- **Module path**: `forge.lthn.ai/core/go-api`
- **Language**: Go 1.25
- **Licence**: EUPL-1.2

## Build & Test Commands

```bash
go test ./...                       # Run all tests
go test -run TestName ./...         # Run a single test
go test -v -race ./...              # Verbose with race detector
go build ./...                      # Build (library — no main package)
go vet ./...                        # Vet
```

## Coding Standards

- **UK English** in comments and user-facing strings (colour, organisation, unauthorised)
- **Conventional commits**: `type(scope): description`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Error handling**: Return wrapped errors with context, never panic
- **Test naming**: `_Good` (happy path), `_Bad` (expected errors), `_Ugly` (panics/edge cases)
- **Licence**: EUPL-1.2
```

**Step 2: Write README.md**

Brief README with quick start and links to design doc.

**Step 3: Commit**

```bash
cd /Users/snider/Code/go-api
git add CLAUDE.md README.md
git commit -m "docs: add CLAUDE.md and README.md"
```

---

### Task 9: Create Forge Repo + Push

**Step 1: Create repo on Forge**

```bash
curl -s -X POST "https://forge.lthn.ai/api/v1/orgs/core/repos" \
  -H "Authorization: token 375068d101922dd1cf269e8b8cb77a0f99d1b486" \
  -H "Content-Type: application/json" \
  -d '{"name":"go-api","description":"REST framework + OpenAPI SDK generation for the Lethean Go ecosystem","default_branch":"main","auto_init":false,"license":"EUPL-1.2"}'
```

**Step 2: Add remote and push**

```bash
cd /Users/snider/Code/go-api
git remote add forge ssh://git@forge.lthn.ai:2223/core/go-api.git
git branch -M main
git push -u forge main
```

**Step 3: Verify on Forge**

```bash
curl -s "https://forge.lthn.ai/api/v1/repos/core/go-api" \
  -H "Authorization: token 375068d101922dd1cf269e8b8cb77a0f99d1b486" | jq .name
```

Expected: `"go-api"`

---

### Task 10: Integration Test — First Subsystem (go-ml/api)

This task validates the framework by building the first real subsystem integration. It lives in go-ml, not go-api.

**Files:**
- Create: `/Users/snider/Code/go-ml/api/routes.go`
- Create: `/Users/snider/Code/go-ml/api/routes_test.go`

**Step 1: Write the failing test in go-ml**

Create `api/routes_test.go` in go-ml that:
1. Creates a `Routes` with a mock `ml.Service`
2. Registers it on an `api.Engine`
3. Sends `POST /v1/ml/backends` and asserts a 200 response with the response envelope

**Step 2: Implement api/routes.go**

Implement `Routes` struct that wraps `*ml.Service` and exposes:
- `POST /v1/ml/generate`
- `POST /v1/ml/score`
- `GET /v1/ml/backends`
- `GET /v1/ml/status`

Each handler uses `c.ShouldBindJSON()` for input and `api.OK()` / `api.Fail()` for responses.

**Step 3: Run tests**

```bash
cd /Users/snider/Code/go-ml
go test ./api/... -v
```

**Step 4: Commit in go-ml**

```bash
cd /Users/snider/Code/go-ml
git add api/
git commit -m "feat(api): add REST route group for ML endpoints via go-api"
```

---

## Dependency Summary

```
Task 1 (scaffold) → Task 2 (response) → Task 3 (group) → Task 4 (engine)
    → Task 5 (middleware) → Task 6 (websocket) → Task 7 (swagger)
    → Task 8 (docs) → Task 9 (forge) → Task 10 (integration)
```

All tasks are sequential — each builds on the previous.

## Estimated Timeline

- Tasks 1-7: Core go-api package (~820 LOC)
- Task 8: Documentation
- Task 9: Forge deployment
- Task 10: First subsystem integration proof
