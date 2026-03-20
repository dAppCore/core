// Sandboxed local filesystem I/O for the Core framework.
package core

import (
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// Fs is a sandboxed local filesystem backend.
type Fs struct {
	root string
}

// path sanitises and returns the full path.
// Absolute paths are sandboxed under root (unless root is "/").
func (m *Fs) path(p string) string {
	if p == "" {
		return m.root
	}

	// If the path is relative and the medium is rooted at "/",
	// treat it as relative to the current working directory.
	// This makes io.Local behave more like the standard 'os' package.
	if m.root == "/" && !filepath.IsAbs(p) {
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, p)
	}

	// Use filepath.Clean with a leading slash to resolve all .. and . internally
	// before joining with the root. This is a standard way to sandbox paths.
	clean := filepath.Clean("/" + p)

	// If root is "/", allow absolute paths through
	if m.root == "/" {
		return clean
	}

	// Strip leading "/" so Join works correctly with root
	return filepath.Join(m.root, clean[1:])
}

// validatePath ensures the path is within the sandbox, following symlinks if they exist.
func (m *Fs) validatePath(p string) Result {
	if m.root == "/" {
		return Result{m.path(p), true}
	}

	// Split the cleaned path into components
	parts := Split(filepath.Clean("/"+p), string(os.PathSeparator))
	current := m.root

	for _, part := range parts {
		if part == "" {
			continue
		}

		next := filepath.Join(current, part)
		realNext, err := filepath.EvalSymlinks(next)
		if err != nil {
			if os.IsNotExist(err) {
				// Part doesn't exist, we can't follow symlinks anymore.
				// Since the path is already Cleaned and current is safe,
				// appending a component to current will not escape.
				current = next
				continue
			}
			return Result{err, false}
		}

		// Verify the resolved part is still within the root
		rel, err := filepath.Rel(m.root, realNext)
		if err != nil || HasPrefix(rel, "..") {
			// Security event: sandbox escape attempt
			username := "unknown"
			if u, err := user.Current(); err == nil {
				username = u.Username
			}
			Print(os.Stderr, "[%s] SECURITY sandbox escape detected root=%s path=%s attempted=%s user=%s",
				time.Now().Format(time.RFC3339), m.root, p, realNext, username)
			return Result{err, false}
		}
		current = realNext
	}

	return Result{current, true}
}

// Read returns file contents as string.
func (m *Fs) Read(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	data, err := os.ReadFile(vp.Value.(string))
	if err != nil {
		return Result{err, false}
	}
	return Result{string(data), true}
}

// Write saves content to file, creating parent directories as needed.
// Files are created with mode 0644. For sensitive files (keys, secrets),
// use WriteMode with 0600.
func (m *Fs) Write(p, content string) Result {
	return m.WriteMode(p, content, 0644)
}

// WriteMode saves content to file with explicit permissions.
// Use 0600 for sensitive files (encryption output, private keys, auth hashes).
func (m *Fs) WriteMode(p, content string, mode os.FileMode) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return Result{err, false}
	}
	if err := os.WriteFile(full, []byte(content), mode); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// EnsureDir creates directory if it doesn't exist.
func (m *Fs) EnsureDir(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	if err := os.MkdirAll(vp.Value.(string), 0755); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// IsDir returns true if path is a directory.
func (m *Fs) IsDir(p string) bool {
	if p == "" {
		return false
	}
	vp := m.validatePath(p)
	if !vp.OK {
		return false
	}
	info, err := os.Stat(vp.Value.(string))
	return err == nil && info.IsDir()
}

// IsFile returns true if path is a regular file.
func (m *Fs) IsFile(p string) bool {
	if p == "" {
		return false
	}
	vp := m.validatePath(p)
	if !vp.OK {
		return false
	}
	info, err := os.Stat(vp.Value.(string))
	return err == nil && info.Mode().IsRegular()
}

// Exists returns true if path exists.
func (m *Fs) Exists(p string) bool {
	vp := m.validatePath(p)
	if !vp.OK {
		return false
	}
	_, err := os.Stat(vp.Value.(string))
	return err == nil
}

// List returns directory entries.
func (m *Fs) List(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.Result(os.ReadDir(vp.Value.(string)))
}

// Stat returns file info.
func (m *Fs) Stat(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.Result(os.Stat(vp.Value.(string)))
}

// Open opens the named file for reading.
func (m *Fs) Open(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.Result(os.Open(vp.Value.(string)))
}

// Create creates or truncates the named file.
func (m *Fs) Create(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return Result{err, false}
	}
	return Result{}.Result(os.Create(full))
}

// Append opens the named file for appending, creating it if it doesn't exist.
func (m *Fs) Append(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return Result{err, false}
	}
	return Result{}.Result(os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644))
}

// ReadStream returns a reader for the file content.
func (m *Fs) ReadStream(path string) Result {
	return m.Open(path)
}

// WriteStream returns a writer for the file content.
func (m *Fs) WriteStream(path string) Result {
	return m.Create(path)
}

// Delete removes a file or empty directory.
func (m *Fs) Delete(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if full == "/" || full == os.Getenv("HOME") {
		return Result{}
	}
	if err := os.Remove(full); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// DeleteAll removes a file or directory recursively.
func (m *Fs) DeleteAll(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if full == "/" || full == os.Getenv("HOME") {
		return Result{}
	}
	if err := os.RemoveAll(full); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// Rename moves a file or directory.
func (m *Fs) Rename(oldPath, newPath string) Result {
	oldVp := m.validatePath(oldPath)
	if !oldVp.OK {
		return oldVp
	}
	newVp := m.validatePath(newPath)
	if !newVp.OK {
		return newVp
	}
	if err := os.Rename(oldVp.Value.(string), newVp.Value.(string)); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}
