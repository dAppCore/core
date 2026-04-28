package core_test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"

	. "dappco.re/go"
)

// --- Mount ---

func mustMountTestFS(t *T, basedir string) *Embed {
	t.Helper()

	r := Mount(testFS, basedir)
	AssertTrue(t, r.OK)
	return r.Value.(*Embed)
}

func TestEmbed_Mount_Good(t *T) {
	r := Mount(testFS, "testdata")
	AssertTrue(t, r.OK)
}

func TestEmbed_Mount_Bad(t *T) {
	r := Mount(testFS, "nonexistent")
	AssertFalse(t, r.OK)
}

// --- Embed methods ---

func TestEmbed_ReadFile_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadFile("test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestEmbed_ReadString_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadString("test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", r.Value.(string))
}

func TestEmbed_Open_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.Open("test.txt")
	AssertTrue(t, r.OK)
}

func TestEmbed_ReadDir_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadDir(".")
	AssertTrue(t, r.OK)
	AssertNotEmpty(t, r.Value)
}

func TestEmbed_Sub_Good(t *T) {
	emb := mustMountTestFS(t, ".")
	r := emb.Sub("testdata")
	AssertTrue(t, r.OK)
	sub := r.Value.(*Embed)
	r2 := sub.ReadFile("test.txt")
	AssertTrue(t, r2.OK)
}

func TestEmbed_BaseDir_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	AssertEqual(t, "testdata", emb.BaseDirectory())
}

func TestEmbed_FS_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	AssertNotNil(t, emb.FS())
}

func TestEmbed_EmbedFS_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	efs := emb.EmbedFS()
	_, err := efs.ReadFile("testdata/test.txt")
	AssertNoError(t, err)
}

// --- Extract ---

func TestEmbed_Extract_Good(t *T) {
	dir := t.TempDir()
	r := Extract(testFS, dir, nil)
	AssertTrue(t, r.OK)

	cr := (&Fs{}).New("/").Read(Path(dir, "testdata/test.txt"))
	AssertTrue(t, cr.OK)
	AssertEqual(t, "hello from testdata\n", cr.Value)
}

// --- Asset Pack ---

func TestEmbed_AddGetAsset_Good(t *T) {
	AddAsset("test-group", "greeting", mustCompress("hello world"))
	r := GetAsset("test-group", "greeting")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello world", r.Value.(string))
}

func TestEmbed_GetAsset_Bad(t *T) {
	r := GetAsset("missing-group", "missing")
	AssertFalse(t, r.OK)
}

func TestEmbed_GetAssetBytes_Good(t *T) {
	AddAsset("bytes-group", "file", mustCompress("binary content"))
	r := GetAssetBytes("bytes-group", "file")
	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("binary content"), r.Value.([]byte))
}

func TestEmbed_MountEmbed_Good(t *T) {
	r := MountEmbed(testFS, "testdata")
	AssertTrue(t, r.OK)
}

// --- ScanAssets ---

func TestEmbed_ScanAssets_Good(t *T) {
	r := ScanAssets([]string{"testdata/scantest/sample.go"})
	AssertTrue(t, r.OK)
	pkgs := r.Value.([]ScannedPackage)
	AssertLen(t, pkgs, 1)
	AssertEqual(t, "scantest", pkgs[0].PackageName)
}

func TestEmbed_ScanAssets_Bad(t *T) {
	r := ScanAssets([]string{"nonexistent.go"})
	AssertFalse(t, r.OK)
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
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadFile("../../etc/passwd")
	AssertFalse(t, r.OK)
}

func TestEmbed_Sub_BaseDir_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.Sub("scantest")
	AssertTrue(t, r.OK)
	sub := r.Value.(*Embed)
	AssertEqual(t, ".", sub.BaseDirectory())
}

func TestEmbed_Open_Bad(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.Open("nonexistent.txt")
	AssertFalse(t, r.OK)
}

func TestEmbed_ReadDir_Bad(t *T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadDir("nonexistent")
	AssertFalse(t, r.OK)
}

func TestEmbed_EmbedFS_Original_Good(t *T) {
	emb := mustMountTestFS(t, "testdata")
	efs := emb.EmbedFS()
	_, err := efs.ReadFile("testdata/test.txt")
	AssertNoError(t, err)
}

func TestEmbed_Extract_NilData_Good(t *T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	(&Fs{}).New("/").Write(Path(srcDir, "file.txt"), "no template")

	r := Extract(DirFS(srcDir), dir, nil)
	AssertTrue(t, r.OK)
}

func mustCompress(input string) string {
	var buf bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &buf)
	gz, _ := gzip.NewWriterLevel(b64, gzip.BestCompression)
	gz.Write([]byte(input))
	gz.Close()
	b64.Close()
	return buf.String()
}
