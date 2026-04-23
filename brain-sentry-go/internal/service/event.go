package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// EventService manages Event records and triggers LLM-based extraction from
// Memory content when requested.
type EventService struct {
	repo     *postgres.EventRepository
	embedder *EmbeddingService
	llm      LLMProvider
	audit    *AuditService
	tracker  ProvenanceTracker
}

// NewEventService wires dependencies; llm and embedder are optional.
func NewEventService(repo *postgres.EventRepository, embedder *EmbeddingService, llm LLMProvider, audit *AuditService) *EventService {
	return &EventService{
		repo:     repo,
		embedder: embedder,
		llm:      llm,
		audit:    audit,
		tracker:  NewProvenanceTracker("event", audit),
	}
}

// RecordEventRequest is the input for Record.
type RecordEventRequest struct {
	EventType      string                    `json:"eventType"`
	Title          string                    `json:"title"`
	Description    string                    `json:"description"`
	OccurredAt     time.Time                 `json:"occurredAt"`
	Participants   []domain.EventParticipant `json:"participants"`
	Attributes     map[string]any            `json:"attributes,omitempty"`
	SourceMemoryID string                    `json:"sourceMemoryId,omitempty"`
}

// Record persists an event and generates its embedding.
func (s *EventService) Record(ctx context.Context, req RecordEventRequest) (*domain.Event, error) {
	if req.EventType == "" {
		return nil, fmt.Errorf("eventType is required")
	}
	e := &domain.Event{
		ID:             uuid.NewString(),
		EventType:      req.EventType,
		Title:          req.Title,
		Description:    req.Description,
		OccurredAt:     req.OccurredAt,
		Participants:   req.Participants,
		SourceMemoryID: req.SourceMemoryID,
	}
	if len(req.Attributes) > 0 {
		raw, _ := json.Marshal(req.Attributes)
		e.Attributes = raw
	}
	if s.embedder != nil {
		e.Embedding = s.embedder.Embed(buildEventText(e))
	}
	if err := s.tracker.Track(ctx, "record", e.ID, func(ctx context.Context) error {
		return s.repo.Create(ctx, e)
	}); err != nil {
		return nil, err
	}
	if s.audit != nil {
		_ = s.audit.LogEvent(ctx, "event.recorded", map[string]any{
			"eventId":   e.ID,
			"eventType": e.EventType,
			"title":     e.Title,
		})
	}
	return e, nil
}

// Get fetches an event.
func (s *EventService) Get(ctx context.Context, id string) (*domain.Event, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns events matching the filter.
func (s *EventService) List(ctx context.Context, f postgres.EventFilter) ([]*domain.Event, error) {
	return s.repo.List(ctx, f)
}

// Delete removes an event.
func (s *EventService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Stats returns the event-type histogram for the tenant.
func (s *EventService) Stats(ctx context.Context) (map[string]int, error) {
	return s.repo.CountByType(ctx)
}

// ExtractFromText asks the LLM to extract events from a free-text content
// and persists the results. Uses best-effort JSON parsing; returns partial
// results on malformed output.
func (s *EventService) ExtractFromText(ctx context.Context, content, sourceMemoryID string) ([]*domain.Event, error) {
	if s.llm == nil {
		return nil, fmt.Errorf("LLM provider not configured")
	}
	prompt := fmt.Sprintf(`Extract structured events from the text below. Return ONLY a JSON array of objects with:
  - eventType (string, UPPER_SNAKE)
  - title (short sentence)
  - description (one to two sentences)
  - occurredAt (ISO8601 or empty)
  - participants: array of { entityId, role, label }
Text:
"""
%s
"""`, content)
	raw, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You extract structured events from text. Respond with JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("LLM did not return a JSON array")
	}
	var payload []RecordEventRequest
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON from LLM: %w", err)
	}
	out := make([]*domain.Event, 0, len(payload))
	for _, p := range payload {
		p.SourceMemoryID = sourceMemoryID
		if p.OccurredAt.IsZero() {
			p.OccurredAt = time.Now()
		}
		e, err := s.Record(ctx, p)
		if err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func buildEventText(e *domain.Event) string {
	parts := []string{e.EventType, e.Title, e.Description}
	for _, p := range e.Participants {
		parts = append(parts, p.Label)
	}
	return strings.Join(trimEmpty(parts), " | ")
}
