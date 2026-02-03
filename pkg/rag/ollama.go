package rag

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// OllamaConfig holds Ollama connection configuration.
type OllamaConfig struct {
	Host  string
	Port  int
	Model string
}

// DefaultOllamaConfig returns default Ollama configuration.
// Host defaults to localhost for local development.
func DefaultOllamaConfig() OllamaConfig {
	return OllamaConfig{
		Host:  "localhost",
		Port:  11434,
		Model: "nomic-embed-text",
	}
}

// OllamaClient wraps the Ollama API client for embeddings.
type OllamaClient struct {
	client *api.Client
	config OllamaConfig
}

// NewOllamaClient creates a new Ollama client.
func NewOllamaClient(cfg OllamaConfig) (*OllamaClient, error) {
	baseURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}

	client := api.NewClient(baseURL, http.DefaultClient)

	return &OllamaClient{
		client: client,
		config: cfg,
	}, nil
}

// EmbedDimension returns the embedding dimension for the configured model.
// nomic-embed-text uses 768 dimensions.
func (o *OllamaClient) EmbedDimension() uint64 {
	switch o.config.Model {
	case "nomic-embed-text":
		return 768
	case "mxbai-embed-large":
		return 1024
	case "all-minilm":
		return 384
	default:
		return 768 // Default to nomic-embed-text dimension
	}
}

// Embed generates embeddings for the given text.
func (o *OllamaClient) Embed(ctx context.Context, text string) ([]float32, error) {
	req := &api.EmbedRequest{
		Model: o.config.Model,
		Input: text,
	}

	resp, err := o.client.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Embeddings) == 0 || len(resp.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	// Convert float64 to float32 for Qdrant
	embedding := resp.Embeddings[0]
	result := make([]float32, len(embedding))
	for i, v := range embedding {
		result[i] = float32(v)
	}

	return result, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (o *OllamaClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := o.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		results[i] = embedding
	}
	return results, nil
}

// VerifyModel checks if the embedding model is available.
func (o *OllamaClient) VerifyModel(ctx context.Context) error {
	_, err := o.Embed(ctx, "test")
	if err != nil {
		return fmt.Errorf("model %s not available: %w (run: ollama pull %s)", o.config.Model, err, o.config.Model)
	}
	return nil
}

// Model returns the configured embedding model name.
func (o *OllamaClient) Model() string {
	return o.config.Model
}
