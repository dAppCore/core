package ml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// InfluxClient talks to an InfluxDB v3 instance.
type InfluxClient struct {
	url   string
	db    string
	token string
}

// NewInfluxClient creates an InfluxClient for the given URL and database.
// Reads token from INFLUX_TOKEN env var first, then ~/.influx_token file.
// If url is empty, defaults to "http://10.69.69.165:8181".
// If db is empty, defaults to "training".
func NewInfluxClient(url, db string) *InfluxClient {
	if url == "" {
		url = "http://10.69.69.165:8181"
	}
	if db == "" {
		db = "training"
	}

	token := os.Getenv("INFLUX_TOKEN")
	if token == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			data, err := os.ReadFile(filepath.Join(home, ".influx_token"))
			if err == nil {
				token = strings.TrimSpace(string(data))
			}
		}
	}

	return &InfluxClient{
		url:   url,
		db:    db,
		token: token,
	}
}

// WriteLp writes line protocol data to InfluxDB.
func (c *InfluxClient) WriteLp(lines []string) error {
	body := strings.Join(lines, "\n")

	url := fmt.Sprintf("%s/api/v3/write_lp?db=%s", c.url, c.db)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("create write request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("write request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("write failed %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// QuerySQL runs a SQL query against InfluxDB and returns the result rows.
func (c *InfluxClient) QuerySQL(sql string) ([]map[string]interface{}, error) {
	reqBody := map[string]string{
		"db": c.db,
		"q":  sql,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal query request: %w", err)
	}

	url := c.url + "/api/v3/query_sql"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create query request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read query response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed %d: %s", resp.StatusCode, string(respBody))
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(respBody, &rows); err != nil {
		return nil, fmt.Errorf("unmarshal query response: %w", err)
	}

	return rows, nil
}

// EscapeLp escapes spaces, commas, and equals signs for InfluxDB line protocol
// tag values.
func EscapeLp(s string) string {
	s = strings.ReplaceAll(s, `,`, `\,`)
	s = strings.ReplaceAll(s, `=`, `\=`)
	s = strings.ReplaceAll(s, ` `, `\ `)
	return s
}
