// SPDX-License-Identifier: EUPL-1.2

// Data is the embedded/stored content system for core packages.
// Packages mount their embedded content here and other packages
// read from it by path.
//
// Mount a package's assets:
//
//	c.Data().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "brain"},
//	    core.Option{Key: "source", Value: brainFS},
//	    core.Option{Key: "path", Value: "prompts"},
//	))
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
)

// Data manages mounted embedded filesystems from core packages.
// Embeds Registry[*Embed] for thread-safe named storage.
type Data struct {
	*Registry[*Embed]
}

// New registers an embedded filesystem under a named prefix.
//
//	c.Data().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "brain"},
//	    core.Option{Key: "source", Value: brainFS},
//	    core.Option{Key: "path", Value: "prompts"},
//	))
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

	mr := Mount(fsys, path)
	if !mr.OK {
		return mr
	}

	emb := mr.Value.(*Embed)
	d.Set(name, emb)
	return Result{emb, true}
}

// resolve splits a path like "brain/coding.md" into mount name + relative path.
func (d *Data) resolve(path string) (*Embed, string) {
	parts := SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, ""
	}
	r := d.Get(parts[0])
	if !r.OK {
		return nil, ""
	}
	return r.Value.(*Embed), parts[1]
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

// Mounts returns the names of all mounted content in registration order.
//
//	names := c.Data().Mounts()
func (d *Data) Mounts() []string {
	return d.Names()
}
