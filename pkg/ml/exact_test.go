package ml

import "testing"

func TestScoreGSM8K(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		correctAnswer string
		wantCorrect   bool
		wantExtracted string
	}{
		{
			name:          "hash delimiter correct",
			response:      "The answer is #### 42",
			correctAnswer: "42",
			wantCorrect:   true,
			wantExtracted: "42",
		},
		{
			name:          "last number match correct",
			response:      "Let me calculate... the result is 42.0",
			correctAnswer: "42",
			wantCorrect:   true,
			wantExtracted: "42.0",
		},
		{
			name:          "last number incorrect",
			response:      "I think it's 43",
			correctAnswer: "42",
			wantCorrect:   false,
			wantExtracted: "43",
		},
		{
			name:          "comma separated correct",
			response:      "#### 1,234",
			correctAnswer: "1234",
			wantCorrect:   true,
			wantExtracted: "1,234",
		},
		{
			name:          "no numbers",
			response:      "No numbers here",
			correctAnswer: "5",
			wantCorrect:   false,
			wantExtracted: "",
		},
		{
			name:          "empty response",
			response:      "",
			correctAnswer: "5",
			wantCorrect:   false,
			wantExtracted: "",
		},
		{
			name:          "error response",
			response:      "ERROR: model timeout",
			correctAnswer: "10",
			wantCorrect:   false,
			wantExtracted: "",
		},
		{
			name:          "multiple numbers picks last",
			response:      "First 10, then 20, finally 30",
			correctAnswer: "30",
			wantCorrect:   true,
			wantExtracted: "30",
		},
		{
			name:          "negative number",
			response:      "The answer is #### -5",
			correctAnswer: "-5",
			wantCorrect:   true,
			wantExtracted: "-5",
		},
		{
			name:          "decimal answer",
			response:      "Result = 3.14",
			correctAnswer: "3.14",
			wantCorrect:   true,
			wantExtracted: "3.14",
		},
		{
			name:          "hash takes priority over last number",
			response:      "Steps: 10 + 20 = 30 #### 30 and some trailing 99",
			correctAnswer: "30",
			wantCorrect:   true,
			wantExtracted: "30",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := scoreGSM8K(tt.response, tt.correctAnswer)

			if scores.Correct == nil {
				t.Fatal("Correct field is nil")
			}
			if *scores.Correct != tt.wantCorrect {
				t.Errorf("correct = %v, want %v", *scores.Correct, tt.wantCorrect)
			}
			if scores.Extracted != tt.wantExtracted {
				t.Errorf("extracted = %q, want %q", scores.Extracted, tt.wantExtracted)
			}
			if scores.Expected != tt.correctAnswer {
				t.Errorf("expected = %q, want %q", scores.Expected, tt.correctAnswer)
			}
		})
	}
}
