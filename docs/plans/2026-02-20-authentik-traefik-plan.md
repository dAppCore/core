# Authentik + Traefik Integration Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deploy Authentik as the identity provider, wire it into Traefik's forward auth, and add OIDC/header middleware to go-api so protected services get authenticated user context.

**Architecture:** Authentik runs alongside existing services on de2 (production). Traefik's file provider loads a `forwardAuth` middleware definition pointing at Authentik's outpost. Services opt-in via Docker label `middlewares: authentik@file`. go-api gains a `WithAuthentik()` option that extracts user identity from Authentik headers (forward auth mode) or validates JWTs directly (API client mode).

**Tech Stack:** Authentik 2025.2, Traefik v3.6, Go 1.25, coreos/go-oidc/v3, golang.org/x/oauth2

**Design doc:** `docs/plans/2026-02-20-go-api-design.md` (Authentik section)

**Key references:**
- Traefik role: `/Users/snider/Code/DevOps/roles/traefik/`
- Authentik role: `/Users/snider/Code/DevOps/roles/authentik/`
- Forward auth template: `/Users/snider/Code/DevOps/roles/traefik/templates/dynamic-authentik.yml.j2`
- go-api repo: `/Users/snider/Code/go-api/`

---

## Current State

The Ansible infrastructure is **already built but not activated**:

| Component | Status | Location |
|-----------|--------|----------|
| Traefik v3.6 role | Deployed on de2 | `roles/traefik/` |
| Authentik 2025.2 role | Written, **never deployed** | `roles/authentik/` |
| Forward auth middleware template | Written, conditional on `traefik_authentik_enabled` | `dynamic-authentik.yml.j2` |
| Outpost routing in Authentik compose | Pre-configured | `roles/authentik/templates/docker-compose.yml.j2` |
| 5 services with `authentik@file` | Labels present, middleware not yet available | `prod_rebuild.yml` |
| go-api Authentik middleware | **Not started** | — |

**Headers Authentik will pass to go-api (via Traefik):**
```
X-authentik-username, X-authentik-groups, X-authentik-entitlements,
X-authentik-email, X-authentik-name, X-authentik-uid, X-authentik-jwt,
X-authentik-meta-jwks, X-authentik-meta-outpost, X-authentik-meta-provider,
X-authentik-meta-app, X-authentik-meta-version
```

---

### Task 1: Enable Authentik in Production Inventory

This task sets the Ansible variables to enable Authentik deployment on the production host.

**Files:**
- Modify: `/Users/snider/Code/DevOps/inventory/host_vars/de2.yml` (or equivalent group_vars)

**Step 1: Find the correct inventory file for de2**

Run:
```bash
find /Users/snider/Code/DevOps/inventory -name "*.yml" -o -name "*.yaml" | head -20
ls /Users/snider/Code/DevOps/inventory/
```

Identify where de2's host vars live.

**Step 2: Add Authentik variables**

Add these variables for the de2 host:

```yaml
# Authentik
traefik_authentik_enabled: true
traefik_authentik_url: "https://auth.host.uk.com"

authentik_host: "auth.host.uk.com"
authentik_bootstrap_password: "<generate with: openssl rand -hex 32>"
authentik_bootstrap_token: "<generate with: openssl rand -hex 32>"
authentik_bootstrap_email: "admin@host.uk.com"
```

Note: `authentik_secret_key` auto-generates and persists on first run. `authentik_pg_password` auto-generates via lookup. The Authentik role handles both.

**Step 3: Verify prerequisites exist on de2**

Authentik requires PostgreSQL + Dragonfly (Redis). Check they're in the prod playbook:
```bash
grep -n "postgres\|dragonfly" /Users/snider/Code/DevOps/playbooks/prod_rebuild.yml | head -10
```

**Step 4: Commit**

```bash
cd /Users/snider/Code/DevOps
git add inventory/
git commit -m "feat(authentik): enable Authentik and Traefik forward auth on de2

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 2: Add Authentik to Production Playbook

The Authentik Ansible role exists but is not included in the prod rebuild playbook. This task adds it.

**Files:**
- Modify: `/Users/snider/Code/DevOps/playbooks/prod_rebuild.yml`

**Step 1: Read the playbook to find the right insertion point**

Authentik must deploy AFTER PostgreSQL + Dragonfly (it needs them) and AFTER Traefik (it needs the proxy network), but BEFORE services that use `authentik@file`.

```bash
grep -n "Phase\|traefik\|postgres\|dragonfly\|portainer\|glance" /Users/snider/Code/DevOps/playbooks/prod_rebuild.yml | head -20
```

**Step 2: Add Authentik role include**

Insert after the Traefik phase, before services:

```yaml
    # ── Phase N: Identity (Authentik) ──
    - name: Deploy Authentik
      ansible.builtin.include_role:
        name: authentik
      tags: [authentik]
```

**Step 3: Verify the playbook parses**

```bash
cd /Users/snider/Code/DevOps
ansible-playbook playbooks/prod_rebuild.yml --syntax-check
```

Expected: No errors.

**Step 4: Commit**

```bash
cd /Users/snider/Code/DevOps
git add playbooks/prod_rebuild.yml
git commit -m "feat(authentik): add Authentik phase to prod rebuild playbook

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 3: Deploy Authentik (Run Playbook)

This is a manual step — run the Ansible playbook to deploy Authentik on de2.

**Step 1: Dry-run the Authentik tag only**

```bash
cd /Users/snider/Code/DevOps
ansible-playbook playbooks/prod_rebuild.yml --tags authentik --check --diff
```

Review the output. Expect: directories created, docker-compose deployed, containers started.

Note: `--check` will skip shell/command tasks (like the PostgreSQL user creation). This is expected — the actual run will handle those.

**Step 2: Deploy Authentik**

```bash
ansible-playbook playbooks/prod_rebuild.yml --tags authentik
```

**Step 3: Re-deploy Traefik to pick up the forward auth middleware**

The Traefik role conditionally deploys `dynamic-authentik.yml` based on `traefik_authentik_enabled`. Re-running the role with the new variable will create the middleware file:

```bash
ansible-playbook playbooks/prod_rebuild.yml --tags traefik
```

**Step 4: Verify Authentik is accessible**

```bash
curl -sI https://auth.host.uk.com | head -5
```

Expected: HTTP 200 or 302 redirect to login page.

**Step 5: Complete initial setup**

Open `https://auth.host.uk.com/if/flow/initial-setup/` in a browser. Set the admin password (the bootstrap password from Task 1 is used for the API token, but the UI setup flow creates the actual admin account).

---

### Task 4: Create Authentik OIDC Application for go-api

This configures Authentik to issue tokens for go-api. Done via the Authentik admin UI or API.

**Step 1: Create an OAuth2/OIDC Provider**

In Authentik Admin → Providers → Create:

| Field | Value |
|-------|-------|
| Name | `Core API` |
| Protocol | OAuth2/OIDC |
| Client type | Confidential |
| Client ID | `core-api` |
| Redirect URIs | `https://api.lthn.ai/auth/callback` (for auth code flow) |
| Signing key | Select auto-generated signing key |
| Scopes | `openid`, `email`, `profile` |
| Subject mode | Based on user's hashed ID |

Record the **Client Secret** — needed for go-api config.

**Step 2: Create an Application**

In Authentik Admin → Applications → Create:

| Field | Value |
|-------|-------|
| Name | `Core API` |
| Slug | `core-api` |
| Provider | Core API (from step 1) |
| Launch URL | `https://api.lthn.ai/` |

**Step 3: Create a Forward Auth (Proxy) Provider for Traefik**

In Authentik Admin → Providers → Create:

| Field | Value |
|-------|-------|
| Name | `Traefik Forward Auth — Core API` |
| Protocol | Proxy |
| Mode | Forward auth (single application) |
| External host | `https://api.lthn.ai` |

**Step 4: Create an Outpost (if not exists)**

In Authentik Admin → Outposts:
- If no outpost exists: Create → Type: Proxy, Integration: Local Docker
- Add both providers to the outpost

**Step 5: Test forward auth is working**

```bash
# This should redirect to Authentik login
curl -sI https://api.lthn.ai/
```

Once authenticated, Traefik passes the X-authentik-* headers through.

---

### Task 5: go-api Authentik User Type (TDD)

**Files:**
- Create: `/Users/snider/Code/go-api/authentik.go`
- Create: `/Users/snider/Code/go-api/authentik_test.go`

**Step 1: Write the failing tests**

Create `authentik_test.go`:
```go
package api_test

import (
	"testing"

	api "forge.lthn.ai/core/go-api"
)

func TestAuthentikUser_Good(t *testing.T) {
	user := &api.AuthentikUser{
		Username: "alice",
		Email:    "alice@example.com",
		Name:     "Alice Smith",
		UID:      "abc-123",
		Groups:   []string{"admins", "developers"},
	}

	if user.Username != "alice" {
		t.Fatalf("expected username alice, got %s", user.Username)
	}
	if len(user.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(user.Groups))
	}
}

func TestAuthentikUserHasGroup_Good(t *testing.T) {
	user := &api.AuthentikUser{
		Groups: []string{"admins", "developers"},
	}

	if !user.HasGroup("admins") {
		t.Fatal("expected user to have admins group")
	}
	if user.HasGroup("viewers") {
		t.Fatal("expected user to not have viewers group")
	}
}

func TestAuthentikUserHasGroup_Bad_Empty(t *testing.T) {
	user := &api.AuthentikUser{}

	if user.HasGroup("admins") {
		t.Fatal("expected empty user to have no groups")
	}
}

func TestAuthentikConfig_Good(t *testing.T) {
	cfg := api.AuthentikConfig{
		Issuer:       "https://auth.host.uk.com/application/o/core-api/",
		ClientID:     "core-api",
		TrustedProxy: true,
	}

	if cfg.Issuer == "" {
		t.Fatal("expected non-empty issuer")
	}
	if !cfg.TrustedProxy {
		t.Fatal("expected TrustedProxy to be true")
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -run TestAuthentik
```

Expected: Compilation errors — `api.AuthentikUser`, `api.AuthentikConfig` not defined.

**Step 3: Implement authentik.go**

Create `authentik.go`:
```go
package api

// AuthentikConfig configures Authentik OIDC integration.
type AuthentikConfig struct {
	// Issuer is the OIDC issuer URL (e.g. "https://auth.host.uk.com/application/o/core-api/").
	// Used for JWT validation via OIDC discovery.
	Issuer string

	// ClientID is the OAuth2 client identifier registered in Authentik.
	ClientID string

	// TrustedProxy enables reading X-authentik-* headers set by Traefik forward auth.
	// Only enable this when go-api sits behind a trusted reverse proxy.
	TrustedProxy bool

	// PublicPaths lists path prefixes that skip authentication entirely.
	// /health and /swagger are always public regardless of this setting.
	PublicPaths []string
}

// AuthentikUser represents an authenticated user extracted from Authentik headers or JWT claims.
type AuthentikUser struct {
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	UID          string   `json:"uid"`
	Groups       []string `json:"groups"`
	Entitlements []string `json:"entitlements,omitempty"`
	JWT          string   `json:"-"`
}

// HasGroup returns true if the user belongs to the named group.
func (u *AuthentikUser) HasGroup(group string) bool {
	for _, g := range u.Groups {
		if g == group {
			return true
		}
	}
	return false
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -run TestAuthentik
```

Expected: All 4 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/go-api
git add authentik.go authentik_test.go
git commit -m "feat: add AuthentikUser and AuthentikConfig types

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 6: go-api Header Extraction Middleware (TDD)

This implements the forward auth path — extracting user identity from X-authentik-* headers set by Traefik.

**Files:**
- Modify: `/Users/snider/Code/go-api/authentik.go`
- Modify: `/Users/snider/Code/go-api/authentik_test.go`

**Step 1: Write the failing tests**

Append to `authentik_test.go`:
```go
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// authentikTestGroup returns JSON with the user from context.
type authentikTestGroup struct{}

func (g *authentikTestGroup) Name() string     { return "authtest" }
func (g *authentikTestGroup) BasePath() string { return "/v1/authtest" }
func (g *authentikTestGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/whoami", func(c *gin.Context) {
		user := api.GetUser(c)
		if user == nil {
			c.JSON(200, api.OK[any](nil))
			return
		}
		c.JSON(200, api.OK(user))
	})
}

func TestForwardAuthHeaders_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		TrustedProxy: true,
	}))
	engine.Register(&authentikTestGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/authtest/whoami", nil)
	req.Header.Set("X-authentik-username", "alice")
	req.Header.Set("X-authentik-email", "alice@example.com")
	req.Header.Set("X-authentik-name", "Alice Smith")
	req.Header.Set("X-authentik-uid", "abc-123")
	req.Header.Set("X-authentik-groups", "admins|developers")
	req.Header.Set("X-authentik-entitlements", "core:read|core:write")
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[*api.AuthentikUser]
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data == nil {
		t.Fatal("expected non-nil user data")
	}
	if resp.Data.Username != "alice" {
		t.Fatalf("expected username alice, got %s", resp.Data.Username)
	}
	if resp.Data.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", resp.Data.Email)
	}
	if len(resp.Data.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(resp.Data.Groups))
	}
	if resp.Data.Groups[0] != "admins" {
		t.Fatalf("expected first group admins, got %s", resp.Data.Groups[0])
	}
}

func TestForwardAuthHeaders_Good_NoHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		TrustedProxy: true,
	}))
	engine.Register(&authentikTestGroup{})
	handler := engine.Handler()

	// Request without Authentik headers — should pass through (middleware is permissive)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/authtest/whoami", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[*api.AuthentikUser]
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data != nil {
		t.Fatal("expected nil user when no headers present")
	}
}

func TestForwardAuthHeaders_Bad_NotTrusted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// TrustedProxy: false — should NOT read X-authentik-* headers
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		TrustedProxy: false,
	}))
	engine.Register(&authentikTestGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/authtest/whoami", nil)
	req.Header.Set("X-authentik-username", "alice")
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[*api.AuthentikUser]
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data != nil {
		t.Fatal("expected nil user when TrustedProxy is false")
	}
}

func TestHealthBypassesAuthentik_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		TrustedProxy: true,
	}))
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for /health, got %d", w.Code)
	}
}

func TestGetUser_Good_NilContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Test GetUser with no user in context (no Authentik middleware)
	engine, _ := api.New()
	engine.Register(&authentikTestGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/authtest/whoami", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -run TestForwardAuth\|TestHealthBypassesAuthentik\|TestGetUser
```

Expected: Compilation errors — `api.WithAuthentik`, `api.GetUser` not defined.

**Step 3: Add GetUser helper and middleware to authentik.go**

Append to `authentik.go`:
```go
import (
	"strings"

	"github.com/gin-gonic/gin"
)

const authentikUserKey = "authentik_user"

// GetUser returns the authenticated Authentik user from the Gin context, or nil
// if no user is authenticated.
func GetUser(c *gin.Context) *AuthentikUser {
	val, exists := c.Get(authentikUserKey)
	if !exists {
		return nil
	}
	user, ok := val.(*AuthentikUser)
	if !ok {
		return nil
	}
	return user
}

// authentikMiddleware extracts user identity from X-authentik-* headers
// (when TrustedProxy is true) and stores it in the Gin context.
// This middleware is PERMISSIVE — it does not reject unauthenticated requests.
// Handlers must check GetUser() and decide whether to require auth.
func authentikMiddleware(cfg AuthentikConfig) gin.HandlerFunc {
	publicPaths := append([]string{"/health", "/swagger"}, cfg.PublicPaths...)

	return func(c *gin.Context) {
		// Skip public paths entirely.
		for _, path := range publicPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Forward auth mode: read trusted headers from Traefik.
		if cfg.TrustedProxy {
			username := c.GetHeader("X-authentik-username")
			if username != "" {
				user := &AuthentikUser{
					Username: username,
					Email:    c.GetHeader("X-authentik-email"),
					Name:     c.GetHeader("X-authentik-name"),
					UID:      c.GetHeader("X-authentik-uid"),
					JWT:      c.GetHeader("X-authentik-jwt"),
				}

				if groups := c.GetHeader("X-authentik-groups"); groups != "" {
					user.Groups = strings.Split(groups, "|")
				}
				if ent := c.GetHeader("X-authentik-entitlements"); ent != "" {
					user.Entitlements = strings.Split(ent, "|")
				}

				c.Set(authentikUserKey, user)
			}
		}

		c.Next()
	}
}
```

**Step 4: Add WithAuthentik option to options.go**

Append to `options.go`:
```go
// WithAuthentik adds Authentik identity middleware.
// When TrustedProxy is true, reads X-authentik-* headers from Traefik forward auth.
// When Issuer is set, also validates JWT Bearer tokens via OIDC discovery.
func WithAuthentik(cfg AuthentikConfig) Option {
	return func(e *Engine) {
		e.middlewares = append(e.middlewares, authentikMiddleware(cfg))
	}
}
```

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1
```

Expected: All tests PASS (existing 36 + new 5).

**Step 6: Commit**

```bash
cd /Users/snider/Code/go-api
git add authentik.go authentik_test.go options.go
git commit -m "feat: add Authentik header extraction middleware and GetUser helper

Forward auth mode reads X-authentik-* headers from Traefik.
Middleware is permissive — handlers decide whether auth is required.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 7: go-api JWT Validation Middleware (TDD)

This implements the direct OIDC path — validating JWT Bearer tokens for API clients.

**Files:**
- Modify: `/Users/snider/Code/go-api/authentik.go`
- Modify: `/Users/snider/Code/go-api/authentik_test.go`
- Modify: `/Users/snider/Code/go-api/go.mod` (new dependency)

**Step 1: Write the failing tests**

Append to `authentik_test.go`:
```go
func TestJWTValidation_Bad_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Use a fake issuer — OIDC discovery will fail, but we test the flow
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		Issuer:   "https://auth.example.com/application/o/test/",
		ClientID: "test-client",
	}))
	engine.Register(&authentikTestGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/authtest/whoami", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt-token")
	handler.ServeHTTP(w, req)

	// Without a reachable OIDC endpoint, JWT validation can't succeed.
	// The middleware should pass through (permissive) with no user.
	if w.Code != 200 {
		t.Fatalf("expected 200 (permissive), got %d", w.Code)
	}

	var resp api.Response[*api.AuthentikUser]
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data != nil {
		t.Fatal("expected nil user for invalid JWT")
	}
}

func TestBearerAndAuthentikCoexist_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Both WithBearerAuth and WithAuthentik should work together.
	// Bearer auth gates access, Authentik extracts user identity.
	engine, _ := api.New(
		api.WithBearerAuth("secret-token"),
		api.WithAuthentik(api.AuthentikConfig{TrustedProxy: true}),
	)
	engine.Register(&authentikTestGroup{})
	handler := engine.Handler()

	// With bearer token + Authentik headers → 200 with user
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/authtest/whoami", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("X-authentik-username", "bob")
	req.Header.Set("X-authentik-email", "bob@example.com")
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp api.Response[*api.AuthentikUser]
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data == nil {
		t.Fatal("expected user data")
	}
	if resp.Data.Username != "bob" {
		t.Fatalf("expected username bob, got %s", resp.Data.Username)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -run TestJWTValidation\|TestBearerAndAuthentikCoexist
```

**Step 3: Add OIDC validation to authentik middleware**

Update `authentikMiddleware` in `authentik.go` to handle JWT Bearer tokens when `Issuer` is configured. Add the go-oidc dependency:

```bash
cd /Users/snider/Code/go-api
go get github.com/coreos/go-oidc/v3/oidc
go get golang.org/x/oauth2
```

Add JWT validation logic to the middleware — after the header extraction block, before `c.Next()`:

```go
// Direct OIDC mode: validate JWT from Authorization header.
if cfg.Issuer != "" && cfg.ClientID != "" {
    // Only attempt JWT validation if no user was extracted from headers
    // (headers take priority — they're pre-validated by Authentik).
    if GetUser(c) == nil {
        authHeader := c.GetHeader("Authorization")
        if strings.HasPrefix(authHeader, "Bearer ") {
            token := strings.TrimPrefix(authHeader, "Bearer ")
            user, err := validateJWT(c.Request.Context(), cfg, token)
            if err == nil && user != nil {
                c.Set(authentikUserKey, user)
            }
            // Permissive: if validation fails, continue without user.
        }
    }
}
```

Add the validation function:
```go
import (
    "context"
    "sync"

    oidc "github.com/coreos/go-oidc/v3/oidc"
)

var (
    oidcProviderMu sync.Mutex
    oidcProviders  = make(map[string]*oidc.Provider)
)

// getOIDCProvider returns a cached OIDC provider for the given issuer.
func getOIDCProvider(ctx context.Context, issuer string) (*oidc.Provider, error) {
    oidcProviderMu.Lock()
    defer oidcProviderMu.Unlock()

    if p, ok := oidcProviders[issuer]; ok {
        return p, nil
    }

    p, err := oidc.NewProvider(ctx, issuer)
    if err != nil {
        return nil, err
    }
    oidcProviders[issuer] = p
    return p, nil
}

// validateJWT verifies a JWT token against the OIDC provider and extracts the user.
func validateJWT(ctx context.Context, cfg AuthentikConfig, rawToken string) (*AuthentikUser, error) {
    provider, err := getOIDCProvider(ctx, cfg.Issuer)
    if err != nil {
        return nil, err
    }

    verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
    idToken, err := verifier.Verify(ctx, rawToken)
    if err != nil {
        return nil, err
    }

    var claims struct {
        PreferredUsername string   `json:"preferred_username"`
        Email            string   `json:"email"`
        Name             string   `json:"name"`
        Sub              string   `json:"sub"`
        Groups           []string `json:"groups"`
    }
    if err := idToken.Claims(&claims); err != nil {
        return nil, err
    }

    return &AuthentikUser{
        Username: claims.PreferredUsername,
        Email:    claims.Email,
        Name:     claims.Name,
        UID:      claims.Sub,
        Groups:   claims.Groups,
        JWT:      rawToken,
    }, nil
}
```

**Step 4: Run go mod tidy**

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
git add authentik.go authentik_test.go go.mod go.sum
git commit -m "feat: add OIDC JWT validation for direct API client auth

Uses coreos/go-oidc for OIDC discovery and JWT verification.
Cached provider instances. Permissive — fails open if OIDC unreachable.
Forward auth headers take priority over JWT when both present.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 8: go-api RequireAuth Middleware Helper (TDD)

The Authentik middleware is permissive. This task adds a helper for routes that REQUIRE authentication.

**Files:**
- Modify: `/Users/snider/Code/go-api/authentik.go`
- Modify: `/Users/snider/Code/go-api/authentik_test.go`

**Step 1: Write the failing tests**

Append to `authentik_test.go`:
```go
// protectedGroup uses RequireAuth on its routes.
type protectedGroup struct{}

func (g *protectedGroup) Name() string     { return "protected" }
func (g *protectedGroup) BasePath() string { return "/v1/protected" }
func (g *protectedGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/data", api.RequireAuth(), func(c *gin.Context) {
		user := api.GetUser(c)
		c.JSON(200, api.OK(user.Username))
	})
}

func TestRequireAuth_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{TrustedProxy: true}))
	engine.Register(&protectedGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/protected/data", nil)
	req.Header.Set("X-authentik-username", "alice")
	req.Header.Set("X-authentik-email", "alice@example.com")
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 with auth, got %d", w.Code)
	}
}

func TestRequireAuth_Bad_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{TrustedProxy: true}))
	engine.Register(&protectedGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/protected/data", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 without auth, got %d", w.Code)
	}
}

func TestRequireAuth_Bad_NoAuthentikMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// No WithAuthentik — RequireAuth should still return 401
	engine, _ := api.New()
	engine.Register(&protectedGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/protected/data", nil)
	handler.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// groupRequireGroup uses RequireGroup.
type groupRequireGroup struct{}

func (g *groupRequireGroup) Name() string     { return "adminonly" }
func (g *groupRequireGroup) BasePath() string { return "/v1/admin" }
func (g *groupRequireGroup) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/panel", api.RequireGroup("admins"), func(c *gin.Context) {
		c.JSON(200, api.OK("admin panel"))
	})
}

func TestRequireGroup_Good(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{TrustedProxy: true}))
	engine.Register(&groupRequireGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/admin/panel", nil)
	req.Header.Set("X-authentik-username", "alice")
	req.Header.Set("X-authentik-groups", "admins|developers")
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 for admin user, got %d", w.Code)
	}
}

func TestRequireGroup_Bad_WrongGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine, _ := api.New(api.WithAuthentik(api.AuthentikConfig{TrustedProxy: true}))
	engine.Register(&groupRequireGroup{})
	handler := engine.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/admin/panel", nil)
	req.Header.Set("X-authentik-username", "bob")
	req.Header.Set("X-authentik-groups", "developers")
	handler.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for non-admin user, got %d", w.Code)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -run TestRequireAuth\|TestRequireGroup
```

Expected: Compilation errors — `api.RequireAuth`, `api.RequireGroup` not defined.

**Step 3: Implement RequireAuth and RequireGroup**

Append to `authentik.go`:
```go
import "net/http"

// RequireAuth is a Gin middleware that returns 401 if no authenticated user
// is present in the context. Use after WithAuthentik() middleware.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetUser(c) == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				Fail("unauthorised", "Authentication required"))
			return
		}
		c.Next()
	}
}

// RequireGroup is a Gin middleware that returns 403 if the authenticated user
// does not belong to the specified group. Implies RequireAuth.
func RequireGroup(group string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				Fail("unauthorised", "Authentication required"))
			return
		}
		if !user.HasGroup(group) {
			c.AbortWithStatusJSON(http.StatusForbidden,
				Fail("forbidden", "Insufficient permissions"))
			return
		}
		c.Next()
	}
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1
```

Expected: All tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/go-api
git add authentik.go authentik_test.go
git commit -m "feat: add RequireAuth and RequireGroup middleware helpers

RequireAuth returns 401 when no user in context.
RequireGroup returns 403 when user lacks the specified group.
Both use British English 'unauthorised' in error responses.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 9: Update go-api Documentation

**Files:**
- Modify: `/Users/snider/Code/go-api/CLAUDE.md`
- Modify: `/Users/snider/Code/go-api/README.md`

**Step 1: Update CLAUDE.md**

Add to the Project Overview section:
```markdown
## Authentik Integration

go-api supports Authentik as the identity provider:

- **Forward auth mode**: Reads `X-authentik-*` headers from Traefik (requires `TrustedProxy: true`)
- **OIDC mode**: Validates JWT Bearer tokens via OIDC discovery
- **Permissive middleware**: `WithAuthentik()` extracts user but doesn't block. Use `RequireAuth()` / `RequireGroup()` on routes that need auth.
- **Coexists with `WithBearerAuth()`** for service-to-service tokens

```go
engine, _ := api.New(
    api.WithAuthentik(api.AuthentikConfig{
        Issuer:       "https://auth.host.uk.com/application/o/core-api/",
        ClientID:     "core-api",
        TrustedProxy: true,
    }),
)
```
```

**Step 2: Update README.md**

Add Authentik section with quick-start example showing `WithAuthentik()`, `GetUser()`, `RequireAuth()`, and `RequireGroup()`.

**Step 3: Commit**

```bash
cd /Users/snider/Code/go-api
git add CLAUDE.md README.md
git commit -m "docs: add Authentik integration guide to CLAUDE.md and README

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 10: Push go-api and DevOps Changes

**Step 1: Push go-api**

```bash
cd /Users/snider/Code/go-api
go test ./... -v -count=1  # Final verification
git push forge main
```

**Step 2: Push DevOps**

```bash
cd /Users/snider/Code/DevOps
git push forge main
```

**Step 3: Update go-ecosystem memory**

Update the go-api entry in the ecosystem inventory to note Authentik middleware.

---

## Dependency Summary

```
Task 1 (enable vars) → Task 2 (playbook) → Task 3 (deploy) → Task 4 (OIDC app)
                                                                     ↓
Task 5 (user type) → Task 6 (header middleware) → Task 7 (JWT) → Task 8 (RequireAuth)
                                                                     ↓
                                                               Task 9 (docs) → Task 10 (push)
```

Tasks 1-4 (DevOps) and Tasks 5-8 (Go) are independent tracks that can run in parallel. Task 9-10 depend on both tracks.

## Estimated Sizes

| Task | LOC | Tests |
|------|-----|-------|
| Task 5: User type | ~50 | 4 |
| Task 6: Header middleware | ~60 | 5 |
| Task 7: JWT validation | ~80 | 2 |
| Task 8: RequireAuth/Group | ~30 | 5 |
| **go-api total** | **~220** | **16** |
