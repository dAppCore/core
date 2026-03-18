// SPDX-License-Identifier: EUPL-1.2

// Mount operations for the Core framework.
//
// Sub provides scoped filesystem access that works with:
//   - go:embed (embed.FS)
//   - any fs.FS implementation
//   - the Core asset registry (AddAsset/GetAsset from embed.go)
//
// Usage:
//
//	sub, _ := core.Mount(myFS, "lib/persona")
//	content, _ := sub.ReadString("secops/developer.md")
//	sub.Extract("/tmp/workspace", data)
package core

import (
	"embed"
	"io/fs"
	"path/filepath"
)

// Sub wraps an fs.FS with a basedir for scoped access.
// All paths are relative to basedir.
type Sub struct {
	basedir string
	fsys    fs.FS
	embedFS *embed.FS // kept for Embed() backwards compat
}

// Mount creates a scoped view of an fs.FS anchored at basedir.
// Works with embed.FS, os.DirFS, or any fs.FS implementation.
func Mount(fsys fs.FS, basedir string) (*Sub, error) {
	s := &Sub{fsys: fsys, basedir: basedir}

	// If it's an embed.FS, keep a reference for Embed()
	if efs, ok := fsys.(embed.FS); ok {
		s.embedFS = &efs
	}

	// Verify the basedir exists
	if _, err := s.ReadDir("."); err != nil {
		return nil, err
	}
	return s, nil
}

// MountEmbed creates a scoped view of an embed.FS.
// Convenience wrapper that preserves the embed.FS type for Embed().
func MountEmbed(efs embed.FS, basedir string) (*Sub, error) {
	return Mount(efs, basedir)
}

func (s *Sub) path(name string) string {
	return filepath.ToSlash(filepath.Join(s.basedir, name))
}

// Open opens the named file for reading.
func (s *Sub) Open(name string) (fs.File, error) {
	return s.fsys.Open(s.path(name))
}

// ReadDir reads the named directory.
func (s *Sub) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(s.fsys, s.path(name))
}

// ReadFile reads the named file.
func (s *Sub) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(s.fsys, s.path(name))
}

// ReadString reads the named file as a string.
func (s *Sub) ReadString(name string) (string, error) {
	data, err := s.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Sub returns a new Sub anchored at a subdirectory within this Sub.
func (s *Sub) Sub(subDir string) (*Sub, error) {
	sub, err := fs.Sub(s.fsys, s.path(subDir))
	if err != nil {
		return nil, err
	}
	return &Sub{fsys: sub, basedir: "."}, nil
}

// FS returns the underlying fs.FS.
func (s *Sub) FS() fs.FS {
	return s.fsys
}

// Embed returns the underlying embed.FS if mounted from one.
// Returns zero embed.FS if mounted from a non-embed source.
func (s *Sub) Embed() embed.FS {
	if s.embedFS != nil {
		return *s.embedFS
	}
	return embed.FS{}
}

// BaseDir returns the basedir this Sub is anchored at.
func (s *Sub) BaseDir() string {
	return s.basedir
}
