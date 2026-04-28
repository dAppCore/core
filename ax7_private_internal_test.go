// SPDX-License-Identifier: EUPL-1.2

package core

import "os"

type ax7FailingWriter struct{}

func (ax7FailingWriter) Write(_ []byte) (int, error) {
	return 0, E("ax7.failingWriter", "write failed", nil)
}

func ax7CrashReport(message string) CrashReport {
	return CrashReport{
		Timestamp: Now(),
		Error:     message,
		Stack:     "agent stack",
		Meta:      map[string]string{"agent": "codex"},
	}
}

func TestAction_Action_safeName_Good(t *T) {
	action := &Action{Name: "agent.dispatch"}

	AssertEqual(t, "agent.dispatch", action.safeName())
}

func TestAction_Action_safeName_Bad(t *T) {
	var action *Action

	AssertEqual(t, "<nil>", action.safeName())
}

func TestAction_Action_safeName_Ugly(t *T) {
	action := &Action{}

	AssertEqual(t, "", action.safeName())
}

func TestAction_Task_safeName_Good(t *T) {
	task := &Task{Name: "deploy.to.homelab"}

	AssertEqual(t, "deploy.to.homelab", task.safeName())
}

func TestAction_Task_safeName_Bad(t *T) {
	var task *Task

	AssertEqual(t, "<nil>", task.safeName())
}

func TestAction_Task_safeName_Ugly(t *T) {
	task := &Task{}

	AssertEqual(t, "", task.safeName())
}

func TestAction_stepOptions_Good(t *T) {
	opts := stepOptions(Step{With: NewOptions(Option{Key: "site", Value: "homelab"})})

	AssertEqual(t, "homelab", opts.String("site"))
}

func TestAction_stepOptions_Bad(t *T) {
	opts := stepOptions(Step{})

	AssertEqual(t, 0, opts.Len())
}

func TestAction_stepOptions_Ugly(t *T) {
	opts := stepOptions(Step{With: NewOptions(
		Option{Key: "agent", Value: "codex"},
		Option{Key: "retry", Value: 3},
	)})

	AssertEqual(t, "codex", opts.String("agent"))
	AssertEqual(t, 3, opts.Int("retry"))
}

func TestConfig_ConfigOptions_init_Good(t *T) {
	var opts ConfigOptions

	opts.init()

	AssertNotNil(t, opts.Settings)
	AssertNotNil(t, opts.Features)
}

func TestConfig_ConfigOptions_init_Bad(t *T) {
	opts := ConfigOptions{Settings: map[string]any{"agent": "codex"}}

	opts.init()

	AssertEqual(t, "codex", opts.Settings["agent"])
	AssertNotNil(t, opts.Features)
}

func TestConfig_ConfigOptions_init_Ugly(t *T) {
	opts := ConfigOptions{}

	opts.init()
	opts.Settings["site"] = "homelab"
	opts.init()

	AssertEqual(t, "homelab", opts.Settings["site"])
}

func TestIpc_Core_broadcast_Good(t *T) {
	c := New()
	called := 0
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called++
		return Result{OK: true}
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called++
		return Result{OK: true}
	})

	r := c.broadcast(ActionTaskStarted{TaskIdentifier: "task-1117"})

	AssertTrue(t, r.OK)
	AssertEqual(t, 2, called)
}

func TestIpc_Core_broadcast_Bad(t *T) {
	r := New().broadcast(ActionTaskStarted{})

	AssertTrue(t, r.OK)
}

func TestIpc_Core_broadcast_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: NewBuffer()}))

	c := New()
	called := 0
	c.RegisterAction(func(_ *Core, _ Message) Result {
		panic("handler refused")
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called++
		return Result{OK: true}
	})

	r := c.broadcast(ActionTaskStarted{TaskIdentifier: "task-1117"})

	AssertTrue(t, r.OK)
	AssertEqual(t, 1, called)
}

func TestData_Data_resolve_Good(t *T) {
	c := New()
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, "prompts"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, "prompts", "agent.md"), []byte("ready"), 0o644).OK)
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "."},
	))
	RequireTrue(t, r.OK)

	embed, rel := c.Data().resolve("agent/prompts/agent.md")

	AssertNotNil(t, embed)
	AssertEqual(t, "prompts/agent.md", rel)
}

func TestData_Data_resolve_Bad(t *T) {
	embed, rel := New().Data().resolve("agent")

	AssertNil(t, embed)
	AssertEqual(t, "", rel)
}

func TestData_Data_resolve_Ugly(t *T) {
	embed, rel := New().Data().resolve("missing/agent.md")

	AssertNil(t, embed)
	AssertEqual(t, "", rel)
}

func TestEmbed_Embed_path_Good(t *T) {
	embed := &Embed{basedir: "assets"}

	r := embed.path("agent/readme.md")

	AssertTrue(t, r.OK)
	AssertEqual(t, "assets/agent/readme.md", r.Value)
}

func TestEmbed_Embed_path_Bad(t *T) {
	embed := &Embed{basedir: "assets"}

	r := embed.path("../../secrets/token")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path traversal rejected")
}

func TestEmbed_Embed_path_Ugly(t *T) {
	embed := &Embed{basedir: "assets"}

	r := embed.path(".")

	AssertTrue(t, r.OK)
	AssertEqual(t, "assets", r.Value)
}

func TestEmbed_compress_Good(t *T) {
	packed, err := compress("agent dispatch ready")
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "agent dispatch ready", plain)
}

func TestEmbed_compress_Bad(t *T) {
	packed, err := compress("")
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "", plain)
}

func TestEmbed_compress_Ugly(t *T) {
	input := Join("\n", "agent", "dispatch", "retry")
	packed, err := compress(input)
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, input, plain)
}

func TestEmbed_compressFile_Good(t *T) {
	path := Path(t.TempDir(), "agent.txt")
	RequireTrue(t, WriteFile(path, []byte("ready"), 0o644).OK)

	packed, err := compressFile(path)
	RequireNoError(t, err)
	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "ready", plain)
}

func TestEmbed_compressFile_Bad(t *T) {
	_, err := compressFile(Path(t.TempDir(), "missing.txt"))

	AssertError(t, err)
}

func TestEmbed_compressFile_Ugly(t *T) {
	path := Path(t.TempDir(), "empty.txt")
	RequireTrue(t, WriteFile(path, nil, 0o644).OK)

	packed, err := compressFile(path)
	RequireNoError(t, err)
	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "", plain)
}

func TestEmbed_decompress_Good(t *T) {
	packed, err := compress("homelab")
	RequireNoError(t, err)

	plain, err := decompress(packed)

	RequireNoError(t, err)
	AssertEqual(t, "homelab", plain)
}

func TestEmbed_decompress_Bad(t *T) {
	_, err := decompress("not base64")

	AssertError(t, err)
}

func TestEmbed_decompress_Ugly(t *T) {
	_, err := decompress(Base64Encode([]byte("plain text")))

	AssertError(t, err)
}

func TestEmbed_getAllFiles_Good(t *T) {
	dir := t.TempDir()
	agent := Path(dir, "agent.txt")
	task := Path(dir, "nested", "task.txt")
	RequireTrue(t, WriteFile(agent, []byte("agent"), 0o644).OK)
	RequireTrue(t, MkdirAll(Path(dir, "nested"), 0o755).OK)
	RequireTrue(t, WriteFile(task, []byte("task"), 0o644).OK)

	files, err := getAllFiles(dir)

	RequireNoError(t, err)
	AssertContains(t, files, agent)
	AssertContains(t, files, task)
}

func TestEmbed_getAllFiles_Bad(t *T) {
	_, err := getAllFiles(Path(t.TempDir(), "missing"))

	AssertError(t, err)
}

func TestEmbed_getAllFiles_Ugly(t *T) {
	files, err := getAllFiles(t.TempDir())

	RequireNoError(t, err)
	AssertEmpty(t, files)
}

func TestEmbed_isTemplate_Good(t *T) {
	AssertTrue(t, isTemplate("README.md.tmpl", []string{".tmpl"}))
}

func TestEmbed_isTemplate_Bad(t *T) {
	AssertFalse(t, isTemplate("README.md", []string{".tmpl"}))
}

func TestEmbed_isTemplate_Ugly(t *T) {
	AssertTrue(t, isTemplate("agent.go.tpl", []string{".tmpl", ".tpl"}))
}

func TestEmbed_renderPath_Good(t *T) {
	path := renderPath("workspace/{{.Name}}/README.md", map[string]string{"Name": "agent"})

	AssertEqual(t, "workspace/agent/README.md", path)
}

func TestEmbed_renderPath_Bad(t *T) {
	path := "workspace/{{.Name/README.md"

	AssertEqual(t, path, renderPath(path, map[string]string{"Name": "agent"}))
}

func TestEmbed_renderPath_Ugly(t *T) {
	path := "workspace/{{.Name}}/README.md"

	AssertEqual(t, path, renderPath(path, nil))
}

func TestEmbed_copyFile_Good(t *T) {
	src := t.TempDir()
	target := Path(t.TempDir(), "agent.txt")
	RequireTrue(t, WriteFile(Path(src, "agent.txt"), []byte("ready"), 0o644).OK)

	err := copyFile(DirFS(src), "agent.txt", target)

	RequireNoError(t, err)
	AssertEqual(t, "ready", string(ReadFile(target).Value.([]byte)))
}

func TestEmbed_copyFile_Bad(t *T) {
	err := copyFile(DirFS(t.TempDir()), "missing.txt", Path(t.TempDir(), "out.txt"))

	AssertError(t, err)
}

func TestEmbed_copyFile_Ugly(t *T) {
	src := t.TempDir()
	target := Path(t.TempDir(), "nested", "agent.txt")
	RequireTrue(t, WriteFile(Path(src, "agent.txt"), []byte("nested"), 0o644).OK)

	err := copyFile(DirFS(src), "agent.txt", target)

	RequireNoError(t, err)
	AssertEqual(t, "nested", string(ReadFile(target).Value.([]byte)))
}

func TestError_ErrorLog_logger_Good(t *T) {
	log := NewLog(LogOptions{Level: LevelDebug, Output: NewBuffer()})

	AssertSame(t, log, (&ErrorLog{log: log}).logger())
}

func TestError_ErrorLog_logger_Bad(t *T) {
	AssertSame(t, Default(), (&ErrorLog{}).logger())
}

func TestError_ErrorLog_logger_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	log := NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()})
	SetDefault(log)

	AssertSame(t, log, (&ErrorLog{}).logger())
}

func TestError_ErrorPanic_appendReport_Good(t *T) {
	handler := &ErrorPanic{filePath: Path(t.TempDir(), "crash.json")}

	handler.appendReport(ax7CrashReport("panic: agent failed"))
	r := handler.Reports(0)

	RequireTrue(t, r.OK)
	reports := r.Value.([]CrashReport)
	AssertLen(t, reports, 1)
	AssertEqual(t, "panic: agent failed", reports[0].Error)
}

func TestError_ErrorPanic_appendReport_Bad(t *T) {
	path := Path(t.TempDir(), "crash.json")
	RequireTrue(t, WriteFile(path, []byte("{"), 0o600).OK)
	handler := &ErrorPanic{filePath: path}

	handler.appendReport(ax7CrashReport("panic: recovered"))
	r := handler.Reports(0)

	RequireTrue(t, r.OK)
	reports := r.Value.([]CrashReport)
	AssertLen(t, reports, 1)
	AssertEqual(t, "panic: recovered", reports[0].Error)
}

func TestError_ErrorPanic_appendReport_Ugly(t *T) {
	handler := &ErrorPanic{filePath: Path(t.TempDir(), "nested", "crash.json")}

	handler.appendReport(ax7CrashReport("panic: nested path"))
	r := handler.Reports(0)

	RequireTrue(t, r.OK)
	reports := r.Value.([]CrashReport)
	AssertLen(t, reports, 1)
	AssertEqual(t, "panic: nested path", reports[0].Error)
}

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

func TestLog_Log_log_Good(t *T) {
	out := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: out})

	log.log(LevelInfo, "[INF]", "agent ready", "site", "homelab")

	AssertContains(t, out.String(), "[INF]")
	AssertContains(t, out.String(), `site="homelab"`)
}

func TestLog_Log_log_Bad(t *T) {
	out := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: out})

	log.log(LevelInfo, "[INF]", "dangling key", "session")

	AssertContains(t, out.String(), "session=<nil>")
}

func TestLog_Log_log_Ugly(t *T) {
	out := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: out, RedactKeys: []string{"token"}})

	log.log(LevelInfo, "[INF]", "auth", "token", "secret", "agent", "codex")

	AssertContains(t, out.String(), `token="[REDACTED]"`)
	AssertNotContains(t, out.String(), "secret")
}

func TestLog_Log_shouldLog_Good(t *T) {
	log := NewLog(LogOptions{Level: LevelInfo})

	AssertTrue(t, log.shouldLog(LevelWarn))
}

func TestLog_Log_shouldLog_Bad(t *T) {
	log := NewLog(LogOptions{Level: LevelWarn})

	AssertFalse(t, log.shouldLog(LevelDebug))
}

func TestLog_Log_shouldLog_Ugly(t *T) {
	log := NewLog(LogOptions{Level: LevelQuiet})

	AssertFalse(t, log.shouldLog(LevelError))
}

func TestLog_identity_Good(t *T) {
	AssertEqual(t, "agent", identity("agent"))
}

func TestLog_identity_Bad(t *T) {
	AssertEqual(t, "", identity(""))
}

func TestLog_identity_Ugly(t *T) {
	AssertEqual(t, "colour", identity("colour"))
}

func TestInfo_init_Good(t *T) {
	AssertNotEmpty(t, Env("OS"))
	AssertNotEmpty(t, Env("ARCH"))
	AssertNotEmpty(t, Env("DIR_HOME"))
}

func TestInfo_init_Bad(t *T) {
	AssertNotEmpty(t, Env("DIR_HOME"))
}

func TestInfo_init_Ugly(t *T) {
	AssertNotEmpty(t, Env("DIR_DATA"))
}

func TestLog_init_Good(t *T) {
	AssertNotNil(t, Default())
}

func TestLog_init_Bad(t *T) {
	RequireTrue(t, Default() != nil)

	AssertTrue(t, Default().shouldLog(LevelError))
}

func TestLog_init_Ugly(t *T) {
	RequireTrue(t, Default() != nil)

	AssertFalse(t, Default().shouldLog(LevelDebug))
}

func TestTable_Table_write_Good(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	table.write("Name\tStatus\n")
	RequireNoError(t, table.Flush())

	AssertContains(t, out.String(), "Name")
	AssertContains(t, out.String(), "Status")
}

func TestTable_Table_write_Bad(t *T) {
	table := &Table{err: AnError}

	table.write("ignored")

	AssertError(t, table.err)
}

func TestTable_Table_write_Ugly(t *T) {
	table := NewTable(ax7FailingWriter{})

	table.write("agent\n")

	AssertError(t, table.Flush())
}

func TestTest_assertCmpFloat64_Good(t *T) {
	AssertEqual(t, -1, assertCmpFloat64(1.25, 2.5))
}

func TestTest_assertCmpFloat64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpFloat64(2.5, 2.5))
}

func TestTest_assertCmpFloat64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpFloat64(-0.5, -1.5))
}

func TestTest_assertCmpInt64_Good(t *T) {
	AssertEqual(t, -1, assertCmpInt64(-1, 1))
}

func TestTest_assertCmpInt64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpInt64(42, 42))
}

func TestTest_assertCmpInt64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpInt64(1<<62, -1<<62))
}

func TestTest_assertCmpUint64_Good(t *T) {
	AssertEqual(t, -1, assertCmpUint64(1, 2))
}

func TestTest_assertCmpUint64_Bad(t *T) {
	AssertEqual(t, 0, assertCmpUint64(42, 42))
}

func TestTest_assertCmpUint64_Ugly(t *T) {
	AssertEqual(t, 1, assertCmpUint64(1<<63, 1))
}

func TestTest_assertCompare_Good(t *T) {
	cmp, ok := assertCompare("agent", "brain")

	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)
}

func TestTest_assertCompare_Bad(t *T) {
	cmp, ok := assertCompare(struct{ Name string }{"agent"}, struct{ Name string }{"agent"})

	AssertFalse(t, ok)
	AssertEqual(t, 0, cmp)
}

func TestTest_assertCompare_Ugly(t *T) {
	cmp, ok := assertCompare(-1, uint(1))

	AssertTrue(t, ok)
	AssertEqual(t, -1, cmp)
}

func TestTest_assertContains_Good(t *T) {
	AssertTrue(t, assertContains([]string{"agent", "dispatch"}, "dispatch"))
}

func TestTest_assertContains_Bad(t *T) {
	AssertFalse(t, assertContains("agent dispatch", "missing"))
}

func TestTest_assertContains_Ugly(t *T) {
	AssertTrue(t, assertContains(map[string]int{"session": 1}, "session"))
}

func TestTest_assertIsEmpty_Good(t *T) {
	AssertTrue(t, assertIsEmpty(""))
}

func TestTest_assertIsEmpty_Bad(t *T) {
	AssertFalse(t, assertIsEmpty("agent"))
}

func TestTest_assertIsEmpty_Ugly(t *T) {
	names := []string{}

	AssertTrue(t, assertIsEmpty(&names))
}

func TestTest_assertIsNil_Good(t *T) {
	AssertTrue(t, assertIsNil(nil))
}

func TestTest_assertIsNil_Bad(t *T) {
	AssertFalse(t, assertIsNil(0))
}

func TestTest_assertIsNil_Ugly(t *T) {
	var sessions map[string]string

	AssertTrue(t, assertIsNil(sessions))
}

func TestTest_assertMsg_Good(t *T) {
	AssertEqual(t, " — agent retry", assertMsg([]string{"agent", "retry"}))
}

func TestTest_assertMsg_Bad(t *T) {
	AssertEqual(t, "", assertMsg(nil))
}

func TestTest_assertMsg_Ugly(t *T) {
	AssertEqual(t, " — lethean degraded", assertMsg([]string{"lethean", "degraded"}))
}

func TestEntitlement_defaultChecker_Good(t *T) {
	entitlement := defaultChecker("process.run", 1, Background())

	AssertTrue(t, entitlement.Allowed)
	AssertTrue(t, entitlement.Unlimited)
}

func TestEntitlement_defaultChecker_Bad(t *T) {
	entitlement := defaultChecker("process.delete", -1, Background())

	AssertTrue(t, entitlement.Allowed)
	AssertTrue(t, entitlement.Unlimited)
}

func TestEntitlement_defaultChecker_Ugly(t *T) {
	entitlement := defaultChecker("", 0, nil)

	AssertTrue(t, entitlement.Allowed)
	AssertTrue(t, entitlement.Unlimited)
}

func TestApi_extractScheme_Good(t *T) {
	AssertEqual(t, "https", extractScheme("https://homelab.lthn.sh/api"))
}

func TestApi_extractScheme_Bad(t *T) {
	AssertEqual(t, "", extractScheme(""))
}

func TestApi_extractScheme_Ugly(t *T) {
	AssertEqual(t, "stdio", extractScheme("stdio"))
}

func TestHash_hashFor_Good(t *T) {
	factory := hashFor("sha256")

	AssertNotNil(t, factory)
	AssertEqual(t, 32, factory().Size())
}

func TestHash_hashFor_Bad(t *T) {
	AssertNil(t, hashFor("blake2"))
}

func TestHash_hashFor_Ugly(t *T) {
	factory := hashFor("sha512")

	AssertNotNil(t, factory)
	AssertEqual(t, 64, factory().Size())
}

func TestApp_isExecutable_Good(t *T) {
	path := Path(t.TempDir(), "agent")
	RequireTrue(t, WriteFile(path, []byte("#!/bin/sh\n"), 0o755).OK)

	AssertTrue(t, isExecutable(path))
}

func TestApp_isExecutable_Bad(t *T) {
	AssertFalse(t, isExecutable(Path(t.TempDir(), "missing")))
}

func TestApp_isExecutable_Ugly(t *T) {
	AssertFalse(t, isExecutable(t.TempDir()))
}

func TestSha3_keccakF1600_Good(t *T) {
	var state [25]uint64

	keccakF1600(&state)

	AssertEqual(t, uint64(0xf1258f7940e1dde7), state[0])
}

func TestSha3_keccakF1600_Bad(t *T) {
	AssertPanics(t, func() {
		keccakF1600(nil)
	})
}

func TestSha3_keccakF1600_Ugly(t *T) {
	left := [25]uint64{0: 1, 24: 1 << 63}
	right := left

	keccakF1600(&left)
	keccakF1600(&right)

	AssertEqual(t, left, right)
}

func TestPath_lastIndex_Good(t *T) {
	AssertEqual(t, 9, lastIndex("deploy/to/homelab", "/"))
}

func TestPath_lastIndex_Bad(t *T) {
	AssertEqual(t, -1, lastIndex("deploy/to/homelab", "."))
}

func TestPath_lastIndex_Ugly(t *T) {
	AssertEqual(t, -1, lastIndex("deploy/to/homelab", ""))
}

func TestCommand_pathName_Good(t *T) {
	AssertEqual(t, "homelab", pathName("deploy/to/homelab"))
}

func TestCommand_pathName_Bad(t *T) {
	AssertEqual(t, "", pathName(""))
}

func TestCommand_pathName_Ugly(t *T) {
	AssertEqual(t, "", pathName("deploy/"))
}

func TestCore_registryProxy_Good(t *T) {
	src := NewRegistry[string]()
	src.Set("agent", "codex")

	proxy := registryProxy(src)

	AssertEqual(t, []string{"agent"}, proxy.Names())
	AssertEqual(t, "codex", proxy.Get("agent").Value)
}

func TestCore_registryProxy_Bad(t *T) {
	src := NewRegistry[string]()
	src.Set("agent", "codex")
	src.Disable("agent")

	proxy := registryProxy(src)

	AssertEqual(t, 0, proxy.Len())
}

func TestCore_registryProxy_Ugly(t *T) {
	AssertPanics(t, func() {
		registryProxy[int](nil)
	})
}

func TestUtils_shortRand_Good(t *T) {
	token := shortRand()

	AssertLen(t, token, 6)
	AssertTrue(t, HexDecode(token).OK)
}

func TestUtils_shortRand_Bad(t *T) {
	token := shortRand()

	AssertNotEmpty(t, token)
	AssertFalse(t, Contains(token, "/"))
}

func TestUtils_shortRand_Ugly(t *T) {
	for i := 0; i < 5; i++ {
		token := shortRand()
		decoded := HexDecode(token)
		RequireTrue(t, decoded.OK)
		AssertLen(t, decoded.Value.([]byte), 3)
	}
}
