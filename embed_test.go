package core_test

import . "dappco.re/go"

// --- Mount ---

func mustMountTestFS(t *T, basedir string) *Embed {
	t.Helper()

	r := Mount(testFS, basedir)
	AssertTrue(t, r.OK)
	return r.Value.(*Embed)
}

func TestEmbed_Mount_Good(t *T) {
	r := Mount(testFS, "tests/data")
	AssertTrue(t, r.OK)
}

func TestEmbed_Mount_Bad(t *T) {
	r := Mount(testFS, "nonexistent")
	AssertFalse(t, r.OK)
}

// --- Embed methods ---

func TestEmbed_ReadFile_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadFile("test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestEmbed_ReadString_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadString("test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", r.Value.(string))
}

func TestEmbed_Open_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Open("test.txt")
	AssertTrue(t, r.OK)
}

func TestEmbed_ReadDir_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadDir(".")
	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value)
}

func TestEmbed_Sub_Good(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.Sub("tests/data")
	AssertTrue(t, r.OK)
	sub := r.Value.(*Embed)
	r2 := sub.ReadFile("test.txt")
	AssertTrue(t, r2.OK)
}

func TestEmbed_BaseDir_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	AssertEqual(t, "tests/data", emb.BaseDirectory())
}

func TestEmbed_FS_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	AssertNotNil(t, emb.FS())
}

func TestEmbed_EmbedFS_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	efs := emb.EmbedFS()
	_, err := efs.ReadFile("tests/data/test.txt")
	AssertNoError(t, err)
}

// --- Extract ---

func TestEmbed_Extract_Good(t *T) {
	dir := t.TempDir()
	r := Extract(testFS, dir, nil)
	AssertTrue(t, r.OK)

	cr := (&Fs{}).New("/").Read(Path(dir, "tests/data/test.txt"))
	AssertTrue(t, cr.OK)
	AssertEqual(t, "hello from testdata\n", cr.Value)
}

// --- Asset Pack ---

func TestEmbed_AddGetAsset_Good(t *T) {
	AddAsset("test-group", "greeting", MustCompressTestAsset(t, "hello world"))
	r := GetAsset("test-group", "greeting")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello world", r.Value.(string))
}

func TestEmbed_GetAsset_Bad(t *T) {
	AddAsset("missing-name-group", "present", MustCompressTestAsset(t, "ready"))
	missingName := GetAsset("missing-name-group", "missing")
	AssertFalse(t, missingName.OK)

	r := GetAsset("missing-group", "missing")
	AssertFalse(t, r.OK)
}

func TestEmbed_GetAssetBytes_Good(t *T) {
	AddAsset("bytes-group", "file", MustCompressTestAsset(t, "binary content"))
	r := GetAssetBytes("bytes-group", "file")
	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("binary content"), r.Value.([]byte))
}

func TestEmbed_MountEmbed_Good(t *T) {
	r := MountEmbed(testFS, "tests/data")
	AssertTrue(t, r.OK)
}

// --- ScanAssets ---

func TestEmbed_ScanAssets_Good(t *T) {
	r := ScanAssets([]string{"tests/data/_scantest/sample.go"})
	AssertTrue(t, r.OK)
	pkgs := r.Value.([]ScannedPackage)
	AssertLen(t, pkgs, 1)
	AssertEqual(t, "scantest", pkgs[0].PackageName)
}

func TestEmbed_ScanAssets_Bad(t *T) {
	r := ScanAssets([]string{"nonexistent.go"})
	AssertFalse(t, r.OK)
}

func TestEmbed_ScanAssetsGroup_Good(t *T) {
	dir := t.TempDir()
	source := `package agent

import core "dappco.re/go"

func assets() {
	_ = core.Group("assets")
}
`
	goFile := Path(dir, "agent.go")
	(&Fs{}).New("/").Write(goFile, source)

	r := ScanAssets([]string{goFile})

	AssertTrue(t, r.OK)
	pkgs := r.Value.([]ScannedPackage)
	AssertLen(t, pkgs, 1)
	AssertLen(t, pkgs[0].Groups, 1)
	AssertContains(t, pkgs[0].Groups[0], "assets")
}

func TestEmbed_GeneratePack_Empty_Good(t *T) {
	pkg := ScannedPackage{PackageName: "empty"}
	r := GeneratePack(pkg)
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "package empty")
}

func TestEmbed_GeneratePack_WithFiles_Good(t *T) {
	dir := t.TempDir()
	assetDir := Path(dir, "mygroup")
	(&Fs{}).New("/").EnsureDir(assetDir)
	(&Fs{}).New("/").Write(Path(assetDir, "hello.txt"), "hello world")

	source := "package test\nimport \"dappco.re/go\"\nfunc example() {\n\t_, _ = core.GetAsset(\"mygroup\", \"hello.txt\")\n}\n"
	goFile := Path(dir, "test.go")
	(&Fs{}).New("/").Write(goFile, source)

	sr := ScanAssets([]string{goFile})
	AssertTrue(t, sr.OK)
	pkgs := sr.Value.([]ScannedPackage)

	r := GeneratePack(pkgs[0])
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "core.AddAsset")
}

// --- Extract (template + nested) ---

func TestEmbed_Extract_WithTemplate_Good(t *T) {
	dir := t.TempDir()

	// Create an in-memory FS with a template file and a plain file
	tmplDir := DirFS(t.TempDir())

	// Use a real temp dir with files
	srcDir := t.TempDir()
	(&Fs{}).New("/").Write(Path(srcDir, "plain.txt"), "static content")
	(&Fs{}).New("/").Write(Path(srcDir, "greeting.tmpl"), "Hello {{.Name}}!")
	(&Fs{}).New("/").EnsureDir(Path(srcDir, "sub"))
	(&Fs{}).New("/").Write(Path(srcDir, "sub/nested.txt"), "nested")

	_ = tmplDir
	fsys := DirFS(srcDir)
	data := map[string]string{"Name": "World"}

	r := Extract(fsys, dir, data)
	AssertTrue(t, r.OK)

	f := (&Fs{}).New("/")

	// Plain file copied
	cr := f.Read(Path(dir, "plain.txt"))
	AssertTrue(t, cr.OK)
	AssertEqual(t, "static content", cr.Value)

	// Template processed and .tmpl stripped
	gr := f.Read(Path(dir, "greeting"))
	AssertTrue(t, gr.OK)
	AssertEqual(t, "Hello World!", gr.Value)

	// Nested directory preserved
	nr := f.Read(Path(dir, "sub/nested.txt"))
	AssertTrue(t, nr.OK)
	AssertEqual(t, "nested", nr.Value)
}

func TestEmbed_Extract_BadTargetDir_Ugly(t *T) {
	srcDir := t.TempDir()
	(&Fs{}).New("/").Write(Path(srcDir, "f.txt"), "x")
	r := Extract(DirFS(srcDir), "/nonexistent/deeply/nested/impossible", nil)
	// Should fail gracefully, not panic
	_ = r
}

func TestEmbed_PathTraversal_Ugly(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadFile("../../etc/passwd")
	AssertFalse(t, r.OK)
}

func TestEmbed_Sub_BaseDir_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Sub("_scantest")
	AssertTrue(t, r.OK)
	sub := r.Value.(*Embed)
	AssertEqual(t, ".", sub.BaseDirectory())
}

func TestEmbed_Open_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Open("nonexistent.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_ReadDir_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadDir("nonexistent")
	AssertFalse(t, r.OK)
}

func TestEmbed_EmbedFS_Original_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	efs := emb.EmbedFS()
	_, err := efs.ReadFile("tests/data/test.txt")
	AssertNoError(t, err)
}

func TestEmbed_Extract_NilData_Good(t *T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	(&Fs{}).New("/").Write(Path(srcDir, "file.txt"), "no template")

	r := Extract(DirFS(srcDir), dir, nil)
	AssertTrue(t, r.OK)
}

// --- AX-7 canonical triplets ---

func TestEmbed_AddAsset_Good(t *T) {
	AddAsset("lane-c-agent", "persona/developer.md", MustCompressTestAsset(t, "agent ready"))
	r := GetAsset("lane-c-agent", "persona/developer.md")
	AssertTrue(t, r.OK)
	AssertEqual(t, "agent ready", r.Value.(string))
}

func TestEmbed_AddAsset_Bad(t *T) {
	AddAsset("lane-c-bad", "broken.txt", "not-base64")
	r := GetAsset("lane-c-bad", "broken.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_AddAsset_Ugly(t *T) {
	AddAsset("lane-c-overwrite", "status.txt", MustCompressTestAsset(t, "old"))
	AddAsset("lane-c-overwrite", "status.txt", MustCompressTestAsset(t, "new"))
	r := GetAsset("lane-c-overwrite", "status.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "new", r.Value.(string))
}

func TestEmbed_GetAsset_Good(t *T) {
	AddAsset("lane-c-get", "agent.txt", MustCompressTestAsset(t, "codex"))
	r := GetAsset("lane-c-get", "agent.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "codex", r.Value.(string))
}

func TestEmbed_GetAsset_Ugly(t *T) {
	AddAsset("lane-c-corrupt", "agent.txt", "%%%")
	r := GetAsset("lane-c-corrupt", "agent.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_GetAssetBytes_Bad(t *T) {
	r := GetAssetBytes("lane-c-missing", "missing.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_GetAssetBytes_Ugly(t *T) {
	AddAsset("lane-c-bytes-corrupt", "agent.bin", "not-gzip")
	r := GetAssetBytes("lane-c-bytes-corrupt", "agent.bin")
	AssertFalse(t, r.OK)
}

func TestEmbed_ScanAssets_Ugly(t *T) {
	r := ScanAssets([]string{})
	AssertTrue(t, r.OK)
	AssertEmpty(t, r.Value.([]ScannedPackage))
}

func TestEmbed_GeneratePack_Good(t *T) {
	dir := t.TempDir()
	asset := Path(dir, "agent.txt")
	(&Fs{}).New("/").Write(asset, "dispatch ready")
	pkg := ScannedPackage{
		PackageName:   "agentpack",
		BaseDirectory: dir,
		Assets: []AssetRef{{
			Name:     "agent.txt",
			Group:    "persona",
			FullPath: asset,
		}},
	}
	r := GeneratePack(pkg)
	AssertTrue(t, r.OK)
	source := r.Value.(string)
	AssertContains(t, source, "package agentpack")
	AssertContains(t, source, "core.AddAsset")
}

func TestEmbed_GeneratePack_Bad(t *T) {
	dir := t.TempDir()
	pkg := ScannedPackage{
		PackageName:   "agentpack",
		BaseDirectory: dir,
		Assets: []AssetRef{{
			Name:     "missing.txt",
			Group:    "persona",
			FullPath: Path(dir, "missing.txt"),
		}},
	}
	r := GeneratePack(pkg)
	AssertFalse(t, r.OK)
}

func TestEmbed_GeneratePack_Ugly(t *T) {
	dir := t.TempDir()
	group := Path(dir, "assets")
	f := (&Fs{}).New("/")
	f.EnsureDir(group)
	f.Write(Path(group, "agent.txt"), "dispatch ready")
	pkg := ScannedPackage{
		PackageName:   "agentpack",
		BaseDirectory: dir,
		Groups:        []string{group},
	}
	r := GeneratePack(pkg)
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), `core.AddAsset("assets", "agent.txt"`)
}

func TestEmbed_GeneratePackDeduplicates_Ugly(t *T) {
	dir := t.TempDir()
	group := Path(dir, "assets")
	f := (&Fs{}).New("/")
	f.EnsureDir(group)
	asset := Path(group, "agent.txt")
	f.Write(asset, "dispatch ready")
	pkg := ScannedPackage{
		PackageName:   "agentpack",
		BaseDirectory: dir,
		Groups:        []string{group},
		Assets: []AssetRef{{
			Name:     "agent.txt",
			Group:    "assets",
			FullPath: asset,
		}},
	}

	r := GeneratePack(pkg)

	AssertTrue(t, r.OK)
	source := r.Value.(string)
	AssertLen(t, Split(source, `core.AddAsset("assets", "agent.txt"`), 2)
}

func TestEmbed_Mount_Ugly(t *T) {
	dir := t.TempDir()
	r := Mount(DirFS(dir), ".")
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(*Embed).BaseDirectory())
}

func TestEmbed_MountEmbed_Bad(t *T) {
	r := MountEmbed(testFS, "missing")
	AssertFalse(t, r.OK)
}

func TestEmbed_MountEmbed_Ugly(t *T) {
	r := MountEmbed(testFS, ".")
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(*Embed).BaseDirectory())
}

func TestEmbed_Embed_Open_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Open("test.txt")
	AssertTrue(t, r.OK)
	read := ReadAll(r.Value)
	AssertTrue(t, read.OK)
	AssertEqual(t, "hello from testdata\n", read.Value.(string))
}

func TestEmbed_Embed_Open_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Open("missing.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_Open_Ugly(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Open("../../etc/passwd")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_OpenTraversal_Bad(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.Open("../secrets")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path traversal rejected")
}

func TestEmbed_Embed_ReadDir_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadDir(".")
	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value.([]FsDirEntry))
}

func TestEmbed_Embed_ReadDir_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadDir("missing")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_ReadDir_Ugly(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadDir("test.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_ReadDirTraversal_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadDir("../../secrets")

	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_ReadDirParentTraversal_Ugly(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.ReadDir("../secrets")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path traversal rejected")
}

func TestEmbed_Embed_ReadFile_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadFile("test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestEmbed_Embed_ReadFile_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadFile("missing.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_ReadFile_Ugly(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadFile("../../secrets/token")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_ReadFileTraversal_Bad(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.ReadFile("../secrets/token")

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path traversal rejected")
}

func TestEmbed_Embed_ReadString_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadString("test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", r.Value.(string))
}

func TestEmbed_Embed_ReadString_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.ReadString("missing.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_ReadString_Ugly(t *T) {
	dir := t.TempDir()
	(&Fs{}).New("/").Write(Path(dir, "empty.txt"), "")
	r := Mount(DirFS(dir), ".")
	AssertTrue(t, r.OK)
	read := r.Value.(*Embed).ReadString("empty.txt")
	AssertTrue(t, read.OK)
	AssertEqual(t, "", read.Value.(string))
}

func TestEmbed_Embed_Sub_Good(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.Sub("tests/data")
	AssertTrue(t, r.OK)
	AssertEqual(t, ".", r.Value.(*Embed).BaseDirectory())
}

func TestEmbed_Embed_Sub_Bad(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	r := emb.Sub("missing")
	AssertTrue(t, r.OK)
	AssertFalse(t, r.Value.(*Embed).ReadDir(".").OK)
}

func TestEmbed_Embed_Sub_Ugly(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.Sub("../../")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_FS_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	AssertNotNil(t, emb.FS())
	r := ReadDir(emb.FS(), "tests/data")
	AssertTrue(t, r.OK)
}

func TestEmbed_Embed_FS_Bad(t *T) {
	dir := t.TempDir()
	emb := Mount(DirFS(dir), ".").Value.(*Embed)
	r := ReadDir(emb.FS(), "missing")
	AssertFalse(t, r.OK)
}

func TestEmbed_Embed_FS_Ugly(t *T) {
	dir := t.TempDir()
	(&Fs{}).New("/").Write(Path(dir, "agent.txt"), "ready")
	emb := Mount(DirFS(dir), ".").Value.(*Embed)
	r := ReadFSFile(emb.FS(), "agent.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "ready", string(r.Value.([]byte)))
}

func TestEmbed_Embed_EmbedFS_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	_, err := emb.EmbedFS().ReadFile("tests/data/test.txt")
	AssertNoError(t, err)
}

func TestEmbed_Embed_EmbedFS_Bad(t *T) {
	dir := t.TempDir()
	emb := Mount(DirFS(dir), ".").Value.(*Embed)
	_, err := emb.EmbedFS().ReadFile("agent.txt")
	AssertError(t, err)
}

func TestEmbed_Embed_EmbedFS_Ugly(t *T) {
	emb := mustMountTestFS(t, ".")
	_, err := emb.EmbedFS().ReadFile("tests/data/test.txt")
	AssertNoError(t, err)
}

func TestEmbed_Embed_BaseDirectory_Good(t *T) {
	emb := mustMountTestFS(t, "tests/data")
	AssertEqual(t, "tests/data", emb.BaseDirectory())
}

func TestEmbed_Embed_BaseDirectory_Bad(t *T) {
	emb := Mount(DirFS(t.TempDir()), ".").Value.(*Embed)
	AssertEqual(t, ".", emb.BaseDirectory())
}

func TestEmbed_Embed_BaseDirectory_Ugly(t *T) {
	emb := mustMountTestFS(t, "tests/data/")
	AssertEqual(t, "tests/data/", emb.BaseDirectory())
}

func TestEmbed_Extract_Bad(t *T) {
	dir := t.TempDir()
	src := t.TempDir()
	(&Fs{}).New("/").Write(Path(src, "broken.tmpl"), "hello {{")
	r := Extract(DirFS(src), dir, map[string]string{"Agent": "codex"})
	AssertFalse(t, r.OK)
}

func TestEmbed_Extract_Ugly(t *T) {
	target := t.TempDir()
	src := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(src, "README.tmpl"), "agent {{.Agent}}")
	f.Write(Path(src, "skip.txt"), "skip")
	r := Extract(DirFS(src), target, map[string]string{"Agent": "codex"}, ExtractOptions{
		IgnoreFiles: map[string]struct{}{"skip.txt": {}},
		RenameFiles: map[string]string{"README": "README.md"},
	})
	AssertTrue(t, r.OK)
	read := f.Read(Path(target, "README.md"))
	AssertTrue(t, read.OK)
	AssertEqual(t, "agent codex", read.Value)
	AssertFalse(t, f.Exists(Path(target, "skip.txt")))
}

func TestEmbed_ExtractCustomFilter_Good(t *T) {
	target := t.TempDir()
	src := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(src, "agent.gotmpl"), "agent {{.Agent}}")

	r := Extract(DirFS(src), target, map[string]string{"Agent": "codex"}, ExtractOptions{
		TemplateFilters: []string{".gotmpl"},
		RenameFiles:     map[string]string{"agent": "agent.txt"},
	})

	AssertTrue(t, r.OK)
	read := f.Read(Path(target, "agent.txt"))
	AssertTrue(t, read.OK)
	AssertEqual(t, "agent codex", read.Value)
}

func TestEmbed_ExtractRenderedPathEscape_Bad(t *T) {
	target := t.TempDir()
	src := t.TempDir()
	f := (&Fs{}).New("/")
	f.EnsureDir(Path(src, "{{.Name}}"))

	r := Extract(DirFS(src), target, map[string]string{"Name": "../escape"})

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path escapes target")
}

func TestEmbed_ExtractRenamedPathEscape_Bad(t *T) {
	target := t.TempDir()
	src := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(src, "agent.txt"), "ready")

	r := Extract(DirFS(src), target, nil, ExtractOptions{
		RenameFiles: map[string]string{"agent.txt": "../escape.txt"},
	})

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "path escapes target")
}

func TestEmbed_ExtractTemplateExecute_Bad(t *T) {
	target := t.TempDir()
	src := t.TempDir()
	f := (&Fs{}).New("/")
	f.Write(Path(src, "agent.tmpl"), "{{call .Agent}}")

	r := Extract(DirFS(src), target, map[string]string{"Agent": "codex"})

	AssertFalse(t, r.OK)
}

func TestEmbed_ExtractCopyParentFile_Ugly(t *T) {
	target := t.TempDir()
	src := t.TempDir()
	f := (&Fs{}).New("/")
	f.EnsureDir(Path(src, "blocked"))
	f.Write(Path(src, "blocked", "agent.txt"), "ready")
	f.Write(Path(target, "blocked"), "file")

	r := Extract(DirFS(src), target, nil)

	AssertFalse(t, r.OK)
}
