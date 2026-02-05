package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubSource_Name_Good(t *testing.T) {
	src := NewGitHubSource(Config{Repos: []string{"owner/repo"}})
	assert.Equal(t, "github", src.Name())
}

func TestGitHubSource_Poll_Good(t *testing.T) {
	epic := ghIssue{
		Number: 10,
		Title:  "Epic: feature rollout",
		Body:   "Tasks:\n- [ ] #5\n- [x] #6\n- [ ] #7",
		Labels: []ghLabel{{Name: "epic"}},
		State:  "open",
	}

	pr5 := ghPR{
		Number:         50,
		Title:          "Implement child #5",
		Body:           "Closes #5",
		State:          "open",
		Draft:          false,
		MergeableState: "clean",
		Head:           ghRef{SHA: "abc123", Ref: "feature-5"},
	}

	// PR 7 has no linked reference to any child, so child #7 should not produce a signal.
	pr99 := ghPR{
		Number:         99,
		Title:          "Unrelated PR",
		Body:           "No issue links here",
		State:          "open",
		Draft:          false,
		MergeableState: "dirty",
		Head:           ghRef{SHA: "def456", Ref: "feature-other"},
	}

	checkSuites := ghCheckSuites{
		TotalCount: 1,
		CheckSuites: []ghCheckSuite{
			{ID: 1, Status: "completed", Conclusion: "success"},
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /repos/test-org/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "epic", r.URL.Query().Get("labels"))
		assert.Equal(t, "open", r.URL.Query().Get("state"))
		w.Header().Set("ETag", `"epic-etag-1"`)
		_ = json.NewEncoder(w).Encode([]ghIssue{epic})
	})

	mux.HandleFunc("GET /repos/test-org/test-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "open", r.URL.Query().Get("state"))
		_ = json.NewEncoder(w).Encode([]ghPR{pr5, pr99})
	})

	mux.HandleFunc("GET /repos/test-org/test-repo/commits/abc123/check-suites", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(checkSuites)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := NewGitHubSource(Config{
		Repos:  []string{"test-org/test-repo"},
		APIURL: srv.URL,
	})

	signals, err := src.Poll(context.Background())
	require.NoError(t, err)

	// Only child #5 has a linked PR (pr5 references #5 in body).
	// Child #7 is unchecked but no PR references it.
	// Child #6 is checked so it's ignored.
	require.Len(t, signals, 1)

	sig := signals[0]
	assert.Equal(t, 10, sig.EpicNumber)
	assert.Equal(t, 5, sig.ChildNumber)
	assert.Equal(t, 50, sig.PRNumber)
	assert.Equal(t, "test-org", sig.RepoOwner)
	assert.Equal(t, "test-repo", sig.RepoName)
	assert.Equal(t, "OPEN", sig.PRState)
	assert.Equal(t, false, sig.IsDraft)
	assert.Equal(t, "MERGEABLE", sig.Mergeable)
	assert.Equal(t, "SUCCESS", sig.CheckStatus)
	assert.Equal(t, "abc123", sig.LastCommitSHA)
}

func TestGitHubSource_Poll_Good_NotModified(t *testing.T) {
	callCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/test-org/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("ETag", `"etag-v1"`)
			_ = json.NewEncoder(w).Encode([]ghIssue{})
		} else {
			// Second call should have If-None-Match.
			assert.Equal(t, `"etag-v1"`, r.Header.Get("If-None-Match"))
			w.WriteHeader(http.StatusNotModified)
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := NewGitHubSource(Config{
		Repos:  []string{"test-org/test-repo"},
		APIURL: srv.URL,
	})

	// First poll — gets empty list, stores ETag.
	signals, err := src.Poll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, signals)

	// Second poll — sends If-None-Match, gets 304.
	signals, err = src.Poll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, signals)

	assert.Equal(t, 2, callCount)
}

func TestParseEpicChildren_Good(t *testing.T) {
	body := `## Epic

Tasks to complete:
- [ ] #1
- [x] #2
- [ ] #3
- [x] #4
- [ ] #5
`

	unchecked, checked := parseEpicChildren(body)

	assert.Equal(t, []int{1, 3, 5}, unchecked)
	assert.Equal(t, []int{2, 4}, checked)
}

func TestParseEpicChildren_Good_Empty(t *testing.T) {
	unchecked, checked := parseEpicChildren("No checklist here")
	assert.Nil(t, unchecked)
	assert.Nil(t, checked)
}

func TestFindLinkedPR_Good(t *testing.T) {
	prs := []ghPR{
		{Number: 10, Body: "Unrelated work"},
		{Number: 20, Body: "Fixes #42 and updates docs"},
		{Number: 30, Body: "Closes #99"},
	}

	pr := findLinkedPR(prs, 42)
	require.NotNil(t, pr)
	assert.Equal(t, 20, pr.Number)
}

func TestFindLinkedPR_Good_NoMatch(t *testing.T) {
	prs := []ghPR{
		{Number: 10, Body: "Unrelated work"},
		{Number: 20, Body: "Closes #99"},
	}

	pr := findLinkedPR(prs, 42)
	assert.Nil(t, pr)
}

func TestAggregateCheckStatus_Good(t *testing.T) {
	tests := []struct {
		name   string
		suites []ghCheckSuite
		want   string
	}{
		{
			name:   "all success",
			suites: []ghCheckSuite{{Status: "completed", Conclusion: "success"}},
			want:   "SUCCESS",
		},
		{
			name:   "one failure",
			suites: []ghCheckSuite{{Status: "completed", Conclusion: "failure"}},
			want:   "FAILURE",
		},
		{
			name:   "in progress",
			suites: []ghCheckSuite{{Status: "in_progress", Conclusion: ""}},
			want:   "PENDING",
		},
		{
			name:   "empty",
			suites: nil,
			want:   "PENDING",
		},
		{
			name: "mixed completed",
			suites: []ghCheckSuite{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "failure"},
			},
			want: "FAILURE",
		},
		{
			name: "neutral is success",
			suites: []ghCheckSuite{
				{Status: "completed", Conclusion: "neutral"},
				{Status: "completed", Conclusion: "success"},
			},
			want: "SUCCESS",
		},
		{
			name: "skipped is success",
			suites: []ghCheckSuite{
				{Status: "completed", Conclusion: "skipped"},
			},
			want: "SUCCESS",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := aggregateCheckStatus(tc.suites)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMergeableToString_Good(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"clean", "MERGEABLE"},
		{"has_hooks", "MERGEABLE"},
		{"unstable", "MERGEABLE"},
		{"dirty", "CONFLICTING"},
		{"blocked", "CONFLICTING"},
		{"unknown", "UNKNOWN"},
		{"", "UNKNOWN"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := mergeableToString(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGitHubSource_Report_Good(t *testing.T) {
	src := NewGitHubSource(Config{Repos: []string{"owner/repo"}})
	err := src.Report(context.Background(), nil)
	assert.NoError(t, err)
}
