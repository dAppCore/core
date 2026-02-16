package dev

import (
	"os"
	"path/filepath"
	"strings"

	"forge.lthn.ai/core/cli/internal/cmd/workspace"
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/i18n"
	"forge.lthn.ai/core/cli/pkg/io"
	"forge.lthn.ai/core/cli/pkg/repos"
)

// loadRegistryWithConfig loads the registry and applies workspace configuration.
func loadRegistryWithConfig(registryPath string) (*repos.Registry, string, error) {
	var reg *repos.Registry
	var err error
	var registryDir string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(io.Local, registryPath)
		if err != nil {
			return nil, "", cli.Wrap(err, "failed to load registry")
		}
		cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
		registryDir = filepath.Dir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry(io.Local)
		if err == nil {
			reg, err = repos.LoadRegistry(io.Local, registryPath)
			if err != nil {
				return nil, "", cli.Wrap(err, "failed to load registry")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
			registryDir = filepath.Dir(registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(io.Local, cwd)
			if err != nil {
				return nil, "", cli.Wrap(err, "failed to scan directory")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
			registryDir = cwd
		}
	}
	// Load workspace config to respect packages_dir (only if config exists)
	if wsConfig, err := workspace.LoadConfig(registryDir); err == nil && wsConfig != nil {
		if wsConfig.PackagesDir != "" {
			pkgDir := wsConfig.PackagesDir
			// Expand ~
			if strings.HasPrefix(pkgDir, "~/") {
				home, _ := os.UserHomeDir()
				pkgDir = filepath.Join(home, pkgDir[2:])
			}
			if !filepath.IsAbs(pkgDir) {
				pkgDir = filepath.Join(registryDir, pkgDir)
			}

			// Update repo paths
			for _, repo := range reg.Repos {
				repo.Path = filepath.Join(pkgDir, repo.Name)
			}
		}
	}

	return reg, registryDir, nil
}
