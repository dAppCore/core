package ci

import (
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/i18n"
	"forge.lthn.ai/core/cli/pkg/release"
)

// runCIReleaseVersion shows the determined version.
func runCIReleaseVersion() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return cli.WrapVerb(err, "get", "working directory")
	}

	version, err := release.DetermineVersion(projectDir)
	if err != nil {
		return cli.WrapVerb(err, "determine", "version")
	}

	cli.Print("%s %s\n", i18n.Label("version"), releaseValueStyle.Render(version))
	return nil
}
