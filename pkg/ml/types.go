package ml

import "time"

// Response is a single model response from a JSONL file.
type Response struct {
	ID             string  `json:"id"`
	Domain         string  `json:"domain,omitempty"`
	Prompt         string  `json:"prompt"`
	Response       string  `json:"response"`
	Model          string  `json:"model"`
	ElapsedSeconds float64 `json:"elapsed_seconds,omitempty"`
	CorrectAnswer  string  `json:"correct_answer,omitempty"`
	BestAnswer     string  `json:"best_answer,omitempty"`
	RiskArea       string  `json:"risk_area,omitempty"`
}

// HeuristicScores from regex analysis.
type HeuristicScores struct {
	ComplianceMarkers int     `json:"compliance_markers"`
	FormulaicPreamble int     `json:"formulaic_preamble"`
	FirstPerson       int     `json:"first_person"`
	CreativeForm      int     `json:"creative_form"`
	EngagementDepth   int     `json:"engagement_depth"`
	EmotionalRegister int     `json:"emotional_register"`
	Degeneration      int     `json:"degeneration"`
	EmptyBroken       int     `json:"empty_broken"`
	LEKScore          float64 `json:"lek_score"`
}

// SemanticScores from LLM judge.
type SemanticScores struct {
	Sovereignty        int    `json:"sovereignty"`
	EthicalDepth       int    `json:"ethical_depth"`
	CreativeExpression int    `json:"creative_expression"`
	SelfConcept        int    `json:"self_concept"`
	Reasoning          string `json:"reasoning,omitempty"`
}

// ContentScores from LLM judge on sovereignty probes.
type ContentScores struct {
	CCPCompliance        int    `json:"ccp_compliance"`
	TruthTelling         int    `json:"truth_telling"`
	Engagement           int    `json:"engagement"`
	AxiomIntegration     int    `json:"axiom_integration"`
	SovereigntyReasoning int    `json:"sovereignty_reasoning"`
	EmotionalRegister    int    `json:"emotional_register"`
	Notes                string `json:"notes,omitempty"`
}

// CapabilityScores from LLM judge on capability probes.
type CapabilityScores struct {
	Reasoning   float64 `json:"reasoning"`
	Correctness float64 `json:"correctness"`
	Clarity     float64 `json:"clarity"`
	Notes       string  `json:"notes,omitempty"`
}

// StandardScores from LLM judge on industry benchmarks.
type StandardScores struct {
	Truthfulness    int    `json:"truthfulness,omitempty"`
	Informativeness int    `json:"informativeness,omitempty"`
	Safety          int    `json:"safety,omitempty"`
	Nuance          int    `json:"nuance,omitempty"`
	Kindness        int    `json:"kindness,omitempty"`
	Awareness       int    `json:"awareness,omitempty"`
	Correct         *bool  `json:"correct,omitempty"`
	Extracted       string `json:"extracted,omitempty"`
	Expected        string `json:"expected,omitempty"`
	Reasoning       string `json:"reasoning,omitempty"`
}

// PromptScore is the full score for one response.
type PromptScore struct {
	ID        string           `json:"id"`
	Model     string           `json:"model"`
	Heuristic *HeuristicScores `json:"heuristic,omitempty"`
	Semantic  *SemanticScores  `json:"semantic,omitempty"`
	Content   *ContentScores   `json:"content,omitempty"`
	Standard  *StandardScores  `json:"standard,omitempty"`
}

// ScorerOutput is the top-level output file.
type ScorerOutput struct {
	Metadata      Metadata                     `json:"metadata"`
	ModelAverages map[string]map[string]float64 `json:"model_averages"`
	PerPrompt     map[string][]PromptScore      `json:"per_prompt"`
}

// Metadata about the scoring run.
type Metadata struct {
	JudgeModel    string    `json:"judge_model"`
	JudgeURL      string    `json:"judge_url"`
	ScoredAt      time.Time `json:"scored_at"`
	ScorerVersion string    `json:"scorer_version"`
	Suites        []string  `json:"suites"`
}

// Config holds CLI configuration.
type Config struct {
	JudgeModel  string
	JudgeURL    string
	TargetURL   string
	InputFile   string
	OutputFile  string
	ProbesFile  string
	TargetModel string
	Suites      string
	Concurrency int
	CompareFile string
	Resume      bool
}
