package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleJoinPath() {
	fmt.Println(JoinPath("deploy", "to", "homelab"))
	// Output: deploy/to/homelab
}

func ExamplePathBase() {
	fmt.Println(PathBase("/srv/workspaces/alpha"))
	// Output: alpha
}

func ExamplePathDir() {
	fmt.Println(PathDir("/srv/workspaces/alpha"))
	// Output: /srv/workspaces
}

func ExamplePathExt() {
	fmt.Println(PathExt("report.pdf"))
	// Output: .pdf
}

