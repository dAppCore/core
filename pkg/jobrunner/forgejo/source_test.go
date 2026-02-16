package forgejo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"forge.lthn.ai/core/go/pkg/forge"
	"forge.lthn.ai/core/go/pkg/jobrunner"
)

// withVersion wraps an HTTP handler to serve the Forgejo /api/v1/version
// endpoint that the SDK calls during NewClient initialization.
func withVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/version") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"9.0.0"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newTestClient(t *testing.T, url string) *forge.Client {
	t.Helper()
	client, err := forge.New(url, "test-token")
	require.NoError(t, err)
	return client
}

func TestForgejoSource_Name(t *testing.T) {
	s := New(Config{}, nil)
	assert.Equal(t, "forgejo", s.Name())
}

func TestForgejoSource_Poll_Good(t *testing.T) {
	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		w.Header().Set("Content-Type", "application/json")

		switch {
		// List issues — return one epic
		case strings.Contains(path, "/issues"):
			issues := []map[string]any{
				{
					"number": 10,
					"body":   "## Tasks\n- [ ] #11\n- [x] #12\n",
					"labels": []map[string]string{{"name": "epic"}},
					"state":  "open",
				},
			}
			_ = json.NewEncoder(w).Encode(issues)

		// List PRs — return one open PR linked to #11
		case strings.Contains(path, "/pulls"):
			prs := []map[string]any{
				{
					"number":    20,
					"body":      "Fixes #11",
					"state":     "open",
					"mergeable": true,
					"merged":    false,
					"head":      map[string]string{"sha": "abc123", "ref": "feature", "label": "feature"},
				},
			}
			_ = json.NewEncoder(w).Encode(prs)

		// Combined status
		case strings.Contains(path, "/status"):
			status := map[string]any{
				"state":       "success",
				"total_count": 1,
				"statuses":    []map[string]any{{"status": "success", "context": "ci"}},
			}
			_ = json.NewEncoder(w).Encode(status)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	s := New(Config{Repos: []string{"test-org/test-repo"}}, client)

	signals, err := s.Poll(context.Background())
	require.NoError(t, err)

	require.Len(t, signals, 1)
	sig := signals[0]
	assert.Equal(t, 10, sig.EpicNumber)
	assert.Equal(t, 11, sig.ChildNumber)
	assert.Equal(t, 20, sig.PRNumber)
	assert.Equal(t, "OPEN", sig.PRState)
	assert.Equal(t, "MERGEABLE", sig.Mergeable)
	assert.Equal(t, "SUCCESS", sig.CheckStatus)
	assert.Equal(t, "test-org", sig.RepoOwner)
	assert.Equal(t, "test-repo", sig.RepoName)
	assert.Equal(t, "abc123", sig.LastCommitSHA)
}

func TestForgejoSource_Poll_NoEpics(t *testing.T) {
	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{})
	})))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	s := New(Config{Repos: []string{"test-org/test-repo"}}, client)

	signals, err := s.Poll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, signals)
}

func TestForgejoSource_Report_Good(t *testing.T) {
	var capturedBody string

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		capturedBody = body["body"]
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1})
	})))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	s := New(Config{}, client)

	result := &jobrunner.ActionResult{
		Action:      "enable_auto_merge",
		RepoOwner:   "test-org",
		RepoName:    "test-repo",
		EpicNumber:  10,
		ChildNumber: 11,
		PRNumber:    20,
		Success:     true,
	}

	err := s.Report(context.Background(), result)
	require.NoError(t, err)
	assert.Contains(t, capturedBody, "enable_auto_merge")
	assert.Contains(t, capturedBody, "succeeded")
}

func TestParseEpicChildren(t *testing.T) {
	body := "## Tasks\n- [x] #1\n- [ ] #7\n- [ ] #8\n- [x] #3\n"
	unchecked, checked := parseEpicChildren(body)
	assert.Equal(t, []int{7, 8}, unchecked)
	assert.Equal(t, []int{1, 3}, checked)
}

func TestFindLinkedPR(t *testing.T) {
	assert.Nil(t, findLinkedPR(nil, 7))
}

func TestSplitRepo(t *testing.T) {
	owner, repo, err := splitRepo("host-uk/core")
	require.NoError(t, err)
	assert.Equal(t, "host-uk", owner)
	assert.Equal(t, "core", repo)

	_, _, err = splitRepo("invalid")
	assert.Error(t, err)

	_, _, err = splitRepo("")
	assert.Error(t, err)
}
