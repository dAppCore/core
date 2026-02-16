package ml

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPBackend_Generate_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if len(req.Messages) != 1 || req.Messages[0].Content != "hello" {
			t.Errorf("unexpected messages: %+v", req.Messages)
		}

		resp := chatResponse{
			Choices: []chatChoice{{Message: Message{Role: "assistant", Content: "world"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	b := NewHTTPBackend(srv.URL, "test-model")
	result, err := b.Generate(context.Background(), "hello", DefaultGenOpts())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if result != "world" {
		t.Errorf("got %q, want %q", result, "world")
	}
}

func TestHTTPBackend_Generate_Bad(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	b := NewHTTPBackend(srv.URL, "test-model")
	_, err := b.Generate(context.Background(), "hello", DefaultGenOpts())
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestHTTPBackend_Retry_Ugly(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}
		resp := chatResponse{
			Choices: []chatChoice{{Message: Message{Role: "assistant", Content: "recovered"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	b := NewHTTPBackend(srv.URL, "test-model")
	result, err := b.Generate(context.Background(), "test", DefaultGenOpts())
	if err != nil {
		t.Fatalf("Generate after retry: %v", err)
	}
	if result != "recovered" {
		t.Errorf("got %q, want %q", result, "recovered")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPBackend_Name(t *testing.T) {
	b := NewHTTPBackend("http://localhost", "model")
	if b.Name() != "http" {
		t.Errorf("Name() = %q, want %q", b.Name(), "http")
	}
}

func TestHTTPBackend_Available(t *testing.T) {
	b := NewHTTPBackend("http://localhost", "model")
	if !b.Available() {
		t.Error("Available() should be true when baseURL is set")
	}

	b2 := NewHTTPBackend("", "model")
	if b2.Available() {
		t.Error("Available() should be false when baseURL is empty")
	}
}
