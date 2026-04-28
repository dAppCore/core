// SPDX-License-Identifier: EUPL-1.2

package core_test

import . "dappco.re/go"

func TestPath_Relative(t *T) {
	home := Env("DIR_HOME")

	ds := Env("DS")
	AssertEqual(t, home+ds+"Code"+ds+".core", Path("Code", ".core"))
}

func TestPath_Absolute(t *T) {
	ds := Env("DS")
	AssertEqual(t, "/tmp"+ds+"workspace", Path("/tmp", "workspace"))
}

func TestPath_Empty(t *T) {
	home := Env("DIR_HOME")

	AssertEqual(t, home, Path())
}

func TestPath_Cleans(t *T) {
	home := Env("DIR_HOME")

	AssertEqual(t, home+Env("DS")+"Code", Path("Code", "sub", ".."))
}

func TestPath_CleanDoubleSlash(t *T) {
	ds := Env("DS")
	AssertEqual(t, ds+"tmp"+ds+"file", Path("/tmp//file"))
}

func TestPath_PathBase(t *T) {
	AssertEqual(t, "core", PathBase("/Users/snider/Code/core"))
	AssertEqual(t, "homelab", PathBase("deploy/to/homelab"))
}

func TestPath_PathBase_Root(t *T) {
	AssertEqual(t, "/", PathBase("/"))
}

func TestPath_PathBase_Empty(t *T) {
	AssertEqual(t, ".", PathBase(""))
}

func TestPath_PathDir(t *T) {
	AssertEqual(t, "/Users/snider/Code", PathDir("/Users/snider/Code/core"))
}

func TestPath_PathDir_Root(t *T) {
	AssertEqual(t, "/", PathDir("/file"))
}

func TestPath_PathDir_NoDir(t *T) {
	AssertEqual(t, ".", PathDir("file.go"))
}

func TestPath_PathExt(t *T) {
	AssertEqual(t, ".go", PathExt("main.go"))
	AssertEqual(t, "", PathExt("Makefile"))
	AssertEqual(t, ".gz", PathExt("archive.tar.gz"))
}

func TestPath_EnvConsistency(t *T) {
	AssertEqual(t, Env("DIR_HOME"), Path())
}

func TestPath_PathGlob_Good(t *T) {
	dir := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(dir, "a.txt"), "a")
	f.Write(Path(dir, "b.txt"), "b")
	f.Write(Path(dir, "c.log"), "c")

	matches := PathGlob(Path(dir, "*.txt"))
	AssertLen(t, matches, 2)
}

func TestPath_PathGlob_NoMatch(t *T) {
	matches := PathGlob("/nonexistent/pattern-*.xyz")
	AssertEmpty(t, matches)
}

func TestPath_PathIsAbs_Good(t *T) {
	AssertTrue(t, PathIsAbs("/tmp"))
	AssertTrue(t, PathIsAbs("/"))
	AssertFalse(t, PathIsAbs("relative"))
	AssertFalse(t, PathIsAbs(""))
}

func TestPath_CleanPath_Good(t *T) {
	AssertEqual(t, "/a/b", CleanPath("/a//b", "/"))
	AssertEqual(t, "/a/c", CleanPath("/a/b/../c", "/"))
	AssertEqual(t, "/", CleanPath("/", "/"))
	AssertEqual(t, ".", CleanPath("", "/"))
}

func TestPath_PathDir_TrailingSlash(t *T) {
	result := PathDir("/Users/snider/Code/")
	AssertEqual(t, "/Users/snider/Code", result)
}

// --- PathRel ---

func TestPath_PathRel_Good_Descendant(t *T) {
	r := PathRel("/var/lib/foo", "/var/lib/foo/bar/baz")
	AssertTrue(t, r.OK)
	AssertEqual(t, "bar/baz", r.Value.(string))
}

func TestPath_PathRel_Good_Sibling(t *T) {
	r := PathRel("/a", "/b")
	AssertTrue(t, r.OK)
	AssertEqual(t, "../b", r.Value.(string))
}

func TestPath_PathRel_Good_Identical(t *T) {
	r := PathRel("/x/y", "/x/y")
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(string))
}

func TestPath_PathRel_Bad_MixedAbsRel(t *T) {
	// filepath.Rel rejects when one is absolute and the other relative.
	r := PathRel("/abs/path", "rel/path")
	AssertFalse(t, r.OK)
}

// --- PathAbs ---

func TestPath_PathAbs_Good_AlreadyAbsolute(t *T) {
	r := PathAbs("/already/absolute")
	AssertTrue(t, r.OK)
	AssertEqual(t, "/already/absolute", r.Value.(string))
}

func TestPath_PathAbs_Good_Relative(t *T) {
	r := PathAbs("./relative/path")
	AssertTrue(t, r.OK)
	abs := r.Value.(string)
	// Resolved against cwd — must end in the relative tail and be absolute.
	AssertTrue(t, len(abs) > 0 && abs[0] == '/', "PathAbs must return absolute: %s", abs)
	AssertContains(t, abs, "relative/path")
}

// --- AX-7 canonical triplets ---

func TestPath_Path_Good(t *T) {
	home := Env("DIR_HOME")
	ds := Env("DS")
	AssertEqual(t, home+ds+"Code"+ds+"core", Path("Code", "core"))
}

func TestPath_Path_Bad(t *T) {
	AssertEqual(t, Env("DIR_HOME"), Path())
}

func TestPath_Path_Ugly(t *T) {
	home := Env("DIR_HOME")
	ds := Env("DS")
	AssertEqual(t, home+ds+"agent", Path("workspace", "..", "agent"))
}

func TestPath_PathJoin_Good(t *T) {
	AssertEqual(t, "deploy/to/homelab", PathToSlash(PathJoin("deploy", "to", "homelab")))
}

func TestPath_PathJoin_Bad(t *T) {
	AssertEqual(t, "", PathJoin())
}

func TestPath_PathJoin_Ugly(t *T) {
	AssertEqual(t, "agent", PathJoin("deploy", "..", "agent"))
}

func TestPath_PathBase_Good(t *T) {
	AssertEqual(t, "agent.json", PathBase("/srv/dappcore/agent.json"))
}

func TestPath_PathBase_Bad(t *T) {
	AssertEqual(t, ".", PathBase(""))
}

func TestPath_PathBase_Ugly(t *T) {
	AssertEqual(t, string(PathSeparator), PathBase(string(PathSeparator)))
}

func TestPath_PathDir_Good(t *T) {
	AssertEqual(t, "/srv/dappcore", PathDir("/srv/dappcore/agent.json"))
}

func TestPath_PathDir_Bad(t *T) {
	AssertEqual(t, ".", PathDir("agent.json"))
}

func TestPath_PathDir_Ugly(t *T) {
	AssertEqual(t, string(PathSeparator), PathDir(string(PathSeparator)+"agent.json"))
}

func TestPath_PathExt_Good(t *T) {
	AssertEqual(t, ".json", PathExt("agent.config.json"))
}

func TestPath_PathExt_Bad(t *T) {
	AssertEqual(t, "", PathExt("Makefile"))
}

func TestPath_PathExt_Ugly(t *T) {
	AssertEqual(t, "", PathExt(".env"))
}

func TestPath_PathIsAbs_Bad(t *T) {
	AssertFalse(t, PathIsAbs("tmp/agent"))
}

func TestPath_PathIsAbs_Ugly(t *T) {
	AssertTrue(t, PathIsAbs(`C:\agent\workspace`))
	AssertFalse(t, PathIsAbs(""))
}

func TestPath_CleanPath_Bad(t *T) {
	AssertEqual(t, ".", CleanPath("", "/"))
}

func TestPath_CleanPath_Ugly(t *T) {
	AssertEqual(t, "/agent", CleanPath("/../../agent", "/"))
}

func TestPath_PathGlob_Bad(t *T) {
	AssertEmpty(t, PathGlob("["))
}

func TestPath_PathGlob_Ugly(t *T) {
	dir := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(dir, "agent one.txt"), "one")
	matches := PathGlob(Path(dir, "agent*.txt"))
	AssertLen(t, matches, 1)
	AssertContains(t, matches[0], "agent one.txt")
}

func TestPath_PathMatch_Good(t *T) {
	r := PathMatch("agent.?", "agent.a")
	AssertTrue(t, r.OK)
	AssertTrue(t, r.Value.(bool))
}

func TestPath_PathMatch_Bad(t *T) {
	r := PathMatch("[", "agent")
	AssertFalse(t, r.OK)
}

func TestPath_PathMatch_Ugly(t *T) {
	r := PathMatch("agent.*", "agent.")
	AssertTrue(t, r.OK)
	AssertTrue(t, r.Value.(bool))
}

func TestPath_PathRel_Good(t *T) {
	r := PathRel("/srv/dappcore", "/srv/dappcore/agent/config.json")
	AssertTrue(t, r.OK)
	AssertEqual(t, "agent/config.json", r.Value.(string))
}

func TestPath_PathRel_Bad(t *T) {
	r := PathRel("/srv/dappcore", "agent/config.json")
	AssertFalse(t, r.OK)
}

func TestPath_PathRel_Ugly(t *T) {
	r := PathRel("/srv/dappcore", "/srv/dappcore")
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(string))
}

func TestPath_PathAbs_Good(t *T) {
	r := PathAbs(".")
	AssertTrue(t, r.OK)
	AssertTrue(t, PathIsAbs(r.Value.(string)))
}

func TestPath_PathAbs_Bad(t *T) {
	r := PathAbs("agent/workspace")
	AssertTrue(t, r.OK, "PathAbs is infallible for ordinary relative input")
	AssertTrue(t, PathIsAbs(r.Value.(string)))
}

func TestPath_PathAbs_Ugly(t *T) {
	r := PathAbs("")
	AssertTrue(t, r.OK)
	AssertTrue(t, PathIsAbs(r.Value.(string)))
}

func TestPath_PathEvalSymlinks_Good(t *T) {
	dir := t.TempDir()
	r := PathEvalSymlinks(dir)
	AssertTrue(t, r.OK)
	AssertTrue(t, PathIsAbs(r.Value.(string)))
	AssertEqual(t, PathBase(dir), PathBase(r.Value.(string)))
}

func TestPath_PathEvalSymlinks_Bad(t *T) {
	r := PathEvalSymlinks(Path(t.TempDir(), "missing"))
	AssertFalse(t, r.OK)
}

func TestPath_PathEvalSymlinks_Ugly(t *T) {
	dir := t.TempDir()
	file := Path(dir, "agent.txt")
	(&Fs{}).New("/").Write(file, "ready")
	r := PathEvalSymlinks(file)
	AssertTrue(t, r.OK)
	AssertTrue(t, PathIsAbs(r.Value.(string)))
	AssertEqual(t, "agent.txt", PathBase(r.Value.(string)))
}

func TestPath_PathToSlash_Good(t *T) {
	AssertEqual(t, "deploy/to/homelab", PathToSlash(PathJoin("deploy", "to", "homelab")))
}

func TestPath_PathToSlash_Bad(t *T) {
	AssertEqual(t, "agent", PathToSlash("agent"))
}

func TestPath_PathToSlash_Ugly(t *T) {
	AssertEqual(t, "agent/dispatch", PathToSlash("agent/dispatch"))
}

func TestPath_PathWalk_Good(t *T) {
	dir := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(dir, "agent.txt"), "ready")
	visited := 0
	err := PathWalk(dir, func(path string, _ FsFileInfo, err error) error {
		AssertNoError(t, err)
		if PathBase(path) == "agent.txt" {
			visited++
		}
		return nil
	})
	AssertNoError(t, err)
	AssertEqual(t, 1, visited)
}

func TestPath_PathWalk_Bad(t *T) {
	err := PathWalk(Path(t.TempDir(), "missing"), func(path string, _ FsFileInfo, err error) error {
		return err
	})
	AssertError(t, err)
}

func TestPath_PathWalk_Ugly(t *T) {
	dir := t.TempDir()
	f := (&Fs{}).New("/")
	f.EnsureDir(Path(dir, "skip"))
	f.Write(Path(dir, "skip/hidden.txt"), "hidden")
	visitedHidden := false
	err := PathWalk(dir, func(path string, info FsFileInfo, err error) error {
		AssertNoError(t, err)
		if info.IsDir() && PathBase(path) == "skip" {
			return PathSkipDir
		}
		if PathBase(path) == "hidden.txt" {
			visitedHidden = true
		}
		return nil
	})
	AssertNoError(t, err)
	AssertFalse(t, visitedHidden)
}

func TestPath_PathWalkDir_Good(t *T) {
	dir := t.TempDir()
	(&Fs{}).New("/").Write(Path(dir, "agent.txt"), "ready")
	visited := 0
	err := PathWalkDir(dir, func(path string, d FsDirEntry, err error) error {
		AssertNoError(t, err)
		if !d.IsDir() && PathBase(path) == "agent.txt" {
			visited++
		}
		return nil
	})
	AssertNoError(t, err)
	AssertEqual(t, 1, visited)
}

func TestPath_PathWalkDir_Bad(t *T) {
	err := PathWalkDir(Path(t.TempDir(), "missing"), func(path string, d FsDirEntry, err error) error {
		return err
	})
	AssertError(t, err)
}

func TestPath_PathWalkDir_Ugly(t *T) {
	dir := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(dir, "first.txt"), "first")
	f.Write(Path(dir, "second.txt"), "second")
	visited := 0
	err := PathWalkDir(dir, func(path string, d FsDirEntry, err error) error {
		AssertNoError(t, err)
		visited++
		if !d.IsDir() {
			return PathSkipAll
		}
		return nil
	})
	AssertNoError(t, err)
	AssertLess(t, visited, 4)
}

func TestPath_PathChangeExt_Good(t *T) {
	AssertEqual(t, "agent.yaml", PathChangeExt("agent.json", "yaml"))
}

func TestPath_PathChangeExt_Bad(t *T) {
	AssertEqual(t, "README.md", PathChangeExt("README", ".md"))
}

func TestPath_PathChangeExt_Ugly(t *T) {
	AssertEqual(t, "agent", PathChangeExt("agent.json", ""))
}
