package bugseti

import "testing"

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
