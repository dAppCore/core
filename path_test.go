// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

func TestPath_Relative(t *testing.T) {
	home := Env("DIR_HOME")

	ds := Env("DS")
	AssertEqual(t, home+ds+"Code"+ds+".core", Path("Code", ".core"))
}

func TestPath_Absolute(t *testing.T) {
	ds := Env("DS")
	AssertEqual(t, "/tmp"+ds+"workspace", Path("/tmp", "workspace"))
}

func TestPath_Empty(t *testing.T) {
	home := Env("DIR_HOME")

	AssertEqual(t, home, Path())
}

func TestPath_Cleans(t *testing.T) {
	home := Env("DIR_HOME")

	AssertEqual(t, home+Env("DS")+"Code", Path("Code", "sub", ".."))
}

func TestPath_CleanDoubleSlash(t *testing.T) {
	ds := Env("DS")
	AssertEqual(t, ds+"tmp"+ds+"file", Path("/tmp//file"))
}

func TestPath_PathBase(t *testing.T) {
	AssertEqual(t, "core", PathBase("/Users/snider/Code/core"))
	AssertEqual(t, "homelab", PathBase("deploy/to/homelab"))
}

func TestPath_PathBase_Root(t *testing.T) {
	AssertEqual(t, "/", PathBase("/"))
}

func TestPath_PathBase_Empty(t *testing.T) {
	AssertEqual(t, ".", PathBase(""))
}

func TestPath_PathDir(t *testing.T) {
	AssertEqual(t, "/Users/snider/Code", PathDir("/Users/snider/Code/core"))
}

func TestPath_PathDir_Root(t *testing.T) {
	AssertEqual(t, "/", PathDir("/file"))
}

func TestPath_PathDir_NoDir(t *testing.T) {
	AssertEqual(t, ".", PathDir("file.go"))
}

func TestPath_PathExt(t *testing.T) {
	AssertEqual(t, ".go", PathExt("main.go"))
	AssertEqual(t, "", PathExt("Makefile"))
	AssertEqual(t, ".gz", PathExt("archive.tar.gz"))
}

func TestPath_EnvConsistency(t *testing.T) {
	AssertEqual(t, Env("DIR_HOME"), Path())
}

func TestPath_PathGlob_Good(t *testing.T) {
	dir := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(dir, "a.txt"), "a")
	f.Write(Path(dir, "b.txt"), "b")
	f.Write(Path(dir, "c.log"), "c")

	matches := PathGlob(Path(dir, "*.txt"))
	AssertLen(t, matches, 2)
}

func TestPath_PathGlob_NoMatch(t *testing.T) {
	matches := PathGlob("/nonexistent/pattern-*.xyz")
	AssertEmpty(t, matches)
}

func TestPath_PathIsAbs_Good(t *testing.T) {
	AssertTrue(t, PathIsAbs("/tmp"))
	AssertTrue(t, PathIsAbs("/"))
	AssertFalse(t, PathIsAbs("relative"))
	AssertFalse(t, PathIsAbs(""))
}

func TestPath_CleanPath_Good(t *testing.T) {
	AssertEqual(t, "/a/b", CleanPath("/a//b", "/"))
	AssertEqual(t, "/a/c", CleanPath("/a/b/../c", "/"))
	AssertEqual(t, "/", CleanPath("/", "/"))
	AssertEqual(t, ".", CleanPath("", "/"))
}

func TestPath_PathDir_TrailingSlash(t *testing.T) {
	result := PathDir("/Users/snider/Code/")
	AssertEqual(t, "/Users/snider/Code", result)
}

// --- PathRel ---

func TestPath_PathRel_Good_Descendant(t *testing.T) {
	r := PathRel("/var/lib/foo", "/var/lib/foo/bar/baz")
	AssertTrue(t, r.OK)
	AssertEqual(t, "bar/baz", r.Value.(string))
}

func TestPath_PathRel_Good_Sibling(t *testing.T) {
	r := PathRel("/a", "/b")
	AssertTrue(t, r.OK)
	AssertEqual(t, "../b", r.Value.(string))
}

func TestPath_PathRel_Good_Identical(t *testing.T) {
	r := PathRel("/x/y", "/x/y")
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(string))
}

func TestPath_PathRel_Bad_MixedAbsRel(t *testing.T) {
	// filepath.Rel rejects when one is absolute and the other relative.
	r := PathRel("/abs/path", "rel/path")
	AssertFalse(t, r.OK)
}

// --- PathAbs ---

func TestPath_PathAbs_Good_AlreadyAbsolute(t *testing.T) {
	r := PathAbs("/already/absolute")
	AssertTrue(t, r.OK)
	AssertEqual(t, "/already/absolute", r.Value.(string))
}

func TestPath_PathAbs_Good_Relative(t *testing.T) {
	r := PathAbs("./relative/path")
	AssertTrue(t, r.OK)
	abs := r.Value.(string)
	// Resolved against cwd — must end in the relative tail and be absolute.
	AssertTrue(t, len(abs) > 0 && abs[0] == '/', "PathAbs must return absolute: %s", abs)
	AssertContains(t, abs, "relative/path")
}
