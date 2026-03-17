// SPDX-License-Identifier: EUPL-1.2

// Package mnt provides mount operations for the Core framework.
//
// Mount operations attach data to/from binaries and watch live filesystems:
//
//   - FS: mount an embed.FS subdirectory for scoped access
//   - Extract: extract a template directory with variable substitution
//   - Watch: observe filesystem changes (file watcher)
//
// Zero external dependencies. All operations use stdlib only.
//
// Usage:
//
//	sub, _ := mnt.FS(myEmbed, "lib/persona")
//	content, _ := sub.ReadFile("secops/developer.md")
//
//	mnt.Extract(sub, "/tmp/workspace", map[string]string{"Name": "myproject"})
package mnt

import (
	"embed"
	"io/fs"
	"path/filepath"
)

// Sub wraps an embed.FS with a basedir for scoped access.
// All paths are relative to basedir.
type Sub struct {
	basedir string
	fs      embed.FS
}

// FS creates a scoped view of an embed.FS anchored at basedir.
// Returns error if basedir doesn't exist in the embedded filesystem.
func FS(efs embed.FS, basedir string) (*Sub, error) {
	s := &Sub{fs: efs, basedir: basedir}
	// Verify the basedir exists
	if _, err := s.ReadDir("."); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Sub) path(name string) string {
	return filepath.ToSlash(filepath.Join(s.basedir, name))
}

// Open opens the named file for reading.
func (s *Sub) Open(name string) (fs.File, error) {
	return s.fs.Open(s.path(name))
}

// ReadDir reads the named directory.
func (s *Sub) ReadDir(name string) ([]fs.DirEntry, error) {
	return s.fs.ReadDir(s.path(name))
}

// ReadFile reads the named file.
func (s *Sub) ReadFile(name string) ([]byte, error) {
	return s.fs.ReadFile(s.path(name))
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
	return FS(s.fs, s.path(subDir))
}

// Embed returns the underlying embed.FS.
func (s *Sub) Embed() embed.FS {
	return s.fs
}

// BaseDir returns the basedir this Sub is anchored at.
func (s *Sub) BaseDir() string {
	return s.basedir
}
