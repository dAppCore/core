# core.help Hugo Documentation Site — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Hugo + Docsy documentation site at core.help that aggregates markdown from 39 repos via `core docs sync --target hugo`.

**Architecture:** Hugo static site with Docsy theme, populated by extending `core docs sync` with a `--target hugo` flag that maps repo docs into Hugo's `content/` tree with auto-injected front matter. Deploy to BunnyCDN.

**Tech Stack:** Hugo (Go SSG), Docsy theme (Hugo module), BunnyCDN, `core docs sync` CLI

---

## Context

The docs sync command lives in `/Users/snider/Code/host-uk/cli/cmd/docs/`. The site will be scaffolded at `/Users/snider/Code/host-uk/docs-site/`. The registry at `/Users/snider/Code/host-uk/.core/repos.yaml` already contains all 39 repos (20 PHP + 18 Go + 1 CLI) with explicit paths for Go repos.

Key files:
- `/Users/snider/Code/host-uk/cli/cmd/docs/cmd_sync.go` — sync command (modify)
- `/Users/snider/Code/host-uk/cli/cmd/docs/cmd_scan.go` — repo scanner (modify)
- `/Users/snider/Code/host-uk/docs-site/` — Hugo site (create)

## Task 1: Scaffold Hugo + Docsy site

**Files:**
- Create: `/Users/snider/Code/host-uk/docs-site/hugo.toml`
- Create: `/Users/snider/Code/host-uk/docs-site/go.mod`
- Create: `/Users/snider/Code/host-uk/docs-site/content/_index.md`
- Create: `/Users/snider/Code/host-uk/docs-site/content/getting-started/_index.md`
- Create: `/Users/snider/Code/host-uk/docs-site/content/cli/_index.md`
- Create: `/Users/snider/Code/host-uk/docs-site/content/go/_index.md`
- Create: `/Users/snider/Code/host-uk/docs-site/content/mcp/_index.md`
- Create: `/Users/snider/Code/host-uk/docs-site/content/php/_index.md`
- Create: `/Users/snider/Code/host-uk/docs-site/content/kb/_index.md`

This is the one-time Hugo scaffolding. No tests — just files.

**`hugo.toml`:**
```toml
baseURL = "https://core.help/"
title = "Core Documentation"
languageCode = "en"
defaultContentLanguage = "en"

enableRobotsTXT = true
enableGitInfo = false

[outputs]
home = ["HTML", "JSON"]
section = ["HTML"]

[params]
description = "Documentation for the Core CLI, Go packages, PHP modules, and MCP tools"
copyright = "Host UK — EUPL-1.2"

[params.ui]
sidebar_menu_compact = true
breadcrumb_disable = false
sidebar_search_disable = false
navbar_logo = false

[params.ui.readingtime]
enable = false

[module]
proxy = "direct"

[module.hugoVersion]
extended = true
min = "0.120.0"

[[module.imports]]
path = "github.com/google/docsy"
disable = false

[markup.goldmark.renderer]
unsafe = true

[menu]
[[menu.main]]
name = "Getting Started"
weight = 10
url = "/getting-started/"
[[menu.main]]
name = "CLI Reference"
weight = 20
url = "/cli/"
[[menu.main]]
name = "Go Packages"
weight = 30
url = "/go/"
[[menu.main]]
name = "MCP Tools"
weight = 40
url = "/mcp/"
[[menu.main]]
name = "PHP Packages"
weight = 50
url = "/php/"
[[menu.main]]
name = "Knowledge Base"
weight = 60
url = "/kb/"
```

**`go.mod`:**
```
module github.com/host-uk/docs-site

go 1.22

require github.com/google/docsy v0.11.0
```

Note: Run `hugo mod get` after creating these files to populate `go.sum` and download Docsy.

**Section `_index.md` files** — each needs Hugo front matter:

`content/_index.md`:
```markdown
---
title: "Core Documentation"
description: "Documentation for the Core CLI, Go packages, PHP modules, and MCP tools"
---

Welcome to the Core ecosystem documentation.

## Sections

- [Getting Started](/getting-started/) — Installation, configuration, and first steps
- [CLI Reference](/cli/) — Command reference for `core` CLI
- [Go Packages](/go/) — Go ecosystem package documentation
- [MCP Tools](/mcp/) — Model Context Protocol tool reference
- [PHP Packages](/php/) — PHP module documentation
- [Knowledge Base](/kb/) — Wiki articles and deep dives
```

`content/getting-started/_index.md`:
```markdown
---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 10
description: "Installation, configuration, and first steps with the Core CLI"
---
```

`content/cli/_index.md`:
```markdown
---
title: "CLI Reference"
linkTitle: "CLI Reference"
weight: 20
description: "Command reference for the core CLI tool"
---
```

`content/go/_index.md`:
```markdown
---
title: "Go Packages"
linkTitle: "Go Packages"
weight: 30
description: "Documentation for the Go ecosystem packages"
---
```

`content/mcp/_index.md`:
```markdown
---
title: "MCP Tools"
linkTitle: "MCP Tools"
weight: 40
description: "Model Context Protocol tool reference — file operations, RAG, ML inference, process management"
---
```

`content/php/_index.md`:
```markdown
---
title: "PHP Packages"
linkTitle: "PHP Packages"
weight: 50
description: "Documentation for the PHP module ecosystem"
---
```

`content/kb/_index.md`:
```markdown
---
title: "Knowledge Base"
linkTitle: "Knowledge Base"
weight: 60
description: "Wiki articles, deep dives, and reference material"
---
```

**Verify:** After creating files, run from `/Users/snider/Code/host-uk/docs-site/`:
```bash
hugo mod get
hugo server
```
The site should start and show the landing page with Docsy theme at `localhost:1313`.

**Commit:**
```bash
cd /Users/snider/Code/host-uk/docs-site
git init
git add .
git commit -m "feat: scaffold Hugo + Docsy documentation site"
```

---

## Task 2: Extend scanRepoDocs to collect KB/ and README

**Files:**
- Modify: `/Users/snider/Code/host-uk/cli/cmd/docs/cmd_scan.go`

Currently `scanRepoDocs` only collects files from `docs/`. For the Hugo target we also need:
- `KB/**/*.md` files (wiki pages from go-mlx, go-i18n)
- `README.md` content (becomes the package _index.md)

Add a `KBFiles []string` field to `RepoDocInfo` and scan `KB/` alongside `docs/`:

```go
type RepoDocInfo struct {
    Name      string
    Path      string
    HasDocs   bool
    Readme    string
    ClaudeMd  string
    Changelog string
    DocsFiles []string // All files in docs/ directory (recursive)
    KBFiles   []string // All files in KB/ directory (recursive)
}
```

In `scanRepoDocs`, after the `docs/` walk, add a second walk for `KB/`:

```go
// Recursively scan KB/ directory for .md files
kbDir := filepath.Join(repo.Path, "KB")
if _, err := io.Local.List(kbDir); err == nil {
    _ = filepath.WalkDir(kbDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
            return nil
        }
        relPath, _ := filepath.Rel(kbDir, path)
        info.KBFiles = append(info.KBFiles, relPath)
        info.HasDocs = true
        return nil
    })
}
```

**Tests:** The existing tests should still pass. No new test file needed — this is a data-collection change.

**Verify:** `cd /Users/snider/Code/host-uk/cli && GOWORK=off go build ./cmd/docs/...`

**Commit:**
```bash
git add cmd/docs/cmd_scan.go
git commit -m "feat(docs): scan KB/ directory alongside docs/"
```

---

## Task 3: Add `--target hugo` flag and Hugo sync logic

**Files:**
- Modify: `/Users/snider/Code/host-uk/cli/cmd/docs/cmd_sync.go`

This is the main task. Add a `--target` flag (default `"php"`) and a new `runHugoSync` function that maps repos to Hugo's content tree.

**Add flag variable and registration:**

```go
var (
    docsSyncRegistryPath string
    docsSyncDryRun       bool
    docsSyncOutputDir    string
    docsSyncTarget       string
)

func init() {
    docsSyncCmd.Flags().StringVar(&docsSyncRegistryPath, "registry", "", i18n.T("common.flag.registry"))
    docsSyncCmd.Flags().BoolVar(&docsSyncDryRun, "dry-run", false, i18n.T("cmd.docs.sync.flag.dry_run"))
    docsSyncCmd.Flags().StringVar(&docsSyncOutputDir, "output", "", i18n.T("cmd.docs.sync.flag.output"))
    docsSyncCmd.Flags().StringVar(&docsSyncTarget, "target", "php", "Target format: php (default) or hugo")
}
```

**Update RunE to pass target:**
```go
RunE: func(cmd *cli.Command, args []string) error {
    return runDocsSync(docsSyncRegistryPath, docsSyncOutputDir, docsSyncDryRun, docsSyncTarget)
},
```

**Update `runDocsSync` signature and add target dispatch:**
```go
func runDocsSync(registryPath string, outputDir string, dryRun bool, target string) error {
    reg, basePath, err := loadRegistry(registryPath)
    if err != nil {
        return err
    }

    switch target {
    case "hugo":
        return runHugoSync(reg, basePath, outputDir, dryRun)
    default:
        return runPHPSync(reg, basePath, outputDir, dryRun)
    }
}
```

**Rename current sync body to `runPHPSync`** — extract lines 67-159 of current `runDocsSync` into `runPHPSync(reg, basePath, outputDir string, dryRun bool) error`. This is a pure extract, no logic changes.

**Add `hugoOutputName` mapping function:**
```go
// hugoOutputName maps repo name to Hugo content section and folder.
// Returns (section, folder) where section is the top-level content dir.
func hugoOutputName(repoName string) (string, string) {
    // CLI guides
    if repoName == "cli" {
        return "getting-started", ""
    }
    // Core CLI command docs
    if repoName == "core" {
        return "cli", ""
    }
    // Go packages
    if strings.HasPrefix(repoName, "go-") {
        return "go", repoName
    }
    // PHP packages
    if strings.HasPrefix(repoName, "core-") {
        return "php", strings.TrimPrefix(repoName, "core-")
    }
    return "go", repoName
}
```

**Add front matter injection helper:**
```go
// injectFrontMatter prepends Hugo front matter to markdown content if missing.
func injectFrontMatter(content []byte, title string, weight int) []byte {
    // Already has front matter
    if bytes.HasPrefix(bytes.TrimSpace(content), []byte("---")) {
        return content
    }
    fm := fmt.Sprintf("---\ntitle: %q\nweight: %d\n---\n\n", title, weight)
    return append([]byte(fm), content...)
}

// titleFromFilename derives a human-readable title from a filename.
func titleFromFilename(filename string) string {
    name := strings.TrimSuffix(filepath.Base(filename), ".md")
    name = strings.ReplaceAll(name, "-", " ")
    name = strings.ReplaceAll(name, "_", " ")
    // Title case
    words := strings.Fields(name)
    for i, w := range words {
        if len(w) > 0 {
            words[i] = strings.ToUpper(w[:1]) + w[1:]
        }
    }
    return strings.Join(words, " ")
}
```

**Add `runHugoSync` function:**
```go
func runHugoSync(reg *repos.Registry, basePath string, outputDir string, dryRun bool) error {
    if outputDir == "" {
        outputDir = filepath.Join(basePath, "docs-site", "content")
    }

    // Scan all repos
    var docsInfo []RepoDocInfo
    for _, repo := range reg.List() {
        if repo.Name == "core-template" || repo.Name == "core-claude" {
            continue
        }
        info := scanRepoDocs(repo)
        if info.HasDocs {
            docsInfo = append(docsInfo, info)
        }
    }

    if len(docsInfo) == 0 {
        cli.Text("No documentation found")
        return nil
    }

    cli.Print("\n  Hugo sync: %d repos with docs → %s\n\n", len(docsInfo), outputDir)

    // Show plan
    for _, info := range docsInfo {
        section, folder := hugoOutputName(info.Name)
        target := section
        if folder != "" {
            target = section + "/" + folder
        }
        fileCount := len(info.DocsFiles) + len(info.KBFiles)
        if info.Readme != "" {
            fileCount++
        }
        cli.Print("  %s → %s/ (%d files)\n", repoNameStyle.Render(info.Name), target, fileCount)
    }

    if dryRun {
        cli.Print("\n  Dry run — no files written\n")
        return nil
    }

    cli.Blank()
    if !confirm("Sync to Hugo content directory?") {
        cli.Text("Aborted")
        return nil
    }

    cli.Blank()
    var synced int
    for _, info := range docsInfo {
        section, folder := hugoOutputName(info.Name)

        // Build destination path
        destDir := filepath.Join(outputDir, section)
        if folder != "" {
            destDir = filepath.Join(destDir, folder)
        }

        // Copy docs/ files
        weight := 10
        docsDir := filepath.Join(info.Path, "docs")
        for _, f := range info.DocsFiles {
            src := filepath.Join(docsDir, f)
            dst := filepath.Join(destDir, f)
            if err := copyWithFrontMatter(src, dst, weight); err != nil {
                cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
                continue
            }
            weight += 10
        }

        // Copy README.md as _index.md (if not CLI/core which use their own index)
        if info.Readme != "" && folder != "" {
            dst := filepath.Join(destDir, "_index.md")
            if err := copyWithFrontMatter(info.Readme, dst, 1); err != nil {
                cli.Print("  %s README: %s\n", errorStyle.Render("✗"), err)
            }
        }

        // Copy KB/ files to kb/{suffix}/
        if len(info.KBFiles) > 0 {
            // Extract suffix: go-mlx → mlx, go-i18n → i18n
            suffix := strings.TrimPrefix(info.Name, "go-")
            kbDestDir := filepath.Join(outputDir, "kb", suffix)
            kbDir := filepath.Join(info.Path, "KB")
            kbWeight := 10
            for _, f := range info.KBFiles {
                src := filepath.Join(kbDir, f)
                dst := filepath.Join(kbDestDir, f)
                if err := copyWithFrontMatter(src, dst, kbWeight); err != nil {
                    cli.Print("  %s KB/%s: %s\n", errorStyle.Render("✗"), f, err)
                    continue
                }
                kbWeight += 10
            }
        }

        cli.Print("  %s %s\n", successStyle.Render("✓"), info.Name)
        synced++
    }

    cli.Print("\n  Synced %d repos to Hugo content\n", synced)
    return nil
}

// copyWithFrontMatter copies a markdown file, injecting front matter if missing.
func copyWithFrontMatter(src, dst string, weight int) error {
    if err := io.Local.EnsureDir(filepath.Dir(dst)); err != nil {
        return err
    }
    content, err := io.Local.Read(src)
    if err != nil {
        return err
    }
    title := titleFromFilename(src)
    result := injectFrontMatter([]byte(content), title, weight)
    return io.Local.Write(dst, string(result))
}
```

**Add imports** at top of file:
```go
import (
    "bytes"
    "fmt"
    "path/filepath"
    "strings"

    "forge.lthn.ai/core/go/pkg/cli"
    "forge.lthn.ai/core/go/pkg/i18n"
    "forge.lthn.ai/core/go/pkg/io"
    "forge.lthn.ai/core/go/pkg/repos"
)
```

**Verify:** `cd /Users/snider/Code/host-uk/cli && GOWORK=off go build ./cmd/docs/...`

**Commit:**
```bash
git add cmd/docs/cmd_sync.go
git commit -m "feat(docs): add --target hugo sync mode for core.help"
```

---

## Task 4: Test the full pipeline

**No code changes.** Run the pipeline end-to-end.

**Step 1:** Sync docs to Hugo:
```bash
cd /Users/snider/Code/host-uk
core docs sync --target hugo --dry-run
```
Verify all 39 repos appear with correct section mappings.

**Step 2:** Run actual sync:
```bash
core docs sync --target hugo
```

**Step 3:** Build and preview:
```bash
cd /Users/snider/Code/host-uk/docs-site
hugo server
```
Open `localhost:1313` and verify:
- Landing page renders with section links
- Getting Started section has CLI guides
- CLI Reference section has command docs
- Go Packages section has 18 packages with architecture/development/history
- PHP Packages section has PHP module docs
- Knowledge Base has MLX and i18n wiki pages
- Navigation works, search works

**Step 4:** Fix any issues found during preview.

**Commit docs-site content:**
```bash
cd /Users/snider/Code/host-uk/docs-site
git add content/
git commit -m "feat: sync initial content from 39 repos"
```

---

## Task 5: BunnyCDN deployment config

**Files:**
- Modify: `/Users/snider/Code/host-uk/docs-site/hugo.toml`

Add deployment target:

```toml
[deployment]
[[deployment.targets]]
name = "production"
URL = "s3://core-help?endpoint=storage.bunnycdn.com&region=auto"
```

Add a `Taskfile.yml` for convenience:

**Create:** `/Users/snider/Code/host-uk/docs-site/Taskfile.yml`
```yaml
version: '3'

tasks:
  dev:
    desc: Start Hugo dev server
    cmds:
      - hugo server --buildDrafts

  build:
    desc: Build static site
    cmds:
      - hugo --minify

  sync:
    desc: Sync docs from all repos
    dir: ..
    cmds:
      - core docs sync --target hugo

  deploy:
    desc: Build and deploy to BunnyCDN
    cmds:
      - task: sync
      - task: build
      - hugo deploy --target production

  clean:
    desc: Remove generated content (keeps _index.md files)
    cmds:
      - find content -name "*.md" ! -name "_index.md" -delete
```

**Verify:** `task dev` starts the site.

**Commit:**
```bash
git add hugo.toml Taskfile.yml
git commit -m "feat: add BunnyCDN deployment config and Taskfile"
```

---

## Dependency Sequencing

```
Task 1 (Hugo scaffold) — independent, do first
Task 2 (scan KB/) — independent, can parallel with Task 1
Task 3 (--target hugo) — depends on Task 2
Task 4 (test pipeline) — depends on Tasks 1 + 3
Task 5 (deploy config) — depends on Task 1
```

## Verification

After all tasks:
1. `core docs sync --target hugo` populates `docs-site/content/` from all repos
2. `cd docs-site && hugo server` renders the full site
3. Navigation has 6 sections: Getting Started, CLI, Go, MCP, PHP, KB
4. All existing markdown renders correctly with auto-injected front matter
5. `hugo build` produces `public/` with no errors
