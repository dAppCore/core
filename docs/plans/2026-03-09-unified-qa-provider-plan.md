# Unified QA Provider Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `core/lint` the unified QA provider for Go and PHP by extracting PHP QA code from `core/php` into `core/lint`.

**Architecture:** Language sub-packages under `pkg/` — `pkg/detect/` for project detection, `pkg/lint/` for Go (unchanged), `pkg/php/` for PHP analysis. `cmd/qa/` gains PHP subcommands. Uses `go-io` Medium interface for filesystem, `go-process` for QA pipeline orchestration.

**Tech Stack:** Go 1.26, go-io (Medium interface), go-process (subprocess runner), go-i18n, cli

---

### Task 1: Create `pkg/detect/` — Project Type Detection

**Files:**
- Create: `pkg/detect/detect.go`
- Create: `pkg/detect/detect_test.go`

**Step 1: Write the failing test**

```go
package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGoProject_Good(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	assert.True(t, IsGoProject(dir))
}

func TestIsGoProject_Bad(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, IsGoProject(dir))
}

func TestIsPHPProject_Good(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}"), 0644)
	assert.True(t, IsPHPProject(dir))
}

func TestIsPHPProject_Bad(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, IsPHPProject(dir))
}

func TestDetectAll_Good(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}"), 0644)
	types := DetectAll(dir)
	assert.Contains(t, types, Go)
	assert.Contains(t, types, PHP)
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code && GOWORK=go.work go test ./core/lint/pkg/detect/... -v`
Expected: FAIL — functions not defined

**Step 3: Write minimal implementation**

```go
// Package detect identifies project types by examining filesystem markers.
package detect

import "os"

// ProjectType identifies a project's language/framework.
type ProjectType string

const (
	Go  ProjectType = "go"
	PHP ProjectType = "php"
)

// IsGoProject returns true if dir contains a go.mod file.
func IsGoProject(dir string) bool {
	_, err := os.Stat(dir + "/go.mod")
	return err == nil
}

// IsPHPProject returns true if dir contains a composer.json file.
func IsPHPProject(dir string) bool {
	_, err := os.Stat(dir + "/composer.json")
	return err == nil
}

// DetectAll returns all detected project types in the directory.
func DetectAll(dir string) []ProjectType {
	var types []ProjectType
	if IsGoProject(dir) {
		types = append(types, Go)
	}
	if IsPHPProject(dir) {
		types = append(types, PHP)
	}
	return types
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/snider/Code && GOWORK=go.work go test ./core/lint/pkg/detect/... -v`
Expected: PASS

**Step 5: Commit**

```
feat(lint): add pkg/detect — project type detection
```

---

### Task 2: Create `pkg/php/format.go` — Pint Formatter

**Files:**
- Create: `pkg/php/format.go`
- Create: `pkg/php/format_test.go`

**Step 1: Write the failing test**

```go
package php

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectFormatter_Good(t *testing.T) {
	dir := t.TempDir()
	// Create vendor/bin/pint
	vendorBin := filepath.Join(dir, "vendor", "bin")
	os.MkdirAll(vendorBin, 0755)
	os.WriteFile(filepath.Join(vendorBin, "pint"), []byte("#!/bin/sh"), 0755)

	ft, found := DetectFormatter(dir)
	assert.True(t, found)
	assert.Equal(t, FormatterPint, ft)
}

func TestDetectFormatter_Bad(t *testing.T) {
	dir := t.TempDir()
	_, found := DetectFormatter(dir)
	assert.False(t, found)
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code && GOWORK=go.work go test ./core/lint/pkg/php/... -v`
Expected: FAIL

**Step 3: Write implementation**

Extract from `core/php/quality.go` lines 16-232 (FormatOptions, FormatterType, DetectFormatter, Format, buildPintCommand). Replace `getMedium()` with `io.Local`. Change package to `php`.

Key changes from source:
- `package php` (not `package php` from core/php — same name, different module)
- Replace `getMedium()` → `io.Local`
- Replace `cli.Sprintf` → `fmt.Sprintf`
- Keep `cli.WrapVerb`, `cli.Err` for error returns

**Step 4: Run test to verify it passes**

Run: `cd /Users/snider/Code && GOWORK=go.work go test ./core/lint/pkg/php/... -v`
Expected: PASS

**Step 5: Commit**

```
feat(lint): add pkg/php/format — Pint formatter detection and execution
```

---

### Task 3: Create `pkg/php/analyse.go` — PHPStan + Psalm

**Files:**
- Create: `pkg/php/analyse.go`
- Create: `pkg/php/analyse_test.go`

Extract from `core/php/quality.go`:
- `AnalyseOptions`, `AnalyserType`, `DetectAnalyser`, `Analyse`, `buildPHPStanCommand` (lines 37-266)
- `PsalmOptions`, `PsalmType`, `DetectPsalm`, `RunPsalm` (lines 272-370)

Same `getMedium()` → `io.Local` replacement.

Test: create temp dirs with mock `vendor/bin/phpstan`, verify `DetectAnalyser` finds it.

**Commit:**
```
feat(lint): add pkg/php/analyse — PHPStan, Larastan, Psalm
```

---

### Task 4: Create `pkg/php/audit.go` — Security Auditing

**Files:**
- Create: `pkg/php/audit.go`
- Create: `pkg/php/audit_test.go`

Extract from `core/php/quality.go`:
- `AuditOptions`, `AuditResult`, `AuditAdvisory`, `RunAudit`, `runComposerAudit`, `runNpmAudit` (lines 376-521)

Test: verify `AuditResult` struct fields, test JSON parsing of mock composer audit output.

**Commit:**
```
feat(lint): add pkg/php/audit — composer and npm security auditing
```

---

### Task 5: Create `pkg/php/security.go` — Security Checks

**Files:**
- Create: `pkg/php/security.go`
- Create: `pkg/php/security_test.go`

Extract from `core/php/quality.go`:
- `SecurityOptions`, `SecurityResult`, `SecurityCheck`, `SecuritySummary` (lines 781-817)
- `RunSecurityChecks`, `runEnvSecurityChecks`, `runFilesystemSecurityChecks` (lines 819-994)

Test: create temp .env with `APP_DEBUG=true`, verify check catches it.

**Commit:**
```
feat(lint): add pkg/php/security — .env and filesystem security checks
```

---

### Task 6: Create `pkg/php/refactor.go` + `pkg/php/mutation.go`

**Files:**
- Create: `pkg/php/refactor.go` — Rector
- Create: `pkg/php/mutation.go` — Infection
- Create: `pkg/php/refactor_test.go`
- Create: `pkg/php/mutation_test.go`

Extract Rector (lines 527-598) and Infection (lines 604-693) from `core/php/quality.go`.

**Commit:**
```
feat(lint): add pkg/php/refactor + mutation — Rector and Infection
```

---

### Task 7: Create `pkg/php/test.go` — Test Runner

**Files:**
- Create: `pkg/php/test.go`
- Create: `pkg/php/test_test.go`

Extract from `core/php/testing.go`:
- `TestOptions`, `TestRunner`, `DetectTestRunner`, `RunTests`, `buildPestCommand`, `buildPHPUnitCommand`

**Commit:**
```
feat(lint): add pkg/php/test — Pest and PHPUnit runner
```

---

### Task 8: Create `pkg/php/pipeline.go` + `pkg/php/runner.go` — QA Orchestration

**Files:**
- Create: `pkg/php/pipeline.go`
- Create: `pkg/php/runner.go`
- Create: `pkg/php/pipeline_test.go`

Extract from `core/php/quality.go`:
- `QAOptions`, `QAStage`, `QACheckResult`, `QAResult`, `GetQAStages`, `GetQAChecks` (lines 699-775) → `pipeline.go`

Extract from `core/php/cmd_qa_runner.go`:
- `QARunner`, `QARunResult`, `QACheckRunResult`, `NewQARunner`, `Run`, `BuildSpecs`, `buildSpec` → `runner.go`

This task adds `go-process` dependency to lint's go.mod.

**Commit:**
```
feat(lint): add pkg/php/pipeline + runner — QA orchestration
```

---

### Task 9: Add PHP commands to `cmd/qa/`

**Files:**
- Create: `cmd/qa/cmd_php.go`
- Modify: `cmd/qa/cmd_qa.go` — add auto-detect logic

Extract CLI commands from `core/php/cmd_quality.go`:
- `addPHPFmtCommand` → `core qa fmt` (detect: PHP → pint, Go → gofmt)
- `addPHPStanCommand` → `core qa stan`
- `addPHPPsalmCommand` → `core qa psalm`
- `addPHPAuditCommand` → `core qa audit`
- `addPHPSecurityCommand` → `core qa security`
- `addPHPRectorCommand` → `core qa rector`
- `addPHPInfectionCommand` → `core qa infection`
- `addPHPTestCommand` → `core qa test` (detect: PHP → pest/phpunit, Go → go test)
- `addPHPQACommand` → integrated into main `core qa` with auto-detect

Auto-detect in `cmd_qa.go`:
```go
types := detect.DetectAll(cwd)
for _, t := range types {
    switch t {
    case detect.Go:
        // run Go checks (existing)
    case detect.PHP:
        // run PHP checks (new)
    }
}
```

**Commit:**
```
feat(lint): add PHP QA commands to core qa
```

---

### Task 10: Update go.mod, build, test

**Files:**
- Modify: `go.mod` — add `go-process` dep
- Modify: `cmd/qa/cmd_qa.go` — register PHP subcommands

**Step 1:** Run `go mod tidy`
**Step 2:** Run `go build ./...`
**Step 3:** Run `go test ./...`
**Step 4:** Verify `core qa --help` shows new PHP commands

**Commit:**
```
feat(lint): wire up PHP QA, update dependencies
```

---

### Task 11: Remove QA code from core/php

**Files:**
- Delete: `core/php/quality.go`
- Delete: `core/php/quality_test.go`
- Delete: `core/php/quality_extended_test.go`
- Delete: `core/php/cmd_quality.go`
- Delete: `core/php/cmd_qa_runner.go`
- Delete: `core/php/qa.yaml`
- Modify: `core/php/cmd_commands.go` — remove QA command registrations

**Step 1:** Remove files
**Step 2:** Remove QA command registrations from `cmd_commands.go`
**Step 3:** Run `go build ./...` on core/php to verify nothing breaks
**Step 4:** Run `go build ./...` on core/lint to verify still compiles

**Commit:**
```
refactor(php): remove QA code — moved to core/lint
```

---

### Task 12: Final verification and tag

**Step 1:** Build all: `GOWORK=go.work go build ./core/lint/... ./core/php/...`
**Step 2:** Test all: `GOWORK=go.work go test ./core/lint/...`
**Step 3:** Verify CLI: build core binary, run `core qa --help`
**Step 4:** Tag lint v0.3.0

**Commit:**
```
chore(lint): tag v0.3.0 — unified QA provider
```
