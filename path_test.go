// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"
	"path/filepath"
	"testing"

	core "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath_Relative(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	ds := core.Env("DS")
	assert.Equal(t, home+ds+"Code"+ds+".core", core.Path("Code", ".core"))
}

func TestPath_Absolute(t *testing.T) {
	ds := core.Env("DS")
	assert.Equal(t, "/tmp"+ds+"workspace", core.Path("/tmp", "workspace"))
}

func TestPath_Empty(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home, core.Path())
}

func TestPath_Cleans(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, home+core.Env("DS")+"Code", core.Path("Code", "sub", ".."))
}

func TestPath_CleanDoubleSlash(t *testing.T) {
	ds := core.Env("DS")
	assert.Equal(t, ds+"tmp"+ds+"file", core.Path("/tmp//file"))
}

func TestPathBase(t *testing.T) {
	assert.Equal(t, "core", core.PathBase("/Users/snider/Code/core"))
	assert.Equal(t, "homelab", core.PathBase("deploy/to/homelab"))
}

func TestPathBase_Root(t *testing.T) {
	assert.Equal(t, "/", core.PathBase("/"))
}

func TestPathBase_Empty(t *testing.T) {
	assert.Equal(t, ".", core.PathBase(""))
}

func TestPathDir(t *testing.T) {
	assert.Equal(t, "/Users/snider/Code", core.PathDir("/Users/snider/Code/core"))
}

func TestPathDir_Root(t *testing.T) {
	assert.Equal(t, "/", core.PathDir("/file"))
}

func TestPathDir_NoDir(t *testing.T) {
	assert.Equal(t, ".", core.PathDir("file.go"))
}

func TestPathExt(t *testing.T) {
	assert.Equal(t, ".go", core.PathExt("main.go"))
	assert.Equal(t, "", core.PathExt("Makefile"))
	assert.Equal(t, ".gz", core.PathExt("archive.tar.gz"))
}

func TestPath_EnvConsistency(t *testing.T) {
	assert.Equal(t, core.Env("DIR_HOME"), core.Path())
}

func TestPathGlob_Good(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "c.log"), []byte("c"), 0644)

	matches := core.PathGlob(filepath.Join(dir, "*.txt"))
	assert.Len(t, matches, 2)
}

func TestPathGlob_NoMatch(t *testing.T) {
	matches := core.PathGlob("/nonexistent/pattern-*.xyz")
	assert.Empty(t, matches)
}

func TestPathIsAbs_Good(t *testing.T) {
	assert.True(t, core.PathIsAbs("/tmp"))
	assert.True(t, core.PathIsAbs("/"))
	assert.False(t, core.PathIsAbs("relative"))
	assert.False(t, core.PathIsAbs(""))
}

func TestCleanPath_Good(t *testing.T) {
	assert.Equal(t, "/a/b", core.CleanPath("/a//b", "/"))
	assert.Equal(t, "/a/c", core.CleanPath("/a/b/../c", "/"))
	assert.Equal(t, "/", core.CleanPath("/", "/"))
	assert.Equal(t, ".", core.CleanPath("", "/"))
}

func TestPathDir_TrailingSlash(t *testing.T) {
	result := core.PathDir("/Users/snider/Code/")
	assert.Equal(t, "/Users/snider/Code", result)
}
