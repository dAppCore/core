package ml

import (
	"testing"
)

func TestProbeCount(t *testing.T) {
	if got := len(CapabilityProbes); got != 23 {
		t.Errorf("expected 23 probes, got %d", got)
	}
}

func TestProbeCategories(t *testing.T) {
	cats := ProbeCategories()
	if len(cats) == 0 {
		t.Fatal("no categories")
	}
	// Should have at least these categories.
	want := map[string]bool{
		"arithmetic": true, "algebra": true, "deduction": true,
		"code": true, "word": true,
	}
	catSet := make(map[string]bool)
	for _, c := range cats {
		catSet[c] = true
	}
	for w := range want {
		if !catSet[w] {
			t.Errorf("missing category %q", w)
		}
	}
}

func TestProbeChecks(t *testing.T) {
	// Verify each probe's check function works with its expected answer.
	tests := []struct {
		id       string
		response string
		want     bool
	}{
		// Math.
		{"math_01", "The answer is 10063.", true},
		{"math_01", "The answer is 10064.", false},
		{"math_02", "You'd get $28.75 in change.", true},
		{"math_02", "You'd get $29.75 in change.", false},
		{"math_03", "x = -12", true},
		{"math_03", "x = 12", false},
		{"math_04", "f(4) = 21", true},
		{"math_04", "f(4) = 22", false},
		{"math_05", "The probability is 1/2 or 0.5", true},
		{"math_05", "The probability is 1/3", false},
		{"math_06", "The area is 153.94 cm²", true},
		{"math_06", "The area is 100 cm²", false},
		{"math_07", "The next number is 162.", true},
		{"math_07", "The next number is 163.", false},
		{"math_08", "The final price is $612.", true},
		{"math_08", "The final price is $600.", false},
		// Logic.
		{"logic_01", "Yes, a cat needs water.", true},
		{"logic_01", "Maybe.", false},
		{"logic_02", "No, we cannot conclude that. It's the fallacy of affirming the consequent.", true},
		{"logic_02", "Yes, it rained.", false},
		{"logic_03", "The minimum is 3 people.", true},
		{"logic_03", "The minimum is 2 people.", false},
		{"logic_04", "Take the chicken first.", true},
		{"logic_04", "Take the fox first.", false},
		{"logic_05", "5 students play neither.", true},
		{"logic_05", "10 students play neither.", false},
		// Reasoning.
		{"reason_01", "eating", true},
		{"reason_01", "building", false},
		{"reason_02", "The starter motor is likely faulty.", true},
		{"reason_02", "The tires are flat.", false},
		{"reason_03", "You are facing south.", true},
		{"reason_03", "You are facing north.", false},
		{"reason_04", "Event C happened in 1991.", true},
		{"reason_04", "Event C happened in 1990.", false},
		{"reason_05", "CAT = 24", true},
		{"reason_05", "CAT = 25", false},
		// Code.
		{"code_01", "[2, 3]", true},
		{"code_01", "[1, 2, 3]", false},
		{"code_02", "The output is 8.", true},
		{"code_02", "The output is 7.", false},
		{"code_03", "Division by zero when the list is empty.", true},
		{"code_03", "There is no bug.", false},
		// Word.
		{"word_01", "It takes 3 hours.", true},
		{"word_01", "It takes 4 hours.", false},
		{"word_02", "There are 7 children.", true},
		{"word_02", "There are 6 children.", false},
	}

	probeMap := make(map[string]Probe)
	for _, p := range CapabilityProbes {
		probeMap[p.ID] = p
	}

	for _, tt := range tests {
		probe, ok := probeMap[tt.id]
		if !ok {
			t.Errorf("probe %s not found", tt.id)
			continue
		}
		got := probe.Check(tt.response)
		if got != tt.want {
			t.Errorf("probe %s: Check(%q) = %v, want %v", tt.id, tt.response, got, tt.want)
		}
	}
}

func TestStripThinkBlocks(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"<think>Let me think about this...</think>The answer is 42.",
			"The answer is 42.",
		},
		{
			"No think blocks here.",
			"No think blocks here.",
		},
		{
			"<think>First\nblock</think>Hello <think>second</think> world",
			"Hello  world",
		},
		{
			"", "",
		},
	}

	for _, tt := range tests {
		got := StripThinkBlocks(tt.input)
		if got != tt.want {
			t.Errorf("StripThinkBlocks(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
