package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Good(t *testing.T) {
	raw := `
code: photo-browser
name: Photo Browser
version: 0.1.0
sign: dGVzdHNpZw==

layout: HLCRF
slots:
  H: nav-breadcrumb
  L: folder-tree
  C: photo-grid
  R: metadata-panel
  F: status-bar

permissions:
  read: ["./photos/"]
  write: []
  net: []
  run: []

modules:
  - core/media
  - core/fs
`
	m, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "photo-browser", m.Code)
	assert.Equal(t, "Photo Browser", m.Name)
	assert.Equal(t, "0.1.0", m.Version)
	assert.Equal(t, "dGVzdHNpZw==", m.Sign)
	assert.Equal(t, "HLCRF", m.Layout)
	assert.Equal(t, "nav-breadcrumb", m.Slots["H"])
	assert.Equal(t, "photo-grid", m.Slots["C"])
	assert.Len(t, m.Permissions.Read, 1)
	assert.Equal(t, "./photos/", m.Permissions.Read[0])
	assert.Len(t, m.Modules, 2)
}

func TestParse_Bad(t *testing.T) {
	_, err := Parse([]byte("not: valid: yaml: ["))
	assert.Error(t, err)
}

func TestManifest_SlotNames_Good(t *testing.T) {
	m := Manifest{
		Slots: map[string]string{
			"H": "nav-bar",
			"C": "main-content",
		},
	}
	names := m.SlotNames()
	assert.Contains(t, names, "nav-bar")
	assert.Contains(t, names, "main-content")
	assert.Len(t, names, 2)
}
