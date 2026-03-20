package core_test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Embed (Mount + ReadFile + Sub) ---

func TestMount_Good(t *testing.T) {
	emb, err := Mount(testFS, "testdata")
	assert.NoError(t, err)
	assert.NotNil(t, emb)
}

func TestMount_Bad(t *testing.T) {
	_, err := Mount(testFS, "nonexistent")
	assert.Error(t, err)
}

func TestEmbed_ReadFile_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	data, err := emb.ReadFile("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(data))
}

func TestEmbed_ReadString_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	s, err := emb.ReadString("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", s)
}

func TestEmbed_Open_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	f, err := emb.Open("test.txt")
	assert.NoError(t, err)
	defer f.Close()
}

func TestEmbed_ReadDir_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	entries, err := emb.ReadDir(".")
	assert.NoError(t, err)
	assert.NotEmpty(t, entries)
}

func TestEmbed_Sub_Good(t *testing.T) {
	emb, _ := Mount(testFS, ".")
	sub, err := emb.Sub("testdata")
	assert.NoError(t, err)
	data, err := sub.ReadFile("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(data))
}

func TestEmbed_BaseDir_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	assert.Equal(t, "testdata", emb.BaseDir())
}

func TestEmbed_FS_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	assert.NotNil(t, emb.FS())
}

func TestEmbed_EmbedFS_Good(t *testing.T) {
	emb, _ := Mount(testFS, "testdata")
	efs := emb.EmbedFS()
	// Should return the original embed.FS
	_, err := efs.ReadFile("testdata/test.txt")
	assert.NoError(t, err)
}

// --- Extract (Template Directory) ---

func TestExtract_Good(t *testing.T) {
	dir := t.TempDir()
	err := Extract(testFS, dir, nil)
	assert.NoError(t, err)

	// testdata/test.txt should be extracted
	content, err := os.ReadFile(filepath.Join(dir, "testdata", "test.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(content))
}

// --- Asset Pack (Build-time) ---

func TestAddGetAsset_Good(t *testing.T) {
	AddAsset("test-group", "greeting", mustCompress("hello world"))
	result, err := GetAsset("test-group", "greeting")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestGetAsset_Bad(t *testing.T) {
	_, err := GetAsset("missing-group", "missing")
	assert.Error(t, err)

	AddAsset("exists", "item", mustCompress("data"))
	_, err = GetAsset("exists", "missing-item")
	assert.Error(t, err)
}

func TestGetAssetBytes_Good(t *testing.T) {
	AddAsset("bytes-group", "file", mustCompress("binary content"))
	data, err := GetAssetBytes("bytes-group", "file")
	assert.NoError(t, err)
	assert.Equal(t, []byte("binary content"), data)
}

// mustCompress is a test helper — compresses a string the way AddAsset expects.
func mustCompress(input string) string {
	// AddAsset stores pre-compressed data. We need to compress it the same way.
	// Use the internal format: base64(gzip(input))
	var buf bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &buf)
	gz, _ := gzip.NewWriterLevel(b64, gzip.BestCompression)
	gz.Write([]byte(input))
	gz.Close()
	b64.Close()
	return buf.String()
}

// --- ScanAssets (Build-time AST) ---

func TestScanAssets_Good(t *testing.T) {
	pkgs, err := ScanAssets([]string{"testdata/scantest/sample.go"})
	assert.NoError(t, err)
	assert.Len(t, pkgs, 1)
	assert.Equal(t, "scantest", pkgs[0].PackageName)
	assert.NotEmpty(t, pkgs[0].Assets)
	assert.Equal(t, "myfile.txt", pkgs[0].Assets[0].Name)
	assert.Equal(t, "mygroup", pkgs[0].Assets[0].Group)
}

func TestScanAssets_Bad(t *testing.T) {
	_, err := ScanAssets([]string{"nonexistent.go"})
	assert.Error(t, err)
}

// --- GeneratePack ---

func TestGeneratePack_Good(t *testing.T) {
	pkgs, _ := ScanAssets([]string{"testdata/scantest/sample.go"})
	if len(pkgs) == 0 {
		t.Skip("no packages scanned")
	}

	// GeneratePack needs the referenced files to exist
	// Since mygroup/myfile.txt doesn't exist, it will error — that's expected
	_, err := GeneratePack(pkgs[0])
	// The error is "file not found" for the asset — that's correct behavior
	assert.Error(t, err)
}

func TestGeneratePack_Empty_Good(t *testing.T) {
	pkg := ScannedPackage{PackageName: "empty"}
	source, err := GeneratePack(pkg)
	assert.NoError(t, err)
	assert.Contains(t, source, "package empty")
}

// --- GeneratePack with real files ---

func TestGeneratePack_WithFiles_Good(t *testing.T) {
	// Create a Go source that references an asset, with the asset file present
	dir := t.TempDir()

	// Create the asset file
	assetDir := dir + "/mygroup"
	os.MkdirAll(assetDir, 0755)
	os.WriteFile(assetDir+"/hello.txt", []byte("hello world"), 0644)

	// Create the Go source referencing it
	source := `package test
import "forge.lthn.ai/core/go/pkg/core"
func example() {
	_, _ = core.GetAsset("mygroup", "hello.txt")
}
`
	goFile := dir + "/test.go"
	os.WriteFile(goFile, []byte(source), 0644)

	pkgs, err := ScanAssets([]string{goFile})
	assert.NoError(t, err)
	assert.Len(t, pkgs, 1)

	// GeneratePack compresses the file and generates init() code
	code, err := GeneratePack(pkgs[0])
	assert.NoError(t, err)
	assert.Contains(t, code, "package test")
	assert.Contains(t, code, "core.AddAsset")
}
