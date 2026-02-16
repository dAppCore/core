// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/cli/pkg/forge"
)

// SubmitService handles the PR submission flow.
type SubmitService struct {
	config *ConfigService
	notify *NotifyService
	stats  *StatsService
	forge  *forge.Client
}

// NewSubmitService creates a new SubmitService.
func NewSubmitService(config *ConfigService, notify *NotifyService, stats *StatsService, forgeClient *forge.Client) *SubmitService {
	return &SubmitService{
		config: config,
		notify: notify,
		stats:  stats,
		forge:  forgeClient,
	}
}

// ServiceName returns the service name for Wails.
func (s *SubmitService) ServiceName() string {
	return "SubmitService"
}

// PRSubmission contains the data for a pull request submission.
type PRSubmission struct {
	Issue     *Issue   `json:"issue"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Branch    string   `json:"branch"`
	CommitMsg string   `json:"commitMsg"`
	Files     []string `json:"files"`
	WorkDir   string   `json:"workDir"`
}

// PRResult contains the result of a PR submission.
type PRResult struct {
	Success   bool   `json:"success"`
	PRURL     string `json:"prUrl,omitempty"`
	PRNumber  int    `json:"prNumber,omitempty"`
	Error     string `json:"error,omitempty"`
	ForkOwner string `json:"forkOwner,omitempty"`
}

// Submit creates a pull request for the given issue.
// Flow: Fork -> Branch -> Commit -> Push -> PR
func (s *SubmitService) Submit(submission *PRSubmission) (*PRResult, error) {
	if submission == nil || submission.Issue == nil {
		return nil, fmt.Errorf("invalid submission")
	}

	issue := submission.Issue
	workDir := submission.WorkDir
	if workDir == "" {
		return nil, fmt.Errorf("work directory not specified")
	}

	guard := getEthicsGuardWithRoot(context.Background(), s.config.GetMarketplaceMCPRoot())
	issueTitle := guard.SanitizeTitle(issue.Title)

	owner, repoName, err := splitRepo(issue.Repo)
	if err != nil {
		return &PRResult{Success: false, Error: err.Error()}, err
	}

	// Step 1: Ensure we have a fork
	forkOwner, err := s.ensureFork(owner, repoName)
	if err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("fork failed: %v", err)}, err
	}

	// Step 2: Create branch
	branch := submission.Branch
	if branch == "" {
		branch = fmt.Sprintf("bugseti/issue-%d", issue.Number)
	}
	if err := s.createBranch(workDir, branch); err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("branch creation failed: %v", err)}, err
	}

	// Step 3: Stage and commit changes
	commitMsg := submission.CommitMsg
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("fix: resolve issue #%d\n\n%s\n\nFixes #%d", issue.Number, issueTitle, issue.Number)
	} else {
		commitMsg = guard.SanitizeBody(commitMsg)
	}
	if err := s.commitChanges(workDir, submission.Files, commitMsg); err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("commit failed: %v", err)}, err
	}

	// Step 4: Push to fork
	if err := s.pushToFork(workDir, forkOwner, repoName, branch); err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("push failed: %v", err)}, err
	}

	// Step 5: Create PR
	prTitle := submission.Title
	if prTitle == "" {
		prTitle = fmt.Sprintf("Fix #%d: %s", issue.Number, issueTitle)
	} else {
		prTitle = guard.SanitizeTitle(prTitle)
	}
	prBody := submission.Body
	if prBody == "" {
		prBody = s.generatePRBody(issue)
	}
	prBody = guard.SanitizeBody(prBody)

	prURL, prNumber, err := s.createPR(owner, repoName, forkOwner, branch, prTitle, prBody)
	if err != nil {
		return &PRResult{Success: false, Error: fmt.Sprintf("PR creation failed: %v", err)}, err
	}

	// Update stats
	s.stats.RecordPRSubmitted(issue.Repo)

	// Notify user
	s.notify.Notify("BugSETI", fmt.Sprintf("PR #%d submitted for issue #%d", prNumber, issue.Number))

	return &PRResult{
		Success:   true,
		PRURL:     prURL,
		PRNumber:  prNumber,
		ForkOwner: forkOwner,
	}, nil
}

// ensureFork ensures a fork exists for the repo, returns the fork owner's username.
func (s *SubmitService) ensureFork(owner, repo string) (string, error) {
	// Get current user
	user, err := s.forge.GetCurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	username := user.UserName

	// Check if fork already exists
	_, err = s.forge.GetRepo(username, repo)
	if err == nil {
		return username, nil
	}

	// Fork doesn't exist, create it
	log.Printf("Creating fork of %s/%s...", owner, repo)
	_, err = s.forge.ForkRepo(owner, repo, "")
	if err != nil {
		return "", fmt.Errorf("failed to create fork: %w", err)
	}

	// Wait for Forgejo to process the fork
	time.Sleep(2 * time.Second)

	return username, nil
}

// createBranch creates a new branch in the repository.
func (s *SubmitService) createBranch(workDir, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch latest from upstream
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin")
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		log.Printf("WARNING: git fetch origin failed in %s: %v (proceeding with potentially stale data)", workDir, err)
	}

	// Create and checkout new branch
	cmd = exec.CommandContext(ctx, "git", "checkout", "-b", branch)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Branch might already exist, try to checkout
		cmd = exec.CommandContext(ctx, "git", "checkout", branch)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create/checkout branch: %s: %w", stderr.String(), err)
		}
	}

	return nil
}

// commitChanges stages and commits the specified files.
func (s *SubmitService) commitChanges(workDir string, files []string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stage files
	if len(files) == 0 {
		// Stage all changes
		cmd := exec.CommandContext(ctx, "git", "add", "-A")
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}
	} else {
		// Stage specific files
		args := append([]string{"add"}, files...)
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}
	}

	// Check if there are changes to commit
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = workDir
	if err := cmd.Run(); err == nil {
		return fmt.Errorf("no changes to commit")
	}

	// Commit
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %s: %w", stderr.String(), err)
	}

	return nil
}

// pushToFork pushes the branch to the user's fork.
func (s *SubmitService) pushToFork(workDir, forkOwner, repoName, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Add fork as remote if not exists
	forkRemote := "fork"
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", forkRemote)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		// Construct fork URL using the forge instance URL
		forkURL := fmt.Sprintf("%s/%s/%s.git", strings.TrimRight(s.forge.URL(), "/"), forkOwner, repoName)

		// Embed token for HTTPS push auth
		if s.forge.Token() != "" {
			forkURL = strings.Replace(forkURL, "://", fmt.Sprintf("://bugseti:%s@", s.forge.Token()), 1)
		}

		cmd = exec.CommandContext(ctx, "git", "remote", "add", forkRemote, forkURL)
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add fork remote: %w", err)
		}
	}

	// Push to fork
	cmd = exec.CommandContext(ctx, "git", "push", "-u", forkRemote, branch)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %s: %w", stderr.String(), err)
	}

	return nil
}

// createPR creates a pull request using the Forgejo API.
func (s *SubmitService) createPR(owner, repo, forkOwner, branch, title, body string) (string, int, error) {
	pr, err := s.forge.CreatePullRequest(owner, repo, forgejo.CreatePullRequestOption{
		Head:  fmt.Sprintf("%s:%s", forkOwner, branch),
		Base:  "main",
		Title: title,
		Body:  body,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr.HTMLURL, int(pr.Index), nil
}

// generatePRBody creates a default PR body for an issue.
func (s *SubmitService) generatePRBody(issue *Issue) string {
	var body strings.Builder

	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("This PR addresses issue #%d.\n\n", issue.Number))

	if issue.Context != nil && issue.Context.Summary != "" {
		body.WriteString("## Context\n\n")
		body.WriteString(issue.Context.Summary)
		body.WriteString("\n\n")
	}

	body.WriteString("## Changes\n\n")
	body.WriteString("<!-- Describe your changes here -->\n\n")

	body.WriteString("## Testing\n\n")
	body.WriteString("<!-- Describe how you tested your changes -->\n\n")

	body.WriteString("---\n\n")
	body.WriteString("*Submitted via [BugSETI](https://bugseti.app) - Distributed Bug Fixing*\n")

	return body.String()
}

// GetPRStatus checks the status of a submitted PR.
func (s *SubmitService) GetPRStatus(repo string, prNumber int) (*PRStatus, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	pr, err := s.forge.GetPullRequest(owner, repoName, int64(prNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get PR status: %w", err)
	}

	status := &PRStatus{
		State:     string(pr.State),
		Mergeable: pr.Mergeable,
	}

	// Check CI status via combined commit status
	if pr.Head != nil {
		combined, err := s.forge.GetCombinedStatus(owner, repoName, pr.Head.Sha)
		if err == nil && combined != nil {
			status.CIPassing = combined.State == forgejo.StatusSuccess
		}
	}

	// Check review status
	reviews, err := s.forge.ListPRReviews(owner, repoName, int64(prNumber))
	if err == nil {
		for _, review := range reviews {
			if review.State == forgejo.ReviewStateApproved {
				status.Approved = true
				break
			}
		}
	}

	return status, nil
}

// PRStatus represents the current status of a PR.
type PRStatus struct {
	State     string `json:"state"`
	Mergeable bool   `json:"mergeable"`
	CIPassing bool   `json:"ciPassing"`
	Approved  bool   `json:"approved"`
}
