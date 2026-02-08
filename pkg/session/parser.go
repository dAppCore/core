package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Event represents a single action in a session timeline.
type Event struct {
	Timestamp time.Time
	Type      string // "tool_use", "user", "assistant", "error"
	Tool      string // "Bash", "Read", "Edit", "Write", "Grep", "Glob", etc.
	ToolID    string
	Input     string // Command, file path, or message text
	Output    string // Result text
	Duration  time.Duration
	Success   bool
	ErrorMsg  string
}

// Session holds parsed session metadata and events.
type Session struct {
	ID        string
	Path      string
	StartTime time.Time
	EndTime   time.Time
	Events    []Event
}

// rawEntry is the top-level structure of a Claude Code JSONL line.
type rawEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Message   json.RawMessage `json:"message"`
	UserType  string          `json:"userType"`
}

type rawMessage struct {
	Role    string            `json:"role"`
	Content []json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type      string          `json:"type"`
	Name      string          `json:"name,omitempty"`
	ID        string          `json:"id,omitempty"`
	Text      string          `json:"text,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   interface{}     `json:"content,omitempty"`
	IsError   *bool           `json:"is_error,omitempty"`
}

type bashInput struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Timeout     int    `json:"timeout"`
}

type readInput struct {
	FilePath string `json:"file_path"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

type editInput struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

type writeInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

type grepInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

type globInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

type taskInput struct {
	Prompt       string `json:"prompt"`
	Description  string `json:"description"`
	SubagentType string `json:"subagent_type"`
}

// ListSessions returns all sessions found in the Claude projects directory.
func ListSessions(projectsDir string) ([]Session, error) {
	matches, err := filepath.Glob(filepath.Join(projectsDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("glob sessions: %w", err)
	}

	var sessions []Session
	for _, path := range matches {
		base := filepath.Base(path)
		id := strings.TrimSuffix(base, ".jsonl")

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		s := Session{
			ID:   id,
			Path: path,
		}

		// Quick scan for first and last timestamps
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		var firstTS, lastTS string
		for scanner.Scan() {
			var entry rawEntry
			if json.Unmarshal(scanner.Bytes(), &entry) != nil {
				continue
			}
			if entry.Timestamp == "" {
				continue
			}
			if firstTS == "" {
				firstTS = entry.Timestamp
			}
			lastTS = entry.Timestamp
		}
		f.Close()

		if firstTS != "" {
			s.StartTime, _ = time.Parse(time.RFC3339Nano, firstTS)
		}
		if lastTS != "" {
			s.EndTime, _ = time.Parse(time.RFC3339Nano, lastTS)
		}
		if s.StartTime.IsZero() {
			s.StartTime = info.ModTime()
		}

		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions, nil
}

// ParseTranscript reads a JSONL session file and returns structured events.
func ParseTranscript(path string) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	base := filepath.Base(path)
	sess := &Session{
		ID:   strings.TrimSuffix(base, ".jsonl"),
		Path: path,
	}

	// Collect tool_use entries keyed by ID
	type toolUse struct {
		timestamp time.Time
		tool      string
		input     string
	}
	pendingTools := make(map[string]toolUse)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		var entry rawEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)

		if sess.StartTime.IsZero() && !ts.IsZero() {
			sess.StartTime = ts
		}
		if !ts.IsZero() {
			sess.EndTime = ts
		}

		switch entry.Type {
		case "assistant":
			var msg rawMessage
			if json.Unmarshal(entry.Message, &msg) != nil {
				continue
			}
			for _, raw := range msg.Content {
				var block contentBlock
				if json.Unmarshal(raw, &block) != nil {
					continue
				}

				switch block.Type {
				case "text":
					if text := strings.TrimSpace(block.Text); text != "" {
						sess.Events = append(sess.Events, Event{
							Timestamp: ts,
							Type:      "assistant",
							Input:     truncate(text, 500),
						})
					}

				case "tool_use":
					inputStr := extractToolInput(block.Name, block.Input)
					pendingTools[block.ID] = toolUse{
						timestamp: ts,
						tool:      block.Name,
						input:     inputStr,
					}
				}
			}

		case "user":
			var msg rawMessage
			if json.Unmarshal(entry.Message, &msg) != nil {
				continue
			}
			for _, raw := range msg.Content {
				var block contentBlock
				if json.Unmarshal(raw, &block) != nil {
					continue
				}

				switch block.Type {
				case "tool_result":
					if tu, ok := pendingTools[block.ToolUseID]; ok {
						output := extractResultContent(block.Content)
						isError := block.IsError != nil && *block.IsError
						evt := Event{
							Timestamp: tu.timestamp,
							Type:      "tool_use",
							Tool:      tu.tool,
							ToolID:    block.ToolUseID,
							Input:     tu.input,
							Output:    truncate(output, 2000),
							Duration:  ts.Sub(tu.timestamp),
							Success:   !isError,
						}
						if isError {
							evt.ErrorMsg = truncate(output, 500)
						}
						sess.Events = append(sess.Events, evt)
						delete(pendingTools, block.ToolUseID)
					}

				case "text":
					if text := strings.TrimSpace(block.Text); text != "" {
						sess.Events = append(sess.Events, Event{
							Timestamp: ts,
							Type:      "user",
							Input:     truncate(text, 500),
						})
					}
				}
			}
		}
	}

	return sess, scanner.Err()
}

func extractToolInput(toolName string, raw json.RawMessage) string {
	if raw == nil {
		return ""
	}

	switch toolName {
	case "Bash":
		var inp bashInput
		if json.Unmarshal(raw, &inp) == nil {
			desc := inp.Description
			if desc != "" {
				desc = " # " + desc
			}
			return inp.Command + desc
		}
	case "Read":
		var inp readInput
		if json.Unmarshal(raw, &inp) == nil {
			return inp.FilePath
		}
	case "Edit":
		var inp editInput
		if json.Unmarshal(raw, &inp) == nil {
			return fmt.Sprintf("%s (edit)", inp.FilePath)
		}
	case "Write":
		var inp writeInput
		if json.Unmarshal(raw, &inp) == nil {
			return fmt.Sprintf("%s (%d bytes)", inp.FilePath, len(inp.Content))
		}
	case "Grep":
		var inp grepInput
		if json.Unmarshal(raw, &inp) == nil {
			path := inp.Path
			if path == "" {
				path = "."
			}
			return fmt.Sprintf("/%s/ in %s", inp.Pattern, path)
		}
	case "Glob":
		var inp globInput
		if json.Unmarshal(raw, &inp) == nil {
			return inp.Pattern
		}
	case "Task":
		var inp taskInput
		if json.Unmarshal(raw, &inp) == nil {
			desc := inp.Description
			if desc == "" {
				desc = truncate(inp.Prompt, 80)
			}
			return fmt.Sprintf("[%s] %s", inp.SubagentType, desc)
		}
	}

	// Fallback: show raw JSON keys
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) == nil {
		var parts []string
		for k := range m {
			parts = append(parts, k)
		}
		sort.Strings(parts)
		return strings.Join(parts, ", ")
	}

	return ""
}

func extractResultContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok {
			return text
		}
	}
	return fmt.Sprintf("%v", content)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
