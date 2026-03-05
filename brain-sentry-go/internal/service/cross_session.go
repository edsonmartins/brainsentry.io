package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// RedactionLevel controls how much detail is preserved in cross-session entries.
type RedactionLevel string

const (
	RedactionNone    RedactionLevel = "none"    // full content preserved
	RedactionPartial RedactionLevel = "partial" // PII removed, content preserved
	RedactionFull    RedactionLevel = "full"    // only summaries, no raw content
)

// SessionEvent represents a recorded event during a session.
type SessionEvent struct {
	ID        string          `json:"id"`
	SessionID string          `json:"sessionId"`
	Type      domain.ObservationType `json:"type"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
}

// CrossSessionEntry represents a memory entry derived from session analysis.
type CrossSessionEntry struct {
	ID              string                 `json:"id"`
	SourceSessionID string                 `json:"sourceSessionId"`
	Type            domain.ObservationType `json:"type"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	SupersededBy    string                 `json:"supersededBy,omitempty"`
	Provenance      []string               `json:"provenance"` // chain of source IDs
	CreatedAt       time.Time              `json:"createdAt"`
}

// CrossSessionResult represents the outcome of cross-session processing.
type CrossSessionResult struct {
	SessionID          string              `json:"sessionId"`
	EventsRecorded     int                 `json:"eventsRecorded"`
	ObservationsFound  int                 `json:"observationsFound"`
	EntriesCreated     int                 `json:"entriesCreated"`
	EntriesSuperseded  int                 `json:"entriesSuperseded"`
	ContextInjected    string              `json:"contextInjected,omitempty"`
	TokensInjected     int                 `json:"tokensInjected"`
}

// SessionLifecycleHook is a callback for session lifecycle events.
type SessionLifecycleHook func(ctx context.Context, sessionID string) error

// CrossSessionService manages cross-session memory pipeline with lifecycle hooks.
type CrossSessionService struct {
	memoryRepo     *postgres.MemoryRepository
	openRouter     *OpenRouterService
	piiService     *PIIService
	sessionService *SessionService

	// In-memory event buffer per session
	eventBuffers map[string][]SessionEvent

	// Lifecycle hooks
	onSessionStart []SessionLifecycleHook
	onSessionEnd   []SessionLifecycleHook

	// Configuration
	redactionLevel  RedactionLevel
	tokenBudget     int
	maxContextItems int
	lookbackDays    int
}

// NewCrossSessionService creates a new CrossSessionService.
func NewCrossSessionService(
	memoryRepo *postgres.MemoryRepository,
	openRouter *OpenRouterService,
	sessionService *SessionService,
) *CrossSessionService {
	return &CrossSessionService{
		memoryRepo:      memoryRepo,
		openRouter:      openRouter,
		piiService:      NewPIIService(),
		sessionService:  sessionService,
		eventBuffers:    make(map[string][]SessionEvent),
		redactionLevel:  RedactionPartial,
		tokenBudget:     1500,
		maxContextItems: 10,
		lookbackDays:    7,
	}
}

// RegisterStartHook adds a hook called when a session starts.
func (s *CrossSessionService) RegisterStartHook(hook SessionLifecycleHook) {
	s.onSessionStart = append(s.onSessionStart, hook)
}

// RegisterEndHook adds a hook called when a session ends.
func (s *CrossSessionService) RegisterEndHook(hook SessionLifecycleHook) {
	s.onSessionEnd = append(s.onSessionEnd, hook)
}

// OnSessionStart handles session start: injects cross-session context.
func (s *CrossSessionService) OnSessionStart(ctx context.Context, sessionID string) (*CrossSessionResult, error) {
	result := &CrossSessionResult{SessionID: sessionID}

	// Run registered hooks
	for _, hook := range s.onSessionStart {
		if err := hook(ctx, sessionID); err != nil {
			slog.Warn("session start hook failed", "error", err)
		}
	}

	// Inject context from previous sessions
	context, tokens := s.buildCrossSessionContext(ctx)
	result.ContextInjected = context
	result.TokensInjected = tokens

	// Initialize event buffer for this session
	s.eventBuffers[sessionID] = make([]SessionEvent, 0)

	return result, nil
}

// RecordEvent records a typed event during the active session.
func (s *CrossSessionService) RecordEvent(sessionID string, eventType domain.ObservationType, title, content string, metadata map[string]any) {
	event := SessionEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Title:     title,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	s.eventBuffers[sessionID] = append(s.eventBuffers[sessionID], event)
}

// OnSessionEnd handles session end: extracts observations and stores cross-session entries.
func (s *CrossSessionService) OnSessionEnd(ctx context.Context, sessionID string) (*CrossSessionResult, error) {
	result := &CrossSessionResult{SessionID: sessionID}

	events := s.eventBuffers[sessionID]
	result.EventsRecorded = len(events)

	// Run registered hooks
	for _, hook := range s.onSessionEnd {
		if err := hook(ctx, sessionID); err != nil {
			slog.Warn("session end hook failed", "error", err)
		}
	}

	if len(events) == 0 {
		delete(s.eventBuffers, sessionID)
		return result, nil
	}

	// Extract observations from events
	observations := s.extractObservations(ctx, events)
	result.ObservationsFound = len(observations)

	// Store as cross-session memories (requires context and repo)
	if ctx != nil && s.memoryRepo != nil {
		tenantID := tenant.FromContext(ctx)
		for _, obs := range observations {
			entry := s.createCrossSessionEntry(ctx, sessionID, obs, tenantID)
			if entry != nil {
				result.EntriesCreated++
			}
		}

		// Check for supersession of existing entries
		superseded := s.checkSupersession(ctx, observations, tenantID)
		result.EntriesSuperseded = superseded
	}

	// Cleanup event buffer
	delete(s.eventBuffers, sessionID)

	slog.Info("cross-session processing completed",
		"session", sessionID,
		"events", result.EventsRecorded,
		"observations", result.ObservationsFound,
		"entries_created", result.EntriesCreated,
		"superseded", result.EntriesSuperseded,
	)

	return result, nil
}

// GetSessionEvents returns recorded events for a session.
func (s *CrossSessionService) GetSessionEvents(sessionID string) []SessionEvent {
	return s.eventBuffers[sessionID]
}

// buildCrossSessionContext builds context from previous sessions within lookback window.
func (s *CrossSessionService) buildCrossSessionContext(ctx context.Context) (string, int) {
	if s.memoryRepo == nil {
		return "", 0
	}

	// Fetch recent cross-session memories (EPISODIC type with recent observations)
	memories, err := s.memoryRepo.FullTextSearch(ctx, "", s.maxContextItems)
	if err != nil || len(memories) == 0 {
		return "", 0
	}

	// Filter to lookback window and active memories
	cutoff := time.Now().AddDate(0, 0, -s.lookbackDays)
	var relevant []domain.Memory
	for _, m := range memories {
		if m.CreatedAt.Before(cutoff) {
			continue
		}
		if m.SupersededBy != "" {
			continue
		}
		relevant = append(relevant, m)
	}

	if len(relevant) == 0 {
		return "", 0
	}

	// Build context with token budget
	var sb strings.Builder
	sb.WriteString("<cross_session_context>\nRelevant context from previous sessions:\n\n")
	usedTokens := estimateTokens(sb.String()) + estimateTokens("</cross_session_context>")

	for _, m := range relevant {
		var entry strings.Builder
		entry.WriteString(fmt.Sprintf("- [%s] %s", m.MemoryType, m.Summary))
		if m.Summary == "" {
			entry.WriteString(truncate(m.Content, 150))
		}
		entry.WriteString("\n")

		entryTokens := estimateTokens(entry.String())
		if usedTokens+entryTokens > s.tokenBudget {
			break
		}
		sb.WriteString(entry.String())
		usedTokens += entryTokens
	}

	sb.WriteString("</cross_session_context>")

	contextStr := sb.String()

	// Apply redaction
	contextStr = s.applyRedaction(contextStr)

	return contextStr, estimateTokens(contextStr)
}

// extractObservations extracts typed observations from session events.
func (s *CrossSessionService) extractObservations(ctx context.Context, events []SessionEvent) []CrossSessionEntry {
	if len(events) == 0 {
		return nil
	}

	// If we have LLM access, use it for richer extraction
	if s.openRouter != nil {
		return s.llmExtractObservations(ctx, events)
	}

	// Fallback: convert events directly to entries
	return s.directExtractObservations(events)
}

func (s *CrossSessionService) llmExtractObservations(ctx context.Context, events []SessionEvent) []CrossSessionEntry {
	var eventSummary strings.Builder
	for _, e := range events {
		eventSummary.WriteString(fmt.Sprintf("[%s] %s: %s\n", e.Type, e.Title, truncate(e.Content, 200)))
	}

	prompt := fmt.Sprintf(`Analyze these session events and extract key observations.
For each observation, identify:
1. Type: DECISION, BUGFIX, FEATURE, REFACTOR, DISCOVERY, or CHANGE
2. A concise title
3. A self-contained description (no pronouns, include full context)

Events:
%s

Respond as plain text, one observation per line in format:
TYPE|title|description`, eventSummary.String())

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You extract structured observations from session events. Be concise and factual."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		slog.Warn("LLM observation extraction failed, using direct", "error", err)
		return s.directExtractObservations(events)
	}

	var entries []CrossSessionEntry
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}

		obsType := parseObservationType(strings.TrimSpace(parts[0]))
		entries = append(entries, CrossSessionEntry{
			ID:    uuid.New().String(),
			Type:  obsType,
			Title: strings.TrimSpace(parts[1]),
			Content: strings.TrimSpace(parts[2]),
			CreatedAt: time.Now(),
		})
	}

	if len(entries) == 0 {
		return s.directExtractObservations(events)
	}

	return entries
}

func (s *CrossSessionService) directExtractObservations(events []SessionEvent) []CrossSessionEntry {
	entries := make([]CrossSessionEntry, 0, len(events))
	for _, e := range events {
		entries = append(entries, CrossSessionEntry{
			ID:      uuid.New().String(),
			Type:    e.Type,
			Title:   e.Title,
			Content: e.Content,
			CreatedAt: time.Now(),
		})
	}
	return entries
}

// createCrossSessionEntry stores a cross-session observation as a memory.
func (s *CrossSessionService) createCrossSessionEntry(ctx context.Context, sessionID string, entry CrossSessionEntry, tenantID string) *CrossSessionEntry {
	if s.memoryRepo == nil {
		return nil
	}

	content := entry.Content
	if s.redactionLevel == RedactionFull {
		content = entry.Title // only title for full redaction
	} else if s.redactionLevel == RedactionPartial && s.piiService != nil {
		content, _ = s.piiService.MaskForLLM(content)
	}

	entry.SourceSessionID = sessionID
	entry.Provenance = []string{sessionID, entry.ID}

	// Create as EPISODIC memory with observation type in tags
	memory := &domain.Memory{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Content:   content,
		Summary:   fmt.Sprintf("[%s] %s", entry.Type, entry.Title),
		Category:  domain.CategoryKnowledge,
		MemoryType: domain.MemoryTypeEpisodic,
		Importance: observationImportance(entry.Type),
		Tags:      []string{"cross-session", string(entry.Type), "session:" + sessionID},
		CreatedAt: time.Now(),
		Version:   1,
	}

	if err := s.memoryRepo.Create(ctx, memory); err != nil {
		slog.Warn("failed to store cross-session entry", "error", err)
		return nil
	}

	return &entry
}

// checkSupersession checks if new observations supersede existing entries.
func (s *CrossSessionService) checkSupersession(ctx context.Context, observations []CrossSessionEntry, tenantID string) int {
	if s.memoryRepo == nil || len(observations) == 0 {
		return 0
	}

	superseded := 0
	for _, obs := range observations {
		// Search for similar existing memories
		existing, err := s.memoryRepo.FullTextSearch(ctx, obs.Title, 3)
		if err != nil {
			continue
		}

		for _, m := range existing {
			// Only supersede same-type cross-session entries
			if m.MemoryType != domain.MemoryTypeEpisodic {
				continue
			}
			if !hasTag(m.Tags, "cross-session") {
				continue
			}
			if !hasTag(m.Tags, string(obs.Type)) {
				continue
			}
			// Same type observation with similar title → supersede
			if m.SupersededBy == "" {
				if err := s.memoryRepo.SupersedeMemory(ctx, m.ID, obs.ID); err == nil {
					superseded++
				}
			}
		}
	}

	return superseded
}

// applyRedaction applies the configured redaction level.
func (s *CrossSessionService) applyRedaction(text string) string {
	switch s.redactionLevel {
	case RedactionPartial:
		if s.piiService != nil {
			masked, _ := s.piiService.MaskForLLM(text)
			return masked
		}
	case RedactionFull:
		// Strip all content, keep only structure markers
		lines := strings.Split(text, "\n")
		var result []string
		for _, line := range lines {
			if strings.HasPrefix(line, "<") || strings.HasPrefix(line, "- [") || strings.Contains(line, "context") {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n")
	}
	return text
}

func parseObservationType(s string) domain.ObservationType {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DECISION":
		return domain.ObservationDecision
	case "BUGFIX":
		return domain.ObservationBugfix
	case "FEATURE":
		return domain.ObservationFeature
	case "REFACTOR":
		return domain.ObservationRefactor
	case "DISCOVERY":
		return domain.ObservationDiscovery
	case "CHANGE":
		return domain.ObservationChange
	default:
		return domain.ObservationDiscovery
	}
}

func observationImportance(t domain.ObservationType) domain.ImportanceLevel {
	switch t {
	case domain.ObservationDecision:
		return domain.ImportanceCritical
	case domain.ObservationBugfix:
		return domain.ImportanceImportant
	case domain.ObservationFeature:
		return domain.ImportanceImportant
	case domain.ObservationRefactor:
		return domain.ImportanceMinor
	case domain.ObservationDiscovery:
		return domain.ImportanceMinor
	case domain.ObservationChange:
		return domain.ImportanceMinor
	default:
		return domain.ImportanceMinor
	}
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
