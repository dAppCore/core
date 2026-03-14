# Lint Pattern Catalog & Polish Skill Design

> **Partial implementation (14 Mar 2026):** Layer 1 (`core/lint` -- catalog, matcher, scanner, CLI) is fully implemented and documented at `docs/tools/lint/index.md`. Layer 2 (MCP subsystem in `go-ai`) and Layer 3 (Claude Code polish skill in `core/agent`) are NOT implemented. This plan is retained for those remaining layers.

**Goal:** A structured pattern catalog (`core/lint`) that captures recurring code quality findings as regex rules, exposes them via MCP tools in `go-ai`, and orchestrates multi-AI code review via a Claude Code skill in `core/agent`.

**Architecture:** Three layers — a standalone catalog+matcher library (`core/lint`), an MCP subsystem in `go-ai` that exposes lint tools to agents, and a Claude Code plugin in `core/agent` that orchestrates the "polish" workflow (deterministic checks + AI reviewers + feedback loop into the catalog).

**Tech Stack:** Go (catalog, matcher, CLI, MCP subsystem), YAML (rule definitions), JSONL (findings output, compatible with `~/.core/ai/metrics/`), Claude Code plugin format (hooks.json, commands/*.md, plugin.json).

---

## Context

During a code review sweep of 18 Go repos (March 2026), AI reviewers (Gemini, Claude) found ~20 recurring patterns: SQL injection, path traversal, XSS, missing constant-time comparison, goroutine leaks, Go 1.26 modernisation opportunities, and more. Many of these patterns repeat across repos.

Currently these findings exist only as commit messages. This design captures them as a reusable, machine-readable catalog that:
1. Deterministic tools can run immediately (regex matching)
2. MCP-connected agents can query and apply
3. LEM models can train on for "does this comply with CoreGo standards?" judgements
4. Grows automatically as AI reviewers find new patterns

## Layer 1: `core/lint` — Pattern Catalog & Matcher

### Repository Structure

```
core/lint/
├── go.mod                        # forge.lthn.ai/core/lint
├── catalog/
│   ├── go-security.yaml          # SQL injection, path traversal, XSS, constant-time
│   ├── go-modernise.yaml         # Go 1.26: slices.Clone, wg.Go, maps.Keys, range-over-int
│   ├── go-correctness.yaml       # Deadlocks, goroutine leaks, nil guards, error handling
│   ├── php-security.yaml         # XSS, CSRF, mass assignment, SQL injection
│   ├── ts-security.yaml          # DOM XSS, prototype pollution
│   └── cpp-safety.yaml           # Buffer overflow, use-after-free
├── pkg/lint/
│   ├── catalog.go                # Load + parse YAML catalog files
│   ├── rule.go                   # Rule struct definition
│   ├── matcher.go                # Regex matcher against file contents
│   ├── report.go                 # Structured findings output (JSON/JSONL/text)
│   ├── catalog_test.go
│   ├── matcher_test.go
│   └── report_test.go
├── cmd/core-lint/
│   └── main.go                   # `core-lint check ./...` CLI
└── .core/
    └── build.yaml                # Produces core-lint binary
```

### Rule Schema (YAML)

```yaml
- id: go-sec-001
  title: "SQL wildcard injection in LIKE clauses"
  severity: high          # critical, high, medium, low, info
  languages: [go]
  tags: [security, injection, owasp-a03]
  pattern: 'LIKE\s+\?\s*,\s*["\x60]%\s*\+'
  exclude_pattern: 'EscapeLike'              # suppress if this also matches
  fix: "Use parameterised LIKE with explicit escaping of % and _ characters"
  found_in: [go-store]                       # repos where first discovered
  example_bad: |
    db.Where("name LIKE ?", "%"+input+"%")
  example_good: |
    db.Where("name LIKE ?", EscapeLike(input))
  first_seen: "2026-03-09"
  detection: regex        # future: ast, semantic
  auto_fixable: false     # future: true when we add codemods
```

### Rule Struct (Go)

```go
type Rule struct {
    ID             string   `yaml:"id"`
    Title          string   `yaml:"title"`
    Severity       string   `yaml:"severity"`
    Languages      []string `yaml:"languages"`
    Tags           []string `yaml:"tags"`
    Pattern        string   `yaml:"pattern"`
    ExcludePattern string   `yaml:"exclude_pattern,omitempty"`
    Fix            string   `yaml:"fix"`
    FoundIn        []string `yaml:"found_in,omitempty"`
    ExampleBad     string   `yaml:"example_bad,omitempty"`
    ExampleGood    string   `yaml:"example_good,omitempty"`
    FirstSeen      string   `yaml:"first_seen"`
    Detection      string   `yaml:"detection"`     // regex | ast | semantic
    AutoFixable    bool     `yaml:"auto_fixable"`
}
```

### Finding Struct (Go)

Designed to align with go-ai's `ScanAlert` shape and `~/.core/ai/metrics/` JSONL format:

```go
type Finding struct {
    RuleID   string `json:"rule_id"`
    Title    string `json:"title"`
    Severity string `json:"severity"`
    File     string `json:"file"`
    Line     int    `json:"line"`
    Match    string `json:"match"`        // matched text
    Fix      string `json:"fix"`
    Repo     string `json:"repo,omitempty"`
}
```

### CLI Interface

```bash
# Check current directory against all catalogs for detected languages
core-lint check ./...

# Check specific languages/catalogs
core-lint check --lang go --catalog go-security ./pkg/...

# Output as JSON (for piping to other tools)
core-lint check --format json ./...

# List available rules
core-lint catalog list
core-lint catalog list --lang go --severity high

# Show a specific rule with examples
core-lint catalog show go-sec-001
```

## Layer 2: `go-ai` Lint MCP Subsystem

New subsystem registered alongside files/rag/ml/brain:

```go
type LintSubsystem struct {
    catalog *lint.Catalog
    root    string  // workspace root for scanning
}

func (s *LintSubsystem) Name() string { return "lint" }

func (s *LintSubsystem) RegisterTools(server *mcp.Server) {
    // lint_check  - run rules against workspace files
    // lint_catalog - list/search available rules
    // lint_report - get findings summary for a path
}
```

### MCP Tools

| Tool | Input | Output | Group |
|------|-------|--------|-------|
| `lint_check` | `{path: string, lang?: string, severity?: string}` | `{findings: []Finding}` | lint |
| `lint_catalog` | `{lang?: string, tags?: []string, severity?: string}` | `{rules: []Rule}` | lint |
| `lint_report` | `{path: string, format?: "summary" or "detailed"}` | `{summary: ReportSummary}` | lint |

This means any MCP-connected agent (Claude, LEM, Codex) can call `lint_check` to scan code against the catalog.

## Layer 3: `core/agent` Polish Skill

Claude Code plugin at `core/agent/claude/polish/`:

```
core/agent/claude/polish/
├── plugin.json
├── hooks.json              # optional: PostToolUse after git commit
├── commands/
│   └── polish.md           # /polish slash command
└── scripts/
    └── run-lint.sh         # shells out to core-lint
```

### `/polish` Command Flow

1. Run `core-lint check ./...` for fast deterministic findings
2. Report findings to user
3. Optionally run AI reviewers (Gemini CLI, Codex) for deeper analysis
4. Deduplicate AI findings against catalog (already-known patterns)
5. Propose new patterns as catalog additions (PR to core/lint)

### Subagent Configuration (`.core/agents/`)

Repos can configure polish behaviour:

```yaml
# any-repo/.core/agents/polish.yaml
languages: [go]
catalogs: [go-security, go-modernise, go-correctness]
reviewers: [gemini]          # which AI tools to invoke
exclude: [vendor/, testdata/, *_test.go]
severity_threshold: medium   # only report medium+ findings
```

## Findings to LEM Pipeline

```
core-lint check -> findings.json
    |
    v
~/.core/ai/metrics/YYYY-MM-DD.jsonl    (audit trail)
    |
    v
LEM training data:
  - Rule examples (bad/good pairs) -> supervised training signal
  - Finding frequency -> pattern importance weighting
  - Rule descriptions -> natural language understanding of "why"
    |
    v
LEM tool: "does this code comply with CoreGo standards?"
  -> queries lint_catalog via MCP
  -> applies learned pattern recognition
  -> reports violations with rule IDs and fixes
```

## Initial Catalog Seed

From the March 2026 ecosystem sweep:

| ID | Title | Severity | Language | Found In |
|----|-------|----------|----------|----------|
| go-sec-001 | SQL wildcard injection | high | go | go-store |
| go-sec-002 | Path traversal in cache keys | high | go | go-cache |
| go-sec-003 | XSS in HTML output | high | go | go-html |
| go-sec-004 | Non-constant-time auth comparison | high | go | go-crypt |
| go-sec-005 | Log injection via unescaped input | medium | go | go-log |
| go-sec-006 | Key material in log output | high | go | go-log |
| go-cor-001 | Goroutine leak (no WaitGroup) | high | go | core/go |
| go-cor-002 | Shutdown deadlock (wg.Wait no timeout) | high | go | core/go |
| go-cor-003 | Silent error swallowing | medium | go | go-process, go-ratelimit |
| go-cor-004 | Panic in library code | medium | go | go-i18n |
| go-cor-005 | Delete without path validation | high | go | go-io |
| go-mod-001 | Manual slice clone (append nil pattern) | low | go | core/go |
| go-mod-002 | Manual sort instead of slices.Sorted | low | go | core/go |
| go-mod-003 | Manual reverse loop instead of slices.Backward | low | go | core/go |
| go-mod-004 | sync.WaitGroup Add+Done instead of Go() | low | go | core/go |
| go-mod-005 | Manual map key collection instead of maps.Keys | low | go | core/go |
| go-cor-006 | Missing error return from API calls | medium | go | go-forge, go-git |
| go-cor-007 | Signal handler uses wrong type | medium | go | go-process |

## Dependencies

```
core/lint (standalone, zero core deps)
    ^
    |
go-ai/mcp/lint/ (imports core/lint for catalog + matcher)
    ^
    |
core/agent/claude/polish/ (shells out to core-lint CLI)
```

`core/lint` has no dependency on `core/go` or any other framework module. It is a standalone library + CLI, like `core/go-io`.

## Future Extensions (Not Built Now)

- **AST-based detection** (layer 2): Parse Go/PHP AST, match structural patterns
- **Semantic detection** (layer 3): LEM judges code against rule descriptions
- **Auto-fix codemods**: `core-lint fix` applies known fixes automatically
- **CI integration**: GitHub Actions workflow runs `core-lint check` on PRs
- **CodeRabbit integration**: Import CodeRabbit findings as catalog entries
- **Cross-repo dashboard**: Aggregate findings across all repos in workspace
