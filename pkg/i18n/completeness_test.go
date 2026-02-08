package i18n

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTranslationCompleteness_Good verifies every T() key in the source code
// has a translation in en_GB.json. Catches missing keys at test time instead
// of showing raw keys like "cmd.collect.short" in the CLI.
func TestTranslationCompleteness_Good(t *testing.T) {
	svc, err := New(WithMode(ModeStrict))
	require.NoError(t, err)

	// Find repo root (walk up from pkg/i18n/ to find go.mod)
	root := findRepoRoot(t)

	// Extract all T("key") calls from Go source
	keys := extractTranslationKeys(t, root)
	require.NotEmpty(t, keys, "should find translation keys in source code")

	var missing []string
	for _, key := range keys {
		// ModeStrict panics on missing — use recover to collect them all
		func() {
			defer func() {
				if r := recover(); r != nil {
					missing = append(missing, key)
				}
			}()
			svc.T(key)
		}()
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("found %d missing translation keys in en_GB.json:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

// findRepoRoot walks up from the test directory to find the repo root (containing go.mod).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// tCallRegex matches i18n.T("key"), T("key"), and cli.T("key") patterns.
var tCallRegex = regexp.MustCompile(`(?:i18n|cli)\.T\("([^"]+)"`)

// extractTranslationKeys scans all .go files (excluding tests and vendors)
// for T() calls and returns the unique set of translation keys.
func extractTranslationKeys(t *testing.T, root string) []string {
	t.Helper()
	seen := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		// Skip vendor, .git, and test files
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		matches := tCallRegex.FindAllSubmatch(data, -1)
		for _, m := range matches {
			key := string(m[1])
			// Only track cmd.* and common.* keys (skip dynamic/template keys)
			if strings.HasPrefix(key, "cmd.") || strings.HasPrefix(key, "common.") {
				seen[key] = true
			}
		}
		return nil
	})
	require.NoError(t, err)

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
