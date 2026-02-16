// Package datanode provides an in-memory io.Medium backed by Borg's DataNode.
//
// DataNode is an in-memory fs.FS that serializes to tar. Wrapping it as a
// Medium lets any code that works with io.Medium transparently operate on
// an in-memory filesystem that can be snapshotted, shipped as a crash report,
// or wrapped in a TIM container for runc execution.
package datanode

import (
	goio "io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Snider/Borg/pkg/datanode"
	coreerr "forge.lthn.ai/core/cli/pkg/framework/core"
)

// Medium is an in-memory storage backend backed by a Borg DataNode.
// All paths are relative (no leading slash). Thread-safe via RWMutex.
type Medium struct {
	dn   *datanode.DataNode
	dirs map[string]bool // explicit directory tracking
	mu   sync.RWMutex
}

// New creates a new empty DataNode Medium.
func New() *Medium {
	return &Medium{
		dn:   datanode.New(),
		dirs: make(map[string]bool),
	}
}

// FromTar creates a Medium from a tarball, restoring all files.
func FromTar(data []byte) (*Medium, error) {
	dn, err := datanode.FromTar(data)
	if err != nil {
		return nil, coreerr.E("datanode.FromTar", "failed to restore", err)
	}
	return &Medium{
		dn:   dn,
		dirs: make(map[string]bool),
	}, nil
}

// Snapshot serializes the entire filesystem to a tarball.
// Use this for crash reports, workspace packaging, or TIM creation.
func (m *Medium) Snapshot() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, err := m.dn.ToTar()
	if err != nil {
		return nil, coreerr.E("datanode.Snapshot", "tar failed", err)
	}
	return data, nil
}

// Restore replaces the filesystem contents from a tarball.
func (m *Medium) Restore(data []byte) error {
	dn, err := datanode.FromTar(data)
	if err != nil {
		return coreerr.E("datanode.Restore", "tar failed", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dn = dn
	m.dirs = make(map[string]bool)
	return nil
}

// DataNode returns the underlying Borg DataNode.
// Use this to wrap the filesystem in a TIM container.
func (m *Medium) DataNode() *datanode.DataNode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dn
}

// clean normalizes a path: strips leading slash, cleans traversal.
func clean(p string) string {
	p = strings.TrimPrefix(p, "/")
	p = path.Clean(p)
	if p == "." {
		return ""
	}
	return p
}

// --- io.Medium interface ---

func (m *Medium) Read(p string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	f, err := m.dn.Open(p)
	if err != nil {
		return "", coreerr.E("datanode.Read", "not found: "+p, os.ErrNotExist)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", coreerr.E("datanode.Read", "stat failed: "+p, err)
	}
	if info.IsDir() {
		return "", coreerr.E("datanode.Read", "is a directory: "+p, os.ErrInvalid)
	}

	data, err := goio.ReadAll(f)
	if err != nil {
		return "", coreerr.E("datanode.Read", "read failed: "+p, err)
	}
	return string(data), nil
}

func (m *Medium) Write(p, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p = clean(p)
	if p == "" {
		return coreerr.E("datanode.Write", "empty path", os.ErrInvalid)
	}
	m.dn.AddData(p, []byte(content))

	// ensure parent dirs are tracked
	m.ensureDirsLocked(path.Dir(p))
	return nil
}

func (m *Medium) EnsureDir(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p = clean(p)
	if p == "" {
		return nil
	}
	m.ensureDirsLocked(p)
	return nil
}

// ensureDirsLocked marks a directory and all ancestors as existing.
// Caller must hold m.mu.
func (m *Medium) ensureDirsLocked(p string) {
	for p != "" && p != "." {
		m.dirs[p] = true
		p = path.Dir(p)
		if p == "." {
			break
		}
	}
}

func (m *Medium) IsFile(p string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	info, err := m.dn.Stat(p)
	return err == nil && !info.IsDir()
}

func (m *Medium) FileGet(p string) (string, error) {
	return m.Read(p)
}

func (m *Medium) FileSet(p, content string) error {
	return m.Write(p, content)
}

func (m *Medium) Delete(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p = clean(p)
	if p == "" {
		return coreerr.E("datanode.Delete", "cannot delete root", os.ErrPermission)
	}

	// Check if it's a file in the DataNode
	info, err := m.dn.Stat(p)
	if err != nil {
		// Check explicit dirs
		if m.dirs[p] {
			// Check if dir is empty
			if m.hasPrefixLocked(p + "/") {
				return coreerr.E("datanode.Delete", "directory not empty: "+p, os.ErrExist)
			}
			delete(m.dirs, p)
			return nil
		}
		return coreerr.E("datanode.Delete", "not found: "+p, os.ErrNotExist)
	}

	if info.IsDir() {
		if m.hasPrefixLocked(p + "/") {
			return coreerr.E("datanode.Delete", "directory not empty: "+p, os.ErrExist)
		}
		delete(m.dirs, p)
		return nil
	}

	// Remove the file by creating a new DataNode without it
	m.removeFileLocked(p)
	return nil
}

func (m *Medium) DeleteAll(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p = clean(p)
	if p == "" {
		return coreerr.E("datanode.DeleteAll", "cannot delete root", os.ErrPermission)
	}

	prefix := p + "/"
	found := false

	// Check if p itself is a file
	info, err := m.dn.Stat(p)
	if err == nil && !info.IsDir() {
		m.removeFileLocked(p)
		found = true
	}

	// Remove all files under prefix
	entries, _ := m.collectAllLocked()
	for _, name := range entries {
		if name == p || strings.HasPrefix(name, prefix) {
			m.removeFileLocked(name)
			found = true
		}
	}

	// Remove explicit dirs under prefix
	for d := range m.dirs {
		if d == p || strings.HasPrefix(d, prefix) {
			delete(m.dirs, d)
			found = true
		}
	}

	if !found {
		return coreerr.E("datanode.DeleteAll", "not found: "+p, os.ErrNotExist)
	}
	return nil
}

func (m *Medium) Rename(oldPath, newPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldPath = clean(oldPath)
	newPath = clean(newPath)

	// Check if source is a file
	info, err := m.dn.Stat(oldPath)
	if err != nil {
		return coreerr.E("datanode.Rename", "not found: "+oldPath, os.ErrNotExist)
	}

	if !info.IsDir() {
		// Read old, write new, delete old
		f, err := m.dn.Open(oldPath)
		if err != nil {
			return coreerr.E("datanode.Rename", "open failed: "+oldPath, err)
		}
		data, err := goio.ReadAll(f)
		f.Close()
		if err != nil {
			return coreerr.E("datanode.Rename", "read failed: "+oldPath, err)
		}
		m.dn.AddData(newPath, data)
		m.ensureDirsLocked(path.Dir(newPath))
		m.removeFileLocked(oldPath)
		return nil
	}

	// Directory rename: move all files under oldPath to newPath
	oldPrefix := oldPath + "/"
	newPrefix := newPath + "/"

	entries, _ := m.collectAllLocked()
	for _, name := range entries {
		if strings.HasPrefix(name, oldPrefix) {
			newName := newPrefix + strings.TrimPrefix(name, oldPrefix)
			f, err := m.dn.Open(name)
			if err != nil {
				continue
			}
			data, _ := goio.ReadAll(f)
			f.Close()
			m.dn.AddData(newName, data)
			m.removeFileLocked(name)
		}
	}

	// Move explicit dirs
	dirsToMove := make(map[string]string)
	for d := range m.dirs {
		if d == oldPath || strings.HasPrefix(d, oldPrefix) {
			newD := newPath + strings.TrimPrefix(d, oldPath)
			dirsToMove[d] = newD
		}
	}
	for old, nw := range dirsToMove {
		delete(m.dirs, old)
		m.dirs[nw] = true
	}

	return nil
}

func (m *Medium) List(p string) ([]fs.DirEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)

	entries, err := m.dn.ReadDir(p)
	if err != nil {
		// Check explicit dirs
		if p == "" || m.dirs[p] {
			return []fs.DirEntry{}, nil
		}
		return nil, coreerr.E("datanode.List", "not found: "+p, os.ErrNotExist)
	}

	// Also include explicit subdirectories not discovered via files
	prefix := p
	if prefix != "" {
		prefix += "/"
	}
	seen := make(map[string]bool)
	for _, e := range entries {
		seen[e.Name()] = true
	}

	for d := range m.dirs {
		if !strings.HasPrefix(d, prefix) {
			continue
		}
		rest := strings.TrimPrefix(d, prefix)
		if rest == "" {
			continue
		}
		first := strings.SplitN(rest, "/", 2)[0]
		if !seen[first] {
			seen[first] = true
			entries = append(entries, &dirEntry{name: first})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

func (m *Medium) Stat(p string) (fs.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	if p == "" {
		return &fileInfo{name: ".", isDir: true, mode: fs.ModeDir | 0755}, nil
	}

	info, err := m.dn.Stat(p)
	if err == nil {
		return info, nil
	}

	if m.dirs[p] {
		return &fileInfo{name: path.Base(p), isDir: true, mode: fs.ModeDir | 0755}, nil
	}
	return nil, coreerr.E("datanode.Stat", "not found: "+p, os.ErrNotExist)
}

func (m *Medium) Open(p string) (fs.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	return m.dn.Open(p)
}

func (m *Medium) Create(p string) (goio.WriteCloser, error) {
	p = clean(p)
	if p == "" {
		return nil, coreerr.E("datanode.Create", "empty path", os.ErrInvalid)
	}
	return &writeCloser{m: m, path: p}, nil
}

func (m *Medium) Append(p string) (goio.WriteCloser, error) {
	p = clean(p)
	if p == "" {
		return nil, coreerr.E("datanode.Append", "empty path", os.ErrInvalid)
	}

	// Read existing content
	var existing []byte
	m.mu.RLock()
	f, err := m.dn.Open(p)
	if err == nil {
		existing, _ = goio.ReadAll(f)
		f.Close()
	}
	m.mu.RUnlock()

	return &writeCloser{m: m, path: p, buf: existing}, nil
}

func (m *Medium) ReadStream(p string) (goio.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	f, err := m.dn.Open(p)
	if err != nil {
		return nil, coreerr.E("datanode.ReadStream", "not found: "+p, os.ErrNotExist)
	}
	return f.(goio.ReadCloser), nil
}

func (m *Medium) WriteStream(p string) (goio.WriteCloser, error) {
	return m.Create(p)
}

func (m *Medium) Exists(p string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	if p == "" {
		return true // root always exists
	}
	_, err := m.dn.Stat(p)
	if err == nil {
		return true
	}
	return m.dirs[p]
}

func (m *Medium) IsDir(p string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p = clean(p)
	if p == "" {
		return true
	}
	info, err := m.dn.Stat(p)
	if err == nil {
		return info.IsDir()
	}
	return m.dirs[p]
}

// --- internal helpers ---

// hasPrefixLocked checks if any file path starts with prefix. Caller holds lock.
func (m *Medium) hasPrefixLocked(prefix string) bool {
	entries, _ := m.collectAllLocked()
	for _, name := range entries {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	for d := range m.dirs {
		if strings.HasPrefix(d, prefix) {
			return true
		}
	}
	return false
}

// collectAllLocked returns all file paths in the DataNode. Caller holds lock.
func (m *Medium) collectAllLocked() ([]string, error) {
	var names []string
	err := fs.WalkDir(m.dn, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			names = append(names, p)
		}
		return nil
	})
	return names, err
}

// removeFileLocked removes a single file by rebuilding the DataNode.
// This is necessary because Borg's DataNode doesn't expose a Remove method.
// Caller must hold m.mu write lock.
func (m *Medium) removeFileLocked(target string) {
	entries, _ := m.collectAllLocked()
	newDN := datanode.New()
	for _, name := range entries {
		if name == target {
			continue
		}
		f, err := m.dn.Open(name)
		if err != nil {
			continue
		}
		data, err := goio.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}
		newDN.AddData(name, data)
	}
	m.dn = newDN
}

// --- writeCloser buffers writes and flushes to DataNode on Close ---

type writeCloser struct {
	m    *Medium
	path string
	buf  []byte
}

func (w *writeCloser) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *writeCloser) Close() error {
	w.m.mu.Lock()
	defer w.m.mu.Unlock()

	w.m.dn.AddData(w.path, w.buf)
	w.m.ensureDirsLocked(path.Dir(w.path))
	return nil
}

// --- fs types for explicit directories ---

type dirEntry struct {
	name string
}

func (d *dirEntry) Name() string      { return d.name }
func (d *dirEntry) IsDir() bool       { return true }
func (d *dirEntry) Type() fs.FileMode { return fs.ModeDir }
func (d *dirEntry) Info() (fs.FileInfo, error) {
	return &fileInfo{name: d.name, isDir: true, mode: fs.ModeDir | 0755}, nil
}

type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.isDir }
func (fi *fileInfo) Sys() any           { return nil }
