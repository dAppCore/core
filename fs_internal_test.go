// SPDX-License-Identifier: EUPL-1.2

package core

import "os"

func TestFs_Fs_path_Good(t *T) {
	root := t.TempDir()
	fsys := (&Fs{}).New(root)

	AssertEqual(t, PathJoin(root, "logs", "agent.txt"), fsys.path("logs/agent.txt"))
}
func TestFs_Fs_path_Bad(t *T) {
	root := t.TempDir()
	fsys := (&Fs{}).New(root)

	AssertEqual(t, root, fsys.path(""))
}
func TestFs_Fs_path_Ugly(t *T) {
	cwd := Getwd()
	RequireTrue(t, cwd.OK)
	fsys := (&Fs{}).New("/")

	AssertEqual(t, PathJoin(cwd.Value.(string), "relative.txt"), fsys.path("relative.txt"))
}
func TestFs_Fs_validatePath_Good(t *T) {
	root := t.TempDir()
	fsys := (&Fs{}).New(root)

	r := fsys.validatePath("logs/agent.txt")

	AssertTrue(t, r.OK)
	AssertEqual(t, PathJoin(root, "logs", "agent.txt"), r.Value)
}
func TestFs_Fs_validatePath_Bad(t *T) {
	root := t.TempDir()
	outside := t.TempDir()
	RequireNoError(t, os.Symlink(outside, Path(root, "escape")))
	fsys := (&Fs{}).New(root)

	r := fsys.validatePath("escape/agent.txt")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "sandbox escape")
}
func TestFs_Fs_validatePath_Ugly(t *T) {
	root := t.TempDir()
	fsys := (&Fs{}).New(root)

	r := fsys.validatePath("missing/deep/agent.txt")

	AssertTrue(t, r.OK)
	AssertEqual(t, PathJoin(root, "missing", "deep", "agent.txt"), r.Value)
}
func TestFs_Fs_walkSeq_Good(t *T) {
	root := t.TempDir()
	fsys := (&Fs{}).New(root)
	RequireTrue(t, fsys.Write("agent.txt", "ready").OK)

	var names []string
	for entry, err := range fsys.walkSeq(".", nil) {
		RequireNoError(t, err)
		names = append(names, entry.Name)
	}

	AssertContains(t, names, "agent.txt")
}
func TestFs_Fs_walkSeq_Bad(t *T) {
	fsys := (&Fs{}).New(t.TempDir())

	var walkErr error
	for _, err := range fsys.walkSeq("missing", nil) {
		walkErr = err
		break
	}

	AssertError(t, walkErr)
}
func TestFs_Fs_walkSeq_Ugly(t *T) {
	root := t.TempDir()
	fsys := (&Fs{}).New(root)
	RequireTrue(t, fsys.Write("app/agent.txt", "ready").OK)
	RequireTrue(t, fsys.Write("vendor/agent.txt", "skip").OK)

	var paths []string
	for entry, err := range fsys.walkSeq(".", map[string]struct{}{"vendor": {}}) {
		RequireNoError(t, err)
		paths = append(paths, entry.Path)
	}

	AssertContains(t, paths, PathJoin("app", "agent.txt"))
	AssertNotContains(t, paths, PathJoin("vendor", "agent.txt"))
}
