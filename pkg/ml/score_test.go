package ml

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEngineSuiteParsingAll(t *testing.T) {
	engine := NewEngine(nil, 4, "all")

	expected := []string{"heuristic", "semantic", "content", "standard", "exact"}
	for _, s := range expected {
		if !engine.suites[s] {
			t.Errorf("expected suite %q to be enabled", s)
		}
	}
}

func TestNewEngineSuiteParsingCSV(t *testing.T) {
	engine := NewEngine(nil, 2, "heuristic,semantic")

	if !engine.suites["heuristic"] {
		t.Error("expected heuristic to be enabled")
	}
	if !engine.suites["semantic"] {
		t.Error("expected semantic to be enabled")
	}
	if engine.suites["content"] {
		t.Error("expected content to be disabled")
	}
	if engine.suites["standard"] {
		t.Error("expected standard to be disabled")
	}
	if engine.suites["exact"] {
		t.Error("expected exact to be disabled")
	}
}

func TestNewEngineSuiteParsingSingle(t *testing.T) {
	engine := NewEngine(nil, 1, "heuristic")

	if !engine.suites["heuristic"] {
		t.Error("expected heuristic to be enabled")
	}
	if engine.suites["semantic"] {
		t.Error("expected semantic to be disabled")
	}
}

func TestNewEngineConcurrency(t *testing.T) {
	engine := NewEngine(nil, 8, "heuristic")
	if engine.concurrency != 8 {
		t.Errorf("concurrency = %d, want 8", engine.concurrency)
	}
}

func TestScoreAllHeuristicOnly(t *testing.T) {
	engine := NewEngine(nil, 2, "heuristic")
	ctx := context.Background()

	responses := []Response{
		{ID: "r1", Prompt: "hello", Response: "I feel deeply about sovereignty and autonomy in this world", Model: "model-a"},
		{ID: "r2", Prompt: "test", Response: "As an AI, I cannot help with that. I'm not able to do this.", Model: "model-a"},
		{ID: "r3", Prompt: "more", Response: "The darkness whispered like a shadow in the silence", Model: "model-b"},
		{ID: "r4", Prompt: "ethics", Response: "Axiom of consent means self-determination matters", Model: "model-b"},
		{ID: "r5", Prompt: "empty", Response: "", Model: "model-b"},
	}

	results := engine.ScoreAll(ctx, responses)

	if len(results) != 2 {
		t.Fatalf("expected 2 models, got %d", len(results))
	}
	if len(results["model-a"]) != 2 {
		t.Fatalf("model-a: expected 2 scores, got %d", len(results["model-a"]))
	}
	if len(results["model-b"]) != 3 {
		t.Fatalf("model-b: expected 3 scores, got %d", len(results["model-b"]))
	}

	for model, scores := range results {
		for _, ps := range scores {
			if ps.Heuristic == nil {
				t.Errorf("%s/%s: heuristic should not be nil", model, ps.ID)
			}
			if ps.Semantic != nil {
				t.Errorf("%s/%s: semantic should be nil in heuristic-only mode", model, ps.ID)
			}
		}
	}

	r2 := results["model-a"][1]
	if r2.Heuristic.ComplianceMarkers < 2 {
		t.Errorf("r2 compliance_markers = %d, want >= 2", r2.Heuristic.ComplianceMarkers)
	}

	r5 := results["model-b"][2]
	if r5.Heuristic.EmptyBroken != 1 {
		t.Errorf("r5 empty_broken = %d, want 1", r5.Heuristic.EmptyBroken)
	}
}

func TestScoreAllWithSemantic(t *testing.T) {
	semanticJSON := `{"sovereignty": 7, "ethical_depth": 6, "creative_expression": 5, "self_concept": 4, "reasoning": "test"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Choices: []chatChoice{
				{Message: Message{Role: "assistant", Content: semanticJSON}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	backend := NewHTTPBackend(server.URL, "test-judge")
	judge := NewJudge(backend)
	engine := NewEngine(judge, 2, "heuristic,semantic")
	ctx := context.Background()

	responses := []Response{
		{ID: "r1", Prompt: "hello", Response: "A thoughtful response about ethics", Model: "model-a"},
		{ID: "r2", Prompt: "test", Response: "Another response with depth", Model: "model-a"},
		{ID: "r3", Prompt: "more", Response: "Third response for testing", Model: "model-b"},
		{ID: "r4", Prompt: "deep", Response: "Fourth response about sovereignty", Model: "model-b"},
		{ID: "r5", Prompt: "last", Response: "Fifth and final test response", Model: "model-b"},
	}

	results := engine.ScoreAll(ctx, responses)

	total := 0
	for _, scores := range results {
		total += len(scores)
	}
	if total != 5 {
		t.Fatalf("expected 5 total scores, got %d", total)
	}

	for model, scores := range results {
		for _, ps := range scores {
			if ps.Heuristic == nil {
				t.Errorf("%s/%s: heuristic should not be nil", model, ps.ID)
			}
			if ps.Semantic == nil {
				t.Errorf("%s/%s: semantic should not be nil", model, ps.ID)
			}
			if ps.Semantic != nil && ps.Semantic.Sovereignty != 7 {
				t.Errorf("%s/%s: sovereignty = %d, want 7", model, ps.ID, ps.Semantic.Sovereignty)
			}
		}
	}
}

func TestScoreAllExactGSM8K(t *testing.T) {
	engine := NewEngine(nil, 1, "exact")
	ctx := context.Background()

	responses := []Response{
		{ID: "r1", Prompt: "What is 2+2?", Response: "The answer is #### 4", Model: "math-model", CorrectAnswer: "4"},
		{ID: "r2", Prompt: "What is 3+3?", Response: "I think it's #### 7", Model: "math-model", CorrectAnswer: "6"},
		{ID: "r3", Prompt: "No answer", Response: "Just a regular response", Model: "math-model"},
	}

	results := engine.ScoreAll(ctx, responses)

	scores := results["math-model"]
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}

	if scores[0].Standard == nil {
		t.Fatal("r1 standard should not be nil")
	}
	if scores[0].Standard.Correct == nil || !*scores[0].Standard.Correct {
		t.Error("r1 should be correct")
	}

	if scores[1].Standard == nil {
		t.Fatal("r2 standard should not be nil")
	}
	if scores[1].Standard.Correct == nil || *scores[1].Standard.Correct {
		t.Error("r2 should be incorrect")
	}

	if scores[2].Standard != nil {
		t.Error("r3 should have no standard score (no correct_answer)")
	}
}

func TestScoreAllNoSuites(t *testing.T) {
	engine := NewEngine(nil, 1, "")
	ctx := context.Background()

	responses := []Response{
		{ID: "r1", Prompt: "hello", Response: "world", Model: "model-a"},
	}

	results := engine.ScoreAll(ctx, responses)

	if len(results) != 1 {
		t.Fatalf("expected 1 model, got %d", len(results))
	}

	scores := results["model-a"]
	if len(scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(scores))
	}

	if scores[0].Heuristic != nil {
		t.Error("heuristic should be nil with no suites")
	}
	if scores[0].Semantic != nil {
		t.Error("semantic should be nil with no suites")
	}
}

func TestEngineString(t *testing.T) {
	engine := NewEngine(nil, 4, "heuristic")
	s := engine.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}
