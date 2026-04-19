package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAnthropicProvider_Name(t *testing.T) {
	p := NewAnthropicProvider(DefaultAnthropicConfig("k"))
	if p.Name() != "anthropic" {
		t.Errorf("expected 'anthropic', got %q", p.Name())
	}
}

func TestAnthropicProvider_RequiresAPIKey(t *testing.T) {
	p := NewAnthropicProvider(AnthropicConfig{})
	_, err := p.Chat(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})
	if err == nil {
		t.Error("expected error when API key is missing")
	}
}

func TestAnthropicProvider_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			http.Error(w, "missing x-api-key", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("anthropic-version") == "" {
			http.Error(w, "missing version", http.StatusBadRequest)
			return
		}

		// Verify body has system and messages
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["system"] != "You are helpful." {
			http.Error(w, "wrong system", http.StatusBadRequest)
			return
		}

		resp := map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": "Hello world!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:  "test-key",
		BaseURL: srv.URL,
		Model:   "test-model",
		Timeout: 5 * time.Second,
	})

	text, err := p.Chat(context.Background(), []ChatMessage{
		{Role: "system", Content: "You are helpful."},
		{Role: "user", Content: "Hi"},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if text != "Hello world!" {
		t.Errorf("unexpected response: %q", text)
	}
}

func TestAnthropicProvider_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"type":"rate_limit","message":"too many"}}`))
	}))
	defer srv.Close()

	p := NewAnthropicProvider(AnthropicConfig{
		APIKey:  "k",
		BaseURL: srv.URL,
		Timeout: 5 * time.Second,
	})

	_, err := p.Chat(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})
	if err == nil {
		t.Error("expected error on 429")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("expected status in error, got %v", err)
	}
}

func TestGeminiProvider_Name(t *testing.T) {
	p := NewGeminiProvider(DefaultGeminiConfig("k"))
	if p.Name() != "gemini" {
		t.Errorf("expected 'gemini', got %q", p.Name())
	}
}

func TestGeminiProvider_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key in query
		if r.URL.Query().Get("key") != "test-key" {
			http.Error(w, "missing key", http.StatusUnauthorized)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)

		// Verify systemInstruction present
		if _, ok := req["systemInstruction"]; !ok {
			http.Error(w, "missing systemInstruction", http.StatusBadRequest)
			return
		}

		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"role":  "model",
						"parts": []map[string]any{{"text": "Hi from Gemini"}},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := NewGeminiProvider(GeminiConfig{
		APIKey:  "test-key",
		BaseURL: srv.URL,
		Model:   "gemini-2.0-flash",
		Timeout: 5 * time.Second,
	})

	text, err := p.Chat(context.Background(), []ChatMessage{
		{Role: "system", Content: "Be concise."},
		{Role: "user", Content: "Hello"},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if text != "Hi from Gemini" {
		t.Errorf("unexpected response: %q", text)
	}
}

func TestGeminiProvider_RequiresAPIKey(t *testing.T) {
	p := NewGeminiProvider(GeminiConfig{})
	_, err := p.Chat(context.Background(), []ChatMessage{{Role: "user", Content: "x"}})
	if err == nil {
		t.Error("expected error when API key is missing")
	}
}

func TestJoinSystemParts(t *testing.T) {
	if got := joinSystemParts([]string{"a"}); got != "a" {
		t.Errorf("single: got %q", got)
	}
	if got := joinSystemParts([]string{"a", "b"}); got != "a\n\nb" {
		t.Errorf("multi: got %q", got)
	}
}

func TestAnthropicProvider_Defaults(t *testing.T) {
	p := NewAnthropicProvider(AnthropicConfig{APIKey: "k"})
	if p.config.BaseURL == "" {
		t.Error("BaseURL default missing")
	}
	if p.config.Model == "" {
		t.Error("Model default missing")
	}
	if p.config.MaxTokens <= 0 {
		t.Error("MaxTokens default missing")
	}
	if p.config.Timeout == 0 {
		t.Error("Timeout default missing")
	}
}

// Silence unused imports.
var _ = bytes.NewReader
