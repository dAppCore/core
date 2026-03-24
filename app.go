// SPDX-License-Identifier: EUPL-1.2

// Application identity for the Core framework.

package core

import (
	"os/exec"
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
//
//	r := core.App{}.Find("node", "Node.js")
//	if r.OK { app := r.Value.(*App) }
func (a App) Find(filename, name string) Result {
	path, err := exec.LookPath(filename)
	if err != nil {
		return Result{err, false}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return Result{err, false}
	}
	return Result{&App{
		Name:     name,
		Filename: filename,
		Path:     abs,
	}, true}
}
