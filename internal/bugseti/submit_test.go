package bugseti

import (
	"strings"
	"testing"
)

// helpers to build minimal service dependencies without touching disk

func testConfigService(t *testing.T) *ConfigService {
	t.Helper()
	dir := t.TempDir()
	return &ConfigService{
		path: dir + "/config.json",
		config: &Config{
			DataDir: dir,
		},
	}
}

func testSubmitService(t *testing.T) *SubmitService {
	t.Helper()
	cfg := testConfigService(t)
	notify := &NotifyService{enabled: false, config: cfg}
	stats := &StatsService{
		config: cfg,
		stats: &Stats{
			ReposContributed: make(map[string]*RepoStats),
			DailyActivity:    make(map[string]*DayStats),
		},
	}
	return NewSubmitService(cfg, notify, stats)
}

// --- NewSubmitService / ServiceName ---

func TestNewSubmitService_Good(t *testing.T) {
	s := testSubmitService(t)
	if s == nil {
		t.Fatal("expected non-nil SubmitService")
	}
	if s.config == nil || s.notify == nil || s.stats == nil {
		t.Fatal("expected all dependencies set")
	}
}

func TestServiceName_Good(t *testing.T) {
	s := testSubmitService(t)
	if got := s.ServiceName(); got != "SubmitService" {
		t.Fatalf("expected %q, got %q", "SubmitService", got)
	}
}

// --- Submit validation ---

func TestSubmit_Bad_NilSubmission(t *testing.T) {
	s := testSubmitService(t)
	_, err := s.Submit(nil)
	if err == nil {
		t.Fatal("expected error for nil submission")
	}
	if !strings.Contains(err.Error(), "invalid submission") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubmit_Bad_NilIssue(t *testing.T) {
	s := testSubmitService(t)
	_, err := s.Submit(&PRSubmission{Issue: nil})
	if err == nil {
		t.Fatal("expected error for nil issue")
	}
	if !strings.Contains(err.Error(), "invalid submission") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubmit_Bad_EmptyWorkDir(t *testing.T) {
	s := testSubmitService(t)
	_, err := s.Submit(&PRSubmission{
		Issue:   &Issue{Number: 1, Repo: "owner/repo", Title: "test"},
		WorkDir: "",
	})
	if err == nil {
		t.Fatal("expected error for empty work directory")
	}
	if !strings.Contains(err.Error(), "work directory not specified") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- buildForkURL ---

func TestBuildForkURL_Good_HTTPS(t *testing.T) {
	origin := "https://github.com/upstream-owner/my-repo.git"
	got := buildForkURL(origin, "myfork")
	want := "https://github.com/myfork/my-repo.git"
	if got != want {
		t.Fatalf("HTTPS fork URL:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestBuildForkURL_Good_HTTPSNoGitSuffix(t *testing.T) {
	origin := "https://github.com/upstream-owner/my-repo"
	got := buildForkURL(origin, "myfork")
	want := "https://github.com/myfork/my-repo"
	if got != want {
		t.Fatalf("HTTPS fork URL without .git:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestBuildForkURL_Good_SSH(t *testing.T) {
	origin := "git@github.com:upstream-owner/my-repo.git"
	got := buildForkURL(origin, "myfork")
	want := "git@github.com:myfork/my-repo.git"
	if got != want {
		t.Fatalf("SSH fork URL:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestBuildForkURL_Good_SSHNoGitSuffix(t *testing.T) {
	origin := "git@github.com:upstream-owner/my-repo"
	got := buildForkURL(origin, "myfork")
	want := "git@github.com:myfork/my-repo"
	if got != want {
		t.Fatalf("SSH fork URL without .git:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestBuildForkURL_Bad_ShortHTTPS(t *testing.T) {
	// URL with fewer than 4 parts after split returns unchanged
	origin := "https://x"
	got := buildForkURL(origin, "fork")
	if got != origin {
		t.Fatalf("expected unchanged URL for short HTTPS, got: %s", got)
	}
}

// --- generatePRBody ---

func TestGeneratePRBody_Good_Basic(t *testing.T) {
	s := testSubmitService(t)
	issue := &Issue{Number: 42, Repo: "owner/repo", Title: "A bug"}
	body := s.generatePRBody(issue)

	if !strings.Contains(body, "#42") {
		t.Fatal("PR body should reference issue number")
	}
	if !strings.Contains(body, "## Summary") {
		t.Fatal("PR body should have Summary section")
	}
	if !strings.Contains(body, "## Changes") {
		t.Fatal("PR body should have Changes section")
	}
	if !strings.Contains(body, "## Testing") {
		t.Fatal("PR body should have Testing section")
	}
	if !strings.Contains(body, "BugSETI") {
		t.Fatal("PR body should have BugSETI attribution")
	}
}

func TestGeneratePRBody_Good_WithContext(t *testing.T) {
	s := testSubmitService(t)
	issue := &Issue{
		Number: 7,
		Repo:   "owner/repo",
		Title:  "Fix login",
		Context: &IssueContext{
			Summary: "The login endpoint returns 500 on empty password.",
		},
	}
	body := s.generatePRBody(issue)

	if !strings.Contains(body, "## Context") {
		t.Fatal("PR body should have Context section when context exists")
	}
	if !strings.Contains(body, "login endpoint returns 500") {
		t.Fatal("PR body should include context summary")
	}
}

func TestGeneratePRBody_Good_WithoutContext(t *testing.T) {
	s := testSubmitService(t)
	issue := &Issue{Number: 7, Repo: "owner/repo", Title: "Fix login"}
	body := s.generatePRBody(issue)

	if strings.Contains(body, "## Context") {
		t.Fatal("PR body should omit Context section when no context")
	}
}

func TestGeneratePRBody_Good_EmptyContextSummary(t *testing.T) {
	s := testSubmitService(t)
	issue := &Issue{
		Number:  7,
		Repo:    "owner/repo",
		Title:   "Fix login",
		Context: &IssueContext{Summary: ""},
	}
	body := s.generatePRBody(issue)

	if strings.Contains(body, "## Context") {
		t.Fatal("PR body should omit Context section when summary is empty")
	}
}

// --- PRSubmission / PRResult struct tests ---

func TestPRSubmission_Good_Defaults(t *testing.T) {
	sub := &PRSubmission{
		Issue:   &Issue{Number: 10, Repo: "o/r"},
		WorkDir: "/tmp/work",
	}
	if sub.Branch != "" {
		t.Fatal("expected empty branch to be default")
	}
	if sub.Title != "" {
		t.Fatal("expected empty title to be default")
	}
	if sub.CommitMsg != "" {
		t.Fatal("expected empty commit msg to be default")
	}
}

func TestPRResult_Good_Success(t *testing.T) {
	r := &PRResult{
		Success:   true,
		PRURL:     "https://github.com/o/r/pull/1",
		PRNumber:  1,
		ForkOwner: "me",
	}
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Error != "" {
		t.Fatal("expected no error on success")
	}
}

func TestPRResult_Good_Failure(t *testing.T) {
	r := &PRResult{
		Success: false,
		Error:   "fork failed: something",
	}
	if r.Success {
		t.Fatal("expected failure")
	}
	if r.Error == "" {
		t.Fatal("expected error message")
	}
}

// --- PRStatus struct ---

func TestPRStatus_Good(t *testing.T) {
	s := &PRStatus{
		State:     "OPEN",
		Mergeable: true,
		CIPassing: true,
		Approved:  false,
	}
	if s.State != "OPEN" {
		t.Fatalf("expected OPEN, got %s", s.State)
	}
	if !s.Mergeable {
		t.Fatal("expected mergeable")
	}
	if s.Approved {
		t.Fatal("expected not approved")
	}
}

// --- ensureFork validation ---

func TestEnsureFork_Bad_InvalidRepoFormat(t *testing.T) {
	s := testSubmitService(t)
	_, err := s.ensureFork("invalidrepo")
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
	if !strings.Contains(err.Error(), "invalid repo format") {
		t.Fatalf("unexpected error: %v", err)
	}
}
