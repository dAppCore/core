# Core Go Ecosystem

The Core Go ecosystem is a set of 19 standalone Go modules that form the infrastructure backbone for the host-uk platform and the Lethean network. All modules are hosted under the `forge.lthn.ai/core/` organisation. Each module has its own repository, independent versioning, and a `docs/` directory.

The CLI framework documented in the rest of this site (`forge.lthn.ai/core/cli`) is one node in this graph. The satellite packages listed here are separate repositories that the CLI imports or that stand alone as libraries.

---

## Module Index

| Package | Module Path | Managed By |
|---------|-------------|-----------|
| [go-inference](#go-inference) | `forge.lthn.ai/core/go-inference` | Virgil |
| [go-mlx](#go-mlx) | `forge.lthn.ai/core/go-mlx` | Virgil |
| [go-rocm](#go-rocm) | `forge.lthn.ai/core/go-rocm` | Charon |
| [go-ml](#go-ml) | `forge.lthn.ai/core/go-ml` | Virgil |
| [go-ai](#go-ai) | `forge.lthn.ai/core/go-ai` | Virgil |
| [go-agentic](#go-agentic) | `forge.lthn.ai/core/go-agentic` | Charon |
| [go-rag](#go-rag) | `forge.lthn.ai/core/go-rag` | Charon |
| [go-i18n](#go-i18n) | `forge.lthn.ai/core/go-i18n` | Virgil |
| [go-html](#go-html) | `forge.lthn.ai/core/go-html` | Charon |
| [go-crypt](#go-crypt) | `forge.lthn.ai/core/go-crypt` | Virgil |
| [go-scm](#go-scm) | `forge.lthn.ai/core/go-scm` | Charon |
| [go-p2p](#go-p2p) | `forge.lthn.ai/core/go-p2p` | Charon |
| [go-devops](#go-devops) | `forge.lthn.ai/core/go-devops` | Virgil |
| [go-help](#go-help) | `forge.lthn.ai/core/go-help` | Charon |
| [go-ratelimit](#go-ratelimit) | `forge.lthn.ai/core/go-ratelimit` | Charon |
| [go-session](#go-session) | `forge.lthn.ai/core/go-session` | Charon |
| [go-store](#go-store) | `forge.lthn.ai/core/go-store` | Charon |
| [go-ws](#go-ws) | `forge.lthn.ai/core/go-ws` | Charon |
| [go-webview](#go-webview) | `forge.lthn.ai/core/go-webview` | Charon |

---

## Dependency Graph

The graph below shows import relationships. An arrow `A → B` means A imports B.

```
go-inference          (no dependencies — foundation contract)
     ↑
     ├── go-mlx       (CGO, Apple Silicon Metal GPU)
     ├── go-rocm      (AMD ROCm, llama-server subprocess)
     └── go-ml        (scoring engine, backends, orchestrator)
              ↑
              └── go-ai  (MCP hub, 49 tools)
                    ↑
                    └── go-agentic  (service lifecycle, allowances)

go-rag               (Qdrant + Ollama, standalone)
     ↑
     └── go-ai

go-i18n              (grammar engine, standalone; Phase 2a imports go-mlx)

go-crypt             (standalone)
     ↑
     ├── go-p2p       (UEPS wire protocol)
     └── go-scm       (AgentCI dispatch)

go-store             (SQLite KV, standalone)
     ↑
     ├── go-ratelimit (sliding window limiter)
     ├── go-session   (transcript parser)
     └── go-agentic

go-ws                (WebSocket hub, standalone)
     ↑
     └── go-ai

go-webview           (CDP client, standalone)
     ↑
     └── go-ai

go-html              (DOM compositor, standalone)

go-help              (help catalogue, standalone)

go-devops            (Ansible, build, infrastructure — imports go-scm)
```

The CLI framework (`forge.lthn.ai/core/cli`) has internal equivalents of several of these packages (`pkg/rag`, `pkg/ws`, `pkg/webview`, `pkg/i18n`) that were developed in parallel. The satellite packages are the canonical standalone versions intended for use outside the CLI binary.

---

## Package Descriptions

### go-inference

**Module:** `forge.lthn.ai/core/go-inference`

Zero-dependency interface package that defines the common contract for all inference backends in the ecosystem:

- `TextModel` — the top-level model interface (`Generate`, `Stream`, `Close`)
- `Backend` — hardware/runtime abstraction (Metal, ROCm, CPU, remote)
- `Token` — streaming token type with metadata

No concrete implementations live here. Any package that needs to call inference without depending on a specific hardware library imports `go-inference` and receives an implementation at runtime.

---

### go-mlx

**Module:** `forge.lthn.ai/core/go-mlx`

Native Metal GPU inference for Apple Silicon using CGO bindings to `mlx-c` (the C API for Apple's MLX framework). Implements the `go-inference` interfaces.

Build requirements:
- macOS 13+ (Ventura) on Apple Silicon
- `mlx-c` installed (`brew install mlx`)
- CGO enabled: `CGO_CFLAGS` and `CGO_LDFLAGS` must reference the mlx-c headers and library

Features:
- Loads GGUF and MLX-format models
- Streaming token generation directly on GPU
- Quantised model support (Q4, Q8)
- Phase 4 backend abstraction in progress — will allow hot-swapping backends at runtime

Local path: `/Users/snider/Code/go-mlx`

---

### go-rocm

**Module:** `forge.lthn.ai/core/go-rocm`

AMD ROCm GPU inference for Linux. Rather than using CGO, this package manages a `llama-server` subprocess (from llama.cpp) compiled with ROCm support and communicates over its HTTP API.

Features:
- Subprocess lifecycle management (start, health-check, restart on crash)
- OpenAI-compatible HTTP client wrapping llama-server's API
- Implements `go-inference` interfaces
- Targeted at the homelab RX 7800 XT running Ubuntu 24.04

Managed by Charon (Linux homelab).

---

### go-ml

**Module:** `forge.lthn.ai/core/go-ml`

Scoring engine, backend registry, and agent orchestration layer. The hub that connects models from `go-mlx`, `go-rocm`, and future backends into a unified interface.

Features:
- Backend registry: register multiple inference backends, select by capability
- Scoring pipeline: evaluate model outputs against rubrics
- Agent orchestrator: coordinate multi-step inference tasks
- ~3.5K LOC

---

### go-ai

**Module:** `forge.lthn.ai/core/go-ai`

MCP (Model Context Protocol) server hub with 49 registered tools. Acts as the primary facade for AI capabilities in the ecosystem.

Features:
- 49 MCP tools covering file operations, RAG, metrics, process management, WebSocket, and CDP/webview
- Imports `go-ml`, `go-rag`, `go-mlx`
- Can run as stdio MCP server or TCP MCP server
- AI usage metrics recorded to JSONL

Run the MCP server:

```bash
# stdio (for Claude Desktop / Claude Code)
core mcp serve

# TCP
MCP_ADDR=:9000 core mcp serve
```

---

### go-agentic

**Module:** `forge.lthn.ai/core/go-agentic`

Service lifecycle and allowance management for autonomous agents. Handles:

- Agent session tracking and state persistence
- Allowance system: budget constraints on tool calls, token usage, and wall-clock time
- Integration with `go-store` for persistence
- REST client for the PHP `core-agentic` backend

Managed by Charon.

---

### go-rag

**Module:** `forge.lthn.ai/core/go-rag`

Retrieval-Augmented Generation pipeline using Qdrant for vector storage and Ollama for embeddings.

Features:
- `ChunkMarkdown`: semantic splitting by H2 headers and paragraphs with overlap
- `Ingest`: crawl a directory of Markdown files, embed, and store in Qdrant
- `Query`: semantic search returning ranked `QueryResult` slices
- `FormatResultsContext`: formats results as XML tags for LLM prompt injection
- Clients: `QdrantClient` and `OllamaClient` wrapping their respective Go SDKs

Managed by Charon.

---

### go-i18n

**Module:** `forge.lthn.ai/core/go-i18n`

Grammar engine for natural-language generation. Goes beyond key-value lookup tables to handle pluralisation, verb conjugation, past tense, gerunds, and semantic sentence construction ("Subject verbed object").

Features:
- `T(key, args...)` — main translation function
- `S(noun, value)` — semantic subject with grammatical context
- Language rules defined in JSON; algorithmic fallbacks for irregular verbs
- **GrammarImprint**: a linguistic hash (reversal of the grammar engine) used as a semantic fingerprint — part of the Lethean identity verification stack
- Phase 2a (imports `go-mlx` for language model-assisted reversal) currently blocked on `go-mlx` Phase 4

Local path: `/Users/snider/Code/go-i18n`

---

### go-html

**Module:** `forge.lthn.ai/core/go-html`

HLCRF DOM compositor — a programmatic HTML/DOM construction library targeting both server-side rendering and WASM (browser).

HLCRF stands for Header, Left, Content, Right, Footer — the region layout model used throughout the CLI's terminal UI and web rendering layer.

Features:
- Composable region-based layout (mirrors the terminal `Composite` in `pkg/cli`)
- WASM build target: runs in the browser without JavaScript
- Used by the LEM Chat UI and web SDK generation

Managed by Charon.

---

### go-crypt

**Module:** `forge.lthn.ai/core/go-crypt`

Cryptographic primitives, authentication, and trust policy enforcement.

Features:
- Password hashing (Argon2id with tuned parameters)
- Symmetric encryption (ChaCha20-Poly1305, AES-GCM)
- Key derivation (HKDF, Scrypt)
- OpenPGP challenge-response authentication
- Trust policies: define and evaluate access rules
- Foundation for the UEPS (User-controlled Encryption Policy System) wire protocol in `go-p2p`

---

### go-scm

**Module:** `forge.lthn.ai/core/go-scm`

Source control management and CI integration, including the AgentCI dispatch system.

Features:
- Forgejo and Gitea API clients (typed wrappers)
- GitHub integration via the `gh` CLI
- `AgentCI`: dispatches AI work items to agent runners over SSH using Charm stack libraries (`soft-serve`, `keygen`, `melt`, `wishlist`)
- PR lifecycle management: create, review, merge, label
- JSONL job journal for audit trails

Managed by Charon.

---

### go-p2p

**Module:** `forge.lthn.ai/core/go-p2p`

Peer-to-peer mesh networking implementing the UEPS (User-controlled Encryption Policy System) wire protocol.

Features:
- UEPS: consent-gated TLV frames with Ed25519 consent tokens and an Intent-Broker
- Peer discovery and mesh routing
- Encrypted relay transport
- Integration with `go-crypt` for all cryptographic operations

This is a core component of the Lethean Web3 network layer.

Managed by Charon (Linux homelab).

---

### go-devops

**Module:** `forge.lthn.ai/core/go-devops`

Infrastructure automation, build tooling, and release pipeline utilities, intended as a standalone library form of what the Core CLI provides as commands.

Features:
- Ansible-lite engine (native Go SSH playbook execution)
- LinuxKit image building and VM lifecycle
- Multi-target binary build and release
- Integration with `go-scm` for repository operations

---

### go-help

**Module:** `forge.lthn.ai/core/go-help`

Embedded documentation catalogue with full-text search and an optional HTTP server for serving help content.

Features:
- YAML-frontmatter Markdown topic parsing
- In-memory reverse index with title/heading/body scoring
- Snippet extraction with keyword highlighting
- `HTTP server` mode: serve the catalogue as a documentation site
- Used by the `core pkg search` command and the `pkg/help` package inside the CLI

Managed by Charon.

---

### go-ratelimit

**Module:** `forge.lthn.ai/core/go-ratelimit`

Sliding-window rate limiter with a SQLite persistence backend.

Features:
- Token bucket and sliding-window algorithms
- SQLite backend via `go-store` for durable rate state across restarts
- HTTP middleware helper
- Used by `go-ai` and `go-agentic` to enforce per-agent API quotas

Managed by Charon.

---

### go-session

**Module:** `forge.lthn.ai/core/go-session`

Claude Code JSONL transcript parser and visualisation toolkit (standalone version of `pkg/session` inside the CLI).

Features:
- `ParseTranscript(path)`: reads `.jsonl` session files and reconstructs tool use timelines
- `ListSessions(dir)`: scans a Claude projects directory for session files
- `Search(dir, query)`: full-text search across sessions
- `RenderHTML(sess, path)`: single-file HTML visualisation
- `RenderMP4(sess, path)`: terminal video replay via VHS

Managed by Charon.

---

### go-store

**Module:** `forge.lthn.ai/core/go-store`

SQLite-backed key-value store with reactive change notification.

Features:
- `Get`, `Set`, `Delete`, `List` over typed keys
- `Watch(key, handler)`: register a callback that fires on change
- `OnChange(handler)`: subscribe to all changes
- Used by `go-ratelimit`, `go-session`, and `go-agentic` for lightweight persistence

Managed by Charon.

---

### go-ws

**Module:** `forge.lthn.ai/core/go-ws`

WebSocket hub with channel-based subscriptions and an optional Redis pub/sub bridge for multi-instance deployments.

Features:
- Hub pattern: central registry of connected clients
- Channel routing: `SendToChannel(topic, msg)` delivers only to subscribers
- Redis bridge: publish messages from one instance, receive on all
- HTTP handler: `hub.Handler()` for embedding in any Go HTTP server
- `SendProcessOutput(id, line)`: convenience method for streaming process logs

Managed by Charon.

---

### go-webview

**Module:** `forge.lthn.ai/core/go-webview`

Chrome DevTools Protocol (CDP) client for browser automation, testing, and AI-driven web interaction (standalone version of `pkg/webview` inside the CLI).

Features:
- Navigation, click, type, screenshot
- `Evaluate(script)`: arbitrary JavaScript execution with result capture
- Console capture and filtering
- Angular-aware helpers: `WaitForAngular()`, `GetNgModel(selector)`
- `ActionSequence`: chain interactions into a single call
- Used by `go-ai` to expose browser tools to MCP agents

Managed by Charon.

---

## Forge Repository Paths

All repositories are hosted at `forge.lthn.ai` (Forgejo). SSH access uses port 2223:

```
ssh://git@forge.lthn.ai:2223/core/go-inference.git
ssh://git@forge.lthn.ai:2223/core/go-mlx.git
ssh://git@forge.lthn.ai:2223/core/go-rocm.git
ssh://git@forge.lthn.ai:2223/core/go-ml.git
ssh://git@forge.lthn.ai:2223/core/go-ai.git
ssh://git@forge.lthn.ai:2223/core/go-agentic.git
ssh://git@forge.lthn.ai:2223/core/go-rag.git
ssh://git@forge.lthn.ai:2223/core/go-i18n.git
ssh://git@forge.lthn.ai:2223/core/go-html.git
ssh://git@forge.lthn.ai:2223/core/go-crypt.git
ssh://git@forge.lthn.ai:2223/core/go-scm.git
ssh://git@forge.lthn.ai:2223/core/go-p2p.git
ssh://git@forge.lthn.ai:2223/core/go-devops.git
ssh://git@forge.lthn.ai:2223/core/go-help.git
ssh://git@forge.lthn.ai:2223/core/go-ratelimit.git
ssh://git@forge.lthn.ai:2223/core/go-session.git
ssh://git@forge.lthn.ai:2223/core/go-store.git
ssh://git@forge.lthn.ai:2223/core/go-ws.git
ssh://git@forge.lthn.ai:2223/core/go-webview.git
```

HTTPS authentication is not available on Forge. Always use SSH remotes.

---

## Go Workspace Setup

The satellite packages can be used together in a Go workspace. After cloning the repositories you need:

```bash
go work init
go work use ./go-inference ./go-mlx ./go-rag ./go-ai   # add as needed
go work sync
```

The CLI repository already uses a Go workspace that includes `cmd/core-gui`, `cmd/bugseti`, and `cmd/examples/*`.

---

## See Also

- [index.md](index.md) — Main documentation hub
- [getting-started.md](getting-started.md) — CLI installation
- [configuration.md](configuration.md) — `repos.yaml` registry format
