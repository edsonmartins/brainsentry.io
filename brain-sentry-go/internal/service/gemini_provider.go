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

// GeminiConfig configures the native Google Gemini provider.
type GeminiConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
	Temperature float64
	Timeout    time.Duration
}

// DefaultGeminiConfig returns sensible defaults.
func DefaultGeminiConfig(apiKey string) GeminiConfig {
	return GeminiConfig{
		APIKey:      apiKey,
		BaseURL:     "https://generativelanguage.googleapis.com",
		Model:       "gemini-2.0-flash",
		MaxTokens:   4096,
		Temperature: 0.7,
		Timeout:     60 * time.Second,
	}
}

// GeminiProvider implements LLMProvider for Google Gemini via the native REST API.
type GeminiProvider struct {
	config GeminiConfig
	client *http.Client
}

// NewGeminiProvider creates a new GeminiProvider.
func NewGeminiProvider(config GeminiConfig) *GeminiProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com"
	}
	if config.Model == "" {
		config.Model = "gemini-2.0-flash"
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	return &GeminiProvider{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// Name returns the provider identifier.
func (p *GeminiProvider) Name() string { return "gemini" }

type geminiContent struct {
	Role  string       `json:"role"` // "user" or "model"
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent      `json:"systemInstruction,omitempty"`
	Contents          []geminiContent     `json:"contents"`
	GenerationConfig  *geminiGenerationCfg `json:"generationConfig,omitempty"`
}

type geminiGenerationCfg struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends a chat completion request to Gemini.
func (p *GeminiProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("gemini: API key not set")
	}

	var systemParts []string
	contents := make([]geminiContent, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			systemParts = append(systemParts, m.Content)
			continue
		}
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	reqBody := geminiRequest{
		Contents: contents,
		GenerationConfig: &geminiGenerationCfg{
			Temperature:     p.config.Temperature,
			MaxOutputTokens: p.config.MaxTokens,
		},
	}
	if len(systemParts) > 0 {
		reqBody.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: joinSystemParts(systemParts)}},
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini: marshal: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s",
		p.config.BaseURL, p.config.Model, p.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini: read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result geminiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("gemini: parse: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("gemini api error: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}

	var text string
	for _, p := range result.Candidates[0].Content.Parts {
		text += p.Text
	}
	return text, nil
}
