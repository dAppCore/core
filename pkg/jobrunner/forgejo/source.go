package forgejo

import (
	"context"
	"fmt"
	"strings"

	"forge.lthn.ai/core/go/pkg/forge"
	"forge.lthn.ai/core/go/pkg/jobrunner"
	"forge.lthn.ai/core/go/pkg/log"
)

// Config configures a ForgejoSource.
type Config struct {
	Repos []string // "owner/repo" format
}

// ForgejoSource polls a Forgejo instance for pipeline signals from epic issues.
type ForgejoSource struct {
	repos []string
	forge *forge.Client
}

// New creates a ForgejoSource using the given forge client.
func New(cfg Config, client *forge.Client) *ForgejoSource {
	return &ForgejoSource{
		repos: cfg.Repos,
		forge: client,
	}
}

// Name returns the source identifier.
func (s *ForgejoSource) Name() string {
	return "forgejo"
}

// Poll fetches epics and their linked PRs from all configured repositories,
// returning a PipelineSignal for each unchecked child that has a linked PR.
func (s *ForgejoSource) Poll(ctx context.Context) ([]*jobrunner.PipelineSignal, error) {
	var signals []*jobrunner.PipelineSignal

	for _, repoFull := range s.repos {
		owner, repo, err := splitRepo(repoFull)
		if err != nil {
			log.Error("invalid repo format", "repo", repoFull, "err", err)
			continue
		}

		repoSignals, err := s.pollRepo(ctx, owner, repo)
		if err != nil {
			log.Error("poll repo failed", "repo", repoFull, "err", err)
			continue
		}

		signals = append(signals, repoSignals...)
	}

	return signals, nil
}

// Report posts the action result as a comment on the epic issue.
func (s *ForgejoSource) Report(ctx context.Context, result *jobrunner.ActionResult) error {
	if result == nil {
		return nil
	}

	status := "succeeded"
	if !result.Success {
		status = "failed"
	}

	body := fmt.Sprintf("**jobrunner** `%s` %s for #%d (PR #%d)", result.Action, status, result.ChildNumber, result.PRNumber)
	if result.Error != "" {
		body += fmt.Sprintf("\n\n```\n%s\n```", result.Error)
	}

	return s.forge.CreateIssueComment(result.RepoOwner, result.RepoName, int64(result.EpicNumber), body)
}

// pollRepo fetches epics and PRs for a single repository.
func (s *ForgejoSource) pollRepo(_ context.Context, owner, repo string) ([]*jobrunner.PipelineSignal, error) {
	// Fetch epic issues (label=epic, state=open).
	issues, err := s.forge.ListIssues(owner, repo, forge.ListIssuesOpts{State: "open"})
	if err != nil {
		return nil, log.E("forgejo.pollRepo", "fetch issues", err)
	}

	// Filter to epics only.
	var epics []epicInfo
	for _, issue := range issues {
		for _, label := range issue.Labels {
			if label.Name == "epic" {
				epics = append(epics, epicInfo{
					Number: int(issue.Index),
					Body:   issue.Body,
				})
				break
			}
		}
	}

	if len(epics) == 0 {
		return nil, nil
	}

	// Fetch all open PRs (and also merged/closed to catch MERGED state).
	prs, err := s.forge.ListPullRequests(owner, repo, "all")
	if err != nil {
		return nil, log.E("forgejo.pollRepo", "fetch PRs", err)
	}

	var signals []*jobrunner.PipelineSignal

	for _, epic := range epics {
		unchecked, _ := parseEpicChildren(epic.Body)
		for _, childNum := range unchecked {
			pr := findLinkedPR(prs, childNum)

			if pr == nil {
				// No PR yet — check if the child issue is assigned (needs coding).
				childIssue, err := s.forge.GetIssue(owner, repo, int64(childNum))
				if err != nil {
					log.Error("fetch child issue failed", "repo", owner+"/"+repo, "issue", childNum, "err", err)
					continue
				}
				if len(childIssue.Assignees) > 0 && childIssue.Assignees[0].UserName != "" {
					sig := &jobrunner.PipelineSignal{
						EpicNumber:  epic.Number,
						ChildNumber: childNum,
						RepoOwner:   owner,
						RepoName:    repo,
						NeedsCoding: true,
						Assignee:    childIssue.Assignees[0].UserName,
						IssueTitle:  childIssue.Title,
						IssueBody:   childIssue.Body,
					}
					signals = append(signals, sig)
				}
				continue
			}

			// Get combined commit status for the PR's head SHA.
			checkStatus := "PENDING"
			if pr.Head != nil && pr.Head.Sha != "" {
				cs, err := s.forge.GetCombinedStatus(owner, repo, pr.Head.Sha)
				if err != nil {
					log.Error("fetch combined status failed", "repo", owner+"/"+repo, "sha", pr.Head.Sha, "err", err)
				} else {
					checkStatus = mapCombinedStatus(cs)
				}
			}

			sig := buildSignal(owner, repo, epic.Number, childNum, pr, checkStatus)
			signals = append(signals, sig)
		}
	}

	return signals, nil
}

type epicInfo struct {
	Number int
	Body   string
}

// splitRepo parses "owner/repo" into its components.
func splitRepo(full string) (string, string, error) {
	parts := strings.SplitN(full, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", log.E("forgejo.splitRepo", fmt.Sprintf("expected owner/repo format, got %q", full), nil)
	}
	return parts[0], parts[1], nil
}
