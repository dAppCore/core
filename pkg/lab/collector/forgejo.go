package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"forge.lthn.ai/core/go/pkg/lab"
)

type Forgejo struct {
	url   string
	token string
	store *lab.Store
}

func NewForgejo(forgeURL, token string, s *lab.Store) *Forgejo {
	return &Forgejo{url: forgeURL, token: token, store: s}
}

func (f *Forgejo) Name() string { return "forgejo" }

func (f *Forgejo) Collect(ctx context.Context) error {
	if f.token == "" {
		return nil
	}

	commits, err := f.recentActivity(ctx)
	if err != nil {
		f.store.SetError("forgejo", err)
		return err
	}

	f.store.SetCommits(commits)
	f.store.SetError("forgejo", nil)
	return nil
}

type forgeRepo struct {
	FullName  string    `json:"full_name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type forgeCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

func (f *Forgejo) recentActivity(ctx context.Context) ([]lab.Commit, error) {
	// Get recently updated repos
	repos, err := f.apiGet(ctx, "/api/v1/repos/search?sort=updated&order=desc&limit=5")
	if err != nil {
		return nil, err
	}

	var repoList []forgeRepo
	if err := json.Unmarshal(repos, &repoList); err != nil {
		// The search API wraps in {"data": [...], "ok": true}
		var wrapped struct {
			Data []forgeRepo `json:"data"`
		}
		if err2 := json.Unmarshal(repos, &wrapped); err2 != nil {
			return nil, err
		}
		repoList = wrapped.Data
	}

	var commits []lab.Commit
	for _, repo := range repoList {
		if len(commits) >= 10 {
			break
		}
		data, err := f.apiGet(ctx, fmt.Sprintf("/api/v1/repos/%s/commits?limit=2", repo.FullName))
		if err != nil {
			continue
		}
		var fc []forgeCommit
		if err := json.Unmarshal(data, &fc); err != nil {
			continue
		}
		for _, c := range fc {
			msg := c.Commit.Message
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			commits = append(commits, lab.Commit{
				SHA:       c.SHA[:8],
				Message:   msg,
				Author:    c.Commit.Author.Name,
				Repo:      repo.FullName,
				Timestamp: c.Commit.Author.Date,
			})
		}
	}

	return commits, nil
}

func (f *Forgejo) apiGet(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.url+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+f.token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("forgejo %s returned %d", path, resp.StatusCode)
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}
