// SPDX-License-Identifier: EUPL-1.2

// Application identity for the Core framework.
// Based on leaanthony/sail — Name, Filename, Path.

package core

import (
	"os"
	"os/exec"
	"path/filepath"
)

// App holds the application identity and optional GUI runtime.
type App struct {
	// Name is the human-readable application name (e.g., "Core CLI").
	Name string

	// Version is the application version string (e.g., "1.2.3").
	Version string

	// Description is a short description of the application.
	Description string

	// Filename is the executable filename (e.g., "core").
	Filename string

	// Path is the absolute path to the executable.
	Path string

	// Runtime is the GUI runtime (e.g., Wails App).
	// Nil for CLI-only applications.
	Runtime any
}

// NewApp creates a App with the given identity.
// Filename and Path are auto-detected from the running binary.
func NewApp(name, description, version string) *App {
	app := &App{
		Name:        name,
		Version:     version,
		Description: description,
	}

	// Auto-detect executable identity
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			app.Path = abs
			app.Filename = filepath.Base(abs)
		}
	}

	return app
}

// Find locates a program on PATH and returns a App for it.
// Returns nil if not found.
func Find(filename, name string) *App {
	path, err := exec.LookPath(filename)
	if err != nil {
		return nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	return &App{
		Name:     name,
		Filename: filename,
		Path:     abs,
	}
}
