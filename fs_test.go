package core_test

import (
	"io/fs"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Fs (Sandboxed Filesystem) ---

func TestFs_WriteRead_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := Path(dir, "test.txt")
	assert.True(t, c.Fs().Write(path, "hello core").OK)

	r := c.Fs().Read(path)
	assert.True(t, r.OK)
	assert.Equal(t, "hello core", r.Value.(string))
}

func TestFs_Read_Bad(t *testing.T) {
	c := New()
	r := c.Fs().Read("/nonexistent/path/to/file.txt")
	assert.False(t, r.OK)
}

func TestFs_EnsureDir_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "dir")
	assert.True(t, c.Fs().EnsureDir(path).OK)
	assert.True(t, c.Fs().IsDir(path))
}

func TestFs_IsDir_Good(t *testing.T) {
	c := New()
	dir := t.TempDir()
	assert.True(t, c.Fs().IsDir(dir))
	assert.False(t, c.Fs().IsDir(Path(dir, "nonexistent")))
	assert.False(t, c.Fs().IsDir(""))
}

func TestFs_IsFile_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "test.txt")
	c.Fs().Write(path, "data")
	assert.True(t, c.Fs().IsFile(path))
	assert.False(t, c.Fs().IsFile(dir))
	assert.False(t, c.Fs().IsFile(""))
}

func TestFs_Exists_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "exists.txt")
	c.Fs().Write(path, "yes")
	assert.True(t, c.Fs().Exists(path))
	assert.True(t, c.Fs().Exists(dir))
	assert.False(t, c.Fs().Exists(Path(dir, "nope")))
}

func TestFs_List_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	c.Fs().Write(Path(dir, "a.txt"), "a")
	c.Fs().Write(Path(dir, "b.txt"), "b")
	r := c.Fs().List(dir)
	assert.True(t, r.OK)
	assert.Len(t, r.Value.([]fs.DirEntry), 2)
}

func TestFs_Stat_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stat.txt")
	c.Fs().Write(path, "data")
	r := c.Fs().Stat(path)
	assert.True(t, r.OK)
	assert.Equal(t, "stat.txt", r.Value.(fs.FileInfo).Name())
}

func TestFs_Open_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "open.txt")
	c.Fs().Write(path, "content")
	r := c.Fs().Open(path)
	assert.True(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_Create_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "created.txt")
	r := c.Fs().Create(path)
	assert.True(t, r.OK)
	WriteAll(r.Value, "hello")
	rr := c.Fs().Read(path)
	assert.Equal(t, "hello", rr.Value.(string))
}

func TestFs_Append_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "append.txt")
	c.Fs().Write(path, "first")
	r := c.Fs().Append(path)
	assert.True(t, r.OK)
	WriteAll(r.Value, " second")
	rr := c.Fs().Read(path)
	assert.Equal(t, "first second", rr.Value.(string))
}

func TestFs_ReadStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stream.txt")
	c.Fs().Write(path, "streamed")
	r := c.Fs().ReadStream(path)
	assert.True(t, r.OK)
	CloseStream(r.Value)
}

func TestFs_WriteStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "ws.txt")
	r := c.Fs().WriteStream(path)
	assert.True(t, r.OK)
	WriteAll(r.Value, "stream")
}

func TestFs_Delete_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "delete.txt")
	c.Fs().Write(path, "gone")
	assert.True(t, c.Fs().Delete(path).OK)
	assert.False(t, c.Fs().Exists(path))
}

func TestFs_DeleteAll_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	sub := Path(dir, "deep", "nested")
	c.Fs().EnsureDir(sub)
	c.Fs().Write(Path(sub, "file.txt"), "data")
	assert.True(t, c.Fs().DeleteAll(Path(dir, "deep")).OK)
	assert.False(t, c.Fs().Exists(Path(dir, "deep")))
}

func TestFs_Rename_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	old := Path(dir, "old.txt")
	nw := Path(dir, "new.txt")
	c.Fs().Write(old, "data")
	assert.True(t, c.Fs().Rename(old, nw).OK)
	assert.False(t, c.Fs().Exists(old))
	assert.True(t, c.Fs().Exists(nw))
}

func TestFs_WriteMode_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "secret.txt")
	assert.True(t, c.Fs().WriteMode(path, "secret", 0600).OK)
	r := c.Fs().Stat(path)
	assert.True(t, r.OK)
	assert.Equal(t, "secret.txt", r.Value.(fs.FileInfo).Name())
}

// --- Zero Value ---

func TestFs_ZeroValue_Good(t *testing.T) {
	dir := t.TempDir()
	zeroFs := &Fs{}

	path := Path(dir, "zero.txt")
	assert.True(t, zeroFs.Write(path, "zero value works").OK)
	r := zeroFs.Read(path)
	assert.True(t, r.OK)
	assert.Equal(t, "zero value works", r.Value.(string))
	assert.True(t, zeroFs.IsFile(path))
	assert.True(t, zeroFs.Exists(path))
	assert.True(t, zeroFs.IsDir(dir))
}

func TestFs_ZeroValue_List_Good(t *testing.T) {
	dir := t.TempDir()
	zeroFs := &Fs{}

	(&Fs{}).New("/").Write(Path(dir, "a.txt"), "a")
	r := zeroFs.List(dir)
	assert.True(t, r.OK)
	entries := r.Value.([]fs.DirEntry)
	assert.Len(t, entries, 1)
}

func TestFs_Exists_NotFound_Bad(t *testing.T) {
	c := New()
	assert.False(t, c.Fs().Exists("/nonexistent/path/xyz"))
}

// --- Fs path/validatePath edge cases ---

func TestFs_Read_EmptyPath_Ugly(t *testing.T) {
	c := New()
	r := c.Fs().Read("")
	assert.False(t, r.OK)
}

func TestFs_Write_EmptyPath_Ugly(t *testing.T) {
	c := New()
	r := c.Fs().Write("", "data")
	assert.False(t, r.OK)
}

func TestFs_Delete_Protected_Ugly(t *testing.T) {
	c := New()
	r := c.Fs().Delete("/")
	assert.False(t, r.OK)
}

func TestFs_DeleteAll_Protected_Ugly(t *testing.T) {
	c := New()
	r := c.Fs().DeleteAll("/")
	assert.False(t, r.OK)
}

func TestFs_ReadStream_WriteStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "stream.txt")
	c.Fs().Write(path, "streamed")

	r := c.Fs().ReadStream(path)
	assert.True(t, r.OK)

	w := c.Fs().WriteStream(path)
	assert.True(t, w.OK)
}

// --- WriteAtomic ---

func TestFs_WriteAtomic_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "status.json")
	r := c.Fs().WriteAtomic(path, `{"status":"completed"}`)
	assert.True(t, r.OK)

	read := c.Fs().Read(path)
	assert.True(t, read.OK)
	assert.Equal(t, `{"status":"completed"}`, read.Value)
}

func TestFs_WriteAtomic_Good_Overwrite(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "data.txt")
	c.Fs().WriteAtomic(path, "first")
	c.Fs().WriteAtomic(path, "second")

	read := c.Fs().Read(path)
	assert.Equal(t, "second", read.Value)
}

func TestFs_WriteAtomic_Bad_ReadOnlyDir(t *testing.T) {
	// Write to a non-existent root that can't be created
	m := (&Fs{}).New("/proc/nonexistent")
	r := m.WriteAtomic("file.txt", "data")
	assert.False(t, r.OK, "WriteAtomic must fail when parent dir cannot be created")
}

func TestFs_WriteAtomic_Ugly_NoTempFileLeftOver(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "clean.txt")
	c.Fs().WriteAtomic(path, "content")

	// Check no .tmp files remain
	lr := c.Fs().List(dir)
	entries, _ := lr.Value.([]fs.DirEntry)
	for _, e := range entries {
		assert.False(t, Contains(e.Name(), ".tmp."), "temp file should not remain after successful atomic write")
	}
}

func TestFs_WriteAtomic_Good_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := Path(dir, "sub", "dir", "file.txt")
	r := c.Fs().WriteAtomic(path, "nested")
	assert.True(t, r.OK)

	read := c.Fs().Read(path)
	assert.Equal(t, "nested", read.Value)
}

// --- NewUnrestricted ---

func TestFs_NewUnrestricted_Good(t *testing.T) {
	sandboxed := (&Fs{}).New(t.TempDir())
	unrestricted := sandboxed.NewUnrestricted()
	assert.Equal(t, "/", unrestricted.Root())
}

func TestFs_NewUnrestricted_Good_CanReadOutsideSandbox(t *testing.T) {
	dir := t.TempDir()
	outside := Path(dir, "outside.txt")
	(&Fs{}).New("/").Write(outside, "hello")

	sandboxed := (&Fs{}).New(Path(dir, "sandbox"))
	unrestricted := sandboxed.NewUnrestricted()

	r := unrestricted.Read(outside)
	assert.True(t, r.OK, "unrestricted Fs must read paths outside the original sandbox")
	assert.Equal(t, "hello", r.Value)
}

func TestFs_NewUnrestricted_Ugly_OriginalStaysSandboxed(t *testing.T) {
	dir := t.TempDir()
	sandbox := Path(dir, "sandbox")
	(&Fs{}).New("/").EnsureDir(sandbox)

	sandboxed := (&Fs{}).New(sandbox)
	_ = sandboxed.NewUnrestricted() // getting unrestricted doesn't affect original

	assert.Equal(t, sandbox, sandboxed.Root(), "original Fs must remain sandboxed")
}

// --- Root ---

func TestFs_Root_Good(t *testing.T) {
	m := (&Fs{}).New("/home/agent")
	assert.Equal(t, "/home/agent", m.Root())
}

func TestFs_Root_Good_Default(t *testing.T) {
	m := (&Fs{}).New("")
	assert.Equal(t, "/", m.Root())
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
func walkSeqSeed(t *testing.T, dir string) {
	t.Helper()
	c := New()
	c.Fs().Write(Path(dir, "a.txt"), "alpha")
	c.Fs().Write(Path(dir, "sub", "b.txt"), "bravo")
	c.Fs().Write(Path(dir, "vendor", "skipme.txt"), "skip")
	c.Fs().Write(Path(dir, ".git", "HEAD"), "ref: refs/heads/main")
}

func TestFs_WalkSeq_Good(t *testing.T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	seen := map[string]bool{}
	for entry, err := range c.Fs().WalkSeq(dir) {
		assert.NoError(t, err)
		seen[entry.Name] = true
	}
	assert.True(t, seen["a.txt"])
	assert.True(t, seen["b.txt"])
	assert.True(t, seen["sub"])
	assert.True(t, seen["vendor"])
	assert.True(t, seen["skipme.txt"])
	assert.True(t, seen[".git"])
	assert.True(t, seen["HEAD"])
}

func TestFs_WalkSeq_Good_BreakStops(t *testing.T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	count := 0
	for entry, err := range c.Fs().WalkSeq(dir) {
		assert.NoError(t, err)
		_ = entry
		count++
		if count == 2 {
			break
		}
	}
	assert.Equal(t, 2, count)
}

func TestFs_WalkSeq_Good_RelPath(t *testing.T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	rels := []string{}
	for entry, err := range c.Fs().WalkSeq(dir) {
		assert.NoError(t, err)
		if !entry.IsDir {
			rels = append(rels, entry.Path)
		}
	}
	// Paths are relative to the walk root, never start with the absolute dir.
	for _, p := range rels {
		assert.False(t, len(p) > 0 && p[0] == '/', "rel path leaked absolute: %s", p)
	}
}

func TestFs_WalkSeqSkip_Good(t *testing.T) {
	dir := t.TempDir()
	walkSeqSeed(t, dir)
	c := New()
	seen := map[string]bool{}
	for entry, err := range c.Fs().WalkSeqSkip(dir, "vendor", ".git") {
		assert.NoError(t, err)
		seen[entry.Name] = true
	}
	assert.True(t, seen["a.txt"])
	assert.True(t, seen["sub"])
	assert.True(t, seen["b.txt"])
	// Skipped directory names themselves still appear (as the entry where
	// skip is decided), but their contents do not.
	assert.False(t, seen["skipme.txt"], "vendor contents should be skipped")
	assert.False(t, seen["HEAD"], ".git contents should be skipped")
}

func TestFs_WalkSeqSkip_Good_RootBasenameNotSkipped(t *testing.T) {
	// If the user explicitly walks a directory whose basename matches a
	// skipName, the walk root itself is still entered.
	parent := t.TempDir()
	c := New()
	c.Fs().Write(Path(parent, "vendor", "kept.txt"), "kept")
	seen := map[string]bool{}
	for entry, err := range c.Fs().WalkSeqSkip(Path(parent, "vendor"), "vendor") {
		assert.NoError(t, err)
		seen[entry.Name] = true
	}
	assert.True(t, seen["kept.txt"], "walk root must not be self-skipped")
}

func TestFs_WalkSeq_Good_SandboxContainment(t *testing.T) {
	// Fs sandboxed to a temp dir; walking ".." must NOT escape — every
	// yielded entry path must remain inside the sandbox root.
	root := t.TempDir()
	sandbox := (&Fs{}).New(root)
	c := New()
	c.Fs().Write(Path(root, "inside.txt"), "ok")

	for entry, err := range sandbox.WalkSeq("..") {
		assert.NoError(t, err)
		// Entry paths are relative to the (resolved) walk root; they
		// must never start with ".." which would indicate escape.
		assert.False(t, len(entry.Path) >= 2 && entry.Path[:2] == "..",
			"sandbox escape: %s", entry.Path)
	}
}

func TestFs_WalkSeq_Good_FileMode(t *testing.T) {
	dir := t.TempDir()
	c := New()
	c.Fs().WriteMode(Path(dir, "secret.key"), "x", 0o600)
	for entry, err := range c.Fs().WalkSeq(dir) {
		assert.NoError(t, err)
		if entry.Name == "secret.key" {
			assert.Equal(t, fs.FileMode(0o600), entry.Mode.Perm())
		}
	}
}
