# Core Go Framework — Documentation

Core is a Go framework and unified CLI for the host-uk ecosystem. It provides two complementary things: a **dependency injection container** for building Go services and Wails v3 desktop applications, and a **command-line tool** for managing the full development lifecycle across Go, PHP, and container workloads.

The `core` binary is the single entry point for all development tasks: testing, building, releasing, multi-repo management, MCP servers, and AI-assisted workflows.

---

## Getting Started

| Document | Description |
|----------|-------------|
| [Getting Started](getting-started.md) | Install the CLI, run your first build, and set up a multi-repo workspace |
| [User Guide](user-guide.md) | Key concepts and daily workflow patterns |
| [Workflows](workflows.md) | End-to-end task sequences for common scenarios |
| [FAQ](faq.md) | Answers to common questions |

---

## Architecture

| Document | Description |
|----------|-------------|
| [Package Standards](pkg/PACKAGE_STANDARDS.md) | Canonical patterns for creating packages: Service struct, factory, IPC, thread safety |
| [pkg/i18n — Grammar](pkg/i18n/GRAMMAR.md) | Grammar engine internals and language rule format |
| [pkg/i18n — Extending](pkg/i18n/EXTENDING.md) | How to add new locales and translation files |

### Framework Architecture Summary

The Core framework (`pkg/framework`) is a dependency injection container built around three ideas:

**Service registry.** Services are registered via factory functions and retrieved with type-safe generics:

```go
core, _ := framework.New(
    framework.WithService(mypackage.NewService(opts)),
)
svc, _ := framework.ServiceFor[*mypackage.Service](core, "mypackage")
```

**Lifecycle.** Services implementing `Startable` or `Stoppable` are called automatically during boot and shutdown.

**ACTION bus.** Services communicate by broadcasting typed messages via `core.ACTION(msg)` and registering handlers via `core.RegisterAction()`. This decouples packages without requiring direct imports between them.

---

## Command Reference

The `core` CLI is documented command-by-command in `docs/cmd/`:

| Command | Description |
|---------|-------------|
| [cmd/](cmd/) | Full command index |
| [cmd/go/](cmd/go/) | Go development: test, fmt, lint, coverage, mod, work |
| [cmd/php/](cmd/php/) | Laravel/PHP development: dev server, test, deploy |
| [cmd/build/](cmd/build/) | Build Go, Wails, Docker, LinuxKit projects |
| [cmd/ci/](cmd/ci/) | Publish releases to GitHub, Docker, npm, Homebrew |
| [cmd/sdk/](cmd/sdk/) | SDK generation and OpenAPI validation |
| [cmd/dev/](cmd/dev/) | Multi-repo workflow and sandboxed dev environment |
| [cmd/ai/](cmd/ai/) | AI task management and Claude integration |
| [cmd/pkg/](cmd/pkg/) | Package search and install |
| [cmd/vm/](cmd/vm/) | LinuxKit VM management |
| [cmd/docs/](cmd/docs/) | Documentation sync and management |
| [cmd/setup/](cmd/setup/) | Clone repositories from a registry |
| [cmd/doctor/](cmd/doctor/) | Verify development environment |
| [cmd/test/](cmd/test/) | Run Go tests with coverage reporting |

---

## Packages

The Core repository contains the following internal packages. Full API analysis for each is available in the batch analysis documents listed under [Reference](#reference).

### Foundation

| Package | Description |
|---------|-------------|
| `pkg/framework` | Dependency injection container; re-exports `pkg/framework/core` |
| `pkg/log` | Structured logger with `Err` error type, operation chains, and log rotation |
| `pkg/config` | 12-factor config management layered over Viper; accepts `io.Medium` |
| `pkg/io` | Filesystem abstraction (`Medium` interface); `NewSandboxed`, `MockMedium` |
| `pkg/crypt` | Opinionated crypto: Argon2id passwords, ChaCha20 encryption, HMAC |
| `pkg/cache` | File-based JSON cache with TTL expiry |
| `pkg/i18n` | Grammar engine with pluralisation, verb conjugation, semantic sentences |

### CLI and Interaction

| Package | Description |
|---------|-------------|
| `pkg/cli` | CLI runtime: Cobra wrapping, ANSI styling, prompts, daemon lifecycle |
| `pkg/help` | Embedded documentation catalogue with in-memory full-text search |
| `pkg/session` | Claude Code JSONL transcript parser; HTML and MP4 export |
| `pkg/workspace` | Isolated, PGP-keyed workspace environments with IPC control |

### Build and Release

| Package | Description |
|---------|-------------|
| `pkg/build` | Project type detection, cross-compilation, archiving, checksums |
| `pkg/release` | Semantic versioning, conventional-commit changelogs, multi-target publishing |
| `pkg/container` | LinuxKit VM lifecycle via QEMU/Hyperkit; template management |
| `pkg/process` | `os/exec` wrapper with ring-buffer output, DAG task runner, ACTION streaming |
| `pkg/jobrunner` | Poll-dispatch automation engine with JSONL audit journal |

### Source Control and Hosting

| Package | Description |
|---------|-------------|
| `pkg/git` | Multi-repo status, push, pull; concurrent status checks |
| `pkg/repos` | `repos.yaml` registry loader; topological dependency ordering |
| `pkg/gitea` | Gitea API client with PR metadata extraction |
| `pkg/forge` | Forgejo API client with PR metadata extraction |
| `pkg/plugin` | Git-based CLI extension system |

### AI and Agentic

| Package | Description |
|---------|-------------|
| `pkg/mcp` | MCP server exposing file, process, RAG, and CDP tools to AI agents |
| `pkg/rag` | RAG pipeline: Markdown chunking, Ollama embeddings, Qdrant vector search |
| `pkg/ai` | Facade over RAG and metrics; `QueryRAGForTask` for prompt enrichment |
| `pkg/agentic` | REST client for core-agentic; `AutoCommit`, `CreatePR`, `BuildTaskContext` |
| `pkg/agentci` | Configuration bridge for AgentCI dispatch targets |
| `pkg/collect` | Data collection pipeline from GitHub, forums, market APIs |

### Infrastructure and Networking

| Package | Description |
|---------|-------------|
| `pkg/devops` | LinuxKit dev environment lifecycle; SSH bridging; project auto-detection |
| `pkg/ansible` | Native Go Ansible-lite engine; SSH playbook execution without the CLI |
| `pkg/webview` | Chrome DevTools Protocol client; Angular-aware automation |
| `pkg/ws` | WebSocket hub with channel-based subscriptions |
| `pkg/unifi` | UniFi controller client for network management |
| `pkg/auth` | OpenPGP challenge-response authentication; air-gapped flow |

---

## Workflows

| Document | Description |
|----------|-------------|
| [Workflows](workflows.md) | Go build and release, PHP deploy, multi-repo daily workflow, hotfix |
| [Migration](migration.md) | Migrating from `push-all.sh`, raw `go` commands, `goreleaser`, or manual git |

---

## Reference

| Document | Description |
|----------|-------------|
| [Configuration](configuration.md) | `.core/` directory, `release.yaml`, `build.yaml`, `php.yaml`, `repos.yaml`, environment variables |
| [Glossary](glossary.md) | Term definitions: target, workspace, registry, publisher, dry-run |
| [Troubleshooting](troubleshooting.md) | Installation failures, build errors, release issues, multi-repo problems, PHP issues |
| [Claude Code Skill](skill/) | Install the `core` skill to teach Claude Code how to use this CLI |

### Historical Package Analysis

The following documents were generated by an automated analysis pipeline (Gemini, February 2026) to extract architecture, public API, and test coverage notes from each package. They remain valid as architectural reference.

| Document | Packages Covered |
|----------|-----------------|
| [pkg-batch1-analysis.md](pkg-batch1-analysis.md) | `pkg/log`, `pkg/config`, `pkg/io`, `pkg/crypt`, `pkg/auth` |
| [pkg-batch2-analysis.md](pkg-batch2-analysis.md) | `pkg/cli`, `pkg/help`, `pkg/session`, `pkg/workspace` |
| [pkg-batch3-analysis.md](pkg-batch3-analysis.md) | `pkg/build`, `pkg/container`, `pkg/process`, `pkg/jobrunner` |
| [pkg-batch4-analysis.md](pkg-batch4-analysis.md) | `pkg/git`, `pkg/repos`, `pkg/gitea`, `pkg/forge`, `pkg/release` |
| [pkg-batch5-analysis.md](pkg-batch5-analysis.md) | `pkg/agentci`, `pkg/agentic`, `pkg/ai`, `pkg/rag` |
| [pkg-batch6-analysis.md](pkg-batch6-analysis.md) | `pkg/ansible`, `pkg/devops`, `pkg/framework`, `pkg/mcp`, `pkg/plugin`, `pkg/unifi`, `pkg/webview`, `pkg/ws`, `pkg/collect`, `pkg/i18n`, `pkg/cache` |

### Design Plans

| Document | Description |
|----------|-------------|
| [plans/2026-02-05-core-ide-job-runner-design.md](plans/2026-02-05-core-ide-job-runner-design.md) | Autonomous job runner design for core-ide: poller, dispatcher, MCP handler registry, JSONL training data |
| [plans/2026-02-05-core-ide-job-runner-plan.md](plans/2026-02-05-core-ide-job-runner-plan.md) | Implementation plan for the job runner |
| [plans/2026-02-05-mcp-integration.md](plans/2026-02-05-mcp-integration.md) | MCP integration design notes |
| [plans/2026-02-17-lem-chat-design.md](plans/2026-02-17-lem-chat-design.md) | LEM Chat Web Components design: streaming SSE, zero-dependency vanilla UI |

---

## Satellite Packages

The Core ecosystem extends across 19 standalone Go modules, all hosted under `forge.lthn.ai/core/`. Each has its own repository and `docs/` directory.

See [ecosystem.md](ecosystem.md) for the full map, module paths, and dependency graph.

| Package | Purpose |
|---------|---------|
| [go-inference](ecosystem.md#go-inference) | Shared `TextModel`/`Backend`/`Token` interfaces — the common contract |
| [go-mlx](ecosystem.md#go-mlx) | Native Metal GPU inference via CGO/mlx-c (Apple Silicon) |
| [go-rocm](ecosystem.md#go-rocm) | AMD ROCm GPU inference via llama-server subprocess |
| [go-ml](ecosystem.md#go-ml) | Scoring engine, backends, agent orchestrator |
| [go-ai](ecosystem.md#go-ai) | MCP hub with 49 registered tools |
| [go-agentic](ecosystem.md#go-agentic) | Service lifecycle and allowance management for agents |
| [go-rag](ecosystem.md#go-rag) | Qdrant vector search and Ollama embeddings |
| [go-i18n](ecosystem.md#go-i18n) | Grammar engine, reversal, GrammarImprint |
| [go-html](ecosystem.md#go-html) | HLCRF DOM compositor and WASM target |
| [go-crypt](ecosystem.md#go-crypt) | Cryptographic primitives, auth, trust policies |
| [go-scm](ecosystem.md#go-scm) | SCM/CI integration and AgentCI dispatch |
| [go-p2p](ecosystem.md#go-p2p) | P2P mesh networking and UEPS wire protocol |
| [go-devops](ecosystem.md#go-devops) | Ansible automation, build tooling, infrastructure, release |
| [go-help](ecosystem.md#go-help) | YAML help catalogue with full-text search and HTTP server |
| [go-ratelimit](ecosystem.md#go-ratelimit) | Sliding-window rate limiter with SQLite backend |
| [go-session](ecosystem.md#go-session) | Claude Code JSONL transcript parser |
| [go-store](ecosystem.md#go-store) | SQLite key-value store with `Watch`/`OnChange` |
| [go-ws](ecosystem.md#go-ws) | WebSocket hub with Redis bridge |
| [go-webview](ecosystem.md#go-webview) | Chrome DevTools Protocol automation client |
