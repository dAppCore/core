// Package session provides commands for replaying and searching Claude Code session transcripts.
package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/session"
)

func init() {
	cli.RegisterCommands(AddSessionCommands)
}

// AddSessionCommands registers the 'session' command group.
func AddSessionCommands(root *cli.Command) {
	sessionCmd := &cli.Command{
		Use:   "session",
		Short: "Session recording and replay",
	}
	root.AddCommand(sessionCmd)

	addListCommand(sessionCmd)
	addReplayCommand(sessionCmd)
	addSearchCommand(sessionCmd)
}

func projectsDir() string {
	home, _ := os.UserHomeDir()
	// Walk .claude/projects/ looking for dirs with .jsonl files
	base := filepath.Join(home, ".claude", "projects")
	entries, err := os.ReadDir(base)
	if err != nil {
		return base
	}
	// Return the first project dir that has .jsonl files
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(base, e.Name())
		matches, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		if len(matches) > 0 {
			return dir
		}
	}
	return base
}

func addListCommand(parent *cli.Command) {
	listCmd := &cli.Command{
		Use:   "list",
		Short: "List recent sessions",
		RunE: func(cmd *cli.Command, args []string) error {
			sessions, err := session.ListSessions(projectsDir())
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				cli.Print("No sessions found")
				return nil
			}

			cli.Print(cli.HeaderStyle.Render("Recent Sessions"))
			cli.Print("")
			for i, s := range sessions {
				if i >= 20 {
					cli.Print(cli.DimStyle.Render(fmt.Sprintf("  ... and %d more", len(sessions)-20)))
					break
				}
				dur := s.EndTime.Sub(s.StartTime)
				durStr := ""
				if dur > 0 {
					durStr = fmt.Sprintf(" (%s)", formatDur(dur))
				}
				id := s.ID
				if len(id) > 8 {
					id = id[:8]
				}
				cli.Print(fmt.Sprintf("  %s  %s%s",
					cli.ValueStyle.Render(id),
					s.StartTime.Format("2006-01-02 15:04"),
					cli.DimStyle.Render(durStr)))
			}
			return nil
		},
	}
	parent.AddCommand(listCmd)
}

func addReplayCommand(parent *cli.Command) {
	var mp4 bool
	var output string

	replayCmd := &cli.Command{
		Use:   "replay <session-id>",
		Short: "Generate HTML timeline (and optional MP4) from a session",
		Args:  cli.MinimumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			id := args[0]
			path := findSession(id)
			if path == "" {
				return fmt.Errorf("session not found: %s", id)
			}

			cli.Print(fmt.Sprintf("Parsing %s...", cli.ValueStyle.Render(filepath.Base(path))))

			sess, err := session.ParseTranscript(path)
			if err != nil {
				return fmt.Errorf("parse: %w", err)
			}

			toolCount := 0
			for _, e := range sess.Events {
				if e.Type == "tool_use" {
					toolCount++
				}
			}
			cli.Print(fmt.Sprintf("  %d events, %d tool calls",
				len(sess.Events), toolCount))

			// HTML output
			htmlPath := output
			if htmlPath == "" {
				htmlPath = fmt.Sprintf("session-%s.html", shortID(sess.ID))
			}
			if err := session.RenderHTML(sess, htmlPath); err != nil {
				return fmt.Errorf("render html: %w", err)
			}
			cli.Print(cli.SuccessStyle.Render(fmt.Sprintf("  HTML: %s", htmlPath)))

			// MP4 output
			if mp4 {
				mp4Path := strings.TrimSuffix(htmlPath, ".html") + ".mp4"
				if err := session.RenderMP4(sess, mp4Path); err != nil {
					cli.Print(cli.ErrorStyle.Render(fmt.Sprintf("  MP4: %s", err)))
				} else {
					cli.Print(cli.SuccessStyle.Render(fmt.Sprintf("  MP4: %s", mp4Path)))
				}
			}

			return nil
		},
	}
	replayCmd.Flags().BoolVar(&mp4, "mp4", false, "Also generate MP4 video (requires vhs + ffmpeg)")
	replayCmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	parent.AddCommand(replayCmd)
}

func addSearchCommand(parent *cli.Command) {
	searchCmd := &cli.Command{
		Use:   "search <query>",
		Short: "Search across session transcripts",
		Args:  cli.MinimumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			query := strings.ToLower(strings.Join(args, " "))
			results, err := session.Search(projectsDir(), query)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				cli.Print("No matches found")
				return nil
			}

			cli.Print(cli.HeaderStyle.Render(fmt.Sprintf("Found %d matches", len(results))))
			cli.Print("")
			for _, r := range results {
				id := r.SessionID
				if len(id) > 8 {
					id = id[:8]
				}
				cli.Print(fmt.Sprintf("  %s  %s  %s",
					cli.ValueStyle.Render(id),
					r.Timestamp.Format("15:04:05"),
					cli.DimStyle.Render(r.Tool)))
				cli.Print(fmt.Sprintf("    %s", truncateStr(r.Match, 100)))
				cli.Print("")
			}
			return nil
		},
	}
	parent.AddCommand(searchCmd)
}

func findSession(id string) string {
	dir := projectsDir()
	// Try exact match first
	path := filepath.Join(dir, id+".jsonl")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	// Try prefix match
	matches, _ := filepath.Glob(filepath.Join(dir, id+"*.jsonl"))
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatDur(d interface {
	Hours() float64
	Minutes() float64
	Seconds() float64
}) string {
	type dur interface {
		Hours() float64
		Minutes() float64
		Seconds() float64
	}
	dd := d.(dur)
	h := int(dd.Hours())
	m := int(dd.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	s := int(dd.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
