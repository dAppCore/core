// SPDX-License-Identifier: EUPL-1.2

// Data is the embedded/stored content system for core packages.
// Packages mount their embedded content here and other packages
// read from it by path.
//
// Mount a package's assets:
//
//	c.Data().New(core.Options{
//	    {Key: "name", Value: "brain"},
//	    {Key: "source", Value: brainFS},
//	    {Key: "path", Value: "prompts"},
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
//	    {Key: "name", Value: "brain"},
//	    {Key: "source", Value: brainFS},
//	    {Key: "path", Value: "prompts"},
//	})
func (d *Data) New(opts Options) Result {
	name := opts.String("name")
	if name == "" {
		return Result{}
	}

	r := opts.Get("source")
	if !r.OK {
		return r
	}

	fsys, ok := r.Value.(fs.FS)
	if !ok {
		return Result{E("data.New", "source is not fs.FS", nil), false}
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

	mr := Mount(fsys, path)
	if !mr.OK {
		return mr
	}

	emb := mr.Value.(*Embed)
	d.mounts[name] = emb
	return Result{emb, true}
}

// Get returns the Embed for a named mount point.
//
//	r := c.Data().Get("brain")
//	if r.OK { emb := r.Value.(*Embed) }
func (d *Data) Get(name string) Result {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.mounts == nil {
		return Result{}
	}
	emb, ok := d.mounts[name]
	if !ok {
		return Result{}
	}
	return Result{emb, true}
}

// resolve splits a path like "brain/coding.md" into mount name + relative path.
func (d *Data) resolve(path string) (*Embed, string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	parts := SplitN(path, "/", 2)
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
//	r := c.Data().ReadFile("brain/prompts/coding.md")
//	if r.OK { data := r.Value.([]byte) }
func (d *Data) ReadFile(path string) Result {
	emb, rel := d.resolve(path)
	if emb == nil {
		return Result{}
	}
	return emb.ReadFile(rel)
}

// ReadString reads a file as a string.
//
//	r := c.Data().ReadString("agent/flow/deploy/to/homelab.yaml")
//	if r.OK { content := r.Value.(string) }
func (d *Data) ReadString(path string) Result {
	r := d.ReadFile(path)
	if !r.OK {
		return r
	}
	return Result{string(r.Value.([]byte)), true}
}

// List returns directory entries at a path.
//
//	r := c.Data().List("agent/persona/code")
//	if r.OK { entries := r.Value.([]fs.DirEntry) }
func (d *Data) List(path string) Result {
	emb, rel := d.resolve(path)
	if emb == nil {
		return Result{}
	}
	r := emb.ReadDir(rel)
	if !r.OK {
		return r
	}
	return Result{r.Value, true}
}

// ListNames returns filenames (without extensions) at a path.
//
//	r := c.Data().ListNames("agent/flow")
//	if r.OK { names := r.Value.([]string) }
func (d *Data) ListNames(path string) Result {
	r := d.List(path)
	if !r.OK {
		return r
	}
	entries := r.Value.([]fs.DirEntry)
	var names []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() {
			name = TrimSuffix(name, filepath.Ext(name))
		}
		names = append(names, name)
	}
	return Result{names, true}
}

// Extract copies a template directory to targetDir.
//
//	r := c.Data().Extract("agent/workspace/default", "/tmp/ws", templateData)
func (d *Data) Extract(path, targetDir string, templateData any) Result {
	emb, rel := d.resolve(path)
	if emb == nil {
		return Result{}
	}
	r := emb.Sub(rel)
	if !r.OK {
		return r
	}
	return Extract(r.Value.(*Embed).FS(), targetDir, templateData)
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
