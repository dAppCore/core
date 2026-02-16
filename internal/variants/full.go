//go:build !ci && !php && !minimal

// full.go imports all packages for the full development binary.
//
// Build with: go build (default)
//
// This is the default build variant with all development tools:
//   - dev: Multi-repo git workflows (commit, push, pull, sync)
//   - ai: AI agent task management + RAG + metrics
//   - go: Go module and build tools
//   - php: Laravel/Composer development tools
//   - build: Cross-platform compilation
//   - ci: Release publishing
//   - sdk: API compatibility checks
//   - pkg: Package management
//   - vm: LinuxKit VM management
//   - docs: Documentation generation
//   - setup: Repository cloning and setup
//   - doctor: Environment health checks
//   - test: Test runner with coverage
//   - qa: Quality assurance workflows
//   - monitor: Security monitoring aggregation
//   - forge: Forgejo instance management
//   - prod: Production infrastructure
//   - mcp: MCP server management
//   - daemon: Background service daemon
//   - session: Session management

package variants

import (
	// Commands via self-registration
	_ "forge.lthn.ai/core/cli/internal/cmd/ai"
	_ "forge.lthn.ai/core/cli/internal/cmd/ci"
	_ "forge.lthn.ai/core/cli/internal/cmd/collect"
	_ "forge.lthn.ai/core/cli/internal/cmd/config"
	_ "forge.lthn.ai/core/cli/internal/cmd/crypt"
	_ "forge.lthn.ai/core/cli/internal/cmd/daemon"
	_ "forge.lthn.ai/core/cli/internal/cmd/deploy"
	_ "forge.lthn.ai/core/cli/internal/cmd/dev"
	_ "forge.lthn.ai/core/cli/internal/cmd/docs"
	_ "forge.lthn.ai/core/cli/internal/cmd/doctor"
	_ "forge.lthn.ai/core/cli/internal/cmd/forge"
	_ "forge.lthn.ai/core/cli/internal/cmd/gitcmd"
	_ "forge.lthn.ai/core/cli/internal/cmd/go"
	_ "forge.lthn.ai/core/cli/internal/cmd/help"
	_ "forge.lthn.ai/core/cli/internal/cmd/mcpcmd"
	_ "forge.lthn.ai/core/cli/internal/cmd/ml"
	_ "forge.lthn.ai/core/cli/internal/cmd/monitor"
	_ "forge.lthn.ai/core/cli/internal/cmd/php"
	_ "forge.lthn.ai/core/cli/internal/cmd/pkgcmd"
	_ "forge.lthn.ai/core/cli/internal/cmd/plugin"
	_ "forge.lthn.ai/core/cli/internal/cmd/prod"
	_ "forge.lthn.ai/core/cli/internal/cmd/qa"
	_ "forge.lthn.ai/core/cli/internal/cmd/sdk"
	_ "forge.lthn.ai/core/cli/internal/cmd/security"
	_ "forge.lthn.ai/core/cli/internal/cmd/session"
	_ "forge.lthn.ai/core/cli/internal/cmd/setup"
	_ "forge.lthn.ai/core/cli/internal/cmd/test"
	_ "forge.lthn.ai/core/cli/internal/cmd/updater"
	_ "forge.lthn.ai/core/cli/internal/cmd/vm"
	_ "forge.lthn.ai/core/cli/internal/cmd/workspace"
	_ "forge.lthn.ai/core/cli/pkg/build/buildcmd"
)
