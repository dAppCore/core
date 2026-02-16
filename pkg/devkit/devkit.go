// Package devkit provides a developer toolkit for common automation commands.
// Designed by Gemini 3 Pro (Hypnos) + Claude Opus (Charon), signed LEK-1 | lthn.ai | EUPL-1.2
package devkit

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// --- Code Quality ---

// Finding represents a single issue found by a linting tool.
type Finding struct {
	File    string
	Line    int
	Message string
	Tool    string
}

// CoverageReport holds the test coverage percentage for a package.
type CoverageReport struct {
	Package    string
	Percentage float64
}

// RaceCondition represents a data race detected by the Go race detector.
type RaceCondition struct {
	File string
	Line int
	Desc string
}

// TODO represents a tracked code comment like TODO, FIXME, or HACK.
type TODO struct {
	File    string
	Line    int
	Type    string
	Message string
}

// --- Security ---

// Vulnerability represents a dependency vulnerability.
type Vulnerability struct {
	ID          string
	Package     string
	Version     string
	Description string
}

// SecretLeak represents a potential secret found in the codebase.
type SecretLeak struct {
	File   string
	Line   int
	RuleID string
	Match  string
}

// PermIssue represents a file permission issue.
type PermIssue struct {
	File       string
	Permission string
	Issue      string
}

// --- Git Operations ---

// DiffSummary provides a summary of changes.
type DiffSummary struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

// Commit represents a single git commit.
type Commit struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
}

// --- Build & Dependencies ---

// BuildResult holds the outcome of a single build target.
type BuildResult struct {
	Target string
	Path   string
	Error  error
}

// Graph represents a dependency graph.
type Graph struct {
	Nodes []string
	Edges map[string][]string
}

// --- Metrics ---

// ComplexFunc represents a function with its cyclomatic complexity score.
type ComplexFunc struct {
	Package  string
	FuncName string
	File     string
	Line     int
	Score    int
}

// Toolkit wraps common dev automation commands into structured Go APIs.
type Toolkit struct {
	Dir string // Working directory for commands
}

// New creates a Toolkit rooted at the given directory.
func New(dir string) *Toolkit {
	return &Toolkit{Dir: dir}
}

// Run executes a command and captures stdout, stderr, and exit code.
func (t *Toolkit) Run(name string, args ...string) (stdout, stderr string, exitCode int, err error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = t.Dir
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return
}

// FindTODOs greps for TODO/FIXME/HACK comments within a directory.
func (t *Toolkit) FindTODOs(dir string) ([]TODO, error) {
	pattern := `\b(TODO|FIXME|HACK)\b(\(.*\))?:`
	stdout, stderr, exitCode, err := t.Run("git", "grep", "--line-number", "-E", pattern, "--", dir)

	if exitCode == 1 && stdout == "" {
		return nil, nil
	}
	if err != nil && exitCode != 1 {
		return nil, fmt.Errorf("git grep failed (exit %d): %s\n%s", exitCode, err, stderr)
	}

	var todos []TODO
	re := regexp.MustCompile(pattern)

	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		lineNum, _ := strconv.Atoi(parts[1])
		match := re.FindStringSubmatch(parts[2])
		todoType := ""
		if len(match) > 1 {
			todoType = match[1]
		}
		msg := strings.TrimSpace(re.Split(parts[2], 2)[1])

		todos = append(todos, TODO{
			File:    parts[0],
			Line:    lineNum,
			Type:    todoType,
			Message: msg,
		})
	}
	return todos, nil
}

// AuditDeps runs govulncheck to find dependency vulnerabilities.
func (t *Toolkit) AuditDeps() ([]Vulnerability, error) {
	stdout, stderr, exitCode, err := t.Run("govulncheck", "./...")
	if err != nil && exitCode != 0 && !strings.Contains(stdout, "Vulnerability") {
		return nil, fmt.Errorf("govulncheck failed (exit %d): %s\n%s", exitCode, err, stderr)
	}

	var vulns []Vulnerability
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	var cur Vulnerability
	inBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Vulnerability #") {
			if cur.ID != "" {
				vulns = append(vulns, cur)
			}
			fields := strings.Fields(line)
			cur = Vulnerability{}
			if len(fields) > 1 {
				cur.ID = fields[1]
			}
			inBlock = true
		} else if inBlock {
			switch {
			case strings.Contains(line, "Package:"):
				cur.Package = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			case strings.Contains(line, "Found in version:"):
				cur.Version = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			case line == "":
				if cur.ID != "" {
					vulns = append(vulns, cur)
					cur = Vulnerability{}
				}
				inBlock = false
			default:
				if !strings.HasPrefix(line, "  ") && cur.Description == "" {
					cur.Description = strings.TrimSpace(line)
				}
			}
		}
	}
	if cur.ID != "" {
		vulns = append(vulns, cur)
	}
	return vulns, nil
}

// DiffStat returns a summary of uncommitted changes.
func (t *Toolkit) DiffStat() (DiffSummary, error) {
	stdout, stderr, exitCode, err := t.Run("git", "diff", "--stat")
	if err != nil && exitCode != 0 {
		return DiffSummary{}, fmt.Errorf("git diff failed (exit %d): %s\n%s", exitCode, err, stderr)
	}

	var s DiffSummary
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return s, nil
	}

	last := lines[len(lines)-1]
	for _, part := range strings.Split(last, ",") {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.Atoi(fields[0])
		switch {
		case strings.Contains(part, "file"):
			s.FilesChanged = val
		case strings.Contains(part, "insertion"):
			s.Insertions = val
		case strings.Contains(part, "deletion"):
			s.Deletions = val
		}
	}
	return s, nil
}

// UncommittedFiles returns paths of files with uncommitted changes.
func (t *Toolkit) UncommittedFiles() ([]string, error) {
	stdout, stderr, exitCode, err := t.Run("git", "status", "--porcelain")
	if err != nil && exitCode != 0 {
		return nil, fmt.Errorf("git status failed: %s\n%s", err, stderr)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if len(line) > 3 {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	return files, nil
}

// Lint runs go vet on the given package pattern.
func (t *Toolkit) Lint(pkg string) ([]Finding, error) {
	_, stderr, exitCode, err := t.Run("go", "vet", pkg)
	if exitCode == 0 {
		return nil, nil
	}
	if err != nil && exitCode != 2 {
		return nil, fmt.Errorf("go vet failed: %w", err)
	}

	var findings []Finding
	for _, line := range strings.Split(strings.TrimSpace(stderr), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}
		lineNum, _ := strconv.Atoi(parts[1])
		findings = append(findings, Finding{
			File:    parts[0],
			Line:    lineNum,
			Message: strings.TrimSpace(parts[3]),
			Tool:    "go vet",
		})
	}
	return findings, nil
}

// ScanSecrets runs gitleaks to find potential secret leaks.
func (t *Toolkit) ScanSecrets(dir string) ([]SecretLeak, error) {
	stdout, _, exitCode, err := t.Run("gitleaks", "detect", "--source", dir, "--report-format", "csv", "--no-git")
	if exitCode == 0 {
		return nil, nil
	}
	if err != nil && exitCode != 1 {
		return nil, fmt.Errorf("gitleaks failed: %w", err)
	}

	var leaks []SecretLeak
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" || strings.HasPrefix(line, "RuleID") {
			continue
		}
		parts := strings.SplitN(line, ",", 4)
		if len(parts) < 4 {
			continue
		}
		lineNum, _ := strconv.Atoi(parts[2])
		leaks = append(leaks, SecretLeak{
			RuleID: parts[0],
			File:   parts[1],
			Line:   lineNum,
			Match:  parts[3],
		})
	}
	return leaks, nil
}

// ModTidy runs go mod tidy.
func (t *Toolkit) ModTidy() error {
	_, stderr, exitCode, err := t.Run("go", "mod", "tidy")
	if err != nil && exitCode != 0 {
		return fmt.Errorf("go mod tidy failed: %s", stderr)
	}
	return nil
}

// Build compiles the given targets.
func (t *Toolkit) Build(targets ...string) ([]BuildResult, error) {
	var results []BuildResult
	for _, target := range targets {
		_, stderr, _, err := t.Run("go", "build", "-o", "/dev/null", target)
		r := BuildResult{Target: target}
		if err != nil {
			r.Error = fmt.Errorf("%s", strings.TrimSpace(stderr))
		}
		results = append(results, r)
	}
	return results, nil
}

// TestCount returns the number of test functions in a package.
func (t *Toolkit) TestCount(pkg string) (int, error) {
	stdout, stderr, exitCode, err := t.Run("go", "test", "-list", ".*", pkg)
	if err != nil && exitCode != 0 {
		return 0, fmt.Errorf("go test -list failed: %s\n%s", err, stderr)
	}
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if strings.HasPrefix(line, "Test") || strings.HasPrefix(line, "Benchmark") {
			count++
		}
	}
	return count, nil
}

// Coverage runs go test -cover and parses per-package coverage percentages.
func (t *Toolkit) Coverage(pkg string) ([]CoverageReport, error) {
	if pkg == "" {
		pkg = "./..."
	}
	stdout, stderr, exitCode, err := t.Run("go", "test", "-cover", pkg)
	if err != nil && exitCode != 0 && !strings.Contains(stdout, "coverage:") {
		return nil, fmt.Errorf("go test -cover failed (exit %d): %s\n%s", exitCode, err, stderr)
	}

	var reports []CoverageReport
	re := regexp.MustCompile(`ok\s+(\S+)\s+.*coverage:\s+([\d.]+)%`)
	scanner := bufio.NewScanner(strings.NewReader(stdout))

	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) == 3 {
			pct, _ := strconv.ParseFloat(matches[2], 64)
			reports = append(reports, CoverageReport{
				Package:    matches[1],
				Percentage: pct,
			})
		}
	}
	return reports, nil
}

// RaceDetect runs go test -race and parses data race warnings.
func (t *Toolkit) RaceDetect(pkg string) ([]RaceCondition, error) {
	if pkg == "" {
		pkg = "./..."
	}
	_, stderr, _, err := t.Run("go", "test", "-race", pkg)
	if err != nil && !strings.Contains(stderr, "WARNING: DATA RACE") {
		return nil, fmt.Errorf("go test -race failed: %w", err)
	}

	var races []RaceCondition
	lines := strings.Split(stderr, "\n")
	reFile := regexp.MustCompile(`\s+(.*\.go):(\d+)`)

	for i, line := range lines {
		if strings.Contains(line, "WARNING: DATA RACE") {
			rc := RaceCondition{Desc: "Data race detected"}
			for j := i + 1; j < len(lines) && j < i+15; j++ {
				if match := reFile.FindStringSubmatch(lines[j]); len(match) == 3 {
					rc.File = strings.TrimSpace(match[1])
					rc.Line, _ = strconv.Atoi(match[2])
					break
				}
			}
			races = append(races, rc)
		}
	}
	return races, nil
}

// Complexity runs gocyclo and returns functions exceeding the threshold.
func (t *Toolkit) Complexity(threshold int) ([]ComplexFunc, error) {
	stdout, stderr, exitCode, err := t.Run("gocyclo", "-over", strconv.Itoa(threshold), ".")
	if err != nil && exitCode == -1 {
		return nil, fmt.Errorf("gocyclo not available: %s\n%s", err, stderr)
	}

	var funcs []ComplexFunc
	scanner := bufio.NewScanner(strings.NewReader(stdout))

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		score, _ := strconv.Atoi(fields[0])
		fileParts := strings.Split(fields[3], ":")
		line := 0
		if len(fileParts) > 1 {
			line, _ = strconv.Atoi(fileParts[1])
		}

		funcs = append(funcs, ComplexFunc{
			Score:    score,
			Package:  fields[1],
			FuncName: fields[2],
			File:     fileParts[0],
			Line:     line,
		})
	}
	return funcs, nil
}

// DepGraph runs go mod graph and builds a dependency graph.
func (t *Toolkit) DepGraph(pkg string) (*Graph, error) {
	stdout, stderr, exitCode, err := t.Run("go", "mod", "graph")
	if err != nil && exitCode != 0 {
		return nil, fmt.Errorf("go mod graph failed (exit %d): %s\n%s", exitCode, err, stderr)
	}

	graph := &Graph{Edges: make(map[string][]string)}
	nodes := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(stdout))

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			src, dst := parts[0], parts[1]
			graph.Edges[src] = append(graph.Edges[src], dst)
			nodes[src] = struct{}{}
			nodes[dst] = struct{}{}
		}
	}

	for node := range nodes {
		graph.Nodes = append(graph.Nodes, node)
	}
	return graph, nil
}

// GitLog returns the last n commits from git history.
func (t *Toolkit) GitLog(n int) ([]Commit, error) {
	stdout, stderr, exitCode, err := t.Run("git", "log", fmt.Sprintf("-n%d", n), "--format=%H|%an|%aI|%s")
	if err != nil && exitCode != 0 {
		return nil, fmt.Errorf("git log failed (exit %d): %s\n%s", exitCode, err, stderr)
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(stdout))

	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "|", 4)
		if len(parts) < 4 {
			continue
		}
		date, _ := time.Parse(time.RFC3339, parts[2])
		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    date,
			Message: parts[3],
		})
	}
	return commits, nil
}

// CheckPerms walks a directory and flags files with overly permissive modes.
func (t *Toolkit) CheckPerms(dir string) ([]PermIssue, error) {
	var issues []PermIssue
	err := filepath.Walk(filepath.Join(t.Dir, dir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		mode := info.Mode().Perm()
		if mode&0o002 != 0 {
			issues = append(issues, PermIssue{
				File:       path,
				Permission: fmt.Sprintf("%04o", mode),
				Issue:      "World-writable",
			})
		} else if mode&0o020 != 0 && mode&0o002 != 0 {
			issues = append(issues, PermIssue{
				File:       path,
				Permission: fmt.Sprintf("%04o", mode),
				Issue:      "Group and world-writable",
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk failed: %w", err)
	}
	return issues, nil
}

// LEK-1 | lthn.ai | EUPL-1.2
