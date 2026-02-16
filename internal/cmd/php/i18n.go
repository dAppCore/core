// Package php provides PHP/Laravel development tools.
package php

import (
	"embed"

	"forge.lthn.ai/core/cli/pkg/i18n"
)

//go:embed locales/*.json
var localeFS embed.FS

func init() {
	// Register PHP translations with the i18n system
	i18n.RegisterLocales(localeFS, "locales")
}
