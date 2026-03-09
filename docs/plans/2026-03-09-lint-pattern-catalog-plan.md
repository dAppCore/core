# Lint Pattern Catalog Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build `core/lint` — a standalone Go library + CLI that loads YAML pattern catalogs and runs regex-based code checks, seeded with 18 patterns from the March 2026 ecosystem sweep.

**Architecture:** Standalone Go module (`forge.lthn.ai/core/lint`) with zero framework deps. YAML catalog files define rules (id, severity, regex pattern, fix). `pkg/lint` loads catalogs and matches patterns against files. `cmd/core-lint` is a Cobra CLI. Uses `cli.Main()` + `cli.WithCommands()` from `core/cli`.

**Tech Stack:** Go 1.26, `gopkg.in/yaml.v3` (YAML parsing), `forge.lthn.ai/core/cli` (CLI framework), `github.com/stretchr/testify` (testing), `embed` (catalog embedding).

---

### Task 1: Create repo and Go module

**Files:**
- Create: `/Users/snider/Code/core/lint/go.mod`
- Create: `/Users/snider/Code/core/lint/.core/build.yaml`
- Create: `/Users/snider/Code/core/lint/CLAUDE.md`

**Step 1: Create repo on forge**

```bash
ssh -p 2223 git@forge.lthn.ai
```

If SSH repo creation isn't available, create via Forgejo API:
```bash
curl -X POST "https://forge.lthn.ai/api/v1/orgs/core/repos" \
  -H "Authorization: token $FORGE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"lint","description":"Pattern catalog & regex matcher for code quality","auto_init":true,"default_branch":"main"}'
```

Or manually create on forge.lthn.ai web UI under the `core` org.

**Step 2: Clone and initialise Go module**

```bash
cd ~/Code/core
git clone ssh://git@forge.lthn.ai:2223/core/lint.git
cd lint
go mod init forge.lthn.ai/core/lint
```

Set Go version in go.mod:
```
module forge.lthn.ai/core/lint

go 1.26.0
```

**Step 3: Create `.core/build.yaml`**

```yaml
version: 1

project:
  name: core-lint
  description: Pattern catalog and regex code checker
  main: ./cmd/core-lint
  binary: core-lint

build:
  cgo: false
  flags:
    - -trimpath
  ldflags:
    - -s
    - -w

targets:
  - os: linux
    arch: amd64
  - os: linux
    arch: arm64
  - os: darwin
    arch: arm64
  - os: windows
    arch: amd64
```

**Step 4: Create `CLAUDE.md`**

```markdown
# CLAUDE.md

## Project Overview

`core/lint` is a standalone pattern catalog and regex-based code checker. It loads YAML rule definitions and matches them against source files. Zero framework dependencies.

## Build & Development

```bash
core go test
core go qa
core build          # produces ./bin/core-lint
```

## Architecture

- `catalog/` — YAML rule files (embedded at compile time)
- `pkg/lint/` — Library: Rule, Catalog, Matcher, Report types
- `cmd/core-lint/` — CLI binary using `cli.Main()`

## Rule Schema

Each YAML file contains an array of rules with: id, title, severity, languages, tags, pattern (regex), exclude_pattern, fix, example_bad, example_good, detection type.

## Coding Standards

- UK English
- `declare(strict_types=1)` equivalent: all functions have typed params/returns
- Tests use testify
- License: EUPL-1.2
```

**Step 5: Add to go.work**

Add `./core/lint` to `~/Code/go.work` under the Core framework section.

**Step 6: Commit**

```bash
git add go.mod .core/ CLAUDE.md
git commit -m "feat: initialise core/lint module"
```

---

### Task 2: Rule struct and YAML parsing

**Files:**
- Create: `/Users/snider/Code/core/lint/pkg/lint/rule.go`
- Create: `/Users/snider/Code/core/lint/pkg/lint/rule_test.go`

**Step 1: Write the failing test**

```go
package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRules(t *testing.T) {
	yaml := `
- id: test-001
  title: "Test rule"
  severity: high
  languages: [go]
  tags: [security]
  pattern: 'fmt\.Println'
  fix: "Use structured logging"
  detection: regex
`
	rules, err := ParseRules([]byte(yaml))
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, "test-001", rules[0].ID)
	assert.Equal(t, "high", rules[0].Severity)
	assert.Equal(t, []string{"go"}, rules[0].Languages)
	assert.Equal(t, `fmt\.Println`, rules[0].Pattern)
}

func TestParseRules_Invalid(t *testing.T) {
	_, err := ParseRules([]byte("not: valid: yaml: ["))
	assert.Error(t, err)
}

func TestRule_Validate(t *testing.T) {
	good := Rule{ID: "x-001", Title: "T", Severity: "high", Languages: []string{"go"}, Pattern: "foo", Detection: "regex"}
	assert.NoError(t, good.Validate())

	bad := Rule{} // missing required fields
	assert.Error(t, bad.Validate())
}

func TestRule_Validate_BadRegex(t *testing.T) {
	r := Rule{ID: "x-001", Title: "T", Severity: "high", Languages: []string{"go"}, Pattern: "[invalid", Detection: "regex"}
	assert.Error(t, r.Validate())
}
```

**Step 2: Run test to verify it fails**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v`
Expected: FAIL — `ParseRules` and `Rule` not defined

**Step 3: Write minimal implementation**

```go
package lint

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Rule defines a single lint pattern check.
type Rule struct {
	ID             string   `yaml:"id"              json:"id"`
	Title          string   `yaml:"title"           json:"title"`
	Severity       string   `yaml:"severity"        json:"severity"`
	Languages      []string `yaml:"languages"       json:"languages"`
	Tags           []string `yaml:"tags"            json:"tags"`
	Pattern        string   `yaml:"pattern"         json:"pattern"`
	ExcludePattern string   `yaml:"exclude_pattern" json:"exclude_pattern,omitempty"`
	Fix            string   `yaml:"fix"             json:"fix"`
	FoundIn        []string `yaml:"found_in"        json:"found_in,omitempty"`
	ExampleBad     string   `yaml:"example_bad"     json:"example_bad,omitempty"`
	ExampleGood    string   `yaml:"example_good"    json:"example_good,omitempty"`
	FirstSeen      string   `yaml:"first_seen"      json:"first_seen,omitempty"`
	Detection      string   `yaml:"detection"       json:"detection"`
	AutoFixable    bool     `yaml:"auto_fixable"    json:"auto_fixable"`
}

// Validate checks that a Rule has all required fields and a compilable regex pattern.
func (r *Rule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("rule missing id")
	}
	if r.Title == "" {
		return fmt.Errorf("rule %s: missing title", r.ID)
	}
	if r.Severity == "" {
		return fmt.Errorf("rule %s: missing severity", r.ID)
	}
	if len(r.Languages) == 0 {
		return fmt.Errorf("rule %s: missing languages", r.ID)
	}
	if r.Pattern == "" {
		return fmt.Errorf("rule %s: missing pattern", r.ID)
	}
	if r.Detection == "regex" {
		if _, err := regexp.Compile(r.Pattern); err != nil {
			return fmt.Errorf("rule %s: invalid regex: %w", r.ID, err)
		}
	}
	return nil
}

// ParseRules parses YAML bytes into a slice of Rules.
func ParseRules(data []byte) ([]Rule, error) {
	var rules []Rule
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	return rules, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v`
Expected: PASS (4 tests)

**Step 5: Add yaml dependency**

```bash
cd ~/Code/core/lint && go get gopkg.in/yaml.v3 && go get github.com/stretchr/testify
```

**Step 6: Commit**

```bash
git add pkg/lint/rule.go pkg/lint/rule_test.go go.mod go.sum
git commit -m "feat: add Rule struct with YAML parsing and validation"
```

---

### Task 3: Catalog loader with embed support

**Files:**
- Create: `/Users/snider/Code/core/lint/pkg/lint/catalog.go`
- Create: `/Users/snider/Code/core/lint/pkg/lint/catalog_test.go`
- Create: `/Users/snider/Code/core/lint/catalog/go-security.yaml` (minimal test file)

**Step 1: Create a minimal test catalog file**

Create `/Users/snider/Code/core/lint/catalog/go-security.yaml`:
```yaml
- id: go-sec-001
  title: "SQL wildcard injection in LIKE clauses"
  severity: high
  languages: [go]
  tags: [security, injection]
  pattern: 'LIKE\s+\?\s*,\s*["%].*\+'
  fix: "Use parameterised LIKE with EscapeLike()"
  found_in: [go-store]
  first_seen: "2026-03-09"
  detection: regex
```

**Step 2: Write the failing test**

```go
package lint

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalog_LoadDir(t *testing.T) {
	// Find the catalog/ dir relative to the module root
	dir := filepath.Join("..", "..", "catalog")
	cat, err := LoadDir(dir)
	require.NoError(t, err)
	assert.Greater(t, len(cat.Rules), 0)
	assert.Equal(t, "go-sec-001", cat.Rules[0].ID)
}

func TestCatalog_LoadDir_NotExist(t *testing.T) {
	_, err := LoadDir("/nonexistent")
	assert.Error(t, err)
}

func TestCatalog_Filter_Language(t *testing.T) {
	cat := &Catalog{Rules: []Rule{
		{ID: "go-001", Languages: []string{"go"}, Severity: "high"},
		{ID: "php-001", Languages: []string{"php"}, Severity: "high"},
	}}
	filtered := cat.ForLanguage("go")
	assert.Len(t, filtered, 1)
	assert.Equal(t, "go-001", filtered[0].ID)
}

func TestCatalog_Filter_Severity(t *testing.T) {
	cat := &Catalog{Rules: []Rule{
		{ID: "a", Severity: "high"},
		{ID: "b", Severity: "low"},
		{ID: "c", Severity: "medium"},
	}}
	filtered := cat.AtSeverity("medium")
	assert.Len(t, filtered, 2) // high + medium
}

func TestCatalog_LoadFS(t *testing.T) {
	// Write temp yaml
	dir := t.TempDir()
	data := []byte(`- id: fs-001
  title: "FS test"
  severity: low
  languages: [go]
  tags: []
  pattern: 'test'
  fix: "fix"
  detection: regex
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.yaml"), data, 0644))

	cat, err := LoadDir(dir)
	require.NoError(t, err)
	assert.Len(t, cat.Rules, 1)
}
```

**Step 3: Write minimal implementation**

```go
package lint

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Catalog holds a collection of lint rules loaded from YAML files.
type Catalog struct {
	Rules []Rule
}

// severityOrder maps severity names to numeric priority (higher = more severe).
var severityOrder = map[string]int{
	"critical": 5,
	"high":     4,
	"medium":   3,
	"low":      2,
	"info":     1,
}

// LoadDir loads all .yaml files from a directory path into a Catalog.
func LoadDir(dir string) (*Catalog, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("load catalog dir: %w", err)
	}

	cat := &Catalog{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		rules, err := ParseRules(data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		cat.Rules = append(cat.Rules, rules...)
	}
	return cat, nil
}

// LoadFS loads all .yaml files from an embed.FS into a Catalog.
func LoadFS(fsys embed.FS, dir string) (*Catalog, error) {
	cat := &Catalog{}
	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		rules, err := ParseRules(data)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		cat.Rules = append(cat.Rules, rules...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cat, nil
}

// ForLanguage returns rules that apply to the given language.
func (c *Catalog) ForLanguage(lang string) []Rule {
	var out []Rule
	for _, r := range c.Rules {
		if slices.Contains(r.Languages, lang) {
			out = append(out, r)
		}
	}
	return out
}

// AtSeverity returns rules at or above the given severity threshold.
func (c *Catalog) AtSeverity(threshold string) []Rule {
	minLevel := severityOrder[threshold]
	var out []Rule
	for _, r := range c.Rules {
		if severityOrder[r.Severity] >= minLevel {
			out = append(out, r)
		}
	}
	return out
}

// ByID returns a rule by its ID, or nil if not found.
func (c *Catalog) ByID(id string) *Rule {
	for i := range c.Rules {
		if c.Rules[i].ID == id {
			return &c.Rules[i]
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v`
Expected: PASS (all tests)

**Step 5: Commit**

```bash
git add pkg/lint/catalog.go pkg/lint/catalog_test.go catalog/go-security.yaml
git commit -m "feat: add Catalog loader with dir/embed/filter support"
```

---

### Task 4: Regex matcher

**Files:**
- Create: `/Users/snider/Code/core/lint/pkg/lint/matcher.go`
- Create: `/Users/snider/Code/core/lint/pkg/lint/matcher_test.go`

**Step 1: Write the failing test**

```go
package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatcher_Match(t *testing.T) {
	rules := []Rule{
		{
			ID:        "test-001",
			Title:     "fmt.Println usage",
			Severity:  "low",
			Languages: []string{"go"},
			Pattern:   `fmt\.Println`,
			Fix:       "Use structured logging",
			Detection: "regex",
		},
	}
	m, err := NewMatcher(rules)
	require.NoError(t, err)

	content := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	findings := m.Match("main.go", []byte(content))
	require.Len(t, findings, 1)
	assert.Equal(t, "test-001", findings[0].RuleID)
	assert.Equal(t, "main.go", findings[0].File)
	assert.Equal(t, 6, findings[0].Line)
	assert.Contains(t, findings[0].Match, "fmt.Println")
}

func TestMatcher_ExcludePattern(t *testing.T) {
	rules := []Rule{
		{
			ID:             "test-002",
			Title:          "Println with exclude",
			Severity:       "low",
			Languages:      []string{"go"},
			Pattern:        `fmt\.Println`,
			ExcludePattern: `// lint:ignore`,
			Fix:            "Use logging",
			Detection:      "regex",
		},
	}
	m, err := NewMatcher(rules)
	require.NoError(t, err)

	content := `package main
func a() { fmt.Println("bad") }
func b() { fmt.Println("ok") // lint:ignore }
`
	findings := m.Match("main.go", []byte(content))
	// Line 2 matches, line 3 is excluded
	assert.Len(t, findings, 1)
	assert.Equal(t, 2, findings[0].Line)
}

func TestMatcher_NoMatch(t *testing.T) {
	rules := []Rule{
		{ID: "test-003", Title: "T", Severity: "low", Languages: []string{"go"}, Pattern: `NEVER_MATCH_THIS`, Detection: "regex"},
	}
	m, err := NewMatcher(rules)
	require.NoError(t, err)

	findings := m.Match("main.go", []byte("package main\n"))
	assert.Empty(t, findings)
}

func TestMatcher_InvalidRegex(t *testing.T) {
	rules := []Rule{
		{ID: "bad", Title: "T", Severity: "low", Languages: []string{"go"}, Pattern: `[invalid`, Detection: "regex"},
	}
	_, err := NewMatcher(rules)
	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v -run TestMatcher`
Expected: FAIL — `NewMatcher` not defined

**Step 3: Write minimal implementation**

```go
package lint

import (
	"fmt"
	"regexp"
	"strings"
)

// Finding represents a single match of a rule against source code.
type Finding struct {
	RuleID   string `json:"rule_id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Match    string `json:"match"`
	Fix      string `json:"fix"`
	Repo     string `json:"repo,omitempty"`
}

// compiledRule is a rule with its regex pre-compiled.
type compiledRule struct {
	rule    Rule
	pattern *regexp.Regexp
	exclude *regexp.Regexp
}

// Matcher runs compiled rules against file contents.
type Matcher struct {
	rules []compiledRule
}

// NewMatcher compiles all rule patterns and returns a Matcher.
func NewMatcher(rules []Rule) (*Matcher, error) {
	compiled := make([]compiledRule, 0, len(rules))
	for _, r := range rules {
		if r.Detection != "regex" {
			continue // skip non-regex rules
		}
		p, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rule %s: invalid pattern: %w", r.ID, err)
		}
		cr := compiledRule{rule: r, pattern: p}
		if r.ExcludePattern != "" {
			ex, err := regexp.Compile(r.ExcludePattern)
			if err != nil {
				return nil, fmt.Errorf("rule %s: invalid exclude_pattern: %w", r.ID, err)
			}
			cr.exclude = ex
		}
		compiled = append(compiled, cr)
	}
	return &Matcher{rules: compiled}, nil
}

// Match checks file contents against all rules and returns findings.
func (m *Matcher) Match(filename string, content []byte) []Finding {
	lines := strings.Split(string(content), "\n")
	var findings []Finding

	for _, cr := range m.rules {
		for i, line := range lines {
			if !cr.pattern.MatchString(line) {
				continue
			}
			if cr.exclude != nil && cr.exclude.MatchString(line) {
				continue
			}
			findings = append(findings, Finding{
				RuleID:   cr.rule.ID,
				Title:    cr.rule.Title,
				Severity: cr.rule.Severity,
				File:     filename,
				Line:     i + 1,
				Match:    strings.TrimSpace(line),
				Fix:      cr.rule.Fix,
			})
		}
	}
	return findings
}
```

**Step 4: Run test to verify it passes**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v -run TestMatcher`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add pkg/lint/matcher.go pkg/lint/matcher_test.go
git commit -m "feat: add regex Matcher with exclude pattern support"
```

---

### Task 5: Report output (JSON, text, JSONL)

**Files:**
- Create: `/Users/snider/Code/core/lint/pkg/lint/report.go`
- Create: `/Users/snider/Code/core/lint/pkg/lint/report_test.go`

**Step 1: Write the failing test**

```go
package lint

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReport_JSON(t *testing.T) {
	findings := []Finding{
		{RuleID: "x-001", Title: "Test", Severity: "high", File: "a.go", Line: 10, Match: "bad code", Fix: "fix it"},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteJSON(&buf, findings))

	var parsed []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Len(t, parsed, 1)
	assert.Equal(t, "x-001", parsed[0].RuleID)
}

func TestReport_JSONL(t *testing.T) {
	findings := []Finding{
		{RuleID: "a-001", File: "a.go", Line: 1},
		{RuleID: "b-001", File: "b.go", Line: 2},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteJSONL(&buf, findings))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)
}

func TestReport_Text(t *testing.T) {
	findings := []Finding{
		{RuleID: "x-001", Title: "Test rule", Severity: "high", File: "main.go", Line: 42, Match: "bad()", Fix: "use good()"},
	}
	var buf bytes.Buffer
	WriteText(&buf, findings)

	out := buf.String()
	assert.Contains(t, out, "main.go:42")
	assert.Contains(t, out, "x-001")
	assert.Contains(t, out, "high")
}

func TestReport_Summary(t *testing.T) {
	findings := []Finding{
		{Severity: "high"},
		{Severity: "high"},
		{Severity: "low"},
	}
	s := Summarise(findings)
	assert.Equal(t, 3, s.Total)
	assert.Equal(t, 2, s.BySeverity["high"])
	assert.Equal(t, 1, s.BySeverity["low"])
}
```

**Step 2: Run test to verify it fails**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v -run TestReport`
Expected: FAIL — functions not defined

**Step 3: Write minimal implementation**

```go
package lint

import (
	"encoding/json"
	"fmt"
	"io"
)

// Summary holds aggregate stats about findings.
type Summary struct {
	Total      int            `json:"total"`
	BySeverity map[string]int `json:"by_severity"`
}

// Summarise creates a Summary from a list of findings.
func Summarise(findings []Finding) Summary {
	s := Summary{
		Total:      len(findings),
		BySeverity: make(map[string]int),
	}
	for _, f := range findings {
		s.BySeverity[f.Severity]++
	}
	return s
}

// WriteJSON writes findings as a JSON array.
func WriteJSON(w io.Writer, findings []Finding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(findings)
}

// WriteJSONL writes findings as newline-delimited JSON (one object per line).
// Compatible with ~/.core/ai/metrics/ format.
func WriteJSONL(w io.Writer, findings []Finding) error {
	enc := json.NewEncoder(w)
	for _, f := range findings {
		if err := enc.Encode(f); err != nil {
			return err
		}
	}
	return nil
}

// WriteText writes findings as human-readable text.
func WriteText(w io.Writer, findings []Finding) {
	for _, f := range findings {
		fmt.Fprintf(w, "%s:%d  [%s] %s (%s)\n", f.File, f.Line, f.Severity, f.Title, f.RuleID)
		if f.Fix != "" {
			fmt.Fprintf(w, "  fix: %s\n", f.Fix)
		}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v -run TestReport`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add pkg/lint/report.go pkg/lint/report_test.go
git commit -m "feat: add report output (JSON, JSONL, text, summary)"
```

---

### Task 6: Scanner (walk files + match)

**Files:**
- Create: `/Users/snider/Code/core/lint/pkg/lint/scanner.go`
- Create: `/Users/snider/Code/core/lint/pkg/lint/scanner_test.go`

**Step 1: Write the failing test**

```go
package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner_ScanDir(t *testing.T) {
	// Set up temp dir with a .go file containing a known pattern
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`), 0644))

	rules := []Rule{
		{ID: "test-001", Title: "Println", Severity: "low", Languages: []string{"go"}, Pattern: `fmt\.Println`, Fix: "log", Detection: "regex"},
	}

	s, err := NewScanner(rules)
	require.NoError(t, err)

	findings, err := s.ScanDir(dir)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "test-001", findings[0].RuleID)
}

func TestScanner_ScanDir_ExcludesVendor(t *testing.T) {
	dir := t.TempDir()
	vendor := filepath.Join(dir, "vendor")
	require.NoError(t, os.MkdirAll(vendor, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(vendor, "lib.go"), []byte("package lib\nfunc x() { fmt.Println() }\n"), 0644))

	rules := []Rule{
		{ID: "test-001", Title: "Println", Severity: "low", Languages: []string{"go"}, Pattern: `fmt\.Println`, Fix: "log", Detection: "regex"},
	}

	s, err := NewScanner(rules)
	require.NoError(t, err)

	findings, err := s.ScanDir(dir)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestScanner_LanguageDetection(t *testing.T) {
	assert.Equal(t, "go", DetectLanguage("main.go"))
	assert.Equal(t, "php", DetectLanguage("app.php"))
	assert.Equal(t, "ts", DetectLanguage("index.ts"))
	assert.Equal(t, "ts", DetectLanguage("index.tsx"))
	assert.Equal(t, "cpp", DetectLanguage("engine.cpp"))
	assert.Equal(t, "cpp", DetectLanguage("engine.cc"))
	assert.Equal(t, "", DetectLanguage("README.md"))
}
```

**Step 2: Run test to verify it fails**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v -run TestScanner`
Expected: FAIL — `NewScanner` not defined

**Step 3: Write minimal implementation**

```go
package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// defaultExcludes are directories skipped during scanning.
var defaultExcludes = []string{"vendor", "node_modules", ".git", "testdata", ".core"}

// extToLang maps file extensions to language identifiers.
var extToLang = map[string]string{
	".go":   "go",
	".php":  "php",
	".ts":   "ts",
	".tsx":  "ts",
	".js":   "js",
	".jsx":  "js",
	".cpp":  "cpp",
	".cc":   "cpp",
	".cxx":  "cpp",
	".c":    "cpp",
	".h":    "cpp",
	".hpp":  "cpp",
}

// DetectLanguage returns the language identifier for a filename, or "" if unknown.
func DetectLanguage(filename string) string {
	ext := filepath.Ext(filename)
	return extToLang[ext]
}

// Scanner walks directories and matches files against rules.
type Scanner struct {
	matcher  *Matcher
	rules    []Rule
	excludes []string
}

// NewScanner creates a Scanner from a set of rules.
func NewScanner(rules []Rule) (*Scanner, error) {
	m, err := NewMatcher(rules)
	if err != nil {
		return nil, err
	}
	return &Scanner{
		matcher:  m,
		rules:    rules,
		excludes: defaultExcludes,
	}, nil
}

// ScanDir walks a directory tree and returns all findings.
func (s *Scanner) ScanDir(root string) ([]Finding, error) {
	var all []Finding

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if d.IsDir() {
			for _, ex := range s.excludes {
				if d.Name() == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Only scan files with known language extensions
		lang := DetectLanguage(path)
		if lang == "" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// Make path relative to root for cleaner output
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}

		findings := s.matcher.Match(rel, content)
		all = append(all, findings...)
		return nil
	})

	return all, err
}

// ScanFile scans a single file and returns findings.
func (s *Scanner) ScanFile(path string) ([]Finding, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return s.matcher.Match(path, content), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v -run TestScanner`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add pkg/lint/scanner.go pkg/lint/scanner_test.go
git commit -m "feat: add Scanner with directory walking and language detection"
```

---

### Task 7: Seed the catalog YAML files

**Files:**
- Create: `/Users/snider/Code/core/lint/catalog/go-security.yaml` (expand from task 3)
- Create: `/Users/snider/Code/core/lint/catalog/go-correctness.yaml`
- Create: `/Users/snider/Code/core/lint/catalog/go-modernise.yaml`

**Step 1: Write `catalog/go-security.yaml`**

```yaml
- id: go-sec-001
  title: "SQL wildcard injection in LIKE clauses"
  severity: high
  languages: [go]
  tags: [security, injection, owasp-a03]
  pattern: 'LIKE\s+\?.*["%`]\s*\%.*\+'
  exclude_pattern: 'EscapeLike'
  fix: "Use parameterised LIKE with explicit escaping of % and _ characters"
  found_in: [go-store]
  example_bad: |
    db.Where("name LIKE ?", "%"+input+"%")
  example_good: |
    db.Where("name LIKE ?", EscapeLike(input))
  first_seen: "2026-03-09"
  detection: regex

- id: go-sec-002
  title: "Path traversal in file/cache key operations"
  severity: high
  languages: [go]
  tags: [security, path-traversal, owasp-a01]
  pattern: 'filepath\.Join\(.*,\s*\w+\)'
  exclude_pattern: 'filepath\.Clean|securejoin|ValidatePath'
  fix: "Validate path components do not contain .. before joining"
  found_in: [go-cache]
  example_bad: |
    path := filepath.Join(cacheDir, userInput)
  example_good: |
    if strings.Contains(key, "..") { return ErrInvalidKey }
    path := filepath.Join(cacheDir, key)
  first_seen: "2026-03-09"
  detection: regex

- id: go-sec-003
  title: "XSS via unescaped HTML output"
  severity: high
  languages: [go]
  tags: [security, xss, owasp-a03]
  pattern: 'fmt\.Sprintf\(.*<.*>.*%s'
  exclude_pattern: 'html\.EscapeString|template\.HTMLEscapeString'
  fix: "Use html.EscapeString() for user-supplied values in HTML output"
  found_in: [go-html]
  example_bad: |
    out := fmt.Sprintf("<div>%s</div>", userInput)
  example_good: |
    out := fmt.Sprintf("<div>%s</div>", html.EscapeString(userInput))
  first_seen: "2026-03-09"
  detection: regex

- id: go-sec-004
  title: "Non-constant-time comparison for authentication"
  severity: high
  languages: [go]
  tags: [security, timing-attack, owasp-a02]
  pattern: '==\s*\w*(token|key|secret|password|hash|digest|hmac|mac|sig)'
  exclude_pattern: 'subtle\.ConstantTimeCompare|hmac\.Equal'
  fix: "Use crypto/subtle.ConstantTimeCompare for security-sensitive comparisons"
  found_in: [go-crypt]
  example_bad: |
    if providedToken == storedToken {
  example_good: |
    if subtle.ConstantTimeCompare([]byte(provided), []byte(stored)) == 1 {
  first_seen: "2026-03-09"
  detection: regex

- id: go-sec-005
  title: "Log injection via unescaped newlines"
  severity: medium
  languages: [go]
  tags: [security, injection, logging]
  pattern: 'log\.\w+\(.*\+.*\)'
  exclude_pattern: 'strings\.ReplaceAll.*\\n|slog\.'
  fix: "Use structured logging (slog) or sanitise newlines from user input"
  found_in: [go-log]
  example_bad: |
    log.Printf("user login: " + username)
  example_good: |
    slog.Info("user login", "username", username)
  first_seen: "2026-03-09"
  detection: regex

- id: go-sec-006
  title: "Sensitive key material in log output"
  severity: high
  languages: [go]
  tags: [security, secrets, logging]
  pattern: 'log\.\w+\(.*(?i)(password|secret|token|apikey|private.?key|credential)'
  exclude_pattern: 'REDACTED|\*\*\*|redact'
  fix: "Redact sensitive fields before logging"
  found_in: [go-log]
  example_bad: |
    log.Printf("config: token=%s", cfg.Token)
  example_good: |
    log.Printf("config: token=%s", redact(cfg.Token))
  first_seen: "2026-03-09"
  detection: regex
```

**Step 2: Write `catalog/go-correctness.yaml`**

```yaml
- id: go-cor-001
  title: "Goroutine without WaitGroup or context"
  severity: high
  languages: [go]
  tags: [correctness, goroutine-leak]
  pattern: 'go\s+func\s*\('
  exclude_pattern: 'wg\.|\.Go\(|context\.|done\s*<-|select\s*\{'
  fix: "Use sync.WaitGroup.Go() or ensure goroutine has a shutdown signal"
  found_in: [core/go]
  example_bad: |
    go func() { doWork() }()
  example_good: |
    wg.Go(func() { doWork() })
  first_seen: "2026-03-09"
  detection: regex

- id: go-cor-002
  title: "WaitGroup.Wait without context/timeout"
  severity: high
  languages: [go]
  tags: [correctness, deadlock]
  pattern: '\.Wait\(\)'
  exclude_pattern: 'select\s*\{|ctx\.Done|context\.With|time\.After'
  fix: "Wrap wg.Wait() in a select with context.Done() or timeout"
  found_in: [core/go]
  example_bad: |
    wg.Wait() // blocks forever if goroutine hangs
  example_good: |
    done := make(chan struct{})
    go func() { wg.Wait(); close(done) }()
    select {
    case <-done:
    case <-ctx.Done():
    }
  first_seen: "2026-03-09"
  detection: regex

- id: go-cor-003
  title: "Silent error swallowing"
  severity: medium
  languages: [go]
  tags: [correctness, error-handling]
  pattern: '^\s*_\s*=\s*\w+\.\w+\('
  exclude_pattern: 'defer|Close\(|Flush\('
  fix: "Handle or propagate errors instead of discarding with _"
  found_in: [go-process, go-ratelimit]
  example_bad: |
    _ = db.Save(record)
  example_good: |
    if err := db.Save(record); err != nil {
        return fmt.Errorf("save record: %w", err)
    }
  first_seen: "2026-03-09"
  detection: regex

- id: go-cor-004
  title: "Panic in library code"
  severity: medium
  languages: [go]
  tags: [correctness, panic]
  pattern: '\bpanic\('
  exclude_pattern: '_test\.go|// unreachable|Must\w+\('
  fix: "Return errors instead of panicking in library code"
  found_in: [go-i18n]
  example_bad: |
    func Parse(s string) *Node { panic("not implemented") }
  example_good: |
    func Parse(s string) (*Node, error) { return nil, fmt.Errorf("not implemented") }
  first_seen: "2026-03-09"
  detection: regex

- id: go-cor-005
  title: "File deletion without path validation"
  severity: high
  languages: [go]
  tags: [correctness, safety]
  pattern: 'os\.Remove(All)?\('
  exclude_pattern: 'filepath\.Clean|ValidatePath|strings\.Contains.*\.\.'
  fix: "Validate path does not escape base directory before deletion"
  found_in: [go-io]
  example_bad: |
    os.RemoveAll(filepath.Join(base, userInput))
  example_good: |
    clean := filepath.Clean(filepath.Join(base, userInput))
    if !strings.HasPrefix(clean, base) { return ErrPathTraversal }
    os.RemoveAll(clean)
  first_seen: "2026-03-09"
  detection: regex

- id: go-cor-006
  title: "Missing error return from API/network calls"
  severity: medium
  languages: [go]
  tags: [correctness, error-handling]
  pattern: 'resp,\s*_\s*:=.*\.(Get|Post|Do|Send)\('
  fix: "Check and handle HTTP/API errors"
  found_in: [go-forge, go-git]
  example_bad: |
    resp, _ := client.Get(url)
  example_good: |
    resp, err := client.Get(url)
    if err != nil { return fmt.Errorf("api call: %w", err) }
  first_seen: "2026-03-09"
  detection: regex

- id: go-cor-007
  title: "Signal handler uses wrong type"
  severity: medium
  languages: [go]
  tags: [correctness, signals]
  pattern: 'syscall\.Signal\b'
  exclude_pattern: 'os\.Signal'
  fix: "Use os.Signal for portable signal handling"
  found_in: [go-process]
  example_bad: |
    func Handle(sig syscall.Signal) { ... }
  example_good: |
    func Handle(sig os.Signal) { ... }
  first_seen: "2026-03-09"
  detection: regex
```

**Step 3: Write `catalog/go-modernise.yaml`**

```yaml
- id: go-mod-001
  title: "Manual slice clone via append([]T(nil)...)"
  severity: low
  languages: [go]
  tags: [modernise, go126]
  pattern: 'append\(\[\]\w+\(nil\),\s*\w+\.\.\.\)'
  fix: "Use slices.Clone() from Go 1.21+"
  found_in: [core/go]
  example_bad: |
    copy := append([]string(nil), original...)
  example_good: |
    copy := slices.Clone(original)
  first_seen: "2026-03-09"
  detection: regex

- id: go-mod-002
  title: "Manual sort of string/int slices"
  severity: low
  languages: [go]
  tags: [modernise, go126]
  pattern: 'sort\.Strings\(|sort\.Ints\(|sort\.Slice\('
  exclude_pattern: 'sort\.SliceStable'
  fix: "Use slices.Sort() or slices.Sorted(iter) from Go 1.21+"
  found_in: [core/go]
  example_bad: |
    sort.Strings(names)
  example_good: |
    slices.Sort(names)
  first_seen: "2026-03-09"
  detection: regex

- id: go-mod-003
  title: "Manual reverse iteration loop"
  severity: low
  languages: [go]
  tags: [modernise, go126]
  pattern: 'for\s+\w+\s*:=\s*len\(\w+\)\s*-\s*1'
  fix: "Use slices.Backward() from Go 1.23+"
  found_in: [core/go]
  example_bad: |
    for i := len(items) - 1; i >= 0; i-- { use(items[i]) }
  example_good: |
    for _, item := range slices.Backward(items) { use(item) }
  first_seen: "2026-03-09"
  detection: regex

- id: go-mod-004
  title: "WaitGroup Add+Done instead of Go()"
  severity: low
  languages: [go]
  tags: [modernise, go126]
  pattern: 'wg\.Add\(1\)'
  fix: "Use sync.WaitGroup.Go() from Go 1.26"
  found_in: [core/go]
  example_bad: |
    wg.Add(1)
    go func() { defer wg.Done(); work() }()
  example_good: |
    wg.Go(func() { work() })
  first_seen: "2026-03-09"
  detection: regex

- id: go-mod-005
  title: "Manual map key collection"
  severity: low
  languages: [go]
  tags: [modernise, go126]
  pattern: 'for\s+\w+\s*:=\s*range\s+\w+\s*\{\s*\n\s*\w+\s*=\s*append'
  exclude_pattern: 'maps\.Keys'
  fix: "Use maps.Keys() or slices.Sorted(maps.Keys()) from Go 1.23+"
  found_in: [core/go]
  example_bad: |
    var keys []string
    for k := range m { keys = append(keys, k) }
  example_good: |
    keys := slices.Sorted(maps.Keys(m))
  first_seen: "2026-03-09"
  detection: regex
```

**Step 4: Run all tests to verify catalog loads correctly**

Run: `cd ~/Code/core/lint && go test ./pkg/lint/ -v`
Expected: PASS (all tests, including TestCatalog_LoadDir which reads the catalog/ dir)

**Step 5: Commit**

```bash
git add catalog/
git commit -m "feat: seed catalog with 18 patterns from ecosystem sweep"
```

---

### Task 8: CLI binary with `cli.Main()`

**Files:**
- Create: `/Users/snider/Code/core/lint/cmd/core-lint/main.go`
- Create: `/Users/snider/Code/core/lint/lint.go` (embed catalog + public API)

**Step 1: Create the embed entry point**

Create `/Users/snider/Code/core/lint/lint.go`:

```go
package lint

import (
	"embed"

	lintpkg "forge.lthn.ai/core/lint/pkg/lint"
)

//go:embed catalog/*.yaml
var catalogFS embed.FS

// LoadEmbeddedCatalog loads the built-in catalog from embedded YAML files.
func LoadEmbeddedCatalog() (*lintpkg.Catalog, error) {
	return lintpkg.LoadFS(catalogFS, "catalog")
}
```

**Step 2: Create the CLI entry point**

Create `/Users/snider/Code/core/lint/cmd/core-lint/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	lint "forge.lthn.ai/core/lint"
	lintpkg "forge.lthn.ai/core/lint/pkg/lint"
)

func main() {
	cli.Main(
		cli.WithCommands("lint", addLintCommands),
	)
}

func addLintCommands(root *cli.Command) {
	lintCmd := &cli.Command{
		Use:   "lint",
		Short: "Pattern-based code checker",
	}
	root.AddCommand(lintCmd)

	// core-lint lint check [path...]
	lintCmd.AddCommand(cli.NewCommand(
		"check [path...]",
		"Run pattern checks against source files",
		"Scans files for known anti-patterns from the catalog",
		func(cmd *cli.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			lang, _ := cmd.Flags().GetString("lang")
			severity, _ := cmd.Flags().GetString("severity")

			cat, err := lint.LoadEmbeddedCatalog()
			if err != nil {
				return fmt.Errorf("load catalog: %w", err)
			}

			rules := cat.Rules
			if lang != "" {
				rules = cat.ForLanguage(lang)
			}
			if severity != "" {
				filtered := (&lintpkg.Catalog{Rules: rules}).AtSeverity(severity)
				rules = filtered
			}

			scanner, err := lintpkg.NewScanner(rules)
			if err != nil {
				return fmt.Errorf("create scanner: %w", err)
			}

			paths := args
			if len(paths) == 0 {
				paths = []string{"."}
			}

			var allFindings []lintpkg.Finding
			for _, p := range paths {
				findings, err := scanner.ScanDir(p)
				if err != nil {
					return fmt.Errorf("scan %s: %w", p, err)
				}
				allFindings = append(allFindings, findings...)
			}

			switch format {
			case "json":
				return lintpkg.WriteJSON(os.Stdout, allFindings)
			case "jsonl":
				return lintpkg.WriteJSONL(os.Stdout, allFindings)
			default:
				lintpkg.WriteText(os.Stdout, allFindings)
			}

			if len(allFindings) > 0 {
				s := lintpkg.Summarise(allFindings)
				fmt.Fprintf(os.Stderr, "\n%d findings", s.Total)
				for sev, count := range s.BySeverity {
					fmt.Fprintf(os.Stderr, " | %s: %d", sev, count)
				}
				fmt.Fprintln(os.Stderr)
			}
			return nil
		},
	))

	// Add flags to check command
	checkCmd := lintCmd.Commands()[0]
	checkCmd.Flags().StringP("format", "f", "text", "Output format: text, json, jsonl")
	checkCmd.Flags().StringP("lang", "l", "", "Filter by language: go, php, ts, cpp")
	checkCmd.Flags().StringP("severity", "s", "", "Minimum severity: critical, high, medium, low, info")

	// core-lint lint catalog
	catalogCmd := &cli.Command{
		Use:   "catalog",
		Short: "Browse the pattern catalog",
	}
	lintCmd.AddCommand(catalogCmd)

	// core-lint lint catalog list
	catalogCmd.AddCommand(cli.NewCommand(
		"list",
		"List available rules",
		"",
		func(cmd *cli.Command, args []string) error {
			lang, _ := cmd.Flags().GetString("lang")

			cat, err := lint.LoadEmbeddedCatalog()
			if err != nil {
				return err
			}

			rules := cat.Rules
			if lang != "" {
				rules = cat.ForLanguage(lang)
			}

			for _, r := range rules {
				fmt.Printf("%-12s [%s] %s\n", r.ID, r.Severity, r.Title)
			}
			fmt.Fprintf(os.Stderr, "\n%d rules\n", len(rules))
			return nil
		},
	))
	catalogCmd.Commands()[0].Flags().StringP("lang", "l", "", "Filter by language")

	// core-lint lint catalog show <id>
	catalogCmd.AddCommand(cli.NewCommand(
		"show [rule-id]",
		"Show details for a specific rule",
		"",
		func(cmd *cli.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("rule ID required")
			}
			cat, err := lint.LoadEmbeddedCatalog()
			if err != nil {
				return err
			}
			r := cat.ByID(args[0])
			if r == nil {
				return fmt.Errorf("rule %s not found", args[0])
			}
			fmt.Printf("ID:        %s\n", r.ID)
			fmt.Printf("Title:     %s\n", r.Title)
			fmt.Printf("Severity:  %s\n", r.Severity)
			fmt.Printf("Languages: %v\n", r.Languages)
			fmt.Printf("Tags:      %v\n", r.Tags)
			fmt.Printf("Pattern:   %s\n", r.Pattern)
			if r.ExcludePattern != "" {
				fmt.Printf("Exclude:   %s\n", r.ExcludePattern)
			}
			fmt.Printf("Fix:       %s\n", r.Fix)
			if r.ExampleBad != "" {
				fmt.Printf("\nBad:\n%s\n", r.ExampleBad)
			}
			if r.ExampleGood != "" {
				fmt.Printf("Good:\n%s\n", r.ExampleGood)
			}
			return nil
		},
	))
}
```

**Step 3: Add cli dependency**

```bash
cd ~/Code/core/lint
go get forge.lthn.ai/core/cli
go mod tidy
```

**Step 4: Build and smoke test**

```bash
cd ~/Code/core/lint
go build -o ./bin/core-lint ./cmd/core-lint
./bin/core-lint lint catalog list
./bin/core-lint lint catalog show go-sec-001
./bin/core-lint lint check --lang go --format json ~/Code/host-uk/core/pkg/core/
```

Expected: Binary builds, catalog lists 18 rules, show displays rule details, check scans files.

**Step 5: Commit**

```bash
git add lint.go cmd/core-lint/main.go go.mod go.sum
git commit -m "feat: add core-lint CLI with check, catalog list, catalog show"
```

---

### Task 9: Run all tests, push to forge

**Step 1: Run full test suite**

```bash
cd ~/Code/core/lint
go test -race -count=1 ./...
```

Expected: PASS with race detector

**Step 2: Run go vet**

```bash
go vet ./...
```

Expected: No issues

**Step 3: Build binary**

```bash
go build -trimpath -o ./bin/core-lint ./cmd/core-lint
```

**Step 4: Smoke test against a real repo**

```bash
./bin/core-lint lint check --lang go ~/Code/host-uk/core/pkg/core/
./bin/core-lint lint check --lang go --severity high ~/Code/core/go-io/
```

Expected: Any findings are displayed (or no findings if the repos are already clean from our sweep)

**Step 5: Update go.work**

```bash
# Add ./core/lint to ~/Code/go.work if not already there
cd ~/Code && go work sync
```

**Step 6: Push to forge**

```bash
cd ~/Code/core/lint
git push -u origin main
```

**Step 7: Tag initial release**

```bash
git tag v0.1.0
git push origin v0.1.0
```
