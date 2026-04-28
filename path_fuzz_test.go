// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func FuzzPathRel(f *F) {
	f.Add("/var/lib/foo", "/var/lib/foo/bar")
	f.Add("/a", "/b")
	f.Add("relative/base", "relative/base/file")
	f.Add("base", "../target")
	f.Add("C:\\foo", "C:\\foo\\bar")
	f.Add("", "")

	f.Fuzz(func(t *T, base, target string) {
		r := PathRel(base, target)
		if r.OK {
			_ = r.Value.(string)
		}
	})
}

func FuzzPathAbs(f *F) {
	f.Add("/var/lib/foo")
	f.Add("./bar")
	f.Add("../../etc")
	f.Add("/")
	f.Add("a/./b/../c")
	f.Add("C:\\foo")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		r := PathAbs(raw)
		if r.OK {
			abs := r.Value.(string)
			if abs == "" {
				t.Errorf("PathAbs OK with empty value raw=%q", raw)
			}
		}
	})
}

func FuzzCleanPath(f *F) {
	f.Add("/var/lib/foo")
	f.Add("./bar")
	f.Add("../../etc")
	f.Add("/")
	f.Add("a/./b/../c")
	f.Add("C:\\foo")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		cleaned := CleanPath(raw, Env("DS"))
		if cleaned == "" {
			t.Errorf("CleanPath returned empty path raw=%q", raw)
		}
		_ = CleanPath(raw, "/")
		_ = CleanPath(raw, "\\")
	})
}

func FuzzPathBase(f *F) {
	f.Add("/var/lib/foo")
	f.Add("./bar")
	f.Add("../../etc")
	f.Add("/")
	f.Add("a/./b/../c")
	f.Add("C:\\foo")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		_ = Concat(PathBase(raw), "")
	})
}

func FuzzPathDir(f *F) {
	f.Add("/var/lib/foo")
	f.Add("./bar")
	f.Add("../../etc")
	f.Add("/")
	f.Add("a/./b/../c")
	f.Add("C:\\foo")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		_ = Concat(PathDir(raw), "")
	})
}

func FuzzPathExt(f *F) {
	f.Add("/var/lib/foo")
	f.Add("./bar")
	f.Add("../../etc")
	f.Add("/")
	f.Add("a/./b/../c")
	f.Add("C:\\foo")
	f.Add("")

	f.Fuzz(func(t *T, raw string) {
		_ = Concat(PathExt(raw), "")
	})
}
