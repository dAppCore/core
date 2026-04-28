package core_test

import . "dappco.re/go"

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
	AssertLen(t, r.Value.([]FsDirEntry), 2)
}

func TestFs_Stat_Good(t *T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stat.txt")
	c.Fs().Write(path, "data")
	r := c.Fs().Stat(path)
	AssertTrue(t, r.OK)
	AssertEqual(t, "stat.txt", r.Value.(FsFileInfo).Name())
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
	AssertEqual(t, "secret.txt", r.Value.(FsFileInfo).Name())
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
	entries := r.Value.([]FsDirEntry)
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
	entries, _ := lr.Value.([]FsDirEntry)
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
	sandboxed := (&Fs{}).New(ax7TempRoot(t))
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

func ax7TempRoot(t *T) string {
	t.Helper()
	dir := t.TempDir()
	r := PathEvalSymlinks(dir)
	if r.OK {
		return r.Value.(string)
	}
	return dir
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
			AssertEqual(t, FileMode(0o600), entry.Mode.Perm())
		}
	}
}

func TestFs_CloseStream_Good(t *T) {
	dir := ax7TempRoot(t)
	fsys := (&Fs{}).New(dir)
	AssertTrue(t, fsys.Write("agent.log", "ready").OK)
	r := fsys.ReadStream("agent.log")
	RequireTrue(t, r.OK)
	closer := r.Value.(ReadCloser)

	CloseStream(closer)

	AssertError(t, closer.Close())
}

func TestFs_CloseStream_Bad(t *T) {
	CloseStream("not a stream")

	AssertTrue(t, true)
}

func TestFs_CloseStream_Ugly(t *T) {
	CloseStream(nil)

	AssertTrue(t, true)
}

func TestFs_Fs_New_Good(t *T) {
	dir := ax7TempRoot(t)

	fsys := (&Fs{}).New(dir)

	AssertEqual(t, dir, fsys.Root())
}

func TestFs_Fs_New_Bad(t *T) {
	fsys := (&Fs{}).New("")

	AssertEqual(t, "/", fsys.Root())
}

func TestFs_Fs_New_Ugly(t *T) {
	first := t.TempDir()
	second := t.TempDir()
	fsys := (&Fs{}).New(first)

	fsys.New(second)

	AssertEqual(t, second, fsys.Root())
}

func TestFs_Fs_NewUnrestricted_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t)).NewUnrestricted()

	AssertEqual(t, "/", fsys.Root())
}

func TestFs_Fs_NewUnrestricted_Bad(t *T) {
	sandboxed := (&Fs{}).New(ax7TempRoot(t))
	unrestricted := sandboxed.NewUnrestricted()

	AssertNotEqual(t, sandboxed.Root(), unrestricted.Root())
}

func TestFs_Fs_NewUnrestricted_Ugly(t *T) {
	root := t.TempDir()
	sandboxed := (&Fs{}).New(root)

	_ = sandboxed.NewUnrestricted()

	AssertEqual(t, root, sandboxed.Root())
}

func TestFs_Fs_Root_Good(t *T) {
	root := t.TempDir()

	AssertEqual(t, root, (&Fs{}).New(root).Root())
}

func TestFs_Fs_Root_Bad(t *T) {
	var fsys Fs

	AssertEqual(t, "/", fsys.Root())
}

func TestFs_Fs_Root_Ugly(t *T) {
	fsys := (&Fs{}).New("")

	AssertEqual(t, "/", fsys.Root())
}

func TestFs_Fs_TempDir_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	dir := fsys.TempDir("agent-")
	defer RemoveAll(dir)

	AssertNotEmpty(t, dir)
	AssertTrue(t, (&Fs{}).New("/").IsDir(dir))
}

func TestFs_Fs_TempDir_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	dir := fsys.TempDir("")
	defer RemoveAll(dir)

	AssertNotEmpty(t, dir)
}

func TestFs_Fs_TempDir_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	first := fsys.TempDir("agent-")
	second := fsys.TempDir("agent-")
	defer RemoveAll(first)
	defer RemoveAll(second)

	AssertNotEqual(t, first, second)
}

func TestFs_Fs_Write_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Write("config/agent.json", `{"status":"ready"}`)

	AssertTrue(t, r.OK)
	read := fsys.Read("config/agent.json")
	AssertTrue(t, read.OK)
	AssertEqual(t, `{"status":"ready"}`, read.Value.(string))
}

func TestFs_Fs_Write_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("config", "file").OK)

	r := fsys.Write("config/agent.json", "blocked")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Write_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Write("empty.txt", "")

	AssertTrue(t, r.OK)
	read := fsys.Read("empty.txt")
	AssertTrue(t, read.OK)
	AssertEqual(t, "", read.Value.(string))
}

func TestFs_Fs_WriteMode_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.WriteMode("secrets/session.token", "token", 0o600)

	AssertTrue(t, r.OK)
	stat := fsys.Stat("secrets/session.token")
	AssertTrue(t, stat.OK)
	info := stat.Value.(FsFileInfo)
	AssertEqual(t, FileMode(0o600), info.Mode().Perm())
}

func TestFs_Fs_WriteMode_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("secrets", "file").OK)

	r := fsys.WriteMode("secrets/session.token", "token", 0o600)

	AssertFalse(t, r.OK)
}

func TestFs_Fs_WriteMode_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.WriteMode("public.txt", "open", 0o644)

	AssertTrue(t, r.OK)
	stat := fsys.Stat("public.txt")
	AssertTrue(t, stat.OK)
	info := stat.Value.(FsFileInfo)
	AssertEqual(t, FileMode(0o644), info.Mode().Perm())
}

func TestFs_Fs_WriteAtomic_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.WriteAtomic("status/agent.json", `{"ok":true}`)

	AssertTrue(t, r.OK)
	read := fsys.Read("status/agent.json")
	AssertTrue(t, read.OK)
	AssertEqual(t, `{"ok":true}`, read.Value.(string))
}

func TestFs_Fs_WriteAtomic_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("status", "file").OK)

	r := fsys.WriteAtomic("status/agent.json", `{"ok":true}`)

	AssertFalse(t, r.OK)
}

func TestFs_Fs_WriteAtomic_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	AssertTrue(t, fsys.WriteAtomic("status.json", "first").OK)
	AssertTrue(t, fsys.WriteAtomic("status.json", "second").OK)
	read := fsys.Read("status.json")

	AssertTrue(t, read.OK)
	AssertEqual(t, "second", read.Value.(string))
}

func TestFs_Fs_Read_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("config/agent.json", "ready").OK)

	r := fsys.Read("config/agent.json")

	AssertTrue(t, r.OK)
	AssertEqual(t, "ready", r.Value.(string))
}

func TestFs_Fs_Read_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Read("missing.txt")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Read_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("empty.txt", "").OK)

	r := fsys.Read("empty.txt")

	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value.(string))
}

func TestFs_Fs_EnsureDir_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.EnsureDir("logs/agent")

	AssertTrue(t, r.OK)
	AssertTrue(t, fsys.IsDir("logs/agent"))
}

func TestFs_Fs_EnsureDir_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("logs", "file").OK)

	r := fsys.EnsureDir("logs/agent")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_EnsureDir_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.EnsureDir("")

	AssertTrue(t, r.OK)
	AssertFalse(t, fsys.IsDir(""))
}

func TestFs_Fs_IsDir_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.EnsureDir("logs").OK)

	AssertTrue(t, fsys.IsDir("logs"))
}

func TestFs_Fs_IsDir_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("logs.txt", "file").OK)

	AssertFalse(t, fsys.IsDir("logs.txt"))
}

func TestFs_Fs_IsDir_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	AssertFalse(t, fsys.IsDir(""))
}

func TestFs_Fs_IsFile_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("config.json", "file").OK)

	AssertTrue(t, fsys.IsFile("config.json"))
}

func TestFs_Fs_IsFile_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.EnsureDir("config").OK)

	AssertFalse(t, fsys.IsFile("config"))
}

func TestFs_Fs_IsFile_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	AssertFalse(t, fsys.IsFile(""))
}

func TestFs_Fs_Exists_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("config.json", "file").OK)

	AssertTrue(t, fsys.Exists("config.json"))
}

func TestFs_Fs_Exists_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	AssertFalse(t, fsys.Exists("missing.json"))
}

func TestFs_Fs_Exists_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	AssertTrue(t, fsys.Exists(""))
}

func TestFs_Fs_List_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agents/a.json", "a").OK)
	AssertTrue(t, fsys.Write("agents/b.json", "b").OK)

	r := fsys.List("agents")

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]FsDirEntry), 2)
}

func TestFs_Fs_List_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.List("missing")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_List_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.EnsureDir("agents").OK)

	r := fsys.List("agents")

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]FsDirEntry), 0)
}

func TestFs_Fs_Stat_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.json", "ready").OK)

	r := fsys.Stat("agent.json")

	AssertTrue(t, r.OK)
	AssertEqual(t, "agent.json", r.Value.(FsFileInfo).Name())
}

func TestFs_Fs_Stat_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Stat("missing.json")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Stat_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.EnsureDir("agents").OK)

	r := fsys.Stat("agents")

	AssertTrue(t, r.OK)
	AssertTrue(t, r.Value.(FsFileInfo).IsDir())
}

func TestFs_Fs_Open_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.json", "ready").OK)

	r := fsys.Open("agent.json")

	AssertTrue(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_Fs_Open_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Open("missing.json")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Open_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("empty.txt", "").OK)

	r := fsys.Open("empty.txt")

	AssertTrue(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_Fs_Create_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Create("logs/agent.log")
	RequireTrue(t, r.OK)

	AssertTrue(t, WriteAll(r.Value, "ready").OK)
	AssertEqual(t, "ready", fsys.Read("logs/agent.log").Value.(string))
}

func TestFs_Fs_Create_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("logs", "file").OK)

	r := fsys.Create("logs/agent.log")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Create_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.log", "old").OK)

	r := fsys.Create("agent.log")
	RequireTrue(t, r.OK)
	CloseStream(r.Value)
	read := fsys.Read("agent.log")

	AssertTrue(t, read.OK)
	AssertEqual(t, "", read.Value.(string))
}

func TestFs_Fs_Append_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.log", "start").OK)

	r := fsys.Append("agent.log")
	RequireTrue(t, r.OK)

	AssertTrue(t, WriteAll(r.Value, " ready").OK)
	AssertEqual(t, "start ready", fsys.Read("agent.log").Value.(string))
}

func TestFs_Fs_Append_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("logs", "file").OK)

	r := fsys.Append("logs/agent.log")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Append_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Append("new.log")
	RequireTrue(t, r.OK)

	AssertTrue(t, WriteAll(r.Value, "created").OK)
	AssertEqual(t, "created", fsys.Read("new.log").Value.(string))
}

func TestFs_Fs_ReadStream_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.log", "ready").OK)

	r := fsys.ReadStream("agent.log")

	AssertTrue(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_Fs_ReadStream_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.ReadStream("missing.log")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_ReadStream_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("empty.log", "").OK)

	r := fsys.ReadStream("empty.log")
	RequireTrue(t, r.OK)
	read := ReadAll(r.Value)

	AssertTrue(t, read.OK)
	AssertEqual(t, "", read.Value.(string))
}

func TestFs_Fs_WriteStream_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.WriteStream("logs/agent.log")
	RequireTrue(t, r.OK)

	AssertTrue(t, WriteAll(r.Value, "ready").OK)
	AssertEqual(t, "ready", fsys.Read("logs/agent.log").Value.(string))
}

func TestFs_Fs_WriteStream_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("logs", "file").OK)

	r := fsys.WriteStream("logs/agent.log")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_WriteStream_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.WriteStream("empty.log")
	RequireTrue(t, r.OK)

	AssertTrue(t, WriteAll(r.Value, "").OK)
	AssertEqual(t, "", fsys.Read("empty.log").Value.(string))
}

func TestFs_Fs_Delete_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("old.log", "gone").OK)

	r := fsys.Delete("old.log")

	AssertTrue(t, r.OK)
	AssertFalse(t, fsys.Exists("old.log"))
}

func TestFs_Fs_Delete_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Delete("missing.log")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Delete_Ugly(t *T) {
	fsys := (&Fs{}).New("/")

	r := fsys.Delete("/")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_DeleteAll_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("sessions/one/status.json", "done").OK)

	r := fsys.DeleteAll("sessions")

	AssertTrue(t, r.OK)
	AssertFalse(t, fsys.Exists("sessions"))
}

func TestFs_Fs_DeleteAll_Bad(t *T) {
	fsys := (&Fs{}).New("/")

	r := fsys.DeleteAll("/")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_DeleteAll_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.DeleteAll("missing")

	AssertTrue(t, r.OK)
}

func TestFs_Fs_Rename_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.tmp", "ready").OK)

	r := fsys.Rename("agent.tmp", "agent.json")

	AssertTrue(t, r.OK)
	AssertFalse(t, fsys.Exists("agent.tmp"))
	AssertTrue(t, fsys.Exists("agent.json"))
}

func TestFs_Fs_Rename_Bad(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))

	r := fsys.Rename("missing.tmp", "agent.json")

	AssertFalse(t, r.OK)
}

func TestFs_Fs_Rename_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	AssertTrue(t, fsys.Write("agent.tmp", "new").OK)
	AssertTrue(t, fsys.Write("agent.json", "old").OK)

	r := fsys.Rename("agent.tmp", "agent.json")

	AssertTrue(t, r.OK)
	AssertEqual(t, "new", fsys.Read("agent.json").Value.(string))
}

func TestFs_Fs_WalkSeq_Good(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	fsys := (&Fs{}).New("/")
	seen := map[string]bool{}

	for entry, err := range fsys.WalkSeq(dir) {
		AssertNoError(t, err)
		seen[entry.Name] = true
	}

	AssertTrue(t, seen["a.txt"])
	AssertTrue(t, seen["b.txt"])
}

func TestFs_Fs_WalkSeq_Bad(t *T) {
	fsys := (&Fs{}).New("/")
	seenError := false

	for _, err := range fsys.WalkSeq(Path(t.TempDir(), "missing")) {
		if err != nil {
			seenError = true
			break
		}
	}

	AssertTrue(t, seenError)
}

func TestFs_Fs_WalkSeq_Ugly(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	fsys := (&Fs{}).New("/")
	count := 0

	for _, err := range fsys.WalkSeq(dir) {
		AssertNoError(t, err)
		count++
		break
	}

	AssertEqual(t, 1, count)
}

func TestFs_Fs_WalkSeqSkip_Good(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	fsys := (&Fs{}).New("/")
	seen := map[string]bool{}

	for entry, err := range fsys.WalkSeqSkip(dir, "vendor", ".git") {
		AssertNoError(t, err)
		seen[entry.Name] = true
	}

	AssertTrue(t, seen["a.txt"])
	AssertFalse(t, seen["skipme.txt"])
	AssertFalse(t, seen["HEAD"])
}

func TestFs_Fs_WalkSeqSkip_Bad(t *T) {
	fsys := (&Fs{}).New("/")
	seenError := false

	for _, err := range fsys.WalkSeqSkip(Path(t.TempDir(), "missing"), "vendor") {
		if err != nil {
			seenError = true
			break
		}
	}

	AssertTrue(t, seenError)
}

func TestFs_Fs_WalkSeqSkip_Ugly(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	fsys := (&Fs{}).New("/")
	seen := map[string]bool{}

	for entry, err := range fsys.WalkSeqSkip(dir, "") {
		AssertNoError(t, err)
		seen[entry.Name] = true
	}

	AssertTrue(t, seen["skipme.txt"])
}

func TestFs_ReadDir_Ugly(t *T) {
	r := ReadDir(DirFS(t.TempDir()), ".")

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]FsDirEntry), 0)
}

func TestFs_ReadFSFile_Good(t *T) {
	dir := t.TempDir()
	AssertTrue(t, WriteFile(Path(dir, "agent.txt"), []byte("ready"), 0o644).OK)

	r := ReadFSFile(DirFS(dir), "agent.txt")

	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("ready"), r.Value.([]byte))
}

func TestFs_ReadFSFile_Bad(t *T) {
	r := ReadFSFile(DirFS(t.TempDir()), "missing.txt")

	AssertFalse(t, r.OK)
}

func TestFs_ReadFSFile_Ugly(t *T) {
	dir := t.TempDir()
	AssertTrue(t, WriteFile(Path(dir, "empty.txt"), nil, 0o644).OK)

	r := ReadFSFile(DirFS(dir), "empty.txt")

	AssertTrue(t, r.OK)
	AssertEqual(t, []byte{}, r.Value.([]byte))
}

func TestFs_Sub_Bad(t *T) {
	r := Sub(DirFS(t.TempDir()), "../escape")

	AssertFalse(t, r.OK)
}

func TestFs_Sub_Ugly(t *T) {
	dir := t.TempDir()
	AssertTrue(t, WriteFile(Path(dir, "agent.txt"), []byte("ready"), 0o644).OK)
	r := Sub(DirFS(dir), ".")
	RequireTrue(t, r.OK)

	read := ReadFSFile(r.Value.(FS), "agent.txt")

	AssertTrue(t, read.OK)
	AssertEqual(t, []byte("ready"), read.Value.([]byte))
}

func TestFs_WalkDir_Good(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	count := 0

	err := WalkDir(DirFS(dir), ".", func(_ string, _ FsDirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		return nil
	})

	AssertNoError(t, err)
	AssertGreater(t, count, 1)
}

func TestFs_WalkDir_Bad(t *T) {
	err := WalkDir(DirFS(t.TempDir()), "missing", func(_ string, _ FsDirEntry, err error) error {
		return err
	})

	AssertError(t, err)
}

func TestFs_WalkDir_Ugly(t *T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	seenVendorChild := false

	err := WalkDir(DirFS(dir), ".", func(_ string, entry FsDirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == "vendor" {
			return PathSkipDir
		}
		if entry.Name() == "skipme.txt" {
			seenVendorChild = true
		}
		return nil
	})

	AssertNoError(t, err)
	AssertFalse(t, seenVendorChild)
}

func TestFs_WriteAll_Good(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	r := fsys.WriteStream("agent.log")
	RequireTrue(t, r.OK)

	w := WriteAll(r.Value, "ready")

	AssertTrue(t, w.OK)
	AssertEqual(t, "ready", fsys.Read("agent.log").Value.(string))
}

func TestFs_WriteAll_Bad(t *T) {
	r := WriteAll("not a writer", "ready")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "not a writer")
}

func TestFs_WriteAll_Ugly(t *T) {
	fsys := (&Fs{}).New(ax7TempRoot(t))
	r := fsys.WriteStream("empty.log")
	RequireTrue(t, r.OK)

	w := WriteAll(r.Value, "")

	AssertTrue(t, w.OK)
	AssertEqual(t, "", fsys.Read("empty.log").Value.(string))
}
