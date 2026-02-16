// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"forge.lthn.ai/core/cli/pkg/forge"
)

// FetcherService fetches issues from configured OSS repositories.
type FetcherService struct {
	config   *ConfigService
	notify   *NotifyService
	forge    *forge.Client
	running  bool
	mu       sync.RWMutex
	stopCh   chan struct{}
	issuesCh chan []*Issue
}

// NewFetcherService creates a new FetcherService.
func NewFetcherService(config *ConfigService, notify *NotifyService, forgeClient *forge.Client) *FetcherService {
	return &FetcherService{
		config:   config,
		notify:   notify,
		forge:    forgeClient,
		issuesCh: make(chan []*Issue, 10),
	}
}

// ServiceName returns the service name for Wails.
func (f *FetcherService) ServiceName() string {
	return "FetcherService"
}

// Start begins fetching issues from configured repositories.
func (f *FetcherService) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return nil
	}

	f.running = true
	f.stopCh = make(chan struct{})

	go f.fetchLoop()
	log.Println("FetcherService started")
	return nil
}

// Pause stops fetching issues.
func (f *FetcherService) Pause() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return
	}

	f.running = false
	close(f.stopCh)
	log.Println("FetcherService paused")
}

// IsRunning returns whether the fetcher is actively running.
func (f *FetcherService) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.running
}

// Issues returns a channel that receives batches of fetched issues.
func (f *FetcherService) Issues() <-chan []*Issue {
	return f.issuesCh
}

// fetchLoop periodically fetches issues from all configured repositories.
func (f *FetcherService) fetchLoop() {
	// Initial fetch
	f.fetchAll()

	// Set up ticker for periodic fetching
	interval := f.config.GetFetchInterval()
	if interval < time.Minute {
		interval = 15 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-f.stopCh:
			return
		case <-ticker.C:
			// Check if within work hours
			if f.config.IsWithinWorkHours() {
				f.fetchAll()
			}
		}
	}
}

// fetchAll fetches issues from all configured repositories.
func (f *FetcherService) fetchAll() {
	repos := f.config.GetWatchedRepos()
	if len(repos) == 0 {
		log.Println("No repositories configured")
		return
	}

	var allIssues []*Issue
	for _, repo := range repos {
		issues, err := f.fetchFromRepo(repo)
		if err != nil {
			log.Printf("Error fetching from %s: %v", repo, err)
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	if len(allIssues) > 0 {
		select {
		case f.issuesCh <- allIssues:
			f.notify.Notify("BugSETI", fmt.Sprintf("Found %d new issues", len(allIssues)))
		default:
			// Channel full, skip
		}
	}
}

// fetchFromRepo fetches issues from a single repository using the Forgejo API.
func (f *FetcherService) fetchFromRepo(repo string) ([]*Issue, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	labels := f.config.GetLabels()
	if len(labels) == 0 {
		labels = []string{"good first issue", "help wanted", "beginner-friendly"}
	}

	forgeIssues, err := f.forge.ListIssues(owner, repoName, forge.ListIssuesOpts{
		State:  "open",
		Labels: labels,
		Limit:  20,
	})
	if err != nil {
		return nil, fmt.Errorf("forge list issues failed: %w", err)
	}

	issues := make([]*Issue, 0, len(forgeIssues))
	for _, fi := range forgeIssues {
		labelNames := make([]string, len(fi.Labels))
		for i, l := range fi.Labels {
			labelNames[i] = l.Name
		}

		author := ""
		if fi.Poster != nil {
			author = fi.Poster.UserName
		}

		issues = append(issues, &Issue{
			ID:        fmt.Sprintf("%s#%d", repo, fi.Index),
			Number:    int(fi.Index),
			Repo:      repo,
			Title:     fi.Title,
			Body:      fi.Body,
			URL:       fi.HTMLURL,
			Labels:    labelNames,
			Author:    author,
			CreatedAt: fi.Created,
			Priority:  calculatePriority(labelNames),
		})
	}

	return issues, nil
}

// FetchIssue fetches a single issue by repo and number.
func (f *FetcherService) FetchIssue(repo string, number int) (*Issue, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	fi, err := f.forge.GetIssue(owner, repoName, int64(number))
	if err != nil {
		return nil, fmt.Errorf("forge get issue failed: %w", err)
	}

	labelNames := make([]string, len(fi.Labels))
	for i, l := range fi.Labels {
		labelNames[i] = l.Name
	}

	author := ""
	if fi.Poster != nil {
		author = fi.Poster.UserName
	}

	// Fetch comments
	forgeComments, err := f.forge.ListIssueComments(owner, repoName, int64(number))
	if err != nil {
		log.Printf("Warning: could not fetch comments for %s#%d: %v", repo, number, err)
	}

	comments := make([]Comment, 0, len(forgeComments))
	for _, c := range forgeComments {
		commentAuthor := ""
		if c.Poster != nil {
			commentAuthor = c.Poster.UserName
		}
		comments = append(comments, Comment{
			Author: commentAuthor,
			Body:   c.Body,
		})
	}

	return &Issue{
		ID:        fmt.Sprintf("%s#%d", repo, fi.Index),
		Number:    int(fi.Index),
		Repo:      repo,
		Title:     fi.Title,
		Body:      fi.Body,
		URL:       fi.HTMLURL,
		Labels:    labelNames,
		Author:    author,
		CreatedAt: fi.Created,
		Priority:  calculatePriority(labelNames),
		Comments:  comments,
	}, nil
}

// splitRepo splits "owner/repo" into owner and repo parts.
func splitRepo(repo string) (string, string, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format %q, expected owner/repo", repo)
	}
	return parts[0], parts[1], nil
}

// calculatePriority assigns a priority score based on labels.
func calculatePriority(labels []string) int {
	priority := 50 // Default priority

	for _, label := range labels {
		lower := strings.ToLower(label)
		switch {
		case strings.Contains(lower, "good first issue"):
			priority += 30
		case strings.Contains(lower, "help wanted"):
			priority += 20
		case strings.Contains(lower, "beginner"):
			priority += 25
		case strings.Contains(lower, "easy"):
			priority += 20
		case strings.Contains(lower, "bug"):
			priority += 10
		case strings.Contains(lower, "documentation"):
			priority += 5
		case strings.Contains(lower, "priority"):
			priority += 15
		}
	}

	return priority
}
