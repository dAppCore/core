// Sandboxed local filesystem I/O for the Core framework.
package core

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Fs is a sandboxed local filesystem backend.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.EnsureDir("logs")
//	if !r.OK { return r }
type Fs struct {
	root string
}

// FS is a generic filesystem accepted by Mount and Extract.
//
//	fsys := core.DirFS("templates")
//	r := core.Mount(fsys, ".")
type FS = fs.FS

// FsFile is a file opened from an FS.
//
//	r := emb.Open("README.md")
//	if r.OK { file := r.Value.(core.FsFile); _ = file }
type FsFile = fs.File

// FsDirEntry is a directory entry returned by filesystem walkers.
//
//	r := emb.ReadDir(".")
//	if r.OK { entries := r.Value.([]core.FsDirEntry); _ = entries }
type FsDirEntry = fs.DirEntry

// WalkDirFunc visits one path during a filesystem walk.
//
//	fn := func(path string, d core.FsDirEntry, err error) error { return err }
//	_ = fn
type WalkDirFunc = fs.WalkDirFunc

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
			if r := UserCurrent(); r.OK {
				username = r.Value.(*User).Username
			}
			Print(os.Stderr, "[%s] SECURITY sandbox escape detected root=%s path=%s attempted=%s user=%s",
				Now().Format(TimeRFC3339), root, p, realNext, username)
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.Read("config/agent.json")
//	if r.OK { core.Println(r.Value.(string)) }
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.Write("config/agent.json", `{"host":"homelab.lthn.sh"}`)
//	if !r.OK { return r }
func (m *Fs) Write(p, content string) Result {
	return m.WriteMode(p, content, 0644)
}

// WriteMode saves content to file with explicit permissions.
// Use 0600 for sensitive files (encryption output, private keys, auth hashes).
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.WriteMode("secrets/token", "lethean-token", 0o600)
//	if !r.OK { return r }
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
func DirFS(dir string) FS {
	return os.DirFS(dir)
}

// ReadDir reads a directory from fsys.
//
//	r := core.ReadDir(core.DirFS("templates"), ".")
func ReadDir(fsys FS, name string) Result {
	return Result{}.New(fs.ReadDir(fsys, name))
}

// ReadFile reads a file from fsys.
//
//	r := core.ReadFile(core.DirFS("templates"), "README.md")
func ReadFile(fsys FS, name string) Result {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return Result{err, false}
	}
	return Result{data, true}
}

// Sub returns an FS rooted at dir inside fsys.
//
//	r := core.Sub(core.DirFS("templates"), "agent")
func Sub(fsys FS, dir string) Result {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		return Result{err, false}
	}
	return Result{sub, true}
}

// WalkDir walks fsys from root, calling fn for each file or directory.
//
//	err := core.WalkDir(core.DirFS("templates"), ".", fn)
func WalkDir(fsys FS, root string, fn WalkDirFunc) error {
	return fs.WalkDir(fsys, root, fn)
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.EnsureDir("logs/agent")
//	if !r.OK { return r }
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	if fsys.IsDir("logs") { core.Println("logs ready") }
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	if fsys.IsFile("config/agent.json") { core.Println("config ready") }
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	if fsys.Exists("config/agent.json") { core.Println("config present") }
func (m *Fs) Exists(p string) bool {
	vp := m.validatePath(p)
	if !vp.OK {
		return false
	}
	_, err := os.Stat(vp.Value.(string))
	return err == nil
}

// List returns directory entries.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.List("config")
//	if !r.OK { return r }
func (m *Fs) List(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.New(os.ReadDir(vp.Value.(string)))
}

// Stat returns file info.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.Stat("config/agent.json")
//	if !r.OK { return r }
func (m *Fs) Stat(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.New(os.Stat(vp.Value.(string)))
}

// Open opens the named file for reading.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.ReadStream("config/agent.json")
//	if !r.OK { return r }
//	defer r.Value.(core.ReadCloser).Close()
func (m *Fs) Open(p string) Result {
	vp := m.validatePath(p)
	if !vp.OK {
		return vp
	}
	return Result{}.New(os.Open(vp.Value.(string)))
}

// Create creates or truncates the named file.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.WriteStream("logs/agent.log")
//	if !r.OK { return r }
//	defer r.Value.(core.WriteCloser).Close()
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.Append("logs/agent.log")
//	if !r.OK { return r }
//	defer r.Value.(core.WriteCloser).Close()
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.ReadStream("config/agent.json")
//	if !r.OK { return r }
//	defer r.Value.(core.ReadCloser).Close()
func (m *Fs) ReadStream(path string) Result {
	return m.Open(path)
}

// WriteStream returns a writer for the file content.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.WriteStream("logs/agent.log")
//	if !r.OK { return r }
//	defer r.Value.(core.WriteCloser).Close()
func (m *Fs) WriteStream(path string) Result {
	return m.Create(path)
}

// WriteAll writes content to a writer and closes it if it implements Closer.
//
//	r := fs.WriteStream(path)
//	core.WriteAll(r.Value, "content")
func WriteAll(writer any, content string) Result {
	wc, ok := writer.(Writer)
	if !ok {
		return Result{E("core.WriteAll", "not a writer", nil), false}
	}
	_, err := wc.Write([]byte(content))
	if closer, ok := writer.(Closer); ok {
		closer.Close()
	}
	if err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}

// CloseStream closes any value that implements Closer.
//
//	core.CloseStream(r.Value)
func CloseStream(v any) {
	if closer, ok := v.(Closer); ok {
		closer.Close()
	}
}

// Delete removes a file or empty directory.
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.Delete("logs/old-agent.log")
//	if !r.OK { return r }
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.DeleteAll("tmp/session-42")
//	if !r.OK { return r }
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
//
//	fsys := (&core.Fs{}).New("/tmp/agent-workspace")
//	r := fsys.Rename("config/agent.tmp", "config/agent.json")
//	if !r.OK { return r }
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
//
//	entry := core.FsEntry{Path: "config/agent.json", Name: "agent.json", IsDir: false, Mode: 0o644}
//	core.Println(entry.Path)
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
func (m *Fs) WalkSeq(root string) Seq2[FsEntry, error] {
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
func (m *Fs) WalkSeqSkip(root string, skipNames ...string) Seq2[FsEntry, error] {
	skip := make(map[string]struct{}, len(skipNames))
	for _, name := range skipNames {
		if name != "" {
			skip[name] = struct{}{}
		}
	}
	return m.walkSeq(root, skip)
}

// walkSeq is the shared implementation behind WalkSeq and WalkSeqSkip.
func (m *Fs) walkSeq(root string, skip map[string]struct{}) Seq2[FsEntry, error] {
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
