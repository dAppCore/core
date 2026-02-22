# CLI Meta-Package Restructure — Completed

**Completed:** 22 Feb 2026

## What Was Done

`pkg/cli` was extracted from `core/go` into its own Go module at `forge.lthn.ai/core/cli`. This made the CLI SDK a first-class, independently versioned package rather than a subdirectory of the Go foundation repo.

Following the extraction, an ecosystem-wide import path migration updated all consumers from the old path to the new one:

- Old: `forge.lthn.ai/core/go/pkg/cli`
- New: `forge.lthn.ai/core/cli/pkg/cli`

## Scope

- **147+ files** updated across **10 repos**
- All repos build clean after migration

## Repos Migrated

`core/cli`, `core/go`, `go-devops`, `go-ai`, `go-agentic`, `go-crypt`, `go-rag`, `go-scm`, `go-api`, `go-update`

## Key Outcomes

- `forge.lthn.ai/core/cli/pkg/cli` is the single import for all CLI concerns across the ecosystem
- Domain repos are insulated from cobra, lipgloss, and bubbletea — only `pkg/cli` imports them
- Command registration uses the Core framework lifecycle via `cli.WithCommands()` — no `init()`, no global state
- `core/cli` is a thin assembly repo (~2K LOC) with 7 meta packages; all business logic lives in domain repos
- Variant binary pattern established: multiple `main.go` files can wire different `WithCommands` sets for targeted binaries (core-ci, core-mlx, core-ops, etc.)
- Command migration from the old `core/cli` monolith to domain repos was completed in full (13 command groups moved)
