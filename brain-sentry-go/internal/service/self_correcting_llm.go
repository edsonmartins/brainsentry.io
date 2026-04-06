package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// SelfCorrectingLLM wraps an LLM provider with output validation and retry with feedback.
type SelfCorrectingLLM struct {
	provider   LLMProvider
	maxRetries int
}

// NewSelfCorrectingLLM creates a new SelfCorrectingLLM wrapper.
func NewSelfCorrectingLLM(provider LLMProvider, maxRetries int) *SelfCorrectingLLM {
	if maxRetries <= 0 {
		maxRetries = 2
	}
	return &SelfCorrectingLLM{
		provider:   provider,
		maxRetries: maxRetries,
	}
}

// ChatWithValidation sends a chat request and validates the response as JSON.
// If validation fails, retries with error feedback.
func (s *SelfCorrectingLLM) ChatWithValidation(
	ctx context.Context,
	messages []ChatMessage,
	validateFn func(json.RawMessage) error,
) (json.RawMessage, error) {
	currentMessages := make([]ChatMessage, len(messages))
	copy(currentMessages, messages)

	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		response, err := s.provider.Chat(ctx, currentMessages)
		if err != nil {
			lastErr = err
			slog.Warn("self-correcting LLM: chat failed",
				"attempt", attempt+1,
				"error", err,
			)
			continue
		}

		// Try to parse as JSON
		cleaned := cleanJSON(response)
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
			lastErr = fmt.Errorf("invalid JSON: %w", err)
			slog.Warn("self-correcting LLM: invalid JSON, retrying with feedback",
				"attempt", attempt+1,
				"error", err,
			)
			// Add feedback message for retry
			currentMessages = append(currentMessages,
				ChatMessage{Role: "assistant", Content: response},
				ChatMessage{Role: "user", Content: fmt.Sprintf(
					"Your response was not valid JSON. Error: %s\nPlease respond with valid JSON only, no markdown or explanation.",
					err.Error(),
				)},
			)
			continue
		}

		// Run custom validation
		if validateFn != nil {
			if err := validateFn(raw); err != nil {
				lastErr = fmt.Errorf("validation failed: %w", err)
				slog.Warn("self-correcting LLM: validation failed, retrying with feedback",
					"attempt", attempt+1,
					"error", err,
				)
				currentMessages = append(currentMessages,
					ChatMessage{Role: "assistant", Content: response},
					ChatMessage{Role: "user", Content: fmt.Sprintf(
						"Your JSON response had validation errors: %s\nPlease fix and respond with corrected JSON only.",
						err.Error(),
					)},
				)
				continue
			}
		}

		return raw, nil
	}

	return nil, fmt.Errorf("self-correcting LLM failed after %d attempts: %w", s.maxRetries+1, lastErr)
}

// ChatJSON sends a chat request and parses the result into the provided struct.
// Retries with feedback on parse errors.
func (s *SelfCorrectingLLM) ChatJSON(ctx context.Context, messages []ChatMessage, result any) error {
	raw, err := s.ChatWithValidation(ctx, messages, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, result)
}

func (s *SelfCorrectingLLM) Name() string {
	return "self-correcting-" + s.provider.Name()
}

func (s *SelfCorrectingLLM) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	return s.provider.Chat(ctx, messages)
}
