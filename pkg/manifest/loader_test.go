package manifest

import (
	"crypto/ed25519"
	"testing"

	"forge.lthn.ai/core/go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Good(t *testing.T) {
	fs := io.NewMockMedium()
	fs.Files[".core/view.yml"] = `
code: test-app
name: Test App
version: 1.0.0
layout: HLCRF
slots:
  C: main-content
`
	m, err := Load(fs, ".")
	require.NoError(t, err)
	assert.Equal(t, "test-app", m.Code)
	assert.Equal(t, "main-content", m.Slots["C"])
}

func TestLoad_Bad_NoManifest(t *testing.T) {
	fs := io.NewMockMedium()
	_, err := Load(fs, ".")
	assert.Error(t, err)
}

func TestLoadVerified_Good(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := &Manifest{
		Code: "signed-app", Name: "Signed", Version: "1.0.0",
		Layout: "HLCRF", Slots: map[string]string{"C": "main"},
	}
	_ = Sign(m, priv)

	raw, _ := MarshalYAML(m)
	fs := io.NewMockMedium()
	fs.Files[".core/view.yml"] = string(raw)

	loaded, err := LoadVerified(fs, ".", pub)
	require.NoError(t, err)
	assert.Equal(t, "signed-app", loaded.Code)
}

func TestLoadVerified_Bad_Tampered(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := &Manifest{Code: "app", Version: "1.0.0"}
	_ = Sign(m, priv)

	raw, _ := MarshalYAML(m)
	tampered := "code: evil\n" + string(raw)[6:]
	fs := io.NewMockMedium()
	fs.Files[".core/view.yml"] = tampered

	_, err := LoadVerified(fs, ".", pub)
	assert.Error(t, err)
}
