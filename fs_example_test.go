package core_test

import (
	. "dappco.re/go"
)

func ExampleFs_New() {
	f := (&Fs{}).New("/srv/workspaces")
	Println(f.Root())
	// Output: /srv/workspaces
}

func ExampleFs_Read() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "hello.txt"), "hello")

	Println(f.Read(Path(dir, "hello.txt")).Value)
	// Output: hello
}

func ExampleFs_Write() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)

	r := f.Write(Path(dir, "hello.txt"), "hello")
	Println(r.OK)
	Println(f.Read(Path(dir, "hello.txt")).Value)
	// Output:
	// true
	// hello
}

func ExampleFs_WriteMode() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)

	r := f.WriteMode(Path(dir, "secret.txt"), "secret", 0600)
	Println(r.OK)
	// Output: true
}

func ExampleFs_TempDir() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)

	Println(PathBase(dir) != "")
	// Output: true
}

func ExampleDirFS() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "hello.txt"), "hello")

	emb := Mount(DirFS(dir), ".").Value.(*Embed)
	Println(emb.ReadString("hello.txt").Value)
	// Output: hello
}

func ExampleFs_WriteAtomic() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("example")
	defer f.DeleteAll(dir)

	path := Path(dir, "status.json")
	f.WriteAtomic(path, `{"status":"completed"}`)

	r := f.Read(path)
	Println(r.Value)
	// Output: {"status":"completed"}
}

func ExampleFs_EnsureDir() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)

	r := f.EnsureDir(Path(dir, "logs"))
	Println(r.OK)
	Println(f.IsDir(Path(dir, "logs")))
	// Output:
	// true
	// true
}

func ExampleFs_IsDir() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	Println(f.IsDir(dir))
	// Output: true
}

func ExampleFs_IsFile() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.IsFile(path))
	// Output: true
}

func ExampleFs_Exists() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.Exists(path))
	// Output: true
}

func ExampleFs_List() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "hello.txt"), "hello")
	Println(f.List(dir).OK)
	// Output: true
}

func ExampleFs_Stat() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.Stat(path).OK)
	// Output: true
}

func ExampleFs_Open() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")

	r := f.Open(path)
	Println(r.OK)
	CloseStream(r.Value)
	// Output: true
}

func ExampleFs_Create() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")

	r := f.Create(path)
	WriteAll(r.Value, "hello")
	Println(f.Read(path).Value)
	// Output: hello
}

func ExampleFs_Append() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")

	r := f.Append(path)
	WriteAll(r.Value, " world")
	Println(f.Read(path).Value)
	// Output: hello world
}

func ExampleFs_ReadStream() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")

	r := f.ReadStream(path)
	Println(ReadAll(r.Value).Value)
	// Output: hello
}

func ExampleFs_WriteStream() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")

	r := f.WriteStream(path)
	WriteAll(r.Value, "hello")
	Println(f.Read(path).Value)
	// Output: hello
}

func ExampleReadAll() {
	r := ReadAll(NewReader("hello"))
	Println(r.Value)
	// Output: hello
}

func ExampleWriteAll() {
	buf := NewBuffer()
	r := WriteAll(buf, "hello")
	Println(r.OK)
	Println(buf.String())
	// Output:
	// true
	// hello
}

func ExampleCloseStream() {
	CloseStream(NewReader("not a closer"))
	Println("ok")
	// Output: ok
}

func ExampleFs_Delete() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.Delete(path).OK)
	Println(f.Exists(path))
	// Output:
	// true
	// false
}

func ExampleFs_DeleteAll() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	f.Write(Path(dir, "nested", "hello.txt"), "hello")
	Println(f.DeleteAll(dir).OK)
	Println(f.Exists(dir))
	// Output:
	// true
	// false
}

func ExampleFs_Rename() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	oldPath := Path(dir, "old.txt")
	newPath := Path(dir, "new.txt")
	f.Write(oldPath, "hello")
	Println(f.Rename(oldPath, newPath).OK)
	Println(f.Read(newPath).Value)
	// Output:
	// true
	// hello
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
	Println(r.Value)
	// Output: hello
}

func ExampleFs_Root() {
	f := (&Fs{}).New("/srv/workspaces")
	Println(f.Root())
	// Output: /srv/workspaces
}

func ExampleFsEntry() {
	entry := FsEntry{Path: "src/main.go", Name: "main.go", IsDir: false}
	Println(entry.Path)
	Println(entry.Name)
	Println(entry.IsDir)
	// Output:
	// src/main.go
	// main.go
	// false
}

func ExampleFs_WalkSeq() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "main.go"), "package main")

	for entry, err := range f.WalkSeq(dir) {
		if err == nil && !entry.IsDir {
			Println(entry.Name)
		}
	}
	// Output: main.go
}

func ExampleFs_WalkSeqSkip() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "src", "main.go"), "package main")
	f.Write(Path(dir, "vendor", "dep.go"), "package dep")

	var names []string
	for entry, err := range f.WalkSeqSkip(dir, "vendor") {
		if err == nil && !entry.IsDir {
			names = append(names, entry.Name)
		}
	}
	Println(names)
	// Output: [main.go]
}
