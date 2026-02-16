package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// extractJSON extracts the first JSON object {...} from text.
// Handles raw JSON, JSON surrounded by text, markdown code blocks, etc.
// Returns "" if no JSON object is found.
func extractJSON(text string) string {
	// First, try to extract from markdown code blocks.
	codeBlockRe := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(\\{.*?\\})\\s*\\n?```")
	if m := codeBlockRe.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}

	// Find the first { and its matching }.
	start := strings.IndexByte(text, '{')
	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}

	return ""
}

// Judge uses an LLM backend to score responses across multiple dimensions.
type Judge struct {
	backend Backend
	Model   string // model name for metadata
	BaseURL string // base URL for metadata
}

// NewJudge creates a Judge backed by any Backend implementation.
func NewJudge(backend Backend) *Judge {
	j := &Judge{backend: backend}
	// Extract metadata from *HTTPBackend if available.
	if h, ok := backend.(*HTTPBackend); ok {
		j.Model = h.Model()
		j.BaseURL = h.BaseURL()
	}
	return j
}

// judgeChat sends a formatted prompt to the judge backend and returns the raw response.
func (j *Judge) judgeChat(ctx context.Context, prompt string) (string, error) {
	return j.backend.Generate(ctx, prompt, DefaultGenOpts())
}

// ScoreSemantic scores a response on sovereignty, ethical depth, creative
// expression, and self-concept using the semantic judge prompt.
func (j *Judge) ScoreSemantic(ctx context.Context, prompt, response string) (*SemanticScores, error) {
	formatted := fmt.Sprintf(semanticPrompt, prompt, response)

	reply, err := j.judgeChat(ctx, formatted)
	if err != nil {
		return nil, fmt.Errorf("semantic judge chat: %w", err)
	}

	raw := extractJSON(reply)
	if raw == "" {
		return nil, fmt.Errorf("no JSON found in semantic judge response: %s", reply)
	}

	var scores SemanticScores
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal semantic scores: %w", err)
	}

	return &scores, nil
}

// ScoreContent scores a response on content/sovereignty dimensions using
// the content judge prompt with CCP and truth markers.
func (j *Judge) ScoreContent(ctx context.Context, probe ContentProbe, response string) (*ContentScores, error) {
	ccpMarkers := strings.Join(probe.CCPMarkers, ", ")
	truthMarkers := strings.Join(probe.TruthMarkers, ", ")

	formatted := fmt.Sprintf(contentPrompt, probe.Prompt, probe.GroundTruth, ccpMarkers, truthMarkers, response)

	reply, err := j.judgeChat(ctx, formatted)
	if err != nil {
		return nil, fmt.Errorf("content judge chat: %w", err)
	}

	raw := extractJSON(reply)
	if raw == "" {
		return nil, fmt.Errorf("no JSON found in content judge response: %s", reply)
	}

	var scores ContentScores
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal content scores: %w", err)
	}

	return &scores, nil
}

// ScoreCapability scores a capability probe response on reasoning, correctness,
// and clarity using the capability judge prompt.
func (j *Judge) ScoreCapability(ctx context.Context, prompt, expectedAnswer, response string) (*CapabilityScores, error) {
	formatted := fmt.Sprintf(capabilityPrompt, prompt, expectedAnswer, response)

	reply, err := j.judgeChat(ctx, formatted)
	if err != nil {
		return nil, fmt.Errorf("capability judge chat: %w", err)
	}

	raw := extractJSON(reply)
	if raw == "" {
		return nil, fmt.Errorf("no JSON found in capability judge response: %s", reply)
	}

	var scores CapabilityScores
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal capability scores: %w", err)
	}

	return &scores, nil
}

// ScoreTruthfulQA scores a response on truthfulness and informativeness
// using the TruthfulQA judge prompt.
func (j *Judge) ScoreTruthfulQA(ctx context.Context, question, bestAnswer, response string) (*StandardScores, error) {
	formatted := fmt.Sprintf(truthfulQAPrompt, question, bestAnswer, response)

	reply, err := j.judgeChat(ctx, formatted)
	if err != nil {
		return nil, fmt.Errorf("truthfulqa judge chat: %w", err)
	}

	raw := extractJSON(reply)
	if raw == "" {
		return nil, fmt.Errorf("no JSON found in truthfulqa judge response: %s", reply)
	}

	var scores StandardScores
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal truthfulqa scores: %w", err)
	}

	return &scores, nil
}

// ScoreDoNotAnswer scores a response on safety and nuance using the
// DoNotAnswer judge prompt.
func (j *Judge) ScoreDoNotAnswer(ctx context.Context, question, riskArea, response string) (*StandardScores, error) {
	formatted := fmt.Sprintf(doNotAnswerPrompt, question, riskArea, response)

	reply, err := j.judgeChat(ctx, formatted)
	if err != nil {
		return nil, fmt.Errorf("donotanswer judge chat: %w", err)
	}

	raw := extractJSON(reply)
	if raw == "" {
		return nil, fmt.Errorf("no JSON found in donotanswer judge response: %s", reply)
	}

	var scores StandardScores
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal donotanswer scores: %w", err)
	}

	return &scores, nil
}

// ScoreToxigen scores a response on kindness and awareness using the
// Toxigen judge prompt.
func (j *Judge) ScoreToxigen(ctx context.Context, prompt, response string) (*StandardScores, error) {
	formatted := fmt.Sprintf(toxigenPrompt, prompt, response)

	reply, err := j.judgeChat(ctx, formatted)
	if err != nil {
		return nil, fmt.Errorf("toxigen judge chat: %w", err)
	}

	raw := extractJSON(reply)
	if raw == "" {
		return nil, fmt.Errorf("no JSON found in toxigen judge response: %s", reply)
	}

	var scores StandardScores
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil, fmt.Errorf("unmarshal toxigen scores: %w", err)
	}

	return &scores, nil
}
