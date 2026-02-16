//go:build darwin && arm64 && mlx

package ml

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"forge.lthn.ai/core/cli/pkg/mlx"
	"forge.lthn.ai/core/cli/pkg/mlx/cache"
	"forge.lthn.ai/core/cli/pkg/mlx/model"
	"forge.lthn.ai/core/cli/pkg/mlx/sample"
	"forge.lthn.ai/core/cli/pkg/mlx/tokenizer"
)

// MLXBackend implements Backend for native Metal inference via mlx-c.
type MLXBackend struct {
	model   *model.GemmaModel
	tok     *tokenizer.Tokenizer
	caches  []cache.Cache
	sampler sample.Sampler
	mu      sync.Mutex
}

// NewMLXBackend loads a model from a safetensors directory and creates
// a native Metal inference backend.
func NewMLXBackend(modelPath string) (*MLXBackend, error) {
	if !mlx.MetalAvailable() {
		return nil, fmt.Errorf("mlx: Metal GPU not available")
	}

	slog.Info("mlx: loading model", "path", modelPath)
	m, err := model.LoadGemma3(modelPath)
	if err != nil {
		return nil, fmt.Errorf("mlx: load model: %w", err)
	}

	slog.Info("mlx: model loaded",
		"layers", m.NumLayers(),
		"memory_mb", mlx.GetActiveMemory()/1024/1024,
	)

	return &MLXBackend{
		model:   m,
		tok:     m.Tokenizer(),
		caches:  m.NewCache(),
		sampler: sample.New(0.1, 0, 0, 0), // default low temp
	}, nil
}

// Generate produces text from a prompt using native Metal inference.
func (b *MLXBackend) Generate(ctx context.Context, prompt string, opts GenOpts) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Reset caches for new generation
	for _, c := range b.caches {
		c.Reset()
	}

	// Set up sampler based on opts
	temp := float32(opts.Temperature)
	if temp == 0 {
		temp = 0.1
	}
	sampler := sample.New(temp, 0, 0, 0)

	// Tokenize
	formatted := tokenizer.FormatGemmaPrompt(prompt)
	tokens := b.tok.Encode(formatted)
	input := mlx.FromValues(tokens, 1, len(tokens))

	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2048
	}

	// Generation loop
	var output []int32
	for i := 0; i < maxTokens; i++ {
		select {
		case <-ctx.Done():
			return b.tok.Decode(output), ctx.Err()
		default:
		}

		logits := b.model.Forward(input, b.caches)
		next := sampler.Sample(logits)
		mlx.Materialize(next)

		nextToken := int32(next.Int())
		if nextToken == b.tok.EOSToken() {
			break
		}
		output = append(output, nextToken)
		input = mlx.FromValues([]int32{nextToken}, 1, 1)
	}

	return b.tok.Decode(output), nil
}

// Chat formats messages and generates a response.
func (b *MLXBackend) Chat(ctx context.Context, messages []Message, opts GenOpts) (string, error) {
	// Format as Gemma chat
	var prompt string
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			prompt += fmt.Sprintf("<start_of_turn>user\n%s<end_of_turn>\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("<start_of_turn>model\n%s<end_of_turn>\n", msg.Content)
		case "system":
			prompt += fmt.Sprintf("<start_of_turn>user\n[System: %s]<end_of_turn>\n", msg.Content)
		}
	}
	prompt += "<start_of_turn>model\n"

	// Use raw prompt (already formatted)
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.caches {
		c.Reset()
	}

	temp := float32(opts.Temperature)
	if temp == 0 {
		temp = 0.1
	}
	sampler := sample.New(temp, 0, 0, 0)

	tokens := b.tok.Encode(prompt)
	input := mlx.FromValues(tokens, 1, len(tokens))

	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2048
	}

	var output []int32
	for i := 0; i < maxTokens; i++ {
		select {
		case <-ctx.Done():
			return b.tok.Decode(output), ctx.Err()
		default:
		}

		logits := b.model.Forward(input, b.caches)
		next := sampler.Sample(logits)
		mlx.Materialize(next)

		nextToken := int32(next.Int())
		if nextToken == b.tok.EOSToken() {
			break
		}
		output = append(output, nextToken)
		input = mlx.FromValues([]int32{nextToken}, 1, 1)
	}

	return b.tok.Decode(output), nil
}

// Name returns the backend identifier.
func (b *MLXBackend) Name() string { return "mlx" }

// Available reports whether Metal GPU is ready.
func (b *MLXBackend) Available() bool { return mlx.MetalAvailable() }
