# BugSETI HubService Design

## Overview

A thin HTTP client service in the BugSETI desktop app that coordinates with the agentic portal's `/api/bugseti/*` endpoints. Prevents duplicate work across the 11 community testers, aggregates stats for leaderboard, and registers client instances.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Target | Direct to portal API | Endpoints built for this purpose |
| Auth | Auto-register via forge token | No manual key management for users |
| Sync strategy | Lazy/manual | User-triggered claims, manual stats sync |
| Offline mode | Offline-first | Queue failed writes, retry on reconnect |
| Approach | Thin HTTP client (net/http) | Matches existing patterns, no deps |

## Architecture

**File:** `internal/bugseti/hub.go` + `hub_test.go`

```
HubService
├── HTTP client (net/http, 10s timeout)
├── Auth: auto-register via forge token → cached ak_ token
├── Config: HubURL, HubToken, ClientID in ConfigService
├── Offline-first: queue failed writes, drain on next success
└── Lazy sync: user-triggered, no background goroutines
```

**Dependencies:** ConfigService only.

**Integration:**
- QueueService calls `hub.ClaimIssue()` when user picks an issue
- SubmitService calls `hub.UpdateStatus("completed")` after PR
- TrayService calls `hub.GetLeaderboard()` from UI
- main.go calls `hub.Register()` on startup

## Data Types

```go
type HubClient struct {
    ClientID  string    // UUID, generated once, persisted in config
    Name      string    // e.g. "Snider's MacBook"
    Version   string    // bugseti.GetVersion()
    OS        string    // runtime.GOOS
    Arch      string    // runtime.GOARCH
}

type HubClaim struct {
    IssueID     string  // "owner/repo#123"
    Repo        string
    IssueNumber int
    Title       string
    URL         string
    Status      string  // claimed|in_progress|completed|skipped
    ClaimedAt   time.Time
    PRUrl       string
    PRNumber    int
}

type LeaderboardEntry struct {
    Rank            int
    ClientName      string
    IssuesCompleted int
    PRsSubmitted    int
    PRsMerged       int
    CurrentStreak   int
}

type GlobalStats struct {
    TotalParticipants  int
    ActiveParticipants int
    TotalIssuesCompleted int
    TotalPRsMerged     int
    ActiveClaims       int
}
```

## API Mapping

| Method | HTTP | Endpoint | Trigger |
|--------|------|----------|---------|
| `Register()` | POST /register | App startup |
| `Heartbeat()` | POST /heartbeat | Manual / periodic if enabled |
| `ClaimIssue(issue)` | POST /issues/claim | User picks issue |
| `UpdateStatus(id, status)` | PATCH /issues/{id}/status | PR submitted, skip |
| `ReleaseClaim(id)` | DELETE /issues/{id}/claim | User abandons |
| `IsIssueClaimed(id)` | GET /issues/{id} | Before showing issue |
| `ListClaims(filters)` | GET /issues/claimed | UI active claims view |
| `SyncStats(stats)` | POST /stats/sync | Manual from UI |
| `GetLeaderboard(limit)` | GET /leaderboard | UI leaderboard view |
| `GetGlobalStats()` | GET /stats | UI stats dashboard |

## Auto-Register Flow

New endpoint on portal:

```
POST /api/bugseti/auth/forge
Body: { "forge_url": "https://forge.lthn.io", "forge_token": "..." }
```

Portal validates token against Forgejo API (`/api/v1/user`), creates an AgentApiKey with `bugseti.read` + `bugseti.write` scopes, returns `{ "api_key": "ak_..." }`.

HubService caches the `ak_` token in config.json. On 401, clears cached token and re-registers.

## Error Handling

| Error | Behaviour |
|-------|-----------|
| Network unreachable | Log, queue write ops, return cached reads |
| 401 Unauthorised | Clear token, re-register via forge |
| 409 Conflict (claim) | Return "already claimed" — not an error |
| 404 (claim not found) | Return nil |
| 429 Rate limited | Back off, queue the op |
| 5xx Server error | Log, queue write ops |

**Pending operations queue:**
- Failed writes stored in `[]PendingOp`, persisted to `$DataDir/hub_pending.json`
- Drained on next successful user-triggered call (no background goroutine)
- Each op has: method, path, body, created_at

## Config Changes

New fields in `Config` struct:

```go
HubURL    string `json:"hubUrl,omitempty"`    // portal API base URL
HubToken  string `json:"hubToken,omitempty"`  // cached ak_ token
ClientID  string `json:"clientId,omitempty"`  // UUID, generated once
ClientName string `json:"clientName,omitempty"` // display name
```

## Files Changed

| File | Action |
|------|--------|
| `internal/bugseti/hub.go` | New — HubService |
| `internal/bugseti/hub_test.go` | New — httptest-based tests |
| `internal/bugseti/config.go` | Edit — add Hub* + ClientID fields |
| `cmd/bugseti/main.go` | Edit — create + register HubService |
| `cmd/bugseti/tray.go` | Edit — leaderboard/stats menu items |
| Laravel: auth controller | New — `/api/bugseti/auth/forge` |

## Testing

- `httptest.NewServer` mocks for all endpoints
- Test success, network error, 409 conflict, 401 re-auth flows
- Test pending ops queue: add when offline, drain on reconnect
- `_Good`, `_Bad`, `_Ugly` naming convention
