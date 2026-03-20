package core_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
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
