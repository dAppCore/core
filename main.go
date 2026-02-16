package main

import (
	"forge.lthn.ai/core/cli/pkg/cli"

	// Build variants import commands via self-registration.
	// See internal/variants/ for available variants: full, ci, php, minimal.
	_ "forge.lthn.ai/core/cli/internal/variants"
)

func main() {
	cli.Main()
}
