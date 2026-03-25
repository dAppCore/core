package core_test

import (
	"fmt"

	. "dappco.re/go/core"
)

func ExampleFs_WriteAtomic() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("example")
	defer f.DeleteAll(dir)

	path := Path(dir, "status.json")
	f.WriteAtomic(path, `{"status":"completed"}`)

	r := f.Read(path)
	fmt.Println(r.Value)
	// Output: {"status":"completed"}
}

func ExampleFs_NewUnrestricted() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("example")
	defer f.DeleteAll(dir)

	// Write outside sandbox using Core's Fs
	outside := Path(dir, "outside.txt")
	f.Write(outside, "hello")

	sandbox := (&Fs{}).New(Path(dir, "sandbox"))
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
