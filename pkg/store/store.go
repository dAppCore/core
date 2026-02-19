package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"text/template"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a key does not exist in the store.
var ErrNotFound = errors.New("store: not found")

// Store is a group-namespaced key-value store backed by SQLite.
type Store struct {
	db *sql.DB
}

// New creates a Store at the given SQLite path. Use ":memory:" for tests.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store.New: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.New: WAL: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS kv (
		grp   TEXT NOT NULL,
		key   TEXT NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY (grp, key)
	)`); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.New: schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Get retrieves a value by group and key.
func (s *Store) Get(group, key string) (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM kv WHERE grp = ? AND key = ?", group, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("store.Get: %s/%s: %w", group, key, ErrNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("store.Get: %w", err)
	}
	return val, nil
}

// Set stores a value by group and key, overwriting if exists.
func (s *Store) Set(group, key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO kv (grp, key, value) VALUES (?, ?, ?)
		 ON CONFLICT(grp, key) DO UPDATE SET value = excluded.value`,
		group, key, value,
	)
	if err != nil {
		return fmt.Errorf("store.Set: %w", err)
	}
	return nil
}

// Delete removes a single key from a group.
func (s *Store) Delete(group, key string) error {
	_, err := s.db.Exec("DELETE FROM kv WHERE grp = ? AND key = ?", group, key)
	if err != nil {
		return fmt.Errorf("store.Delete: %w", err)
	}
	return nil
}

// Count returns the number of keys in a group.
func (s *Store) Count(group string) (int, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM kv WHERE grp = ?", group).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("store.Count: %w", err)
	}
	return n, nil
}

// DeleteGroup removes all keys in a group.
func (s *Store) DeleteGroup(group string) error {
	_, err := s.db.Exec("DELETE FROM kv WHERE grp = ?", group)
	if err != nil {
		return fmt.Errorf("store.DeleteGroup: %w", err)
	}
	return nil
}

// GetAll returns all key-value pairs in a group.
func (s *Store) GetAll(group string) (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, value FROM kv WHERE grp = ?", group)
	if err != nil {
		return nil, fmt.Errorf("store.GetAll: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("store.GetAll: scan: %w", err)
		}
		result[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetAll: rows: %w", err)
	}
	return result, nil
}

// Render loads all key-value pairs from a group and renders a Go template.
func (s *Store) Render(tmplStr, group string) (string, error) {
	rows, err := s.db.Query("SELECT key, value FROM kv WHERE grp = ?", group)
	if err != nil {
		return "", fmt.Errorf("store.Render: query: %w", err)
	}
	defer rows.Close()

	vars := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return "", fmt.Errorf("store.Render: scan: %w", err)
		}
		vars[k] = v
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("store.Render: rows: %w", err)
	}

	tmpl, err := template.New("render").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("store.Render: parse: %w", err)
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, vars); err != nil {
		return "", fmt.Errorf("store.Render: exec: %w", err)
	}
	return b.String(), nil
}
