package mcp

import (
	"context"
	"fmt"
	"strings"

	"forge.lthn.ai/core/go/pkg/log"
	"forge.lthn.ai/core/go/pkg/ml"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MLSubsystem exposes ML inference and scoring tools via MCP.
type MLSubsystem struct {
	service *ml.Service
	logger  *log.Logger
}

// NewMLSubsystem creates an MCP subsystem for ML tools.
func NewMLSubsystem(svc *ml.Service) *MLSubsystem {
	return &MLSubsystem{
		service: svc,
		logger:  log.Default(),
	}
}

func (m *MLSubsystem) Name() string { return "ml" }

// RegisterTools adds ML tools to the MCP server.
func (m *MLSubsystem) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ml_generate",
		Description: "Generate text via a configured ML inference backend.",
	}, m.mlGenerate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ml_score",
		Description: "Score a prompt/response pair using heuristic and LLM judge suites.",
	}, m.mlScore)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ml_probe",
		Description: "Run capability probes against an inference backend.",
	}, m.mlProbe)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ml_status",
		Description: "Show training and generation progress from InfluxDB.",
	}, m.mlStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ml_backends",
		Description: "List available inference backends and their status.",
	}, m.mlBackends)
}

// --- Input/Output types ---

// MLGenerateInput contains parameters for text generation.
type MLGenerateInput struct {
	Prompt      string  `json:"prompt"`                // The prompt to generate from
	Backend     string  `json:"backend,omitempty"`     // Backend name (default: service default)
	Model       string  `json:"model,omitempty"`       // Model override
	Temperature float64 `json:"temperature,omitempty"` // Sampling temperature
	MaxTokens   int     `json:"max_tokens,omitempty"`  // Maximum tokens to generate
}

// MLGenerateOutput contains the generation result.
type MLGenerateOutput struct {
	Response string `json:"response"`
	Backend  string `json:"backend"`
	Model    string `json:"model,omitempty"`
}

// MLScoreInput contains parameters for scoring a response.
type MLScoreInput struct {
	Prompt   string `json:"prompt"`           // The original prompt
	Response string `json:"response"`         // The model response to score
	Suites   string `json:"suites,omitempty"` // Comma-separated suites (default: heuristic)
}

// MLScoreOutput contains the scoring result.
type MLScoreOutput struct {
	Heuristic *ml.HeuristicScores `json:"heuristic,omitempty"`
	Semantic  *ml.SemanticScores  `json:"semantic,omitempty"`
	Content   *ml.ContentScores   `json:"content,omitempty"`
}

// MLProbeInput contains parameters for running probes.
type MLProbeInput struct {
	Backend    string `json:"backend,omitempty"`    // Backend name
	Categories string `json:"categories,omitempty"` // Comma-separated categories to run
}

// MLProbeOutput contains probe results.
type MLProbeOutput struct {
	Total   int                 `json:"total"`
	Results []MLProbeResultItem `json:"results"`
}

// MLProbeResultItem is a single probe result.
type MLProbeResultItem struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Response string `json:"response"`
}

// MLStatusInput contains parameters for the status query.
type MLStatusInput struct {
	InfluxURL string `json:"influx_url,omitempty"` // InfluxDB URL override
	InfluxDB  string `json:"influx_db,omitempty"`  // InfluxDB database override
}

// MLStatusOutput contains pipeline status.
type MLStatusOutput struct {
	Status string `json:"status"`
}

// MLBackendsInput is empty — lists all backends.
type MLBackendsInput struct{}

// MLBackendsOutput lists available backends.
type MLBackendsOutput struct {
	Backends []MLBackendInfo `json:"backends"`
	Default  string          `json:"default"`
}

// MLBackendInfo describes a single backend.
type MLBackendInfo struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// --- Tool handlers ---

func (m *MLSubsystem) mlGenerate(ctx context.Context, req *mcp.CallToolRequest, input MLGenerateInput) (*mcp.CallToolResult, MLGenerateOutput, error) {
	m.logger.Info("MCP tool execution", "tool", "ml_generate", "backend", input.Backend, "user", log.Username())

	if input.Prompt == "" {
		return nil, MLGenerateOutput{}, fmt.Errorf("prompt cannot be empty")
	}

	opts := ml.GenOpts{
		Temperature: input.Temperature,
		MaxTokens:   input.MaxTokens,
		Model:       input.Model,
	}

	response, err := m.service.Generate(ctx, input.Backend, input.Prompt, opts)
	if err != nil {
		return nil, MLGenerateOutput{}, fmt.Errorf("generate: %w", err)
	}

	return nil, MLGenerateOutput{
		Response: response,
		Backend:  input.Backend,
		Model:    input.Model,
	}, nil
}

func (m *MLSubsystem) mlScore(ctx context.Context, req *mcp.CallToolRequest, input MLScoreInput) (*mcp.CallToolResult, MLScoreOutput, error) {
	m.logger.Info("MCP tool execution", "tool", "ml_score", "suites", input.Suites, "user", log.Username())

	if input.Prompt == "" || input.Response == "" {
		return nil, MLScoreOutput{}, fmt.Errorf("prompt and response cannot be empty")
	}

	suites := input.Suites
	if suites == "" {
		suites = "heuristic"
	}

	output := MLScoreOutput{}

	for _, suite := range strings.Split(suites, ",") {
		suite = strings.TrimSpace(suite)
		switch suite {
		case "heuristic":
			output.Heuristic = ml.ScoreHeuristic(input.Response)
		case "semantic":
			judge := m.service.Judge()
			if judge == nil {
				return nil, MLScoreOutput{}, fmt.Errorf("semantic scoring requires a judge backend")
			}
			s, err := judge.ScoreSemantic(ctx, input.Prompt, input.Response)
			if err != nil {
				return nil, MLScoreOutput{}, fmt.Errorf("semantic score: %w", err)
			}
			output.Semantic = s
		case "content":
			return nil, MLScoreOutput{}, fmt.Errorf("content scoring requires a ContentProbe — use ml_probe instead")
		}
	}

	return nil, output, nil
}

func (m *MLSubsystem) mlProbe(ctx context.Context, req *mcp.CallToolRequest, input MLProbeInput) (*mcp.CallToolResult, MLProbeOutput, error) {
	m.logger.Info("MCP tool execution", "tool", "ml_probe", "backend", input.Backend, "user", log.Username())

	// Filter probes by category if specified.
	probes := ml.CapabilityProbes
	if input.Categories != "" {
		cats := make(map[string]bool)
		for _, c := range strings.Split(input.Categories, ",") {
			cats[strings.TrimSpace(c)] = true
		}
		var filtered []ml.Probe
		for _, p := range probes {
			if cats[p.Category] {
				filtered = append(filtered, p)
			}
		}
		probes = filtered
	}

	var results []MLProbeResultItem
	for _, probe := range probes {
		resp, err := m.service.Generate(ctx, input.Backend, probe.Prompt, ml.GenOpts{Temperature: 0.7, MaxTokens: 2048})
		if err != nil {
			resp = fmt.Sprintf("error: %v", err)
		}
		results = append(results, MLProbeResultItem{
			ID:       probe.ID,
			Category: probe.Category,
			Response: resp,
		})
	}

	return nil, MLProbeOutput{
		Total:   len(results),
		Results: results,
	}, nil
}

func (m *MLSubsystem) mlStatus(ctx context.Context, req *mcp.CallToolRequest, input MLStatusInput) (*mcp.CallToolResult, MLStatusOutput, error) {
	m.logger.Info("MCP tool execution", "tool", "ml_status", "user", log.Username())

	url := input.InfluxURL
	db := input.InfluxDB
	if url == "" {
		url = "http://localhost:8086"
	}
	if db == "" {
		db = "lem"
	}

	influx := ml.NewInfluxClient(url, db)
	var buf strings.Builder
	if err := ml.PrintStatus(influx, &buf); err != nil {
		return nil, MLStatusOutput{}, fmt.Errorf("status: %w", err)
	}

	return nil, MLStatusOutput{Status: buf.String()}, nil
}

func (m *MLSubsystem) mlBackends(ctx context.Context, req *mcp.CallToolRequest, input MLBackendsInput) (*mcp.CallToolResult, MLBackendsOutput, error) {
	m.logger.Info("MCP tool execution", "tool", "ml_backends", "user", log.Username())

	names := m.service.Backends()
	backends := make([]MLBackendInfo, len(names))
	defaultName := ""
	for i, name := range names {
		b := m.service.Backend(name)
		backends[i] = MLBackendInfo{
			Name:      name,
			Available: b != nil && b.Available(),
		}
	}

	if db := m.service.DefaultBackend(); db != nil {
		defaultName = db.Name()
	}

	return nil, MLBackendsOutput{
		Backends: backends,
		Default:  defaultName,
	}, nil
}
