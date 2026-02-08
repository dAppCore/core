package session

import (
	"path/filepath"
	"strings"
	"time"
)

// SearchResult represents a match found in a session transcript.
type SearchResult struct {
	SessionID string
	Timestamp time.Time
	Tool      string
	Match     string
}

// Search finds events matching the query across all sessions in the directory.
func Search(projectsDir, query string) ([]SearchResult, error) {
	matches, err := filepath.Glob(filepath.Join(projectsDir, "*.jsonl"))
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	query = strings.ToLower(query)

	for _, path := range matches {
		sess, err := ParseTranscript(path)
		if err != nil {
			continue
		}

		for _, evt := range sess.Events {
			if evt.Type != "tool_use" {
				continue
			}
			text := strings.ToLower(evt.Input + " " + evt.Output)
			if strings.Contains(text, query) {
				matchCtx := evt.Input
				if matchCtx == "" {
					matchCtx = truncate(evt.Output, 120)
				}
				results = append(results, SearchResult{
					SessionID: sess.ID,
					Timestamp: evt.Timestamp,
					Tool:      evt.Tool,
					Match:     matchCtx,
				})
			}
		}
	}

	return results, nil
}
