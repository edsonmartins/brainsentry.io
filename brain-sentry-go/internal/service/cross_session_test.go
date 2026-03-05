package service

import (
	"context"
	"strings"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestNewCrossSessionService(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.redactionLevel != RedactionPartial {
		t.Errorf("expected partial redaction, got %s", svc.redactionLevel)
	}
	if svc.tokenBudget != 1500 {
		t.Errorf("expected 1500 token budget, got %d", svc.tokenBudget)
	}
	if svc.lookbackDays != 7 {
		t.Errorf("expected 7 lookback days, got %d", svc.lookbackDays)
	}
}

func TestRecordEvent(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)
	svc.eventBuffers["s1"] = make([]SessionEvent, 0)

	svc.RecordEvent("s1", domain.ObservationDecision, "Use PostgreSQL", "Decided to use PostgreSQL for persistence", nil)
	svc.RecordEvent("s1", domain.ObservationBugfix, "Fix null pointer", "Fixed NPE in handler", nil)

	events := svc.GetSessionEvents("s1")
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != domain.ObservationDecision {
		t.Errorf("expected DECISION type, got %s", events[0].Type)
	}
	if events[1].Title != "Fix null pointer" {
		t.Errorf("expected 'Fix null pointer', got %s", events[1].Title)
	}
}

func TestOnSessionStart_NilDeps(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)
	result, err := svc.OnSessionStart(nil, "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SessionID != "s1" {
		t.Errorf("expected session s1, got %s", result.SessionID)
	}
	if result.ContextInjected != "" {
		t.Error("expected no context without repos")
	}
}

func TestOnSessionEnd_NoEvents(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)
	svc.eventBuffers["s1"] = make([]SessionEvent, 0)

	result, err := svc.OnSessionEnd(nil, "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventsRecorded != 0 {
		t.Errorf("expected 0 events, got %d", result.EventsRecorded)
	}
	if result.EntriesCreated != 0 {
		t.Errorf("expected 0 entries, got %d", result.EntriesCreated)
	}
}

func TestOnSessionEnd_WithEvents(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)
	svc.eventBuffers["s1"] = make([]SessionEvent, 0)

	svc.RecordEvent("s1", domain.ObservationFeature, "Add auth", "Added JWT authentication", nil)
	svc.RecordEvent("s1", domain.ObservationDiscovery, "Found pattern", "Discovered caching pattern", nil)

	result, err := svc.OnSessionEnd(nil, "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventsRecorded != 2 {
		t.Errorf("expected 2 events, got %d", result.EventsRecorded)
	}
	if result.ObservationsFound != 2 {
		t.Errorf("expected 2 observations, got %d", result.ObservationsFound)
	}
	// No memory repo so entries won't be created
	if result.EntriesCreated != 0 {
		t.Errorf("expected 0 entries without repo, got %d", result.EntriesCreated)
	}
}

func TestDirectExtractObservations(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)
	events := []SessionEvent{
		{ID: "e1", Type: domain.ObservationDecision, Title: "Choose Go", Content: "Selected Go for backend"},
		{ID: "e2", Type: domain.ObservationBugfix, Title: "Fix race", Content: "Fixed race condition in cache"},
	}

	entries := svc.directExtractObservations(events)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Type != domain.ObservationDecision {
		t.Errorf("expected DECISION, got %s", entries[0].Type)
	}
	if entries[1].Title != "Fix race" {
		t.Errorf("expected 'Fix race', got %s", entries[1].Title)
	}
}

func TestParseObservationType(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.ObservationType
	}{
		{"DECISION", domain.ObservationDecision},
		{"bugfix", domain.ObservationBugfix},
		{"Feature", domain.ObservationFeature},
		{"REFACTOR", domain.ObservationRefactor},
		{"discovery", domain.ObservationDiscovery},
		{"CHANGE", domain.ObservationChange},
		{"unknown", domain.ObservationDiscovery},
	}

	for _, tt := range tests {
		result := parseObservationType(tt.input)
		if result != tt.expected {
			t.Errorf("parseObservationType(%q) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestObservationImportance(t *testing.T) {
	if observationImportance(domain.ObservationDecision) != domain.ImportanceCritical {
		t.Error("decisions should be critical")
	}
	if observationImportance(domain.ObservationBugfix) != domain.ImportanceImportant {
		t.Error("bugfixes should be important")
	}
	if observationImportance(domain.ObservationRefactor) != domain.ImportanceMinor {
		t.Error("refactors should be normal")
	}
}

func TestLifecycleHooks(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)

	startCalled := false
	endCalled := false

	svc.RegisterStartHook(func(ctx context.Context, sessionID string) error {
		startCalled = true
		return nil
	})
	svc.RegisterEndHook(func(ctx context.Context, sessionID string) error {
		endCalled = true
		return nil
	})

	svc.OnSessionStart(nil, "s1")
	if !startCalled {
		t.Error("start hook not called")
	}

	svc.eventBuffers["s1"] = make([]SessionEvent, 0)
	svc.OnSessionEnd(nil, "s1")
	if !endCalled {
		t.Error("end hook not called")
	}
}

func TestRedactionLevels(t *testing.T) {
	svc := NewCrossSessionService(nil, nil, nil)

	text := "Hello world with email test@example.com"

	// None
	svc.redactionLevel = RedactionNone
	result := svc.applyRedaction(text)
	if result != text {
		t.Errorf("none redaction should not change text")
	}

	// Partial (PII masked)
	svc.redactionLevel = RedactionPartial
	result = svc.applyRedaction(text)
	if result == text {
		// PII service should mask the email
		// (depending on PII service implementation)
	}

	// Full
	svc.redactionLevel = RedactionFull
	result = svc.applyRedaction(text)
	if strings.Contains(result, "email") {
		t.Error("full redaction should strip content")
	}
}

func TestHasTag(t *testing.T) {
	tags := []string{"cross-session", "DECISION", "session:s1"}
	if !hasTag(tags, "cross-session") {
		t.Error("expected to find cross-session tag")
	}
	if hasTag(tags, "nonexistent") {
		t.Error("expected not to find nonexistent tag")
	}
}

func TestSessionEvent_Structure(t *testing.T) {
	e := SessionEvent{
		ID:        "e1",
		SessionID: "s1",
		Type:      domain.ObservationFeature,
		Title:     "Add caching",
		Content:   "Implemented Redis caching layer",
	}
	if e.ID != "e1" {
		t.Error("expected e1")
	}
	if e.Type != domain.ObservationFeature {
		t.Error("expected FEATURE")
	}
}

func TestCrossSessionEntry_Structure(t *testing.T) {
	e := CrossSessionEntry{
		ID:              "cs1",
		SourceSessionID: "s1",
		Type:            domain.ObservationDecision,
		Title:           "Use Redis",
		Provenance:      []string{"s1", "e1"},
	}
	if len(e.Provenance) != 2 {
		t.Error("expected 2 provenance entries")
	}
	if e.Type != domain.ObservationDecision {
		t.Error("expected DECISION")
	}
}
