# go-forge Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a full-coverage Go client for the Forgejo API (450 endpoints) using a generic Resource[T,C,U] pattern and types generated from swagger.v1.json.

**Architecture:** A code generator (`cmd/forgegen/`) parses Forgejo's Swagger 2.0 spec and emits typed Go structs. A generic `Resource[T,C,U]` provides List/Get/Create/Update/Delete for 411 CRUD endpoints. 18 service structs embed the generic resource and add 39 hand-written action methods. An HTTP client handles auth, pagination, rate limiting, and context.Context.

**Tech Stack:** Go 1.25, `net/http`, `text/template`, generics, Swagger 2.0 (JSON)

---

## Context

**This is a NEW repo** at `forge.lthn.ai/core/go-forge`. Create it locally at `/Users/snider/Code/go-forge`.

**Extracted from:** `/Users/snider/Code/go-scm/forge/` (45 methods covering 10% of API). The config resolution pattern (env → file → flags) comes from there.

**Swagger spec:** Download from `https://forge.lthn.ai/swagger.v1.json` — Swagger 2.0 format, 229 type definitions, 450 operations across 284 paths. Pin it at `testdata/swagger.v1.json`.

**Forgejo version:** 10.0.3 (Gitea 1.22.0 compatible)

**Dependencies:** None (pure `net/http`). Config uses `forge.lthn.ai/core/go` for `pkg/config` and `pkg/log` — same as go-scm.

**Key insight:** 91% of endpoints are generic CRUD (List/Get/Create/Update/Delete). The generic `Resource[T,C,U]` pattern means each service is a struct definition + path constant + optional action methods. The code generator handles 229 type definitions.

**Test command:** `go test ./...` from the repo root.

**The forge remote for this repo will be:** `ssh://git@forge.lthn.ai:2223/core/go-forge.git`

---

## Wave 1: Foundation (Tasks 1-6)

### Task 1: Repo scaffolding + go.mod

**Files:**
- Create: `go.mod`
- Create: `go.sum` (auto-generated)
- Create: `doc.go`
- Create: `testdata/swagger.v1.json` (downloaded)

**Step 1: Create directory and initialise module**

```bash
mkdir -p /Users/snider/Code/go-forge/testdata
cd /Users/snider/Code/go-forge
git init
go mod init forge.lthn.ai/core/go-forge
```

**Step 2: Download and pin swagger spec**

```bash
curl -s https://forge.lthn.ai/swagger.v1.json > testdata/swagger.v1.json
```

Verify: `python3 -c "import json; d=json.load(open('testdata/swagger.v1.json')); print(f'{len(d[\"definitions\"])} types, {len(d[\"paths\"])} paths')"`
Expected: `229 types, 284 paths`

**Step 3: Write doc.go**

```go
// Package forge provides a full-coverage Go client for the Forgejo API.
//
// Usage:
//
//	f := forge.NewForge("https://forge.lthn.ai", "your-token")
//	repos, err := f.Repos.List(ctx, forge.Params{"org": "core"}, forge.DefaultList)
//
// Types are generated from Forgejo's swagger.v1.json spec via cmd/forgegen/.
// Run `go generate ./types/...` to regenerate after a Forgejo upgrade.
package forge
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: scaffold go-forge repo with pinned swagger spec

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 2: HTTP Client

**Files:**
- Create: `client.go`
- Create: `client_test.go`

**Step 1: Write client tests**

```go
package forge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Good_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("missing auth header")
		}
		if r.URL.Path != "/api/v1/user" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]string{"login": "virgil"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	var out map[string]string
	err := c.Get(context.Background(), "/api/v1/user", &out)
	if err != nil {
		t.Fatal(err)
	}
	if out["login"] != "virgil" {
		t.Errorf("got login=%q", out["login"])
	}
}

func TestClient_Good_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test-repo" {
			t.Errorf("wrong body: %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "test-repo"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	body := map[string]string{"name": "test-repo"}
	var out map[string]any
	err := c.Post(context.Background(), "/api/v1/orgs/core/repos", body, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out["name"] != "test-repo" {
		t.Errorf("got name=%v", out["name"])
	}
}

func TestClient_Good_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	err := c.Delete(context.Background(), "/api/v1/repos/core/test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Bad_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	err := c.Get(context.Background(), "/api/v1/user", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("got status=%d", apiErr.StatusCode)
	}
}

func TestClient_Bad_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	err := c.Get(context.Background(), "/api/v1/repos/x/y", nil)
	if !IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestClient_Good_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	err := c.Get(ctx, "/api/v1/user", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestClient_Good_Options(t *testing.T) {
	c := NewClient("https://forge.lthn.ai", "tok",
		WithUserAgent("go-forge/1.0"),
	)
	if c.userAgent != "go-forge/1.0" {
		t.Errorf("got user agent=%q", c.userAgent)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestClient`
Expected: Compilation errors (types don't exist yet)

**Step 3: Write client.go**

```go
package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError represents an error response from the Forgejo API.
type APIError struct {
	StatusCode int
	Message    string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("forge: %s %d: %s", e.URL, e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// IsForbidden returns true if the error is a 403 response.
func IsForbidden(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusForbidden
}

// IsConflict returns true if the error is a 409 response.
func IsConflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// Client is a low-level HTTP client for the Forgejo API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new Forgejo API client.
func NewClient(url, token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(url, "/"),
		token:      token,
		httpClient: http.DefaultClient,
		userAgent:  "go-forge/0.1",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPatch, path, body, out)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("forge: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("forge: create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("forge: request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.parseError(resp, path)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("forge: decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) parseError(resp *http.Response, path string) error {
	var errBody struct {
		Message string `json:"message"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&errBody)
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    errBody.Message,
		URL:        path,
	}
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestClient`
Expected: All 7 tests PASS

**Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: HTTP client with auth, context, error handling

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 3: Pagination

**Files:**
- Create: `pagination.go`
- Create: `pagination_test.go`

**Step 1: Write pagination tests**

```go
package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPagination_Good_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "2")
		json.NewEncoder(w).Encode([]map[string]int{{"id": 1}, {"id": 2}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	result, err := ListAll[map[string]int](context.Background(), c, "/api/v1/repos", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("got %d items", len(result))
	}
}

func TestPagination_Good_MultiPage(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("X-Total-Count", "100")
		items := make([]map[string]int, 50)
		for i := range items {
			items[i] = map[string]int{"id": (page-1)*50 + i + 1}
		}
		json.NewEncoder(w).Encode(items)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	result, err := ListAll[map[string]int](context.Background(), c, "/api/v1/repos", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 100 {
		t.Errorf("got %d items, want 100", len(result))
	}
}

func TestPagination_Good_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "0")
		json.NewEncoder(w).Encode([]map[string]int{})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	result, err := ListAll[map[string]int](context.Background(), c, "/api/v1/repos", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("got %d items", len(result))
	}
}

func TestListPage_Good_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("page")
		l := r.URL.Query().Get("limit")
		s := r.URL.Query().Get("state")
		if p != "2" || l != "25" || s != "open" {
			t.Errorf("wrong params: page=%s limit=%s state=%s", p, l, s)
		}
		w.Header().Set("X-Total-Count", "50")
		json.NewEncoder(w).Encode([]map[string]int{})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	_, err := ListPage[map[string]int](context.Background(), c, "/api/v1/repos",
		map[string]string{"state": "open"}, ListOptions{Page: 2, Limit: 25})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPagination_Bad_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"message": "fail"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	_, err := ListAll[map[string]int](context.Background(), c, "/api/v1/repos", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestPagination -run TestListPage`
Expected: Compilation errors

**Step 3: Write pagination.go**

```go
package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListOptions controls pagination.
type ListOptions struct {
	Page  int // 1-based page number
	Limit int // items per page (default 50)
}

// DefaultList returns sensible default pagination.
var DefaultList = ListOptions{Page: 1, Limit: 50}

// PagedResult holds a single page of results with metadata.
type PagedResult[T any] struct {
	Items      []T
	TotalCount int
	Page       int
	HasMore    bool
}

// ListPage fetches a single page of results.
// Extra query params can be passed via the query map.
func ListPage[T any](ctx context.Context, c *Client, path string, query map[string]string, opts ListOptions) (*PagedResult[T], error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Limit < 1 {
		opts.Limit = 50
	}

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("forge: parse url: %w", err)
	}

	q := u.Query()
	q.Set("page", strconv.Itoa(opts.Page))
	q.Set("limit", strconv.Itoa(opts.Limit))
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("forge: create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("forge: request GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, c.parseError(resp, path)
	}

	var items []T
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("forge: decode response: %w", err)
	}

	totalCount, _ := strconv.Atoi(resp.Header.Get("X-Total-Count"))

	return &PagedResult[T]{
		Items:      items,
		TotalCount: totalCount,
		Page:       opts.Page,
		HasMore:    len(items) >= opts.Limit && opts.Page*opts.Limit < totalCount,
	}, nil
}

// ListAll fetches all pages of results.
func ListAll[T any](ctx context.Context, c *Client, path string, query map[string]string) ([]T, error) {
	var all []T
	page := 1

	for {
		result, err := ListPage[T](ctx, c, path, query, ListOptions{Page: page, Limit: 50})
		if err != nil {
			return nil, err
		}
		all = append(all, result.Items...)
		if !result.HasMore {
			break
		}
		page++
	}

	return all, nil
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v -run "TestPagination|TestListPage"`
Expected: All 5 tests PASS

**Step 5: Commit**

```bash
git add pagination.go pagination_test.go
git commit -m "feat: generic pagination with ListAll and ListPage

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 4: Params and path resolution

**Files:**
- Create: `params.go`
- Create: `params_test.go`

**Step 1: Write tests**

```go
package forge

import "testing"

func TestResolvePath_Good_Simple(t *testing.T) {
	got := ResolvePath("/api/v1/repos/{owner}/{repo}", Params{"owner": "core", "repo": "go-forge"})
	want := "/api/v1/repos/core/go-forge"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolvePath_Good_NoParams(t *testing.T) {
	got := ResolvePath("/api/v1/user", nil)
	if got != "/api/v1/user" {
		t.Errorf("got %q", got)
	}
}

func TestResolvePath_Good_WithID(t *testing.T) {
	got := ResolvePath("/api/v1/repos/{owner}/{repo}/issues/{index}", Params{
		"owner": "core", "repo": "go-forge", "index": "42",
	})
	want := "/api/v1/repos/core/go-forge/issues/42"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolvePath_Good_URLEncoding(t *testing.T) {
	got := ResolvePath("/api/v1/repos/{owner}/{repo}", Params{"owner": "my org", "repo": "my repo"})
	want := "/api/v1/repos/my%20org/my%20repo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestResolvePath`
Expected: Compilation errors

**Step 3: Write params.go**

```go
package forge

import (
	"net/url"
	"strings"
)

// Params maps path variable names to values.
// Example: Params{"owner": "core", "repo": "go-forge"}
type Params map[string]string

// ResolvePath substitutes {placeholders} in path with values from params.
func ResolvePath(path string, params Params) string {
	for k, v := range params {
		path = strings.ReplaceAll(path, "{"+k+"}", url.PathEscape(v))
	}
	return path
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestResolvePath`
Expected: All 4 tests PASS

**Step 5: Commit**

```bash
git add params.go params_test.go
git commit -m "feat: path parameter resolution with URL encoding

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 5: Generic Resource[T, C, U]

**Files:**
- Create: `resource.go`
- Create: `resource_test.go`

**Step 1: Write resource tests**

```go
package forge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test types
type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type testCreate struct {
	Name string `json:"name"`
}

type testUpdate struct {
	Name *string `json:"name,omitempty"`
}

func TestResource_Good_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/orgs/core/repos" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.Header().Set("X-Total-Count", "2")
		json.NewEncoder(w).Encode([]testItem{{1, "a"}, {2, "b"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	res := NewResource[testItem, testCreate, testUpdate](c, "/api/v1/orgs/{org}/repos")

	items, err := res.List(context.Background(), Params{"org": "core"}, DefaultList)
	if err != nil {
		t.Fatal(err)
	}
	if len(items.Items) != 2 {
		t.Errorf("got %d items", len(items.Items))
	}
}

func TestResource_Good_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repos/core/go-forge" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(testItem{1, "go-forge"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	res := NewResource[testItem, testCreate, testUpdate](c, "/api/v1/repos/{owner}/{repo}")

	item, err := res.Get(context.Background(), Params{"owner": "core", "repo": "go-forge"})
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "go-forge" {
		t.Errorf("got name=%q", item.Name)
	}
}

func TestResource_Good_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body testCreate
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(testItem{1, body.Name})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	res := NewResource[testItem, testCreate, testUpdate](c, "/api/v1/orgs/{org}/repos")

	item, err := res.Create(context.Background(), Params{"org": "core"}, &testCreate{Name: "new-repo"})
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "new-repo" {
		t.Errorf("got name=%q", item.Name)
	}
}

func TestResource_Good_Update(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(testItem{1, "updated"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	res := NewResource[testItem, testCreate, testUpdate](c, "/api/v1/repos/{owner}/{repo}")

	name := "updated"
	item, err := res.Update(context.Background(), Params{"owner": "core", "repo": "old"}, &testUpdate{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "updated" {
		t.Errorf("got name=%q", item.Name)
	}
}

func TestResource_Good_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	res := NewResource[testItem, testCreate, testUpdate](c, "/api/v1/repos/{owner}/{repo}")

	err := res.Delete(context.Background(), Params{"owner": "core", "repo": "old"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestResource_Good_ListAll(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("X-Total-Count", "3")
		if page == 1 {
			json.NewEncoder(w).Encode([]testItem{{1, "a"}, {2, "b"}})
		} else {
			json.NewEncoder(w).Encode([]testItem{{3, "c"}})
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	res := NewResource[testItem, testCreate, testUpdate](c, "/api/v1/repos")

	items, err := res.ListAll(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Errorf("got %d items, want 3", len(items))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestResource`
Expected: Compilation errors

**Step 3: Write resource.go**

```go
package forge

import "context"

// Resource provides generic CRUD operations for a Forgejo API resource.
// T is the resource type, C is the create options type, U is the update options type.
type Resource[T any, C any, U any] struct {
	client *Client
	path   string
}

// NewResource creates a new Resource for the given path pattern.
// The path may contain {placeholders} that are resolved via Params.
func NewResource[T any, C any, U any](c *Client, path string) *Resource[T, C, U] {
	return &Resource[T, C, U]{client: c, path: path}
}

// List returns a single page of resources.
func (r *Resource[T, C, U]) List(ctx context.Context, params Params, opts ListOptions) (*PagedResult[T], error) {
	return ListPage[T](ctx, r.client, ResolvePath(r.path, params), nil, opts)
}

// ListAll returns all resources across all pages.
func (r *Resource[T, C, U]) ListAll(ctx context.Context, params Params) ([]T, error) {
	return ListAll[T](ctx, r.client, ResolvePath(r.path, params), nil)
}

// Get returns a single resource by appending id to the path.
func (r *Resource[T, C, U]) Get(ctx context.Context, params Params) (*T, error) {
	var out T
	if err := r.client.Get(ctx, ResolvePath(r.path, params), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create creates a new resource.
func (r *Resource[T, C, U]) Create(ctx context.Context, params Params, body *C) (*T, error) {
	var out T
	if err := r.client.Post(ctx, ResolvePath(r.path, params), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies an existing resource.
func (r *Resource[T, C, U]) Update(ctx context.Context, params Params, body *U) (*T, error) {
	var out T
	if err := r.client.Patch(ctx, ResolvePath(r.path, params), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a resource.
func (r *Resource[T, C, U]) Delete(ctx context.Context, params Params) error {
	return r.client.Delete(ctx, ResolvePath(r.path, params))
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestResource`
Expected: All 6 tests PASS

**Step 5: Commit**

```bash
git add resource.go resource_test.go
git commit -m "feat: generic Resource[T,C,U] for CRUD operations

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 6: Config resolution (extracted from go-scm)

**Files:**
- Create: `config.go`
- Create: `config_test.go`

**Step 1: Write config tests**

```go
package forge

import (
	"os"
	"testing"
)

func TestResolveConfig_Good_EnvOverrides(t *testing.T) {
	t.Setenv("FORGE_URL", "https://forge.example.com")
	t.Setenv("FORGE_TOKEN", "env-token")

	url, token, err := ResolveConfig("", "")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://forge.example.com" {
		t.Errorf("got url=%q", url)
	}
	if token != "env-token" {
		t.Errorf("got token=%q", token)
	}
}

func TestResolveConfig_Good_FlagOverridesEnv(t *testing.T) {
	t.Setenv("FORGE_URL", "https://env.example.com")
	t.Setenv("FORGE_TOKEN", "env-token")

	url, token, err := ResolveConfig("https://flag.example.com", "flag-token")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://flag.example.com" {
		t.Errorf("got url=%q", url)
	}
	if token != "flag-token" {
		t.Errorf("got token=%q", token)
	}
}

func TestResolveConfig_Good_DefaultURL(t *testing.T) {
	// Clear env vars to test defaults
	os.Unsetenv("FORGE_URL")
	os.Unsetenv("FORGE_TOKEN")

	url, _, err := ResolveConfig("", "")
	if err != nil {
		t.Fatal(err)
	}
	if url != DefaultURL {
		t.Errorf("got url=%q, want %q", url, DefaultURL)
	}
}

func TestNewForgeFromConfig_Bad_NoToken(t *testing.T) {
	os.Unsetenv("FORGE_URL")
	os.Unsetenv("FORGE_TOKEN")

	_, err := NewForgeFromConfig("", "")
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v -run TestResolveConfig -run TestNewForgeFromConfig`
Expected: Compilation errors

**Step 3: Write config.go**

```go
package forge

import (
	"fmt"
	"os"
)

const (
	// DefaultURL is used when no URL is configured.
	DefaultURL = "http://localhost:3000"
)

// ResolveConfig resolves Forge URL and token from multiple sources.
// Priority (highest to lowest): flags → environment → defaults.
func ResolveConfig(flagURL, flagToken string) (url, token string, err error) {
	// Environment variables
	url = os.Getenv("FORGE_URL")
	token = os.Getenv("FORGE_TOKEN")

	// Flag overrides
	if flagURL != "" {
		url = flagURL
	}
	if flagToken != "" {
		token = flagToken
	}

	// Default URL
	if url == "" {
		url = DefaultURL
	}

	return url, token, nil
}

// NewForgeFromConfig creates a Forge client using resolved configuration.
func NewForgeFromConfig(flagURL, flagToken string, opts ...Option) (*Forge, error) {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("forge: no API token configured (set FORGE_TOKEN or pass --token)")
	}
	return NewForge(url, token, opts...), nil
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v -run "TestResolveConfig|TestNewForgeFromConfig"`
Expected: All 4 tests PASS (Note: `NewForge` doesn't exist yet — if this fails, create a stub `NewForge` function that just returns `&Forge{client: NewClient(url, token, opts...)}`)

**Step 5: Commit**

```bash
git add config.go config_test.go
git commit -m "feat: config resolution from env vars and flags

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

## Wave 2: Code Generator (Tasks 7-9)

### Task 7: Swagger spec parser

**Files:**
- Create: `cmd/forgegen/main.go`
- Create: `cmd/forgegen/parser.go`
- Create: `cmd/forgegen/parser_test.go`

The parser reads swagger.v1.json and extracts type definitions into an intermediate representation.

**Step 1: Write parser tests**

```go
package main

import (
	"os"
	"testing"
)

func TestParser_Good_LoadSpec(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Swagger != "2.0" {
		t.Errorf("got swagger=%q", spec.Swagger)
	}
	if len(spec.Definitions) < 200 {
		t.Errorf("got %d definitions, expected 200+", len(spec.Definitions))
	}
}

func TestParser_Good_ExtractTypes(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	types := ExtractTypes(spec)
	if len(types) < 200 {
		t.Errorf("got %d types", len(types))
	}

	// Check a known type
	repo, ok := types["Repository"]
	if !ok {
		t.Fatal("Repository type not found")
	}
	if len(repo.Fields) < 50 {
		t.Errorf("Repository has %d fields, expected 50+", len(repo.Fields))
	}
}

func TestParser_Good_FieldTypes(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	types := ExtractTypes(spec)
	repo := types["Repository"]

	// Check specific field mappings
	for _, f := range repo.Fields {
		switch f.JSONName {
		case "id":
			if f.GoType != "int64" {
				t.Errorf("id: got %q, want int64", f.GoType)
			}
		case "name":
			if f.GoType != "string" {
				t.Errorf("name: got %q, want string", f.GoType)
			}
		case "private":
			if f.GoType != "bool" {
				t.Errorf("private: got %q, want bool", f.GoType)
			}
		case "created_at":
			if f.GoType != "time.Time" {
				t.Errorf("created_at: got %q, want time.Time", f.GoType)
			}
		case "owner":
			if f.GoType != "*User" {
				t.Errorf("owner: got %q, want *User", f.GoType)
			}
		}
	}
}

func TestParser_Good_DetectCreateEditPairs(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	pairs := DetectCRUDPairs(spec)
	// Should find Repository, Issue, PullRequest, etc.
	if len(pairs) < 10 {
		t.Errorf("got %d pairs, expected 10+", len(pairs))
	}

	found := false
	for _, p := range pairs {
		if p.Base == "Repository" {
			found = true
			if p.Create != "CreateRepoOption" {
				t.Errorf("repo create=%q", p.Create)
			}
		}
	}
	if !found {
		t.Fatal("Repository pair not found")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v ./cmd/forgegen/ -run TestParser`
Expected: Compilation errors

**Step 3: Write parser.go**

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Spec represents a Swagger 2.0 specification.
type Spec struct {
	Swagger     string                        `json:"swagger"`
	Info        SpecInfo                       `json:"info"`
	Definitions map[string]SchemaDefinition   `json:"definitions"`
	Paths       map[string]map[string]any     `json:"paths"`
}

type SpecInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// SchemaDefinition represents a type definition in the spec.
type SchemaDefinition struct {
	Description string                       `json:"description"`
	Type        string                       `json:"type"`
	Properties  map[string]SchemaProperty    `json:"properties"`
	Required    []string                     `json:"required"`
	Enum        []any                        `json:"enum"`
	XGoName     string                       `json:"x-go-name"`
}

// SchemaProperty represents a field in a type definition.
type SchemaProperty struct {
	Type        string          `json:"type"`
	Format      string          `json:"format"`
	Description string          `json:"description"`
	Ref         string          `json:"$ref"`
	Items       *SchemaProperty `json:"items"`
	Enum        []any           `json:"enum"`
	XGoName     string          `json:"x-go-name"`
}

// GoType represents a Go type extracted from the spec.
type GoType struct {
	Name        string
	Description string
	Fields      []GoField
	IsEnum      bool
	EnumValues  []string
}

// GoField represents a field in a Go struct.
type GoField struct {
	GoName   string
	GoType   string
	JSONName string
	Comment  string
	Required bool
}

// CRUDPair maps a base type to its Create and Edit option types.
type CRUDPair struct {
	Base   string // e.g. "Repository"
	Create string // e.g. "CreateRepoOption"
	Edit   string // e.g. "EditRepoOption"
}

// LoadSpec reads and parses a Swagger 2.0 JSON file.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	return &spec, nil
}

// ExtractTypes converts spec definitions to Go types.
func ExtractTypes(spec *Spec) map[string]*GoType {
	result := make(map[string]*GoType)

	for name, def := range spec.Definitions {
		gt := &GoType{
			Name:        name,
			Description: def.Description,
		}

		if len(def.Enum) > 0 {
			gt.IsEnum = true
			for _, v := range def.Enum {
				gt.EnumValues = append(gt.EnumValues, fmt.Sprintf("%v", v))
			}
			sort.Strings(gt.EnumValues)
			result[name] = gt
			continue
		}

		required := make(map[string]bool)
		for _, r := range def.Required {
			required[r] = true
		}

		for fieldName, prop := range def.Properties {
			goName := prop.XGoName
			if goName == "" {
				goName = pascalCase(fieldName)
			}

			gf := GoField{
				GoName:   goName,
				GoType:   resolveGoType(prop),
				JSONName: fieldName,
				Comment:  prop.Description,
				Required: required[fieldName],
			}
			gt.Fields = append(gt.Fields, gf)
		}

		// Sort fields alphabetically for stable output
		sort.Slice(gt.Fields, func(i, j int) bool {
			return gt.Fields[i].GoName < gt.Fields[j].GoName
		})

		result[name] = gt
	}

	return result
}

// DetectCRUDPairs finds Create/Edit option pairs.
func DetectCRUDPairs(spec *Spec) []CRUDPair {
	var pairs []CRUDPair

	for name := range spec.Definitions {
		if !strings.HasPrefix(name, "Create") || !strings.HasSuffix(name, "Option") {
			continue
		}

		// CreateXxxOption → Xxx → EditXxxOption
		inner := strings.TrimPrefix(name, "Create")
		inner = strings.TrimSuffix(inner, "Option")

		editName := "Edit" + inner + "Option"

		pair := CRUDPair{
			Base:   inner,
			Create: name,
		}

		if _, ok := spec.Definitions[editName]; ok {
			pair.Edit = editName
		}

		pairs = append(pairs, pair)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Base < pairs[j].Base
	})

	return pairs
}

func resolveGoType(prop SchemaProperty) string {
	if prop.Ref != "" {
		parts := strings.Split(prop.Ref, "/")
		return "*" + parts[len(parts)-1]
	}

	switch prop.Type {
	case "string":
		switch prop.Format {
		case "date-time":
			return "time.Time"
		case "binary":
			return "[]byte"
		default:
			return "string"
		}
	case "integer":
		switch prop.Format {
		case "int64":
			return "int64"
		case "int32":
			return "int32"
		default:
			return "int"
		}
	case "number":
		switch prop.Format {
		case "float":
			return "float32"
		default:
			return "float64"
		}
	case "boolean":
		return "bool"
	case "array":
		if prop.Items != nil {
			itemType := resolveGoType(*prop.Items)
			return "[]" + itemType
		}
		return "[]any"
	case "object":
		return "map[string]any"
	default:
		if prop.Type == "" && prop.Ref == "" {
			return "any"
		}
		return "any"
	}
}

func pascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		// Handle common acronyms
		upper := strings.ToUpper(p)
		switch upper {
		case "ID", "URL", "HTML", "SSH", "HTTP", "HTTPS", "API", "URI", "GPG", "IP", "CSS", "JS":
			parts[i] = upper
		default:
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
```

**Step 4: Write main.go stub**

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	specPath := flag.String("spec", "testdata/swagger.v1.json", "path to swagger.v1.json")
	outDir := flag.String("out", "types", "output directory for generated types")
	flag.Parse()

	spec, err := LoadSpec(*specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	types := ExtractTypes(spec)
	pairs := DetectCRUDPairs(spec)

	fmt.Printf("Loaded %d types, %d CRUD pairs\n", len(types), len(pairs))
	fmt.Printf("Output dir: %s\n", *outDir)

	// Generation happens in Task 8
	if err := Generate(types, pairs, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 5: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v ./cmd/forgegen/ -run TestParser`
Expected: All 4 tests PASS (Note: `Generate` doesn't exist yet — add a stub: `func Generate(...) error { return nil }`)

**Step 6: Commit**

```bash
git add cmd/forgegen/
git commit -m "feat: swagger spec parser for type extraction

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 8: Code generator — Go source emission

**Files:**
- Create: `cmd/forgegen/generator.go`
- Create: `cmd/forgegen/generator_test.go`

**Step 1: Write generator tests**

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_Good_CreatesFiles(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	types := ExtractTypes(spec)
	pairs := DetectCRUDPairs(spec)

	outDir := t.TempDir()
	if err := Generate(types, pairs, outDir); err != nil {
		t.Fatal(err)
	}

	// Should create at least one .go file
	entries, _ := os.ReadDir(outDir)
	goFiles := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".go") {
			goFiles++
		}
	}
	if goFiles == 0 {
		t.Fatal("no .go files generated")
	}
}

func TestGenerate_Good_ValidGoSyntax(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	types := ExtractTypes(spec)
	pairs := DetectCRUDPairs(spec)

	outDir := t.TempDir()
	if err := Generate(types, pairs, outDir); err != nil {
		t.Fatal(err)
	}

	// Read a generated file and verify basic Go syntax markers
	data, err := os.ReadFile(filepath.Join(outDir, "repo.go"))
	if err != nil {
		// Try another name
		entries, _ := os.ReadDir(outDir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".go") {
				data, err = os.ReadFile(filepath.Join(outDir, e.Name()))
				break
			}
		}
	}
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "package types") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(content, "// Code generated") {
		t.Error("missing generated comment")
	}
}

func TestGenerate_Good_RepositoryType(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	types := ExtractTypes(spec)
	pairs := DetectCRUDPairs(spec)

	outDir := t.TempDir()
	if err := Generate(types, pairs, outDir); err != nil {
		t.Fatal(err)
	}

	// Find file containing Repository type
	var content string
	entries, _ := os.ReadDir(outDir)
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(outDir, e.Name()))
		if strings.Contains(string(data), "type Repository struct") {
			content = string(data)
			break
		}
	}

	if content == "" {
		t.Fatal("Repository type not found in any generated file")
	}

	// Check essential fields exist
	checks := []string{
		"`json:\"id\"`",
		"`json:\"name\"`",
		"`json:\"full_name\"`",
		"`json:\"private\"`",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("missing field with tag %s", check)
		}
	}
}

func TestGenerate_Good_TimeImport(t *testing.T) {
	spec, err := LoadSpec("../../testdata/swagger.v1.json")
	if err != nil {
		t.Fatal(err)
	}

	types := ExtractTypes(spec)
	pairs := DetectCRUDPairs(spec)

	outDir := t.TempDir()
	if err := Generate(types, pairs, outDir); err != nil {
		t.Fatal(err)
	}

	// Files with time.Time fields should import "time"
	entries, _ := os.ReadDir(outDir)
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(outDir, e.Name()))
		content := string(data)
		if strings.Contains(content, "time.Time") && !strings.Contains(content, "\"time\"") {
			t.Errorf("file %s uses time.Time but doesn't import time", e.Name())
		}
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v ./cmd/forgegen/ -run TestGenerate`
Expected: Failures (Generate is stub)

**Step 3: Write generator.go**

The generator groups types by logical domain and writes one `.go` file per group. Type grouping uses name prefixes and the CRUD pairs.

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// typeGrouping maps types to their output file.
var typeGrouping = map[string]string{
	"Repository":      "repo",
	"Repo":            "repo",
	"Issue":           "issue",
	"PullRequest":     "pr",
	"Pull":            "pr",
	"User":            "user",
	"Organization":    "org",
	"Org":             "org",
	"Team":            "team",
	"Label":           "label",
	"Milestone":       "milestone",
	"Release":         "release",
	"Tag":             "tag",
	"Branch":          "branch",
	"Hook":            "hook",
	"Deploy":          "key",
	"PublicKey":       "key",
	"GPGKey":          "key",
	"Key":             "key",
	"Notification":    "notification",
	"Package":         "package",
	"Action":          "action",
	"Commit":          "commit",
	"Git":             "git",
	"Contents":        "content",
	"File":            "content",
	"Wiki":            "wiki",
	"Comment":         "comment",
	"Review":          "review",
	"Reaction":        "reaction",
	"Topic":           "topic",
	"Status":          "status",
	"Combined":        "status",
	"Cron":            "admin",
	"Quota":           "quota",
	"OAuth2":          "oauth",
	"AccessToken":     "oauth",
	"API":             "error",
	"Forbidden":       "error",
	"NotFound":        "error",
	"NodeInfo":        "federation",
	"Activity":        "activity",
	"Feed":            "activity",
	"StopWatch":       "time_tracking",
	"TrackedTime":     "time_tracking",
	"Blocked":         "user",
	"Email":           "user",
	"Settings":        "settings",
	"GeneralAPI":      "settings",
	"GeneralAttachment": "settings",
	"GeneralRepo":     "settings",
	"GeneralUI":       "settings",
	"Markdown":        "misc",
	"Markup":          "misc",
	"License":         "misc",
	"Gitignore":       "misc",
	"Annotated":       "git",
	"Note":            "git",
	"ChangedFile":     "git",
	"ExternalTracker": "repo",
	"ExternalWiki":    "repo",
	"InternalTracker": "repo",
	"Permission":      "common",
	"RepoTransfer":    "repo",
	"PayloadCommit":   "hook",
	"Dispatch":        "action",
	"Secret":          "action",
	"Variable":        "action",
	"Push":            "repo",
	"Mirror":          "repo",
	"Attachment":      "common",
	"EditDeadline":    "issue",
	"IssueDeadline":   "issue",
	"IssueLabels":     "issue",
	"IssueMeta":       "issue",
	"IssueTemplate":   "issue",
	"StateType":       "common",
	"TimeStamp":       "common",
	"Rename":          "admin",
	"Unadopted":       "admin",
}

// classifyType determines which file a type belongs in.
func classifyType(name string) string {
	// Direct match
	if group, ok := typeGrouping[name]; ok {
		return group
	}

	// Prefix match (longest first)
	for prefix, group := range typeGrouping {
		if strings.HasPrefix(name, prefix) {
			return group
		}
	}

	// Try common suffixes
	if strings.HasSuffix(name, "Option") || strings.HasSuffix(name, "Options") {
		// Strip Create/Edit prefix to find base
		trimmed := name
		trimmed = strings.TrimPrefix(trimmed, "Create")
		trimmed = strings.TrimPrefix(trimmed, "Edit")
		trimmed = strings.TrimPrefix(trimmed, "Delete")
		trimmed = strings.TrimPrefix(trimmed, "Update")
		trimmed = strings.TrimSuffix(trimmed, "Option")
		trimmed = strings.TrimSuffix(trimmed, "Options")
		if group, ok := typeGrouping[trimmed]; ok {
			return group
		}
	}

	return "misc"
}

// Generate writes Go source files for all types.
func Generate(types map[string]*GoType, pairs []CRUDPair, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// Group types by file
	groups := make(map[string][]*GoType)
	for _, gt := range types {
		file := classifyType(gt.Name)
		groups[file] = append(groups[file], gt)
	}

	// Sort types within each group
	for file := range groups {
		sort.Slice(groups[file], func(i, j int) bool {
			return groups[file][i].Name < groups[file][j].Name
		})
	}

	// Write each file
	for file, fileTypes := range groups {
		if err := writeFile(filepath.Join(outDir, file+".go"), fileTypes); err != nil {
			return fmt.Errorf("write %s.go: %w", file, err)
		}
	}

	return nil
}

var fileTmpl = template.Must(template.New("file").Parse(`// Code generated by forgegen from swagger.v1.json — DO NOT EDIT.

package types
{{if .NeedsTime}}
import "time"
{{end}}
{{range .Types}}
{{if .Description}}// {{.Name}} — {{.Description}}{{else}}// {{.Name}} represents a Forgejo API type.{{end}}
{{if .IsEnum}}type {{.Name}} string

const (
{{range .EnumValues}}	{{$.EnumConst .Name .}}	{{$.EnumType .Name}} = "{{.}}"
{{end}})
{{else}}type {{.Name}} struct {
{{range .Fields}}	{{.GoName}}	{{.GoType}}	` + "`" + `json:"{{.JSONName}}{{if not .Required}},omitempty{{end}}"` + "`" + `{{if .Comment}} // {{.Comment}}{{end}}
{{end}}}
{{end}}
{{end}}`))

type fileData struct {
	Types     []*GoType
	NeedsTime bool
}

func (fd fileData) EnumConst(typeName, value string) string {
	return typeName + pascalCase(value)
}

func (fd fileData) EnumType(typeName string) string {
	return typeName
}

func writeFile(path string, types []*GoType) error {
	needsTime := false
	for _, gt := range types {
		for _, f := range gt.Fields {
			if strings.Contains(f.GoType, "time.Time") {
				needsTime = true
				break
			}
		}
		if needsTime {
			break
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return fileTmpl.Execute(f, fileData{
		Types:     types,
		NeedsTime: needsTime,
	})
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v ./cmd/forgegen/ -run TestGenerate`
Expected: All 4 tests PASS

**Step 5: Commit**

```bash
git add cmd/forgegen/generator.go cmd/forgegen/generator_test.go
git commit -m "feat: Go source code generator from Swagger types

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 9: Generate types + verify compilation

**Files:**
- Create: `types/` directory with generated files
- Create: `types/generate.go` (go:generate directive)

**Step 1: Run the generator**

```bash
cd /Users/snider/Code/go-forge
mkdir -p types
go run ./cmd/forgegen/ -spec testdata/swagger.v1.json -out types/
```

**Step 2: Add go:generate directive**

Create `types/generate.go`:
```go
package types

//go:generate go run ../cmd/forgegen/ -spec ../testdata/swagger.v1.json -out .
```

**Step 3: Verify compilation**

Run: `cd /Users/snider/Code/go-forge && go build ./types/`
Expected: Compiles without errors

If there are compilation errors, fix the generator (`cmd/forgegen/generator.go`) and regenerate. Common issues:
- Missing imports (time)
- Duplicate field names (GoName collision)
- Invalid Go identifiers (reserved words, starting with numbers)

**Step 4: Run all tests**

Run: `cd /Users/snider/Code/go-forge && go test ./...`
Expected: All tests pass

**Step 5: Commit**

```bash
git add types/
git commit -m "feat: generate all 229 Forgejo API types from swagger spec

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

## Wave 3: Core Services (Tasks 10-13)

Each service follows the same pattern: embed `Resource[T,C,U]`, add action methods. The first service (Task 10) is fully detailed as a template. Subsequent services follow the same structure with less repetition.

### Task 10: Forge client + RepoService (template service)

**Files:**
- Create: `forge.go`
- Create: `repos.go`
- Create: `forge_test.go`

**Step 1: Write tests**

```go
package forge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forge.lthn.ai/core/go-forge/types"
)

func TestForge_Good_NewForge(t *testing.T) {
	f := NewForge("https://forge.lthn.ai", "tok")
	if f.Repos == nil {
		t.Fatal("Repos service is nil")
	}
	if f.Issues == nil {
		t.Fatal("Issues service is nil")
	}
}

func TestRepoService_Good_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "1")
		json.NewEncoder(w).Encode([]types.Repository{{Name: "go-forge"}})
	}))
	defer srv.Close()

	f := NewForge(srv.URL, "tok")
	result, err := f.Repos.List(context.Background(), Params{"org": "core"}, DefaultList)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "go-forge" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestRepoService_Good_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.Repository{Name: "go-forge", FullName: "core/go-forge"})
	}))
	defer srv.Close()

	f := NewForge(srv.URL, "tok")
	repo, err := f.Repos.Get(context.Background(), Params{"owner": "core", "repo": "go-forge"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.Name != "go-forge" {
		t.Errorf("got name=%q", repo.Name)
	}
}

func TestRepoService_Good_Fork(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(types.Repository{Name: "go-forge", Fork: true})
	}))
	defer srv.Close()

	f := NewForge(srv.URL, "tok")
	repo, err := f.Repos.Fork(context.Background(), "core", "go-forge", "my-org")
	if err != nil {
		t.Fatal(err)
	}
	if !repo.Fork {
		t.Error("expected fork=true")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/go-forge && go test -v -run "TestForge|TestRepoService"`
Expected: Compilation errors

**Step 3: Write forge.go**

```go
package forge

import "forge.lthn.ai/core/go-forge/types"

// Forge is the top-level client for the Forgejo API.
type Forge struct {
	client *Client

	Repos         *RepoService
	Issues        *IssueService
	Pulls         *PullService
	Orgs          *OrgService
	Users         *UserService
	Teams         *TeamService
	Admin         *AdminService
	Branches      *BranchService
	Releases      *ReleaseService
	Labels        *LabelService
	Webhooks      *WebhookService
	Notifications *NotificationService
	Packages      *PackageService
	Actions       *ActionsService
	Contents      *ContentService
	Wiki          *WikiService
	Misc          *MiscService
}

// NewForge creates a new Forge client.
func NewForge(url, token string, opts ...Option) *Forge {
	c := NewClient(url, token, opts...)
	f := &Forge{client: c}
	f.Repos = newRepoService(c)
	// Other services initialised in their respective tasks.
	// Stub them here so tests compile:
	f.Issues = &IssueService{}
	f.Pulls = &PullService{}
	f.Orgs = &OrgService{}
	f.Users = &UserService{}
	f.Teams = &TeamService{}
	f.Admin = &AdminService{}
	f.Branches = &BranchService{}
	f.Releases = &ReleaseService{}
	f.Labels = &LabelService{}
	f.Webhooks = &WebhookService{}
	f.Notifications = &NotificationService{}
	f.Packages = &PackageService{}
	f.Actions = &ActionsService{}
	f.Contents = &ContentService{}
	f.Wiki = &WikiService{}
	f.Misc = &MiscService{}
	return f
}

// Client returns the underlying HTTP client.
func (f *Forge) Client() *Client { return f.client }
```

**Step 4: Write repos.go**

```go
package forge

import (
	"context"

	"forge.lthn.ai/core/go-forge/types"
)

// RepoService handles repository operations.
type RepoService struct {
	Resource[types.Repository, types.CreateRepoOption, types.EditRepoOption]
}

func newRepoService(c *Client) *RepoService {
	return &RepoService{
		Resource: *NewResource[types.Repository, types.CreateRepoOption, types.EditRepoOption](
			c, "/api/v1/repos/{owner}/{repo}",
		),
	}
}

// ListOrgRepos returns all repositories for an organisation.
func (s *RepoService) ListOrgRepos(ctx context.Context, org string) ([]types.Repository, error) {
	return ListAll[types.Repository](ctx, s.client, "/api/v1/orgs/"+org+"/repos", nil)
}

// ListUserRepos returns all repositories for the authenticated user.
func (s *RepoService) ListUserRepos(ctx context.Context) ([]types.Repository, error) {
	return ListAll[types.Repository](ctx, s.client, "/api/v1/user/repos", nil)
}

// Fork forks a repository. If org is non-empty, forks into that organisation.
func (s *RepoService) Fork(ctx context.Context, owner, repo, org string) (*types.Repository, error) {
	body := map[string]string{}
	if org != "" {
		body["organization"] = org
	}
	var out types.Repository
	err := s.client.Post(ctx, "/api/v1/repos/"+owner+"/"+repo+"/forks", body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Migrate imports a repository from an external service.
func (s *RepoService) Migrate(ctx context.Context, opts *types.MigrateRepoOptions) (*types.Repository, error) {
	var out types.Repository
	err := s.client.Post(ctx, "/api/v1/repos/migrate", opts, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Transfer initiates a repository transfer.
func (s *RepoService) Transfer(ctx context.Context, owner, repo string, opts map[string]any) error {
	return s.client.Post(ctx, "/api/v1/repos/"+owner+"/"+repo+"/transfer", opts, nil)
}

// AcceptTransfer accepts a pending repository transfer.
func (s *RepoService) AcceptTransfer(ctx context.Context, owner, repo string) error {
	return s.client.Post(ctx, "/api/v1/repos/"+owner+"/"+repo+"/transfer/accept", nil, nil)
}

// RejectTransfer rejects a pending repository transfer.
func (s *RepoService) RejectTransfer(ctx context.Context, owner, repo string) error {
	return s.client.Post(ctx, "/api/v1/repos/"+owner+"/"+repo+"/transfer/reject", nil, nil)
}

// MirrorSync triggers a mirror sync.
func (s *RepoService) MirrorSync(ctx context.Context, owner, repo string) error {
	return s.client.Post(ctx, "/api/v1/repos/"+owner+"/"+repo+"/mirror-sync", nil, nil)
}
```

**Step 5: Write stub service types** so `forge.go` compiles. Create `services_stub.go`:

```go
package forge

// Stub service types — replaced as each service is implemented.

type IssueService struct{}
type PullService struct{}
type OrgService struct{}
type UserService struct{}
type TeamService struct{}
type AdminService struct{}
type BranchService struct{}
type ReleaseService struct{}
type LabelService struct{}
type WebhookService struct{}
type NotificationService struct{}
type PackageService struct{}
type ActionsService struct{}
type ContentService struct{}
type WikiService struct{}
type MiscService struct{}
```

**Step 6: Run tests**

Run: `cd /Users/snider/Code/go-forge && go test -v -run "TestForge|TestRepoService"`
Expected: All tests PASS (if generated types compile — if `types.CreateRepoOption` or `types.MigrateRepoOptions` don't exist, adjust field names to match generated types)

**Step 7: Commit**

```bash
git add forge.go repos.go services_stub.go forge_test.go
git commit -m "feat: Forge client + RepoService with CRUD and actions

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 11: IssueService + PullService

**Files:**
- Create: `issues.go`
- Create: `pulls.go`
- Create: `issues_test.go`
- Create: `pulls_test.go`
- Modify: `forge.go` (wire up services)
- Modify: `services_stub.go` (remove IssueService, PullService stubs)

Follow the same pattern as Task 10. Key points:

**IssueService** embeds `Resource[types.Issue, types.CreateIssueOption, types.EditIssueOption]`.
Path: `/api/v1/repos/{owner}/{repo}/issues/{index}`

Action methods (9):
- `Pin(ctx, owner, repo, index)` — POST `.../issues/{index}/pin`
- `Unpin(ctx, owner, repo, index)` — DELETE `.../issues/{index}/pin`
- `SetDeadline(ctx, owner, repo, index, deadline)` — POST `.../issues/{index}/deadline`
- `AddReaction(ctx, owner, repo, index, reaction)` — POST `.../issues/{index}/reactions`
- `DeleteReaction(ctx, owner, repo, index, reaction)` — DELETE `.../issues/{index}/reactions`
- `StartStopwatch(ctx, owner, repo, index)` — POST `.../issues/{index}/stopwatch/start`
- `StopStopwatch(ctx, owner, repo, index)` — POST `.../issues/{index}/stopwatch/stop`
- `AddLabels(ctx, owner, repo, index, labelIDs)` — POST `.../issues/{index}/labels`
- `RemoveLabel(ctx, owner, repo, index, labelID)` — DELETE `.../issues/{index}/labels/{id}`
- `ListComments(ctx, owner, repo, index)` — GET `.../issues/{index}/comments`
- `CreateComment(ctx, owner, repo, index, body)` — POST `.../issues/{index}/comments`

**PullService** embeds `Resource[types.PullRequest, types.CreatePullRequestOption, types.EditPullRequestOption]`.
Path: `/api/v1/repos/{owner}/{repo}/pulls/{index}`

Action methods (6):
- `Merge(ctx, owner, repo, index, method)` — POST `.../pulls/{index}/merge`
- `Update(ctx, owner, repo, index)` — POST `.../pulls/{index}/update`
- `ListReviews(ctx, owner, repo, index)` — GET `.../pulls/{index}/reviews`
- `SubmitReview(ctx, owner, repo, index, reviewID)` — POST `.../pulls/{index}/reviews/{id}`
- `DismissReview(ctx, owner, repo, index, reviewID, msg)` — POST `.../pulls/{index}/reviews/{id}/dismissals`
- `UndismissReview(ctx, owner, repo, index, reviewID)` — POST `.../pulls/{index}/reviews/{id}/undismissals`

Write tests for at least: List, Get, Create for each service + one action method each.

Run: `cd /Users/snider/Code/go-forge && go test ./... -v`
Commit: `git commit -m "feat: IssueService and PullService with actions"`

---

### Task 12: OrgService + TeamService + UserService

**Files:**
- Create: `orgs.go`, `teams.go`, `users.go`
- Create: `orgs_test.go`, `teams_test.go`, `users_test.go`
- Modify: `forge.go` (wire up)
- Modify: `services_stub.go` (remove stubs)

**OrgService** — `Resource[types.Organization, types.CreateOrgOption, types.EditOrgOption]`
Path: `/api/v1/orgs/{org}`
Actions: ListMembers, AddMember, RemoveMember, SetAvatar, Block, Unblock

**TeamService** — `Resource[types.Team, types.CreateTeamOption, types.EditTeamOption]`
Path: `/api/v1/teams/{id}`
Actions: ListMembers, AddMember, RemoveMember, ListRepos, AddRepo, RemoveRepo

**UserService** — `Resource[types.User, struct{}, struct{}]` (no create/edit via this path)
Path: `/api/v1/users/{username}`
Custom: `GetCurrent(ctx)`, `ListFollowers(ctx)`, `ListStarred(ctx)`, keys, GPG keys, settings

Run: `cd /Users/snider/Code/go-forge && go test ./... -v`
Commit: `git commit -m "feat: OrgService, TeamService, UserService"`

---

### Task 13: AdminService

**Files:**
- Create: `admin.go`
- Create: `admin_test.go`
- Modify: `forge.go` (wire up)
- Modify: `services_stub.go` (remove stub)

**AdminService** — No generic Resource (admin endpoints are heterogeneous).
Direct methods:
- `ListUsers(ctx)` — GET `/api/v1/admin/users`
- `CreateUser(ctx, opts)` — POST `/api/v1/admin/users`
- `EditUser(ctx, username, opts)` — PATCH `/api/v1/admin/users/{username}`
- `DeleteUser(ctx, username)` — DELETE `/api/v1/admin/users/{username}`
- `RenameUser(ctx, username, newName)` — POST `.../users/{username}/rename`
- `ListOrgs(ctx)` — GET `/api/v1/admin/orgs`
- `RunCron(ctx, task)` — POST `/api/v1/admin/cron/{task}`
- `ListCron(ctx)` — GET `/api/v1/admin/cron`
- `AdoptRepo(ctx, owner, repo)` — POST `.../unadopted/{owner}/{repo}`
- `GenerateRunnerToken(ctx)` — POST `/api/v1/admin/runners/registration-token`

Run: `cd /Users/snider/Code/go-forge && go test ./... -v`
Commit: `git commit -m "feat: AdminService with user, org, cron, runner operations"`

---

## Wave 4: Extended Services (Tasks 14-17)

### Task 14: BranchService + ReleaseService

**BranchService** — `Resource[types.Branch, types.CreateBranchRepoOption, struct{}]`
Path: `/api/v1/repos/{owner}/{repo}/branches/{branch}`
Additional: BranchProtection CRUD at `.../branch_protections/{name}`

**ReleaseService** — `Resource[types.Release, types.CreateReleaseOption, types.EditReleaseOption]`
Path: `/api/v1/repos/{owner}/{repo}/releases/{id}`
Additional: Asset upload/download at `.../releases/{id}/assets`

### Task 15: LabelService + WebhookService + ContentService

**LabelService** — Handles repo labels, org labels, and issue labels.
- `ListRepoLabels(ctx, owner, repo)`
- `CreateRepoLabel(ctx, owner, repo, opts)`
- `ListOrgLabels(ctx, org)`

**WebhookService** — `Resource[types.Hook, types.CreateHookOption, types.EditHookOption]`
Actions: `TestHook(ctx, owner, repo, id)`

**ContentService** — File read/write via API
- `GetFile(ctx, owner, repo, path)` — GET `.../contents/{path}`
- `CreateFile(ctx, owner, repo, path, opts)` — POST `.../contents/{path}`
- `UpdateFile(ctx, owner, repo, path, opts)` — PUT `.../contents/{path}`
- `DeleteFile(ctx, owner, repo, path, opts)` — DELETE `.../contents/{path}`

### Task 16: ActionsService + NotificationService + PackageService

**ActionsService** — runners, secrets, variables, workflow dispatch
- Repo-level: `.../repos/{owner}/{repo}/actions/{secrets,variables,runners}`
- Org-level: `.../orgs/{org}/actions/{secrets,variables,runners}`
- `DispatchWorkflow(ctx, owner, repo, workflow, opts)`

**NotificationService** — list, mark read
- `List(ctx)` — GET `/api/v1/notifications`
- `MarkRead(ctx)` — PUT `/api/v1/notifications`
- `GetThread(ctx, id)` — GET `.../notifications/threads/{id}`

**PackageService** — list, get, delete
- `List(ctx, owner)` — GET `/api/v1/packages/{owner}`
- `Get(ctx, owner, type, name, version)` — GET `.../packages/{owner}/{type}/{name}/{version}`

### Task 17: WikiService + MiscService + CommitService

**WikiService** — pages
- `ListPages(ctx, owner, repo)`
- `GetPage(ctx, owner, repo, pageName)`
- `CreatePage(ctx, owner, repo, opts)`
- `EditPage(ctx, owner, repo, pageName, opts)`
- `DeletePage(ctx, owner, repo, pageName)`

**MiscService** — markdown, licenses, gitignore, nodeinfo
- `RenderMarkdown(ctx, text, mode)` — POST `/api/v1/markdown`
- `ListLicenses(ctx)` — GET `/api/v1/licenses`
- `ListGitignoreTemplates(ctx)` — GET `/api/v1/gitignore/templates`
- `NodeInfo(ctx)` — GET `/api/v1/nodeinfo`

**CommitService** — status and notes
- `GetCombinedStatus(ctx, owner, repo, ref)`
- `CreateStatus(ctx, owner, repo, sha, opts)`
- `SetNote(ctx, owner, repo, sha, opts)`

For each task in Wave 4: write tests first, implement, verify all tests pass, commit.

Run after each task: `cd /Users/snider/Code/go-forge && go test ./... -v`

---

## Wave 5: Clean Up + Services Stub Removal (Task 18)

### Task 18: Remove stubs + final wiring

**Files:**
- Delete: `services_stub.go`
- Modify: `forge.go` — replace all stub initialisations with real `newXxxService(c)` calls

**Step 1: Remove services_stub.go**

Delete the file. All service types should now be defined in their own files.

**Step 2: Wire all services in forge.go**

Update `NewForge()` to call `newXxxService(c)` for every service.

**Step 3: Run all tests**

Run: `cd /Users/snider/Code/go-forge && go test ./... -v -count=1`
Expected: All tests pass

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: wire all 17 services, remove stubs

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

## Wave 6: Integration + Forge Repo Setup (Tasks 19-20)

### Task 19: Create Forge repo + push

**Step 1: Create repo on Forge**

Use the Forgejo API or web UI to create `core/go-forge` on `forge.lthn.ai`.

**Step 2: Add remote and push**

```bash
cd /Users/snider/Code/go-forge
git remote add forge ssh://git@forge.lthn.ai:2223/core/go-forge.git
git push -u forge main
```

### Task 20: Wiki documentation (go-ai treatment)

Create wiki pages for go-forge on Forge, matching the go-ai documentation pattern:

1. **Home** — Overview, install, quick start
2. **Architecture** — Generic Resource[T,C,U], codegen pipeline, service pattern
3. **Services** — All 17 services with example usage
4. **Code Generation** — How to regenerate types, upgrade Forgejo version
5. **Configuration** — Env vars, config file, flags
6. **Error Handling** — APIError, IsNotFound, IsForbidden
7. **Development** — Contributing, testing, releasing

Use the Forge wiki API: `POST /api/v1/repos/core/go-forge/wiki/new` with `{"content_base64":"...","title":"..."}`.

---

## Dependency Sequencing

```
Task 1 (scaffold) ← Task 2 (client) ← Task 3 (pagination) ← Task 4 (params) ← Task 5 (resource)
Task 1 ← Task 7 (parser) ← Task 8 (generator) ← Task 9 (generate types)
Task 5 + Task 9 ← Task 6 (config) ← Task 10 (forge + repos)
Task 10 ← Task 11 (issues + PRs)
Task 10 ← Task 12 (orgs + teams + users)
Task 10 ← Task 13 (admin)
Task 10 ← Task 14-17 (extended services)
Task 14-17 ← Task 18 (remove stubs)
Task 18 ← Task 19 (forge push)
Task 19 ← Task 20 (wiki)
```

**Wave 1 (Tasks 1-6)**: Foundation — all independent once scaffolded
**Wave 2 (Tasks 7-9)**: Codegen — sequential (parser → generator → run)
**Wave 3 (Tasks 10-13)**: Core services — Task 10 first (creates Forge + stubs), then 11-13 parallel
**Wave 4 (Tasks 14-17)**: Extended services — all parallel after Task 10
**Wave 5 (Task 18)**: Clean up — after all services done
**Wave 6 (Tasks 19-20)**: Ship — after clean up

## Verification

After all tasks:

1. `cd /Users/snider/Code/go-forge && go test ./... -count=1` — all pass
2. `go build ./...` — compiles cleanly
3. `go vet ./...` — no issues
4. Verify `types/` contains generated files with `Repository`, `Issue`, `PullRequest`, etc.
5. Verify `NewForge()` creates client with all 17 services populated
6. Verify action methods exist (Fork, Merge, Pin, etc.)
