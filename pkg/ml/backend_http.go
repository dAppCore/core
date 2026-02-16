package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"forge.lthn.ai/core/go/pkg/log"
)

// HTTPBackend talks to an OpenAI-compatible chat completions API.
type HTTPBackend struct {
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// chatRequest is the request body for /v1/chat/completions.
type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// chatChoice is a single completion choice.
type chatChoice struct {
	Message Message `json:"message"`
}

// chatResponse is the response from /v1/chat/completions.
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

// retryableError marks errors that should be retried.
type retryableError struct {
	err error
}

func (e *retryableError) Error() string { return e.err.Error() }
func (e *retryableError) Unwrap() error { return e.err }

// NewHTTPBackend creates an HTTPBackend for the given base URL and model.
func NewHTTPBackend(baseURL, model string) *HTTPBackend {
	return &HTTPBackend{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// Name returns "http".
func (b *HTTPBackend) Name() string { return "http" }

// Available always returns true for HTTP backends.
func (b *HTTPBackend) Available() bool { return b.baseURL != "" }

// Model returns the configured model name.
func (b *HTTPBackend) Model() string { return b.model }

// BaseURL returns the configured base URL.
func (b *HTTPBackend) BaseURL() string { return b.baseURL }

// SetMaxTokens sets the maximum token count for requests.
func (b *HTTPBackend) SetMaxTokens(n int) { b.maxTokens = n }

// Generate sends a single prompt and returns the response.
func (b *HTTPBackend) Generate(ctx context.Context, prompt string, opts GenOpts) (string, error) {
	return b.Chat(ctx, []Message{{Role: "user", Content: prompt}}, opts)
}

// Chat sends a multi-turn conversation and returns the response.
// Retries up to 3 times with exponential backoff on transient failures.
func (b *HTTPBackend) Chat(ctx context.Context, messages []Message, opts GenOpts) (string, error) {
	model := b.model
	if opts.Model != "" {
		model = opts.Model
	}
	maxTokens := b.maxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}
	temp := opts.Temperature

	req := chatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temp,
		MaxTokens:   maxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", log.E("ml.HTTPBackend.Chat", "marshal request", err)
	}

	const maxAttempts = 3
	var lastErr error

	for attempt := range maxAttempts {
		if attempt > 0 {
			backoff := time.Duration(100<<uint(attempt-1)) * time.Millisecond
			time.Sleep(backoff)
		}

		result, err := b.doRequest(ctx, body)
		if err == nil {
			return result, nil
		}
		lastErr = err

		var re *retryableError
		if !errors.As(err, &re) {
			return "", err
		}
	}

	return "", log.E("ml.HTTPBackend.Chat", fmt.Sprintf("exhausted %d retries", maxAttempts), lastErr)
}

// doRequest sends a single HTTP request and parses the response.
func (b *HTTPBackend) doRequest(ctx context.Context, body []byte) (string, error) {
	url := b.baseURL + "/v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return "", &retryableError{fmt.Errorf("http request: %w", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &retryableError{fmt.Errorf("read response: %w", err)}
	}

	if resp.StatusCode >= 500 {
		return "", &retryableError{fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))}
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}
