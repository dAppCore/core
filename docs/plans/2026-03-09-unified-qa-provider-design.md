# Unified QA Provider — core/lint

**Date:** 2026-03-09
**Status:** Approved

## Problem

PHP QA tooling (~2,150 LOC) lives in `core/php` alongside dev server, deploy, and service management code. Go QA tooling (~3,500 LOC) lives in `core/lint`. Two separate QA entry points (`core php qa` and `core qa`) fragment the developer experience.

## Solution

Make `core/lint` the unified QA provider for all languages. Extract PHP QA library and CLI code from `core/php` into `core/lint` under language sub-packages.

## Architecture

```
lint/
├── pkg/detect/          # Project type detection
│   └── detect.go        # IsPHPProject(), IsGoProject(), DetectAll()
├── pkg/lint/            # Go analysis (unchanged)
│   ├── complexity.go
│   ├── coverage.go
│   ├── scanner.go
│   ├── tools.go
│   ├── vulncheck.go
│   └── ...
├── pkg/php/             # PHP analysis (from core/php quality.go)
│   ├── format.go        # Pint (DetectFormatter, Format)
│   ├── analyse.go       # PHPStan/Larastan, Psalm
│   ├── audit.go         # Composer/npm audit
│   ├── security.go      # .env + filesystem security checks
│   ├── refactor.go      # Rector
│   ├── mutation.go      # Infection
│   ├── pipeline.go      # QA pipeline stages
│   └── runner.go        # QARunner orchestration (go-process)
├── cmd/qa/              # Unified CLI
│   ├── cmd_qa.go        # Root — auto-detects project type
│   ├── cmd_docblock.go  # (existing Go)
│   ├── cmd_health.go    # (existing Go)
│   ├── cmd_php.go       # PHP: fmt, stan, psalm, audit, security, rector, infection
│   └── ...
└── cmd/core-lint/main.go
```

## Key Decisions

### Project Detection (`pkg/detect/`)
- Uses `go-io` Medium for filesystem checks
- Exports `IsPHPProject(dir)`, `IsGoProject(dir)`, `DetectAll(dir) []ProjectType`
- Both `pkg/lint` and `pkg/php` import this shared package

### PHP Library (`pkg/php/`)
- Pure library, no CLI coupling
- Option structs in, result structs out
- Replaces `getMedium()` with `io.NewMedium()` directly
- No dependency on `core/php` — fully standalone
- Tools: Pint, PHPStan/Larastan, Psalm, Rector, Infection, composer/npm audit, security

### QA Runner (`pkg/php/runner.go`)
- Uses `go-process` for subprocess orchestration with dependency ordering
- Stages: quick (audit, fmt, stan), standard (psalm, test), full (rector, infection)
- JSON output mode for CI

### Unified CLI (`cmd/qa/`)
- `core qa` auto-detects: Go project → Go checks, PHP project → PHP checks, both → both
- Individual tools: `core qa fmt`, `core qa stan`, `core qa psalm`, etc.
- Existing Go commands unchanged

### core/php Cleanup
- Remove: `quality.go`, `cmd_quality.go`, `cmd_qa_runner.go`, `qa.yaml`
- `core php qa` removed (users run `core qa`)
- core/php retains: dev server, deploy, build, services, container, FrankenPHP

## Dependencies

lint gains:
- `go-io` (already present)
- `go-process` (new — for QA runner subprocess orchestration)
- `go-i18n` (already present)

## Migration

| Source (core/php) | Destination (core/lint) |
|---|---|
| `quality.go` (format section) | `pkg/php/format.go` |
| `quality.go` (analyse section) | `pkg/php/analyse.go` |
| `quality.go` (audit section) | `pkg/php/audit.go` |
| `quality.go` (security section) | `pkg/php/security.go` |
| `quality.go` (rector section) | `pkg/php/refactor.go` |
| `quality.go` (infection section) | `pkg/php/mutation.go` |
| `quality.go` (pipeline section) | `pkg/php/pipeline.go` |
| `cmd_qa_runner.go` | `pkg/php/runner.go` |
| `cmd_quality.go` (all commands) | `cmd/qa/cmd_php.go` |
| `qa.yaml` | `pkg/php/qa.yaml` (embedded) |
| `IsPHPProject()` from detect.go | `pkg/detect/detect.go` |
