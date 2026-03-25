package core_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "dappco.re/go/core"
)

func ExampleFs_WriteAtomic() {
	dir, _ := os.MkdirTemp("", "example")
	defer os.RemoveAll(dir)

	f := (&Fs{}).New("/")
	path := filepath.Join(dir, "status.json")
	f.WriteAtomic(path, `{"status":"completed"}`)

	r := f.Read(path)
	fmt.Println(r.Value)
	// Output: {"status":"completed"}
}

func ExampleFs_NewUnrestricted() {
	dir, _ := os.MkdirTemp("", "example")
	defer os.RemoveAll(dir)

	// Write outside sandbox
	outside := filepath.Join(dir, "outside.txt")
	os.WriteFile(outside, []byte("hello"), 0644)

	sandbox := (&Fs{}).New(filepath.Join(dir, "sandbox"))
	unrestricted := sandbox.NewUnrestricted()

	r := unrestricted.Read(outside)
	fmt.Println(r.Value)
	// Output: hello
}

func ExampleFs_Root() {
	f := (&Fs{}).New("/srv/workspaces")
	fmt.Println(f.Root())
	// Output: /srv/workspaces
}
