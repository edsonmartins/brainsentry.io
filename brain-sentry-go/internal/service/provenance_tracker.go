package service

import (
	"context"
	"time"
)

// ProvenanceEvent is a minimal, language-level description of an action taken
// by any service. When a service wants its operations to appear in the audit
// trail + PROV-O export without writing boilerplate, it embeds a tracker and
// calls Track around the unit of work.
type ProvenanceEvent struct {
	Activity    string         // short stable id, e.g. "memory.create"
	Subject     string         // entity id impacted, when known
	Inputs      map[string]any // sanitised inputs (no raw prompts)
	Outputs     map[string]any // sanitised outputs
	Reason      string         // why this was executed
	StartedAt   time.Time
	EndedAt     time.Time
	Status      string // "success" | "error"
	ErrorMessage string
}

// ProvenanceTracker wraps AuditService with a generic Track helper. Any
// service can embed it and gain consistent audit logging via one call.
type ProvenanceTracker struct {
	audit *AuditService
	name  string
}

// NewProvenanceTracker constructs a tracker scoped to the given logical name.
func NewProvenanceTracker(name string, audit *AuditService) ProvenanceTracker {
	return ProvenanceTracker{audit: audit, name: name}
}

// Track runs fn and records a structured event afterwards. Returns whatever
// fn returned. Safe to use even when audit is nil — events are then discarded.
func (t ProvenanceTracker) Track(ctx context.Context, activity string, subject string, fn func(ctx context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	t.Record(ctx, ProvenanceEvent{
		Activity:    t.name + "." + activity,
		Subject:     subject,
		StartedAt:   start,
		EndedAt:     time.Now(),
		Status:      statusFromErr(err),
		ErrorMessage: errString(err),
	})
	return err
}

// Record persists an event manually (for non-fn-based call sites).
func (t ProvenanceTracker) Record(ctx context.Context, ev ProvenanceEvent) {
	if t.audit == nil {
		return
	}
	payload := map[string]any{
		"activity":    ev.Activity,
		"subject":     ev.Subject,
		"startedAt":   ev.StartedAt.UTC().Format(time.RFC3339),
		"endedAt":     ev.EndedAt.UTC().Format(time.RFC3339),
		"status":      ev.Status,
		"durationMs":  ev.EndedAt.Sub(ev.StartedAt).Milliseconds(),
	}
	if ev.Reason != "" {
		payload["reason"] = ev.Reason
	}
	if ev.ErrorMessage != "" {
		payload["error"] = ev.ErrorMessage
	}
	if len(ev.Inputs) > 0 {
		payload["inputs"] = ev.Inputs
	}
	if len(ev.Outputs) > 0 {
		payload["outputs"] = ev.Outputs
	}
	_ = t.audit.LogEvent(ctx, ev.Activity, payload)
}

func statusFromErr(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
