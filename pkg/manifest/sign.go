package manifest

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"gopkg.in/yaml.v3"
)

// signable returns the canonical bytes to sign (manifest without sign field).
func signable(m *Manifest) ([]byte, error) {
	tmp := *m
	tmp.Sign = ""
	return yaml.Marshal(&tmp)
}

// Sign computes the ed25519 signature and stores it in m.Sign (base64).
func Sign(m *Manifest, priv ed25519.PrivateKey) error {
	msg, err := signable(m)
	if err != nil {
		return fmt.Errorf("manifest.Sign: marshal: %w", err)
	}
	sig := ed25519.Sign(priv, msg)
	m.Sign = base64.StdEncoding.EncodeToString(sig)
	return nil
}

// Verify checks the ed25519 signature in m.Sign against the public key.
func Verify(m *Manifest, pub ed25519.PublicKey) (bool, error) {
	if m.Sign == "" {
		return false, fmt.Errorf("manifest.Verify: no signature present")
	}
	sig, err := base64.StdEncoding.DecodeString(m.Sign)
	if err != nil {
		return false, fmt.Errorf("manifest.Verify: decode: %w", err)
	}
	msg, err := signable(m)
	if err != nil {
		return false, fmt.Errorf("manifest.Verify: marshal: %w", err)
	}
	return ed25519.Verify(pub, msg, sig), nil
}
