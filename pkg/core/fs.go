// Sandboxed local filesystem I/O for the Core framework.
package core

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// Fs is a sandboxed local filesystem backend.
type Fs struct {
	root string
}

// NewIO creates a Fs rooted at the given directory.
// Pass "/" for full filesystem access, or a specific path to sandbox.
func NewIO(root string) (*Fs, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	// Resolve symlinks so sandbox checks compare like-for-like.
	// On macOS, /var is a symlink to /private/var — without this,
	// EvalSymlinks on child paths resolves to /private/var/... while
	// root stays /var/..., causing false sandbox escape detections.
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return &Fs{root: abs}, nil
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
func (m *Fs) validatePath(p string) (string, error) {
	if m.root == "/" {
		return m.path(p), nil
	}

	// Split the cleaned path into components
	parts := strings.Split(filepath.Clean("/"+p), string(os.PathSeparator))
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
			return "", err
		}

		// Verify the resolved part is still within the root
		rel, err := filepath.Rel(m.root, realNext)
		if err != nil || strings.HasPrefix(rel, "..") {
			// Security event: sandbox escape attempt
			username := "unknown"
			if u, err := user.Current(); err == nil {
				username = u.Username
			}
			fmt.Fprintf(os.Stderr, "[%s] SECURITY sandbox escape detected root=%s path=%s attempted=%s user=%s\n",
				time.Now().Format(time.RFC3339), m.root, p, realNext, username)
			return "", os.ErrPermission // Path escapes sandbox
		}
		current = realNext
	}

	return current, nil
}

// Read returns file contents as string.
func (m *Fs) Read(p string) (string, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write saves content to file, creating parent directories as needed.
// Files are created with mode 0644. For sensitive files (keys, secrets),
// use WriteMode with 0600.
func (m *Fs) Write(p, content string) error {
	return m.WriteMode(p, content, 0644)
}

// WriteMode saves content to file with explicit permissions.
// Use 0600 for sensitive files (encryption output, private keys, auth hashes).
func (m *Fs) WriteMode(p, content string, mode os.FileMode) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), mode)
}

// EnsureDir creates directory if it doesn't exist.
func (m *Fs) EnsureDir(p string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	return os.MkdirAll(full, 0755)
}

// IsDir returns true if path is a directory.
func (m *Fs) IsDir(p string) bool {
	if p == "" {
		return false
	}
	full, err := m.validatePath(p)
	if err != nil {
		return false
	}
	info, err := os.Stat(full)
	return err == nil && info.IsDir()
}

// IsFile returns true if path is a regular file.
func (m *Fs) IsFile(p string) bool {
	if p == "" {
		return false
	}
	full, err := m.validatePath(p)
	if err != nil {
		return false
	}
	info, err := os.Stat(full)
	return err == nil && info.Mode().IsRegular()
}

// Exists returns true if path exists.
func (m *Fs) Exists(p string) bool {
	full, err := m.validatePath(p)
	if err != nil {
		return false
	}
	_, err = os.Stat(full)
	return err == nil
}

// List returns directory entries.
func (m *Fs) List(p string) ([]fs.DirEntry, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	return os.ReadDir(full)
}

// Stat returns file info.
func (m *Fs) Stat(p string) (fs.FileInfo, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	return os.Stat(full)
}

// Open opens the named file for reading.
func (m *Fs) Open(p string) (fs.File, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	return os.Open(full)
}

// Create creates or truncates the named file.
func (m *Fs) Create(p string) (io.WriteCloser, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return nil, err
	}
	return os.Create(full)
}

// Append opens the named file for appending, creating it if it doesn't exist.
func (m *Fs) Append(p string) (io.WriteCloser, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// ReadStream returns a reader for the file content.
func (m *Fs) ReadStream(path string) (io.ReadCloser, error) {
	return m.Open(path)
}

// WriteStream returns a writer for the file content.
func (m *Fs) WriteStream(path string) (io.WriteCloser, error) {
	return m.Create(path)
}

// Delete removes a file or empty directory.
func (m *Fs) Delete(p string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	if full == "/" || full == os.Getenv("HOME") {
		return E("core.Delete", "refusing to delete protected path: "+full, nil)
	}
	return os.Remove(full)
}

// DeleteAll removes a file or directory recursively.
func (m *Fs) DeleteAll(p string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	if full == "/" || full == os.Getenv("HOME") {
		return E("core.DeleteAll", "refusing to delete protected path: "+full, nil)
	}
	return os.RemoveAll(full)
}

// Rename moves a file or directory.
func (m *Fs) Rename(oldPath, newPath string) error {
	oldFull, err := m.validatePath(oldPath)
	if err != nil {
		return err
	}
	newFull, err := m.validatePath(newPath)
	if err != nil {
		return err
	}
	return os.Rename(oldFull, newFull)
}
