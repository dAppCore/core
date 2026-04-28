package core_test

import (
	. "dappco.re/go"
)

func ExampleJoinPath() {
	Println(JoinPath("deploy", "to", "homelab"))
	// Output: deploy/to/homelab
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

func ExampleCleanPath() {
	Println(CleanPath("/tmp//file", "/"))
	Println(CleanPath("a/b/../c", "/"))
	Println(CleanPath("deploy/to/homelab", "/"))
	// Output:
	// /tmp/file
	// a/c
	// deploy/to/homelab
}
