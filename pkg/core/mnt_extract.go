// SPDX-License-Identifier: EUPL-1.2

package core

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

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
			name = strings.ReplaceAll(name, filter, "")
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
		if strings.Contains(filename, f) {
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
