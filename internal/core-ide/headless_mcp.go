package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"forge.lthn.ai/core/cli/pkg/jobrunner"
)

// startHeadlessMCP starts a minimal MCP HTTP server for headless mode.
// It exposes job handler tools and health endpoints.
func startHeadlessMCP(poller *jobrunner.Poller) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"mode":   "headless",
			"cycle":  poller.Cycle(),
		})
	})

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"name":    "core-ide",
			"version": "0.1.0",
			"mode":    "headless",
		})
	})

	mux.HandleFunc("/mcp/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tools := []map[string]string{
			{"name": "job_status", "description": "Get poller status (cycle count, dry-run)"},
			{"name": "job_set_dry_run", "description": "Enable/disable dry-run mode"},
			{"name": "job_run_once", "description": "Trigger a single poll-dispatch cycle"},
		}
		json.NewEncoder(w).Encode(map[string]any{"tools": tools})
	})

	mux.HandleFunc("/mcp/call", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Tool   string         `json:"tool"`
			Params map[string]any `json:"params"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		switch req.Tool {
		case "job_status":
			json.NewEncoder(w).Encode(map[string]any{
				"cycle":   poller.Cycle(),
				"dry_run": poller.DryRun(),
			})
		case "job_set_dry_run":
			if v, ok := req.Params["enabled"].(bool); ok {
				poller.SetDryRun(v)
			}
			json.NewEncoder(w).Encode(map[string]any{"dry_run": poller.DryRun()})
		case "job_run_once":
			err := poller.RunOnce(context.Background())
			json.NewEncoder(w).Encode(map[string]any{
				"success": err == nil,
				"cycle":   poller.Cycle(),
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{"error": "unknown tool"})
		}
	})

	addr := fmt.Sprintf("127.0.0.1:%d", mcpPort)
	log.Printf("Headless MCP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Headless MCP server error: %v", err)
	}
}
