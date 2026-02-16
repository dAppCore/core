// Designed by Gemini 3 Pro (Hypnos) + Claude Opus (Charon), signed LEK-1 | lthn.ai | EUPL-1.2
package devkit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupMockCmd creates a shell script in a temp dir that echoes predetermined
// content, and prepends that dir to PATH so Run() picks it up.
func setupMockCmd(t *testing.T, name, content string) {
	t.Helper()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, name)

	script := fmt.Sprintf("#!/bin/sh\ncat <<'MOCK_EOF'\n%s\nMOCK_EOF\n", content)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write mock command %s: %v", name, err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
}

// setupMockCmdExit creates a mock that echoes to stdout/stderr and exits with a code.
func setupMockCmdExit(t *testing.T, name, stdout, stderr string, exitCode int) {
	t.Helper()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, name)

	script := fmt.Sprintf("#!/bin/sh\ncat <<'MOCK_EOF'\n%s\nMOCK_EOF\ncat <<'MOCK_ERR' >&2\n%s\nMOCK_ERR\nexit %d\n", stdout, stderr, exitCode)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write mock command %s: %v", name, err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
}

func TestCoverage_Good(t *testing.T) {
	output := `?   	example.com/skipped	[no test files]
ok  	example.com/pkg1	0.5s	coverage: 85.0% of statements
ok  	example.com/pkg2	0.2s	coverage: 100.0% of statements`

	setupMockCmd(t, "go", output)

	tk := New(t.TempDir())
	reports, err := tk.Coverage("./...")
	if err != nil {
		t.Fatalf("Coverage failed: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0].Package != "example.com/pkg1" || reports[0].Percentage != 85.0 {
		t.Errorf("report 0: want pkg1@85%%, got %s@%.1f%%", reports[0].Package, reports[0].Percentage)
	}
	if reports[1].Package != "example.com/pkg2" || reports[1].Percentage != 100.0 {
		t.Errorf("report 1: want pkg2@100%%, got %s@%.1f%%", reports[1].Package, reports[1].Percentage)
	}
}

func TestCoverage_Bad(t *testing.T) {
	// No coverage lines in output
	setupMockCmd(t, "go", "FAIL\texample.com/broken [build failed]")

	tk := New(t.TempDir())
	reports, err := tk.Coverage("./...")
	if err != nil {
		t.Fatalf("Coverage should not error on partial output: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("expected 0 reports from failed build, got %d", len(reports))
	}
}

func TestGitLog_Good(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	nowStr := now.Format(time.RFC3339)

	output := fmt.Sprintf("abc123|Alice|%s|Fix the bug\ndef456|Bob|%s|Add feature", nowStr, nowStr)
	setupMockCmd(t, "git", output)

	tk := New(t.TempDir())
	commits, err := tk.GitLog(2)
	if err != nil {
		t.Fatalf("GitLog failed: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc123" {
		t.Errorf("hash: want abc123, got %s", commits[0].Hash)
	}
	if commits[0].Author != "Alice" {
		t.Errorf("author: want Alice, got %s", commits[0].Author)
	}
	if commits[0].Message != "Fix the bug" {
		t.Errorf("message: want 'Fix the bug', got %q", commits[0].Message)
	}
	if !commits[0].Date.Equal(now) {
		t.Errorf("date: want %v, got %v", now, commits[0].Date)
	}
}

func TestGitLog_Bad(t *testing.T) {
	// Malformed lines should be skipped
	setupMockCmd(t, "git", "incomplete|line\nabc|Bob|2025-01-01T00:00:00Z|Good commit")

	tk := New(t.TempDir())
	commits, err := tk.GitLog(5)
	if err != nil {
		t.Fatalf("GitLog failed: %v", err)
	}
	if len(commits) != 1 {
		t.Errorf("expected 1 valid commit (skip malformed), got %d", len(commits))
	}
}

func TestComplexity_Good(t *testing.T) {
	output := "15 main ComplexFunc file.go:10:1\n20 pkg VeryComplex other.go:50:1"
	setupMockCmd(t, "gocyclo", output)

	tk := New(t.TempDir())
	funcs, err := tk.Complexity(10)
	if err != nil {
		t.Fatalf("Complexity failed: %v", err)
	}
	if len(funcs) != 2 {
		t.Fatalf("expected 2 funcs, got %d", len(funcs))
	}
	if funcs[0].Score != 15 || funcs[0].FuncName != "ComplexFunc" || funcs[0].File != "file.go" || funcs[0].Line != 10 {
		t.Errorf("func 0: unexpected %+v", funcs[0])
	}
	if funcs[1].Score != 20 || funcs[1].Package != "pkg" {
		t.Errorf("func 1: unexpected %+v", funcs[1])
	}
}

func TestComplexity_Bad(t *testing.T) {
	// No functions above threshold = empty output
	setupMockCmd(t, "gocyclo", "")

	tk := New(t.TempDir())
	funcs, err := tk.Complexity(50)
	if err != nil {
		t.Fatalf("Complexity should not error on empty output: %v", err)
	}
	if len(funcs) != 0 {
		t.Errorf("expected 0 funcs, got %d", len(funcs))
	}
}

func TestDepGraph_Good(t *testing.T) {
	output := "modA@v1 modB@v2\nmodA@v1 modC@v3\nmodB@v2 modD@v1"
	setupMockCmd(t, "go", output)

	tk := New(t.TempDir())
	graph, err := tk.DepGraph("./...")
	if err != nil {
		t.Fatalf("DepGraph failed: %v", err)
	}
	if len(graph.Nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d: %v", len(graph.Nodes), graph.Nodes)
	}
	edgesA := graph.Edges["modA@v1"]
	if len(edgesA) != 2 {
		t.Errorf("expected 2 edges from modA@v1, got %d", len(edgesA))
	}
}

func TestRaceDetect_Good(t *testing.T) {
	// No races = clean run
	setupMockCmd(t, "go", "ok\texample.com/safe\t0.1s")

	tk := New(t.TempDir())
	races, err := tk.RaceDetect("./...")
	if err != nil {
		t.Fatalf("RaceDetect failed on clean run: %v", err)
	}
	if len(races) != 0 {
		t.Errorf("expected 0 races, got %d", len(races))
	}
}

func TestRaceDetect_Bad(t *testing.T) {
	stderrOut := `WARNING: DATA RACE
Read at 0x00c000123456 by goroutine 7:
      /home/user/project/main.go:42
Previous write at 0x00c000123456 by goroutine 6:
      /home/user/project/main.go:38`

	setupMockCmdExit(t, "go", "", stderrOut, 1)

	tk := New(t.TempDir())
	races, err := tk.RaceDetect("./...")
	if err != nil {
		t.Fatalf("RaceDetect should parse races, not error: %v", err)
	}
	if len(races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(races))
	}
	if races[0].File != "/home/user/project/main.go" || races[0].Line != 42 {
		t.Errorf("race: unexpected %+v", races[0])
	}
}

func TestDiffStat_Good(t *testing.T) {
	output := ` file1.go | 10 +++++++---
 file2.go |  5 +++++
 2 files changed, 12 insertions(+), 3 deletions(-)`
	setupMockCmd(t, "git", output)

	tk := New(t.TempDir())
	s, err := tk.DiffStat()
	if err != nil {
		t.Fatalf("DiffStat failed: %v", err)
	}
	if s.FilesChanged != 2 {
		t.Errorf("files: want 2, got %d", s.FilesChanged)
	}
	if s.Insertions != 12 {
		t.Errorf("insertions: want 12, got %d", s.Insertions)
	}
	if s.Deletions != 3 {
		t.Errorf("deletions: want 3, got %d", s.Deletions)
	}
}

func TestCheckPerms_Good(t *testing.T) {
	dir := t.TempDir()

	// Create a world-writable file
	badFile := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(badFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(badFile, 0666); err != nil {
		t.Fatal(err)
	}
	// Create a safe file
	goodFile := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(goodFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tk := New("/")
	issues, err := tk.CheckPerms(dir)
	if err != nil {
		t.Fatalf("CheckPerms failed: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue (world-writable), got %d", len(issues))
	}
	if issues[0].Issue != "World-writable" {
		t.Errorf("issue: want 'World-writable', got %q", issues[0].Issue)
	}
}

func TestNew(t *testing.T) {
	tk := New("/tmp")
	if tk.Dir != "/tmp" {
		t.Errorf("Dir: want /tmp, got %s", tk.Dir)
	}
}

// LEK-1 | lthn.ai | EUPL-1.2
