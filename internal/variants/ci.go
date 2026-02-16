//go:build ci

// ci.go imports packages for the minimal CI/release binary.
//
// Build with: go build -tags ci
//
// This variant includes only commands needed for CI pipelines:
//   - build: Cross-platform compilation
//   - ci: Release publishing
//   - sdk: API compatibility checks
//   - doctor: Environment verification
//
// Use this build to reduce binary size and attack surface in production.

package variants

import (
	// Commands via self-registration
	_ "forge.lthn.ai/core/cli/internal/cmd/ci"
	_ "forge.lthn.ai/core/cli/internal/cmd/doctor"
	_ "forge.lthn.ai/core/cli/internal/cmd/sdk"
	_ "forge.lthn.ai/core/cli/pkg/build/buildcmd"
)
