package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OpenRouterService handles LLM API calls via OpenRouter.
type OpenRouterService struct {
	apiKey      string
	baseURL     string
	model       string
	temperature float64
	maxTokens   int
	timeout     time.Duration
	maxRetries  int
	client      *http.Client
}

// NewOpenRouterService creates a new OpenRouterService.
func NewOpenRouterService(apiKey, baseURL, model string, temperature float64, maxTokens int, timeout time.Duration, maxRetries int) *OpenRouterService {
	return &OpenRouterService{
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		timeout:     timeout,
		maxRetries:  maxRetries,
		client:      &http.Client{Timeout: timeout},
	}
}

// ChatMessage represents a message in the chat.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends a chat completion request to OpenRouter.
func (s *OpenRouterService) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := openRouterRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: s.temperature,
		MaxTokens:   s.maxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second) // backoff
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.apiKey)

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(respBody))
			continue
		}

		var result openRouterResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			lastErr = fmt.Errorf("parsing response: %w", err)
			continue
		}

		if result.Error != nil {
			lastErr = fmt.Errorf("OpenRouter error: %s", result.Error.Message)
			continue
		}

		if len(result.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			continue
		}

		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("OpenRouter request failed after %d retries: %w", s.maxRetries, lastErr)
}

// ImportanceAnalysis represents the result of importance analysis.
type ImportanceAnalysis struct {
	Category   string `json:"category"`
	Importance string `json:"importance"`
	Summary    string `json:"summary"`
}

// AnalyzeImportance classifies content category and importance using LLM.
func (s *OpenRouterService) AnalyzeImportance(ctx context.Context, content string) (*ImportanceAnalysis, error) {
	prompt := fmt.Sprintf(`Analyze the following content and classify it. Respond in JSON format only, with no additional text.

IMPORTANT RULES for the summary:
- Use full entity names (NO pronouns like "he", "she", "it", "they")
- Use absolute dates in ISO 8601 format (NO relative references like "yesterday", "last week")
- The summary must be self-contained and understandable without any surrounding context
- Write a single lossless restatement of the key facts

{
  "category": "one of: INSIGHT, DECISION, WARNING, KNOWLEDGE, ACTION, CONTEXT, REFERENCE",
  "importance": "one of: CRITICAL, IMPORTANT, MINOR",
  "summary": "brief self-contained summary in up to 100 words with no pronouns and absolute dates"
}

Content:
%s`, content)

	response, err := s.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a content classifier. Always respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	var result ImportanceAnalysis
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		slog.Warn("failed to parse importance analysis, using defaults", "error", err)
		return &ImportanceAnalysis{
			Category:   "KNOWLEDGE",
			Importance: "MINOR",
			Summary:    truncate(content, 200),
		}, nil
	}

	return &result, nil
}

// RelevanceAnalysis represents the result of relevance analysis.
type RelevanceAnalysis struct {
	Relevant   bool    `json:"relevant"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// AnalyzeRelevance determines if memories are relevant to a prompt.
func (s *OpenRouterService) AnalyzeRelevance(ctx context.Context, prompt string, memorySummaries []string) (*RelevanceAnalysis, error) {
	systemPrompt := `You are a relevance analyzer. Given a user prompt and memory summaries, determine if any memories are relevant. Respond in JSON format only:
{
  "relevant": true/false,
  "confidence": 0.0-1.0,
  "reasoning": "brief explanation"
}`

	userPrompt := fmt.Sprintf("User prompt: %s\n\nMemory summaries:\n", prompt)
	for i, s := range memorySummaries {
		userPrompt += fmt.Sprintf("%d. %s\n", i+1, s)
	}

	response, err := s.Chat(ctx, []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return nil, err
	}

	var result RelevanceAnalysis
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		return &RelevanceAnalysis{Relevant: false, Confidence: 0, Reasoning: "parse error"}, nil
	}

	return &result, nil
}

// EntityExtractionResult represents extracted entities and relationships.
type EntityExtractionResult struct {
	Entities      []ExtractedEntity       `json:"entities"`
	Relationships []ExtractedRelationship `json:"relationships"`
}

// ExtractedEntity represents an entity extracted from text.
type ExtractedEntity struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
}

// ExtractedRelationship represents a relationship between extracted entities.
type ExtractedRelationship struct {
	Source     string            `json:"source"`
	Target    string            `json:"target"`
	Type      string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
}

// ExtractEntities extracts entities and relationships from text using LLM.
func (s *OpenRouterService) ExtractEntities(ctx context.Context, content string) (*EntityExtractionResult, error) {
	prompt := fmt.Sprintf(`Extract entities and relationships from the following text. Respond in JSON format only.

IMPORTANT RULES:
- Use full entity names (never abbreviations or pronouns)
- Entity names must be complete and unambiguous
- If the text mentions dates, convert to ISO 8601 format in the properties

{
  "entities": [{"name": "full entity name", "type": "TECHNOLOGY|PERSON|PROJECT|CONCEPT|LIBRARY|LANGUAGE|TOOL|SERVICE", "properties": {}}],
  "relationships": [{"source": "full entity1 name", "target": "full entity2 name", "type": "USES|DEPENDS_ON|IMPLEMENTS|EXTENDS|RELATED_TO", "properties": {}}]
}

Text:
%s`, content)

	response, err := s.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are an entity extraction system. Always respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	var result EntityExtractionResult
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		return &EntityExtractionResult{}, nil
	}

	return &result, nil
}

// cleanJSON attempts to extract JSON from a response that may contain markdown code blocks.
func cleanJSON(s string) string {
	// Remove markdown code blocks
	if idx := findSubstring(s, "```json"); idx >= 0 {
		s = s[idx+7:]
	} else if idx := findSubstring(s, "```"); idx >= 0 {
		s = s[idx+3:]
	}
	if idx := findSubstring(s, "```"); idx >= 0 {
		s = s[:idx]
	}

	// Find first { and last }
	start := findSubstring(s, "{")
	end := lastIndex(s, "}")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}

	return s
}

func findSubstring(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func lastIndex(s, sub string) int {
	for i := len(s) - len(sub); i >= 0; i-- {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
