package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/forge"
	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/host-uk/core/pkg/log"
)

// AgentTarget maps a Forgejo username to an SSH-reachable agent machine.
type AgentTarget struct {
	Host     string // SSH destination (e.g., "claude@192.168.0.201")
	QueueDir string // Remote queue directory (e.g., "~/ai-work/queue")
}

// DispatchTicket is the JSON payload written to the agent's queue.
type DispatchTicket struct {
	ID           string `json:"id"`
	RepoOwner    string `json:"repo_owner"`
	RepoName     string `json:"repo_name"`
	IssueNumber  int    `json:"issue_number"`
	IssueTitle   string `json:"issue_title"`
	IssueBody    string `json:"issue_body"`
	TargetBranch string `json:"target_branch"`
	EpicNumber   int    `json:"epic_number"`
	ForgeURL     string `json:"forge_url"`
	ForgeToken   string `json:"forge_token"`
	CreatedAt    string `json:"created_at"`
}

// DispatchHandler dispatches coding work to remote agent machines via SSH/SCP.
type DispatchHandler struct {
	forge    *forge.Client
	forgeURL string
	token    string
	agents   map[string]AgentTarget
}

// NewDispatchHandler creates a handler that dispatches tickets to agent machines.
func NewDispatchHandler(client *forge.Client, forgeURL, token string, agents map[string]AgentTarget) *DispatchHandler {
	return &DispatchHandler{
		forge:    client,
		forgeURL: forgeURL,
		token:    token,
		agents:   agents,
	}
}

// Name returns the handler identifier.
func (h *DispatchHandler) Name() string {
	return "dispatch"
}

// Match returns true for signals where a child issue needs coding (no PR yet)
// and the assignee is a known agent.
func (h *DispatchHandler) Match(signal *jobrunner.PipelineSignal) bool {
	if !signal.NeedsCoding {
		return false
	}
	_, ok := h.agents[signal.Assignee]
	return ok
}

// Execute creates a ticket JSON and SCPs it to the agent's queue directory.
func (h *DispatchHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	agent, ok := h.agents[signal.Assignee]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", signal.Assignee)
	}

	// Determine target branch (default to repo default).
	targetBranch := "new" // TODO: resolve from epic or repo default

	ticket := DispatchTicket{
		ID:           fmt.Sprintf("%s-%s-%d-%d", signal.RepoOwner, signal.RepoName, signal.ChildNumber, time.Now().Unix()),
		RepoOwner:    signal.RepoOwner,
		RepoName:     signal.RepoName,
		IssueNumber:  signal.ChildNumber,
		IssueTitle:   signal.IssueTitle,
		IssueBody:    signal.IssueBody,
		TargetBranch: targetBranch,
		EpicNumber:   signal.EpicNumber,
		ForgeURL:     h.forgeURL,
		ForgeToken:   h.token,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	ticketJSON, err := json.MarshalIndent(ticket, "", "  ")
	if err != nil {
		return &jobrunner.ActionResult{
			Action:      "dispatch",
			RepoOwner:   signal.RepoOwner,
			RepoName:    signal.RepoName,
			EpicNumber:  signal.EpicNumber,
			ChildNumber: signal.ChildNumber,
			Success:     false,
			Error:       fmt.Sprintf("marshal ticket: %v", err),
			Timestamp:   time.Now(),
			Duration:    time.Since(start),
		}, nil
	}

	// Check if ticket already exists on agent (dedup).
	ticketName := fmt.Sprintf("ticket-%s-%s-%d.json", signal.RepoOwner, signal.RepoName, signal.ChildNumber)
	if h.ticketExists(agent, ticketName) {
		log.Info("ticket already queued, skipping", "ticket", ticketName, "agent", signal.Assignee)
		return &jobrunner.ActionResult{
			Action:      "dispatch",
			RepoOwner:   signal.RepoOwner,
			RepoName:    signal.RepoName,
			EpicNumber:  signal.EpicNumber,
			ChildNumber: signal.ChildNumber,
			Success:     true,
			Timestamp:   time.Now(),
			Duration:    time.Since(start),
		}, nil
	}

	// SCP ticket to agent queue.
	remotePath := filepath.Join(agent.QueueDir, ticketName)
	if err := h.scpTicket(ctx, agent.Host, remotePath, ticketJSON); err != nil {
		return &jobrunner.ActionResult{
			Action:      "dispatch",
			RepoOwner:   signal.RepoOwner,
			RepoName:    signal.RepoName,
			EpicNumber:  signal.EpicNumber,
			ChildNumber: signal.ChildNumber,
			Success:     false,
			Error:       fmt.Sprintf("scp ticket: %v", err),
			Timestamp:   time.Now(),
			Duration:    time.Since(start),
		}, nil
	}

	// Comment on issue.
	comment := fmt.Sprintf("Dispatched to **%s** agent queue.", signal.Assignee)
	_ = h.forge.CreateIssueComment(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), comment)

	return &jobrunner.ActionResult{
		Action:      "dispatch",
		RepoOwner:   signal.RepoOwner,
		RepoName:    signal.RepoName,
		EpicNumber:  signal.EpicNumber,
		ChildNumber: signal.ChildNumber,
		Success:     true,
		Timestamp:   time.Now(),
		Duration:    time.Since(start),
	}, nil
}

// scpTicket writes ticket data to a remote path via SSH.
func (h *DispatchHandler) scpTicket(ctx context.Context, host, remotePath string, data []byte) error {
	// Use ssh + cat instead of scp for piping stdin.
	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		host,
		fmt.Sprintf("cat > %s", remotePath),
	)
	cmd.Stdin = strings.NewReader(string(data))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("dispatch.scp", fmt.Sprintf("ssh to %s failed: %s", host, string(output)), err)
	}
	return nil
}

// ticketExists checks if a ticket file already exists in queue, active, or done.
func (h *DispatchHandler) ticketExists(agent AgentTarget, ticketName string) bool {
	cmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		agent.Host,
		fmt.Sprintf("test -f %s/%s || test -f %s/../active/%s || test -f %s/../done/%s",
			agent.QueueDir, ticketName,
			agent.QueueDir, ticketName,
			agent.QueueDir, ticketName),
	)
	return cmd.Run() == nil
}
