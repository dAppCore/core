# go-forge Design Document

## Overview

**go-forge** is a full-coverage Go client for the Forgejo API (450 endpoints, 284 paths, 229 types). It uses a generic `Resource[T, C, U]` pattern for CRUD operations (91% of endpoints) and hand-written methods for 39 unique action endpoints. Types are generated from Forgejo's `swagger.v1.json` spec.

**Module path:** `forge.lthn.ai/core/go-forge`

**Origin:** Extracted from `go-scm/forge/` (45 methods covering 10% of API), expanded to full coverage.

## Architecture

```
forge.lthn.ai/core/go-forge
├── client.go          # HTTP client: auth, headers, rate limiting, context.Context
├── pagination.go      # Generic paginated request helper
├── resource.go        # Resource[T, C, U] generic CRUD (List/Get/Create/Update/Delete)
├── errors.go          # Typed error handling (APIError, NotFound, Forbidden, etc.)
├── forge.go           # Top-level Forge client aggregating all services
│
├── types/             # Generated from swagger.v1.json
│   ├── generate.go    # //go:generate directive
│   ├── repo.go        # Repository, CreateRepoOption, EditRepoOption
│   ├── issue.go       # Issue, CreateIssueOption, EditIssueOption
│   ├── pr.go          # PullRequest, CreatePullRequestOption
│   ├── user.go        # User, CreateUserOption
│   ├── org.go         # Organisation, CreateOrgOption
│   ├── team.go        # Team, CreateTeamOption
│   ├── label.go       # Label, CreateLabelOption
│   ├── release.go     # Release, CreateReleaseOption
│   ├── branch.go      # Branch, BranchProtection
│   ├── milestone.go   # Milestone, CreateMilestoneOption
│   ├── hook.go        # Hook, CreateHookOption
│   ├── key.go         # DeployKey, PublicKey, GPGKey
│   ├── notification.go # NotificationThread, NotificationSubject
│   ├── package.go     # Package, PackageFile
│   ├── action.go      # ActionRunner, ActionSecret, ActionVariable
│   ├── commit.go      # Commit, CommitStatus, CombinedStatus
│   ├── content.go     # ContentsResponse, FileOptions
│   ├── wiki.go        # WikiPage, WikiPageMetaData
│   ├── review.go      # PullReview, PullReviewComment
│   ├── reaction.go    # Reaction
│   ├── topic.go       # TopicResponse
│   ├── misc.go        # Markdown, License, GitignoreTemplate, NodeInfo
│   ├── admin.go       # Cron, QuotaGroup, QuotaRule
│   ├── activity.go    # Activity, Feed
│   └── common.go      # Shared types: Permission, ExternalTracker, etc.
│
├── repos.go           # RepoService: CRUD + fork, mirror, transfer, template
├── issues.go          # IssueService: CRUD + pin, deadline, reactions, stopwatch
├── pulls.go           # PullService: CRUD + merge, update, reviews, dismiss
├── orgs.go            # OrgService: CRUD + members, avatar, block, hooks
├── users.go           # UserService: CRUD + keys, followers, starred, settings
├── teams.go           # TeamService: CRUD + members, repos
├── admin.go           # AdminService: users, orgs, cron, runners, quota, unadopted
├── branches.go        # BranchService: CRUD + protection rules
├── releases.go        # ReleaseService: CRUD + assets
├── labels.go          # LabelService: repo + org + issue labels
├── webhooks.go        # WebhookService: CRUD + test hook
├── notifications.go   # NotificationService: list, mark read
├── packages.go        # PackageService: list, get, delete
├── actions.go         # ActionsService: runners, secrets, variables, workflow dispatch
├── contents.go        # ContentService: file read/write/delete via API
├── wiki.go            # WikiService: pages
├── commits.go         # CommitService: status, notes, diff
├── misc.go            # MiscService: markdown, licenses, gitignore, nodeinfo
│
├── config.go          # URL/token resolution: env → config file → flags
│
├── cmd/forgegen/      # Code generator: swagger.v1.json → types/*.go
│   ├── main.go
│   ├── parser.go      # Parse OpenAPI 2.0 definitions
│   ├── generator.go   # Render Go source files
│   └── templates/     # Go text/template files for codegen
│
└── testdata/
    └── swagger.v1.json  # Pinned spec for testing + generation
```

## Key Design Decisions

### 1. Generic Resource[T, C, U]

Three type parameters: T (resource type), C (create options), U (update options).

```go
type Resource[T any, C any, U any] struct {
    client *Client
    path   string // e.g. "/api/v1/repos/{owner}/{repo}/issues"
}

func (r *Resource[T, C, U]) List(ctx context.Context, params Params, opts ListOptions) ([]T, error)
func (r *Resource[T, C, U]) Get(ctx context.Context, params Params, id string) (*T, error)
func (r *Resource[T, C, U]) Create(ctx context.Context, params Params, body *C) (*T, error)
func (r *Resource[T, C, U]) Update(ctx context.Context, params Params, id string, body *U) (*T, error)
func (r *Resource[T, C, U]) Delete(ctx context.Context, params Params, id string) error
```

`Params` is `map[string]string` resolving path variables: `{"owner": "core", "repo": "go-forge"}`.

This covers 411 of 450 endpoints (91%).

### 2. Service Structs Embed Resource

```go
type IssueService struct {
    Resource[types.Issue, types.CreateIssueOption, types.EditIssueOption]
}

// CRUD comes free. Actions are hand-written:
func (s *IssueService) Pin(ctx context.Context, owner, repo string, index int64) error
func (s *IssueService) SetDeadline(ctx context.Context, owner, repo string, index int64, deadline *time.Time) error
```

### 3. Top-Level Forge Client

```go
type Forge struct {
    client        *Client
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
    Commits       *CommitService
    Misc          *MiscService
}

func NewForge(url, token string, opts ...Option) *Forge
```

### 4. Codegen from swagger.v1.json

The `cmd/forgegen/` tool reads the OpenAPI 2.0 spec and generates:
- Go struct definitions with JSON tags and doc comments
- Enum constants
- Type mapping (OpenAPI → Go)

229 type definitions → ~25 grouped Go files in `types/`.

Type mapping rules:
| OpenAPI | Go |
|---------|-----|
| `string` | `string` |
| `string` + `date-time` | `time.Time` |
| `integer` + `int64` | `int64` |
| `integer` | `int` |
| `boolean` | `bool` |
| `array` of T | `[]T` |
| `$ref` | `*T` (pointer) |
| nullable | pointer type |
| `binary` | `[]byte` |

### 5. HTTP Client

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    userAgent  string
}

func New(url, token string, opts ...Option) *Client

func (c *Client) Get(ctx context.Context, path string, out any) error
func (c *Client) Post(ctx context.Context, path string, body, out any) error
func (c *Client) Patch(ctx context.Context, path string, body, out any) error
func (c *Client) Put(ctx context.Context, path string, body, out any) error
func (c *Client) Delete(ctx context.Context, path string) error
```

Options: `WithHTTPClient`, `WithUserAgent`, `WithRateLimit`, `WithLogger`.

### 6. Pagination

Forgejo uses `page` + `limit` query params and `X-Total-Count` response header.

```go
type ListOptions struct {
    Page  int
    Limit int // default 50, max configurable
}

type PagedResult[T any] struct {
    Items      []T
    TotalCount int
    Page       int
    HasMore    bool
}

// ListAll fetches all pages automatically.
func (r *Resource[T, C, U]) ListAll(ctx context.Context, params Params) ([]T, error)
```

### 7. Error Handling

```go
type APIError struct {
    StatusCode int
    Message    string
    URL        string
}

func IsNotFound(err error) bool
func IsForbidden(err error) bool
func IsConflict(err error) bool
```

### 8. Config Resolution (from go-scm/forge)

Priority: flags → environment → config file.

```go
func NewFromConfig(flagURL, flagToken string) (*Forge, error)
func ResolveConfig(flagURL, flagToken string) (url, token string, err error)
func SaveConfig(url, token string) error
```

Env vars: `FORGE_URL`, `FORGE_TOKEN`. Config file: `~/.config/forge/config.json`.

## API Coverage

| Category | Endpoints | CRUD | Actions |
|----------|-----------|------|---------|
| Repository | 175 | 165 | 10 (fork, mirror, transfer, template, avatar, diffpatch) |
| User | 74 | 70 | 4 (avatar, GPG verify) |
| Issue | 67 | 57 | 10 (pin, deadline, reactions, stopwatch, labels) |
| Organisation | 63 | 59 | 4 (avatar, block/unblock) |
| Admin | 39 | 35 | 4 (cron run, rename, adopt, quota set) |
| Miscellaneous | 12 | 7 | 5 (markdown render, markup, nodeinfo) |
| Notification | 7 | 7 | 0 |
| ActivityPub | 6 | 3 | 3 (inbox POST) |
| Package | 4 | 4 | 0 |
| Settings | 4 | 4 | 0 |
| **Total** | **450** | **411** | **39** |

## Integration Points

### go-api

Services implement `DescribableGroup` from go-api Phase 3, enabling:
- REST endpoint generation via ToolBridge
- Auto-generated OpenAPI spec
- Multi-language SDK codegen

### go-scm

go-scm/forge/ becomes a thin adapter importing go-forge types. Existing go-scm users are unaffected — the multi-provider abstraction layer stays.

### go-ai/mcp

The MCP subsystem can register go-forge operations as MCP tools, giving AI agents full Forgejo API access.

## 39 Unique Action Methods

These require hand-written implementation:

**Repository:** migrate, fork, generate (template), transfer, accept/reject transfer, mirror sync, push mirror sync, avatar, diffpatch, contents (multi-file modify)

**Pull Requests:** merge, update (rebase), submit review, dismiss/undismiss review

**Issues:** pin, set deadline, add reaction, start/stop stopwatch, add issue labels

**Comments:** add reaction

**Admin:** run cron task, adopt unadopted, rename user, set quota groups

**Misc:** render markdown, render raw markdown, render markup, GPG key verify

**ActivityPub:** inbox POST (actor, repo, user)

**Actions:** dispatch workflow

**Git:** set note on commit, test webhook
