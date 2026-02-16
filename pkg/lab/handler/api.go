package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"forge.lthn.ai/core/cli/pkg/lab"
)

type APIHandler struct {
	store *lab.Store
}

func NewAPIHandler(s *lab.Store) *APIHandler {
	return &APIHandler{store: s}
}

type apiResponse struct {
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      any       `json:"data"`
}

func (h *APIHandler) writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiResponse{
		Status:    "ok",
		UpdatedAt: time.Now(),
		Data:      data,
	})
}

func (h *APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.Overview())
}

func (h *APIHandler) Models(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.GetModels())
}

func (h *APIHandler) Training(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.GetTraining())
}

func (h *APIHandler) Agents(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.GetAgents())
}

func (h *APIHandler) Services(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.GetServices())
}

func (h *APIHandler) GoldenSet(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.GetGoldenSet())
}

func (h *APIHandler) Runs(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, h.store.GetBenchmarks())
}

func (h *APIHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
