package plugin

import (
	"context"
	"testing"

	"forge.lthn.ai/core/go-io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── NewInstaller ───────────────────────────────────────────────────

func TestNewInstaller_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	assert.NotNil(t, inst)
	assert.Equal(t, m, inst.medium)
	assert.Equal(t, reg, inst.registry)
}

// ── Install error paths ────────────────────────────────────────────

func TestInstall_Bad_InvalidSource(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Install(context.Background(), "bad-source")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source")
}

func TestInstall_Bad_AlreadyInstalled(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "my-plugin", Version: "1.0.0"})

	inst := NewInstaller(m, reg)
	err := inst.Install(context.Background(), "org/my-plugin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

// ── Remove ─────────────────────────────────────────────────────────

func TestRemove_Good(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "removable", Version: "1.0.0"})

	// Create plugin directory.
	_ = m.EnsureDir("/plugins/removable")
	_ = m.Write("/plugins/removable/plugin.json", `{"name":"removable"}`)

	inst := NewInstaller(m, reg)
	err := inst.Remove("removable")
	require.NoError(t, err)

	// Plugin removed from registry.
	_, ok := reg.Get("removable")
	assert.False(t, ok)

	// Directory cleaned up.
	assert.False(t, m.Exists("/plugins/removable"))
}

func TestRemove_Good_DirAlreadyGone(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	_ = reg.Add(&PluginConfig{Name: "ghost", Version: "1.0.0"})
	// No directory exists — should still succeed.

	inst := NewInstaller(m, reg)
	err := inst.Remove("ghost")
	require.NoError(t, err)

	_, ok := reg.Get("ghost")
	assert.False(t, ok)
}

func TestRemove_Bad_NotFound(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

// ── Update error paths ─────────────────────────────────────────────

func TestUpdate_Bad_NotFound(t *testing.T) {
	m := io.NewMockMedium()
	reg := NewRegistry(m, "/plugins")
	inst := NewInstaller(m, reg)

	err := inst.Update(context.Background(), "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

// ── ParseSource ────────────────────────────────────────────────────

func TestParseSource_Good_OrgRepo(t *testing.T) {
	org, repo, version, err := ParseSource("host-uk/core-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "host-uk", org)
	assert.Equal(t, "core-plugin", repo)
	assert.Equal(t, "", version)
}

func TestParseSource_Good_OrgRepoVersion(t *testing.T) {
	org, repo, version, err := ParseSource("host-uk/core-plugin@v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "host-uk", org)
	assert.Equal(t, "core-plugin", repo)
	assert.Equal(t, "v1.0.0", version)
}

func TestParseSource_Good_VersionWithoutPrefix(t *testing.T) {
	org, repo, version, err := ParseSource("org/repo@1.2.3")
	assert.NoError(t, err)
	assert.Equal(t, "org", org)
	assert.Equal(t, "repo", repo)
	assert.Equal(t, "1.2.3", version)
}

func TestParseSource_Bad_Empty(t *testing.T) {
	_, _, _, err := ParseSource("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is empty")
}

func TestParseSource_Bad_NoSlash(t *testing.T) {
	_, _, _, err := ParseSource("just-a-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_TooManySlashes(t *testing.T) {
	_, _, _, err := ParseSource("a/b/c")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_EmptyOrg(t *testing.T) {
	_, _, _, err := ParseSource("/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_EmptyRepo(t *testing.T) {
	_, _, _, err := ParseSource("org/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org/repo")
}

func TestParseSource_Bad_EmptyVersion(t *testing.T) {
	_, _, _, err := ParseSource("org/repo@")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version is empty")
}
