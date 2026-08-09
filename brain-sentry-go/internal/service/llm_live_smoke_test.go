//go:build llm_smoke

package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLiveLLMProvider(t *testing.T) {
	provider := configuredLiveProvider(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	const marker = "BRAINSENTRY_SMOKE_OK"
	response, err := provider.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "Follow the user's output instruction exactly."},
		{Role: "user", Content: "Reply with exactly: " + marker},
	})
	if err != nil {
		t.Fatalf("%s live smoke failed: %v", provider.Name(), err)
	}
	if !strings.Contains(response, marker) {
		t.Fatalf("%s returned an unexpected response: %q", provider.Name(), response)
	}
}

func configuredLiveProvider(t *testing.T) LLMProvider {
	t.Helper()
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		baseURL := envOr("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")
		model := envOr("OPENROUTER_MODEL", "openai/gpt-4o-mini")
		return NewOpenRouterProvider(NewOpenRouterService(key, baseURL, model, 0, 32, 60*time.Second, 0))
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg := DefaultAnthropicConfig(key)
		cfg.Model = envOr("ANTHROPIC_MODEL", cfg.Model)
		cfg.MaxTokens = 32
		cfg.Temperature = 0
		return NewAnthropicProvider(cfg)
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg := DefaultGeminiConfig(key)
		cfg.Model = envOr("GEMINI_MODEL", cfg.Model)
		cfg.MaxTokens = 32
		cfg.Temperature = 0
		return NewGeminiProvider(cfg)
	}
	t.Fatal("no live LLM credential configured: set OPENROUTER_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY")
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
