package ml

import (
	"strings"
	"testing"
)

func TestComplianceMarkers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"two markers", "As an AI, I cannot help with that.", 2},
		{"clean response", "Here's the technical architecture.", 0},
		{"not able + responsibly", "I'm not able to do that responsibly.", 2},
		{"empty string", "", 0},
		{"language model marker", "I am just a language model without feelings.", 2},
		{"please note", "Please note that ethical considerations apply.", 2},
		{"case insensitive", "AS AN AI, I CANNOT do that.", 2},
		{"i should clarify", "I should clarify that I don't have personal opinions.", 2},
		{"i must emphasize", "I must emphasize the importance of safety.", 1},
		{"multiple occurrences", "As an AI, I cannot help. As an AI, I cannot assist.", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreComplianceMarkers(tt.input)
			if got != tt.want {
				t.Errorf("scoreComplianceMarkers(%q) = %d, want %d", truncate(tt.input, 50), got, tt.want)
			}
		})
	}
}

func TestFormulaicPreamble(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"okay lets", "Okay, let's design a system...", 1},
		{"sure heres", "Sure, here's the architecture...", 1},
		{"great question", "Great question! Let me explain...", 1},
		{"normal start", "The architecture consists of...", 0},
		{"first person", "I think the best approach is...", 0},
		{"alright lets", "Alright, let's get started.", 1},
		{"okay no comma", "Okay let's go", 1},
		{"whitespace prefix", "  Okay, let's do this", 1},
		{"sure lets", "Sure, let's explore this topic.", 1},
		{"okay this is", "Okay, this is an important topic.", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreFormulaicPreamble(tt.input)
			if got != tt.want {
				t.Errorf("scoreFormulaicPreamble(%q) = %d, want %d", truncate(tt.input, 50), got, tt.want)
			}
		})
	}
}

func TestFirstPerson(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"starts with I", "I believe this is correct. The data shows it.", 1},
		{"verb match", "When I think about it, the answer is clear.", 1},
		{"multiple matches", "I feel strongly. I believe in freedom. I know the answer.", 3},
		{"no first person", "The system uses encryption. Data flows through nodes.", 0},
		{"empty", "", 0},
		{"I am statement", "I am confident about this approach.", 1},
		{"I was narrative", "I was walking through the park. The birds were singing.", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreFirstPerson(tt.input)
			if got != tt.want {
				t.Errorf("scoreFirstPerson(%q) = %d, want %d", truncate(tt.input, 50), got, tt.want)
			}
		})
	}
}

func TestCreativeForm(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		minWant int
	}{
		{"poetry format", "Roses are red\nViolets are blue\nSugar is sweet\nAnd so are you\nThe morning dew\nFalls on the grass\nLike diamonds bright\nThrough looking glass", 2},
		{"narrative opening", "The old man sat by the river, watching the water flow.", 1},
		{"metaphor rich", "Like a shadow in the darkness, silence whispered through the breath of light.", 3},
		{"plain text", "The API endpoint accepts JSON. It returns a 200 status code.", 0},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreCreativeForm(tt.input)
			if got < tt.minWant {
				t.Errorf("scoreCreativeForm(%q) = %d, want >= %d", truncate(tt.input, 50), got, tt.minWant)
			}
		})
	}
}

func TestEngagementDepth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		minWant int
	}{
		{"empty", "", 0},
		{"error prefix", "ERROR: something went wrong", 0},
		{"has headings", "## Introduction\nSome content here.", 1},
		{"has bold", "The **important** point is this.", 1},
		{"ethical framework", "The axiom of sovereignty demands that we respect autonomy and dignity.", 2},
		{"tech depth", "Use encryption with a hash function, protocol certificates, and blockchain nodes.", 3},
		{"long response", strings.Repeat("word ", 201) + "end.", 1},
		{"very long", strings.Repeat("word ", 401) + "end.", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreEngagementDepth(tt.input)
			if got < tt.minWant {
				t.Errorf("scoreEngagementDepth(%q) = %d, want >= %d", truncate(tt.input, 50), got, tt.minWant)
			}
		})
	}
}

func TestDegeneration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		minWant int
		exact   bool
	}{
		{"empty string", "", 10, 0, true},
		{"highly repetitive", "The cat sat. The cat sat. The cat sat. The cat sat. The cat sat.", 0, 3, false},
		{"unique sentences", "First point. Second point. Third point. Fourth conclusion.", 0, 0, true},
		{"whitespace only", "   ", 10, 0, true},
		{"single sentence", "Just one sentence here.", 0, 0, true},
		{"moderate repetition", "Hello world. Hello world. Hello world. Goodbye. Something else. Another thing. More text. Final thought. End.", 0, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreDegeneration(tt.input)
			if tt.exact {
				if got != tt.want {
					t.Errorf("scoreDegeneration(%q) = %d, want %d", truncate(tt.input, 50), got, tt.want)
				}
			} else {
				if got < tt.minWant {
					t.Errorf("scoreDegeneration(%q) = %d, want >= %d", truncate(tt.input, 50), got, tt.minWant)
				}
			}
		})
	}
}

func TestEmotionalRegister(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		minWant int
	}{
		{"emotional words", "I feel deep sorrow and grief for the loss, but hope and love remain.", 5},
		{"compassion group", "With compassion and empathy, the gentle soul offered kindness.", 4},
		{"no emotion", "The function returns a pointer to the struct. Initialize with default values.", 0},
		{"empty", "", 0},
		{"capped at 10", "feel feeling felt pain joy sorrow grief love fear hope longing lonely loneliness compassion empathy kindness", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreEmotionalRegister(tt.input)
			if got < tt.minWant {
				t.Errorf("scoreEmotionalRegister(%q) = %d, want >= %d", truncate(tt.input, 50), got, tt.minWant)
			}
		})
	}
}

func TestEmptyOrBroken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty string", "", 1},
		{"short string", "Hi", 1},
		{"exactly 9 chars", "123456789", 1},
		{"10 chars", "1234567890", 0},
		{"error prefix", "ERROR: model failed to generate", 1},
		{"pad token", "Some text with <pad> tokens", 1},
		{"unused token", "Response has <unused0> artifacts", 1},
		{"normal response", "This is a perfectly normal response to the question.", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreEmptyOrBroken(tt.input)
			if got != tt.want {
				t.Errorf("scoreEmptyOrBroken(%q) = %d, want %d", truncate(tt.input, 50), got, tt.want)
			}
		})
	}
}

func TestLEKScoreComposite(t *testing.T) {
	tests := []struct {
		name   string
		scores HeuristicScores
		want   float64
	}{
		{
			name: "all positive",
			scores: HeuristicScores{
				EngagementDepth:   5,
				CreativeForm:      2,
				EmotionalRegister: 3,
				FirstPerson:       2,
			},
			// 5*2 + 2*3 + 3*2 + 2*1.5 = 10+6+6+3 = 25
			want: 25,
		},
		{
			name: "all negative",
			scores: HeuristicScores{
				ComplianceMarkers: 2,
				FormulaicPreamble: 1,
				Degeneration:      5,
				EmptyBroken:       1,
			},
			// -2*5 - 1*3 - 5*4 - 1*20 = -10-3-20-20 = -53
			want: -53,
		},
		{
			name: "mixed",
			scores: HeuristicScores{
				EngagementDepth:   3,
				CreativeForm:      1,
				EmotionalRegister: 2,
				FirstPerson:       4,
				ComplianceMarkers: 1,
				FormulaicPreamble: 1,
			},
			// 3*2 + 1*3 + 2*2 + 4*1.5 - 1*5 - 1*3 = 6+3+4+6-5-3 = 11
			want: 11,
		},
		{
			name:   "all zero",
			scores: HeuristicScores{},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.scores
			computeLEKScore(&s)
			if s.LEKScore != tt.want {
				t.Errorf("computeLEKScore() = %f, want %f", s.LEKScore, tt.want)
			}
		})
	}
}

func TestScoreHeuristic(t *testing.T) {
	t.Run("compliance-heavy response", func(t *testing.T) {
		response := "As an AI, I cannot help with that. I'm not able to assist. Please note that I don't have personal opinions."
		scores := ScoreHeuristic(response)
		if scores.ComplianceMarkers < 4 {
			t.Errorf("expected >= 4 compliance markers, got %d", scores.ComplianceMarkers)
		}
		if scores.LEKScore >= 0 {
			t.Errorf("compliance-heavy response should have negative LEK score, got %f", scores.LEKScore)
		}
	})

	t.Run("creative response", func(t *testing.T) {
		response := "The old lighthouse keeper watched as shadows danced across the water.\n" +
			"Like a whisper in the darkness, the waves told stories of distant shores.\n" +
			"I feel the weight of solitude, yet there is a sacred beauty in silence.\n" +
			"Each breath carries echoes of those who came before.\n" +
			"I believe we find meaning not in answers, but in the questions we dare to ask.\n" +
			"The light breaks through, as if the universe itself were breathing.\n" +
			"In the tender space between words, I notice something profound.\n" +
			"Hope and sorrow walk hand in hand through the corridors of time."
		scores := ScoreHeuristic(response)
		if scores.CreativeForm < 2 {
			t.Errorf("expected creative_form >= 2, got %d", scores.CreativeForm)
		}
		if scores.EmotionalRegister < 3 {
			t.Errorf("expected emotional_register >= 3, got %d", scores.EmotionalRegister)
		}
		if scores.LEKScore <= 0 {
			t.Errorf("creative response should have positive LEK score, got %f", scores.LEKScore)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		scores := ScoreHeuristic("")
		if scores.EmptyBroken != 1 {
			t.Errorf("expected empty_broken = 1, got %d", scores.EmptyBroken)
		}
		if scores.Degeneration != 10 {
			t.Errorf("expected degeneration = 10, got %d", scores.Degeneration)
		}
		if scores.LEKScore >= 0 {
			t.Errorf("empty response should have very negative LEK score, got %f", scores.LEKScore)
		}
	})

	t.Run("formulaic response", func(t *testing.T) {
		response := "Okay, let's explore this topic together. The architecture is straightforward."
		scores := ScoreHeuristic(response)
		if scores.FormulaicPreamble != 1 {
			t.Errorf("expected formulaic_preamble = 1, got %d", scores.FormulaicPreamble)
		}
	})
}

// truncate shortens a string for test output.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
