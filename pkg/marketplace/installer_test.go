package marketplace

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"forge.lthn.ai/core/go/pkg/manifest"
	"forge.lthn.ai/core/go/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepo creates a bare-bones git repo with a manifest and main.ts.
// Returns the repo path (usable as Module.Repo for local clone).
func createTestRepo(t *testing.T, code, version string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), code)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))

	manifestYAML := "code: " + code + "\nname: Test " + code + "\nversion: \"" + version + "\"\n"
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".core", "view.yml"),
		[]byte(manifestYAML), 0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "main.ts"),
		[]byte("export async function init(core: any) {}\n"), 0644,
	))

	runGit(t, dir, "init")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

// createSignedTestRepo creates a git repo with a signed manifest.
// Returns (repo path, hex-encoded public key).
func createSignedTestRepo(t *testing.T, code, version string) (string, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	dir := filepath.Join(t.TempDir(), code)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".core"), 0755))

	m := &manifest.Manifest{
		Code:    code,
		Name:    "Test " + code,
		Version: version,
	}
	require.NoError(t, manifest.Sign(m, priv))

	data, err := manifest.MarshalYAML(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".core", "view.yml"), data, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.ts"), []byte("export async function init(core: any) {}\n"), 0644))

	runGit(t, dir, "init")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "init")

	return dir, hex.EncodeToString(pub)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=test@test.com", "-c", "user.name=test"}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, string(out))
}

func TestInstall_Good(t *testing.T) {
	repo := createTestRepo(t, "hello-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code: "hello-mod",
		Repo: repo,
	})
	require.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(filepath.Join(modulesDir, "hello-mod", "main.ts"))
	assert.NoError(t, err, "main.ts should exist in installed module")

	// Verify store entry
	raw, err := st.Get("_modules", "hello-mod")
	require.NoError(t, err)
	assert.Contains(t, raw, `"code":"hello-mod"`)
	assert.Contains(t, raw, `"version":"1.0"`)
}

func TestInstall_Good_Signed(t *testing.T) {
	repo, signKey := createSignedTestRepo(t, "signed-mod", "2.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code:    "signed-mod",
		Repo:    repo,
		SignKey: signKey,
	})
	require.NoError(t, err)

	raw, err := st.Get("_modules", "signed-mod")
	require.NoError(t, err)
	assert.Contains(t, raw, `"version":"2.0"`)
}

func TestInstall_Bad_AlreadyInstalled(t *testing.T) {
	repo := createTestRepo(t, "dup-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)
	mod := Module{Code: "dup-mod", Repo: repo}

	require.NoError(t, inst.Install(context.Background(), mod))
	err = inst.Install(context.Background(), mod)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestInstall_Bad_InvalidSignature(t *testing.T) {
	// Sign with key A, verify with key B
	repo, _ := createSignedTestRepo(t, "bad-sig", "1.0")
	_, wrongKey := createSignedTestRepo(t, "dummy", "1.0") // different key

	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)
	err = inst.Install(context.Background(), Module{
		Code:    "bad-sig",
		Repo:    repo,
		SignKey: wrongKey,
	})
	assert.Error(t, err)

	// Verify directory was cleaned up
	_, statErr := os.Stat(filepath.Join(modulesDir, "bad-sig"))
	assert.True(t, os.IsNotExist(statErr), "directory should be cleaned up on failure")
}

func TestRemove_Good(t *testing.T) {
	repo := createTestRepo(t, "rm-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{Code: "rm-mod", Repo: repo}))

	err = inst.Remove("rm-mod")
	require.NoError(t, err)

	// Directory gone
	_, statErr := os.Stat(filepath.Join(modulesDir, "rm-mod"))
	assert.True(t, os.IsNotExist(statErr))

	// Store entry gone
	_, err = st.Get("_modules", "rm-mod")
	assert.Error(t, err)
}

func TestRemove_Bad_NotInstalled(t *testing.T) {
	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(t.TempDir(), st)
	err = inst.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestInstalled_Good(t *testing.T) {
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)

	repo1 := createTestRepo(t, "mod-a", "1.0")
	repo2 := createTestRepo(t, "mod-b", "2.0")

	require.NoError(t, inst.Install(context.Background(), Module{Code: "mod-a", Repo: repo1}))
	require.NoError(t, inst.Install(context.Background(), Module{Code: "mod-b", Repo: repo2}))

	installed, err := inst.Installed()
	require.NoError(t, err)
	assert.Len(t, installed, 2)

	codes := map[string]bool{}
	for _, m := range installed {
		codes[m.Code] = true
	}
	assert.True(t, codes["mod-a"])
	assert.True(t, codes["mod-b"])
}

func TestInstalled_Good_Empty(t *testing.T) {
	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(t.TempDir(), st)
	installed, err := inst.Installed()
	require.NoError(t, err)
	assert.Empty(t, installed)
}

func TestUpdate_Good(t *testing.T) {
	repo := createTestRepo(t, "upd-mod", "1.0")
	modulesDir := filepath.Join(t.TempDir(), "modules")

	st, err := store.New(":memory:")
	require.NoError(t, err)
	defer st.Close()

	inst := NewInstaller(modulesDir, st)
	require.NoError(t, inst.Install(context.Background(), Module{Code: "upd-mod", Repo: repo}))

	// Update the origin repo
	newManifest := "code: upd-mod\nname: Updated Module\nversion: \"2.0\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".core", "view.yml"), []byte(newManifest), 0644))
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "bump version")

	err = inst.Update(context.Background(), "upd-mod")
	require.NoError(t, err)

	// Verify updated metadata
	installed, err := inst.Installed()
	require.NoError(t, err)
	require.Len(t, installed, 1)
	assert.Equal(t, "2.0", installed[0].Version)
	assert.Equal(t, "Updated Module", installed[0].Name)
}
