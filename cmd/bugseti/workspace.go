// Package main provides the BugSETI system tray application.
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Snider/Borg/pkg/tim"
	"github.com/host-uk/core/internal/bugseti"
	"github.com/host-uk/core/pkg/io/datanode"
)

// WorkspaceService manages DataNode-backed workspaces for issues.
// Each issue gets a sandboxed in-memory filesystem that can be
// snapshotted, packaged as a TIM container, or shipped as a crash report.
type WorkspaceService struct {
	config     *bugseti.ConfigService
	workspaces map[string]*Workspace // issue ID → workspace
	mu         sync.RWMutex
}

// Workspace tracks a DataNode-backed workspace for an issue.
type Workspace struct {
	Issue     *bugseti.Issue `json:"issue"`
	Medium    *datanode.Medium
	DiskPath  string    `json:"diskPath"`
	CreatedAt time.Time `json:"createdAt"`
	Snapshots int       `json:"snapshots"`
}

// CrashReport contains a packaged workspace state for debugging.
type CrashReport struct {
	IssueID   string    `json:"issueId"`
	Repo      string    `json:"repo"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"` // tar snapshot
	Files     int       `json:"files"`
	Size      int64     `json:"size"`
}

// NewWorkspaceService creates a new WorkspaceService.
func NewWorkspaceService(config *bugseti.ConfigService) *WorkspaceService {
	return &WorkspaceService{
		config:     config,
		workspaces: make(map[string]*Workspace),
	}
}

// ServiceName returns the service name for Wails.
func (w *WorkspaceService) ServiceName() string {
	return "WorkspaceService"
}

// Capture loads a filesystem workspace into a DataNode Medium.
// Call this after git clone to create the in-memory snapshot.
func (w *WorkspaceService) Capture(issue *bugseti.Issue, diskPath string) error {
	if issue == nil {
		return fmt.Errorf("issue is nil")
	}

	m := datanode.New()

	// Walk the filesystem and load all files into the DataNode
	err := filepath.WalkDir(diskPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}

		// Get relative path
		rel, err := filepath.Rel(diskPath, path)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}

		// Skip .git internals (keep .git marker but not the pack files)
		if rel == ".git" {
			return fs.SkipDir
		}

		if d.IsDir() {
			return m.EnsureDir(rel)
		}

		// Skip large files (>1MB) to keep DataNode lightweight
		info, err := d.Info()
		if err != nil || info.Size() > 1<<20 {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		return m.Write(rel, string(content))
	})
	if err != nil {
		return fmt.Errorf("failed to capture workspace: %w", err)
	}

	w.mu.Lock()
	w.workspaces[issue.ID] = &Workspace{
		Issue:     issue,
		Medium:    m,
		DiskPath:  diskPath,
		CreatedAt: time.Now(),
	}
	w.mu.Unlock()

	log.Printf("Captured workspace for issue #%d (%s)", issue.Number, issue.Repo)
	return nil
}

// GetMedium returns the DataNode Medium for an issue's workspace.
func (w *WorkspaceService) GetMedium(issueID string) *datanode.Medium {
	w.mu.RLock()
	defer w.mu.RUnlock()

	ws := w.workspaces[issueID]
	if ws == nil {
		return nil
	}
	return ws.Medium
}

// Snapshot takes a tar snapshot of the workspace.
func (w *WorkspaceService) Snapshot(issueID string) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	ws := w.workspaces[issueID]
	if ws == nil {
		return nil, fmt.Errorf("workspace not found: %s", issueID)
	}

	data, err := ws.Medium.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("snapshot failed: %w", err)
	}

	ws.Snapshots++
	return data, nil
}

// PackageCrashReport captures the current workspace state as a crash report.
// Re-reads from disk to get the latest state (including git changes).
func (w *WorkspaceService) PackageCrashReport(issue *bugseti.Issue, errMsg string) (*CrashReport, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue is nil")
	}

	w.mu.RLock()
	ws := w.workspaces[issue.ID]
	w.mu.RUnlock()

	var diskPath string
	if ws != nil {
		diskPath = ws.DiskPath
	} else {
		// Try to find the workspace on disk
		baseDir := w.config.GetWorkspaceDir()
		if baseDir == "" {
			baseDir = filepath.Join(os.TempDir(), "bugseti")
		}
		diskPath = filepath.Join(baseDir, sanitizeForPath(issue.Repo), fmt.Sprintf("issue-%d", issue.Number))
	}

	// Re-capture from disk to get latest state
	if err := w.Capture(issue, diskPath); err != nil {
		return nil, fmt.Errorf("capture failed: %w", err)
	}

	// Snapshot the captured workspace
	data, err := w.Snapshot(issue.ID)
	if err != nil {
		return nil, fmt.Errorf("snapshot failed: %w", err)
	}

	return &CrashReport{
		IssueID:   issue.ID,
		Repo:      issue.Repo,
		Number:    issue.Number,
		Title:     issue.Title,
		Error:     errMsg,
		Timestamp: time.Now(),
		Data:      data,
		Size:      int64(len(data)),
	}, nil
}

// PackageTIM wraps the workspace as a TIM container (runc-compatible bundle).
// The resulting TIM can be executed via runc or encrypted to .stim for transit.
func (w *WorkspaceService) PackageTIM(issueID string) (*tim.TerminalIsolationMatrix, error) {
	w.mu.RLock()
	ws := w.workspaces[issueID]
	w.mu.RUnlock()

	if ws == nil {
		return nil, fmt.Errorf("workspace not found: %s", issueID)
	}

	dn := ws.Medium.DataNode()
	return tim.FromDataNode(dn)
}

// SaveCrashReport writes a crash report to the data directory.
func (w *WorkspaceService) SaveCrashReport(report *CrashReport) (string, error) {
	dataDir := w.config.GetDataDir()
	if dataDir == "" {
		dataDir = filepath.Join(os.TempDir(), "bugseti")
	}

	crashDir := filepath.Join(dataDir, "crash-reports")
	if err := os.MkdirAll(crashDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create crash dir: %w", err)
	}

	filename := fmt.Sprintf("crash-%s-issue-%d-%s.tar",
		sanitizeForPath(report.Repo),
		report.Number,
		report.Timestamp.Format("20060102-150405"),
	)
	path := filepath.Join(crashDir, filename)

	if err := os.WriteFile(path, report.Data, 0644); err != nil {
		return "", fmt.Errorf("failed to write crash report: %w", err)
	}

	log.Printf("Crash report saved: %s (%d bytes)", path, report.Size)
	return path, nil
}

// Release removes a workspace from memory.
func (w *WorkspaceService) Release(issueID string) {
	w.mu.Lock()
	delete(w.workspaces, issueID)
	w.mu.Unlock()
}

// ActiveWorkspaces returns the count of active workspaces.
func (w *WorkspaceService) ActiveWorkspaces() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.workspaces)
}

// sanitizeForPath converts owner/repo to a safe directory name.
func sanitizeForPath(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range s {
		if c == '/' || c == '\\' || c == ':' {
			result = append(result, '-')
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
