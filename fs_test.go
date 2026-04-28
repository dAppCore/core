package core_test

import (
	"io/fs"

	. "dappco.re/go/core"
)

// --- Fs (Sandboxed Filesystem) ---

func TestFs_WriteRead_Good(t *T) {
	dir := t.TempDir()
	c := New()

	path := Path(dir, "test.txt")
	AssertTrue(t, c.Fs().Write(path, "hello core").OK)

	r := c.Fs().Read(path)
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello core", r.Value.(string))
}

func TestFs_Read_Bad(t *T) {
	c := New()
	r := c.Fs().Read("/nonexistent/path/to/file.txt")
	AssertFalse(t, r.OK)
}

func TestFs_EnsureDir_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "dir")
	AssertTrue(t, c.Fs().EnsureDir(path).OK)
	AssertTrue(t, c.Fs().IsDir(path))
}

func TestFs_IsDir_Good(t *T) {
	c := New()
	dir := t.TempDir()
	AssertTrue(t, c.Fs().IsDir(dir))
	AssertFalse(t, c.Fs().IsDir(Path(dir, "nonexistent")))
	AssertFalse(t, c.Fs().IsDir(""))
}

func TestFs_IsFile_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "test.txt")
	c.Fs().Write(path, "data")
	AssertTrue(t, c.Fs().IsFile(path))
	AssertFalse(t, c.Fs().IsFile(dir))
	AssertFalse(t, c.Fs().IsFile(""))
}

func TestFs_Exists_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "exists.txt")
	c.Fs().Write(path, "yes")
	AssertTrue(t, c.Fs().Exists(path))
	AssertTrue(t, c.Fs().Exists(dir))
	AssertFalse(t, c.Fs().Exists(Path(dir, "nope")))
}

func TestFs_List_Good(t *T) {
	dir := t.TempDir()
	c := New()
	c.Fs().Write(Path(dir, "a.txt"), "a")
	c.Fs().Write(Path(dir, "b.txt"), "b")
	r := c.Fs().List(dir)
	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]fs.DirEntry), 2)
}

func TestFs_Stat_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stat.txt")
	c.Fs().Write(path, "data")
	r := c.Fs().Stat(path)
	AssertTrue(t, r.OK)
	AssertEqual(t, "stat.txt", r.Value.(fs.FileInfo).Name())
}

func TestFs_Open_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "open.txt")
	c.Fs().Write(path, "content")
	r := c.Fs().Open(path)
	AssertTrue(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_Create_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "created.txt")
	r := c.Fs().Create(path)
	AssertTrue(t, r.OK)
	WriteAll(r.Value, "hello")
	rr := c.Fs().Read(path)
	AssertEqual(t, "hello", rr.Value.(string))
}

func TestFs_Append_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "append.txt")
	c.Fs().Write(path, "first")
	r := c.Fs().Append(path)
	AssertTrue(t, r.OK)
	WriteAll(r.Value, " second")
	rr := c.Fs().Read(path)
	AssertEqual(t, "first second", rr.Value.(string))
}

func TestFs_ReadStream_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stream.txt")
	c.Fs().Write(path, "streamed")
	r := c.Fs().ReadStream(path)
	AssertTrue(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_WriteStream_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "ws.txt")
	r := c.Fs().WriteStream(path)
	AssertTrue(t, r.OK)
	WriteAll(r.Value, "stream")
}

func TestFs_Delete_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "delete.txt")
	c.Fs().Write(path, "gone")
	AssertTrue(t, c.Fs().Delete(path).OK)
	AssertFalse(t, c.Fs().Exists(path))
}

func TestFs_DeleteAll_Good(t *T) {
	dir := t.TempDir()
	c := New()
	sub := Path(dir, "deep", "nested")
	c.Fs().EnsureDir(sub)
	c.Fs().Write(Path(sub, "file.txt"), "data")
	AssertTrue(t, c.Fs().DeleteAll(Path(dir, "deep")).OK)
	AssertFalse(t, c.Fs().Exists(Path(dir, "deep")))
}

func TestFs_Rename_Good(t *T) {
	dir := t.TempDir()
	c := New()
	old := Path(dir, "old.txt")
	nw := Path(dir, "new.txt")
	c.Fs().Write(old, "data")
	AssertTrue(t, c.Fs().Rename(old, nw).OK)
	AssertFalse(t, c.Fs().Exists(old))
	AssertTrue(t, c.Fs().Exists(nw))
}

func TestFs_WriteMode_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "secret.txt")
	AssertTrue(t, c.Fs().WriteMode(path, "secret", 0600).OK)
	r := c.Fs().Stat(path)
	AssertTrue(t, r.OK)
	AssertEqual(t, "secret.txt", r.Value.(fs.FileInfo).Name())
}

// --- Zero Value ---

func TestFs_ZeroValue_Good(t *T) {
	dir := t.TempDir()
	zeroFs := &Fs{}

	path := Path(dir, "zero.txt")
	AssertTrue(t, zeroFs.Write(path, "zero value works").OK)
	r := zeroFs.Read(path)
	AssertTrue(t, r.OK)
	AssertEqual(t, "zero value works", r.Value.(string))
	AssertTrue(t, zeroFs.IsFile(path))
	AssertTrue(t, zeroFs.Exists(path))
	AssertTrue(t, zeroFs.IsDir(dir))
}

func TestFs_ZeroValue_List_Good(t *T) {
	dir := t.TempDir()
	zeroFs := &Fs{}

	(&Fs{}).New("/").Write(Path(dir, "a.txt"), "a")
	r := zeroFs.List(dir)
	AssertTrue(t, r.OK)
	entries := r.Value.([]fs.DirEntry)
	AssertLen(t, entries, 1)
}

func TestFs_Exists_NotFound_Bad(t *T) {
	c := New()
	AssertFalse(t, c.Fs().Exists("/nonexistent/path/xyz"))
}

// --- Fs path/validatePath edge cases ---

func TestFs_Read_EmptyPath_Ugly(t *T) {
	c := New()
	r := c.Fs().Read("")
	AssertFalse(t, r.OK)
}

func TestFs_Write_EmptyPath_Ugly(t *T) {
	c := New()
	r := c.Fs().Write("", "data")
	AssertFalse(t, r.OK)
}

func TestFs_Delete_Protected_Ugly(t *T) {
	c := New()
	r := c.Fs().Delete("/")
	AssertFalse(t, r.OK)
}

func TestFs_DeleteAll_Protected_Ugly(t *T) {
	c := New()
	r := c.Fs().DeleteAll("/")
	AssertFalse(t, r.OK)
}

func TestFs_ReadStream_WriteStream_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stream.txt")
	c.Fs().Write(path, "streamed")

	r := c.Fs().ReadStream(path)
	AssertTrue(t, r.OK)

	w := c.Fs().WriteStream(path)
	AssertTrue(t, w.OK)
}

// --- WriteAtomic ---

func TestFs_WriteAtomic_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "status.json")
	r := c.Fs().WriteAtomic(path, `{"status":"completed"}`)
	AssertTrue(t, r.OK)

	read := c.Fs().Read(path)
	AssertTrue(t, read.OK)
	AssertEqual(t, `{"status":"completed"}`, read.Value)
}

func TestFs_WriteAtomic_Good_Overwrite(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "data.txt")
	c.Fs().WriteAtomic(path, "first")
	c.Fs().WriteAtomic(path, "second")

	read := c.Fs().Read(path)
	AssertEqual(t, "second", read.Value)
}

func TestFs_WriteAtomic_Bad_ReadOnlyDir(t *T) {
	// Write to a non-existent root that can't be created
	m := (&Fs{}).New("/proc/nonexistent")
	r := m.WriteAtomic("file.txt", "data")
	AssertFalse(t, r.OK, "WriteAtomic must fail when parent dir cannot be created")
}

func TestFs_WriteAtomic_Ugly_NoTempFileLeftOver(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "clean.txt")
	c.Fs().WriteAtomic(path, "content")

	// Check no .tmp files remain
	lr := c.Fs().List(dir)
	entries, _ := lr.Value.([]fs.DirEntry)
	for _, e := range entries {
		AssertFalse(t, Contains(e.Name(), ".tmp."), "temp file should not remain after successful atomic write")
	}
}

func TestFs_WriteAtomic_Good_CreatesParentDir(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "dir", "file.txt")
	r := c.Fs().WriteAtomic(path, "nested")
	AssertTrue(t, r.OK)

	read := c.Fs().Read(path)
	AssertEqual(t, "nested", read.Value)
}

// --- NewUnrestricted ---

func TestFs_NewUnrestricted_Good(t *T) {
	sandboxed := (&Fs{}).New(t.TempDir())
	unrestricted := sandboxed.NewUnrestricted()
	AssertEqual(t, "/", unrestricted.Root())
}

func TestFs_NewUnrestricted_Good_CanReadOutsideSandbox(t *T) {
	dir := t.TempDir()
	outside := Path(dir, "outside.txt")
	(&Fs{}).New("/").Write(outside, "hello")

	sandboxed := (&Fs{}).New(Path(dir, "sandbox"))
	unrestricted := sandboxed.NewUnrestricted()

	r := unrestricted.Read(outside)
	AssertTrue(t, r.OK, "unrestricted Fs must read paths outside the original sandbox")
	AssertEqual(t, "hello", r.Value)
}

func TestFs_NewUnrestricted_Ugly_OriginalStaysSandboxed(t *T) {
	dir := t.TempDir()
	sandbox := Path(dir, "sandbox")
	(&Fs{}).New("/").EnsureDir(sandbox)

	sandboxed := (&Fs{}).New(sandbox)
	_ = sandboxed.NewUnrestricted() // getting unrestricted doesn't affect original

	AssertEqual(t, sandbox, sandboxed.Root(), "original Fs must remain sandboxed")
}

// --- Root ---

func TestFs_Root_Good(t *T) {
	m := (&Fs{}).New("/home/agent")
	AssertEqual(t, "/home/agent", m.Root())
}

func TestFs_Root_Good_Default(t *T) {
	m := (&Fs{}).New("")
	AssertEqual(t, "/", m.Root())
}

// --- WalkSeq / WalkSeqSkip ---

// walkSeqSeed builds a small fixture tree under dir for the walk tests.
//
//	root/
//	  a.txt
//	  sub/
//	    b.txt
//	  vendor/
//	    skipme.txt
//	  .git/
//	    HEAD
func walkSeqSeed(t *T, dir string) {
	t.Helper()
	c := New()
	c.Fs().Write(Path(dir, "a.txt"), "alpha")
	c.Fs().Write(Path(dir, "sub", "b.txt"), "bravo")
	c.Fs().Write(Path(dir, "vendor", "skipme.txt"), "skip")
	c.Fs().Write(Path(dir, ".git", "HEAD"), "ref: refs/heads/main")
}

func TestFs_WalkSeq_Good(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	seen := map[string]bool{}
	for entry, err := range c.Fs().WalkSeq(dir) {
		AssertNoError(t, err)
		seen[entry.Name] = true
	}
	AssertTrue(t, seen["a.txt"])
	AssertTrue(t, seen["b.txt"])
	AssertTrue(t, seen["sub"])
	AssertTrue(t, seen["vendor"])
	AssertTrue(t, seen["skipme.txt"])
	AssertTrue(t, seen[".git"])
	AssertTrue(t, seen["HEAD"])
}

func TestFs_WalkSeq_Good_BreakStops(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	count := 0
	for entry, err := range c.Fs().WalkSeq(dir) {
		AssertNoError(t, err)
		_ = entry
		count++
		if count == 2 {
			break
		}
	}
	AssertEqual(t, 2, count)
}

func TestFs_WalkSeq_Good_RelPath(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	rels := []string{}
	for entry, err := range c.Fs().WalkSeq(dir) {
		AssertNoError(t, err)
		if !entry.IsDir {
			rels = append(rels, entry.Path)
		}
	}
	// Paths are relative to the walk root, never start with the absolute dir.
	for _, p := range rels {
		AssertFalse(t, len(p) > 0 && p[0] == '/', "rel path leaked absolute: %s", p)
	}
}

func TestFs_WalkSeqSkip_Good(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	seen := map[string]bool{}
	for entry, err := range c.Fs().WalkSeqSkip(dir, "vendor", ".git") {
		AssertNoError(t, err)
		seen[entry.Name] = true
	}
	AssertTrue(t, seen["a.txt"])
	AssertTrue(t, seen["sub"])
	AssertTrue(t, seen["b.txt"])
	// Skipped directory names themselves still appear (as the entry where
	// skip is decided), but their contents do not.
	AssertFalse(t, seen["skipme.txt"], "vendor contents should be skipped")
	AssertFalse(t, seen["HEAD"], ".git contents should be skipped")
}

func TestFs_WalkSeqSkip_Good_RootBasenameNotSkipped(t *T) {
	// If the user explicitly walks a directory whose basename matches a
	// skipName, the walk root itself is still entered.
	parent := t.TempDir()
	c := New()
	c.Fs().Write(Path(parent, "vendor", "kept.txt"), "kept")
	seen := map[string]bool{}
	for entry, err := range c.Fs().WalkSeqSkip(Path(parent, "vendor"), "vendor") {
		AssertNoError(t, err)
		seen[entry.Name] = true
	}
	AssertTrue(t, seen["kept.txt"], "walk root must not be self-skipped")
}

func TestFs_WalkSeq_Good_SandboxContainment(t *T) {
	// Fs sandboxed to a temp dir; walking ".." must NOT escape — every
	// yielded entry path must remain inside the sandbox root.
	root := t.TempDir()
	sandbox := (&Fs{}).New(root)
	c := New()
	c.Fs().Write(Path(root, "inside.txt"), "ok")

	for entry, err := range sandbox.WalkSeq("..") {
		AssertNoError(t, err)
		// Entry paths are relative to the (resolved) walk root; they
		// must never start with ".." which would indicate escape.
		AssertFalse(t, len(entry.Path) >= 2 && entry.Path[:2] == "..",
			"sandbox escape: %s", entry.Path)
	}
}

func TestFs_WalkSeq_Good_FileMode(t *T) {
	dir := t.TempDir()
	c := New()
	c.Fs().WriteMode(Path(dir, "secret.key"), "x", 0o600)
	for entry, err := range c.Fs().WalkSeq(dir) {
		AssertNoError(t, err)
		if entry.Name == "secret.key" {
			AssertEqual(t, fs.FileMode(0o600), entry.Mode.Perm())
		}
	}
}
