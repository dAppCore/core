package marketplace

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"forge.lthn.ai/core/go/pkg/io"
	"forge.lthn.ai/core/go/pkg/manifest"
	"forge.lthn.ai/core/go/pkg/store"
)

const storeGroup = "_modules"

// Installer handles module installation from Git repos.
type Installer struct {
	modulesDir string
	store      *store.Store
}

// NewInstaller creates a new module installer.
func NewInstaller(modulesDir string, st *store.Store) *Installer {
	return &Installer{
		modulesDir: modulesDir,
		store:      st,
	}
}

// InstalledModule holds stored metadata about an installed module.
type InstalledModule struct {
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Repo        string               `json:"repo"`
	EntryPoint  string               `json:"entry_point"`
	Permissions manifest.Permissions `json:"permissions"`
	InstalledAt string               `json:"installed_at"`
}

// Install clones a module repo, verifies its manifest signature, and registers it.
func (i *Installer) Install(ctx context.Context, mod Module) error {
	// Check if already installed
	if _, err := i.store.Get(storeGroup, mod.Code); err == nil {
		return fmt.Errorf("marketplace: module %q already installed", mod.Code)
	}

	dest := filepath.Join(i.modulesDir, mod.Code)
	if err := os.MkdirAll(i.modulesDir, 0755); err != nil {
		return fmt.Errorf("marketplace: mkdir: %w", err)
	}
	if err := gitClone(ctx, mod.Repo, dest); err != nil {
		return fmt.Errorf("marketplace: clone %s: %w", mod.Repo, err)
	}

	// On any error after clone, clean up the directory
	cleanup := true
	defer func() {
		if cleanup {
			os.RemoveAll(dest)
		}
	}()

	medium, err := io.NewSandboxed(dest)
	if err != nil {
		return fmt.Errorf("marketplace: medium: %w", err)
	}

	m, err := loadManifest(medium, mod.SignKey)
	if err != nil {
		return err
	}

	entryPoint := filepath.Join(dest, "main.ts")
	installed := InstalledModule{
		Code:        mod.Code,
		Name:        m.Name,
		Version:     m.Version,
		Repo:        mod.Repo,
		EntryPoint:  entryPoint,
		Permissions: m.Permissions,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(installed)
	if err != nil {
		return fmt.Errorf("marketplace: marshal: %w", err)
	}

	if err := i.store.Set(storeGroup, mod.Code, string(data)); err != nil {
		return fmt.Errorf("marketplace: store: %w", err)
	}

	cleanup = false
	return nil
}

// Remove uninstalls a module by deleting its files and store entry.
func (i *Installer) Remove(code string) error {
	if _, err := i.store.Get(storeGroup, code); err != nil {
		return fmt.Errorf("marketplace: module %q not installed", code)
	}

	dest := filepath.Join(i.modulesDir, code)
	os.RemoveAll(dest)

	return i.store.Delete(storeGroup, code)
}

// Update pulls latest changes and re-verifies the manifest.
func (i *Installer) Update(ctx context.Context, code string) error {
	raw, err := i.store.Get(storeGroup, code)
	if err != nil {
		return fmt.Errorf("marketplace: module %q not installed", code)
	}

	var installed InstalledModule
	if err := json.Unmarshal([]byte(raw), &installed); err != nil {
		return fmt.Errorf("marketplace: unmarshal: %w", err)
	}

	dest := filepath.Join(i.modulesDir, code)

	cmd := exec.CommandContext(ctx, "git", "-C", dest, "pull", "--ff-only")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("marketplace: pull: %s: %w", strings.TrimSpace(string(output)), err)
	}

	// Reload manifest
	medium, mErr := io.NewSandboxed(dest)
	if mErr != nil {
		return fmt.Errorf("marketplace: medium: %w", mErr)
	}
	m, mErr := manifest.Load(medium, ".")
	if mErr != nil {
		return fmt.Errorf("marketplace: reload manifest: %w", mErr)
	}

	// Update stored metadata
	installed.Name = m.Name
	installed.Version = m.Version
	installed.Permissions = m.Permissions

	data, err := json.Marshal(installed)
	if err != nil {
		return fmt.Errorf("marketplace: marshal: %w", err)
	}

	return i.store.Set(storeGroup, code, string(data))
}

// Installed returns all installed module metadata.
func (i *Installer) Installed() ([]InstalledModule, error) {
	all, err := i.store.GetAll(storeGroup)
	if err != nil {
		return nil, fmt.Errorf("marketplace: list: %w", err)
	}

	var modules []InstalledModule
	for _, raw := range all {
		var m InstalledModule
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			continue
		}
		modules = append(modules, m)
	}
	return modules, nil
}

// loadManifest loads and optionally verifies a module manifest.
func loadManifest(medium io.Medium, signKey string) (*manifest.Manifest, error) {
	if signKey != "" {
		pubBytes, err := hex.DecodeString(signKey)
		if err != nil {
			return nil, fmt.Errorf("marketplace: decode sign key: %w", err)
		}
		return manifest.LoadVerified(medium, ".", pubBytes)
	}
	return manifest.Load(medium, ".")
}

// gitClone clones a repository with --depth=1.
func gitClone(ctx context.Context, repo, dest string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", repo, dest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}
