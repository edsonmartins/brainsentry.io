package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// TripletExtractionService extracts (Subject, Predicate, Object) triplets from
// memory content via LLM. Each triplet is stored as an embeddable unit that
// enables graph-aware semantic search ("X causes Y" style queries).
type TripletExtractionService struct {
	llm        LLMProvider
	maxRetries int
}

// NewTripletExtractionService creates a new TripletExtractionService.
func NewTripletExtractionService(llm LLMProvider) *TripletExtractionService {
	return &TripletExtractionService{
		llm:        llm,
		maxRetries: 1,
	}
}

const tripletExtractionSystemPrompt = `You extract atomic factual triplets from text. A triplet is a (Subject, Predicate, Object) tuple capturing one discrete relationship.

Rules:
- Extract every distinct factual claim as a separate triplet.
- Subject and Object should be concrete nouns/entities (people, technologies, concepts, files, organizations, events).
- Predicate should be a short verb phrase describing the relationship (e.g., "uses", "depends_on", "causes", "is_part_of", "implements", "extends", "authored_by").
- Use canonical/singular forms. "PostgreSQL" not "postgres" or "PostgreSQL database".
- Never use pronouns. Resolve them from context.
- Assign confidence 0.0-1.0 based on how explicit the claim is.
- If the text has no factual relationships (pure opinion, question, greeting), return empty triplets array.

Respond with valid JSON only:
{
  "triplets": [
    {"subject": "...", "predicate": "...", "object": "...", "confidence": 0.9}
  ]
}`

// ExtractedTriplet is the raw LLM output before domain mapping.
type ExtractedTriplet struct {
	Subject    string  `json:"subject"`
	Predicate  string  `json:"predicate"`
	Object     string  `json:"object"`
	Confidence float64 `json:"confidence"`
}

type tripletExtractionResponse struct {
	Triplets []ExtractedTriplet `json:"triplets"`
}

// ExtractFromContent uses the LLM to extract triplets from content.
// Returns an empty slice (not nil) and no error if content has no triplets.
func (s *TripletExtractionService) ExtractFromContent(ctx context.Context, content string) ([]ExtractedTriplet, error) {
	if s.llm == nil {
		return nil, nil
	}
	if strings.TrimSpace(content) == "" {
		return []ExtractedTriplet{}, nil
	}

	userPrompt := fmt.Sprintf("Text:\n\n%s", truncateForLLM(content, 4000))

	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		response, err := s.llm.Chat(ctx, []ChatMessage{
			{Role: "system", Content: tripletExtractionSystemPrompt},
			{Role: "user", Content: userPrompt},
		})
		if err != nil {
			lastErr = err
			continue
		}

		var parsed tripletExtractionResponse
		if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
			slog.Warn("triplet extraction parse failed",
				"attempt", attempt+1,
				"error", err,
			)
			userPrompt = fmt.Sprintf("Your response was not valid JSON (error: %s). Return valid JSON only.\n\nText:\n\n%s",
				err.Error(), truncateForLLM(content, 3500))
			lastErr = err
			continue
		}

		// Filter out invalid triplets (empty fields).
		valid := make([]ExtractedTriplet, 0, len(parsed.Triplets))
		for _, t := range parsed.Triplets {
			if strings.TrimSpace(t.Subject) == "" ||
				strings.TrimSpace(t.Predicate) == "" ||
				strings.TrimSpace(t.Object) == "" {
				continue
			}
			if t.Confidence <= 0 || t.Confidence > 1 {
				t.Confidence = 0.5
			}
			valid = append(valid, t)
		}
		return valid, nil
	}

	return nil, fmt.Errorf("triplet extraction failed after %d attempts: %w", s.maxRetries+1, lastErr)
}

// BuildTriplets maps extracted triplets to domain.Triplet instances with
// deterministic UUID5 IDs and formatted embeddable text.
func (s *TripletExtractionService) BuildTriplets(ctx context.Context, memoryID string, extracted []ExtractedTriplet) []domain.Triplet {
	tenantID := tenant.FromContext(ctx)
	now := time.Now()

	triplets := make([]domain.Triplet, 0, len(extracted))
	seen := make(map[string]bool, len(extracted))

	for _, t := range extracted {
		id := GenerateTripletID(t.Subject, t.Predicate, t.Object).String()
		if seen[id] {
			continue // dedup within same extraction batch
		}
		seen[id] = true

		triplets = append(triplets, domain.Triplet{
			ID:             id,
			TenantID:       tenantID,
			MemoryID:       memoryID,
			Subject:        t.Subject,
			Predicate:      t.Predicate,
			Object:         t.Object,
			Text:           FormatTripletText(t.Subject, t.Predicate, t.Object),
			Confidence:     t.Confidence,
			CreatedAt:      now,
			FeedbackWeight: 0.5,
		})
	}

	return triplets
}

// ExtractAndBuild is a convenience that runs extraction and mapping in one call.
func (s *TripletExtractionService) ExtractAndBuild(ctx context.Context, memoryID, content string) ([]domain.Triplet, error) {
	extracted, err := s.ExtractFromContent(ctx, content)
	if err != nil {
		return nil, err
	}
	return s.BuildTriplets(ctx, memoryID, extracted), nil
}
