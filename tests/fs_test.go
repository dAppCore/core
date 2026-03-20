package core_test

import (
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
	err := c.Fs().Write(path, "hello core")
	assert.NoError(t, err)

	content, err := c.Fs().Read(path)
	assert.NoError(t, err)
	assert.Equal(t, "hello core", content)
}

func TestFs_Read_Bad(t *testing.T) {
	c := New()
	_, err := c.Fs().Read("/nonexistent/path/to/file.txt")
	assert.Error(t, err)
}

func TestFs_EnsureDir_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()
	path := filepath.Join(dir, "sub", "dir")
	err := c.Fs().EnsureDir(path)
	assert.NoError(t, err)
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
	assert.False(t, c.Fs().IsFile(dir)) // dir, not file
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

	entries, err := c.Fs().List(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestFs_Stat_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "stat.txt")
	c.Fs().Write(path, "data")

	info, err := c.Fs().Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, "stat.txt", info.Name())
}

func TestFs_Open_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "open.txt")
	c.Fs().Write(path, "content")

	file, err := c.Fs().Open(path)
	assert.NoError(t, err)
	file.Close()
}

func TestFs_Create_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "sub", "created.txt")
	w, err := c.Fs().Create(path)
	assert.NoError(t, err)
	w.Write([]byte("hello"))
	w.Close()

	content, _ := c.Fs().Read(path)
	assert.Equal(t, "hello", content)
}

func TestFs_Append_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "append.txt")
	c.Fs().Write(path, "first")

	w, err := c.Fs().Append(path)
	assert.NoError(t, err)
	w.Write([]byte(" second"))
	w.Close()

	content, _ := c.Fs().Read(path)
	assert.Equal(t, "first second", content)
}

func TestFs_ReadStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "stream.txt")
	c.Fs().Write(path, "streamed")

	r, err := c.Fs().ReadStream(path)
	assert.NoError(t, err)
	r.Close()
}

func TestFs_WriteStream_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "sub", "ws.txt")
	w, err := c.Fs().WriteStream(path)
	assert.NoError(t, err)
	w.Write([]byte("stream"))
	w.Close()
}

func TestFs_Delete_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "delete.txt")
	c.Fs().Write(path, "gone")

	err := c.Fs().Delete(path)
	assert.NoError(t, err)
	assert.False(t, c.Fs().Exists(path))
}

func TestFs_DeleteAll_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	sub := filepath.Join(dir, "deep", "nested")
	c.Fs().EnsureDir(sub)
	c.Fs().Write(filepath.Join(sub, "file.txt"), "data")

	err := c.Fs().DeleteAll(filepath.Join(dir, "deep"))
	assert.NoError(t, err)
	assert.False(t, c.Fs().Exists(filepath.Join(dir, "deep")))
}

func TestFs_Rename_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	old := filepath.Join(dir, "old.txt")
	new := filepath.Join(dir, "new.txt")
	c.Fs().Write(old, "data")

	err := c.Fs().Rename(old, new)
	assert.NoError(t, err)
	assert.False(t, c.Fs().Exists(old))
	assert.True(t, c.Fs().Exists(new))
}

func TestFs_WriteMode_Good(t *testing.T) {
	dir := t.TempDir()
	c := New()

	path := filepath.Join(dir, "secret.txt")
	err := c.Fs().WriteMode(path, "secret", 0600)
	assert.NoError(t, err)

	info, _ := c.Fs().Stat(path)
	assert.Equal(t, "secret.txt", info.Name())
}
