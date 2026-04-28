// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"os"

	. "dappco.re/go"
)

func TestOs_FileMode_Good_Alias(t *T) {
	// FileMode alias must be assignable to/from os.FileMode.
	var coreMode FileMode = 0o644
	var osMode os.FileMode = coreMode
	AssertEqual(t, os.FileMode(0o644), osMode)
}

func TestOs_ModePerm_Good(t *T) {
	AssertEqual(t, FileMode(0o777), ModePerm)
}

func TestOs_ModeDir_Good(t *T) {
	mode := ModeDir | 0o755
	AssertTrue(t, mode.IsDir())
	AssertEqual(t, FileMode(0o755), mode.Perm())
}

func TestOs_Stdin_Good_NotNil(t *T) {
	AssertNotNil(t, Stdin())
}

func TestOs_Stdout_Good_NotNil(t *T) {
	AssertNotNil(t, Stdout())
}

func TestOs_Stderr_Good_NotNil(t *T) {
	AssertNotNil(t, Stderr())
}

func TestOs_Args_Good(t *T) {
	args := Args()

	AssertNotEmpty(t, args)
	AssertNotEmpty(t, args[0])
}

func TestOs_Args_Bad(t *T) {
	args := Args()

	AssertNotNil(t, args)
}

func TestOs_Args_Ugly(t *T) {
	first := Args()
	second := Args()

	AssertEqual(t, first[0], second[0])
}

func TestOs_Chdir_Good(t *T) {
	cwd := Getwd()
	RequireTrue(t, cwd.OK)
	defer func() { AssertTrue(t, Chdir(cwd.Value.(string)).OK) }()
	dir := t.TempDir()
	realDir := PathEvalSymlinks(dir)
	RequireTrue(t, realDir.OK)

	r := Chdir(dir)

	AssertTrue(t, r.OK)
	after := Getwd()
	AssertTrue(t, after.OK)
	AssertEqual(t, realDir.Value.(string), after.Value.(string))
}

func TestOs_Chdir_Bad(t *T) {
	r := Chdir(Path(t.TempDir(), "missing"))

	AssertFalse(t, r.OK)
}

func TestOs_Chdir_Ugly(t *T) {
	cwd := Getwd()
	RequireTrue(t, cwd.OK)

	r := Chdir(".")

	AssertTrue(t, r.OK)
	after := Getwd()
	AssertTrue(t, after.OK)
	AssertEqual(t, cwd.Value.(string), after.Value.(string))
}

func TestOs_Create_Bad(t *T) {
	r := Create(Path(t.TempDir(), "missing", "agent.log"))

	AssertFalse(t, r.OK)
}

func TestOs_Create_Ugly(t *T) {
	path := Path(t.TempDir(), "agent.log")
	r := Create(path)
	RequireTrue(t, r.OK)
	CloseStream(r.Value)

	second := Create(path)
	RequireTrue(t, second.OK)
	CloseStream(second.Value)
	read := ReadFile(path)
	AssertTrue(t, read.OK)
	AssertEqual(t, []byte{}, read.Value.([]byte))
}

func TestOs_DirFS_Good(t *T) {
	dir := t.TempDir()
	path := Path(dir, "agent.txt")
	AssertTrue(t, WriteFile(path, []byte("ready"), 0o644).OK)

	r := ReadFSFile(DirFS(dir), "agent.txt")

	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("ready"), r.Value.([]byte))
}

func TestOs_DirFS_Bad(t *T) {
	r := ReadFSFile(DirFS(t.TempDir()), "missing.txt")

	AssertFalse(t, r.OK)
}

func TestOs_DirFS_Ugly(t *T) {
	r := ReadDir(DirFS(t.TempDir()), ".")

	AssertTrue(t, r.OK)
	AssertLen(t, r.Value.([]FsDirEntry), 0)
}

func TestOs_Environ_Good(t *T) {
	t.Setenv("CORE_AX7_AGENT", "dispatch")

	env := Environ()

	found := false
	for _, entry := range env {
		if entry == "CORE_AX7_AGENT=dispatch" {
			found = true
		}
	}
	AssertTrue(t, found)
}

func TestOs_Environ_Bad(t *T) {
	env := Environ()

	AssertNotNil(t, env)
}

func TestOs_Environ_Ugly(t *T) {
	t.Setenv("CORE_AX7_EMPTY", "")

	env := Environ()

	found := false
	for _, entry := range env {
		if entry == "CORE_AX7_EMPTY=" {
			found = true
		}
	}
	AssertTrue(t, found)
}

func TestOs_Getenv_Good(t *T) {
	t.Setenv("CORE_AX7_TOKEN", "session-token")

	AssertEqual(t, "session-token", Getenv("CORE_AX7_TOKEN"))
}

func TestOs_Getenv_Bad(t *T) {
	Unsetenv("CORE_AX7_MISSING")

	AssertEqual(t, "", Getenv("CORE_AX7_MISSING"))
}

func TestOs_Getenv_Ugly(t *T) {
	t.Setenv("CORE_AX7_EMPTY", "")

	AssertEqual(t, "", Getenv("CORE_AX7_EMPTY"))
}

func TestOs_Getpid_Good(t *T) {
	AssertGreater(t, Getpid(), 0)
}

func TestOs_Getpid_Bad(t *T) {
	AssertNotEqual(t, 0, Getpid())
}

func TestOs_Getpid_Ugly(t *T) {
	AssertEqual(t, Getpid(), Getpid())
}

func TestOs_Getppid_Good(t *T) {
	AssertGreater(t, Getppid(), 0)
}

func TestOs_Getppid_Bad(t *T) {
	AssertNotEqual(t, 0, Getppid())
}

func TestOs_Getppid_Ugly(t *T) {
	AssertEqual(t, Getppid(), Getppid())
}

func TestOs_Getwd_Good(t *T) {
	r := Getwd()

	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_Getwd_Bad(t *T) {
	r := Getwd()

	AssertTrue(t, r.OK)
}

func TestOs_Getwd_Ugly(t *T) {
	cwd := Getwd()
	RequireTrue(t, cwd.OK)
	defer func() { AssertTrue(t, Chdir(cwd.Value.(string)).OK) }()
	dir := t.TempDir()
	realDir := PathEvalSymlinks(dir)
	RequireTrue(t, realDir.OK)

	AssertTrue(t, Chdir(dir).OK)
	r := Getwd()

	AssertTrue(t, r.OK)
	AssertEqual(t, realDir.Value.(string), r.Value.(string))
}

func TestOs_Hostname_Good(t *T) {
	r := Hostname()

	AssertTrue(t, r.OK)
}

func TestOs_Hostname_Bad(t *T) {
	r := Hostname()

	AssertTrue(t, r.OK)
}

func TestOs_Hostname_Ugly(t *T) {
	first := Hostname()
	second := Hostname()

	AssertTrue(t, first.OK)
	AssertTrue(t, second.OK)
	AssertEqual(t, first.Value.(string), second.Value.(string))
}

func TestOs_IsExist_Good(t *T) {
	r := Mkdir(t.TempDir(), 0o755)

	AssertFalse(t, r.OK)
	AssertTrue(t, IsExist(r.Value.(error)))
}

func TestOs_IsExist_Bad(t *T) {
	AssertFalse(t, IsExist(AnError))
}

func TestOs_IsExist_Ugly(t *T) {
	AssertFalse(t, IsExist(nil))
}

func TestOs_IsNotExist_Good(t *T) {
	r := ReadFile(Path(t.TempDir(), "missing.txt"))

	AssertFalse(t, r.OK)
	AssertTrue(t, IsNotExist(r.Value.(error)))
}

func TestOs_IsNotExist_Bad(t *T) {
	AssertFalse(t, IsNotExist(AnError))
}

func TestOs_IsNotExist_Ugly(t *T) {
	AssertFalse(t, IsNotExist(nil))
}

func TestOs_IsPermission_Good(t *T) {
	AssertTrue(t, IsPermission(os.ErrPermission))
}

func TestOs_IsPermission_Bad(t *T) {
	AssertFalse(t, IsPermission(AnError))
}

func TestOs_IsPermission_Ugly(t *T) {
	AssertFalse(t, IsPermission(nil))
}

func TestOs_LookupEnv_Good(t *T) {
	t.Setenv("CORE_AX7_LOOKUP", "present")

	value, ok := LookupEnv("CORE_AX7_LOOKUP")

	AssertTrue(t, ok)
	AssertEqual(t, "present", value)
}

func TestOs_LookupEnv_Bad(t *T) {
	Unsetenv("CORE_AX7_LOOKUP_MISSING")

	value, ok := LookupEnv("CORE_AX7_LOOKUP_MISSING")

	AssertFalse(t, ok)
	AssertEqual(t, "", value)
}

func TestOs_LookupEnv_Ugly(t *T) {
	t.Setenv("CORE_AX7_LOOKUP_EMPTY", "")

	value, ok := LookupEnv("CORE_AX7_LOOKUP_EMPTY")

	AssertTrue(t, ok)
	AssertEqual(t, "", value)
}

func TestOs_Lstat_Good(t *T) {
	path := Path(t.TempDir(), "agent.txt")
	AssertTrue(t, WriteFile(path, []byte("ready"), 0o644).OK)

	r := Lstat(path)

	AssertTrue(t, r.OK)
	info := r.Value.(interface{ Name() string })
	AssertEqual(t, "agent.txt", info.Name())
}

func TestOs_Lstat_Bad(t *T) {
	r := Lstat(Path(t.TempDir(), "missing.txt"))

	AssertFalse(t, r.OK)
}

func TestOs_Lstat_Ugly(t *T) {
	dir := t.TempDir()
	target := Path(dir, "agent.txt")
	link := Path(dir, "current")
	AssertTrue(t, WriteFile(target, []byte("ready"), 0o644).OK)
	RequireNoError(t, os.Symlink(target, link))

	r := Lstat(link)

	AssertTrue(t, r.OK)
	info := r.Value.(interface{ Mode() FileMode })
	AssertTrue(t, info.Mode()&ModeSymlink != 0)
}

func TestOs_Mkdir_Good(t *T) {
	path := Path(t.TempDir(), "agent")

	r := Mkdir(path, 0o755)

	AssertTrue(t, r.OK)
	AssertTrue(t, Stat(path).OK)
}

func TestOs_Mkdir_Bad(t *T) {
	dir := t.TempDir()

	r := Mkdir(dir, 0o755)

	AssertFalse(t, r.OK)
}

func TestOs_Mkdir_Ugly(t *T) {
	r := Mkdir("", 0o755)

	AssertFalse(t, r.OK)
}

func TestOs_MkdirAll_Good(t *T) {
	path := Path(t.TempDir(), "agent", "dispatch", "logs")

	r := MkdirAll(path, 0o755)

	AssertTrue(t, r.OK)
	AssertTrue(t, Stat(path).OK)
}

func TestOs_MkdirAll_Bad(t *T) {
	dir := t.TempDir()
	blocker := Path(dir, "agent")
	AssertTrue(t, WriteFile(blocker, []byte("file"), 0o644).OK)

	r := MkdirAll(Path(blocker, "dispatch"), 0o755)

	AssertFalse(t, r.OK)
}

func TestOs_MkdirAll_Ugly(t *T) {
	r := MkdirAll("", 0o755)

	AssertFalse(t, r.OK)
}

func TestOs_MkdirTemp_Good(t *T) {
	r := MkdirTemp("", "agent-*")
	RequireTrue(t, r.OK)
	defer RemoveAll(r.Value.(string))

	AssertTrue(t, Stat(r.Value.(string)).OK)
}

func TestOs_MkdirTemp_Bad(t *T) {
	r := MkdirTemp(Path(t.TempDir(), "missing"), "agent-*")

	AssertFalse(t, r.OK)
}

func TestOs_MkdirTemp_Ugly(t *T) {
	r := MkdirTemp("", "")
	RequireTrue(t, r.OK)
	defer RemoveAll(r.Value.(string))

	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_Open_Ugly(t *T) {
	path := Path(t.TempDir(), "empty.txt")
	AssertTrue(t, WriteFile(path, nil, 0o644).OK)

	r := Open(path)

	AssertTrue(t, r.OK)
	CloseStream(r.Value)
}

func TestOs_OpenFile_Good(t *T) {
	path := Path(t.TempDir(), "agent.log")

	r := OpenFile(path, O_CREATE|O_WRONLY, 0o644)
	RequireTrue(t, r.OK)
	defer CloseStream(r.Value)

	AssertTrue(t, WriteAll(r.Value, "ready").OK)
}

func TestOs_OpenFile_Bad(t *T) {
	r := OpenFile(Path(t.TempDir(), "missing.log"), O_RDONLY, 0o644)

	AssertFalse(t, r.OK)
}

func TestOs_OpenFile_Ugly(t *T) {
	path := Path(t.TempDir(), "agent.log")
	AssertTrue(t, WriteFile(path, []byte("ready"), 0o644).OK)

	r := OpenFile(path, O_CREATE|O_EXCL|O_WRONLY, 0o644)

	AssertFalse(t, r.OK)
}

func TestOs_ReadFile_Bad(t *T) {
	r := ReadFile(Path(t.TempDir(), "missing.txt"))

	AssertFalse(t, r.OK)
}

func TestOs_ReadFile_Ugly(t *T) {
	path := Path(t.TempDir(), "empty.txt")
	AssertTrue(t, WriteFile(path, nil, 0o644).OK)

	r := ReadFile(path)

	AssertTrue(t, r.OK)
	AssertEqual(t, []byte{}, r.Value.([]byte))
}

func TestOs_Remove_Ugly(t *T) {
	r := Remove(Path(t.TempDir(), "missing.txt"))

	AssertFalse(t, r.OK)
}

func TestOs_RemoveAll_Good(t *T) {
	dir := t.TempDir()
	path := Path(dir, "agent", "dispatch.log")
	AssertTrue(t, MkdirAll(PathDir(path), 0o755).OK)
	AssertTrue(t, WriteFile(path, []byte("ready"), 0o644).OK)

	r := RemoveAll(Path(dir, "agent"))

	AssertTrue(t, r.OK)
	AssertFalse(t, Stat(Path(dir, "agent")).OK)
}

func TestOs_RemoveAll_Bad(t *T) {
	r := RemoveAll(Path(t.TempDir(), "missing"))

	AssertTrue(t, r.OK)
}

func TestOs_RemoveAll_Ugly(t *T) {
	r := RemoveAll("")

	AssertTrue(t, r.OK)
}

func TestOs_Rename_Bad(t *T) {
	dir := t.TempDir()

	r := Rename(Path(dir, "missing.txt"), Path(dir, "agent.txt"))

	AssertFalse(t, r.OK)
}

func TestOs_Rename_Ugly(t *T) {
	dir := t.TempDir()
	oldPath := Path(dir, "agent.tmp")
	newPath := Path(dir, "agent.json")
	AssertTrue(t, WriteFile(oldPath, []byte("new"), 0o644).OK)
	AssertTrue(t, WriteFile(newPath, []byte("old"), 0o644).OK)

	r := Rename(oldPath, newPath)

	AssertTrue(t, r.OK)
	read := ReadFile(newPath)
	AssertTrue(t, read.OK)
	AssertEqual(t, []byte("new"), read.Value.([]byte))
}

func TestOs_Stat_Bad(t *T) {
	r := Stat(Path(t.TempDir(), "missing.txt"))

	AssertFalse(t, r.OK)
}

func TestOs_Stat_Ugly(t *T) {
	dir := t.TempDir()

	r := Stat(dir)

	AssertTrue(t, r.OK)
	info := r.Value.(interface{ IsDir() bool })
	AssertTrue(t, info.IsDir())
}

func TestOs_Stdin_Good(t *T) {
	AssertNotNil(t, Stdin())
}

func TestOs_Stdin_Bad(t *T) {
	var reader Reader = Stdin()

	AssertNotNil(t, reader)
}

func TestOs_Stdin_Ugly(t *T) {
	first := Stdin()
	second := Stdin()

	AssertEqual(t, first, second)
}

func TestOs_Stdout_Good(t *T) {
	AssertNotNil(t, Stdout())
}

func TestOs_Stdout_Bad(t *T) {
	r := WriteString(Stdout(), "")

	AssertTrue(t, r.OK)
	AssertEqual(t, 0, r.Value.(int))
}

func TestOs_Stdout_Ugly(t *T) {
	first := Stdout()
	second := Stdout()

	AssertEqual(t, first, second)
}

func TestOs_Stderr_Good(t *T) {
	AssertNotNil(t, Stderr())
}

func TestOs_Stderr_Bad(t *T) {
	r := WriteString(Stderr(), "")

	AssertTrue(t, r.OK)
	AssertEqual(t, 0, r.Value.(int))
}

func TestOs_Stderr_Ugly(t *T) {
	first := Stderr()
	second := Stderr()

	AssertEqual(t, first, second)
}

func TestOs_TempDir_Good(t *T) {
	dir := TempDir()

	AssertNotEmpty(t, dir)
	AssertTrue(t, Stat(dir).OK)
}

func TestOs_TempDir_Bad(t *T) {
	AssertNotEmpty(t, TempDir())
}

func TestOs_TempDir_Ugly(t *T) {
	t.Setenv("TMPDIR", t.TempDir())

	AssertNotEmpty(t, TempDir())
}

func TestOs_Unsetenv_Good(t *T) {
	t.Setenv("CORE_AX7_UNSET", "value")

	err := Unsetenv("CORE_AX7_UNSET")
	_, ok := LookupEnv("CORE_AX7_UNSET")

	AssertNoError(t, err)
	AssertFalse(t, ok)
}

func TestOs_Unsetenv_Bad(t *T) {
	err := Unsetenv("")

	AssertNoError(t, err)
}

func TestOs_Unsetenv_Ugly(t *T) {
	err := Unsetenv("CORE_AX7_ALREADY_MISSING")

	AssertNoError(t, err)
}

func TestOs_UserCacheDir_Good(t *T) {
	r := UserCacheDir()

	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_UserCacheDir_Bad(t *T) {
	r := UserCacheDir()

	AssertTrue(t, r.OK)
}

func TestOs_UserCacheDir_Ugly(t *T) {
	t.Setenv("HOME", t.TempDir())

	r := UserCacheDir()

	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_UserConfigDir_Good(t *T) {
	r := UserConfigDir()

	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_UserConfigDir_Bad(t *T) {
	r := UserConfigDir()

	AssertTrue(t, r.OK)
}

func TestOs_UserConfigDir_Ugly(t *T) {
	t.Setenv("HOME", t.TempDir())

	r := UserConfigDir()

	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_UserHomeDir_Good(t *T) {
	r := UserHomeDir()

	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.(string))
}

func TestOs_UserHomeDir_Bad(t *T) {
	r := UserHomeDir()

	AssertTrue(t, r.OK)
}

func TestOs_UserHomeDir_Ugly(t *T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	r := UserHomeDir()

	AssertTrue(t, r.OK)
	AssertEqual(t, home, r.Value.(string))
}

func TestOs_WriteFile_Good(t *T) {
	path := Path(t.TempDir(), "agent.json")

	r := WriteFile(path, []byte(`{"agent":"dispatch"}`), 0o644)

	AssertTrue(t, r.OK)
	read := ReadFile(path)
	AssertTrue(t, read.OK)
	AssertEqual(t, []byte(`{"agent":"dispatch"}`), read.Value.([]byte))
}

func TestOs_WriteFile_Bad(t *T) {
	r := WriteFile(t.TempDir(), []byte("not a file"), 0o644)

	AssertFalse(t, r.OK)
}

func TestOs_WriteFile_Ugly(t *T) {
	path := Path(t.TempDir(), "empty.txt")

	r := WriteFile(path, nil, 0o600)

	AssertTrue(t, r.OK)
	read := ReadFile(path)
	AssertTrue(t, read.OK)
	AssertEqual(t, []byte{}, read.Value.([]byte))
}
