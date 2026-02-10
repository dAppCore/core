package bugseti

import (
	"testing"
)

func TestSanitizeInline_Good(t *testing.T) {
	input := "Hello world"
	output := sanitizeInline(input, 50)
	if output != input {
		t.Fatalf("expected %q, got %q", input, output)
	}
}

func TestSanitizeInline_Bad(t *testing.T) {
	input := "Hello\nworld\t\x00"
	expected := "Hello world"
	output := sanitizeInline(input, 50)
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestSanitizeMultiline_Ugly(t *testing.T) {
	input := "ab\ncd\tef\x00"
	output := sanitizeMultiline(input, 5)
	if output != "ab\ncd" {
		t.Fatalf("expected %q, got %q", "ab\ncd", output)
	}
}

func TestSanitizeEnv_Good(t *testing.T) {
	g := &EthicsGuard{}
	input := "owner/repo-name"
	output := g.SanitizeEnv(input)
	if output != input {
		t.Fatalf("expected %q, got %q", input, output)
	}
}

func TestSanitizeEnv_Bad(t *testing.T) {
	g := &EthicsGuard{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"backtick", "owner/repo`whoami`", "owner/repowhoami"},
		{"dollar", "owner/repo$(id)", "owner/repoid"},
		{"semicolon", "owner/repo;rm -rf /", "owner/reporm -rf /"},
		{"pipe", "owner/repo|cat /etc/passwd", "owner/repocat /etc/passwd"},
		{"ampersand", "owner/repo&&echo pwned", "owner/repoecho pwned"},
		{"mixed", "`$;|&(){}<>!\\'\"\n\r", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := g.SanitizeEnv(tc.input)
			if output != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, output)
			}
		})
	}
}

func TestStripShellMeta_Ugly(t *testing.T) {
	// All metacharacters should be stripped, leaving empty string
	input := "`$;|&(){}<>!\\'\""
	output := stripShellMeta(input)
	if output != "" {
		t.Fatalf("expected empty string, got %q", output)
	}
}
