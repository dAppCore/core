//go:build php

// php.go imports packages for the PHP-only binary.
//
// Build with: go build -tags php
//
// This variant includes only PHP/Laravel development tools:
//   - php: Laravel/Composer development tools
//   - doctor: Environment verification
//
// Use this for PHP-focused workflows without other tooling.

package variants

import (
	// Commands via self-registration
	_ "forge.lthn.ai/core/cli/internal/cmd/doctor"
	_ "forge.lthn.ai/core/cli/internal/cmd/php"
)
