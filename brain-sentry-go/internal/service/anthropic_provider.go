package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicConfig configures the native Anthropic API provider.
type AnthropicConfig struct {
	APIKey     string
	BaseURL    string        // default: https://api.anthropic.com
	Model      string        // default: claude-sonnet-4-6
	MaxTokens  int           // default: 4096
	Temperature float64      // default: 0.7
	Timeout    time.Duration // default: 60s
}

// DefaultAnthropicConfig returns production-ready defaults.
func DefaultAnthropicConfig(apiKey string) AnthropicConfig {
	return AnthropicConfig{
		APIKey:      apiKey,
		BaseURL:     "https://api.anthropic.com",
		Model:       "claude-sonnet-4-6",
		MaxTokens:   4096,
		Temperature: 0.7,
		Timeout:     60 * time.Second,
	}
}

// AnthropicProvider implements LLMProvider for Anthropic Claude via the
// native Messages API (no OpenRouter intermediary). Benefits: lower latency,
// access to prompt caching, and direct model features.
type AnthropicProvider struct {
	config AnthropicConfig
	client *http.Client
}

// NewAnthropicProvider creates a new AnthropicProvider.
func NewAnthropicProvider(config AnthropicConfig) *AnthropicProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4-6"
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	return &AnthropicProvider{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// Name returns the provider identifier.
func (p *AnthropicProvider) Name() string { return "anthropic" }

// anthropicMessage is the API request body.
type anthropicRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    string              `json:"system,omitempty"`
	Messages  []anthropicMessage  `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends a chat completion request. "system" role messages are merged into
// the top-level system field as Anthropic expects.
func (p *AnthropicProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("anthropic: API key not set")
	}

	// Convert to Anthropic format
	var systemParts []string
	userAssistant := make([]anthropicMessage, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			systemParts = append(systemParts, m.Content)
			continue
		}
		role := m.Role
		if role != "user" && role != "assistant" {
			role = "user"
		}
		userAssistant = append(userAssistant, anthropicMessage{Role: role, Content: m.Content})
	}

	reqBody := anthropicRequest{
		Model:       p.config.Model,
		MaxTokens:   p.config.MaxTokens,
		Messages:    userAssistant,
		Temperature: p.config.Temperature,
	}
	if len(systemParts) > 0 {
		reqBody.System = joinSystemParts(systemParts)
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("anthropic: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("anthropic: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("x-api-key", p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic: http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic: read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("anthropic: parse: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic api error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}

	// Concatenate all text blocks
	var text string
	for _, c := range result.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	return text, nil
}

func joinSystemParts(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += "\n\n" + p
	}
	return out
}
