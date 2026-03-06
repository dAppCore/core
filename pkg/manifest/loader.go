package manifest

import (
	"crypto/ed25519"
	"fmt"
	"path/filepath"

	"forge.lthn.ai/core/go-io"
	"gopkg.in/yaml.v3"
)

const manifestPath = ".core/view.yml"

// MarshalYAML serializes a manifest to YAML bytes.
func MarshalYAML(m *Manifest) ([]byte, error) {
	return yaml.Marshal(m)
}

// Load reads and parses a .core/view.yml from the given root directory.
func Load(medium io.Medium, root string) (*Manifest, error) {
	path := filepath.Join(root, manifestPath)
	data, err := medium.Read(path)
	if err != nil {
		return nil, fmt.Errorf("manifest.Load: %w", err)
	}
	return Parse([]byte(data))
}

// LoadVerified reads, parses, and verifies the ed25519 signature.
func LoadVerified(medium io.Medium, root string, pub ed25519.PublicKey) (*Manifest, error) {
	m, err := Load(medium, root)
	if err != nil {
		return nil, err
	}
	ok, err := Verify(m, pub)
	if err != nil {
		return nil, fmt.Errorf("manifest.LoadVerified: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("manifest.LoadVerified: signature verification failed for %q", m.Code)
	}
	return m, nil
}
