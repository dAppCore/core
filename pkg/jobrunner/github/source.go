package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/oauth2"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/host-uk/core/pkg/log"
)

// Config configures a GitHubSource.
type Config struct {
	Repos  []string // "owner/repo" format
	APIURL string   // override for testing (default: https://api.github.com)
}

// GitHubSource polls GitHub for pipeline signals from epic issues.
type GitHubSource struct {
	repos  []string
	apiURL string
	client *http.Client
	etags  map[string]string
	mu     sync.Mutex
}

// NewGitHubSource creates a GitHubSource from the given config.
func NewGitHubSource(cfg Config) *GitHubSource {
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = "https://api.github.com"
	}

	// Build an authenticated HTTP client if GITHUB_TOKEN is set.
	var client *http.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.Background(), ts)
	} else {
		client = http.DefaultClient
	}

	return &GitHubSource{
		repos:  cfg.Repos,
		apiURL: strings.TrimRight(apiURL, "/"),
		client: client,
		etags:  make(map[string]string),
	}
}

// Name returns the source identifier.
func (g *GitHubSource) Name() string {
	return "github"
}

// Poll fetches epics and their linked PRs from all configured repositories,
// returning a PipelineSignal for each unchecked child that has a linked PR.
func (g *GitHubSource) Poll(ctx context.Context) ([]*jobrunner.PipelineSignal, error) {
	var signals []*jobrunner.PipelineSignal

	for _, repoFull := range g.repos {
		owner, repo, err := splitRepo(repoFull)
		if err != nil {
			log.Error("invalid repo format", "repo", repoFull, "err", err)
			continue
		}

		repoSignals, err := g.pollRepo(ctx, owner, repo)
		if err != nil {
			log.Error("poll repo failed", "repo", repoFull, "err", err)
			continue
		}

		signals = append(signals, repoSignals...)
	}

	return signals, nil
}

// Report is a no-op for the GitHub source.
func (g *GitHubSource) Report(_ context.Context, _ *jobrunner.ActionResult) error {
	return nil
}

// pollRepo fetches epics and PRs for a single repository.
func (g *GitHubSource) pollRepo(ctx context.Context, owner, repo string) ([]*jobrunner.PipelineSignal, error) {
	// Fetch epic issues (label=epic).
	epicsURL := fmt.Sprintf("%s/repos/%s/%s/issues?labels=epic&state=open", g.apiURL, owner, repo)
	var epics []ghIssue
	notModified, err := g.fetchJSON(ctx, epicsURL, &epics)
	if err != nil {
		return nil, fmt.Errorf("fetch epics: %w", err)
	}
	if notModified {
		log.Debug("epics not modified", "repo", owner+"/"+repo)
		return nil, nil
	}

	if len(epics) == 0 {
		return nil, nil
	}

	// Fetch open PRs.
	prsURL := fmt.Sprintf("%s/repos/%s/%s/pulls?state=open", g.apiURL, owner, repo)
	var prs []ghPR
	_, err = g.fetchJSON(ctx, prsURL, &prs)
	if err != nil {
		return nil, fmt.Errorf("fetch PRs: %w", err)
	}

	var signals []*jobrunner.PipelineSignal

	for _, epic := range epics {
		unchecked, _ := parseEpicChildren(epic.Body)
		for _, childNum := range unchecked {
			pr := findLinkedPR(prs, childNum)
			if pr == nil {
				continue
			}

			// Fetch check suites for the PR's head SHA.
			checksURL := fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-suites", g.apiURL, owner, repo, pr.Head.SHA)
			var checkResp ghCheckSuites
			_, err := g.fetchJSON(ctx, checksURL, &checkResp)
			if err != nil {
				log.Error("fetch check suites failed", "repo", owner+"/"+repo, "sha", pr.Head.SHA, "err", err)
				continue
			}

			checkStatus := aggregateCheckStatus(checkResp.CheckSuites)
			sig := buildSignal(owner, repo, epic.Number, childNum, pr, checkStatus)
			signals = append(signals, sig)
		}
	}

	return signals, nil
}

// fetchJSON performs a GET request with ETag conditional headers.
// Returns true if the server responded with 304 Not Modified.
func (g *GitHubSource) fetchJSON(ctx context.Context, url string, target any) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	g.mu.Lock()
	if etag, ok := g.etags[url]; ok {
		req.Header.Set("If-None-Match", etag)
	}
	g.mu.Unlock()

	resp, err := g.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified {
		return true, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	// Store ETag for future conditional requests.
	if etag := resp.Header.Get("ETag"); etag != "" {
		g.mu.Lock()
		g.etags[url] = etag
		g.mu.Unlock()
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return false, nil
}

// splitRepo parses "owner/repo" into its components.
func splitRepo(full string) (string, string, error) {
	parts := strings.SplitN(full, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected owner/repo format, got %q", full)
	}
	return parts[0], parts[1], nil
}
