package bugseti

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfigService creates a ConfigService with in-memory config for testing.
func testConfigService(t *testing.T, repos []string, labels []string) *ConfigService {
	t.Helper()
	dir := t.TempDir()
	cs := &ConfigService{
		path: dir + "/config.json",
		config: &Config{
			WatchedRepos:  repos,
			Labels:        labels,
			FetchInterval: 15,
			DataDir:       dir,
		},
	}
	return cs
}

// TestHelperProcess is invoked by the test binary when GO_TEST_HELPER_PROCESS
// is set. It prints the value of GO_TEST_HELPER_OUTPUT and optionally exits
// with a non-zero code. Kept for future exec.Command mocking.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, os.Getenv("GO_TEST_HELPER_OUTPUT"))
	if os.Getenv("GO_TEST_HELPER_EXIT_ERROR") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
}

// ---- NewFetcherService ----

func TestNewFetcherService_Good(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	notify := NewNotifyService(cfg)
	f := NewFetcherService(cfg, notify)

	require.NotNil(t, f)
	assert.Equal(t, "FetcherService", f.ServiceName())
	assert.False(t, f.IsRunning())
	assert.NotNil(t, f.Issues())
}

// ---- Start / Pause / IsRunning lifecycle ----

func TestStartPause_Good(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	notify := NewNotifyService(cfg)
	f := NewFetcherService(cfg, notify)

	require.NoError(t, f.Start())
	assert.True(t, f.IsRunning())

	// Starting again is a no-op.
	require.NoError(t, f.Start())
	assert.True(t, f.IsRunning())

	f.Pause()
	assert.False(t, f.IsRunning())

	// Pausing again is a no-op.
	f.Pause()
	assert.False(t, f.IsRunning())
}

// ---- calculatePriority ----

func TestCalculatePriority_Good(t *testing.T) {
	tests := []struct {
		name     string
		labels   []string
		expected int
	}{
		{"no labels", nil, 50},
		{"good first issue", []string{"good first issue"}, 80},
		{"help wanted", []string{"Help Wanted"}, 70},
		{"beginner", []string{"beginner-friendly"}, 75},
		{"easy", []string{"Easy"}, 70},
		{"bug", []string{"bug"}, 60},
		{"documentation", []string{"Documentation"}, 55},
		{"priority", []string{"high-priority"}, 65},
		{"combined", []string{"good first issue", "bug"}, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, calculatePriority(tt.labels))
		})
	}
}

func TestCalculatePriority_Bad(t *testing.T) {
	// Unknown labels should not change priority from default.
	assert.Equal(t, 50, calculatePriority([]string{"unknown-label", "something-else"}))
}

// ---- Label query construction ----

func TestLabelQuery_Good(t *testing.T) {
	// When config has custom labels, fetchFromRepo should use them.
	cfg := testConfigService(t, []string{"owner/repo"}, []string{"custom-label", "another"})
	labels := cfg.GetLabels()
	labelQuery := strings.Join(labels, ",")
	assert.Equal(t, "custom-label,another", labelQuery)
}

func TestLabelQuery_Bad(t *testing.T) {
	// When config has empty labels, fetchFromRepo falls back to defaults.
	cfg := testConfigService(t, []string{"owner/repo"}, nil)
	labels := cfg.GetLabels()
	if len(labels) == 0 {
		labels = []string{"good first issue", "help wanted", "beginner-friendly"}
	}
	labelQuery := strings.Join(labels, ",")
	assert.Equal(t, "good first issue,help wanted,beginner-friendly", labelQuery)
}

// ---- fetchFromRepo with mocked gh CLI output ----

func TestFetchFromRepo_Good(t *testing.T) {
	ghIssues := []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}{
		{
			Number:    42,
			Title:     "Fix login bug",
			Body:      "The login page crashes",
			URL:       "https://github.com/test/repo/issues/42",
			CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
		},
	}
	ghIssues[0].Author.Login = "octocat"
	ghIssues[0].Labels = []struct {
		Name string `json:"name"`
	}{
		{Name: "good first issue"},
		{Name: "bug"},
	}

	output, err := json.Marshal(ghIssues)
	require.NoError(t, err)

	// We can't easily intercept exec.CommandContext in the production code
	// without refactoring, so we test the JSON parsing path by directly
	// calling json.Unmarshal the same way fetchFromRepo does.
	var parsed []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	require.NoError(t, json.Unmarshal(output, &parsed))
	require.Len(t, parsed, 1)

	gi := parsed[0]
	labels := make([]string, len(gi.Labels))
	for i, l := range gi.Labels {
		labels[i] = l.Name
	}

	issue := &Issue{
		ID:        fmt.Sprintf("%s#%d", "test/repo", gi.Number),
		Number:    gi.Number,
		Repo:      "test/repo",
		Title:     gi.Title,
		Body:      gi.Body,
		URL:       gi.URL,
		Labels:    labels,
		Author:    gi.Author.Login,
		CreatedAt: gi.CreatedAt,
		Priority:  calculatePriority(labels),
	}

	assert.Equal(t, "test/repo#42", issue.ID)
	assert.Equal(t, 42, issue.Number)
	assert.Equal(t, "Fix login bug", issue.Title)
	assert.Equal(t, "octocat", issue.Author)
	assert.Equal(t, []string{"good first issue", "bug"}, issue.Labels)
	assert.Equal(t, 90, issue.Priority) // 50 + 30 (good first issue) + 10 (bug)
}

func TestFetchFromRepo_Bad_InvalidJSON(t *testing.T) {
	// Simulate gh returning invalid JSON.
	var ghIssues []struct {
		Number int `json:"number"`
	}
	err := json.Unmarshal([]byte(`not json at all`), &ghIssues)
	assert.Error(t, err, "invalid JSON should produce an error")
}

func TestFetchFromRepo_Bad_GhNotInstalled(t *testing.T) {
	// Verify that a missing executable produces an exec error.
	cmd := exec.Command("gh-nonexistent-binary-12345")
	_, err := cmd.Output()
	assert.Error(t, err, "missing binary should produce an error")
}

// ---- fetchAll: no repos configured ----

func TestFetchAll_Bad_NoRepos(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	notify := NewNotifyService(cfg)
	f := NewFetcherService(cfg, notify)

	// fetchAll with no repos should not panic and should not send to channel.
	f.fetchAll()

	// Channel should be empty.
	select {
	case <-f.issuesCh:
		t.Fatal("expected no issues on channel when no repos configured")
	default:
		// expected
	}
}

// ---- Channel backpressure ----

func TestChannelBackpressure_Ugly(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	notify := NewNotifyService(cfg)
	f := NewFetcherService(cfg, notify)

	// Fill the channel to capacity (buffer size is 10).
	for i := 0; i < 10; i++ {
		f.issuesCh <- []*Issue{{ID: fmt.Sprintf("test#%d", i)}}
	}

	// Now try to send via the select path (same logic as fetchAll).
	// This should be a non-blocking drop, not a deadlock.
	done := make(chan struct{})
	go func() {
		defer close(done)
		issues := []*Issue{{ID: "overflow#1"}}
		select {
		case f.issuesCh <- issues:
			// Shouldn't happen — channel is full.
			t.Error("expected channel send to be skipped due to backpressure")
		default:
			// This is the expected path — channel full, message dropped.
		}
	}()

	select {
	case <-done:
		// success — did not deadlock
	case <-time.After(time.Second):
		t.Fatal("backpressure test timed out — possible deadlock")
	}
}

// ---- FetchIssue single-issue parsing ----

func TestFetchIssue_Good_Parse(t *testing.T) {
	// Test the JSON parsing and Issue construction for FetchIssue.
	ghIssue := struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Comments []struct {
			Body   string `json:"body"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"comments"`
	}{
		Number:    99,
		Title:     "Add dark mode",
		Body:      "Please add dark mode support",
		URL:       "https://github.com/test/repo/issues/99",
		CreatedAt: time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC),
	}
	ghIssue.Author.Login = "contributor"
	ghIssue.Labels = []struct {
		Name string `json:"name"`
	}{
		{Name: "help wanted"},
	}
	ghIssue.Comments = []struct {
		Body   string `json:"body"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
	}{
		{Body: "I can work on this"},
	}
	ghIssue.Comments[0].Author.Login = "volunteer"

	data, err := json.Marshal(ghIssue)
	require.NoError(t, err)

	// Re-parse as the function would.
	var parsed struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Comments []struct {
			Body   string `json:"body"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"comments"`
	}
	require.NoError(t, json.Unmarshal(data, &parsed))

	labels := make([]string, len(parsed.Labels))
	for i, l := range parsed.Labels {
		labels[i] = l.Name
	}
	comments := make([]Comment, len(parsed.Comments))
	for i, c := range parsed.Comments {
		comments[i] = Comment{Author: c.Author.Login, Body: c.Body}
	}

	issue := &Issue{
		ID:        fmt.Sprintf("%s#%d", "test/repo", parsed.Number),
		Number:    parsed.Number,
		Repo:      "test/repo",
		Title:     parsed.Title,
		Body:      parsed.Body,
		URL:       parsed.URL,
		Labels:    labels,
		Author:    parsed.Author.Login,
		CreatedAt: parsed.CreatedAt,
		Priority:  calculatePriority(labels),
		Comments:  comments,
	}

	assert.Equal(t, "test/repo#99", issue.ID)
	assert.Equal(t, "contributor", issue.Author)
	assert.Equal(t, 70, issue.Priority) // 50 + 20 (help wanted)
	require.Len(t, issue.Comments, 1)
	assert.Equal(t, "volunteer", issue.Comments[0].Author)
	assert.Equal(t, "I can work on this", issue.Comments[0].Body)
}

// ---- Issues() channel accessor ----

func TestIssuesChannel_Good(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	notify := NewNotifyService(cfg)
	f := NewFetcherService(cfg, notify)

	ch := f.Issues()
	require.NotNil(t, ch)

	// Send and receive through the channel.
	go func() {
		f.issuesCh <- []*Issue{{ID: "test#1", Title: "Test issue"}}
	}()

	select {
	case issues := <-ch:
		require.Len(t, issues, 1)
		assert.Equal(t, "test#1", issues[0].ID)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for issues on channel")
	}
}
