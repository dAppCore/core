package core_test

import (
	. "dappco.re/go"
)

func ExampleJoinPath() {
	Println(JoinPath("deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

func ExamplePath() {
	Println(PathBase(Path("Code", "core")))
	Println(PathBase(Path("/tmp", "workspace")))
	// Output:
	// core
	// workspace
}

func ExamplePathBase() {
	Println(PathBase("/srv/workspaces/alpha"))
	// Output: alpha
}

func ExamplePathDir() {
	Println(PathDir("/srv/workspaces/alpha"))
	// Output: /srv/workspaces
}

func ExamplePathExt() {
	Println(PathExt("report.pdf"))
	// Output: .pdf
}

func ExamplePathIsAbs() {
	Println(PathIsAbs("/tmp/workspace"))
	Println(PathIsAbs("workspace"))
	// Output:
	// true
	// false
}

func ExampleCleanPath() {
	Println(CleanPath("/tmp//file", "/"))
	Println(CleanPath("a/b/../c", "/"))
	Println(CleanPath("deploy/to/homelab", "/"))
	// Output:
	// /tmp/file
	// a/c
	// deploy/to/homelab
}

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

func ExamplePathRel() {
	r := PathRel("/srv/app", "/srv/app/config/settings.yaml")
	Println(r.Value)
	// Output: config/settings.yaml
}

func ExamplePathAbs() {
	r := PathAbs(".")
	Println(r.OK)
	Println(PathIsAbs(r.Value.(string)))
	// Output:
	// true
	// true
}

func ExamplePathChangeExt() {
	Println(PathChangeExt("config.json", "yaml"))
	Println(PathChangeExt("README", ".md"))
	// Output:
	// config.yaml
	// README.md
}
