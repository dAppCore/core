# go-devops Decomposition Design

**Goal:** Break `core/go-devops` (31K LOC monolith) into logical, independently-versioned packages now that the Core Go framework is stable as a base layer.

**Architecture:** Extract 4 loosely-coupled packages into dedicated repos. Keep the tightly-coupled build/release pipeline as go-devops's core purpose. Redirect imports so consumers pull from the right place.

**Tech Stack:** Go 1.26, existing Core Go ecosystem conventions (.core/build.yaml, cli.Main(), forge SSH).

---

## Current State

go-devops is a 31K LOC catch-all containing:

| Package | LOC | Purpose | Coupling |
|---------|-----|---------|----------|
| `build/` + `build/builders/` + `build/signing/` | 6.8K | Cross-compilation framework, 8 language builders | **Tight** (core purpose) |
| `release/` + `release/publishers/` | 5.9K | Release orchestration, 8 publishers, changelog | **Tight** (core purpose) |
| `sdk/` + `sdk/generators/` | 1.3K | OpenAPI SDK generation, breaking change detection | **Tight** (release feature) |
| `cmd/` (13 packages) | 13.2K | CLI commands | **Tight** (CLI layer) |
| `devkit/` | 1.2K | Code quality checks (coverage, complexity, vuln) | **Loose** (stdlib only) |
| `infra/` | 1.2K | Hetzner Cloud + CloudNS provider APIs | **Loose** (stdlib only) |
| `ansible/` | 3.7K | Pure Go Ansible playbook engine | **Loose** (go-log only) |
| `container/` | 1.3K | LinuxKit VM/hypervisor abstraction | **Loose** (go-io only) |
| `devops/` + `devops/sources/` | 1.7K | LinuxKit dev environment manager | **Medium** (container/) |
| `deploy/` | 0.4K | Coolify + Python deployment wrappers | **Medium** (release) |

## Decomposition Plan

### Phase 1: Extract to New Repos

#### 1.1 `devkit/` → merge into `core/lint`

**Why:** devkit's Finding, CoverageReport, ComplexityResult types align directly with lint's existing Finding type. The native AST complexity analyzer (`complexity.go`) is exactly lint's planned "layer 2" (AST-based detection). Coverage and vulncheck parsing are structured analysis — same domain.

**What moves:**
- `devkit/complexity.go` → `core/lint/pkg/lint/complexity.go` (AST-based cyclomatic complexity)
- `devkit/coverage.go` → `core/lint/pkg/lint/coverage.go` (coverage snapshot + regression tracking)
- `devkit/vulncheck.go` → `core/lint/pkg/lint/vulncheck.go` (govulncheck JSON parsing)
- `devkit/devkit.go` types → align with existing `lint.Finding`, add `TODO`, `SecretLeak` types
- `devkit/devkit.go` tool wrappers (Lint, RaceDetect, etc.) → `core/lint/pkg/lint/tools.go` (subprocess wrappers)

**What stays:** `cmd/qa/` stays in go-devops but changes imports from `devkit` to `core/lint`.

**Type alignment:**
```
devkit.Finding     → lint.Finding (already equivalent)
devkit.TODO        → new lint catalog rule (detection: regex)
devkit.SecretLeak  → new lint catalog rule (detection: regex)
devkit.ComplexFunc → lint.ComplexityResult (new type)
```

**New lint detection types:** `regex` (existing), `ast` (complexity), `tool` (subprocess wrappers)

#### 1.2 `infra/` → `core/go-infra`

**Why:** Pure stdlib, zero go-devops coupling. Generic infrastructure provider APIs (Hetzner Cloud, CloudNS). Reusable by any service that needs infrastructure management — not just devops.

**What moves:**
- `infra/hetzner.go` → `core/go-infra/hetzner/`
- `infra/cloudns.go` → `core/go-infra/cloudns/`
- `infra/api.go` → `core/go-infra/pkg/api/` (retry + rate-limit HTTP client)

**Dependencies:** stdlib only (net/http, encoding/json, math/rand). No framework deps.

**Consumers:** `cmd/prod/`, `cmd/monitor/` → change imports.

#### 1.3 `ansible/` → `core/go-ansible`

**Why:** Pure Go Ansible playbook engine (3.7K LOC). Only depends on go-log + golang.org/x/crypto for SSH. Generically useful — any Go service that needs to orchestrate remote servers. Currently used by deploy commands, but could power Lethean node provisioning, homelab automation, CI pipelines.

**What moves:**
- `ansible/ansible.go` → `core/go-ansible/ansible.go`
- `ansible/ssh.go` → `core/go-ansible/ssh.go`
- `ansible/playbook.go` → `core/go-ansible/playbook.go`
- `ansible/vars.go` → `core/go-ansible/vars.go`
- `ansible/handlers.go` → `core/go-ansible/handlers.go`

**Dependencies:** go-log, golang.org/x/crypto (SSH).

**Consumers:** `cmd/deploy/` → change imports.

#### 1.4 `container/` + `devops/` → `core/go-container`

**Why:** LinuxKit VM/container abstraction with QEMU (Linux) and Hyperkit (macOS) support. Strategic for Lethean network — TIM (Terminal Isolation Matrix) uses immutable LinuxKit images for node security. Distroless, read-only filesystems, single-binary containers.

**What moves:**
- `container/` → `core/go-container/` (VM manager, hypervisor abstraction)
- `devops/` → `core/go-container/devenv/` (dev environment manager, image sources)
- `devops/sources/` → `core/go-container/sources/` (CDN, GitHub image fetching)

**Dependencies:** go-io, go-config, Borg (for git-based image sources).

**Consumers:** `cmd/vm/` → change imports. Future: Lethean node runtime.

**Lethean context:** Network nodes run from immutable LinuxKit images to guarantee environment security. Read-only root filesystem, signed images, minimal attack surface. The container package provides the local hypervisor abstraction for running these images on dev machines and validators.

### Phase 2: Reorganise Within go-devops

After extraction, go-devops becomes a focused **build + release + deploy** tool:

```
go-devops/
├── build/                    # Project detection + cross-compilation
│   ├── builders/             # 8 language-specific builders
│   └── signing/              # Code signing (GPG, codesign, signtool)
├── release/                  # Release orchestration + changelog
│   └── publishers/           # 8 publisher implementations
├── sdk/                      # OpenAPI SDK generation
│   └── generators/           # 4 language generators
├── deploy/                   # Deployment strategies (Coolify)
└── cmd/                      # CLI commands
    ├── build/                # core build
    ├── release/              # core release
    ├── deploy/               # core deploy
    ├── dev/                  # core dev (multi-repo git workflow)
    ├── setup/                # core setup (GitHub org bootstrap)
    ├── prod/                 # core prod (imports go-infra)
    ├── qa/                   # core qa (imports core/lint)
    ├── ci/                   # core ci
    ├── docs/                 # core docs
    ├── sdk/                  # core sdk
    ├── vm/                   # core vm (imports go-container)
    ├── monitor/              # core monitor
    └── gitcmd/               # core git (aliases)
```

**Reduced dependencies:** go-devops drops stdlib-only packages from its tree. External deps like go-embed-python, kin-openapi, oasdiff stay (they're build/release/SDK specific).

### Phase 3: Update Consumers

| Consumer | Current Import | New Import |
|----------|---------------|------------|
| `cmd/qa/` | `go-devops/devkit` | `core/lint/pkg/lint` |
| `cmd/prod/` | `go-devops/infra` | `core/go-infra` |
| `cmd/deploy/` | `go-devops/ansible` | `core/go-ansible` |
| `cmd/vm/` | `go-devops/container` | `core/go-container` |
| `cmd/vm/` | `go-devops/devops` | `core/go-container/devenv` |
| `core/cli/cmd/gocmd/cmd_qa.go` | `go-devops/cmd/qa` | `core/lint/pkg/lint` |

## Dependency Graph (After)

```
core/lint (standalone, zero deps)
    ^
    |
core/go-infra (standalone, stdlib only)
    ^
    |
core/go-ansible (go-log, x/crypto)
    ^
    |
core/go-container (go-io, go-config, Borg)
    ^
    |
go-devops (build + release + deploy)
    imports: core/lint, go-infra, go-ansible, go-container
    imports: go-io, go-log, go-scm, go-i18n, cli
    imports: kin-openapi, oasdiff, go-embed-python (build/release specific)
```

## Execution Order

1. **devkit → core/lint** (smallest, highest value — already have lint repo)
2. **infra → go-infra** (zero deps, clean extract)
3. **ansible → go-ansible** (go-log only, clean extract)
4. **container + devops → go-container** (slightly more deps, needs Borg)
5. **Update go-devops imports** (remove extracted packages, point to new repos)
6. **Update core/cli imports** (cmd_qa.go uses devkit)

## Vanity Import Server

`cmd/vanity-import/` is a standalone HTTP service for Go vanity imports (dappco.re → forge repos). Extract to its own repo or binary. Not blocking — low priority.

## Risk Mitigation

- **Backward compatibility**: Keep type aliases in go-devops during transition (`type Finding = lint.Finding`)
- **go.work**: All repos in workspace, so local development works immediately
- **Testing**: Each extraction verified by `go test ./...` in both source and destination
- **Incremental**: One extraction at a time, commit + push + tag before next
