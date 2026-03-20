// SPDX-License-Identifier: EUPL-1.2

// Data is the embedded/stored content system for core packages.
// Packages mount their embedded content here and other packages
// read from it by path.
//
// Mount a package's assets:
//
//	c.Data().New(core.Options{
//	    {K: "name", V: "brain"},
//	    {K: "source", V: brainFS},
//	    {K: "path", V: "prompts"},
//	})
//
// Read from any mounted path:
//
//	content := c.Data().ReadString("brain/coding.md")
//	entries := c.Data().List("agent/flow")
//
// Extract a template directory:
//
//	c.Data().Extract("agent/workspace/default", "/tmp/ws", data)
package core

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// Data manages mounted embedded filesystems from core packages.
type Data struct {
	mounts map[string]*Embed
	mu     sync.RWMutex
}

// New registers an embedded filesystem under a named prefix.
//
//	c.Data().New(core.Options{
//	    {K: "name", V: "brain"},
//	    {K: "source", V: brainFS},
//	    {K: "path", V: "prompts"},
//	})
func (d *Data) New(opts Options) Result[*Embed] {
	name := opts.String("name")
	if name == "" {
		return Result[*Embed]{}
	}

	source, ok := opts.Get("source")
	if !ok {
		return Result[*Embed]{}
	}

	fsys, ok := source.(fs.FS)
	if !ok {
		return Result[*Embed]{}
	}

	path := opts.String("path")
	if path == "" {
		path = "."
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.mounts == nil {
		d.mounts = make(map[string]*Embed)
	}

	emb, err := Mount(fsys, path)
	if err != nil {
		return Result[*Embed]{}
	}

	d.mounts[name] = emb
	return Result[*Embed]{Value: emb, OK: true}
}

// Get returns the Embed for a named mount point.
//
//	brain := c.Data().Get("brain")
func (d *Data) Get(name string) *Embed {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.mounts == nil {
		return nil
	}
	return d.mounts[name]
}

// resolve splits a path like "brain/coding.md" into mount name + relative path.
func (d *Data) resolve(path string) (*Embed, string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, ""
	}
	if d.mounts == nil {
		return nil, ""
	}
	emb := d.mounts[parts[0]]
	return emb, parts[1]
}

// ReadFile reads a file by full path.
//
//	bytes := c.Data().ReadFile("brain/prompts/coding.md")
func (d *Data) ReadFile(path string) ([]byte, error) {
	emb, rel := d.resolve(path)
	if emb == nil {
		return nil, E("data.ReadFile", "mount not found: "+path, nil)
	}
	return emb.ReadFile(rel)
}

// ReadString reads a file as a string.
//
//	content := c.Data().ReadString("agent/flow/deploy/to/homelab.yaml")
func (d *Data) ReadString(path string) (string, error) {
	data, err := d.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// List returns directory entries at a path.
//
//	entries := c.Data().List("agent/persona/code")
func (d *Data) List(path string) ([]fs.DirEntry, error) {
	emb, rel := d.resolve(path)
	if emb == nil {
		return nil, E("data.List", "mount not found: "+path, nil)
	}
	return emb.ReadDir(rel)
}

// ListNames returns filenames (without extensions) at a path.
//
//	names := c.Data().ListNames("agent/flow")
func (d *Data) ListNames(path string) ([]string, error) {
	entries, err := d.List(path)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() {
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}
		names = append(names, name)
	}
	return names, nil
}

// Extract copies a template directory to targetDir.
//
//	c.Data().Extract("agent/workspace/default", "/tmp/ws", templateData)
func (d *Data) Extract(path, targetDir string, templateData any) error {
	emb, rel := d.resolve(path)
	if emb == nil {
		return E("data.Extract", "mount not found: "+path, nil)
	}
	sub, err := emb.Sub(rel)
	if err != nil {
		return err
	}
	return Extract(sub.FS(), targetDir, templateData)
}

// Mounts returns the names of all mounted content.
//
//	names := c.Data().Mounts()
func (d *Data) Mounts() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var names []string
	for k := range d.mounts {
		names = append(names, k)
	}
	return names
}
