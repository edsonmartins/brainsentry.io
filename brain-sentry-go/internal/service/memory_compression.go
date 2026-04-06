package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/integraltech/brainsentry/internal/domain"
)

// CompressedMemoryData holds structured data extracted from memory content via LLM.
type CompressedMemoryData struct {
	Facts      []string `json:"facts"`
	Concepts   []string `json:"concepts"`
	Narrative  string   `json:"narrative"`
	Importance int      `json:"importance"` // 1-10
	Title      string   `json:"title"`
	Files      []string `json:"files,omitempty"`
}

// MemoryCompressionService extracts structured data from raw memory content using LLM.
type MemoryCompressionService struct {
	llm        LLMProvider
	maxRetries int
}

// NewMemoryCompressionService creates a new MemoryCompressionService.
func NewMemoryCompressionService(llm LLMProvider) *MemoryCompressionService {
	return &MemoryCompressionService{
		llm:        llm,
		maxRetries: 2,
	}
}

const compressionSystemPrompt = `You are a memory compression engine. Given raw content, extract structured information.
Respond with valid JSON only, no additional text.

Rules:
- "facts": Array of atomic, self-contained factual statements. Use full entity names (no pronouns). Use ISO 8601 dates.
- "concepts": Array of key concepts, technologies, patterns, or domain terms mentioned.
- "narrative": A 1-3 sentence summary capturing the essential meaning. Self-contained, no pronouns, absolute dates.
- "importance": Integer 1-10 scale. 10=critical system decision, 1=trivial observation.
- "title": Short descriptive title (5-10 words).
- "files": Array of file paths mentioned (empty if none).

Output format:
{
  "facts": ["fact1", "fact2"],
  "concepts": ["concept1", "concept2"],
  "narrative": "concise summary",
  "importance": 7,
  "title": "short title",
  "files": []
}`

// Compress extracts structured data from memory content.
func (s *MemoryCompressionService) Compress(ctx context.Context, content string) (*CompressedMemoryData, error) {
	if s.llm == nil {
		return s.fallbackCompress(content), nil
	}

	userPrompt := fmt.Sprintf("Content to compress:\n\n%s", truncateForLLM(content, 4000))

	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		response, err := s.llm.Chat(ctx, []ChatMessage{
			{Role: "system", Content: compressionSystemPrompt},
			{Role: "user", Content: userPrompt},
		})
		if err != nil {
			lastErr = err
			continue
		}

		var data CompressedMemoryData
		if err := json.Unmarshal([]byte(cleanJSON(response)), &data); err != nil {
			slog.Warn("compression parse failed, retrying",
				"attempt", attempt+1,
				"error", err,
			)
			// Self-correcting: retry with feedback
			userPrompt = fmt.Sprintf("Your previous response was invalid JSON. Error: %s\n\nPlease try again with valid JSON.\n\nContent to compress:\n\n%s",
				err.Error(), truncateForLLM(content, 3500))
			lastErr = err
			continue
		}

		// Validate
		if err := s.validate(&data); err != nil {
			slog.Warn("compression validation failed, retrying",
				"attempt", attempt+1,
				"error", err,
			)
			userPrompt = fmt.Sprintf("Your response had validation errors: %s\n\nPlease fix and try again.\n\nContent to compress:\n\n%s",
				err.Error(), truncateForLLM(content, 3500))
			lastErr = err
			continue
		}

		return &data, nil
	}

	slog.Warn("compression failed, using fallback", "error", lastErr)
	return s.fallbackCompress(content), nil
}

func (s *MemoryCompressionService) validate(data *CompressedMemoryData) error {
	if data.Narrative == "" {
		return fmt.Errorf("narrative is required")
	}
	if data.Importance < 1 || data.Importance > 10 {
		return fmt.Errorf("importance must be 1-10, got %d", data.Importance)
	}
	if len(data.Facts) == 0 {
		return fmt.Errorf("at least one fact is required")
	}
	return nil
}

// fallbackCompress creates a basic CompressedMemoryData without LLM.
func (s *MemoryCompressionService) fallbackCompress(content string) *CompressedMemoryData {
	words := strings.Fields(content)

	// Simple title from first words
	title := content
	if len(title) > 60 {
		title = title[:57] + "..."
	}

	// Extract potential concepts (capitalized multi-char words)
	var concepts []string
	seen := make(map[string]bool)
	for _, w := range words {
		clean := strings.Trim(w, ".,;:!?()[]{}\"'")
		if len(clean) > 2 && clean[0] >= 'A' && clean[0] <= 'Z' && !seen[clean] {
			concepts = append(concepts, clean)
			seen[clean] = true
		}
	}

	// Narrative = first 200 chars
	narrative := content
	if len(narrative) > 200 {
		narrative = narrative[:197] + "..."
	}

	return &CompressedMemoryData{
		Facts:      []string{narrative},
		Concepts:   concepts,
		Narrative:  narrative,
		Importance: 5,
		Title:      title,
	}
}

// EnrichMemory applies compressed data to a memory domain object.
func (s *MemoryCompressionService) EnrichMemory(m *domain.Memory, data *CompressedMemoryData) {
	if m.Summary == "" {
		m.Summary = data.Narrative
	}

	// Parse existing metadata or create new
	meta := make(map[string]any)
	if len(m.Metadata) > 0 {
		_ = json.Unmarshal(m.Metadata, &meta)
	}

	meta["facts"] = data.Facts
	meta["concepts"] = data.Concepts
	meta["compressionTitle"] = data.Title
	meta["compressionImportance"] = data.Importance

	if len(data.Files) > 0 {
		meta["files"] = data.Files
	}

	raw, err := json.Marshal(meta)
	if err == nil {
		m.Metadata = raw
	}

	// Add concepts as tags (deduplicated)
	tagSet := make(map[string]bool)
	for _, t := range m.Tags {
		tagSet[t] = true
	}
	for _, c := range data.Concepts {
		lower := strings.ToLower(c)
		if !tagSet[lower] && len(lower) > 1 {
			m.Tags = append(m.Tags, lower)
			tagSet[lower] = true
		}
	}
}

func truncateForLLM(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n...[truncated]"
}
