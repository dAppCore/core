// SPDX-License-Identifier: EUPL-1.2

// Application identity for the Core framework.
// Based on leaanthony/sail — Name, Filename, Path.

package core

import (
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
