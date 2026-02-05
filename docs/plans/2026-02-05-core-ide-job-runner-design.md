# Core-IDE Job Runner Design

**Date:** 2026-02-05
**Status:** Approved
**Author:** @Snider + Claude

---

## Goal

Turn core-ide into an autonomous job runner that polls for actionable pipeline work, executes it via typed MCP tool handlers, captures JSONL training data, and self-updates. Supports 12 nodes running headless on servers and desktop on developer machines.

---

## Architecture Overview

```
+-------------------------------------------------+
|                   core-ide                       |
|                                                  |
|  +----------+   +-----------+   +----------+    |
|  | Poller   |-->| Dispatcher|-->| Handler  |    |
|  | (Source) |   | (MCP route)|  | Registry |    |
|  +----------+   +-----------+   +----------+    |
|       |              |              |            |
|       |         +----v----+    +---v-------+    |
|       |         | Journal |    | JobSource |    |
|       |         | (JSONL) |    | (adapter) |    |
|       |         +---------+    +-----------+    |
|  +----v-----+                                   |
|  | Updater  |  (existing internal/cmd/updater)  |
|  +----------+                                   |
+-------------------------------------------------+
```

**Three components:**
- **Poller** -- Periodic scan via pluggable JobSource adapters. Builds PipelineSignal structs from API responses. Never reads comment bodies (injection vector).
- **Dispatcher** -- Matches signals against handler registry in priority order. One action per signal per cycle (prevents cascades).
- **Journal** -- Appends JSONL after each completed action per issue-epic step 10 spec. Structural signals only -- IDs, SHAs, timestamps, cycle counts, instructions sent, automations performed.

---

## Job Source Abstraction

GitHub is the first adapter. The platform's own Agentic API replaces it later. Handler logic is source-agnostic.

```go
type JobSource interface {
    Name() string
    Poll(ctx context.Context) ([]*PipelineSignal, error)
    Report(ctx context.Context, result *ActionResult) error
}
```

| Adapter           | When  | Transport                              |
|-------------------|-------|----------------------------------------|
| `GitHubSource`    | Now   | REST API + conditional requests (ETag) |
| `HostUKSource`    | Next  | Agentic API (WebSocket or poll)        |
| `HyperswarmSource`| Later | P2P encrypted channels via Holepunch   |

**Multi-source:** Poller runs multiple sources concurrently. Own repos get priority. When idle (zero signals for N consecutive cycles), external project sources activate (WailsApp first).

**API budget:** 50% credit allocation for harvest mode is a config value on the source, not hardcoded.

---

## Pipeline Signal

The structural snapshot passed to handlers. Never contains comment bodies or free text.

```go
type PipelineSignal struct {
    EpicNumber      int
    ChildNumber     int
    PRNumber        int
    RepoOwner       string
    RepoName        string
    PRState         string    // OPEN, MERGED, CLOSED
    IsDraft         bool
    Mergeable       string    // MERGEABLE, CONFLICTING, UNKNOWN
    CheckStatus     string    // SUCCESS, FAILURE, PENDING
    ThreadsTotal    int
    ThreadsResolved int
    LastCommitSHA   string
    LastCommitAt    time.Time
    LastReviewAt    time.Time
}
```

---

## Handler Registry

Each action from the issue-epic flow is a registered handler. All Go functions with typed inputs/outputs.

```go
type JobHandler interface {
    Name() string
    Match(signal *PipelineSignal) bool
    Execute(ctx context.Context, signal *PipelineSignal) (*ActionResult, error)
}
```

| Handler            | Epic Stage | Input Signals                                     | Action                                      |
|--------------------|-----------|---------------------------------------------------|---------------------------------------------|
| `publish_draft`    | 3         | PR draft=true, checks=SUCCESS                     | Mark PR as ready for review                 |
| `send_fix_command` | 4/6       | PR CONFLICTING or threads without fix commit       | Comment "fix merge conflict" / "fix the code reviews" |
| `resolve_threads`  | 5         | Unresolved threads, fix commit exists after review | Resolve all pre-commit threads              |
| `enable_auto_merge`| 7         | PR MERGEABLE, checks passing, threads resolved     | Enable auto-merge via API                   |
| `tick_parent`      | 8         | Child PR merged                                    | Update epic issue checklist                 |
| `close_child`      | 9         | Child PR merged + parent ticked                    | Close child issue                           |
| `capture_journal`  | 10        | Any completed action                               | Append JSONL entry                          |

**ActionResult** carries what was done -- action name, target IDs, success/failure, timestamps. Feeds directly into JSONL journal.

Handlers register at init time, same pattern as CLI commands in the existing codebase.

---

## Headless vs Desktop Mode

Same binary, same handlers, different UI surface.

**Detection:**

```go
func hasDisplay() bool {
    if runtime.GOOS == "windows" { return true }
    return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}
```

**Headless mode** (Linux server, no display):
- Skip Wails window creation
- Start poller immediately
- Start MCP bridge (port 9877) for external tool access
- Log to stdout/file (structured JSON)
- Updater: check on startup, auto-apply + restart via watcher
- Managed by systemd: `Restart=always`

**Desktop mode** (display available):
- Full Wails system tray + webview panel
- Tray icon shows status: idle, polling, executing, error
- Tray menu: Start/Stop poller, Force update, Open journal, Configure sources
- Poller off by default (developer toggle)
- Same MCP bridge, same handlers, same journal

**CLI override:** `core-ide --headless` forces headless. `core-ide --desktop` forces GUI.

**Shared startup:**

```go
func main() {
    // 1. Load config (repos, interval, channel, sources)
    // 2. Build handler registry
    // 3. Init journal
    // 4. Init updater (check on startup)
    // 5. Branch:
    if hasDisplay() {
        startDesktop()  // Wails + tray + optional poller
    } else {
        startHeadless() // Poller + MCP bridge + signal handling
    }
}
```

---

## Poller Configuration

```go
type PollerConfig struct {
    Sources      []JobSource
    Handlers     []JobHandler
    Journal      *Journal
    PollInterval time.Duration // default: 60s
    DryRun       bool          // log without executing
}
```

**Rate limiting:** GitHub API allows 5000 req/hr with token. Full scan of 4 repos with ~30 PRs uses ~150 requests. Poller uses conditional requests (If-None-Match/ETag) to avoid counting unchanged responses. Backs off to 5min interval when idle.

**CLI flags:**
- `--poll-interval` (default: 60s)
- `--repos` (comma-separated: `host-uk/core,host-uk/core-php`)
- `--dry-run` (log actions without executing)
- `--headless` / `--desktop` (mode override)

---

## Self-Update

Uses existing `internal/cmd/updater` package. Binary-safe replacement with platform-specific watcher process, SemVer channel selection (stable/beta/alpha/dev), automatic rollback on failure.

**Integration:**
- Headless: `CheckAndUpdateOnStartup` -- auto-apply + restart
- Desktop: `CheckOnStartup` -- notify via tray, user confirms

---

## Training Data (Journal)

JSONL format per issue-epic step 10. One record per completed action.

```json
{
  "ts": "2026-02-05T12:00:00Z",
  "epic": 299,
  "child": 212,
  "pr": 316,
  "repo": "host-uk/core",
  "action": "publish_draft",
  "signals": {
    "pr_state": "OPEN",
    "is_draft": true,
    "check_status": "SUCCESS",
    "mergeable": "UNKNOWN",
    "threads_total": 0,
    "threads_resolved": 0
  },
  "result": {
    "success": true,
    "duration_ms": 340
  },
  "cycle": 1
}
```

**Rules:**
- NO content (no comments, no messages, no bodies)
- Structural signals only -- safe for training
- Append-only JSONL file per node
- File path: `~/.core/journal/<repo>/<date>.jsonl`

---

## Files Summary

| File | Action |
|------|--------|
| `pkg/jobrunner/types.go` | CREATE -- JobSource, JobHandler, PipelineSignal, ActionResult interfaces |
| `pkg/jobrunner/poller.go` | CREATE -- Poller, Dispatcher, multi-source orchestration |
| `pkg/jobrunner/journal.go` | CREATE -- JSONL writer, append-only, structured records |
| `pkg/jobrunner/github/source.go` | CREATE -- GitHubSource adapter, conditional requests |
| `pkg/jobrunner/github/signals.go` | CREATE -- PR/issue state extraction, signal building |
| `internal/core-ide/handlers/publish_draft.go` | CREATE -- Publish draft PR handler |
| `internal/core-ide/handlers/resolve_threads.go` | CREATE -- Resolve review threads handler |
| `internal/core-ide/handlers/send_fix_command.go` | CREATE -- Send fix command handler |
| `internal/core-ide/handlers/enable_auto_merge.go` | CREATE -- Enable auto-merge handler |
| `internal/core-ide/handlers/tick_parent.go` | CREATE -- Tick epic checklist handler |
| `internal/core-ide/handlers/close_child.go` | CREATE -- Close child issue handler |
| `internal/core-ide/main.go` | MODIFY -- Headless/desktop branching, poller integration |
| `internal/core-ide/mcp_bridge.go` | MODIFY -- Register job handlers as MCP tools |

---

## What Doesn't Ship Yet

- HostUK Agentic API adapter (future -- replaces GitHub)
- Hyperswarm P2P adapter (future)
- External project scanning / harvest mode (future -- WailsApp first)
- LoRA training pipeline (separate concern -- reads JSONL journal)

---

## Testing Strategy

- **Handlers:** Unit-testable. Mock PipelineSignal in, assert API calls out.
- **Poller:** httptest server returning fixture responses.
- **Journal:** Read back JSONL, verify schema.
- **Integration:** Dry-run mode against real repos, verify signals match expected state.
