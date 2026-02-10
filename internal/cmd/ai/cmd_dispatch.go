package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/log"
)

// AddDispatchCommands registers the 'dispatch' subcommand group under 'ai'.
// These commands run ON the agent machine to process the work queue.
func AddDispatchCommands(parent *cli.Command) {
	dispatchCmd := &cli.Command{
		Use:   "dispatch",
		Short: "Agent work queue processor (runs on agent machine)",
	}

	dispatchCmd.AddCommand(dispatchRunCmd())
	dispatchCmd.AddCommand(dispatchWatchCmd())
	dispatchCmd.AddCommand(dispatchStatusCmd())

	parent.AddCommand(dispatchCmd)
}

// dispatchTicket represents the work item JSON structure.
type dispatchTicket struct {
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
	ForgeUser    string `json:"forgejo_user"`
	Model        string `json:"model"`
	Runner       string `json:"runner"`
	Timeout      string `json:"timeout"`
	CreatedAt    string `json:"created_at"`
}

const (
	defaultWorkDir = "ai-work"
	lockFileName   = ".runner.lock"
)

type runnerPaths struct {
	root   string
	queue  string
	active string
	done   string
	logs   string
	jobs   string
	lock   string
}

func getPaths(baseDir string) runnerPaths {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, defaultWorkDir)
	}
	return runnerPaths{
		root:   baseDir,
		queue:  filepath.Join(baseDir, "queue"),
		active: filepath.Join(baseDir, "active"),
		done:   filepath.Join(baseDir, "done"),
		logs:   filepath.Join(baseDir, "logs"),
		jobs:   filepath.Join(baseDir, "jobs"),
		lock:   filepath.Join(baseDir, lockFileName),
	}
}

func dispatchRunCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "run",
		Short: "Process a single ticket from the queue",
		RunE: func(cmd *cli.Command, args []string) error {
			workDir, _ := cmd.Flags().GetString("work-dir")
			paths := getPaths(workDir)

			if err := ensureDispatchDirs(paths); err != nil {
				return err
			}

			if err := acquireLock(paths.lock); err != nil {
				log.Info("Runner locked, skipping run", "lock", paths.lock)
				return nil
			}
			defer releaseLock(paths.lock)

			ticketFile, err := pickOldestTicket(paths.queue)
			if err != nil {
				return err
			}
			if ticketFile == "" {
				return nil
			}

			return processTicket(paths, ticketFile)
		},
	}
	cmd.Flags().String("work-dir", "", "Working directory (default: ~/ai-work)")
	return cmd
}

func dispatchWatchCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "watch",
		Short: "Run as a daemon, polling the queue",
		RunE: func(cmd *cli.Command, args []string) error {
			workDir, _ := cmd.Flags().GetString("work-dir")
			interval, _ := cmd.Flags().GetDuration("interval")
			paths := getPaths(workDir)

			if err := ensureDispatchDirs(paths); err != nil {
				return err
			}

			log.Info("Starting dispatch watcher", "dir", paths.root, "interval", interval)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			runCycle(paths)

			for {
				select {
				case <-ticker.C:
					runCycle(paths)
				case <-sigChan:
					log.Info("Shutting down watcher...")
					return nil
				case <-ctx.Done():
					return nil
				}
			}
		},
	}
	cmd.Flags().String("work-dir", "", "Working directory (default: ~/ai-work)")
	cmd.Flags().Duration("interval", 5*time.Minute, "Polling interval")
	return cmd
}

func dispatchStatusCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "status",
		Short: "Show runner status",
		RunE: func(cmd *cli.Command, args []string) error {
			workDir, _ := cmd.Flags().GetString("work-dir")
			paths := getPaths(workDir)

			lockStatus := "IDLE"
			if data, err := os.ReadFile(paths.lock); err == nil {
				pidStr := strings.TrimSpace(string(data))
				pid, _ := strconv.Atoi(pidStr)
				if isProcessAlive(pid) {
					lockStatus = fmt.Sprintf("RUNNING (PID %d)", pid)
				} else {
					lockStatus = fmt.Sprintf("STALE (PID %d)", pid)
				}
			}

			countFiles := func(dir string) int {
				entries, _ := os.ReadDir(dir)
				count := 0
				for _, e := range entries {
					if !e.IsDir() && strings.HasPrefix(e.Name(), "ticket-") {
						count++
					}
				}
				return count
			}

			fmt.Println("=== Agent Dispatch Status ===")
			fmt.Printf("Work Dir: %s\n", paths.root)
			fmt.Printf("Status:   %s\n", lockStatus)
			fmt.Printf("Queue:    %d\n", countFiles(paths.queue))
			fmt.Printf("Active:   %d\n", countFiles(paths.active))
			fmt.Printf("Done:     %d\n", countFiles(paths.done))

			return nil
		},
	}
	cmd.Flags().String("work-dir", "", "Working directory (default: ~/ai-work)")
	return cmd
}

func runCycle(paths runnerPaths) {
	if err := acquireLock(paths.lock); err != nil {
		log.Debug("Runner locked, skipping cycle")
		return
	}
	defer releaseLock(paths.lock)

	ticketFile, err := pickOldestTicket(paths.queue)
	if err != nil {
		log.Error("Failed to pick ticket", "error", err)
		return
	}
	if ticketFile == "" {
		return
	}

	if err := processTicket(paths, ticketFile); err != nil {
		log.Error("Failed to process ticket", "file", ticketFile, "error", err)
	}
}

func processTicket(paths runnerPaths, ticketPath string) error {
	fileName := filepath.Base(ticketPath)
	log.Info("Processing ticket", "file", fileName)

	activePath := filepath.Join(paths.active, fileName)
	if err := os.Rename(ticketPath, activePath); err != nil {
		return fmt.Errorf("failed to move ticket to active: %w", err)
	}

	data, err := os.ReadFile(activePath)
	if err != nil {
		return fmt.Errorf("failed to read ticket: %w", err)
	}
	var t dispatchTicket
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("failed to unmarshal ticket: %w", err)
	}

	jobDir := filepath.Join(paths.jobs, fmt.Sprintf("%s-%s-%d", t.RepoOwner, t.RepoName, t.IssueNumber))
	repoDir := filepath.Join(jobDir, t.RepoName)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return err
	}

	if err := prepareRepo(t, repoDir); err != nil {
		reportToForge(t, false, fmt.Sprintf("Git setup failed: %v", err))
		moveToDone(paths, activePath, fileName)
		return err
	}

	prompt := buildPrompt(t)

	logFile := filepath.Join(paths.logs, fmt.Sprintf("%s-%s-%d.log", t.RepoOwner, t.RepoName, t.IssueNumber))
	success, exitCode, runErr := runAgent(t, prompt, repoDir, logFile)

	msg := fmt.Sprintf("Agent completed work on #%d. Exit code: %d.", t.IssueNumber, exitCode)
	if !success {
		msg = fmt.Sprintf("Agent failed on #%d (exit code: %d). Check logs on agent machine.", t.IssueNumber, exitCode)
		if runErr != nil {
			msg += fmt.Sprintf(" Error: %v", runErr)
		}
	}
	reportToForge(t, success, msg)

	moveToDone(paths, activePath, fileName)
	log.Info("Ticket complete", "id", t.ID, "success", success)
	return nil
}

func prepareRepo(t dispatchTicket, repoDir string) error {
	user := t.ForgeUser
	if user == "" {
		host, _ := os.Hostname()
		user = fmt.Sprintf("%s-%s", host, os.Getenv("USER"))
	}

	cleanURL := strings.TrimPrefix(t.ForgeURL, "https://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")
	cloneURL := fmt.Sprintf("https://%s:%s@%s/%s/%s.git", user, t.ForgeToken, cleanURL, t.RepoOwner, t.RepoName)

	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		log.Info("Updating existing repo", "dir", repoDir)
		cmds := [][]string{
			{"git", "fetch", "origin"},
			{"git", "checkout", t.TargetBranch},
			{"git", "pull", "origin", t.TargetBranch},
		}
		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = repoDir
			if out, err := cmd.CombinedOutput(); err != nil {
				if args[1] == "checkout" {
					createCmd := exec.Command("git", "checkout", "-b", t.TargetBranch, "origin/"+t.TargetBranch)
					createCmd.Dir = repoDir
					if _, err2 := createCmd.CombinedOutput(); err2 == nil {
						continue
					}
				}
				return fmt.Errorf("git command %v failed: %s", args, string(out))
			}
		}
	} else {
		log.Info("Cloning repo", "url", t.RepoOwner+"/"+t.RepoName)
		cmd := exec.Command("git", "clone", "-b", t.TargetBranch, cloneURL, repoDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone failed: %s", string(out))
		}
	}
	return nil
}

func buildPrompt(t dispatchTicket) string {
	return fmt.Sprintf(`You are working on issue #%d in %s/%s.

Title: %s

Description:
%s

The repo is cloned at the current directory on branch '%s'.
Create a feature branch from '%s', make minimal targeted changes, commit referencing #%d, and push.
Then create a PR targeting '%s' using the forgejo MCP tools or git push.`,
		t.IssueNumber, t.RepoOwner, t.RepoName,
		t.IssueTitle,
		t.IssueBody,
		t.TargetBranch,
		t.TargetBranch, t.IssueNumber,
		t.TargetBranch,
	)
}

func runAgent(t dispatchTicket, prompt, dir, logPath string) (bool, int, error) {
	timeout := 30 * time.Minute
	if t.Timeout != "" {
		if d, err := time.ParseDuration(t.Timeout); err == nil {
			timeout = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	model := t.Model
	if model == "" {
		model = "sonnet"
	}

	log.Info("Running agent", "runner", t.Runner, "model", model)

	// For Gemini runner, wrap with rate limiting.
	if t.Runner == "gemini" {
		return executeWithRateLimit(ctx, model, prompt, func() (bool, int, error) {
			return execAgent(ctx, t.Runner, model, prompt, dir, logPath)
		})
	}

	return execAgent(ctx, t.Runner, model, prompt, dir, logPath)
}

func execAgent(ctx context.Context, runner, model, prompt, dir, logPath string) (bool, int, error) {
	var cmd *exec.Cmd

	switch runner {
	case "codex":
		cmd = exec.CommandContext(ctx, "codex", "exec", "--full-auto", prompt)
	case "gemini":
		args := []string{"-p", "-", "-y", "-m", model}
		cmd = exec.CommandContext(ctx, "gemini", args...)
		cmd.Stdin = strings.NewReader(prompt)
	default: // claude
		cmd = exec.CommandContext(ctx, "claude", "-p", "--model", model, "--dangerously-skip-permissions", "--output-format", "text")
		cmd.Stdin = strings.NewReader(prompt)
	}

	cmd.Dir = dir

	f, err := os.Create(logPath)
	if err != nil {
		return false, -1, err
	}
	defer f.Close()

	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Run(); err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return false, exitCode, err
	}

	return true, 0, nil
}

func reportToForge(t dispatchTicket, success bool, body string) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/issues/%d/comments",
		strings.TrimSuffix(t.ForgeURL, "/"), t.RepoOwner, t.RepoName, t.IssueNumber)

	payload := map[string]string{"body": body}
	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Error("Failed to create request", "err", err)
		return
	}
	req.Header.Set("Authorization", "token "+t.ForgeToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to report to Forge", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Warn("Forge reported error", "status", resp.Status)
	}
}

func moveToDone(paths runnerPaths, activePath, fileName string) {
	donePath := filepath.Join(paths.done, fileName)
	if err := os.Rename(activePath, donePath); err != nil {
		log.Error("Failed to move ticket to done", "err", err)
	}
}

func ensureDispatchDirs(p runnerPaths) error {
	dirs := []string{p.queue, p.active, p.done, p.logs, p.jobs}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("mkdir %s failed: %w", d, err)
		}
	}
	return nil
}

func acquireLock(lockPath string) error {
	if data, err := os.ReadFile(lockPath); err == nil {
		pidStr := strings.TrimSpace(string(data))
		pid, _ := strconv.Atoi(pidStr)
		if isProcessAlive(pid) {
			return fmt.Errorf("locked by PID %d", pid)
		}
		log.Info("Removing stale lock", "pid", pid)
		_ = os.Remove(lockPath)
	}

	return os.WriteFile(lockPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func releaseLock(lockPath string) {
	_ = os.Remove(lockPath)
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func pickOldestTicket(queueDir string) (string, error) {
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		return "", err
	}

	var tickets []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "ticket-") && strings.HasSuffix(e.Name(), ".json") {
			tickets = append(tickets, filepath.Join(queueDir, e.Name()))
		}
	}

	if len(tickets) == 0 {
		return "", nil
	}

	sort.Strings(tickets)
	return tickets[0], nil
}
