package core_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Fs (Sandboxed Filesystem) ---

func TestFs_WriteRead_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "test.txt")
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
	path := filepath.Join(dir, "sub", "dir")
	assert.True(t, c.Fs().EnsureDir(path).OK)
	assert.True(t, c.Fs().IsDir(path))
}

func TestFs_IsDir_Good(t *testing.T) {
	c := New()
	dir := t.TempDir()
	assert.True(t, c.Fs().IsDir(dir))
	assert.False(t, c.Fs().IsDir(filepath.Join(dir, "nonexistent")))
	assert.False(t, c.Fs().IsDir(""))
}

func TestFs_IsFile_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "test.txt")
	c.Fs().Write(path, "data")
	assert.True(t, c.Fs().IsFile(path))
	assert.False(t, c.Fs().IsFile(dir))
	assert.False(t, c.Fs().IsFile(""))
}

func TestFs_Exists_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "exists.txt")
	c.Fs().Write(path, "yes")
	assert.True(t, c.Fs().Exists(path))
	assert.True(t, c.Fs().Exists(dir))
	assert.False(t, c.Fs().Exists(filepath.Join(dir, "nope")))
}

func TestFs_List_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	c.Fs().Write(filepath.Join(dir, "a.txt"), "a")
	c.Fs().Write(filepath.Join(dir, "b.txt"), "b")
	r := c.Fs().List(dir)
	assert.True(t, r.OK)
	assert.Len(t, r.Value.([]fs.DirEntry), 2)
}

func TestFs_Stat_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "stat.txt")
	c.Fs().Write(path, "data")
	r := c.Fs().Stat(path)
	assert.True(t, r.OK)
	assert.Equal(t, "stat.txt", r.Value.(os.FileInfo).Name())
}

func TestFs_Open_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "open.txt")
	c.Fs().Write(path, "content")
	r := c.Fs().Open(path)
	assert.True(t, r.OK)
	r.Value.(io.Closer).Close()
}

func TestFs_Create_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "sub", "created.txt")
	r := c.Fs().Create(path)
	assert.True(t, r.OK)
	w := r.Value.(io.WriteCloser)
	w.Write([]byte("hello"))
	w.Close()
	rr := c.Fs().Read(path)
	assert.Equal(t, "hello", rr.Value.(string))
}

func TestFs_Append_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "append.txt")
	c.Fs().Write(path, "first")
	r := c.Fs().Append(path)
	assert.True(t, r.OK)
	w := r.Value.(io.WriteCloser)
	w.Write([]byte(" second"))
	w.Close()
	rr := c.Fs().Read(path)
	assert.Equal(t, "first second", rr.Value.(string))
}

func TestFs_ReadStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "stream.txt")
	c.Fs().Write(path, "streamed")
	r := c.Fs().ReadStream(path)
	assert.True(t, r.OK)
	r.Value.(io.Closer).Close()
}

func TestFs_WriteStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "sub", "ws.txt")
	r := c.Fs().WriteStream(path)
	assert.True(t, r.OK)
	w := r.Value.(io.WriteCloser)
	w.Write([]byte("stream"))
	w.Close()
}

func TestFs_Delete_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "delete.txt")
	c.Fs().Write(path, "gone")
	assert.True(t, c.Fs().Delete(path).OK)
	assert.False(t, c.Fs().Exists(path))
}

func TestFs_DeleteAll_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	sub := filepath.Join(dir, "deep", "nested")
	c.Fs().EnsureDir(sub)
	c.Fs().Write(filepath.Join(sub, "file.txt"), "data")
	assert.True(t, c.Fs().DeleteAll(filepath.Join(dir, "deep")).OK)
	assert.False(t, c.Fs().Exists(filepath.Join(dir, "deep")))
}

func TestFs_Rename_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	old := filepath.Join(dir, "old.txt")
	nw := filepath.Join(dir, "new.txt")
	c.Fs().Write(old, "data")
	assert.True(t, c.Fs().Rename(old, nw).OK)
	assert.False(t, c.Fs().Exists(old))
	assert.True(t, c.Fs().Exists(nw))
}

func TestFs_WriteMode_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "secret.txt")
	assert.True(t, c.Fs().WriteMode(path, "secret", 0600).OK)
	r := c.Fs().Stat(path)
	assert.True(t, r.OK)
	assert.Equal(t, "secret.txt", r.Value.(os.FileInfo).Name())
}

// --- Zero Value ---

func TestFs_ZeroValue_Good(t *testing.T) {
	dir := t.TempDir()
	zeroFs := &Fs{}

	path := filepath.Join(dir, "zero.txt")
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

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
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
	path := filepath.Join(dir, "stream.txt")
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
	path := filepath.Join(dir, "status.json")
	r := c.Fs().WriteAtomic(path, `{"status":"completed"}`)
	assert.True(t, r.OK)

	read := c.Fs().Read(path)
	assert.True(t, read.OK)
	assert.Equal(t, `{"status":"completed"}`, read.Value)
}

func TestFs_WriteAtomic_Good_Overwrite(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "data.txt")
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
	path := filepath.Join(dir, "clean.txt")
	c.Fs().WriteAtomic(path, "content")

	// Check no .tmp files remain
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		assert.False(t, Contains(e.Name(), ".tmp."), "temp file should not remain after successful atomic write")
	}
}

func TestFs_WriteAtomic_Good_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "sub", "dir", "file.txt")
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
	outside := filepath.Join(dir, "outside.txt")
	os.WriteFile(outside, []byte("hello"), 0644)

	sandboxed := (&Fs{}).New(filepath.Join(dir, "sandbox"))
	unrestricted := sandboxed.NewUnrestricted()

	r := unrestricted.Read(outside)
	assert.True(t, r.OK, "unrestricted Fs must read paths outside the original sandbox")
	assert.Equal(t, "hello", r.Value)
}

func TestFs_NewUnrestricted_Ugly_OriginalStaysSandboxed(t *testing.T) {
	dir := t.TempDir()
	sandbox := filepath.Join(dir, "sandbox")
	os.MkdirAll(sandbox, 0755)

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
