# core/mcp Extraction

**Goal:** Consolidate MCP code into `core/mcp` — Go MCP server from go-ai + PHP MCP from php-mcp. Produces `core-mcp` binary.

**Pattern:** Polyglot repo like `core/agent` — Go at root, PHP in `src/php/`, composer.json at root as `lthn/mcp`.

---

### Task 1: Create core/mcp repo on forge, clone locally

- Create `core/mcp` repo on forge via API
- Clone to `/Users/snider/Code/core/mcp/`
- Add to `~/Code/go.work`

### Task 2: Move Go MCP package from go-ai

**Move:** `go-ai/mcp/` → `core/mcp/pkg/mcp/`

All files: `mcp.go`, `registry.go`, `subsystem.go`, `bridge.go`, transports (stdio, tcp, unix), all `tools_*.go`, all `*_test.go`, `brain/`, `ide/`

**Move:** `go-ai/cmd/mcpcmd/` → `core/mcp/cmd/mcpcmd/`
**Move:** `go-ai/cmd/brain-seed/` → `core/mcp/cmd/brain-seed/`

Create `go.mod` as `forge.lthn.ai/core/mcp`. Deps from go-ai: go-sdk, gorilla/websocket, go-ml, go-rag, go-inference, go-process, go-i18n, gin.

Find-replace `forge.lthn.ai/core/go-ai/mcp` → `forge.lthn.ai/core/mcp/pkg/mcp` in all moved files.

### Task 3: Create cmd/core-mcp/main.go

```go
package main

import (
    "forge.lthn.ai/core/cli/pkg/cli"
    mcpcmd "forge.lthn.ai/core/mcp/cmd/mcpcmd"
)

func main() {
    cli.Main(
        cli.WithCommands("mcp", mcpcmd.AddMCPCommands),
    )
}
```

Add `.core/build.yaml`:
```yaml
project:
  name: core-mcp
  binary: core-mcp
```

### Task 4: Move PHP from php-mcp into src/php/

Copy all PHP source from `/Users/snider/Code/core/php-mcp/` → `core/mcp/src/php/`

Create `composer.json` at root:
```json
{
    "name": "lthn/mcp",
    "description": "Model Context Protocol server for Laravel + standalone Go binary",
    "license": "EUPL-1.2",
    "require": { "php": "^8.2", "lthn/php": "*" },
    "autoload": { "psr-4": { "Core\\Mcp\\": "src/php/" } },
    "replace": { "core/php-mcp": "self.version", "lthn/php-mcp": "self.version" }
}
```

Add `.gitattributes` to exclude Go from composer dist.

### Task 5: Update consumers

- `core/agent` — change import `go-ai/mcp` → `core/mcp/pkg/mcp` in `pkg/loop/tools_mcp.go`
- `core/cli` — change `go-ai/cmd/mcpcmd` → `core/mcp/cmd/mcpcmd` import
- App `composer.json` — change `core/php-mcp` → `lthn/mcp`, VCS url → `core/mcp.git`

### Task 6: Clean up go-ai

- Delete `go-ai/mcp/` directory
- Delete `go-ai/cmd/mcpcmd/` and `go-ai/cmd/brain-seed/`
- Remove unused deps from go-ai's go.mod (go-sdk, websocket, etc.)
- go-ai keeps: `ai/`, `cmd/{daemon,embed-bench,lab,metrics,rag,security}`

### Task 7: Register on Packagist, verify

- Submit `core/mcp` to Packagist as `lthn/mcp`
- `go build ./cmd/core-mcp` — verify binary builds
- `go test ./...` — verify tests pass
- Archive `core/php-mcp` on forge

---

**Repos affected:** core/mcp (new), go-ai (shrinks), core/agent (import update), core/cli (import update), php-mcp (archived), app composer.json
