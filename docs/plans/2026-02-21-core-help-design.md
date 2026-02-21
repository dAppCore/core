# core.help Documentation Website — Design

**Date:** 2026-02-21
**Author:** Virgil
**Status:** Design approved
**Domain:** https://core.help

## Problem

Documentation is scattered across 39 repos (18 Go packages, 20 PHP packages, 1 CLI). There is no unified docs site. Developers need a single entry point to find CLI commands, Go package APIs, MCP tool references, and PHP module guides.

## Solution

A Hugo + Docsy static site at core.help, built from existing markdown docs aggregated by `core docs sync`. No new content — just collect and present what already exists across the ecosystem.

## Architecture

### Stack

- **Hugo** — Go-native static site generator, sub-second builds
- **Docsy theme** — Purpose-built for technical docs (used by Kubernetes, gRPC, Knative)
- **BunnyCDN** — Static hosting with pull zone
- **`core docs sync --target hugo`** — Collects markdown from all repos into Hugo content tree

### Why Hugo + Docsy (not VitePress or mdBook)

- Go-native, no Node.js dependency
- Handles multi-section navigation (CLI, Go packages, PHP modules, MCP tools)
- Sub-second builds for ~250 markdown files
- Docsy has built-in search, versioned nav, API reference sections

## Content Structure

```
docs-site/
├── hugo.toml
├── content/
│   ├── _index.md                # Landing page
│   ├── getting-started/         # CLI top-level guides
│   │   ├── _index.md
│   │   ├── installation.md
│   │   ├── configuration.md
│   │   ├── user-guide.md
│   │   ├── troubleshooting.md
│   │   └── faq.md
│   ├── cli/                     # CLI command reference (43 commands)
│   │   ├── _index.md
│   │   ├── dev/                 # core dev commit, push, pull, etc.
│   │   ├── ai/                  # core ai commands
│   │   ├── go/                  # core go test, lint, etc.
│   │   └── ...
│   ├── go/                      # Go ecosystem packages (18)
│   │   ├── _index.md            # Ecosystem overview
│   │   ├── go-api/              # README + architecture/development/history
│   │   ├── go-ai/
│   │   ├── go-mlx/
│   │   ├── go-i18n/
│   │   └── ...
│   ├── mcp/                     # MCP tool reference (49 tools)
│   │   ├── _index.md
│   │   ├── file-operations.md
│   │   ├── process-management.md
│   │   ├── rag.md
│   │   └── ...
│   ├── php/                     # PHP packages (from core-php/docs/packages/)
│   │   ├── _index.md
│   │   ├── admin/
│   │   ├── tenant/
│   │   ├── commerce/
│   │   └── ...
│   └── kb/                      # Knowledge base (wiki pages from go-mlx, go-i18n)
│       ├── _index.md
│       ├── mlx/
│       └── i18n/
├── static/                      # Logos, favicons
├── layouts/                     # Custom template overrides (minimal)
└── go.mod                       # Hugo modules (Docsy as module dep)
```

## Sync Pipeline

`core docs sync --target hugo --output site/content/` performs:

### Source Mapping

```
cli/docs/index.md              → content/getting-started/_index.md
cli/docs/getting-started.md    → content/getting-started/installation.md
cli/docs/user-guide.md         → content/getting-started/user-guide.md
cli/docs/configuration.md      → content/getting-started/configuration.md
cli/docs/troubleshooting.md    → content/getting-started/troubleshooting.md
cli/docs/faq.md                → content/getting-started/faq.md

core/docs/cmd/**/*.md          → content/cli/**/*.md

go-*/README.md                 → content/go/{name}/_index.md
go-*/docs/*.md                 → content/go/{name}/*.md
go-*/KB/*.md                   → content/kb/{name-suffix}/*.md

core-*/docs/**/*.md            → content/php/{name-suffix}/**/*.md
```

### Front Matter Injection

If a markdown file doesn't start with `---`, prepend:

```yaml
---
title: "{derived from filename}"
linkTitle: "{short name}"
weight: {auto-incremented}
---
```

No other content transformations. Markdown stays as-is.

### Build & Deploy

```bash
core docs sync --target hugo --output docs-site/content/
cd docs-site && hugo build
hugo deploy --target bunnycdn
```

Hugo deploy config in `hugo.toml`:

```toml
[deployment]
[[deployment.targets]]
name = "bunnycdn"
URL = "s3://core-help?endpoint=storage.bunnycdn.com&region=auto"
```

Credentials via env vars.

## Registry

All 39 repos registered in `.core/repos.yaml` with `docs: true`. Go repos use explicit `path:` fields since they live outside the PHP `base_path`. `FindRegistry()` checks `.core/repos.yaml` alongside `repos.yaml`.

## Prerequisites Completed

- [x] `.core/repos.yaml` created with all 39 repos
- [x] `FindRegistry()` updated to find `.core/repos.yaml`
- [x] `Repo.Path` supports explicit YAML override
- [x] go-api docs gap filled (architecture.md, development.md, history.md)
- [x] All 18 Go repos have standard docs trio

## What Remains (Implementation Plan)

1. Create docs-site repo with Hugo + Docsy scaffold
2. Extend `core docs sync` with `--target hugo` mode
3. Write section _index.md files (landing page, section intros)
4. Hugo config (navigation, search, theme colours)
5. BunnyCDN deployment config
6. CI pipeline on Forge (optional — can deploy manually initially)
