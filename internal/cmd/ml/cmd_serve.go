package ml

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var serveCmd = &cli.Command{
	Use:   "serve",
	Short: "Start OpenAI-compatible inference server",
	Long:  "Starts an HTTP server serving /v1/completions and /v1/chat/completions using the configured ML backend.",
	RunE:  runServe,
}

var (
	serveBind      string
	serveModelPath string
)

func init() {
	serveCmd.Flags().StringVar(&serveBind, "bind", "0.0.0.0:8090", "Address to bind")
	serveCmd.Flags().StringVar(&serveModelPath, "model-path", "", "Path to model directory (for mlx backend)")
}

type completionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type completionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []completionChoice `json:"choices"`
	Usage   usageInfo          `json:"usage"`
}

type completionChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type chatRequest struct {
	Model       string       `json:"model"`
	Messages    []ml.Message `json:"messages"`
	MaxTokens   int          `json:"max_tokens"`
	Temperature float64      `json:"temperature"`
}

type chatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message      ml.Message `json:"message"`
	Index        int        `json:"index"`
	FinishReason string     `json:"finish_reason"`
}

type usageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func runServe(cmd *cli.Command, args []string) error {
	// Create a backend — use HTTP backend pointing to configured API URL.
	// On macOS with MLX build tag, this will use the native MLX backend instead.
	backend := ml.NewHTTPBackend(apiURL, modelName)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req completionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		opts := ml.GenOpts{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			Model:       req.Model,
		}

		text, err := backend.Generate(r.Context(), req.Prompt, opts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		resp := completionResponse{
			ID:      fmt.Sprintf("cmpl-%d", time.Now().UnixNano()),
			Object:  "text_completion",
			Created: time.Now().Unix(),
			Model:   backend.Name(),
			Choices: []completionChoice{{Text: text, FinishReason: "stop"}},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req chatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		opts := ml.GenOpts{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			Model:       req.Model,
		}

		text, err := backend.Chat(r.Context(), req.Messages, opts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		resp := chatResponse{
			ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   backend.Name(),
			Choices: []chatChoice{{
				Message:      ml.Message{Role: "assistant", Content: text},
				FinishReason: "stop",
			}},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("GET /v1/models", func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Object string `json:"object"`
			Data   []struct {
				ID string `json:"id"`
			} `json:"data"`
		}{
			Object: "list",
			Data: []struct {
				ID string `json:"id"`
			}{{ID: backend.Name()}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	slog.Info("ml serve: starting", "bind", serveBind, "backend", backend.Name())
	fmt.Printf("Serving on http://%s\n", serveBind)
	return http.ListenAndServe(serveBind, mux)
}
