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
	"strings"
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
func GetAsset(group, name string) (string, error) {
	assetGroupsMu.RLock()
	g, ok := assetGroups[group]
	assetGroupsMu.RUnlock()
	if !ok {
		return "", E("core.GetAsset", Join(" ", "asset group", group, "not found"), nil)
	}
	data, ok := g.assets[name]
	if !ok {
		return "", E("core.GetAsset", Join(" ", "asset", name, "not found in group", group), nil)
	}
	return decompress(data)
}

// GetAssetBytes retrieves a packed asset as bytes.
func GetAssetBytes(group, name string) ([]byte, error) {
	s, err := GetAsset(group, name)
	return []byte(s), err
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
	PackageName string
	BaseDir     string
	Groups      []string
	Assets      []AssetRef
}

// ScanAssets parses Go source files and finds asset references.
// Looks for calls to: core.GetAsset("group", "name"), core.AddAsset, etc.
func ScanAssets(filenames []string) ([]ScannedPackage, error) {
	packageMap := make(map[string]*ScannedPackage)
	var scanErr error

	for _, filename := range filenames {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
		if err != nil {
			return nil, err
		}

		baseDir := filepath.Dir(filename)
		pkg, ok := packageMap[baseDir]
		if !ok {
			pkg = &ScannedPackage{BaseDir: baseDir}
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
			return nil, scanErr
		}
	}

	var result []ScannedPackage
	for _, pkg := range packageMap {
		result = append(result, *pkg)
	}
	return result, nil
}

// GeneratePack creates Go source code that embeds the scanned assets.
func GeneratePack(pkg ScannedPackage) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", pkg.PackageName))
	b.WriteString("// Code generated by core pack. DO NOT EDIT.\n\n")

	if len(pkg.Assets) == 0 && len(pkg.Groups) == 0 {
		return b.String(), nil
	}

	b.WriteString("import \"forge.lthn.ai/core/go/pkg/core\"\n\n")
	b.WriteString("func init() {\n")

	// Pack groups (entire directories)
	packed := make(map[string]bool)
	for _, groupPath := range pkg.Groups {
		files, err := getAllFiles(groupPath)
		if err != nil {
			return "", Wrap(err, "core.GeneratePack", Join(" ", "failed to scan asset group", groupPath))
		}
		for _, file := range files {
			if packed[file] {
				continue
			}
			data, err := compressFile(file)
			if err != nil {
				return "", Wrap(err, "core.GeneratePack", Join(" ", "failed to compress asset", file, "in group", groupPath))
			}
			localPath := TrimPrefix(file, groupPath+"/")
			relGroup, err := filepath.Rel(pkg.BaseDir, groupPath)
			if err != nil {
				return "", Wrap(err, "core.GeneratePack", Join(" ", "could not determine relative path for group", groupPath, "(base", Concat(pkg.BaseDir, ")")))
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
			return "", Wrap(err, "core.GeneratePack", Join(" ", "failed to compress asset", asset.FullPath))
		}
		b.WriteString(fmt.Sprintf("\tcore.AddAsset(%q, %q, %q)\n", asset.Group, asset.Name, data))
		packed[asset.FullPath] = true
	}

	b.WriteString("}\n")
	return b.String(), nil
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
	b64 := base64.NewDecoder(base64.StdEncoding, strings.NewReader(input))
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
// Works with embed.FS, os.DirFS, or any fs.FS implementation.
func Mount(fsys fs.FS, basedir string) (*Embed, error) {
	s := &Embed{fsys: fsys, basedir: basedir}

	// If it's an embed.FS, keep a reference for EmbedFS()
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
func MountEmbed(efs embed.FS, basedir string) (*Embed, error) {
	return Mount(efs, basedir)
}

func (s *Embed) path(name string) string {
	return filepath.ToSlash(filepath.Join(s.basedir, name))
}

// Open opens the named file for reading.
func (s *Embed) Open(name string) (fs.File, error) {
	return s.fsys.Open(s.path(name))
}

// ReadDir reads the named directory.
func (s *Embed) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(s.fsys, s.path(name))
}

// ReadFile reads the named file.
func (s *Embed) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(s.fsys, s.path(name))
}

// ReadString reads the named file as a string.
func (s *Embed) ReadString(name string) (string, error) {
	data, err := s.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Sub returns a new Embed anchored at a subdirectory within this mount.
func (s *Embed) Sub(subDir string) (*Embed, error) {
	sub, err := fs.Sub(s.fsys, s.path(subDir))
	if err != nil {
		return nil, err
	}
	return &Embed{fsys: sub, basedir: "."}, nil
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

// BaseDir returns the basedir this Embed is anchored at.
func (s *Embed) BaseDir() string {
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
func Extract(fsys fs.FS, targetDir string, data any, opts ...ExtractOptions) error {
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
		return err
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
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
		return err
	}

	// Create directories (names may contain templates)
	for _, dir := range dirs {
		target := renderPath(filepath.Join(targetDir, dir), data)
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	// Process template files
	for _, path := range templateFiles {
		tmpl, err := template.ParseFS(fsys, path)
		if err != nil {
			return err
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
		targetFile = filepath.Join(dir, name)

		f, err := os.Create(targetFile)
		if err != nil {
			return err
		}
		if err := tmpl.Execute(f, data); err != nil {
			f.Close()
			return err
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
		target := renderPath(filepath.Join(targetDir, targetPath), data)
		if err := copyFile(fsys, path, target); err != nil {
			return err
		}
	}

	return nil
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
