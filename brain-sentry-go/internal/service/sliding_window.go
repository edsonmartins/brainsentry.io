package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// SlidingWindowEnrichment enriches memories by resolving entities, extracting preferences,
// and adding contextual bridges to make each memory self-contained for retrieval.
type SlidingWindowEnrichment struct {
	llm LLMProvider
}

// NewSlidingWindowEnrichment creates a new SlidingWindowEnrichment.
func NewSlidingWindowEnrichment(llm LLMProvider) *SlidingWindowEnrichment {
	return &SlidingWindowEnrichment{llm: llm}
}

// EnrichmentResult holds the result of enriching a piece of content.
type EnrichmentResult struct {
	EnrichedContent string   `json:"enrichedContent"`
	ResolvedEntities []string `json:"resolvedEntities,omitempty"` // e.g., "it" → "PostgreSQL"
	Preferences     []string `json:"preferences,omitempty"`       // e.g., "prefers Go over Java"
	ContextBridges  []string `json:"contextBridges,omitempty"`    // links to adjacent context
	Changed         bool     `json:"changed"`
}

const enrichmentPrompt = `You are a content enrichment engine. Given a piece of text and optional surrounding context, perform these enrichments:

1. **Entity Resolution**: Replace pronouns (he, she, it, they, this, that) with explicit names from context.
2. **Preference Extraction**: Identify user constraints, opinions, or preferences stated.
3. **Self-Containment**: Ensure the text is understandable without surrounding context. Add brief clarifications if needed.

Rules:
- Do NOT add information that isn't in the original text or context.
- Do NOT change the meaning or tone.
- Convert relative dates to ISO 8601 if the current date is provided.
- Keep the enriched text concise.

Respond with valid JSON only:
{
  "enrichedContent": "the rewritten text with entities resolved",
  "resolvedEntities": ["pronoun → resolved name"],
  "preferences": ["preference statement"],
  "contextBridges": ["brief context link added"]
}`

// Enrich processes a piece of content with surrounding context to make it self-contained.
func (s *SlidingWindowEnrichment) Enrich(ctx context.Context, content string, prevContext string, nextContext string) (*EnrichmentResult, error) {
	if s.llm == nil {
		return &EnrichmentResult{EnrichedContent: content, Changed: false}, nil
	}

	// Skip if content is already self-contained (short, no pronouns)
	if !needsEnrichment(content) {
		return &EnrichmentResult{EnrichedContent: content, Changed: false}, nil
	}

	var contextParts []string
	if prevContext != "" {
		contextParts = append(contextParts, fmt.Sprintf("Previous context: %s", truncateForLLM(prevContext, 500)))
	}
	contextParts = append(contextParts, fmt.Sprintf("Content to enrich: %s", truncateForLLM(content, 2000)))
	if nextContext != "" {
		contextParts = append(contextParts, fmt.Sprintf("Next context: %s", truncateForLLM(nextContext, 500)))
	}

	response, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: enrichmentPrompt},
		{Role: "user", Content: strings.Join(contextParts, "\n\n")},
	})
	if err != nil {
		slog.Warn("sliding window enrichment failed", "error", err)
		return &EnrichmentResult{EnrichedContent: content, Changed: false}, nil
	}

	var result EnrichmentResult
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		slog.Warn("sliding window enrichment parse failed", "error", err)
		return &EnrichmentResult{EnrichedContent: content, Changed: false}, nil
	}

	if result.EnrichedContent == "" {
		result.EnrichedContent = content
	}

	result.Changed = result.EnrichedContent != content
	return &result, nil
}

// EnrichBatch processes multiple content pieces with sliding window context.
func (s *SlidingWindowEnrichment) EnrichBatch(ctx context.Context, contents []string) ([]EnrichmentResult, error) {
	results := make([]EnrichmentResult, len(contents))

	for i, content := range contents {
		var prev, next string
		if i > 0 {
			prev = contents[i-1]
		}
		if i < len(contents)-1 {
			next = contents[i+1]
		}

		result, err := s.Enrich(ctx, content, prev, next)
		if err != nil {
			results[i] = EnrichmentResult{EnrichedContent: content, Changed: false}
			continue
		}
		results[i] = *result
	}

	return results, nil
}

// needsEnrichment checks if content contains pronouns or relative references.
func needsEnrichment(content string) bool {
	lower := strings.ToLower(content)

	// Common pronouns that suggest unresolved references
	pronouns := []string{" it ", " they ", " he ", " she ", " this ", " that ", " these ", " those "}
	for _, p := range pronouns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	// Relative time references
	timeRefs := []string{"yesterday", "last week", "last month", "today", "tomorrow", "next week"}
	for _, t := range timeRefs {
		if strings.Contains(lower, t) {
			return true
		}
	}

	return false
}
