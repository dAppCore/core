// SPDX-License-Identifier: EUPL-1.2

// OS-aware filesystem path operations for the Core framework.
// Uses Env("DS") for the separator and Core string primitives
// for path manipulation. filepath imported only for PathGlob.
//
// Path anchors relative segments to DIR_HOME:
//
//	core.Path("Code", ".core")     // "/Users/snider/Code/.core"
//	core.Path("/tmp", "workspace") // "/tmp/workspace"
//	core.Path()                    // "/Users/snider"
//
// Path component helpers:
//
//	core.PathBase("/Users/snider/Code/core")  // "core"
//	core.PathDir("/Users/snider/Code/core")   // "/Users/snider/Code"
//	core.PathExt("main.go")                   // ".go"
package core

import "path/filepath"

// PathWalkFunc visits one path during a filesystem walk.
//
//	var fn core.PathWalkFunc
//	_ = fn
type PathWalkFunc = filepath.WalkFunc

// PathWalkDirFunc visits one path during a directory walk.
//
//	fn := func(path string, d core.FsDirEntry, err error) error { return err }
//	_ = fn
type PathWalkDirFunc = WalkDirFunc

var (
	// PathSkipDir tells PathWalkDir to skip the current directory.
	PathSkipDir = filepath.SkipDir

	// PathSkipAll tells PathWalkDir to stop walking immediately.
	PathSkipAll = filepath.SkipAll
)

// Path builds a clean, absolute filesystem path from segments.
// Uses Env("DS") for the OS directory separator.
// Relative paths are anchored to DIR_HOME. Absolute paths pass through.
//
//	core.Path("Code", ".core")      // "/Users/snider/Code/.core"
//	core.Path("/tmp", "workspace")  // "/tmp/workspace"
//	core.Path()                     // "/Users/snider"
func Path(segments ...string) string {
	ds := Env("DS")
	home := Env("DIR_HOME")
	if home == "" {
		home = "."
	}
	if len(segments) == 0 {
		return home
	}
	p := Join(ds, segments...)
	if PathIsAbs(p) {
		return CleanPath(p, ds)
	}
	return CleanPath(home+ds+p, ds)
}

// PathJoin joins path segments using OS-native filepath semantics.
// Unlike Path, it preserves relative paths.
//
//	core.PathJoin("deploy", "to", "homelab") // "deploy/to/homelab"
func PathJoin(segments ...string) string {
	return filepath.Join(segments...)
}

// PathBase returns the last element of a path.
//
//	core.PathBase("/Users/snider/Code/core")  // "core"
//	core.PathBase("deploy/to/homelab")        // "homelab"
func PathBase(p string) string {
	if p == "" {
		return "."
	}
	ds := Env("DS")
	p = TrimSuffix(p, ds)
	if p == "" {
		return ds
	}
	parts := Split(p, ds)
	return parts[len(parts)-1]
}

// PathDir returns all but the last element of a path.
//
//	core.PathDir("/Users/snider/Code/core")  // "/Users/snider/Code"
func PathDir(p string) string {
	if p == "" {
		return "."
	}
	ds := Env("DS")
	i := lastIndex(p, ds)
	if i < 0 {
		return "."
	}
	dir := p[:i]
	if dir == "" {
		return ds
	}
	return dir
}

// PathExt returns the file extension including the dot.
//
//	core.PathExt("main.go")   // ".go"
//	core.PathExt("Makefile")  // ""
func PathExt(p string) string {
	base := PathBase(p)
	i := lastIndex(base, ".")
	if i <= 0 {
		return ""
	}
	return base[i:]
}

// PathIsAbs returns true if the path is absolute.
// Handles Unix (starts with /) and Windows (drive letter like C:).
//
//	core.PathIsAbs("/tmp")     // true
//	core.PathIsAbs("C:\\tmp")  // true
//	core.PathIsAbs("relative") // false
func PathIsAbs(p string) bool {
	if p == "" {
		return false
	}
	if p[0] == '/' {
		return true
	}
	// Windows: C:\ or C:/
	if len(p) >= 3 && p[1] == ':' && (p[2] == '/' || p[2] == '\\') {
		return true
	}
	return false
}

// CleanPath removes redundant separators and resolves . and .. elements.
//
//	core.CleanPath("/tmp//file", "/")     // "/tmp/file"
//	core.CleanPath("a/b/../c", "/")       // "a/c"
func CleanPath(p, ds string) string {
	if p == "" {
		return "."
	}

	rooted := HasPrefix(p, ds)
	parts := Split(p, ds)
	var clean []string

	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			if len(clean) > 0 && clean[len(clean)-1] != ".." {
				clean = clean[:len(clean)-1]
			} else if !rooted {
				clean = append(clean, "..")
			}
		default:
			clean = append(clean, part)
		}
	}

	result := Join(ds, clean...)
	if rooted {
		result = ds + result
	}
	if result == "" {
		if rooted {
			return ds
		}
		return "."
	}
	return result
}

// PathGlob returns file paths matching a pattern.
//
//	core.PathGlob("/tmp/agent-*.log")
func PathGlob(pattern string) []string {
	matches, _ := filepath.Glob(pattern)
	return matches
}

// PathMatch reports whether name matches the shell file name pattern.
//
//	r := core.PathMatch("process.*", "process.run")
//	if r.OK && r.Value.(bool) { core.Println("matched") }
func PathMatch(pattern, name string) Result {
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return Result{err, false}
	}
	return Result{matched, true}
}

// PathRel returns target expressed as a path relative to base. Both
// arguments must be either both absolute or both relative; otherwise
// the call returns Result.OK=false with the underlying error.
//
//	core.PathRel("/var/lib/foo", "/var/lib/foo/bar/baz")  // Result.Value="bar/baz"
//	core.PathRel("/a", "/b")                              // Result.Value="../b"
func PathRel(base, target string) Result {
	r, err := filepath.Rel(base, target)
	if err != nil {
		return Result{err, false}
	}
	return Result{r, true}
}

// PathAbs returns the absolute representation of p, anchored to the
// current working directory when p is relative.
//
//	core.PathAbs("./relative/path")   // "/cwd/relative/path"
//	core.PathAbs("/already/absolute") // "/already/absolute"
func PathAbs(p string) Result {
	a, err := filepath.Abs(p)
	if err != nil {
		return Result{err, false}
	}
	return Result{a, true}
}

// PathEvalSymlinks returns p after resolving symbolic links.
//
//	r := core.PathEvalSymlinks("/tmp/current")
//	if r.OK { resolved := r.Value.(string); _ = resolved }
func PathEvalSymlinks(p string) Result {
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return Result{err, false}
	}
	return Result{resolved, true}
}

// PathToSlash converts OS-native separators in p to slash separators.
//
//	p := core.PathToSlash(core.PathJoin("templates", "agent"))
func PathToSlash(p string) string {
	return filepath.ToSlash(p)
}

// PathWalk walks the file tree rooted at root.
//
//	err := core.PathWalk("/tmp/workspace", fn)
func PathWalk(root string, fn PathWalkFunc) error {
	return filepath.Walk(root, fn)
}

// PathWalkDir walks the file tree rooted at root using directory entries.
//
//	err := core.PathWalkDir("/tmp/workspace", fn)
func PathWalkDir(root string, fn PathWalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

// PathChangeExt returns p with its file extension replaced by newExt.
// newExt may include a leading dot or omit it — both forms are accepted.
// If p has no extension, newExt is appended.
//
//	core.PathChangeExt("data.json", ".yaml")  // "data.yaml"
//	core.PathChangeExt("data.json", "yaml")   // "data.yaml"
//	core.PathChangeExt("README", ".md")       // "README.md"
func PathChangeExt(p, newExt string) string {
	if newExt != "" && newExt[0] != '.' {
		newExt = "." + newExt
	}
	ext := PathExt(p)
	if ext == "" {
		return p + newExt
	}
	return p[:len(p)-len(ext)] + newExt
}

// lastIndex returns the index of the last occurrence of substr in s, or -1.
func lastIndex(s, substr string) int {
	if substr == "" || s == "" {
		return -1
	}
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
