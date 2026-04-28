package core_test

import (
	. "dappco.re/go"
)

// ExampleJoinPath joins path segments through `JoinPath` for workspace path handling. Path
// joins, cleanup, globbing, and extension changes use core wrappers.
func ExampleJoinPath() {
	Println(JoinPath("deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

// ExamplePath constructs a clean path through `Path` for workspace path handling. Path
// joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePath() {
	Println(PathBase(Path("Code", "core")))
	Println(PathBase(Path("/tmp", "workspace")))
	// Output:
	// core
	// workspace
}

// ExamplePathBase reads a base name through `PathBase` for workspace path handling. Path
// joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathBase() {
	Println(PathBase("/srv/workspaces/alpha"))
	// Output: alpha
}

// ExamplePathDir reads a directory name through `PathDir` for workspace path handling.
// Path joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathDir() {
	Println(PathDir("/srv/workspaces/alpha"))
	// Output: /srv/workspaces
}

// ExamplePathExt reads an extension through `PathExt` for workspace path handling. Path
// joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathExt() {
	Println(PathExt("report.pdf"))
	// Output: .pdf
}

// ExamplePathIsAbs checks whether a path is absolute through `PathIsAbs` for workspace
// path handling. Path joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathIsAbs() {
	Println(PathIsAbs("/tmp/workspace"))
	Println(PathIsAbs("workspace"))
	// Output:
	// true
	// false
}

// ExampleCleanPath normalises a path through `CleanPath` for workspace path handling. Path
// joins, cleanup, globbing, and extension changes use core wrappers.
func ExampleCleanPath() {
	Println(CleanPath("/tmp//file", "/"))
	Println(CleanPath("a/b/../c", "/"))
	Println(CleanPath("deploy/to/homelab", "/"))
	// Output:
	// /tmp/file
	// a/c
	// deploy/to/homelab
}

// ExamplePathGlob expands a glob through `PathGlob` for workspace path handling. Path
// joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathGlob() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-path-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "a.txt"), "a")
	fs.Write(Path(dir, "b.txt"), "b")

	matches := PathGlob(Path(dir, "*.txt"))
	for i, match := range matches {
		matches[i] = PathBase(match)
	}
	SliceSort(matches)
	Println(matches)
	// Output: [a.txt b.txt]
}

// ExamplePathRel calculates a relative path through `PathRel` for workspace path handling.
// Path joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathRel() {
	r := PathRel("/srv/app", "/srv/app/config/settings.yaml")
	Println(r.Value)
	// Output: config/settings.yaml
}

// ExamplePathAbs calculates an absolute path through `PathAbs` for workspace path
// handling. Path joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathAbs() {
	r := PathAbs(".")
	Println(r.OK)
	Println(PathIsAbs(r.Value.(string)))
	// Output:
	// true
	// true
}

// ExamplePathChangeExt rewrites a file extension through `PathChangeExt` for workspace
// path handling. Path joins, cleanup, globbing, and extension changes use core wrappers.
func ExamplePathChangeExt() {
	Println(PathChangeExt("config.json", "yaml"))
	Println(PathChangeExt("README", ".md"))
	// Output:
	// config.yaml
	// README.md
}
