package core_test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"os"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Mount ---

func mustMountTestFS(t *testing.T, basedir string) *Embed {
	t.Helper()

	r := Mount(testFS, basedir)
	assert.True(t, r.OK)
	return r.Value.(*Embed)
}

func TestMount_Good(t *testing.T) {
	r := Mount(testFS, "testdata")
	assert.True(t, r.OK)
}

func TestMount_Bad(t *testing.T) {
	r := Mount(testFS, "nonexistent")
	assert.False(t, r.OK)
}

// --- Embed methods ---

func TestEmbed_ReadFile_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadFile("test.txt")
	assert.True(t, r.OK)
	assert.Equal(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestEmbed_ReadString_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadString("test.txt")
	assert.True(t, r.OK)
	assert.Equal(t, "hello from testdata\n", r.Value.(string))
}

func TestEmbed_Open_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.Open("test.txt")
	assert.True(t, r.OK)
}

func TestEmbed_ReadDir_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadDir(".")
	assert.True(t, r.OK)
	assert.NotEmpty(t, r.Value)
}

func TestEmbed_Sub_Good(t *testing.T) {
	emb := mustMountTestFS(t, ".")
	r := emb.Sub("testdata")
	assert.True(t, r.OK)
	sub := r.Value.(*Embed)
	r2 := sub.ReadFile("test.txt")
	assert.True(t, r2.OK)
}

func TestEmbed_BaseDir_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	assert.Equal(t, "testdata", emb.BaseDirectory())
}

func TestEmbed_FS_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	assert.NotNil(t, emb.FS())
}

func TestEmbed_EmbedFS_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	efs := emb.EmbedFS()
	_, err := efs.ReadFile("testdata/test.txt")
	assert.NoError(t, err)
}

// --- Extract ---

func TestExtract_Good(t *testing.T) {
	dir := t.TempDir()
	r := Extract(testFS, dir, nil)
	assert.True(t, r.OK)

	content, err := os.ReadFile(dir + "/testdata/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(content))
}

// --- Asset Pack ---

func TestAddGetAsset_Good(t *testing.T) {
	AddAsset("test-group", "greeting", mustCompress("hello world"))
	r := GetAsset("test-group", "greeting")
	assert.True(t, r.OK)
	assert.Equal(t, "hello world", r.Value.(string))
}

func TestGetAsset_Bad(t *testing.T) {
	r := GetAsset("missing-group", "missing")
	assert.False(t, r.OK)
}

func TestGetAssetBytes_Good(t *testing.T) {
	AddAsset("bytes-group", "file", mustCompress("binary content"))
	r := GetAssetBytes("bytes-group", "file")
	assert.True(t, r.OK)
	assert.Equal(t, []byte("binary content"), r.Value.([]byte))
}

func TestMountEmbed_Good(t *testing.T) {
	r := MountEmbed(testFS, "testdata")
	assert.True(t, r.OK)
}

// --- ScanAssets ---

func TestScanAssets_Good(t *testing.T) {
	r := ScanAssets([]string{"testdata/scantest/sample.go"})
	assert.True(t, r.OK)
	pkgs := r.Value.([]ScannedPackage)
	assert.Len(t, pkgs, 1)
	assert.Equal(t, "scantest", pkgs[0].PackageName)
}

func TestScanAssets_Bad(t *testing.T) {
	r := ScanAssets([]string{"nonexistent.go"})
	assert.False(t, r.OK)
}

func TestGeneratePack_Empty_Good(t *testing.T) {
	pkg := ScannedPackage{PackageName: "empty"}
	r := GeneratePack(pkg)
	assert.True(t, r.OK)
	assert.Contains(t, r.Value.(string), "package empty")
}

func TestGeneratePack_WithFiles_Good(t *testing.T) {
	dir := t.TempDir()
	assetDir := dir + "/mygroup"
	os.MkdirAll(assetDir, 0755)
	os.WriteFile(assetDir+"/hello.txt", []byte("hello world"), 0644)

	source := "package test\nimport \"dappco.re/go/core\"\nfunc example() {\n\t_, _ = core.GetAsset(\"mygroup\", \"hello.txt\")\n}\n"
	goFile := dir + "/test.go"
	os.WriteFile(goFile, []byte(source), 0644)

	sr := ScanAssets([]string{goFile})
	assert.True(t, sr.OK)
	pkgs := sr.Value.([]ScannedPackage)

	r := GeneratePack(pkgs[0])
	assert.True(t, r.OK)
	assert.Contains(t, r.Value.(string), "core.AddAsset")
}

// --- Extract (template + nested) ---

func TestExtract_WithTemplate_Good(t *testing.T) {
	dir := t.TempDir()

	// Create an in-memory FS with a template file and a plain file
	tmplDir := os.DirFS(t.TempDir())

	// Use a real temp dir with files
	srcDir := t.TempDir()
	os.WriteFile(srcDir+"/plain.txt", []byte("static content"), 0644)
	os.WriteFile(srcDir+"/greeting.tmpl", []byte("Hello {{.Name}}!"), 0644)
	os.MkdirAll(srcDir+"/sub", 0755)
	os.WriteFile(srcDir+"/sub/nested.txt", []byte("nested"), 0644)

	_ = tmplDir
	fsys := os.DirFS(srcDir)
	data := map[string]string{"Name": "World"}

	r := Extract(fsys, dir, data)
	assert.True(t, r.OK)

	// Plain file copied
	content, err := os.ReadFile(dir + "/plain.txt")
	assert.NoError(t, err)
	assert.Equal(t, "static content", string(content))

	// Template processed and .tmpl stripped
	greeting, err := os.ReadFile(dir + "/greeting")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", string(greeting))

	// Nested directory preserved
	nested, err := os.ReadFile(dir + "/sub/nested.txt")
	assert.NoError(t, err)
	assert.Equal(t, "nested", string(nested))
}

func TestExtract_BadTargetDir_Ugly(t *testing.T) {
	srcDir := t.TempDir()
	os.WriteFile(srcDir+"/f.txt", []byte("x"), 0644)
	r := Extract(os.DirFS(srcDir), "/nonexistent/deeply/nested/impossible", nil)
	// Should fail gracefully, not panic
	_ = r
}

func TestEmbed_PathTraversal_Ugly(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadFile("../../etc/passwd")
	assert.False(t, r.OK)
}

func TestEmbed_Sub_BaseDir_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.Sub("scantest")
	assert.True(t, r.OK)
	sub := r.Value.(*Embed)
	assert.Equal(t, ".", sub.BaseDirectory())
}

func TestEmbed_Open_Bad(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.Open("nonexistent.txt")
	assert.False(t, r.OK)
}

func TestEmbed_ReadDir_Bad(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	r := emb.ReadDir("nonexistent")
	assert.False(t, r.OK)
}

func TestEmbed_EmbedFS_Original_Good(t *testing.T) {
	emb := mustMountTestFS(t, "testdata")
	efs := emb.EmbedFS()
	_, err := efs.ReadFile("testdata/test.txt")
	assert.NoError(t, err)
}

func TestExtract_NilData_Good(t *testing.T) {
	dir := t.TempDir()
	srcDir := t.TempDir()
	os.WriteFile(srcDir+"/file.txt", []byte("no template"), 0644)

	r := Extract(os.DirFS(srcDir), dir, nil)
	assert.True(t, r.OK)
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
