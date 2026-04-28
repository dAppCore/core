// Sandboxed local filesystem I/O for the Core framework.
package core

import (
	"io"
	"io/fs"
	"iter"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// Fs is a sandboxed local filesystem backend.
type Fs struct {
	root string
}

// New initialises an Fs with the given root directory.
// Root "/" means unrestricted access. Empty root defaults to "/".
//
//	fs := (&core.Fs{}).New("/")
func (m *Fs) New(root string) *Fs {
	if root == "" {
		root = "/"
	}
	m.root = root
	return m
}

// NewUnrestricted returns a new Fs with root "/", granting full filesystem access.
// Use this instead of unsafe.Pointer to bypass the sandbox.
//
//	fs := c.Fs().NewUnrestricted()
//	fs.Read("/etc/hostname")  // works — no sandbox
func (m *Fs) NewUnrestricted() *Fs {
	return (&Fs{}).New("/")
}

// Root returns the sandbox root path.
//
//	root := c.Fs().Root()  // e.g. "/home/agent/.core"
func (m *Fs) Root() string {
	if m.root == "" {
		return "/"
	}
	return m.root
}

// path sanitises and returns the full path.
// Absolute paths are sandboxed under root (unless root is "/").
// Empty root defaults to "/" — the zero value of Fs is usable.
func (m *Fs) path(p string) string {
	root := m.root
	if root == "" {
		root = "/"
	}
	if p == "" {
		return root
	}

	// If the path is relative and the medium is rooted at "/",
	// treat it as relative to the current working directory.
	// This makes io.Local behave more like the standard 'os' package.
	if root == "/" && !filepath.IsAbs(p) {
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, p)
	}

	// Use filepath.Clean with a leading slash to resolve all .. and . internally
	// before joining with the root. This is a standard way to sandbox paths.
	clean := filepath.Clean("/" + p)

	// If root is "/", allow absolute paths through
	if root == "/" {
		return clean
	}

	// Strip leading "/" so Join works correctly with root
	return filepath.Join(root, clean[1:])
}

// validatePath ensures the path is within the sandbox, following symlinks if they exist.
func (m *Fs) validatePath(p string) Result {
	root := m.root
	if root == "" {
		root = "/"
	}
	if root == "/" {
		return Result{m.path(p), true}
	}

	// Split the cleaned path into components
	parts := Split(filepath.Clean("/"+p), string(os.PathSeparator))
	current := root

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
		rel, err := filepath.Rel(root, realNext)
		if err != nil || HasPrefix(rel, "..") {
			// Security event: sandbox escape attempt
			username := "unknown"
			if u, err := user.Current(); err == nil {
				username = u.Username
			}
			Print(os.Stderr, "[%s] SECURITY sandbox escape detected root=%s path=%s attempted=%s user=%s",
				time.Now().Format(time.RFC3339), root, p, realNext, username)
			if err == nil {
				err = E("fs.validatePath", Concat("sandbox escape: ", p, " resolves outside ", m.root), nil)
			}
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

// TempDir creates a temporary directory and returns its path.
// The caller is responsible for cleanup via fs.DeleteAll().
//
//	dir := fs.TempDir("agent-workspace")
//	defer fs.DeleteAll(dir)
func (m *Fs) TempDir(prefix string) string {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return ""
	}
	return dir
}

// DirFS returns an fs.FS rooted at the given directory path.
//
//	fsys := core.DirFS("/path/to/templates")
func DirFS(dir string) fs.FS {
	return os.DirFS(dir)
}

// WriteAtomic writes content by writing to a temp file then renaming.
// Rename is atomic on POSIX — concurrent readers never see a partial file.
// Use this for status files, config, or any file read from multiple goroutines.
//
//	r := fs.WriteAtomic("/status.json", jsonData)
func (m *Fs) WriteAtomic(p, content string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return Result{err, false}
	}

	tmp := full + ".tmp." + shortRand()
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return Result{err, false}
	}
	if err := os.Rename(tmp, full); err != nil {
		os.Remove(tmp)
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
	return Result{}.New(os.ReadDir(vp.Value.(string)))
}

// Stat returns file info.
func (m *Fs) Stat(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.New(os.Stat(vp.Value.(string)))
}

// Open opens the named file for reading.
func (m *Fs) Open(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.New(os.Open(vp.Value.(string)))
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
	return Result{}.New(os.Create(full))
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
	return Result{}.New(os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644))
}

// ReadStream returns a reader for the file content.
func (m *Fs) ReadStream(path string) Result {
	return m.Open(path)
}

// WriteStream returns a writer for the file content.
func (m *Fs) WriteStream(path string) Result {
	return m.Create(path)
}

// ReadAll reads all bytes from a ReadCloser and closes it.
// Wraps io.ReadAll so consumers don't import "io".
//
//	r := fs.ReadStream(path)
//	data := core.ReadAll(r.Value)
func ReadAll(reader any) Result {
	rc, ok := reader.(io.Reader)
	if !ok {
		return Result{E("core.ReadAll", "not a reader", nil), false}
	}
	data, err := io.ReadAll(rc)
	if closer, ok := reader.(io.Closer); ok {
		closer.Close()
	}
	if err != nil {
		return Result{err, false}
	}
	return Result{string(data), true}
}

// WriteAll writes content to a writer and closes it if it implements Closer.
//
//	r := fs.WriteStream(path)
//	core.WriteAll(r.Value, "content")
func WriteAll(writer any, content string) Result {
	wc, ok := writer.(io.Writer)
	if !ok {
		return Result{E("core.WriteAll", "not a writer", nil), false}
	}
	_, err := wc.Write([]byte(content))
	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
	if err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// CloseStream closes any value that implements io.Closer.
//
//	core.CloseStream(r.Value)
func CloseStream(v any) {
	if closer, ok := v.(io.Closer); ok {
		closer.Close()
	}
}

// Delete removes a file or empty directory.
func (m *Fs) Delete(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	full := vp.Value.(string)
	if full == "/" || full == os.Getenv("HOME") {
		return Result{E("fs.Delete", Concat("refusing to delete protected path: ", full), nil), false}
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
		return Result{E("fs.DeleteAll", Concat("refusing to delete protected path: ", full), nil), false}
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

// FsEntry is a directory entry yielded by WalkSeq and WalkSeqSkip.
// Path is relative to the walk root.
type FsEntry struct {
	Path  string // relative to walk root, OS-native separator
	Name  string // basename
	IsDir bool
	Mode  fs.FileMode
}

// WalkSeq walks the directory tree rooted at root within the Fs sandbox,
// yielding every entry depth-first. Iteration stops on caller break.
//
// Symlinks are not followed: validatePath rejects any symlink that resolves
// outside the sandbox before descent, and the underlying walker uses
// filepath.WalkDir which does not traverse into symlinked directories.
//
//	for entry, err := range c.Fs().WalkSeq("./") {
//		if err != nil { break }
//		if !entry.IsDir { /* file */ }
//	}
func (m *Fs) WalkSeq(root string) iter.Seq2[FsEntry, error] {
	return m.walkSeq(root, nil)
}

// WalkSeqSkip walks like WalkSeq but skips any directory whose basename
// appears in skipNames (e.g. "vendor", "node_modules", ".git"). Skipped
// directories are not descended into; their contents are never yielded.
// The walk root itself is never skipped, even if its basename matches.
//
//	for entry, err := range c.Fs().WalkSeqSkip("./", "vendor", "node_modules", ".git") {
//		if err != nil { break }
//		if !entry.IsDir { /* file */ }
//	}
func (m *Fs) WalkSeqSkip(root string, skipNames ...string) iter.Seq2[FsEntry, error] {
	skip := make(map[string]struct{}, len(skipNames))
	for _, name := range skipNames {
		if name != "" {
			skip[name] = struct{}{}
		}
	}
	return m.walkSeq(root, skip)
}

// walkSeq is the shared implementation behind WalkSeq and WalkSeqSkip.
func (m *Fs) walkSeq(root string, skip map[string]struct{}) iter.Seq2[FsEntry, error] {
	return func(yield func(FsEntry, error) bool) {
		vp := m.validatePath(root)
		if !vp.OK {
			err, _ := vp.Value.(error)
			if err == nil {
				err = E("fs.WalkSeq", "invalid walk root", nil)
			}
			yield(FsEntry{}, err)
			return
		}
		fullRoot, _ := vp.Value.(string)
		if fullRoot == "" {
			yield(FsEntry{}, E("fs.WalkSeq", "validatePath returned empty root", nil))
			return
		}
		stop := false
		_ = filepath.WalkDir(fullRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if stop {
				return filepath.SkipAll
			}
			if walkErr != nil {
				if !yield(FsEntry{}, walkErr) {
					stop = true
					return filepath.SkipAll
				}
				return nil
			}
			if d.IsDir() && skip != nil && path != fullRoot {
				if _, ok := skip[d.Name()]; ok {
					return filepath.SkipDir
				}
			}
			rel, relErr := filepath.Rel(fullRoot, path)
			if relErr != nil {
				rel = path
			}
			info, _ := d.Info()
			mode := fs.FileMode(0)
			if info != nil {
				mode = info.Mode()
			}
			entry := FsEntry{
				Path:  rel,
				Name:  d.Name(),
				IsDir: d.IsDir(),
				Mode:  mode,
			}
			if !yield(entry, nil) {
				stop = true
				return filepath.SkipAll
			}
			return nil
		})
	}
}
