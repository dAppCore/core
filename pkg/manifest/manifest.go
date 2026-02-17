package manifest

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest represents a .core/view.yml application manifest.
type Manifest struct {
	Code    string            `yaml:"code"`
	Name    string            `yaml:"name"`
	Version string            `yaml:"version"`
	Sign    string            `yaml:"sign"`
	Layout  string            `yaml:"layout"`
	Slots   map[string]string `yaml:"slots"`

	Permissions Permissions `yaml:"permissions"`
	Modules     []string    `yaml:"modules"`
}

// Permissions declares the I/O capabilities a module requires.
type Permissions struct {
	Read  []string `yaml:"read"`
	Write []string `yaml:"write"`
	Net   []string `yaml:"net"`
	Run   []string `yaml:"run"`
}

// Parse decodes YAML bytes into a Manifest.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest.Parse: %w", err)
	}
	return &m, nil
}

// SlotNames returns a deduplicated list of component names from slots.
func (m *Manifest) SlotNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, name := range m.Slots {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}
