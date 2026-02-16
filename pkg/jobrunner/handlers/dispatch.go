package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"forge.lthn.ai/core/cli/pkg/agentci"
	"forge.lthn.ai/core/cli/pkg/forge"
	"forge.lthn.ai/core/cli/pkg/jobrunner"
	"forge.lthn.ai/core/cli/pkg/log"
)

const (
	LabelAgentReady    = "agent-ready"
	LabelInProgress    = "in-progress"
	LabelAgentFailed   = "agent-failed"
	LabelAgentComplete = "agent-completed"

	ColorInProgress  = "#1d76db" // Blue
	ColorAgentFailed = "#c0392b" // Red
)

// DispatchTicket is the JSON payload written to the agent's queue.
// The ForgeToken is transferred separately via a .env file with 0600 permissions.
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
	ForgeUser    string `json:"forgejo_user"`
	Model        string `json:"model,omitempty"`
	Runner       string `json:"runner,omitempty"`
	VerifyModel  string `json:"verify_model,omitempty"`
	DualRun      bool   `json:"dual_run"`
	CreatedAt    string `json:"created_at"`
}

// DispatchHandler dispatches coding work to remote agent machines via SSH.
type DispatchHandler struct {
	forge    *forge.Client
	forgeURL string
	token    string
	spinner  *agentci.Spinner
}

// NewDispatchHandler creates a handler that dispatches tickets to agent machines.
func NewDispatchHandler(client *forge.Client, forgeURL, token string, spinner *agentci.Spinner) *DispatchHandler {
	return &DispatchHandler{
		forge:    client,
		forgeURL: forgeURL,
		token:    token,
		spinner:  spinner,
	}
}

// Name returns the handler identifier.
func (h *DispatchHandler) Name() string {
	return "dispatch"
}

// Match returns true for signals where a child issue needs coding (no PR yet)
// and the assignee is a known agent (by config key or Forgejo username).
func (h *DispatchHandler) Match(signal *jobrunner.PipelineSignal) bool {
	if !signal.NeedsCoding {
		return false
	}
	_, _, ok := h.spinner.FindByForgejoUser(signal.Assignee)
	return ok
}

// Execute creates a ticket JSON and transfers it securely to the agent's queue directory.
func (h *DispatchHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	agentName, agent, ok := h.spinner.FindByForgejoUser(signal.Assignee)
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", signal.Assignee)
	}

	// Sanitize inputs to prevent path traversal.
	safeOwner, err := agentci.SanitizePath(signal.RepoOwner)
	if err != nil {
		return nil, fmt.Errorf("invalid repo owner: %w", err)
	}
	safeRepo, err := agentci.SanitizePath(signal.RepoName)
	if err != nil {
		return nil, fmt.Errorf("invalid repo name: %w", err)
	}

	// Ensure in-progress label exists on repo.
	inProgressLabel, err := h.forge.EnsureLabel(safeOwner, safeRepo, LabelInProgress, ColorInProgress)
	if err != nil {
		return nil, fmt.Errorf("ensure label %s: %w", LabelInProgress, err)
	}

	// Check if already in progress to prevent double-dispatch.
	issue, err := h.forge.GetIssue(safeOwner, safeRepo, int64(signal.ChildNumber))
	if err == nil {
		for _, l := range issue.Labels {
			if l.Name == LabelInProgress || l.Name == LabelAgentComplete {
				log.Info("issue already processed, skipping", "issue", signal.ChildNumber, "label", l.Name)
				return &jobrunner.ActionResult{
					Action:    "dispatch",
					Success:   true,
					Timestamp: time.Now(),
					Duration:  time.Since(start),
				}, nil
			}
		}
	}

	// Assign agent and add in-progress label.
	if err := h.forge.AssignIssue(safeOwner, safeRepo, int64(signal.ChildNumber), []string{signal.Assignee}); err != nil {
		log.Warn("failed to assign agent, continuing", "err", err)
	}

	if err := h.forge.AddIssueLabels(safeOwner, safeRepo, int64(signal.ChildNumber), []int64{inProgressLabel.ID}); err != nil {
		return nil, fmt.Errorf("add in-progress label: %w", err)
	}

	// Remove agent-ready label if present.
	if readyLabel, err := h.forge.GetLabelByName(safeOwner, safeRepo, LabelAgentReady); err == nil {
		_ = h.forge.RemoveIssueLabel(safeOwner, safeRepo, int64(signal.ChildNumber), readyLabel.ID)
	}

	// Clotho planning — determine execution mode.
	runMode := h.spinner.DeterminePlan(signal, agentName)
	verifyModel := ""
	if runMode == agentci.ModeDual {
		verifyModel = h.spinner.GetVerifierModel(agentName)
	}

	// Build ticket.
	targetBranch := "new" // TODO: resolve from epic or repo default
	ticketID := fmt.Sprintf("%s-%s-%d-%d", safeOwner, safeRepo, signal.ChildNumber, time.Now().Unix())

	ticket := DispatchTicket{
		ID:           ticketID,
		RepoOwner:    safeOwner,
		RepoName:     safeRepo,
		IssueNumber:  signal.ChildNumber,
		IssueTitle:   signal.IssueTitle,
		IssueBody:    signal.IssueBody,
		TargetBranch: targetBranch,
		EpicNumber:   signal.EpicNumber,
		ForgeURL:     h.forgeURL,
		ForgeUser:    signal.Assignee,
		Model:        agent.Model,
		Runner:       agent.Runner,
		VerifyModel:  verifyModel,
		DualRun:      runMode == agentci.ModeDual,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	ticketJSON, err := json.MarshalIndent(ticket, "", "  ")
	if err != nil {
		h.failDispatch(signal, "Failed to marshal ticket JSON")
		return nil, fmt.Errorf("marshal ticket: %w", err)
	}

	// Check if ticket already exists on agent (dedup).
	ticketName := fmt.Sprintf("ticket-%s-%s-%d.json", safeOwner, safeRepo, signal.ChildNumber)
	if h.ticketExists(ctx, agent, ticketName) {
		log.Info("ticket already queued, skipping", "ticket", ticketName, "agent", signal.Assignee)
		return &jobrunner.ActionResult{
			Action:      "dispatch",
			RepoOwner:   safeOwner,
			RepoName:    safeRepo,
			EpicNumber:  signal.EpicNumber,
			ChildNumber: signal.ChildNumber,
			Success:     true,
			Timestamp:   time.Now(),
			Duration:    time.Since(start),
		}, nil
	}

	// Transfer ticket JSON.
	remoteTicketPath := filepath.Join(agent.QueueDir, ticketName)
	if err := h.secureTransfer(ctx, agent, remoteTicketPath, ticketJSON, 0644); err != nil {
		h.failDispatch(signal, fmt.Sprintf("Ticket transfer failed: %v", err))
		return &jobrunner.ActionResult{
			Action:      "dispatch",
			RepoOwner:   safeOwner,
			RepoName:    safeRepo,
			EpicNumber:  signal.EpicNumber,
			ChildNumber: signal.ChildNumber,
			Success:     false,
			Error:       fmt.Sprintf("transfer ticket: %v", err),
			Timestamp:   time.Now(),
			Duration:    time.Since(start),
		}, nil
	}

	// Transfer token via separate .env file with 0600 permissions.
	envContent := fmt.Sprintf("FORGE_TOKEN=%s\n", h.token)
	remoteEnvPath := filepath.Join(agent.QueueDir, fmt.Sprintf(".env.%s", ticketID))
	if err := h.secureTransfer(ctx, agent, remoteEnvPath, []byte(envContent), 0600); err != nil {
		// Clean up the ticket if env transfer fails.
		_ = h.runRemote(ctx, agent, fmt.Sprintf("rm -f %s", agentci.EscapeShellArg(remoteTicketPath)))
		h.failDispatch(signal, fmt.Sprintf("Token transfer failed: %v", err))
		return &jobrunner.ActionResult{
			Action:      "dispatch",
			RepoOwner:   safeOwner,
			RepoName:    safeRepo,
			EpicNumber:  signal.EpicNumber,
			ChildNumber: signal.ChildNumber,
			Success:     false,
			Error:       fmt.Sprintf("transfer token: %v", err),
			Timestamp:   time.Now(),
			Duration:    time.Since(start),
		}, nil
	}

	// Comment on issue.
	modeStr := "Standard"
	if runMode == agentci.ModeDual {
		modeStr = "Clotho Verified (Dual Run)"
	}
	comment := fmt.Sprintf("Dispatched to **%s** agent queue.\nMode: **%s**", signal.Assignee, modeStr)
	_ = h.forge.CreateIssueComment(safeOwner, safeRepo, int64(signal.ChildNumber), comment)

	return &jobrunner.ActionResult{
		Action:      "dispatch",
		RepoOwner:   safeOwner,
		RepoName:    safeRepo,
		EpicNumber:  signal.EpicNumber,
		ChildNumber: signal.ChildNumber,
		Success:     true,
		Timestamp:   time.Now(),
		Duration:    time.Since(start),
	}, nil
}

// failDispatch handles cleanup when dispatch fails (adds failed label, removes in-progress).
func (h *DispatchHandler) failDispatch(signal *jobrunner.PipelineSignal, reason string) {
	if failedLabel, err := h.forge.EnsureLabel(signal.RepoOwner, signal.RepoName, LabelAgentFailed, ColorAgentFailed); err == nil {
		_ = h.forge.AddIssueLabels(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), []int64{failedLabel.ID})
	}

	if inProgressLabel, err := h.forge.GetLabelByName(signal.RepoOwner, signal.RepoName, LabelInProgress); err == nil {
		_ = h.forge.RemoveIssueLabel(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), inProgressLabel.ID)
	}

	_ = h.forge.CreateIssueComment(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), fmt.Sprintf("Agent dispatch failed: %s", reason))
}

// secureTransfer writes data to a remote path via SSH stdin, preventing command injection.
func (h *DispatchHandler) secureTransfer(ctx context.Context, agent agentci.AgentConfig, remotePath string, data []byte, mode int) error {
	safeRemotePath := agentci.EscapeShellArg(remotePath)
	remoteCmd := fmt.Sprintf("cat > %s && chmod %o %s", safeRemotePath, mode, safeRemotePath)

	cmd := agentci.SecureSSHCommand(agent.Host, remoteCmd)
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("dispatch.transfer", fmt.Sprintf("ssh to %s failed: %s", agent.Host, string(output)), err)
	}
	return nil
}

// runRemote executes a command on the agent via SSH.
func (h *DispatchHandler) runRemote(ctx context.Context, agent agentci.AgentConfig, cmdStr string) error {
	cmd := agentci.SecureSSHCommand(agent.Host, cmdStr)
	return cmd.Run()
}

// ticketExists checks if a ticket file already exists in queue, active, or done.
func (h *DispatchHandler) ticketExists(ctx context.Context, agent agentci.AgentConfig, ticketName string) bool {
	safeTicket, err := agentci.SanitizePath(ticketName)
	if err != nil {
		return false
	}
	qDir := agent.QueueDir
	checkCmd := fmt.Sprintf(
		"test -f %s/%s || test -f %s/../active/%s || test -f %s/../done/%s",
		qDir, safeTicket, qDir, safeTicket, qDir, safeTicket,
	)
	cmd := agentci.SecureSSHCommand(agent.Host, checkCmd)
	return cmd.Run() == nil
}
