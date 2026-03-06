# core/docs Help Engine — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Merge go-help into core/docs as `pkg/help`, replace Hugo with native static generator using go-html HLCRF layout, deploy to core.help via CLI.

**Architecture:** core/docs becomes a Go module (`forge.lthn.ai/core/docs`) with the help engine at `pkg/help/`. go-html provides the HLCRF layout (header/sidebar/content/footer). go-devops `core docs sync` gains a `gohelp` target that outputs to the static generator instead of Hugo.

**Tech Stack:** Go 1.26, go-html (HLCRF), goldmark (Markdown), yaml.v3 (frontmatter), testify (tests)

---

### Task 1: Initialise core/docs as a Go module

The docs repo is currently content-only (no go.mod). We need to make it a Go module.

**Files:**
- Create: `/Users/snider/Code/core/docs/go.mod`
- Create: `/Users/snider/Code/core/docs/pkg/help/.gitkeep`

**Step 1: Create go.mod**

```bash
cd /Users/snider/Code/core/docs
go mod init forge.lthn.ai/core/docs
```

**Step 2: Create pkg/help directory**

```bash
mkdir -p /Users/snider/Code/core/docs/pkg/help
```

**Step 3: Add to go.work**

Edit `~/Code/go.work` — add `./core/docs` to the use block.

**Step 4: Verify module resolves**

```bash
cd /Users/snider/Code
go work sync
```

Expected: no errors

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/docs
git add go.mod pkg/
git commit -m "feat: initialise Go module for help engine

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 2: Copy go-help source into pkg/help

Move all `.go` files and `templates/` from go-help into `pkg/help/`. Change the package name from `help` to `help` (stays the same — `package help`).

**Files:**
- Copy from: `/Users/snider/Code/core/go-help/*.go` → `/Users/snider/Code/core/docs/pkg/help/`
- Copy from: `/Users/snider/Code/core/go-help/templates/` → `/Users/snider/Code/core/docs/pkg/help/templates/`

**Step 1: Copy all source files**

```bash
cp /Users/snider/Code/core/go-help/*.go /Users/snider/Code/core/docs/pkg/help/
cp -r /Users/snider/Code/core/go-help/templates /Users/snider/Code/core/docs/pkg/help/templates
```

**Step 2: Add dependencies to go.mod**

```bash
cd /Users/snider/Code/core/docs
go get github.com/yuin/goldmark
go get gopkg.in/yaml.v3
go get github.com/stretchr/testify
```

**Step 3: Verify build**

```bash
cd /Users/snider/Code/core/docs
go build ./pkg/help/...
```

Expected: builds cleanly

**Step 4: Verify tests pass**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/... -v
```

Expected: all existing go-help tests pass (~18 tests)

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/docs
git add pkg/help/ go.mod go.sum
git commit -m "feat: import go-help library as pkg/help

All source, tests, and templates copied from forge.lthn.ai/core/go-help.
94% test coverage preserved.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 3: Add go-html dependency and create layout.go

Replace the `html/template` rendering with go-html HLCRF compositor. The layout produces the same dark-theme page structure but with semantic HTML.

**Files:**
- Create: `/Users/snider/Code/core/docs/pkg/help/layout.go`
- Test: `/Users/snider/Code/core/docs/pkg/help/layout_test.go`

**Step 1: Write the failing test**

Create `layout_test.go`:

```go
package help

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderLayout_Good_IndexPage(t *testing.T) {
	topics := []*Topic{
		{ID: "getting-started", Title: "Getting Started", Tags: []string{"guide"}},
		{ID: "config", Title: "Configuration", Tags: []string{"guide"}},
	}
	result := RenderIndexPage(topics)

	assert.Contains(t, result, `role="banner"`)
	assert.Contains(t, result, `role="complementary"`)
	assert.Contains(t, result, `role="main"`)
	assert.Contains(t, result, `role="contentinfo"`)
	assert.Contains(t, result, "core.help")
	assert.Contains(t, result, "Getting Started")
	assert.Contains(t, result, "Configuration")
}

func TestRenderLayout_Good_TopicPage(t *testing.T) {
	topic := &Topic{
		ID:      "rate-limiting",
		Title:   "Rate Limiting",
		Content: "## Overview\n\nRate limiting controls...\n\n## Configuration\n\nSet the limit...",
		Sections: []Section{
			{ID: "overview", Title: "Overview", Level: 2},
			{ID: "configuration", Title: "Configuration", Level: 2},
		},
		Tags: []string{"api"},
	}
	sidebar := []*Topic{
		{ID: "rate-limiting", Title: "Rate Limiting", Tags: []string{"api"}},
		{ID: "auth", Title: "Authentication", Tags: []string{"api"}},
	}
	result := RenderTopicPage(topic, sidebar)

	assert.Contains(t, result, `role="banner"`)
	assert.Contains(t, result, `role="main"`)
	assert.Contains(t, result, "Rate Limiting")
	// Section anchors in ToC
	assert.Contains(t, result, `href="#overview"`)
	assert.Contains(t, result, `href="#configuration"`)
}

func TestRenderLayout_Good_SearchPage(t *testing.T) {
	results := []*SearchResult{
		{Topic: &Topic{ID: "rate-limiting", Title: "Rate Limiting"}, Score: 10.0, Snippet: "Rate limiting controls..."},
	}
	result := RenderSearchPage("rate limit", results)

	assert.Contains(t, result, `role="main"`)
	assert.Contains(t, result, "rate limit")
	assert.Contains(t, result, "Rate Limiting")
}

func TestRenderLayout_Good_HasDoctype(t *testing.T) {
	topics := []*Topic{{ID: "test", Title: "Test"}}
	result := RenderIndexPage(topics)

	assert.True(t, strings.HasPrefix(result, "<!DOCTYPE html>"))
}

func TestRenderLayout_Good_404Page(t *testing.T) {
	result := Render404Page()

	assert.Contains(t, result, `role="main"`)
	assert.Contains(t, result, "Not Found")
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestRenderLayout -v
```

Expected: FAIL — `RenderIndexPage` undefined

**Step 3: Add go-html dependency**

```bash
cd /Users/snider/Code/core/docs
go get forge.lthn.ai/core/go-html
```

**Step 4: Implement layout.go**

Create `layout.go` — uses go-html HLCRF to build full page wrappers:

```go
// SPDX-Licence-Identifier: EUPL-1.2
package help

import (
	"fmt"
	"strings"

	html "forge.lthn.ai/core/go-html"
)

// pageCSS is the inline dark-theme stylesheet for all pages.
const pageCSS = `
*{margin:0;padding:0;box-sizing:border-box}
:root{--bg:#0d1117;--bg-card:#161b22;--fg:#c9d1d9;--fg-muted:#8b949e;--accent:#58a6ff;--border:#30363d;--radius:6px}
body{background:var(--bg);color:var(--fg);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Helvetica,Arial,sans-serif;line-height:1.6}
a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline}
header[role=banner]{background:var(--bg-card);border-bottom:1px solid var(--border);padding:0.75rem 1.5rem;display:flex;align-items:center;justify-content:space-between}
header .brand{font-weight:600;font-size:1.1rem;color:var(--fg)}
header .search-form{display:flex;gap:0.5rem}
header input[type=search]{background:var(--bg);border:1px solid var(--border);border-radius:var(--radius);color:var(--fg);padding:0.4rem 0.75rem;width:220px}
aside[role=complementary]{width:250px;min-width:200px;padding:1rem;border-right:1px solid var(--border);overflow-y:auto;background:var(--bg-card)}
aside .nav-group{margin-bottom:1rem}
aside .nav-group-title{font-size:0.75rem;text-transform:uppercase;color:var(--fg-muted);margin-bottom:0.25rem;font-weight:600}
aside .nav-link{display:block;padding:0.2rem 0.5rem;border-radius:var(--radius);font-size:0.9rem;color:var(--fg)}
aside .nav-link:hover{background:var(--border);text-decoration:none}
main[role=main]{flex:1;padding:2rem;max-width:900px;overflow-y:auto}
main h1,main h2,main h3,main h4,main h5,main h6{margin:1.5rem 0 0.75rem;color:var(--fg)}
main h1{font-size:2rem;border-bottom:1px solid var(--border);padding-bottom:0.3rem}
main h2{font-size:1.5rem}
main code{background:var(--bg-card);padding:0.2em 0.4em;border-radius:3px;font-size:0.85em}
main pre{background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius);padding:1rem;overflow-x:auto;margin:1rem 0}
main pre code{background:none;padding:0}
main table{border-collapse:collapse;width:100%;margin:1rem 0}
main th,main td{border:1px solid var(--border);padding:0.5rem 0.75rem;text-align:left}
main th{background:var(--bg-card)}
.card{background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius);padding:1rem;margin:0.75rem 0}
.card h3{margin:0 0 0.5rem}
.tag{display:inline-block;background:var(--border);color:var(--fg-muted);padding:0.1rem 0.5rem;border-radius:3px;font-size:0.8rem;margin-right:0.25rem}
.toc{margin:1rem 0;padding:0.75rem 1rem;background:var(--bg-card);border:1px solid var(--border);border-radius:var(--radius)}
.toc-title{font-weight:600;margin-bottom:0.5rem;font-size:0.9rem}
.toc a{display:block;padding:0.15rem 0;font-size:0.85rem}
.wrapper{display:flex;min-height:100vh;flex-direction:column}
.body-wrap{display:flex;flex:1}
footer[role=contentinfo]{background:var(--bg-card);border-top:1px solid var(--border);padding:0.75rem 1.5rem;text-align:center;color:var(--fg-muted);font-size:0.85rem}
@media(max-width:768px){aside[role=complementary]{display:none}.body-wrap{flex-direction:column}}
`

// wrapPage wraps HLCRF body content in a full HTML document.
func wrapPage(title string, body string) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>")
	b.WriteString(escapeHTMLText(title))
	b.WriteString(" — core.help</title>\n")
	b.WriteString("<style>")
	b.WriteString(pageCSS)
	b.WriteString("</style>\n")
	b.WriteString("</head>\n<body>\n<div class=\"wrapper\">\n")
	b.WriteString(body)
	b.WriteString("\n</div>\n</body>\n</html>")
	return b.String()
}

// escapeHTMLText escapes text for safe HTML insertion.
func escapeHTMLText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// renderHeader builds the banner header with branding and search.
func renderHeader() html.Node {
	return html.El("div",
		html.Attr(html.El("span", html.Raw("core.help")), "class", "brand"),
		html.Raw(`<form class="search-form" action="/search" method="get"><input type="search" name="q" placeholder="Search docs..." aria-label="Search"></form>`),
	)
}

// renderSidebar builds the left aside with topic groups.
func renderSidebar(topics []*Topic) html.Node {
	groups := groupTopicsByTag(topics)
	var children []html.Node
	for _, g := range groups {
		var links []html.Node
		links = append(links, html.Attr(html.El("div", html.Raw(escapeHTMLText(g.Tag))), "class", "nav-group-title"))
		for _, t := range g.Topics {
			link := html.Attr(
				html.El("a", html.Raw(escapeHTMLText(t.Title))),
				"href", "/topics/"+t.ID,
			)
			link = html.Attr(link, "class", "nav-link")
			links = append(links, link)
		}
		group := html.Attr(html.El("div", links...), "class", "nav-group")
		children = append(children, group)
	}
	return html.El("nav", children...)
}

// renderFooter builds the footer content.
func renderFooter() html.Node {
	return html.Raw("EUPL-1.2 — <a href=\"https://forge.lthn.ai/core/docs\">Source</a>")
}

// buildPage constructs a full HLCRF page.
func buildPage(title string, sidebar []*Topic, content html.Node) string {
	layout := html.NewLayout("HLCF")
	layout.H(renderHeader())
	if len(sidebar) > 0 {
		layout = html.NewLayout("HLCRF")
		layout.H(renderHeader())
		layout.L(renderSidebar(sidebar))
	}
	layout.C(content)
	layout.F(renderFooter())

	ctx := html.NewContext()
	body := layout.Render(ctx)

	// Wrap the aside+main in body-wrap div
	// The HLCRF renders H, then L+C+R, then F
	// We need the body-wrap around L+C for flex layout
	// Since go-html renders sequentially, we post-process
	body = strings.Replace(body, `<aside role="complementary"`, `<div class="body-wrap"><aside role="complementary"`, 1)
	body = strings.Replace(body, `</main>`, `</main></div>`, 1)

	return wrapPage(title, body)
}

// RenderIndexPage renders the topic index page.
func RenderIndexPage(topics []*Topic) string {
	groups := groupTopicsByTag(topics)
	var children []html.Node

	children = append(children, html.El("h1", html.Raw("Documentation")))
	children = append(children, html.El("p",
		html.Raw(fmt.Sprintf("%d %s", len(topics), pluralise(len(topics), "topic", "topics"))),
	))

	for _, g := range groups {
		children = append(children, html.El("h2", html.Raw(escapeHTMLText(g.Tag))))
		for _, t := range g.Topics {
			card := html.Attr(html.El("div",
				html.El("h3", html.Attr(html.El("a", html.Raw(escapeHTMLText(t.Title))), "href", "/topics/"+t.ID)),
				html.El("p", html.Raw(escapeHTMLText(truncateContent(t.Content, 150)))),
			), "class", "card")
			children = append(children, card)
		}
	}

	return buildPage("Documentation", topics, html.El("div", children...))
}

// RenderTopicPage renders a single topic page.
func RenderTopicPage(topic *Topic, sidebar []*Topic) string {
	rendered, _ := RenderMarkdown(topic.Content)

	var children []html.Node
	children = append(children, html.El("h1", html.Raw(escapeHTMLText(topic.Title))))

	// Table of contents
	if len(topic.Sections) > 0 {
		var tocLinks []html.Node
		tocLinks = append(tocLinks, html.Attr(html.El("div", html.Raw("Contents")), "class", "toc-title"))
		for _, s := range topic.Sections {
			tocLinks = append(tocLinks,
				html.Attr(html.El("a", html.Raw(escapeHTMLText(s.Title))), "href", "#"+s.ID),
			)
		}
		children = append(children, html.Attr(html.El("div", tocLinks...), "class", "toc"))
	}

	children = append(children, html.Raw(rendered))

	return buildPage(topic.Title, sidebar, html.El("article", children...))
}

// RenderSearchPage renders search results.
func RenderSearchPage(query string, results []*SearchResult) string {
	var children []html.Node
	children = append(children, html.El("h1", html.Raw("Search Results")))

	if query != "" {
		summary := fmt.Sprintf("Found %d %s for \u201c%s\u201d",
			len(results),
			pluralise(len(results), "result", "results"),
			escapeHTMLText(query))
		children = append(children, html.El("p", html.Raw(summary)))
	}

	for _, r := range results {
		card := html.Attr(html.El("div",
			html.El("h3", html.Attr(html.El("a", html.Raw(escapeHTMLText(r.Topic.Title))), "href", "/topics/"+r.Topic.ID)),
			html.El("p", html.Raw(escapeHTMLText(r.Snippet))),
		), "class", "card")
		children = append(children, card)
	}

	return buildPage("Search", nil, html.El("div", children...))
}

// Render404Page renders the not-found page.
func Render404Page() string {
	content := html.El("div",
		html.El("h1", html.Raw("Not Found")),
		html.El("p", html.Raw("The page you're looking for doesn't exist.")),
		html.El("p", html.Attr(html.El("a", html.Raw("Browse all topics")), "href", "/")),
	)
	return buildPage("Not Found", nil, content)
}

// truncateContent strips headings and truncates to n runes.
func truncateContent(s string, n int) string {
	var clean []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		clean = append(clean, trimmed)
	}
	text := strings.Join(clean, " ")
	runes := []rune(text)
	if len(runes) <= n {
		return text
	}
	return string(runes[:n]) + "..."
}

// pluralise returns singular or plural form.
func pluralise(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
```

**Step 5: Run tests**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestRenderLayout -v
```

Expected: all 5 layout tests pass

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/docs
git add pkg/help/layout.go pkg/help/layout_test.go go.mod go.sum
git commit -m "feat(help): add go-html HLCRF layout

Replaces html/template rendering with go-html compositor.
Dark theme, semantic HTML, ARIA roles, section anchors.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 4: Update server.go to use layout functions

Replace the `html/template`-based rendering in server.go with the new layout functions.

**Files:**
- Modify: `/Users/snider/Code/core/docs/pkg/help/server.go`
- Modify: `/Users/snider/Code/core/docs/pkg/help/server_test.go`

**Step 1: Update server handlers**

Replace the three HTML handlers (`handleIndex`, `handleTopic`, `handleSearch`) to call `RenderIndexPage`, `RenderTopicPage`, `RenderSearchPage`, and `Render404Page` instead of `renderPage`.

```go
func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	setSecurityHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	topics := s.catalog.List()
	_, _ = w.Write([]byte(RenderIndexPage(topics)))
}

func (s *Server) handleTopic(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	id := r.PathValue("id")
	topic, err := s.catalog.Get(id)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(Render404Page()))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(RenderTopicPage(topic, s.catalog.List())))
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query parameter 'q'", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	results := s.catalog.Search(query)
	_, _ = w.Write([]byte(RenderSearchPage(query, results)))
}
```

**Step 2: Run existing server tests**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestServer -v
```

Expected: all server tests pass with new layout output

**Step 3: Commit**

```bash
cd /Users/snider/Code/core/docs
git add pkg/help/server.go
git commit -m "refactor(help): switch server handlers to HLCRF layout

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 5: Update generate.go to use layout functions

Replace the template-based static site generator with the layout functions.

**Files:**
- Modify: `/Users/snider/Code/core/docs/pkg/help/generate.go`
- Test: `/Users/snider/Code/core/docs/pkg/help/generate_test.go`

**Step 1: Update Generate function**

Replace `writeStaticPage` calls with `RenderIndexPage`, `RenderTopicPage`, etc. Keep `writeSearchIndex` (JSON) and `clientSearchScript` (JS) as-is.

The `writeStaticPage` helper becomes:

```go
func writeStaticFile(dir, filename, content string) error {
	path := filepath.Join(dir, filename)
	return os.WriteFile(path, []byte(content), 0o644)
}
```

And `Generate` becomes:

```go
func Generate(catalog *Catalog, outputDir string) error {
	topics := catalog.List()
	topicsDir := filepath.Join(outputDir, "topics")
	if err := os.MkdirAll(topicsDir, 0o755); err != nil {
		return err
	}

	// 1. index.html
	if err := writeStaticFile(outputDir, "index.html", RenderIndexPage(topics)); err != nil {
		return err
	}

	// 2. topics/{id}.html
	for _, t := range topics {
		page := RenderTopicPage(t, topics)
		if err := writeStaticFile(topicsDir, t.ID+".html", page); err != nil {
			return err
		}
	}

	// 3. search.html (with client-side JS)
	searchPage := RenderSearchPage("", nil) + clientSearchScript
	if err := writeStaticFile(outputDir, "search.html", searchPage); err != nil {
		return err
	}

	// 4. search-index.json
	if err := writeSearchIndex(outputDir, topics); err != nil {
		return err
	}

	// 5. 404.html
	if err := writeStaticFile(outputDir, "404.html", Render404Page()); err != nil {
		return err
	}

	return nil
}
```

**Step 2: Run generate tests**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestGenerate -v
```

Expected: all generate tests pass

**Step 3: Commit**

```bash
cd /Users/snider/Code/core/docs
git add pkg/help/generate.go
git commit -m "refactor(help): switch static generator to HLCRF layout

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 6: Remove old template dependency

Now that layout.go handles all rendering, remove the `html/template` based code.

**Files:**
- Delete: `/Users/snider/Code/core/docs/pkg/help/templates/` (all 5 HTML files)
- Modify: `/Users/snider/Code/core/docs/pkg/help/templates.go` — remove `//go:embed`, `parseTemplates`, `renderPage`, template data types
- Modify: `/Users/snider/Code/core/docs/pkg/help/templates_test.go` — remove tests for old template functions

**Step 1: Delete template HTML files**

```bash
rm -r /Users/snider/Code/core/docs/pkg/help/templates/
```

**Step 2: Slim down templates.go**

Keep only `groupTopicsByTag` (used by layout.go). Remove: `templateFS`, `templateFuncs`, `parseTemplates`, `renderPage`, `indexData`, `topicData`, `searchData`, `topicGroup` definition stays.

The `truncate`, `pluralise`, `multiply`, `sub` template funcs are replaced by `truncateContent` and `pluralise` in layout.go. Remove duplicates from templates.go.

**Step 3: Run all tests**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/... -v
```

Expected: all tests pass, no import of `html/template` or `embed` remains in templates.go

**Step 4: Commit**

```bash
cd /Users/snider/Code/core/docs
git add -A pkg/help/
git commit -m "refactor(help): remove html/template dependency

All rendering now uses go-html HLCRF layout.
Templates directory and template parsing code removed.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 7: Add `core docs sync --target gohelp` to go-devops

Add a third sync target alongside `php` and `hugo` that outputs to go-help's content format.

**Files:**
- Modify: `/Users/snider/Code/core/go-devops/cmd/docs/cmd_sync.go` — add `gohelp` case
- Test: `/Users/snider/Code/core/go-devops/cmd/docs/cmd_sync_test.go` (if exists, add test)

**Step 1: Add gohelp target**

In `runDocsSync`, add:

```go
case "gohelp":
    return runGoHelpSync(reg, basePath, outputDir, dryRun)
```

**Step 2: Implement runGoHelpSync**

Similar to `runHugoSync` but simpler — copies `docs/` folders into flat `content/{package-name}/` structure without Hugo frontmatter injection:

```go
func goHelpOutputName(repoName string) string {
	if repoName == "core" {
		return "cli"
	}
	if strings.HasPrefix(repoName, "core-") {
		return strings.TrimPrefix(repoName, "core-")
	}
	return repoName
}

func runGoHelpSync(reg *repos.Registry, basePath string, outputDir string, dryRun bool) error {
	if outputDir == "" {
		outputDir = filepath.Join(basePath, "docs", "content")
	}

	var docsInfo []RepoDocInfo
	for _, repo := range reg.List() {
		if repo.Name == "core-template" {
			continue
		}
		info := scanRepoDocs(repo)
		if info.HasDocs && len(info.DocsFiles) > 0 {
			docsInfo = append(docsInfo, info)
		}
	}

	if len(docsInfo) == 0 {
		cli.Text("No documentation found")
		return nil
	}

	cli.Print("\n  go-help sync: %d repos with docs → %s\n\n", len(docsInfo), outputDir)

	for _, info := range docsInfo {
		outName := goHelpOutputName(info.Name)
		cli.Print("  %s → %s/ (%d files)\n", repoNameStyle.Render(info.Name), outName, len(info.DocsFiles))
	}

	if dryRun {
		cli.Print("\n  Dry run — no files written\n")
		return nil
	}

	cli.Blank()
	if !confirm("Sync to go-help content directory?") {
		cli.Text("Aborted")
		return nil
	}

	cli.Blank()
	var synced int
	for _, info := range docsInfo {
		outName := goHelpOutputName(info.Name)
		destDir := filepath.Join(outputDir, outName)

		_ = io.Local.DeleteAll(destDir)
		if err := io.Local.EnsureDir(destDir); err != nil {
			cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), info.Name, err)
			continue
		}

		docsDir := filepath.Join(info.Path, "docs")
		for _, f := range info.DocsFiles {
			src := filepath.Join(docsDir, f)
			dst := filepath.Join(destDir, f)
			if err := io.Local.EnsureDir(filepath.Dir(dst)); err != nil {
				continue
			}
			if err := io.Copy(io.Local, src, io.Local, dst); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
			}
		}

		cli.Print("  %s %s\n", successStyle.Render("✓"), info.Name)
		synced++
	}

	cli.Print("\n  Synced %d repos to go-help content\n", synced)
	return nil
}
```

**Step 3: Run build**

```bash
cd /Users/snider/Code/core/go-devops
go build ./...
```

Expected: builds cleanly

**Step 4: Commit**

```bash
cd /Users/snider/Code/core/go-devops
git add cmd/docs/cmd_sync.go
git commit -m "feat(docs): add gohelp sync target

core docs sync --target gohelp collects repo docs/ folders
into flat content/{package}/ structure for go-help static gen.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 8: Add LoadContentDir to catalog

The catalog needs to load a directory tree of Markdown files (the collected content).

**Files:**
- Modify: `/Users/snider/Code/core/docs/pkg/help/catalog.go`
- Test: `/Users/snider/Code/core/docs/pkg/help/catalog_test.go`

**Step 1: Write the failing test**

Add to `catalog_test.go`:

```go
func TestCatalog_LoadContentDir_Good(t *testing.T) {
	dir := t.TempDir()
	// Create test content
	os.MkdirAll(filepath.Join(dir, "cli"), 0o755)
	os.WriteFile(filepath.Join(dir, "cli", "dev-work.md"), []byte(`---
title: Dev Work
tags: [cli, dev]
---

## Usage

core dev work syncs your workspace.
`), 0o644)

	os.WriteFile(filepath.Join(dir, "cli", "setup.md"), []byte(`---
title: Setup
tags: [cli]
---

## Installation

Run core setup to get started.
`), 0o644)

	catalog, err := LoadContentDir(dir)
	require.NoError(t, err)

	topics := catalog.List()
	assert.Len(t, topics, 2)

	devWork, err := catalog.Get("dev-work")
	require.NoError(t, err)
	assert.Equal(t, "Dev Work", devWork.Title)
	assert.Contains(t, devWork.Tags, "cli")

	// Search should work
	results := catalog.Search("workspace")
	assert.NotEmpty(t, results)
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestCatalog_LoadContentDir -v
```

Expected: FAIL — `LoadContentDir` undefined

**Step 3: Implement LoadContentDir**

Add to `catalog.go`:

```go
// LoadContentDir recursively loads all .md files from a directory into a Catalog.
func LoadContentDir(dir string) (*Catalog, error) {
	c := &Catalog{
		topics: make(map[string]*Topic),
		index:  newSearchIndex(),
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		topic, err := ParseTopic(path, content)
		if err != nil {
			return err
		}
		c.Add(topic)
		return nil
	})

	return c, err
}
```

Add imports: `"io/fs"`, `"os"`, `"path/filepath"`, `"strings"`.

**Step 4: Run test**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestCatalog_LoadContentDir -v
```

Expected: PASS

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/docs
git add pkg/help/catalog.go pkg/help/catalog_test.go
git commit -m "feat(help): add LoadContentDir for directory-based catalog loading

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 9: Integration test — full pipeline

**Files:**
- Create: `/Users/snider/Code/core/docs/pkg/help/integration_test.go`

**Step 1: Write integration test**

```go
package help

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Good_FullPipeline(t *testing.T) {
	// 1. Create content directory
	contentDir := t.TempDir()
	os.MkdirAll(filepath.Join(contentDir, "cli"), 0o755)
	os.MkdirAll(filepath.Join(contentDir, "go"), 0o755)

	os.WriteFile(filepath.Join(contentDir, "cli", "dev-work.md"), []byte(`---
title: Dev Work
tags: [cli]
order: 1
---

## Usage

core dev work syncs your workspace.

## Flags

--status  Show status only
`), 0o644)

	os.WriteFile(filepath.Join(contentDir, "go", "go-scm.md"), []byte(`---
title: Go SCM
tags: [go, library]
order: 2
---

## Overview

Registry and git operations for the workspace.
`), 0o644)

	// 2. Load into catalog
	catalog, err := LoadContentDir(contentDir)
	require.NoError(t, err)
	assert.Equal(t, 2, len(catalog.List()))

	// 3. Generate static site
	outputDir := t.TempDir()
	err = Generate(catalog, outputDir)
	require.NoError(t, err)

	// 4. Verify file structure
	assert.FileExists(t, filepath.Join(outputDir, "index.html"))
	assert.FileExists(t, filepath.Join(outputDir, "search.html"))
	assert.FileExists(t, filepath.Join(outputDir, "search-index.json"))
	assert.FileExists(t, filepath.Join(outputDir, "404.html"))
	assert.FileExists(t, filepath.Join(outputDir, "topics", "dev-work.html"))
	assert.FileExists(t, filepath.Join(outputDir, "topics", "go-scm.html"))

	// 5. Verify search index
	indexData, err := os.ReadFile(filepath.Join(outputDir, "search-index.json"))
	require.NoError(t, err)
	var entries []searchIndexEntry
	err = json.Unmarshal(indexData, &entries)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	// 6. Verify HTML content
	indexHTML, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	require.NoError(t, err)
	assert.Contains(t, string(indexHTML), "Dev Work")
	assert.Contains(t, string(indexHTML), "Go SCM")
	assert.Contains(t, string(indexHTML), `role="banner"`)

	// 7. Verify topic page has section anchors
	topicHTML, err := os.ReadFile(filepath.Join(outputDir, "topics", "dev-work.html"))
	require.NoError(t, err)
	assert.Contains(t, string(topicHTML), `href="#usage"`)
	assert.Contains(t, string(topicHTML), `href="#flags"`)

	// 8. Search works
	results := catalog.Search("workspace")
	assert.NotEmpty(t, results)
	assert.Equal(t, "dev-work", results[0].Topic.ID)
}
```

**Step 2: Run integration test**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/ -run TestIntegration -v
```

Expected: PASS

**Step 3: Commit**

```bash
cd /Users/snider/Code/core/docs
git add pkg/help/integration_test.go
git commit -m "test(help): add full pipeline integration test

Verifies: content loading → catalog → static generation → file
structure → search index → HTML output with HLCRF layout.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 10: Tag, push, archive old repos

**Step 1: Run all tests one final time**

```bash
cd /Users/snider/Code/core/docs
go test ./pkg/help/... -v -count=1
```

Expected: all tests pass

**Step 2: Push core/docs**

```bash
cd /Users/snider/Code/core/docs
git push origin main
```

**Step 3: Tag**

```bash
cd /Users/snider/Code/core/docs
git tag v0.1.0
git push origin v0.1.0
```

**Step 4: Push go-devops changes**

```bash
cd /Users/snider/Code/core/go-devops
git push origin main
```

**Step 5: Archive go-help and docs-site on forge**

Archive via Forge UI or API:
- `forge.lthn.ai/core/go-help` → Archive
- `forge.lthn.ai/core/docs-site` → Archive

---

## Key References

| File | Role |
|------|------|
| `/Users/snider/Code/core/go-help/*.go` | Source to copy into pkg/help |
| `/Users/snider/Code/core/go-html/node.go` | El, Raw, Attr — DOM building |
| `/Users/snider/Code/core/go-html/layout.go` | HLCRF compositor |
| `/Users/snider/Code/core/go-html/context.go` | Render context |
| `/Users/snider/Code/core/go-devops/cmd/docs/cmd_sync.go` | Existing docs sync (php + hugo targets) |
| `/Users/snider/Code/core/go-devops/cmd/docs/cmd_scan.go` | RepoDocInfo, scanRepoDocs |
| `/Users/snider/Code/core/docs/content/` | Existing aggregated docs |
| `docs/plans/2026-03-06-docs-help-engine-design.md` | Approved design |
