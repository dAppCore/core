// Package builders provides build implementations for different project types.
package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"forge.lthn.ai/core/go/pkg/build"
	"forge.lthn.ai/core/go/pkg/io"
)

// CPPBuilder implements the Builder interface for C++ projects using CMake + Conan.
// It wraps the Makefile-based build system from the .core/build submodule.
type CPPBuilder struct{}

// NewCPPBuilder creates a new CPPBuilder instance.
func NewCPPBuilder() *CPPBuilder {
	return &CPPBuilder{}
}

// Name returns the builder's identifier.
func (b *CPPBuilder) Name() string {
	return "cpp"
}

// Detect checks if this builder can handle the project in the given directory.
func (b *CPPBuilder) Detect(fs io.Medium, dir string) (bool, error) {
	return build.IsCPPProject(fs, dir), nil
}

// Build compiles the C++ project using Make targets.
// The build flow is: make configure → make build → make package.
// Cross-compilation is handled via Conan profiles specified in .core/build.yaml.
func (b *CPPBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	if cfg == nil {
		return nil, fmt.Errorf("builders.CPPBuilder.Build: config is nil")
	}

	// Validate make is available
	if err := b.validateMake(); err != nil {
		return nil, err
	}

	// For C++ projects, the Makefile handles everything.
	// We don't iterate per-target like Go — the Makefile's configure + build
	// produces binaries for the host platform, and cross-compilation uses
	// named Conan profiles (e.g., make gcc-linux-armv8).
	if len(targets) == 0 {
		// Default to host platform
		targets = []build.Target{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
	}

	var artifacts []build.Artifact

	for _, target := range targets {
		built, err := b.buildTarget(ctx, cfg, target)
		if err != nil {
			return artifacts, fmt.Errorf("builders.CPPBuilder.Build: %w", err)
		}
		artifacts = append(artifacts, built...)
	}

	return artifacts, nil
}

// buildTarget compiles for a single target platform.
func (b *CPPBuilder) buildTarget(ctx context.Context, cfg *build.Config, target build.Target) ([]build.Artifact, error) {
	// Determine if this is a cross-compile or host build
	isHostBuild := target.OS == runtime.GOOS && target.Arch == runtime.GOARCH

	if isHostBuild {
		return b.buildHost(ctx, cfg, target)
	}

	return b.buildCross(ctx, cfg, target)
}

// buildHost runs the standard make configure → make build → make package flow.
func (b *CPPBuilder) buildHost(ctx context.Context, cfg *build.Config, target build.Target) ([]build.Artifact, error) {
	fmt.Printf("Building C++ project for %s/%s (host)\n", target.OS, target.Arch)

	// Step 1: Configure (runs conan install + cmake configure)
	if err := b.runMake(ctx, cfg.ProjectDir, "configure"); err != nil {
		return nil, fmt.Errorf("configure failed: %w", err)
	}

	// Step 2: Build
	if err := b.runMake(ctx, cfg.ProjectDir, "build"); err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// Step 3: Package
	if err := b.runMake(ctx, cfg.ProjectDir, "package"); err != nil {
		return nil, fmt.Errorf("package failed: %w", err)
	}

	// Discover artifacts from build/packages/
	return b.findArtifacts(cfg.FS, cfg.ProjectDir, target)
}

// buildCross runs a cross-compilation using a Conan profile name.
// The Makefile supports profile targets like: make gcc-linux-armv8
func (b *CPPBuilder) buildCross(ctx context.Context, cfg *build.Config, target build.Target) ([]build.Artifact, error) {
	// Map target to a Conan profile name
	profile := b.targetToProfile(target)
	if profile == "" {
		return nil, fmt.Errorf("no Conan profile mapped for target %s/%s", target.OS, target.Arch)
	}

	fmt.Printf("Building C++ project for %s/%s (cross: %s)\n", target.OS, target.Arch, profile)

	// The Makefile exposes each profile as a top-level target
	if err := b.runMake(ctx, cfg.ProjectDir, profile); err != nil {
		return nil, fmt.Errorf("cross-compile for %s failed: %w", profile, err)
	}

	return b.findArtifacts(cfg.FS, cfg.ProjectDir, target)
}

// runMake executes a make target in the project directory.
func (b *CPPBuilder) runMake(ctx context.Context, projectDir string, target string) error {
	cmd := exec.CommandContext(ctx, "make", target)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make %s: %w", target, err)
	}
	return nil
}

// findArtifacts searches for built packages in build/packages/.
func (b *CPPBuilder) findArtifacts(fs io.Medium, projectDir string, target build.Target) ([]build.Artifact, error) {
	packagesDir := filepath.Join(projectDir, "build", "packages")

	if !fs.IsDir(packagesDir) {
		// Fall back to searching build/release/src/ for raw binaries
		return b.findBinaries(fs, projectDir, target)
	}

	entries, err := fs.List(packagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages directory: %w", err)
	}

	var artifacts []build.Artifact
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip checksum files and hidden files
		if strings.HasSuffix(name, ".sha256") || strings.HasPrefix(name, ".") {
			continue
		}

		artifacts = append(artifacts, build.Artifact{
			Path: filepath.Join(packagesDir, name),
			OS:   target.OS,
			Arch: target.Arch,
		})
	}

	return artifacts, nil
}

// findBinaries searches for compiled binaries in build/release/src/.
func (b *CPPBuilder) findBinaries(fs io.Medium, projectDir string, target build.Target) ([]build.Artifact, error) {
	binDir := filepath.Join(projectDir, "build", "release", "src")

	if !fs.IsDir(binDir) {
		return nil, fmt.Errorf("no build output found in %s", binDir)
	}

	entries, err := fs.List(binDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list build directory: %w", err)
	}

	var artifacts []build.Artifact
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip non-executable files (libraries, cmake files, etc.)
		if strings.HasSuffix(name, ".a") || strings.HasSuffix(name, ".o") ||
			strings.HasSuffix(name, ".cmake") || strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(binDir, name)

		// On Unix, check if file is executable
		if target.OS != "windows" {
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			if info.Mode()&0111 == 0 {
				continue
			}
		}

		artifacts = append(artifacts, build.Artifact{
			Path: fullPath,
			OS:   target.OS,
			Arch: target.Arch,
		})
	}

	return artifacts, nil
}

// targetToProfile maps a build target to a Conan cross-compilation profile name.
// Profile names match those in .core/build/cmake/profiles/.
func (b *CPPBuilder) targetToProfile(target build.Target) string {
	key := target.OS + "/" + target.Arch
	profiles := map[string]string{
		"linux/amd64":    "gcc-linux-x86_64",
		"linux/x86_64":   "gcc-linux-x86_64",
		"linux/arm64":    "gcc-linux-armv8",
		"linux/armv8":    "gcc-linux-armv8",
		"darwin/arm64":   "apple-clang-armv8",
		"darwin/armv8":   "apple-clang-armv8",
		"darwin/amd64":   "apple-clang-x86_64",
		"darwin/x86_64":  "apple-clang-x86_64",
		"windows/amd64":  "msvc-194-x86_64",
		"windows/x86_64": "msvc-194-x86_64",
	}

	return profiles[key]
}

// validateMake checks if make is available.
func (b *CPPBuilder) validateMake() error {
	if _, err := exec.LookPath("make"); err != nil {
		return fmt.Errorf("cpp: make not found. Install build-essential (Linux) or Xcode Command Line Tools (macOS)")
	}
	return nil
}

// Ensure CPPBuilder implements the Builder interface.
var _ build.Builder = (*CPPBuilder)(nil)
