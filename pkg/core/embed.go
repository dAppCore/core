// SPDX-License-Identifier: EUPL-1.2

// Embedded assets for the Core framework.
//
// Embed provides scoped filesystem access for go:embed and any fs.FS.
// Also includes build-time asset packing (AST scanner + compressor)
// and template-based directory extraction.
//
// Usage (mount):
//
//	sub, _ := core.Mount(myFS, "lib/persona")
//	content, _ := sub.ReadString("secops/developer.md")
//
// Usage (extract):
//
//	core.Extract(fsys, "/tmp/workspace", data)
//
// Usage (pack):
//
//	refs, _ := core.ScanAssets([]string{"main.go"})
//	source, _ := core.GeneratePack(refs)
package core

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/base64"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// --- Runtime: Asset Registry ---

// AssetGroup holds a named collection of packed assets.
type AssetGroup struct {
	assets map[string]string // name → compressed data
}

var (
	assetGroups   = make(map[string]*AssetGroup)
	assetGroupsMu sync.RWMutex
)

// AddAsset registers a packed asset at runtime (called from generated init()).
func AddAsset(group, name, data string) {
	assetGroupsMu.Lock()
	defer assetGroupsMu.Unlock()

	g, ok := assetGroups[group]
	if !ok {
		g = &AssetGroup{assets: make(map[string]string)}
		assetGroups[group] = g
	}
	g.assets[name] = data
}

// GetAsset retrieves and decompresses a packed asset.
//
//	r := core.GetAsset("mygroup", "greeting")
//	if r.OK { content := r.Value.(string) }
func GetAsset(group, name string) Result {
	assetGroupsMu.RLock()
	g, ok := assetGroups[group]
	if !ok {
		assetGroupsMu.RUnlock()
		return Result{}
	}
	data, ok := g.assets[name]
	assetGroupsMu.RUnlock()
	if !ok {
		return Result{}
	}
	s, err := decompress(data)
	if err != nil {
		return Result{err, false}
	}
	return Result{s, true}
}

// GetAssetBytes retrieves a packed asset as bytes.
//
//	r := core.GetAssetBytes("mygroup", "file")
//	if r.OK { data := r.Value.([]byte) }
func GetAssetBytes(group, name string) Result {
	r := GetAsset(group, name)
	if !r.OK {
		return r
	}
	return Result{[]byte(r.Value.(string)), true}
}

// --- Build-time: AST Scanner ---

// AssetRef is a reference to an asset found in source code.
type AssetRef struct {
	Name     string
	Path     string
	Group    string
	FullPath string
}

// ScannedPackage holds all asset references from a set of source files.
type ScannedPackage struct {
	PackageName   string
	BaseDirectory string
	Groups        []string
	Assets        []AssetRef
}

// ScanAssets parses Go source files and finds asset references.
// Looks for calls to: core.GetAsset("group", "name"), core.AddAsset, etc.
func ScanAssets(filenames []string) Result {
	packageMap := make(map[string]*ScannedPackage)
	var scanErr error

	for _, filename := range filenames {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
		if err != nil {
			return Result{err, false}
		}

		baseDir := filepath.Dir(filename)
		pkg, ok := packageMap[baseDir]
		if !ok {
			pkg = &ScannedPackage{BaseDirectory: baseDir}
			packageMap[baseDir] = pkg
		}
		pkg.PackageName = node.Name.Name

		ast.Inspect(node, func(n ast.Node) bool {
			if scanErr != nil {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			// Look for core.GetAsset or mewn.String patterns
			if ident.Name == "core" || ident.Name == "mewn" {
				switch sel.Sel.Name {
				case "GetAsset", "GetAssetBytes", "String", "MustString", "Bytes", "MustBytes":
					if len(call.Args) >= 1 {
						if lit, ok := call.Args[len(call.Args)-1].(*ast.BasicLit); ok {
							path := TrimPrefix(TrimSuffix(lit.Value, "\""), "\"")
							group := "."
							if len(call.Args) >= 2 {
								if glit, ok := call.Args[0].(*ast.BasicLit); ok {
									group = TrimPrefix(TrimSuffix(glit.Value, "\""), "\"")
								}
							}
							fullPath, err := filepath.Abs(filepath.Join(baseDir, group, path))
							if err != nil {
								scanErr = Wrap(err, "core.ScanAssets", Join(" ", "could not determine absolute path for asset", path, "in group", group))
								return false
							}
							pkg.Assets = append(pkg.Assets, AssetRef{
								Name: path,

								Group:    group,
								FullPath: fullPath,
							})
						}
					}
				case "Group":
					// Variable assignment: g := core.Group("./assets")
					if len(call.Args) == 1 {
						if lit, ok := call.Args[0].(*ast.BasicLit); ok {
							path := TrimPrefix(TrimSuffix(lit.Value, "\""), "\"")
							fullPath, err := filepath.Abs(filepath.Join(baseDir, path))
							if err != nil {
								scanErr = Wrap(err, "core.ScanAssets", Join(" ", "could not determine absolute path for group", path))
								return false
							}
							pkg.Groups = append(pkg.Groups, fullPath)
							// Track for variable resolution
						}
					}
				}
			}

			return true
		})
		if scanErr != nil {
			return Result{scanErr, false}
		}
	}

	var result []ScannedPackage
	for _, pkg := range packageMap {
		result = append(result, *pkg)
	}
	return Result{result, true}
}

// GeneratePack creates Go source code that embeds the scanned assets.
func GeneratePack(pkg ScannedPackage) Result {
	b := NewBuilder()

	b.WriteString(fmt.Sprintf("package %s\n\n", pkg.PackageName))
	b.WriteString("// Code generated by core pack. DO NOT EDIT.\n\n")

	if len(pkg.Assets) == 0 && len(pkg.Groups) == 0 {
		return Result{b.String(), true}
	}

	b.WriteString("import \"dappco.re/go/core/pkg/core\"\n\n")
	b.WriteString("func init() {\n")

	// Pack groups (entire directories)
	packed := make(map[string]bool)
	for _, groupPath := range pkg.Groups {
		files, err := getAllFiles(groupPath)
		if err != nil {
			return Result{err, false}
		}
		for _, file := range files {
			if packed[file] {
				continue
			}
			data, err := compressFile(file)
			if err != nil {
				return Result{err, false}
			}
			localPath := TrimPrefix(file, groupPath+"/")
			relGroup, err := filepath.Rel(pkg.BaseDirectory, groupPath)
			if err != nil {
				return Result{err, false}
			}
			b.WriteString(fmt.Sprintf("\tcore.AddAsset(%q, %q, %q)\n", relGroup, localPath, data))
			packed[file] = true
		}
	}

	// Pack individual assets
	for _, asset := range pkg.Assets {
		if packed[asset.FullPath] {
			continue
		}
		data, err := compressFile(asset.FullPath)
		if err != nil {
			return Result{err, false}
		}
		b.WriteString(fmt.Sprintf("\tcore.AddAsset(%q, %q, %q)\n", asset.Group, asset.Name, data))
		packed[asset.FullPath] = true
	}

	b.WriteString("}\n")
	return Result{b.String(), true}
}

// --- Compression ---

func compressFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return compress(string(data))
}

func compress(input string) (string, error) {
	var buf bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &buf)
	gz, err := gzip.NewWriterLevel(b64, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err := gz.Write([]byte(input)); err != nil {
		_ = gz.Close()
		_ = b64.Close()
		return "", err
	}
	if err := gz.Close(); err != nil {
		_ = b64.Close()
		return "", err
	}
	if err := b64.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func decompress(input string) (string, error) {
	b64 := base64.NewDecoder(base64.StdEncoding, NewReader(input))
	gz, err := gzip.NewReader(b64)
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return string(data), nil
}

func getAllFiles(dir string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

// --- Embed: Scoped Filesystem Mount ---

// Embed wraps an fs.FS with a basedir for scoped access.
// All paths are relative to basedir.
type Embed struct {
	basedir string
	fsys    fs.FS
	embedFS *embed.FS // original embed.FS for type-safe access via EmbedFS()
}

// Mount creates a scoped view of an fs.FS anchored at basedir.
//
//	r := core.Mount(myFS, "lib/prompts")
//	if r.OK { emb := r.Value.(*Embed) }
func Mount(fsys fs.FS, basedir string) Result {
	s := &Embed{fsys: fsys, basedir: basedir}

	if efs, ok := fsys.(embed.FS); ok {
		s.embedFS = &efs
	}

	if r := s.ReadDir("."); !r.OK {
		return r
	}
	return Result{s, true}
}

// MountEmbed creates a scoped view of an embed.FS.
//
//	r := core.MountEmbed(myFS, "testdata")
func MountEmbed(efs embed.FS, basedir string) Result {
	return Mount(efs, basedir)
}

func (s *Embed) path(name string) Result {
	joined := filepath.ToSlash(filepath.Join(s.basedir, name))
	if HasPrefix(joined, "..") || Contains(joined, "/../") || HasSuffix(joined, "/..") {
		return Result{E("embed.path", Concat("path traversal rejected: ", name), nil), false}
	}
	return Result{joined, true}
}

// Open opens the named file for reading.
//
//	r := emb.Open("test.txt")
//	if r.OK { file := r.Value.(fs.File) }
func (s *Embed) Open(name string) Result {
	r := s.path(name)
	if !r.OK {
		return r
	}
	f, err := s.fsys.Open(r.Value.(string))
	if err != nil {
		return Result{err, false}
	}
	return Result{f, true}
}

// ReadDir reads the named directory.
func (s *Embed) ReadDir(name string) Result {
	r := s.path(name)
	if !r.OK {
		return r
	}
	return Result{}.Result(fs.ReadDir(s.fsys, r.Value.(string)))
}

// ReadFile reads the named file.
//
//	r := emb.ReadFile("test.txt")
//	if r.OK { data := r.Value.([]byte) }
func (s *Embed) ReadFile(name string) Result {
	r := s.path(name)
	if !r.OK {
		return r
	}
	data, err := fs.ReadFile(s.fsys, r.Value.(string))
	if err != nil {
		return Result{err, false}
	}
	return Result{data, true}
}

// ReadString reads the named file as a string.
//
//	r := emb.ReadString("test.txt")
//	if r.OK { content := r.Value.(string) }
func (s *Embed) ReadString(name string) Result {
	r := s.ReadFile(name)
	if !r.OK {
		return r
	}
	return Result{string(r.Value.([]byte)), true}
}

// Sub returns a new Embed anchored at a subdirectory within this mount.
//
//	r := emb.Sub("testdata")
//	if r.OK { sub := r.Value.(*Embed) }
func (s *Embed) Sub(subDir string) Result {
	r := s.path(subDir)
	if !r.OK {
		return r
	}
	sub, err := fs.Sub(s.fsys, r.Value.(string))
	if err != nil {
		return Result{err, false}
	}
	return Result{&Embed{fsys: sub, basedir: "."}, true}
}

// FS returns the underlying fs.FS.
func (s *Embed) FS() fs.FS {
	return s.fsys
}

// EmbedFS returns the underlying embed.FS if mounted from one.
// Returns zero embed.FS if mounted from a non-embed source.
func (s *Embed) EmbedFS() embed.FS {
	if s.embedFS != nil {
		return *s.embedFS
	}
	return embed.FS{}
}

// BaseDirectory returns the base directory this Embed is anchored at.
func (s *Embed) BaseDirectory() string {
	return s.basedir
}

// --- Template Extraction ---

// ExtractOptions configures template extraction.
type ExtractOptions struct {
	// TemplateFilters identifies template files by substring match.
	// Default: [".tmpl"]
	TemplateFilters []string

	// IgnoreFiles is a set of filenames to skip during extraction.
	IgnoreFiles map[string]struct{}

	// RenameFiles maps original filenames to new names.
	RenameFiles map[string]string
}

// Extract copies a template directory from an fs.FS to targetDir,
// processing Go text/template in filenames and file contents.
//
// Files containing a template filter substring (default: ".tmpl") have
// their contents processed through text/template with the given data.
// The filter is stripped from the output filename.
//
// Directory and file names can contain Go template expressions:
// {{.Name}}/main.go → myproject/main.go
//
// Data can be any struct or map[string]string for template substitution.
func Extract(fsys fs.FS, targetDir string, data any, opts ...ExtractOptions) Result {
	opt := ExtractOptions{
		TemplateFilters: []string{".tmpl"},
		IgnoreFiles:     make(map[string]struct{}),
		RenameFiles:     make(map[string]string),
	}
	if len(opts) > 0 {
		if len(opts[0].TemplateFilters) > 0 {
			opt.TemplateFilters = opts[0].TemplateFilters
		}
		if opts[0].IgnoreFiles != nil {
			opt.IgnoreFiles = opts[0].IgnoreFiles
		}
		if opts[0].RenameFiles != nil {
			opt.RenameFiles = opts[0].RenameFiles
		}
	}

	// Ensure target directory exists
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return Result{err, false}
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return Result{err, false}
	}

	// Categorise files
	var dirs []string
	var templateFiles []string
	var standardFiles []string

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if d.IsDir() {
			dirs = append(dirs, path)
			return nil
		}
		filename := filepath.Base(path)
		if _, ignored := opt.IgnoreFiles[filename]; ignored {
			return nil
		}
		if isTemplate(filename, opt.TemplateFilters) {
			templateFiles = append(templateFiles, path)
		} else {
			standardFiles = append(standardFiles, path)
		}
		return nil
	})
	if err != nil {
		return Result{err, false}
	}

	// safePath ensures a rendered path stays under targetDir.
	safePath := func(rendered string) (string, error) {
		abs, err := filepath.Abs(rendered)
		if err != nil {
			return "", err
		}
		if !HasPrefix(abs, targetDir+string(filepath.Separator)) && abs != targetDir {
			return "", E("embed.Extract", Concat("path escapes target: ", abs), nil)
		}
		return abs, nil
	}

	// Create directories (names may contain templates)
	for _, dir := range dirs {
		target, err := safePath(renderPath(filepath.Join(targetDir, dir), data))
		if err != nil {
			return Result{err, false}
		}
		if err := os.MkdirAll(target, 0755); err != nil {
			return Result{err, false}
		}
	}

	// Process template files
	for _, path := range templateFiles {
		tmpl, err := template.ParseFS(fsys, path)
		if err != nil {
			return Result{err, false}
		}

		targetFile := renderPath(filepath.Join(targetDir, path), data)

		// Strip template filters from filename
		dir := filepath.Dir(targetFile)
		name := filepath.Base(targetFile)
		for _, filter := range opt.TemplateFilters {
			name = Replace(name, filter, "")
		}
		if renamed := opt.RenameFiles[name]; renamed != "" {
			name = renamed
		}
		targetFile, err = safePath(filepath.Join(dir, name))
		if err != nil {
			return Result{err, false}
		}

		f, err := os.Create(targetFile)
		if err != nil {
			return Result{err, false}
		}
		if err := tmpl.Execute(f, data); err != nil {
			f.Close()
			return Result{err, false}
		}
		f.Close()
	}

	// Copy standard files
	for _, path := range standardFiles {
		targetPath := path
		name := filepath.Base(path)
		if renamed := opt.RenameFiles[name]; renamed != "" {
			targetPath = filepath.Join(filepath.Dir(path), renamed)
		}
		target, err := safePath(renderPath(filepath.Join(targetDir, targetPath), data))
		if err != nil {
			return Result{err, false}
		}
		if err := copyFile(fsys, path, target); err != nil {
			return Result{err, false}
		}
	}

	return Result{OK: true}
}

func isTemplate(filename string, filters []string) bool {
	for _, f := range filters {
		if Contains(filename, f) {
			return true
		}
	}
	return false
}

func renderPath(path string, data any) string {
	if data == nil {
		return path
	}
	tmpl, err := template.New("path").Parse(path)
	if err != nil {
		return path
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return path
	}
	return buf.String()
}

func copyFile(fsys fs.FS, source, target string) error {
	s, err := fsys.Open(source)
	if err != nil {
		return err
	}
	defer s.Close()

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	d, err := os.Create(target)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}
