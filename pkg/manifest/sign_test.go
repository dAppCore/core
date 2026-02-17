package manifest

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignAndVerify_Good(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	m := &Manifest{
		Code:    "test-app",
		Name:    "Test App",
		Version: "1.0.0",
		Layout:  "HLCRF",
		Slots:   map[string]string{"C": "main"},
	}

	err = Sign(m, priv)
	require.NoError(t, err)
	assert.NotEmpty(t, m.Sign)

	ok, err := Verify(m, pub)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerify_Bad_Tampered(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := &Manifest{Code: "test-app", Version: "1.0.0"}
	_ = Sign(m, priv)

	m.Code = "evil-app" // tamper

	ok, err := Verify(m, pub)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestVerify_Bad_Unsigned(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	m := &Manifest{Code: "test-app"}

	ok, err := Verify(m, pub)
	assert.Error(t, err)
	assert.False(t, ok)
}
