package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CoreferenceService resolves pronouns and aliases in free text to canonical
// entity labels BEFORE entity/triplet extraction. Reduces hallucination and
// fragmentation of otherwise-identical entities across consecutive passes.
type CoreferenceService struct {
	llm LLMProvider
}

// NewCoreferenceService wires the LLM.
func NewCoreferenceService(llm LLMProvider) *CoreferenceService {
	return &CoreferenceService{llm: llm}
}

// CorefResult holds the rewritten text and the resolution map applied.
type CorefResult struct {
	Original    string            `json:"original"`
	Resolved    string            `json:"resolved"`
	Resolutions map[string]string `json:"resolutions"`
}

// Resolve rewrites the text replacing pronouns/aliases with canonical labels.
// When the LLM is unavailable or the payload is malformed, returns the input
// unchanged so upstream extraction continues with best-effort behaviour.
func (s *CoreferenceService) Resolve(ctx context.Context, text string) (*CorefResult, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return &CorefResult{Original: text, Resolved: text}, nil
	}
	if s == nil || s.llm == nil {
		return &CorefResult{Original: text, Resolved: text}, nil
	}

	prompt := fmt.Sprintf(`Resolve coreferences in the text. Return ONLY JSON with:
  { "resolved": "...text with pronouns and aliases replaced by canonical labels...",
    "resolutions": { "<original span>": "<canonical label>", ... } }

Rules:
- Keep the original voice, tense, and length.
- Replace every pronoun and nickname with its canonical referent when known.
- If a referent is ambiguous, leave the pronoun unchanged.

Text:
"""
%s
"""`, trimmed)

	raw, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You resolve coreferences and normalise aliases. JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return &CorefResult{Original: text, Resolved: text}, err
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return &CorefResult{Original: text, Resolved: text}, nil
	}
	var payload struct {
		Resolved    string            `json:"resolved"`
		Resolutions map[string]string `json:"resolutions"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return &CorefResult{Original: text, Resolved: text}, nil
	}
	if payload.Resolved == "" {
		payload.Resolved = text
	}
	return &CorefResult{Original: text, Resolved: payload.Resolved, Resolutions: payload.Resolutions}, nil
}
