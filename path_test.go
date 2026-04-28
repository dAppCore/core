// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"testing"

	core "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestPath_Relative(t *testing.T) {
	home := core.Env("DIR_HOME")

	ds := core.Env("DS")
	assert.Equal(t, home+ds+"Code"+ds+".core", core.Path("Code", ".core"))
}

func TestPath_Absolute(t *testing.T) {
	ds := core.Env("DS")
	assert.Equal(t, "/tmp"+ds+"workspace", core.Path("/tmp", "workspace"))
}

func TestPath_Empty(t *testing.T) {
	home := core.Env("DIR_HOME")

	assert.Equal(t, home, core.Path())
}

func TestPath_Cleans(t *testing.T) {
	home := core.Env("DIR_HOME")

	assert.Equal(t, home+core.Env("DS")+"Code", core.Path("Code", "sub", ".."))
}

func TestPath_CleanDoubleSlash(t *testing.T) {
	ds := core.Env("DS")
	assert.Equal(t, ds+"tmp"+ds+"file", core.Path("/tmp//file"))
}

func TestPath_PathBase(t *testing.T) {
	assert.Equal(t, "core", core.PathBase("/Users/snider/Code/core"))
	assert.Equal(t, "homelab", core.PathBase("deploy/to/homelab"))
}

func TestPath_PathBase_Root(t *testing.T) {
	assert.Equal(t, "/", core.PathBase("/"))
}

func TestPath_PathBase_Empty(t *testing.T) {
	assert.Equal(t, ".", core.PathBase(""))
}

func TestPath_PathDir(t *testing.T) {
	assert.Equal(t, "/Users/snider/Code", core.PathDir("/Users/snider/Code/core"))
}

func TestPath_PathDir_Root(t *testing.T) {
	assert.Equal(t, "/", core.PathDir("/file"))
}

func TestPath_PathDir_NoDir(t *testing.T) {
	assert.Equal(t, ".", core.PathDir("file.go"))
}

func TestPath_PathExt(t *testing.T) {
	assert.Equal(t, ".go", core.PathExt("main.go"))
	assert.Equal(t, "", core.PathExt("Makefile"))
	assert.Equal(t, ".gz", core.PathExt("archive.tar.gz"))
}

func TestPath_EnvConsistency(t *testing.T) {
	assert.Equal(t, core.Env("DIR_HOME"), core.Path())
}

func TestPath_PathGlob_Good(t *testing.T) {
	dir := t.TempDir()
	f := (&core.Fs{}).New("/")
	f.Write(core.Path(dir, "a.txt"), "a")
	f.Write(core.Path(dir, "b.txt"), "b")
	f.Write(core.Path(dir, "c.log"), "c")

	matches := core.PathGlob(core.Path(dir, "*.txt"))
	assert.Len(t, matches, 2)
}

func TestPath_PathGlob_NoMatch(t *testing.T) {
	matches := core.PathGlob("/nonexistent/pattern-*.xyz")
	assert.Empty(t, matches)
}

func TestPath_PathIsAbs_Good(t *testing.T) {
	assert.True(t, core.PathIsAbs("/tmp"))
	assert.True(t, core.PathIsAbs("/"))
	assert.False(t, core.PathIsAbs("relative"))
	assert.False(t, core.PathIsAbs(""))
}

func TestPath_CleanPath_Good(t *testing.T) {
	assert.Equal(t, "/a/b", core.CleanPath("/a//b", "/"))
	assert.Equal(t, "/a/c", core.CleanPath("/a/b/../c", "/"))
	assert.Equal(t, "/", core.CleanPath("/", "/"))
	assert.Equal(t, ".", core.CleanPath("", "/"))
}

func TestPath_PathDir_TrailingSlash(t *testing.T) {
	result := core.PathDir("/Users/snider/Code/")
	assert.Equal(t, "/Users/snider/Code", result)
}

// --- PathRel ---

func TestPath_PathRel_Good_Descendant(t *testing.T) {
	r := core.PathRel("/var/lib/foo", "/var/lib/foo/bar/baz")
	assert.True(t, r.OK)
	assert.Equal(t, "bar/baz", r.Value.(string))
}

func TestPath_PathRel_Good_Sibling(t *testing.T) {
	r := core.PathRel("/a", "/b")
	assert.True(t, r.OK)
	assert.Equal(t, "../b", r.Value.(string))
}

func TestPath_PathRel_Good_Identical(t *testing.T) {
	r := core.PathRel("/x/y", "/x/y")
	assert.True(t, r.OK)
	assert.Equal(t, ".", r.Value.(string))
}

func TestPath_PathRel_Bad_MixedAbsRel(t *testing.T) {
	// filepath.Rel rejects when one is absolute and the other relative.
	r := core.PathRel("/abs/path", "rel/path")
	assert.False(t, r.OK)
}

// --- PathAbs ---

func TestPath_PathAbs_Good_AlreadyAbsolute(t *testing.T) {
	r := core.PathAbs("/already/absolute")
	assert.True(t, r.OK)
	assert.Equal(t, "/already/absolute", r.Value.(string))
}

func TestPath_PathAbs_Good_Relative(t *testing.T) {
	r := core.PathAbs("./relative/path")
	assert.True(t, r.OK)
	abs := r.Value.(string)
	// Resolved against cwd — must end in the relative tail and be absolute.
	assert.True(t, len(abs) > 0 && abs[0] == '/', "PathAbs must return absolute: %s", abs)
	assert.Contains(t, abs, "relative/path")
}
