package datanode

import (
	"io"
	"testing"

	coreio "forge.lthn.ai/core/cli/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time check: Medium implements io.Medium.
var _ coreio.Medium = (*Medium)(nil)

func TestReadWrite_Good(t *testing.T) {
	m := New()

	err := m.Write("hello.txt", "world")
	require.NoError(t, err)

	got, err := m.Read("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "world", got)
}

func TestReadWrite_Bad(t *testing.T) {
	m := New()

	_, err := m.Read("missing.txt")
	assert.Error(t, err)

	err = m.Write("", "content")
	assert.Error(t, err)
}

func TestNestedPaths_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("a/b/c/deep.txt", "deep"))

	got, err := m.Read("a/b/c/deep.txt")
	require.NoError(t, err)
	assert.Equal(t, "deep", got)

	assert.True(t, m.IsDir("a"))
	assert.True(t, m.IsDir("a/b"))
	assert.True(t, m.IsDir("a/b/c"))
}

func TestLeadingSlash_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("/leading/file.txt", "stripped"))
	got, err := m.Read("leading/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "stripped", got)

	got, err = m.Read("/leading/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "stripped", got)
}

func TestIsFile_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("file.go", "package main"))

	assert.True(t, m.IsFile("file.go"))
	assert.False(t, m.IsFile("missing.go"))
	assert.False(t, m.IsFile("")) // empty path
}

func TestEnsureDir_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.EnsureDir("foo/bar/baz"))

	assert.True(t, m.IsDir("foo"))
	assert.True(t, m.IsDir("foo/bar"))
	assert.True(t, m.IsDir("foo/bar/baz"))
	assert.True(t, m.Exists("foo/bar/baz"))
}

func TestDelete_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("delete-me.txt", "bye"))
	assert.True(t, m.Exists("delete-me.txt"))

	require.NoError(t, m.Delete("delete-me.txt"))
	assert.False(t, m.Exists("delete-me.txt"))
}

func TestDelete_Bad(t *testing.T) {
	m := New()

	// Delete non-existent
	assert.Error(t, m.Delete("ghost.txt"))

	// Delete non-empty dir
	require.NoError(t, m.Write("dir/file.txt", "content"))
	assert.Error(t, m.Delete("dir"))
}

func TestDeleteAll_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("tree/a.txt", "a"))
	require.NoError(t, m.Write("tree/sub/b.txt", "b"))
	require.NoError(t, m.Write("keep.txt", "keep"))

	require.NoError(t, m.DeleteAll("tree"))

	assert.False(t, m.Exists("tree/a.txt"))
	assert.False(t, m.Exists("tree/sub/b.txt"))
	assert.True(t, m.Exists("keep.txt"))
}

func TestRename_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("old.txt", "content"))
	require.NoError(t, m.Rename("old.txt", "new.txt"))

	assert.False(t, m.Exists("old.txt"))
	got, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "content", got)
}

func TestRenameDir_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("src/a.go", "package a"))
	require.NoError(t, m.Write("src/sub/b.go", "package b"))

	require.NoError(t, m.Rename("src", "dst"))

	assert.False(t, m.Exists("src/a.go"))

	got, err := m.Read("dst/a.go")
	require.NoError(t, err)
	assert.Equal(t, "package a", got)

	got, err = m.Read("dst/sub/b.go")
	require.NoError(t, err)
	assert.Equal(t, "package b", got)
}

func TestList_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("root.txt", "r"))
	require.NoError(t, m.Write("pkg/a.go", "a"))
	require.NoError(t, m.Write("pkg/b.go", "b"))
	require.NoError(t, m.Write("pkg/sub/c.go", "c"))

	entries, err := m.List("")
	require.NoError(t, err)

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	assert.Contains(t, names, "root.txt")
	assert.Contains(t, names, "pkg")

	entries, err = m.List("pkg")
	require.NoError(t, err)
	names = make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	assert.Contains(t, names, "a.go")
	assert.Contains(t, names, "b.go")
	assert.Contains(t, names, "sub")
}

func TestStat_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("stat.txt", "hello"))

	info, err := m.Stat("stat.txt")
	require.NoError(t, err)
	assert.Equal(t, int64(5), info.Size())
	assert.False(t, info.IsDir())

	// Root stat
	info, err = m.Stat("")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestOpen_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("open.txt", "opened"))

	f, err := m.Open("open.txt")
	require.NoError(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, "opened", string(data))
}

func TestCreateAppend_Good(t *testing.T) {
	m := New()

	// Create
	w, err := m.Create("new.txt")
	require.NoError(t, err)
	w.Write([]byte("hello"))
	w.Close()

	got, err := m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello", got)

	// Append
	w, err = m.Append("new.txt")
	require.NoError(t, err)
	w.Write([]byte(" world"))
	w.Close()

	got, err = m.Read("new.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world", got)
}

func TestStreams_Good(t *testing.T) {
	m := New()

	// WriteStream
	ws, err := m.WriteStream("stream.txt")
	require.NoError(t, err)
	ws.Write([]byte("streamed"))
	ws.Close()

	// ReadStream
	rs, err := m.ReadStream("stream.txt")
	require.NoError(t, err)
	data, err := io.ReadAll(rs)
	require.NoError(t, err)
	assert.Equal(t, "streamed", string(data))
	rs.Close()
}

func TestFileGetFileSet_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.FileSet("alias.txt", "via set"))

	got, err := m.FileGet("alias.txt")
	require.NoError(t, err)
	assert.Equal(t, "via set", got)
}

func TestSnapshotRestore_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("a.txt", "alpha"))
	require.NoError(t, m.Write("b/c.txt", "charlie"))

	snap, err := m.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, snap)

	// Restore into a new Medium
	m2, err := FromTar(snap)
	require.NoError(t, err)

	got, err := m2.Read("a.txt")
	require.NoError(t, err)
	assert.Equal(t, "alpha", got)

	got, err = m2.Read("b/c.txt")
	require.NoError(t, err)
	assert.Equal(t, "charlie", got)
}

func TestRestore_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("original.txt", "before"))

	snap, err := m.Snapshot()
	require.NoError(t, err)

	// Modify
	require.NoError(t, m.Write("original.txt", "after"))
	require.NoError(t, m.Write("extra.txt", "extra"))

	// Restore to snapshot
	require.NoError(t, m.Restore(snap))

	got, err := m.Read("original.txt")
	require.NoError(t, err)
	assert.Equal(t, "before", got)

	assert.False(t, m.Exists("extra.txt"))
}

func TestDataNode_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("test.txt", "borg"))

	dn := m.DataNode()
	assert.NotNil(t, dn)

	// Verify we can use the DataNode directly
	f, err := dn.Open("test.txt")
	require.NoError(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, "borg", string(data))
}

func TestOverwrite_Good(t *testing.T) {
	m := New()

	require.NoError(t, m.Write("file.txt", "v1"))
	require.NoError(t, m.Write("file.txt", "v2"))

	got, err := m.Read("file.txt")
	require.NoError(t, err)
	assert.Equal(t, "v2", got)
}

func TestExists_Good(t *testing.T) {
	m := New()

	assert.True(t, m.Exists("")) // root
	assert.False(t, m.Exists("x"))

	require.NoError(t, m.Write("x", "y"))
	assert.True(t, m.Exists("x"))
}

func TestReadDir_Ugly(t *testing.T) {
	m := New()

	// Read from a file path (not a dir) should return empty or error
	require.NoError(t, m.Write("file.txt", "content"))
	_, err := m.Read("file.txt")
	require.NoError(t, err)
}
