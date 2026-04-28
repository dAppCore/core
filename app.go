// SPDX-License-Identifier: EUPL-1.2

// Application identity for the Core framework.

package core

import (
	"path/filepath"
)

// App holds the application identity and optional GUI runtime.
//
//	app := core.App{}.New(core.NewOptions(
//	    core.Option{Key: "name", Value: "Core CLI"},
//	    core.Option{Key: "version", Value: "1.0.0"},
//	))
type App struct {
	Name        string
	Version     string
	Description string
	Filename    string
	Path        string
	Runtime     any // GUI runtime (e.g., Wails App). Nil for CLI-only.
}

// New creates an App from Options.
//
//	app := core.App{}.New(core.NewOptions(
//	    core.Option{Key: "name", Value: "myapp"},
//	    core.Option{Key: "version", Value: "1.0.0"},
//	))
func (a App) New(opts Options) App {
	if name := opts.String("name"); name != "" {
		a.Name = name
	}
	if version := opts.String("version"); version != "" {
		a.Version = version
	}
	if desc := opts.String("description"); desc != "" {
		a.Description = desc
	}
	if filename := opts.String("filename"); filename != "" {
		a.Filename = filename
	}
	return a
}

// Find locates a program on PATH and returns a Result containing the App.
// Uses core.Stat to search PATH directories — no os/exec dependency.
//
//	r := core.App{}.Find("node", "Node.js")
//	if r.OK { app := r.Value.(*App) }
func (a App) Find(filename, name string) Result {
	// If filename contains a separator, check it directly
	if Contains(filename, string(PathSeparator)) {
		abs := PathAbs(filename)
		if !abs.OK {
			return abs
		}
		if isExecutable(abs.Value.(string)) {
			return Result{&App{Name: name, Filename: filename, Path: abs.Value.(string)}, true}
		}
		return Result{E("app.Find", Concat(filename, " not found"), nil), false}
	}

	// Search PATH
	pathEnv := Env("PATH")
	if pathEnv == "" {
		return Result{E("app.Find", "PATH is empty", nil), false}
	}
	for _, dir := range Split(pathEnv, string(PathListSeparator)) {
		candidate := filepath.Join(dir, filename)
		if isExecutable(candidate) {
			abs := PathAbs(candidate)
			if !abs.OK {
				continue
			}
			return Result{&App{Name: name, Filename: filename, Path: abs.Value.(string)}, true}
		}
	}
	return Result{E("app.Find", Concat(filename, " not found on PATH"), nil), false}
}

// isExecutable checks if a path exists and is executable.
func isExecutable(path string) bool {
	r := Stat(path)
	if !r.OK {
		return false
	}
	info := r.Value.(interface {
		IsDir() bool
		Mode() FileMode
	})
	// Regular file with at least one execute bit
	return !info.IsDir() && info.Mode()&0111 != 0
}
