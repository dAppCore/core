package core_test

import (
	. "dappco.re/go"
)

// ExampleFs_New creates a sandboxed filesystem rooted at a workspace path. File reads,
// writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_New() {
	f := (&Fs{}).New("/srv/workspaces")
	Println(f.Root())
	// Output: /srv/workspaces
}

// ExampleFs_Read reads content through `Fs.Read` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_Read() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "hello.txt"), "hello")

	Println(f.Read(Path(dir, "hello.txt")).Value)
	// Output: hello
}

// ExampleFs_Write writes content through `Fs.Write` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_WriteMode writes content with explicit permissions through `Fs.WriteMode` for
// sandboxed file operations. File reads, writes, walks, and cleanup stay sandbox-aware
// through Fs.
func ExampleFs_WriteMode() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)

	r := f.WriteMode(Path(dir, "secret.txt"), "secret", 0600)
	Println(r.OK)
	// Output: true
}

// ExampleFs_TempDir creates a temporary directory through `Fs.TempDir` for sandboxed file
// operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_TempDir() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)

	Println(PathBase(dir) != "")
	// Output: true
}

// ExampleDirFS creates a directory-backed filesystem through `DirFS` for sandboxed file
// operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleDirFS() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "hello.txt"), "hello")

	emb := Mount(DirFS(dir), ".").Value.(*Embed)
	Println(emb.ReadString("hello.txt").Value)
	// Output: hello
}

// ExampleFs_WriteAtomic writes content atomically through `Fs.WriteAtomic` for sandboxed
// file operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_EnsureDir creates missing directories through `Fs.EnsureDir` for sandboxed
// file operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_IsDir checks directory state through `Fs.IsDir` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_IsDir() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	Println(f.IsDir(dir))
	// Output: true
}

// ExampleFs_IsFile checks file state through `Fs.IsFile` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_IsFile() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.IsFile(path))
	// Output: true
}

// ExampleFs_Exists checks whether a sandboxed path exists after writing a file. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_Exists() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.Exists(path))
	// Output: true
}

// ExampleFs_List lists entries through `Fs.List` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_List() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	f.Write(Path(dir, "hello.txt"), "hello")
	Println(f.List(dir).OK)
	// Output: true
}

// ExampleFs_Stat reads metadata through `Fs.Stat` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_Stat() {
	f := (&Fs{}).New("/")
	dir := f.TempDir("core-fs-example")
	defer f.DeleteAll(dir)
	path := Path(dir, "hello.txt")
	f.Write(path, "hello")
	Println(f.Stat(path).OK)
	// Output: true
}

// ExampleFs_Open opens a sandboxed file for streaming reads. File reads, writes, walks,
// and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_Create creates a resource through `Fs.Create` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_Append appends content through `Fs.Append` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_ReadStream opens a read stream through `Fs.ReadStream` for sandboxed file
// operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_WriteStream opens a write stream through `Fs.WriteStream` for sandboxed file
// operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleReadAll reads an entire stream through `ReadAll` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleReadAll() {
	r := ReadAll(NewReader("hello"))
	Println(r.Value)
	// Output: hello
}

// ExampleWriteAll writes a complete payload through `WriteAll` for sandboxed file
// operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleWriteAll() {
	buf := NewBuffer()
	r := WriteAll(buf, "hello")
	Println(r.OK)
	Println(buf.String())
	// Output:
	// true
	// hello
}

// ExampleCloseStream closes a stream through `CloseStream` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleCloseStream() {
	CloseStream(NewReader("not a closer"))
	Println("ok")
	// Output: ok
}

// ExampleFs_Delete deletes a value through `Fs.Delete` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_DeleteAll deletes a tree through `Fs.DeleteAll` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_Rename renames a path through `Fs.Rename` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_NewUnrestricted creates an unrestricted filesystem through
// `Fs.NewUnrestricted` for sandboxed file operations. File reads, writes, walks, and
// cleanup stay sandbox-aware through Fs.
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

// ExampleFs_Root reports the root value through `Fs.Root` for sandboxed file operations.
// File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
func ExampleFs_Root() {
	f := (&Fs{}).New("/srv/workspaces")
	Println(f.Root())
	// Output: /srv/workspaces
}

// ExampleFsEntry reads a walked filesystem entry through `FsEntry` for sandboxed file
// operations. File reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_WalkSeq walks a tree through `Fs.WalkSeq` for sandboxed file operations. File
// reads, writes, walks, and cleanup stay sandbox-aware through Fs.
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

// ExampleFs_WalkSeqSkip walks a tree through `Fs.WalkSeqSkip` while skipping a branch for
// sandboxed file operations. File reads, writes, walks, and cleanup stay sandbox-aware
// through Fs.
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
