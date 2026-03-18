// SPDX-License-Identifier: EUPL-1.2

// Build-time asset packing for the Core framework.
// Based on leaanthony/mewn — scans Go source AST for asset references,
// reads files, compresses, and generates Go source with embedded data.
//
// This enables asset embedding WITHOUT go:embed — the packer runs at
// build time and generates a .go file with init() that registers assets.
// This pattern works cross-language (Go, TypeScript, etc).
//
// Usage (build tool):
//
//	refs, _ := core.ScanAssets([]string{"main.go", "app.go"})
//	source, _ := core.GeneratePack(refs)
//	os.WriteFile("pack.go", []byte(source), 0644)
//
// Usage (runtime):
//
//	core.AddAsset(".", "template.html", compressedData)
//	content := core.GetAsset(".", "template.html")
package core

import (
	"compress/gzip"
	"encoding/base64"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// --- Runtime: Asset Registry ---

// AssetGroup holds a named collection of packed assets.
type AssetGroup struct {
	name   string
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
		g = &AssetGroup{name: group, assets: make(map[string]string)}
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
		return "", fmt.Errorf("asset group %q not found", group)
	}
	data, ok := g.assets[name]
	if !ok {
		return "", fmt.Errorf("asset %q not found in group %q", name, group)
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
	Name      string
	Path      string
	Group     string
	FullPath  string
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
	groupPaths := make(map[string]string) // variable name → path

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
							path := strings.Trim(lit.Value, "\"")
							group := "."
							if len(call.Args) >= 2 {
								if glit, ok := call.Args[0].(*ast.BasicLit); ok {
									group = strings.Trim(glit.Value, "\"")
								}
							}
							fullPath, _ := filepath.Abs(filepath.Join(baseDir, group, path))
							pkg.Assets = append(pkg.Assets, AssetRef{
								Name:     path,
								Path:     path,
								Group:    group,
								FullPath: fullPath,
							})
						}
					}
				case "Group":
					// Variable assignment: g := core.Group("./assets")
					if len(call.Args) == 1 {
						if lit, ok := call.Args[0].(*ast.BasicLit); ok {
							path := strings.Trim(lit.Value, "\"")
							fullPath, _ := filepath.Abs(filepath.Join(baseDir, path))
							pkg.Groups = append(pkg.Groups, fullPath)
							// Track for variable resolution
							groupPaths[path] = fullPath
						}
					}
				}
			}

			return true
		})
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
			continue
		}
		for _, file := range files {
			if packed[file] {
				continue
			}
			data, err := compressFile(file)
			if err != nil {
				continue
			}
			localPath := strings.TrimPrefix(file, groupPath+"/")
			relGroup, _ := filepath.Rel(pkg.BaseDir, groupPath)
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
			continue
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
		return "", err
	}
	gz.Close()
	b64.Close()
	return buf.String(), nil
}

func decompress(input string) (string, error) {
	b64 := base64.NewDecoder(base64.StdEncoding, strings.NewReader(input))
	gz, err := gzip.NewReader(b64)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	data, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getAllFiles(dir string) ([]string, error) {
	var result []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}
